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

// Save stores a code submission attempt.
func (r *SubmissionRepo) Save(ctx context.Context, taskID int, code, output, errors string, passed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO submissions (task_id, code, output, errors, passed)
		VALUES ($1, $2, $3, $4, $5)`,
		taskID, code, output, errors, passed)
	return err
}

// IsTaskPassed returns true if there is at least one passed submission for the task.
func (r *SubmissionRepo) IsTaskPassed(ctx context.Context, taskID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM submissions WHERE task_id = $1 AND passed = true)`,
		taskID).Scan(&exists)
	return exists, err
}

// LessonLabStatus returns, for every lesson that has tasks, whether all of its
// tasks have at least one passed submission (i.e. the lab is complete).
func (r *SubmissionRepo) LessonLabStatus(ctx context.Context) (map[int]bool, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.lesson_id,
		       COUNT(*) total,
		       COUNT(DISTINCT CASE WHEN s.passed THEN t.id END) passed
		FROM tasks t LEFT JOIN submissions s ON s.task_id = t.id
		GROUP BY t.lesson_id`)
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

// AllLessonTasksPassed checks if all tasks in a lesson have at least one passed submission.
func (r *SubmissionRepo) AllLessonTasksPassed(ctx context.Context, lessonID int) (bool, error) {
	var totalTasks, passedTasks int

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tasks WHERE lesson_id = $1 AND test_cases != '[]'`,
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
		JOIN submissions s ON s.task_id = t.id AND s.passed = true
		WHERE t.lesson_id = $1 AND t.test_cases != '[]'`,
		lessonID).Scan(&passedTasks)
	if err != nil {
		return false, err
	}

	return passedTasks >= totalTasks, nil
}
