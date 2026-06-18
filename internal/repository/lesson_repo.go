package repository

import (
	"context"
	"encoding/json"

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

const lessonCols = `id, module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source, created_at`

func scanLesson(row pgx.Row) (model.Lesson, error) {
	var l model.Lesson
	err := row.Scan(&l.ID, &l.ModuleID, &l.Slug, &l.Title, &l.Content, &l.OrderNum, &l.Difficulty,
		&l.Track, &l.Kind, &l.Format, &l.VMImage, &l.VMInit, &l.Source, &l.CreatedAt)
	return l, err
}

func (r *LessonRepo) GetByModule(ctx context.Context, moduleID int) ([]model.Lesson, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+lessonCols+` FROM lessons WHERE module_id = $1 ORDER BY order_num`, moduleID)
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
		`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, format, vm_image, vm_init, source)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'admin') RETURNING id`,
		l.ModuleID, l.Slug, l.Title, l.Content, l.OrderNum, l.Difficulty, l.Track, l.Kind, l.Format, l.VMImage, l.VMInit).Scan(&id)
	return id, err
}

func (r *LessonRepo) Update(ctx context.Context, l model.Lesson) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lessons SET slug=$1, title=$2, content=$3, order_num=$4, difficulty=$5, kind=$6, format=$7, vm_image=$8, vm_init=$9 WHERE id=$10`,
		l.Slug, l.Title, l.Content, l.OrderNum, l.Difficulty, l.Kind, l.Format, l.VMImage, l.VMInit, l.ID)
	return err
}

func (r *LessonRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM lessons WHERE id = $1`, id)
	return err
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
