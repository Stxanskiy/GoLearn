package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CourseRepo loads and stores a whole course tree (module + lessons + quiz +
// tasks) for the admin import/export. Upsert is transactional: either the whole
// course lands or nothing does.
type CourseRepo struct {
	pool    *pgxpool.Pool
	modules *ModuleRepo
	lessons *LessonRepo
}

func NewCourseRepo(pool *pgxpool.Pool, modules *ModuleRepo, lessons *LessonRepo) *CourseRepo {
	return &CourseRepo{pool: pool, modules: modules, lessons: lessons}
}

// LessonRef identifies a lesson in an import diff.
type LessonRef struct {
	Slug  string
	Title string
}

// CourseDiff describes what importing a tree would change.
type CourseDiff struct {
	Slug            string
	Title           string
	ModuleID        int // 0 when the course does not exist yet
	Exists          bool
	New             []LessonRef
	Updated         []LessonRef
	Removed         []LessonRef
	LostSubmissions int // submissions on tasks the import deletes
	LostProgress    int // progress rows on lessons the import deletes
}

// Export builds a CourseTree for a module id (read-only).
func (r *CourseRepo) Export(ctx context.Context, moduleID int) (model.CourseTree, error) {
	m, err := r.modules.GetByID(ctx, moduleID)
	if err != nil {
		return model.CourseTree{}, err
	}
	lessons, err := r.lessons.GetByModuleAll(ctx, moduleID)
	if err != nil {
		return model.CourseTree{}, err
	}
	tree := model.CourseTree{Module: *m}
	for _, l := range lessons {
		_, questions, _ := r.lessons.GetQuiz(ctx, l.ID) // no quiz -> nil, not fatal
		tasks, _ := r.lessons.GetTasks(ctx, l.ID)
		tree.Lessons = append(tree.Lessons, model.LessonBundle{Lesson: l, Questions: questions, Tasks: tasks})
	}
	return tree, nil
}

// Diff computes what importing tree into moduleID (0: the course with the tree slug) would change, without writing anything.
func (r *CourseRepo) Diff(ctx context.Context, tree model.CourseTree, moduleID int) (CourseDiff, error) {
	d := CourseDiff{Slug: tree.Module.Slug, Title: tree.Module.Title}
	var existing *model.Module
	var err error
	if moduleID != 0 {
		existing, err = r.modules.GetByID(ctx, moduleID)
	} else {
		existing, err = r.modules.GetBySlug(ctx, tree.Module.Slug)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		for _, lb := range tree.Lessons {
			d.New = append(d.New, LessonRef{lb.Lesson.Slug, lb.Lesson.Title})
		}
		return d, nil
	}
	if err != nil {
		return d, err
	}
	d.Exists, d.ModuleID = true, existing.ID
	cur, err := r.lessons.GetByModuleAll(ctx, existing.ID)
	if err != nil {
		return d, err
	}
	incoming := make(map[string]model.LessonBundle, len(tree.Lessons))
	curSlugs := make(map[string]bool, len(cur))
	for _, l := range cur {
		curSlugs[l.Slug] = true
	}
	for _, lb := range tree.Lessons {
		incoming[lb.Lesson.Slug] = lb
		if curSlugs[lb.Lesson.Slug] {
			d.Updated = append(d.Updated, LessonRef{lb.Lesson.Slug, lb.Lesson.Title})
		} else {
			d.New = append(d.New, LessonRef{lb.Lesson.Slug, lb.Lesson.Title})
		}
	}
	var lostTasks, removedLessons []int
	for _, l := range cur {
		tasks, err := r.lessons.GetTasks(ctx, l.ID)
		if err != nil {
			return d, err
		}
		lb, kept := incoming[l.Slug]
		if !kept {
			d.Removed = append(d.Removed, LessonRef{l.Slug, l.Title})
			removedLessons = append(removedLessons, l.ID)
		}
		_, unused := matchByKey(taskTitles(tasks), taskTitles(lb.Tasks))
		for _, i := range unused {
			lostTasks = append(lostTasks, tasks[i].ID)
		}
	}
	err = r.pool.QueryRow(ctx,
		`SELECT (SELECT count(*) FROM submissions WHERE task_id = ANY($1)),
		        (SELECT count(*) FROM progress WHERE lesson_id = ANY($2))`,
		lostTasks, removedLessons).Scan(&d.LostSubmissions, &d.LostProgress)
	return d, err
}

// matchByKey pairs each incoming key with an unused existing index holding the same key (-1 means insert) and returns the existing indexes left unmatched.
func matchByKey(existing, incoming []string) (pairs, unused []int) {
	free := map[string][]int{}
	for i, k := range existing {
		free[k] = append(free[k], i)
	}
	pairs = make([]int, len(incoming))
	used := make([]bool, len(existing))
	for i, k := range incoming {
		pairs[i] = -1
		if idx := free[k]; len(idx) > 0 {
			pairs[i], free[k] = idx[0], idx[1:]
			used[idx[0]] = true
		}
	}
	for i, u := range used {
		if !u {
			unused = append(unused, i)
		}
	}
	return pairs, unused
}

func taskTitles(tasks []model.Task) []string {
	out := make([]string, len(tasks))
	for i, t := range tasks {
		out[i] = t.Title
	}
	return out
}

// Upsert applies a course tree in one transaction keyed by slug; published and owner apply only to rows it inserts; lessons missing from the tree are deleted, questions and tasks are matched by text and title so student answers and submissions survive.
func (r *CourseRepo) Upsert(ctx context.Context, tree model.CourseTree, moduleID int) (CourseDiff, error) {
	d, err := r.Diff(ctx, tree, moduleID)
	if err != nil {
		return d, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return d, err
	}
	defer tx.Rollback(ctx)

	m := tree.Module
	tags, _ := json.Marshal(m.Tags)
	if m.Tags == nil {
		tags = []byte("[]")
	}
	if moduleID == 0 {
		err = tx.QueryRow(ctx, `SELECT id FROM modules WHERE slug=$1`, m.Slug).Scan(&moduleID)
	}
	switch err {
	case pgx.ErrNoRows:
		err = tx.QueryRow(ctx,
			`INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites,
			   category, label, tags, cover_image, accent, est_minutes, source, published, owner_id)
			 VALUES ($1,$2,$3,$4,$5,$6,'[]',$7,$8,$9,$10,$11,$12,'admin',$13,$14) RETURNING id`,
			m.Slug, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
			m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes, m.Published, m.OwnerID).Scan(&moduleID)
		if err != nil {
			return d, err
		}
	case nil:
		// keep an existing cover when the import omits one (cover is a heavy blob)
		_, err = tx.Exec(ctx,
			`UPDATE modules SET title=$2, description=$3, order_num=$4, track=$5, difficulty=$6,
			   category=$7, label=$8, tags=$9,
			   cover_image=CASE WHEN $10='' THEN cover_image ELSE $10 END,
			   accent=$11, est_minutes=$12 WHERE id=$1`,
			moduleID, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
			m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes)
		if err != nil {
			return d, err
		}
	default:
		return d, err
	}

	seen := make([]string, 0, len(tree.Lessons))
	for _, lb := range tree.Lessons {
		l := lb.Lesson
		seen = append(seen, l.Slug)
		var lessonID int
		err = tx.QueryRow(ctx, `SELECT id FROM lessons WHERE module_id=$1 AND slug=$2`, moduleID, l.Slug).Scan(&lessonID)
		switch err {
		case pgx.ErrNoRows:
			err = tx.QueryRow(ctx,
				`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'admin',$12) RETURNING id`,
				moduleID, l.Slug, l.Title, l.Content, l.OrderNum, l.Difficulty, m.Track, l.Kind, l.Format, l.VMImage, l.VMInit, l.Published).Scan(&lessonID)
			if err != nil {
				return d, err
			}
		case nil:
			_, err = tx.Exec(ctx,
				`UPDATE lessons SET title=$2, content=$3, order_num=$4, difficulty=$5, kind=$6, format=$7, vm_image=$8, vm_init=$9 WHERE id=$1`,
				lessonID, l.Title, l.Content, l.OrderNum, l.Difficulty, l.Kind, l.Format, l.VMImage, l.VMInit)
			if err != nil {
				return d, err
			}
		default:
			return d, err
		}

		if err = upsertQuestions(ctx, tx, lessonID, lb.Questions); err != nil {
			return d, err
		}
		if err = upsertTasks(ctx, tx, lessonID, lb.Tasks); err != nil {
			return d, err
		}
	}

	// drop lessons that are no longer in the tree (cascades quiz/tasks)
	if _, err = tx.Exec(ctx, `DELETE FROM lessons WHERE module_id=$1 AND NOT (slug = ANY($2))`, moduleID, seen); err != nil {
		return d, err
	}

	return d, tx.Commit(ctx)
}

// upsertQuestions makes the lesson quiz match questions, updating rows matched by question text in place.
func upsertQuestions(ctx context.Context, tx pgx.Tx, lessonID int, questions []model.QuizQuestion) error {
	if len(questions) == 0 {
		_, err := tx.Exec(ctx, `DELETE FROM quizzes WHERE lesson_id=$1`, lessonID)
		return err
	}
	var quizID int
	err := tx.QueryRow(ctx, `SELECT id FROM quizzes WHERE lesson_id=$1 ORDER BY id LIMIT 1`, lessonID).Scan(&quizID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `INSERT INTO quizzes (lesson_id, title) VALUES ($1,$2) RETURNING id`, lessonID, "Квиз").Scan(&quizID)
	}
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx,
		`SELECT q.id, q.question FROM quiz_questions q JOIN quizzes z ON z.id = q.quiz_id
		 WHERE z.lesson_id=$1 ORDER BY z.id, q.order_num, q.id`, lessonID)
	if err != nil {
		return err
	}
	var ids []int
	var texts []string
	for rows.Next() {
		var id int
		var text string
		if err := rows.Scan(&id, &text); err != nil {
			rows.Close()
			return err
		}
		ids, texts = append(ids, id), append(texts, text)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	incoming := make([]string, len(questions))
	for i, q := range questions {
		incoming[i] = q.Question
	}
	pairs, unused := matchByKey(texts, incoming)
	for _, i := range unused {
		if _, err := tx.Exec(ctx, `DELETE FROM quiz_questions WHERE id=$1`, ids[i]); err != nil {
			return err
		}
	}
	for i, q := range questions {
		opts, _ := json.Marshal(q.Options)
		oexpl, _ := json.Marshal(q.OptionExpl)
		if pairs[i] >= 0 {
			_, err = tx.Exec(ctx,
				`UPDATE quiz_questions SET quiz_id=$2, question=$3, options=$4, option_explanations=$5, correct_index=$6, explanation=$7, order_num=$8 WHERE id=$1`,
				ids[pairs[i]], quizID, q.Question, opts, oexpl, q.CorrectIndex, q.Explanation, i+1)
		} else {
			_, err = tx.Exec(ctx,
				`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num)
				 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
				quizID, q.Question, opts, oexpl, q.CorrectIndex, q.Explanation, i+1)
		}
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `DELETE FROM quizzes WHERE lesson_id=$1 AND id<>$2`, lessonID, quizID)
	return err
}

// upsertTasks makes the lesson tasks match tasks, updating rows matched by title in place.
func upsertTasks(ctx context.Context, tx pgx.Tx, lessonID int, tasks []model.Task) error {
	rows, err := tx.Query(ctx, `SELECT id, title FROM tasks WHERE lesson_id=$1 ORDER BY order_num, id`, lessonID)
	if err != nil {
		return err
	}
	var ids []int
	var titles []string
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			rows.Close()
			return err
		}
		ids, titles = append(ids, id), append(titles, title)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	pairs, unused := matchByKey(titles, taskTitles(tasks))
	for _, i := range unused {
		if _, err := tx.Exec(ctx, `DELETE FROM tasks WHERE id=$1`, ids[i]); err != nil {
			return err
		}
	}
	for i, tk := range tasks {
		gloss, _ := json.Marshal(tk.Glossary)
		tc, _ := json.Marshal(tk.TestCases)
		if pairs[i] >= 0 {
			_, err = tx.Exec(ctx,
				`UPDATE tasks SET title=$2, description=$3, hints=$4, solution=$5, order_num=$6, difficulty=$7, glossary=$8, test_cases=$9,
				   starter_code=$10, format=$11, kind=$12, sandbox_image=$13, setup_script=$14, check_script=$15 WHERE id=$1`,
				ids[pairs[i]], tk.Title, tk.Description, tk.Hints, tk.Solution, i+1, tk.Difficulty, gloss, tc, tk.StarterCode, tk.Format, tk.Kind, tk.SandboxImage, tk.SetupScript, tk.CheckScript)
		} else {
			_, err = tx.Exec(ctx,
				`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
				lessonID, tk.Title, tk.Description, tk.Hints, tk.Solution, i+1, tk.Difficulty, gloss, tc, tk.StarterCode, tk.Format, tk.Kind, tk.SandboxImage, tk.SetupScript, tk.CheckScript)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
