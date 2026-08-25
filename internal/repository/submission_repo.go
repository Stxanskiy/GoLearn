package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepo struct {
	pool *pgxpool.Pool
}

func NewSubmissionRepo(pool *pgxpool.Pool) *SubmissionRepo {
	return &SubmissionRepo{pool: pool}
}

// Save stores a code submission attempt for a user.
func (r *SubmissionRepo) Save(ctx context.Context, userID, taskID int, code, output, errors string, passed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO submissions (user_id, task_id, code, output, errors, passed)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, taskID, code, output, errors, passed)
	return err
}

// IsTaskPassed reports whether the user has a passing submission for the task.
func (r *SubmissionRepo) IsTaskPassed(ctx context.Context, userID, taskID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM submissions WHERE user_id = $1 AND task_id = $2 AND passed = true)`,
		userID, taskID).Scan(&exists)
	return exists, err
}

// LessonLabStatus returns, per lesson with tasks, whether the user has passed all of them.
func (r *SubmissionRepo) LessonLabStatus(ctx context.Context, userID int) (map[int]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.lesson_id,
		       COUNT(*) total,
		       COUNT(DISTINCT CASE WHEN s.passed THEN t.id END) passed
		FROM tasks t LEFT JOIN submissions s ON s.task_id = t.id AND s.user_id = $1
		GROUP BY t.lesson_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	status := make(map[int]bool)
	for rows.Next() {
		var lessonID, total, passed int
		if err := rows.Scan(&lessonID, &total, &passed); err != nil {
			return nil, err
		}
		status[lessonID] = total > 0 && passed >= total
	}
	return status, rows.Err()
}

// AllLessonTasksPassed reports whether the user has passed every checkable task
// in a lesson (go tasks with test cases OR shell tasks with a check script).
func (r *SubmissionRepo) AllLessonTasksPassed(ctx context.Context, userID, lessonID int) (bool, error) {
	var totalTasks, passedTasks int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tasks
		WHERE lesson_id = $1 AND (test_cases <> '[]' OR check_script <> '')`,
		lessonID).Scan(&totalTasks)
	if err != nil {
		return false, err
	}
	if totalTasks == 0 {
		return false, nil
	}
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT t.id)
		FROM tasks t
		JOIN submissions s ON s.task_id = t.id AND s.passed = true AND s.user_id = $1
		WHERE t.lesson_id = $2 AND (t.test_cases <> '[]' OR t.check_script <> '')`,
		userID, lessonID).Scan(&passedTasks)
	if err != nil {
		return false, err
	}
	return passedTasks >= totalTasks, nil
}

// PassedTaskIDs returns the set of tasks in a lesson the user has already
// solved, so a returning student sees their completed steps instead of a
// blank lab.
func (r *SubmissionRepo) PassedTaskIDs(ctx context.Context, userID, lessonID int) (map[int]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT s.task_id
		FROM submissions s JOIN tasks t ON t.id = s.task_id
		WHERE s.user_id = $1 AND t.lesson_id = $2 AND s.passed = true`, userID, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	passed := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		passed[id] = true
	}
	return passed, rows.Err()
}

// ResetLesson drops the user's attempt history for every task in a lesson, so
// the lab can be taken again from scratch ("Пройти заново").
func (r *SubmissionRepo) ResetLesson(ctx context.Context, userID, lessonID int) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM submissions
		WHERE user_id = $1 AND task_id IN (SELECT id FROM tasks WHERE lesson_id = $2)`,
		userID, lessonID)
	return err
}
