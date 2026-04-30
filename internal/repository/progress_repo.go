package repository

import (
	"context"
	"time"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgressRepo struct {
	pool *pgxpool.Pool
}

func NewProgressRepo(pool *pgxpool.Pool) *ProgressRepo {
	return &ProgressRepo{pool: pool}
}

func (r *ProgressRepo) Get(ctx context.Context, lessonID int) (*model.Progress, error) {
	var p model.Progress
	err := r.pool.QueryRow(ctx, `
		SELECT id, lesson_id, status, quiz_score, quiz_total, notes, completed_at, updated_at
		FROM progress WHERE lesson_id = $1`, lessonID).
		Scan(&p.ID, &p.LessonID, &p.Status, &p.QuizScore, &p.QuizTotal, &p.Notes, &p.CompletedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProgressRepo) GetAll(ctx context.Context) ([]model.Progress, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, status, quiz_score, quiz_total, notes, completed_at, updated_at
		FROM progress ORDER BY lesson_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []model.Progress
	for rows.Next() {
		var p model.Progress
		if err := rows.Scan(&p.ID, &p.LessonID, &p.Status, &p.QuizScore, &p.QuizTotal, &p.Notes, &p.CompletedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		progress = append(progress, p)
	}
	return progress, rows.Err()
}

func (r *ProgressRepo) Upsert(ctx context.Context, lessonID int, status string) error {
	now := time.Now()
	var completedAt *time.Time
	if status == "completed" {
		completedAt = &now
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO progress (lesson_id, status, completed_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (lesson_id) DO UPDATE
		SET status = $2, completed_at = COALESCE($3, progress.completed_at), updated_at = $4`,
		lessonID, status, completedAt, now)
	return err
}

func (r *ProgressRepo) SaveQuizResult(ctx context.Context, lessonID, score, total int) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO progress (lesson_id, status, quiz_score, quiz_total, updated_at)
		VALUES ($1, 'in_progress', $2, $3, $4)
		ON CONFLICT (lesson_id) DO UPDATE
		SET quiz_score = $2, quiz_total = $3, updated_at = $4`,
		lessonID, score, total, now)
	return err
}

func (r *ProgressRepo) SaveNotes(ctx context.Context, lessonID int, notes string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO progress (lesson_id, status, notes, updated_at)
		VALUES ($1, 'in_progress', $2, $3)
		ON CONFLICT (lesson_id) DO UPDATE
		SET notes = $2, updated_at = $3`,
		lessonID, notes, now)
	return err
}

func (r *ProgressRepo) GetStats(ctx context.Context) (*model.DashboardStats, error) {
	var stats model.DashboardStats

	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM lessons`).Scan(&stats.TotalLessons)
	if err != nil {
		return nil, err
	}

	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM progress WHERE status = 'completed'`).Scan(&stats.CompletedCount)
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM progress WHERE status = 'in_progress'`).Scan(&stats.InProgressCount)
	_ = r.pool.QueryRow(ctx, `SELECT COALESCE(AVG(quiz_score * 100.0 / NULLIF(quiz_total, 0)), 0) FROM progress WHERE quiz_score IS NOT NULL`).Scan(&stats.AvgQuizScore)

	return &stats, nil
}
