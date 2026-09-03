package repository

import (
	"context"
	"encoding/json"

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

// CourseDiff describes what importing a tree would change, by lesson title.
type CourseDiff struct {
	Slug     string
	Title    string
	Exists   bool
	New      []string
	Updated  []string
	Removed  []string
	NewCount int
	UpdCount int
	DelCount int
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

// Diff computes what an import of tree would change, without writing anything.
func (r *CourseRepo) Diff(ctx context.Context, tree model.CourseTree) (CourseDiff, error) {
	d := CourseDiff{Slug: tree.Module.Slug, Title: tree.Module.Title}
	existing, err := r.modules.GetBySlug(ctx, tree.Module.Slug)
	if err != nil { // module not found -> everything is new
		for _, lb := range tree.Lessons {
			d.New = append(d.New, lb.Lesson.Title)
		}
		d.NewCount = len(d.New)
		return d, nil
	}
	d.Exists = true
	cur, err := r.lessons.GetByModuleAll(ctx, existing.ID)
	if err != nil {
		return d, err
	}
	curSlugs := make(map[string]bool, len(cur))
	for _, l := range cur {
		curSlugs[l.Slug] = true
	}
	newSlugs := make(map[string]bool, len(tree.Lessons))
	for _, lb := range tree.Lessons {
		newSlugs[lb.Lesson.Slug] = true
		if curSlugs[lb.Lesson.Slug] {
			d.Updated = append(d.Updated, lb.Lesson.Title)
		} else {
			d.New = append(d.New, lb.Lesson.Title)
		}
	}
	for _, l := range cur {
		if !newSlugs[l.Slug] {
			d.Removed = append(d.Removed, l.Title)
		}
	}
	d.NewCount, d.UpdCount, d.DelCount = len(d.New), len(d.Updated), len(d.Removed)
	return d, nil
}

// Upsert applies a course tree in a single transaction, keyed by slug: existing
// module/lessons are updated, missing ones inserted, and lessons no longer in the
// tree deleted (cascading to their quiz/tasks). Quiz and tasks of each kept lesson
// are replaced wholesale. Returns the diff that was applied.
func (r *CourseRepo) Upsert(ctx context.Context, tree model.CourseTree) (CourseDiff, error) {
	d, err := r.Diff(ctx, tree)
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
	var moduleID int
	err = tx.QueryRow(ctx, `SELECT id FROM modules WHERE slug=$1`, m.Slug).Scan(&moduleID)
	switch err {
	case pgx.ErrNoRows:
		err = tx.QueryRow(ctx,
			`INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites,
			   category, label, tags, cover_image, accent, est_minutes, source)
			 VALUES ($1,$2,$3,$4,$5,$6,'[]',$7,$8,$9,$10,$11,$12,'admin') RETURNING id`,
			m.Slug, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
			m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes).Scan(&moduleID)
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
				`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'admin') RETURNING id`,
				moduleID, l.Slug, l.Title, l.Content, l.OrderNum, l.Difficulty, l.Track, l.Kind, l.Format, l.VMImage, l.VMInit).Scan(&lessonID)
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

		// replace quiz (delete cascades questions), recreate only if there are questions
		if _, err = tx.Exec(ctx, `DELETE FROM quizzes WHERE lesson_id=$1`, lessonID); err != nil {
			return d, err
		}
		if len(lb.Questions) > 0 {
			var quizID int
			if err = tx.QueryRow(ctx, `INSERT INTO quizzes (lesson_id, title) VALUES ($1,$2) RETURNING id`, lessonID, "Квиз").Scan(&quizID); err != nil {
				return d, err
			}
			for _, q := range lb.Questions {
				opts, _ := json.Marshal(q.Options)
				oexpl, _ := json.Marshal(q.OptionExpl)
				if _, err = tx.Exec(ctx,
					`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num)
					 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
					quizID, q.Question, opts, oexpl, q.CorrectIndex, q.Explanation, q.OrderNum); err != nil {
					return d, err
				}
			}
		}

		// replace tasks wholesale
		if _, err = tx.Exec(ctx, `DELETE FROM tasks WHERE lesson_id=$1`, lessonID); err != nil {
			return d, err
		}
		for _, tk := range lb.Tasks {
			gloss, _ := json.Marshal(tk.Glossary)
			tc, _ := json.Marshal(tk.TestCases)
			if _, err = tx.Exec(ctx,
				`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
				lessonID, tk.Title, tk.Description, tk.Hints, tk.Solution, tk.OrderNum, tk.Difficulty, gloss, tc, tk.StarterCode, tk.Format, tk.Kind, tk.SandboxImage, tk.SetupScript, tk.CheckScript); err != nil {
				return d, err
			}
		}
	}

	// drop lessons that are no longer in the tree (cascades quiz/tasks)
	if _, err = tx.Exec(ctx, `DELETE FROM lessons WHERE module_id=$1 AND NOT (slug = ANY($2))`, moduleID, seen); err != nil {
		return d, err
	}

	return d, tx.Commit(ctx)
}
