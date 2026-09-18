package repository

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoDraft means the course has no draft.
var ErrNoDraft = errors.New("course has no draft")

// DraftRepo manages editable copies of published courses.
type DraftRepo struct {
	pool *pgxpool.Pool
}

func NewDraftRepo(pool *pgxpool.Pool) *DraftRepo {
	return &DraftRepo{pool: pool}
}

// Create copies a live course with its lessons, quiz and tasks into a hidden draft and returns the draft id.
func (r *DraftRepo) Create(ctx context.Context, liveID int) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	var draftID int
	err = tx.QueryRow(ctx, `
		INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites, category, label,
		  tags, cover_image, icon_url, accent, est_minutes, source, published, owner_id, draft_of)
		SELECT '~draft-' || id, title, description, order_num, track, difficulty, prerequisites, category, label,
		  tags, cover_image, icon_url, accent, est_minutes, 'admin', FALSE, owner_id, id
		FROM modules WHERE id = $1 AND draft_of IS NULL RETURNING id`, liveID).Scan(&draftID)
	if err != nil {
		return 0, err
	}
	steps := []step{
		{`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published, origin_id)
		 SELECT $2, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published, id
		 FROM lessons WHERE module_id = $1`, []any{liveID, draftID}},
		{`INSERT INTO quizzes (lesson_id, title)
		 SELECT DISTINCT ON (q.lesson_id) dl.id, q.title FROM quizzes q
		 JOIN lessons dl ON dl.origin_id = q.lesson_id AND dl.module_id = $1
		 ORDER BY q.lesson_id, q.id`, []any{draftID}},
		{`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num, origin_id)
		 SELECT dq.id, qq.question, qq.options, qq.option_explanations, qq.correct_index, qq.explanation, qq.order_num, qq.id
		 FROM quiz_questions qq
		 JOIN quizzes lq ON lq.id = qq.quiz_id
		 JOIN lessons dl ON dl.origin_id = lq.lesson_id AND dl.module_id = $1
		 JOIN quizzes dq ON dq.lesson_id = dl.id`, []any{draftID}},
		{`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases,
		   starter_code, kind, sandbox_image, setup_script, check_script, format, origin_id)
		 SELECT dl.id, t.title, t.description, t.hints, t.solution, t.order_num, t.difficulty, t.glossary, t.test_cases,
		   t.starter_code, t.kind, t.sandbox_image, t.setup_script, t.check_script, t.format, t.id
		 FROM tasks t JOIN lessons dl ON dl.origin_id = t.lesson_id AND dl.module_id = $1`, []any{draftID}},
	}
	if err := runSteps(ctx, tx, steps); err != nil {
		return 0, err
	}
	return draftID, tx.Commit(ctx)
}

// Discard deletes the draft of a live course and cancels its pending review.
func (r *DraftRepo) Discard(ctx context.Context, liveID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM modules WHERE draft_of = $1`, liveID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoDraft
	}
	if _, err := tx.Exec(ctx,
		`UPDATE review_requests SET status = 'cancelled', decided_at = NOW()
		 WHERE module_id = $1 AND kind = 'changes' AND status = 'pending'`, liveID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// mergeDraft applies a draft to its live course, keeping ids of lessons, questions and tasks that came from it, and deletes the draft.
func mergeDraft(ctx context.Context, tx pgx.Tx, liveID int) error {
	var draftID int
	err := tx.QueryRow(ctx, `SELECT id FROM modules WHERE draft_of = $1 FOR UPDATE`, liveID).Scan(&draftID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoDraft
	}
	if err != nil {
		return err
	}
	steps := []step{
		{`UPDATE modules l SET title = d.title, description = d.description, track = d.track, difficulty = d.difficulty,
		   category = d.category, label = d.label, tags = d.tags, cover_image = d.cover_image, icon_url = d.icon_url,
		   accent = d.accent, est_minutes = d.est_minutes
		 FROM modules d WHERE l.id = $1 AND d.id = $2`, []any{liveID, draftID}},
		{`UPDATE lessons SET slug = '~' || id WHERE module_id = $1`, []any{liveID}},
		{`UPDATE lessons l SET slug = d.slug, title = d.title, content = d.content, order_num = d.order_num, difficulty = d.difficulty,
		   track = d.track, kind = d.kind, format = d.format, vm_image = d.vm_image, vm_init = d.vm_init, published = d.published
		 FROM lessons d WHERE d.module_id = $2 AND l.id = d.origin_id AND l.module_id = $1`, []any{liveID, draftID}},
		{`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published, origin_id)
		 SELECT $1, d.slug, d.title, d.content, d.order_num, d.difficulty, d.track, d.kind, d.format, d.vm_image, d.vm_init, 'admin', d.published, d.id
		 FROM lessons d WHERE d.module_id = $2
		   AND NOT EXISTS (SELECT 1 FROM lessons l WHERE l.id = d.origin_id AND l.module_id = $1)`, []any{liveID, draftID}},
		{`DELETE FROM lessons WHERE module_id = $1 AND slug LIKE '~%'`, []any{liveID}},
	}
	if err := runSteps(ctx, tx, steps); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `
		SELECT d.id, l.id FROM lessons d
		JOIN lessons l ON l.module_id = $1 AND (l.id = d.origin_id OR l.origin_id = d.id)
		WHERE d.module_id = $2`, liveID, draftID)
	if err != nil {
		return err
	}
	type pair struct{ draft, live int }
	var pairs []pair
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.draft, &p.live); err != nil {
			rows.Close()
			return err
		}
		pairs = append(pairs, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, p := range pairs {
		if err := mergeQuiz(ctx, tx, p.draft, p.live); err != nil {
			return err
		}
		if err := mergeTasks(ctx, tx, p.draft, p.live); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE lessons SET origin_id = NULL WHERE module_id = $1`, liveID); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM modules WHERE id = $1`, draftID)
	return err
}

func mergeQuiz(ctx context.Context, tx pgx.Tx, draftLesson, liveLesson int) error {
	var questions int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM quiz_questions q JOIN quizzes z ON z.id = q.quiz_id WHERE z.lesson_id = $1`, draftLesson).Scan(&questions); err != nil {
		return err
	}
	if questions == 0 {
		_, err := tx.Exec(ctx, `DELETE FROM quizzes WHERE lesson_id = $1`, liveLesson)
		return err
	}
	var quizID int
	err := tx.QueryRow(ctx, `SELECT id FROM quizzes WHERE lesson_id = $1 ORDER BY id LIMIT 1`, liveLesson).Scan(&quizID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx,
			`INSERT INTO quizzes (lesson_id, title) SELECT $1, COALESCE((SELECT title FROM quizzes WHERE lesson_id = $2 ORDER BY id LIMIT 1), 'Квиз') RETURNING id`,
			liveLesson, draftLesson).Scan(&quizID)
	}
	if err != nil {
		return err
	}
	return runSteps(ctx, tx, []step{
		{`DELETE FROM quiz_questions l USING quizzes lz WHERE lz.id = l.quiz_id AND lz.lesson_id = $2
		   AND NOT EXISTS (SELECT 1 FROM quiz_questions d JOIN quizzes dz ON dz.id = d.quiz_id WHERE dz.lesson_id = $1 AND d.origin_id = l.id)`,
			[]any{draftLesson, liveLesson}},
		{`UPDATE quiz_questions l SET quiz_id = $3, question = d.question, options = d.options, option_explanations = d.option_explanations,
		   correct_index = d.correct_index, explanation = d.explanation, order_num = d.order_num
		 FROM quiz_questions d JOIN quizzes dz ON dz.id = d.quiz_id
		 WHERE dz.lesson_id = $1 AND l.id = d.origin_id
		   AND l.quiz_id IN (SELECT id FROM quizzes WHERE lesson_id = $2)`, []any{draftLesson, liveLesson, quizID}},
		{`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num)
		 SELECT $3, d.question, d.options, d.option_explanations, d.correct_index, d.explanation, d.order_num
		 FROM quiz_questions d JOIN quizzes dz ON dz.id = d.quiz_id
		 WHERE dz.lesson_id = $1 AND NOT EXISTS (
		   SELECT 1 FROM quiz_questions l JOIN quizzes lz ON lz.id = l.quiz_id WHERE lz.lesson_id = $2 AND l.id = d.origin_id)`,
			[]any{draftLesson, liveLesson, quizID}},
		{`DELETE FROM quizzes WHERE lesson_id = $1 AND id <> $2`, []any{liveLesson, quizID}},
	})
}

func mergeTasks(ctx context.Context, tx pgx.Tx, draftLesson, liveLesson int) error {
	args := []any{draftLesson, liveLesson}
	return runSteps(ctx, tx, []step{
		{`DELETE FROM tasks l WHERE l.lesson_id = $2
		   AND NOT EXISTS (SELECT 1 FROM tasks d WHERE d.lesson_id = $1 AND d.origin_id = l.id)`, args},
		{`UPDATE tasks l SET title = d.title, description = d.description, hints = d.hints, solution = d.solution, order_num = d.order_num,
		   difficulty = d.difficulty, glossary = d.glossary, test_cases = d.test_cases, starter_code = d.starter_code, kind = d.kind,
		   sandbox_image = d.sandbox_image, setup_script = d.setup_script, check_script = d.check_script, format = d.format
		 FROM tasks d WHERE d.lesson_id = $1 AND l.id = d.origin_id AND l.lesson_id = $2`, args},
		{`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases,
		   starter_code, kind, sandbox_image, setup_script, check_script, format)
		 SELECT $2, d.title, d.description, d.hints, d.solution, d.order_num, d.difficulty, d.glossary, d.test_cases,
		   d.starter_code, d.kind, d.sandbox_image, d.setup_script, d.check_script, d.format
		 FROM tasks d WHERE d.lesson_id = $1
		   AND NOT EXISTS (SELECT 1 FROM tasks l WHERE l.lesson_id = $2 AND l.id = d.origin_id)`, args},
	})
}

// step is one SQL statement with its arguments.
type step struct {
	sql  string
	args []any
}

func runSteps(ctx context.Context, tx pgx.Tx, steps []step) error {
	for _, st := range steps {
		if _, err := tx.Exec(ctx, st.sql, st.args...); err != nil {
			return err
		}
	}
	return nil
}

// ItemChanges counts added, removed and modified quiz questions or tasks.
type ItemChanges struct {
	Added    int
	Removed  int
	Modified int
}

// LessonChange describes how a lesson differs between the draft and the live course.
type LessonChange struct {
	DraftLessonID *int
	LiveLessonID  *int
	Slug          string
	Title         string
	Change        string // added | removed | modified
	Fields        []string
	Questions     ItemChanges
	Tasks         ItemChanges
}

// DraftChanges summarizes a draft against its live course.
type DraftChanges struct {
	CourseFields []string
	Lessons      []LessonChange
}

type contentRow struct {
	id, origin, parent int
	fields             map[string]any
}

// Changes compares the draft of a live course with the course.
func (r *DraftRepo) Changes(ctx context.Context, liveID int) (DraftChanges, error) {
	var out DraftChanges
	var draftID int
	err := r.pool.QueryRow(ctx, `SELECT id FROM modules WHERE draft_of = $1`, liveID).Scan(&draftID)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNoDraft
	}
	if err != nil {
		return out, err
	}
	const moduleSQL = `SELECT id, 0, 0, to_jsonb(m) - 'id' - 'slug' - 'order_num' - 'published' - 'owner_id' - 'draft_of' - 'created_at' - 'source' - 'prerequisites'
		FROM modules m WHERE id = $1`
	const lessonSQL = `SELECT id, COALESCE(origin_id, 0), 0, to_jsonb(l) - 'id' - 'module_id' - 'origin_id' - 'created_at' - 'source'
		FROM lessons l WHERE module_id = $1`
	const questionSQL = `SELECT q.id, COALESCE(q.origin_id, 0), z.lesson_id, to_jsonb(q) - 'id' - 'quiz_id' - 'origin_id'
		FROM quiz_questions q JOIN quizzes z ON z.id = q.quiz_id JOIN lessons l ON l.id = z.lesson_id WHERE l.module_id = $1`
	const taskSQL = `SELECT t.id, COALESCE(t.origin_id, 0), t.lesson_id, to_jsonb(t) - 'id' - 'lesson_id' - 'origin_id'
		FROM tasks t JOIN lessons l ON l.id = t.lesson_id WHERE l.module_id = $1`
	load := func(sql string, id int) ([]contentRow, error) { return r.contentRows(ctx, sql, id) }
	liveMod, err := load(moduleSQL, liveID)
	if err != nil {
		return out, err
	}
	draftMod, err := load(moduleSQL, draftID)
	if err != nil || len(liveMod) == 0 || len(draftMod) == 0 {
		return out, err
	}
	out.CourseFields = changedFields(liveMod[0].fields, draftMod[0].fields)

	var sets [6][]contentRow
	for i, q := range []struct {
		sql string
		id  int
	}{{lessonSQL, liveID}, {lessonSQL, draftID}, {questionSQL, liveID}, {questionSQL, draftID}, {taskSQL, liveID}, {taskSQL, draftID}} {
		if sets[i], err = load(q.sql, q.id); err != nil {
			return out, err
		}
	}
	liveLessons, draftLessons := sets[0], sets[1]
	liveByID := map[int]contentRow{}
	for _, l := range liveLessons {
		liveByID[l.id] = l
	}
	matched := map[int]bool{}
	sort.Slice(draftLessons, func(i, j int) bool {
		return num(draftLessons[i].fields["order_num"]) < num(draftLessons[j].fields["order_num"])
	})
	for _, d := range draftLessons {
		draftID := d.id
		ch := LessonChange{DraftLessonID: &draftID, Slug: str(d.fields["slug"]), Title: str(d.fields["title"])}
		live, ok := liveByID[d.origin]
		liveLesson := 0
		if ok {
			matched[live.id] = true
			liveLesson = live.id
			liveID := live.id
			ch.LiveLessonID = &liveID
			ch.Fields = changedFields(live.fields, d.fields)
		}
		ch.Questions = compareItems(sets[2], sets[3], liveLesson, d.id)
		ch.Tasks = compareItems(sets[4], sets[5], liveLesson, d.id)
		switch {
		case !ok:
			ch.Change = "added"
		case len(ch.Fields) > 0 || ch.Questions != (ItemChanges{}) || ch.Tasks != (ItemChanges{}):
			ch.Change = "modified"
		default:
			continue
		}
		out.Lessons = append(out.Lessons, ch)
	}
	for _, l := range liveLessons {
		if !matched[l.id] {
			liveID := l.id
			out.Lessons = append(out.Lessons, LessonChange{LiveLessonID: &liveID, Slug: str(l.fields["slug"]), Title: str(l.fields["title"]), Change: "removed",
				Questions: compareItems(sets[2], nil, l.id, 0), Tasks: compareItems(sets[4], nil, l.id, 0)})
		}
	}
	return out, nil
}

func (r *DraftRepo) contentRows(ctx context.Context, sql string, id int) ([]contentRow, error) {
	rows, err := r.pool.Query(ctx, sql, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []contentRow
	for rows.Next() {
		var c contentRow
		var raw []byte
		if err := rows.Scan(&c.id, &c.origin, &c.parent, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &c.fields); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// compareItems counts differences between the items of a live lesson and a draft lesson (0 means none).
func compareItems(live, draft []contentRow, liveLesson, draftLesson int) ItemChanges {
	var c ItemChanges
	liveItems := map[int]contentRow{}
	for _, it := range live {
		if liveLesson != 0 && it.parent == liveLesson {
			liveItems[it.id] = it
		}
	}
	seen := map[int]bool{}
	for _, it := range draft {
		if draftLesson == 0 || it.parent != draftLesson {
			continue
		}
		orig, ok := liveItems[it.origin]
		switch {
		case !ok:
			c.Added++
		case len(changedFields(orig.fields, it.fields)) > 0:
			c.Modified++
		}
		if ok {
			seen[orig.id] = true
		}
	}
	for id := range liveItems {
		if !seen[id] {
			c.Removed++
		}
	}
	return c
}

func changedFields(a, b map[string]any) []string {
	var out []string
	for k, v := range b {
		if !reflect.DeepEqual(a[k], v) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func num(v any) float64 {
	f, _ := v.(float64)
	return f
}
