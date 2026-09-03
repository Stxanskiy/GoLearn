package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LessonRepo struct {
	pool *pgxpool.Pool
}

func NewLessonRepo(pool *pgxpool.Pool) *LessonRepo {
	return &LessonRepo{pool: pool}
}

const lessonCols = `id, module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published, created_at`

func scanLesson(row pgx.Row) (model.Lesson, error) {
	var l model.Lesson
	err := row.Scan(&l.ID, &l.ModuleID, &l.Slug, &l.Title, &l.Content, &l.OrderNum, &l.Difficulty,
		&l.Track, &l.Kind, &l.Format, &l.VMImage, &l.VMInit, &l.Source, &l.Published, &l.CreatedAt)
	return l, err
}

// SetPublished toggles a lesson's draft/published state.
func (r *LessonRepo) SetPublished(ctx context.Context, id int, published bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE lessons SET published = $1 WHERE id = $2`, published, id)
	return err
}

// GetByModule returns published lessons — the default, student-safe listing.
// Admin screens use GetByModuleAll to also see drafts.
func (r *LessonRepo) GetByModule(ctx context.Context, moduleID int) ([]model.Lesson, error) {
	return r.lessonsByModule(ctx, moduleID, true)
}

// GetByModuleAll returns every lesson of a module including drafts (admin).
func (r *LessonRepo) GetByModuleAll(ctx context.Context, moduleID int) ([]model.Lesson, error) {
	return r.lessonsByModule(ctx, moduleID, false)
}

func (r *LessonRepo) lessonsByModule(ctx context.Context, moduleID int, publishedOnly bool) ([]model.Lesson, error) {
	sql := `SELECT ` + lessonCols + ` FROM lessons WHERE module_id = $1 ORDER BY order_num`
	if publishedOnly {
		sql = `SELECT ` + lessonCols + ` FROM lessons WHERE module_id = $1 AND published ORDER BY order_num`
	}
	rows, err := r.pool.Query(ctx, sql, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lessons []model.Lesson
	for rows.Next() {
		l, err := scanLesson(rows)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (r *LessonRepo) GetBySlug(ctx context.Context, moduleID int, slug string) (*model.Lesson, error) {
	l, err := scanLesson(r.pool.QueryRow(ctx,
		`SELECT `+lessonCols+` FROM lessons WHERE module_id = $1 AND slug = $2`, moduleID, slug))
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// Create inserts an admin-authored lesson and returns its id.
func (r *LessonRepo) Create(ctx context.Context, l model.Lesson) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'admin',$12) RETURNING id`,
		l.ModuleID, l.Slug, l.Title, l.Content, l.OrderNum, l.Difficulty, l.Track, l.Kind, l.Format, l.VMImage, l.VMInit, l.Published).Scan(&id)
	return id, err
}

func (r *LessonRepo) Update(ctx context.Context, l model.Lesson) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lessons SET slug=$1, title=$2, content=$3, order_num=$4, difficulty=$5, kind=$6, format=$7, vm_image=$8, vm_init=$9, published=$10 WHERE id=$11`,
		l.Slug, l.Title, l.Content, l.OrderNum, l.Difficulty, l.Kind, l.Format, l.VMImage, l.VMInit, l.Published, l.ID)
	return err
}

func (r *LessonRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM lessons WHERE id = $1`, id)
	return err
}

// MoveLesson swaps a lesson's order_num with its neighbour (dir "up"/"down")
// inside the same module. At a boundary it is a no-op.
func (r *LessonRepo) MoveLesson(ctx context.Context, id int, dir string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var moduleID, ord int
	if err := tx.QueryRow(ctx, `SELECT module_id, order_num FROM lessons WHERE id=$1`, id).Scan(&moduleID, &ord); err != nil {
		return err
	}
	q := `SELECT id, order_num FROM lessons WHERE module_id=$1 AND order_num > $2 ORDER BY order_num ASC LIMIT 1`
	if dir == "up" {
		q = `SELECT id, order_num FROM lessons WHERE module_id=$1 AND order_num < $2 ORDER BY order_num DESC LIMIT 1`
	}
	var nid, nord int
	if err := tx.QueryRow(ctx, q, moduleID, ord).Scan(&nid, &nord); err != nil {
		if err == pgx.ErrNoRows {
			return tx.Commit(ctx) // already at the edge
		}
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE lessons SET order_num=$1 WHERE id=$2`, nord, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE lessons SET order_num=$1 WHERE id=$2`, ord, nid); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// DuplicateLesson clones a lesson (with its quiz, questions and tasks) to the end
// of the module, under a unique "<slug>-copy" slug. Returns the new lesson id.
func (r *LessonRepo) DuplicateLesson(ctx context.Context, id int) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	l, err := scanLesson(tx.QueryRow(ctx, `SELECT `+lessonCols+` FROM lessons WHERE id=$1`, id))
	if err != nil {
		return 0, err
	}
	base := l.Slug + "-copy"
	slug := base
	for i := 2; ; i++ {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM lessons WHERE module_id=$1 AND slug=$2`, l.ModuleID, slug).Scan(&n); err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		slug = base + "-" + strconv.Itoa(i)
	}
	var maxOrd int
	_ = tx.QueryRow(ctx, `SELECT COALESCE(MAX(order_num),0) FROM lessons WHERE module_id=$1`, l.ModuleID).Scan(&maxOrd)

	var newID int
	if err := tx.QueryRow(ctx,
		`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, published)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'admin',FALSE) RETURNING id`,
		l.ModuleID, slug, l.Title+" (копия)", l.Content, maxOrd+1, l.Difficulty, l.Track, l.Kind, l.Format, l.VMImage, l.VMInit).Scan(&newID); err != nil {
		return 0, err
	}

	// copy quiz + its questions, if any
	var quizID int
	if err := tx.QueryRow(ctx, `SELECT id FROM quizzes WHERE lesson_id=$1`, id).Scan(&quizID); err == nil {
		var newQuiz int
		if err := tx.QueryRow(ctx,
			`INSERT INTO quizzes (lesson_id, title) SELECT $1, title FROM quizzes WHERE id=$2 RETURNING id`, newID, quizID).Scan(&newQuiz); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num)
			 SELECT $1, question, options, option_explanations, correct_index, explanation, order_num FROM quiz_questions WHERE quiz_id=$2`,
			newQuiz, quizID); err != nil {
			return 0, err
		}
	}

	// copy tasks
	if _, err := tx.Exec(ctx,
		`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script)
		 SELECT $1, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script FROM tasks WHERE lesson_id=$2`,
		newID, id); err != nil {
		return 0, err
	}
	return newID, tx.Commit(ctx)
}

// CountsForLesson returns the number of quiz questions and tasks for a lesson.
func (r *LessonRepo) CountsForLesson(ctx context.Context, lessonID int) (questions, tasks int) {
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE((SELECT count(*) FROM quiz_questions qq JOIN quizzes z ON qq.quiz_id=z.id WHERE z.lesson_id=$1),0),
		COALESCE((SELECT count(*) FROM tasks WHERE lesson_id=$1),0)`, lessonID).Scan(&questions, &tasks)
	return
}

func (r *LessonRepo) GetByID(ctx context.Context, id int) (*model.Lesson, error) {
	l, err := scanLesson(r.pool.QueryRow(ctx, `SELECT `+lessonCols+` FROM lessons WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LessonRepo) GetQuiz(ctx context.Context, lessonID int) (*model.Quiz, []model.QuizQuestion, error) {
	var quiz model.Quiz
	err := r.pool.QueryRow(ctx, `
		SELECT id, lesson_id, title FROM quizzes WHERE lesson_id = $1`, lessonID).
		Scan(&quiz.ID, &quiz.LessonID, &quiz.Title)
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, quiz_id, question, options, option_explanations, correct_index, explanation, order_num
		FROM quiz_questions WHERE quiz_id = $1 ORDER BY order_num`, quiz.ID)
	if err != nil {
		return &quiz, nil, err
	}
	defer rows.Close()

	var questions []model.QuizQuestion
	for rows.Next() {
		var q model.QuizQuestion
		var optionsJSON, optExplJSON []byte
		if err := rows.Scan(&q.ID, &q.QuizID, &q.Question, &optionsJSON, &optExplJSON, &q.CorrectIndex, &q.Explanation, &q.OrderNum); err != nil {
			return &quiz, nil, err
		}
		_ = json.Unmarshal(optionsJSON, &q.Options)
		_ = json.Unmarshal(optExplJSON, &q.OptionExpl)
		questions = append(questions, q)
	}
	return &quiz, questions, rows.Err()
}

func (r *LessonRepo) GetTasks(ctx context.Context, lessonID int) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script
		FROM tasks WHERE lesson_id = $1 ORDER BY order_num`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		var glossaryJSON, testCasesJSON []byte
		if err := rows.Scan(&t.ID, &t.LessonID, &t.Title, &t.Description, &t.Hints, &t.Solution, &t.OrderNum, &t.Difficulty, &glossaryJSON, &testCasesJSON, &t.StarterCode, &t.Format, &t.Kind, &t.SandboxImage, &t.SetupScript, &t.CheckScript); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(glossaryJSON, &t.Glossary)
		_ = json.Unmarshal(testCasesJSON, &t.TestCases)
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// LessonSandbox returns the image and setup script for a lesson's shell lab.
// The whole lab shares one container, so the setup is the concatenation of
// every distinct task setup script (in task order) and the image is the first
// non-empty one — all fixture files exist before the student's first command.
func (r *LessonRepo) LessonSandbox(ctx context.Context, lessonID int) (image string, setup string, err error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sandbox_image, setup_script FROM tasks
		WHERE lesson_id = $1 AND kind = 'shell' ORDER BY order_num, id`, lessonID)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	seen := make(map[string]bool)
	var parts []string
	for rows.Next() {
		var img, s string
		if err := rows.Scan(&img, &s); err != nil {
			return "", "", err
		}
		if image == "" {
			image = img
		}
		if s = strings.TrimSpace(s); s != "" && !seen[s] {
			seen[s] = true
			parts = append(parts, s)
		}
	}
	return image, strings.Join(parts, "\n"), rows.Err()
}

// ── Admin: quiz question CRUD ──

// EnsureQuiz returns the quiz id for a lesson, creating the quiz row if needed.
func (r *LessonRepo) EnsureQuiz(ctx context.Context, lessonID int, title string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, `SELECT id FROM quizzes WHERE lesson_id = $1`, lessonID).Scan(&id)
	if err == nil {
		return id, nil
	}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO quizzes (lesson_id, title) VALUES ($1, $2) RETURNING id`, lessonID, title).Scan(&id)
	return id, err
}

func (r *LessonRepo) AddQuestion(ctx context.Context, quizID int, q model.QuizQuestion) error {
	opts, _ := json.Marshal(q.Options)
	oexpl, _ := json.Marshal(q.OptionExpl)
	var maxOrder int
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(order_num),0) FROM quiz_questions WHERE quiz_id=$1`, quizID).Scan(&maxOrder)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		quizID, q.Question, opts, oexpl, q.CorrectIndex, q.Explanation, maxOrder+1)
	return err
}

func (r *LessonRepo) UpdateQuestion(ctx context.Context, q model.QuizQuestion) error {
	opts, _ := json.Marshal(q.Options)
	oexpl, _ := json.Marshal(q.OptionExpl)
	_, err := r.pool.Exec(ctx,
		`UPDATE quiz_questions SET question=$1, options=$2, option_explanations=$3, correct_index=$4, explanation=$5 WHERE id=$6`,
		q.Question, opts, oexpl, q.CorrectIndex, q.Explanation, q.ID)
	return err
}

func (r *LessonRepo) DeleteQuestion(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM quiz_questions WHERE id = $1`, id)
	return err
}

func (r *LessonRepo) GetQuestionByID(ctx context.Context, id int) (*model.QuizQuestion, error) {
	var q model.QuizQuestion
	var opts, oexpl []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, quiz_id, question, options, option_explanations, correct_index, explanation, order_num FROM quiz_questions WHERE id=$1`, id).
		Scan(&q.ID, &q.QuizID, &q.Question, &opts, &oexpl, &q.CorrectIndex, &q.Explanation, &q.OrderNum)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(opts, &q.Options)
	_ = json.Unmarshal(oexpl, &q.OptionExpl)
	return &q, nil
}

// LessonIDForQuiz returns the lesson a quiz belongs to (for redirects).
func (r *LessonRepo) LessonIDForQuiz(ctx context.Context, quizID int) (int, error) {
	var lid int
	err := r.pool.QueryRow(ctx, `SELECT lesson_id FROM quizzes WHERE id=$1`, quizID).Scan(&lid)
	return lid, err
}

// ── Admin: task CRUD ──

func (r *LessonRepo) CreateTask(ctx context.Context, t model.Task) (int, error) {
	gloss, _ := json.Marshal(t.Glossary)
	tc, _ := json.Marshal(t.TestCases)
	var maxOrder int
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(order_num),0) FROM tasks WHERE lesson_id=$1`, t.LessonID).Scan(&maxOrder)
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING id`,
		t.LessonID, t.Title, t.Description, t.Hints, t.Solution, maxOrder+1, t.Difficulty, gloss, tc, t.StarterCode, t.Format, t.Kind, t.SandboxImage, t.SetupScript, t.CheckScript).Scan(&id)
	return id, err
}

func (r *LessonRepo) UpdateTask(ctx context.Context, t model.Task) error {
	gloss, _ := json.Marshal(t.Glossary)
	tc, _ := json.Marshal(t.TestCases)
	_, err := r.pool.Exec(ctx,
		`UPDATE tasks SET title=$1, description=$2, hints=$3, solution=$4, difficulty=$5, glossary=$6,
		   test_cases=$7, starter_code=$8, format=$9, kind=$10, sandbox_image=$11, setup_script=$12, check_script=$13 WHERE id=$14`,
		t.Title, t.Description, t.Hints, t.Solution, t.Difficulty, gloss, tc, t.StarterCode, t.Format, t.Kind, t.SandboxImage, t.SetupScript, t.CheckScript, t.ID)
	return err
}

func (r *LessonRepo) DeleteTask(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}

func (r *LessonRepo) GetTaskByID(ctx context.Context, taskID int) (*model.Task, error) {
	var t model.Task
	var glossaryJSON, testCasesJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, format, kind, sandbox_image, setup_script, check_script
		FROM tasks WHERE id = $1`, taskID).
		Scan(&t.ID, &t.LessonID, &t.Title, &t.Description, &t.Hints, &t.Solution, &t.OrderNum, &t.Difficulty, &glossaryJSON, &testCasesJSON, &t.StarterCode, &t.Format, &t.Kind, &t.SandboxImage, &t.SetupScript, &t.CheckScript)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(glossaryJSON, &t.Glossary)
	_ = json.Unmarshal(testCasesJSON, &t.TestCases)
	return &t, nil
}
