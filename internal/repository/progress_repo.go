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

// Overview gathers the rich "Мой прогресс" metrics plus a day-by-day activity map.
func (r *ProgressRepo) Overview(ctx context.Context) (*model.ProgressOverview, error) {
	o := &model.ProgressOverview{
		Activity:      make(map[string]int),
		SimulatorsTot: 4, // placeholder until Phase D
		TrainersTot:   3, // placeholder until Phase C
	}

	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM lessons`).Scan(&o.ArticlesTotal)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM progress WHERE status IN ('in_progress','completed')`).Scan(&o.ArticlesRead)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM progress WHERE quiz_score IS NOT NULL`).Scan(&o.TestsPassed)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT task_id) FROM submissions WHERE passed = true`).Scan(&o.TasksSolved)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT lesson_id) FROM tasks`).Scan(&o.LabsTotal)
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT t.lesson_id,
			       COUNT(*) total,
			       COUNT(DISTINCT CASE WHEN s.passed THEN t.id END) passed
			FROM tasks t LEFT JOIN submissions s ON s.task_id = t.id
			GROUP BY t.lesson_id
		) x WHERE total > 0 AND passed >= total`).Scan(&o.LabsDone)

	// Activity by day (union of lesson progress + code submissions).
	rows, err := r.pool.Query(ctx, `
		SELECT to_char(t::date,'YYYY-MM-DD') d, COUNT(*) c FROM (
			SELECT updated_at t FROM progress
			UNION ALL
			SELECT created_at t FROM submissions
		) s GROUP BY 1`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day string
			var c int
			if err := rows.Scan(&day, &c); err == nil {
				o.Activity[day] = c
			}
		}
	}

	o.ActiveDays = len(o.Activity)
	o.Streak = computeStreak(o.Activity)
	return o, nil
}

// computeStreak counts consecutive days with activity ending today or yesterday.
func computeStreak(activity map[string]int) int {
	if len(activity) == 0 {
		return 0
	}
	day := time.Now()
	if _, ok := activity[day.Format("2006-01-02")]; !ok {
		day = day.AddDate(0, 0, -1) // allow streak that ended yesterday
	}
	streak := 0
	for {
		if _, ok := activity[day.Format("2006-01-02")]; !ok {
			break
		}
		streak++
		day = day.AddDate(0, 0, -1)
	}
	return streak
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
