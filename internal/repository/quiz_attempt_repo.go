package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// QuizAttemptRepo keeps a history of quiz submissions.
type QuizAttemptRepo struct {
	pool *pgxpool.Pool
}

func NewQuizAttemptRepo(pool *pgxpool.Pool) *QuizAttemptRepo {
	return &QuizAttemptRepo{pool: pool}
}

// AttemptAnswer is one question's outcome in an attempt; Selected is nil when unanswered.
type AttemptAnswer struct {
	QuestionID int  `json:"question_id"`
	Selected   *int `json:"selected_index"`
	Correct    bool `json:"correct"`
}

// Save records an attempt and returns its 1-based number for this user and lesson.
func (r *QuizAttemptRepo) Save(ctx context.Context, userID, lessonID, score, total int, answers []AttemptAnswer) (int, error) {
	if answers == nil {
		answers = []AttemptAnswer{}
	}
	raw, err := json.Marshal(answers)
	if err != nil {
		return 0, err
	}
	var n int
	err = r.pool.QueryRow(ctx, `
		WITH ins AS (
			INSERT INTO quiz_attempts (user_id, lesson_id, score, total, answers) VALUES ($1, $2, $3, $4, $5)
		)
		SELECT COUNT(*) + 1 FROM quiz_attempts WHERE user_id = $1 AND lesson_id = $2`,
		userID, lessonID, score, total, raw).Scan(&n)
	return n, err
}

// Count returns how many attempts the user made on a lesson's quiz.
func (r *QuizAttemptRepo) Count(ctx context.Context, userID, lessonID int) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM quiz_attempts WHERE user_id = $1 AND lesson_id = $2`, userID, lessonID).Scan(&n)
	return n, err
}
