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

func (r *ProgressRepo) Get(ctx context.Context, userID, lessonID int) (*model.Progress, error) {
	var p model.Progress
	err := r.pool.QueryRow(ctx, `
		SELECT id, lesson_id, status, quiz_score, quiz_total, notes, completed_at, updated_at
		FROM progress WHERE user_id = $1 AND lesson_id = $2`, userID, lessonID).
		Scan(&p.ID, &p.LessonID, &p.Status, &p.QuizScore, &p.QuizTotal, &p.Notes, &p.CompletedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProgressRepo) GetAll(ctx context.Context, userID int) ([]model.Progress, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, lesson_id, status, quiz_score, quiz_total, notes, completed_at, updated_at
		FROM progress WHERE user_id = $1 ORDER BY lesson_id`, userID)
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

func (r *ProgressRepo) Upsert(ctx context.Context, userID, lessonID int, status string) error {
	now := time.Now()
	var completedAt *time.Time
	if status == "completed" {
		completedAt = &now
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO progress (user_id, lesson_id, status, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, lesson_id) DO UPDATE
		SET status = $3, completed_at = COALESCE($4, progress.completed_at), updated_at = $5`,
		userID, lessonID, status, completedAt, now)
	return err
}

func (r *ProgressRepo) SaveQuizResult(ctx context.Context, userID, lessonID, score, total int) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO progress (user_id, lesson_id, status, quiz_score, quiz_total, updated_at)
		VALUES ($1, $2, 'in_progress', $3, $4, $5)
		ON CONFLICT (user_id, lesson_id) DO UPDATE
		SET quiz_score = $3, quiz_total = $4, updated_at = $5`,
		userID, lessonID, score, total, now)
	return err
}

func (r *ProgressRepo) SaveNotes(ctx context.Context, userID, lessonID int, notes string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO progress (user_id, lesson_id, status, notes, updated_at)
		VALUES ($1, $2, 'in_progress', $3, $4)
		ON CONFLICT (user_id, lesson_id) DO UPDATE
		SET notes = $3, updated_at = $4`,
		userID, lessonID, notes, now)
	return err
}

// Overview gathers the rich "Мой прогресс" metrics plus a day-by-day activity map.
func (r *ProgressRepo) Overview(ctx context.Context, userID int) (*model.ProgressOverview, error) {
	o := &model.ProgressOverview{
		Activity:      make(map[string]int),
		SimulatorsTot: 4, // overridden by handler
		TrainersTot:   3, // overridden by handler
	}

	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM lessons`).Scan(&o.ArticlesTotal)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM progress WHERE user_id=$1 AND status IN ('in_progress','completed')`, userID).Scan(&o.ArticlesRead)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM progress WHERE user_id=$1 AND quiz_score IS NOT NULL`, userID).Scan(&o.TestsPassed)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT task_id) FROM submissions WHERE user_id=$1 AND passed = true`, userID).Scan(&o.TasksSolved)
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT lesson_id) FROM tasks`).Scan(&o.LabsTotal)
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM (
			SELECT t.lesson_id,
			       COUNT(*) total,
			       COUNT(DISTINCT CASE WHEN s.passed THEN t.id END) passed
			FROM tasks t LEFT JOIN submissions s ON s.task_id = t.id AND s.user_id = $1
			GROUP BY t.lesson_id
		) x WHERE total > 0 AND passed >= total`, userID).Scan(&o.LabsDone)

	// Activity by day (this user's lesson progress + code submissions).
	rows, err := r.pool.Query(ctx, `
		SELECT to_char(t::date,'YYYY-MM-DD') d, COUNT(*) c FROM (
			SELECT updated_at t FROM progress WHERE user_id = $1
			UNION ALL
			SELECT created_at t FROM submissions WHERE user_id = $1
		) s GROUP BY 1`, userID)
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

// ResetLesson clears the user's status and quiz score for one lesson while
// keeping their notes — the lesson can be retaken from a clean slate.
func (r *ProgressRepo) ResetLesson(ctx context.Context, userID, lessonID int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE progress
		SET status = 'not_started', quiz_score = NULL, quiz_total = NULL,
		    completed_at = NULL, updated_at = NOW()
		WHERE user_id = $1 AND lesson_id = $2`, userID, lessonID)
	return err
}
