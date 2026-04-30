package repository

import (
	"context"
	"encoding/json"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LessonRepo struct {
	pool *pgxpool.Pool
}

func NewLessonRepo(pool *pgxpool.Pool) *LessonRepo {
	return &LessonRepo{pool: pool}
}

func (r *LessonRepo) GetByModule(ctx context.Context, moduleID int) ([]model.Lesson, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, module_id, slug, title, content, order_num, created_at
		FROM lessons WHERE module_id = $1 ORDER BY order_num`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []model.Lesson
	for rows.Next() {
		var l model.Lesson
		if err := rows.Scan(&l.ID, &l.ModuleID, &l.Slug, &l.Title, &l.Content, &l.OrderNum, &l.CreatedAt); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	return lessons, rows.Err()
}

func (r *LessonRepo) GetBySlug(ctx context.Context, moduleID int, slug string) (*model.Lesson, error) {
	var l model.Lesson
	err := r.pool.QueryRow(ctx, `
		SELECT id, module_id, slug, title, content, order_num, created_at
		FROM lessons WHERE module_id = $1 AND slug = $2`, moduleID, slug).
		Scan(&l.ID, &l.ModuleID, &l.Slug, &l.Title, &l.Content, &l.OrderNum, &l.CreatedAt)
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
		SELECT id, quiz_id, question, options, correct_index, explanation, order_num
		FROM quiz_questions WHERE quiz_id = $1 ORDER BY order_num`, quiz.ID)
	if err != nil {
		return &quiz, nil, err
	}
	defer rows.Close()

	var questions []model.QuizQuestion
	for rows.Next() {
		var q model.QuizQuestion
		var optionsJSON []byte
		if err := rows.Scan(&q.ID, &q.QuizID, &q.Question, &optionsJSON, &q.CorrectIndex, &q.Explanation, &q.OrderNum); err != nil {
			return &quiz, nil, err
		}
		_ = json.Unmarshal(optionsJSON, &q.Options)
		questions = append(questions, q)
	}
	return &quiz, questions, rows.Err()
}

func (r *LessonRepo) GetTasks(ctx context.Context, lessonID int) ([]model.Task, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, title, description, hints, solution, order_num
		FROM tasks WHERE lesson_id = $1 ORDER BY order_num`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.LessonID, &t.Title, &t.Description, &t.Hints, &t.Solution, &t.OrderNum); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
