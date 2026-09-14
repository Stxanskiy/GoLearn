package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QuizAnswerRepo stores users' answers to quiz questions.
type QuizAnswerRepo struct {
	pool *pgxpool.Pool
}

func NewQuizAnswerRepo(pool *pgxpool.Pool) *QuizAnswerRepo {
	return &QuizAnswerRepo{pool: pool}
}

// Record stores the first answer to a question and returns the stored selection; created is false if one existed.
func (r *QuizAnswerRepo) Record(ctx context.Context, userID, questionID, selected int) (stored int, created bool, err error) {
	err = r.pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO quiz_answers (user_id, question_id, selected_index) VALUES ($1, $2, $3)
			ON CONFLICT (user_id, question_id) DO NOTHING
			RETURNING selected_index
		)
		SELECT selected_index, TRUE FROM ins
		UNION ALL
		SELECT selected_index, FALSE FROM quiz_answers
		WHERE user_id = $1 AND question_id = $2 AND NOT EXISTS (SELECT 1 FROM ins)`,
		userID, questionID, selected).Scan(&stored, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		// a concurrent insert is not visible in this statement's snapshot; read it back
		err = r.pool.QueryRow(ctx, `SELECT selected_index FROM quiz_answers WHERE user_id = $1 AND question_id = $2`,
			userID, questionID).Scan(&stored)
	}
	return stored, created, err
}

// ForLesson returns question id → selected index for the user's answers in a lesson's quiz.
func (r *QuizAnswerRepo) ForLesson(ctx context.Context, userID, lessonID int) (map[int]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.question_id, a.selected_index
		FROM quiz_answers a
		JOIN quiz_questions q ON q.id = a.question_id JOIN quizzes z ON z.id = q.quiz_id
		WHERE a.user_id = $1 AND z.lesson_id = $2`, userID, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int]int)
	for rows.Next() {
		var qid, sel int
		if err := rows.Scan(&qid, &sel); err != nil {
			return nil, err
		}
		out[qid] = sel
	}
	return out, rows.Err()
}

// ResetLesson deletes the user's answers in a lesson's quiz.
func (r *QuizAnswerRepo) ResetLesson(ctx context.Context, userID, lessonID int) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM quiz_answers
		WHERE user_id = $1 AND question_id IN (
			SELECT q.id FROM quiz_questions q JOIN quizzes z ON z.id = q.quiz_id WHERE z.lesson_id = $2)`,
		userID, lessonID)
	return err
}
