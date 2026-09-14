package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Review request statuses.
const (
	ReviewPending   = "pending"
	ReviewApproved  = "approved"
	ReviewRejected  = "rejected"
	ReviewCancelled = "cancelled"
)

// ErrReviewClosed means the review request is no longer pending.
var ErrReviewClosed = errors.New("review request is not pending")

// Review is a request to publish a course (LessonID nil) or a lesson, with display fields joined in.
type Review struct {
	ID             int
	ModuleID       int
	CourseSlug     string
	CourseTitle    string
	LessonID       *int
	LessonSlug     string
	LessonTitle    string
	Status         string
	Note           string
	RequestedBy    *int
	RequesterName  string
	RequesterEmail string
	DecisionNote   string
	DecidedBy      *int
	DeciderName    string
	DeciderEmail   string
	CreatedAt      time.Time
	DecidedAt      *time.Time
}

// ReviewFilter selects review requests; zero values match everything.
type ReviewFilter struct {
	Status   string
	ModuleID int
	EditorID int // only courses this user owns or co-authors
}

// ReviewRepo stores moderation requests.
type ReviewRepo struct {
	pool *pgxpool.Pool
}

func NewReviewRepo(pool *pgxpool.Pool) *ReviewRepo {
	return &ReviewRepo{pool: pool}
}

const reviewSelect = `SELECT r.id, r.module_id, m.slug, m.title, r.lesson_id, COALESCE(l.slug, ''), COALESCE(l.title, ''),
	r.status, r.note, r.requested_by, COALESCE(ru.name, ''), COALESCE(ru.email, ''),
	r.decision_note, r.decided_by, COALESCE(du.name, ''), COALESCE(du.email, ''), r.created_at, r.decided_at
	FROM review_requests r
	JOIN modules m ON m.id = r.module_id
	LEFT JOIN lessons l ON l.id = r.lesson_id
	LEFT JOIN users ru ON ru.id = r.requested_by
	LEFT JOIN users du ON du.id = r.decided_by`

func scanReview(row pgx.Row) (Review, error) {
	var v Review
	err := row.Scan(&v.ID, &v.ModuleID, &v.CourseSlug, &v.CourseTitle, &v.LessonID, &v.LessonSlug, &v.LessonTitle,
		&v.Status, &v.Note, &v.RequestedBy, &v.RequesterName, &v.RequesterEmail,
		&v.DecisionNote, &v.DecidedBy, &v.DeciderName, &v.DeciderEmail, &v.CreatedAt, &v.DecidedAt)
	return v, err
}

// Create opens a pending request; a second pending request for the same target is a unique violation.
func (r *ReviewRepo) Create(ctx context.Context, moduleID int, lessonID *int, userID int, note string) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO review_requests (module_id, lesson_id, requested_by, note) VALUES ($1, $2, $3, $4) RETURNING id`,
		moduleID, lessonID, userID, note).Scan(&id)
	return id, err
}

func (r *ReviewRepo) Get(ctx context.Context, id int) (*Review, error) {
	v, err := scanReview(r.pool.QueryRow(ctx, reviewSelect+` WHERE r.id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// List returns requests matching the filter, newest first, at most 200.
func (r *ReviewRepo) List(ctx context.Context, f ReviewFilter) ([]Review, error) {
	return r.query(ctx, reviewSelect+`
		WHERE ($1 = '' OR r.status = $1) AND ($2 = 0 OR r.module_id = $2)
		  AND ($3 = 0 OR m.owner_id = $3 OR EXISTS (SELECT 1 FROM course_authors ca WHERE ca.module_id = m.id AND ca.user_id = $3))
		ORDER BY r.created_at DESC, r.id DESC LIMIT 200`, f.Status, f.ModuleID, f.EditorID)
}

// Latest returns the newest request per target (course or lesson) of the given courses, keyed by lesson id with 0 for the course.
func (r *ReviewRepo) Latest(ctx context.Context, moduleIDs []int) (map[int]map[int]Review, error) {
	rows, err := r.query(ctx, reviewSelect+` WHERE r.module_id = ANY($1) ORDER BY r.created_at, r.id`, moduleIDs)
	if err != nil {
		return nil, err
	}
	out := map[int]map[int]Review{}
	for _, v := range rows {
		key := 0
		if v.LessonID != nil {
			key = *v.LessonID
		}
		if out[v.ModuleID] == nil {
			out[v.ModuleID] = map[int]Review{}
		}
		out[v.ModuleID][key] = v
	}
	return out, nil
}

func (r *ReviewRepo) query(ctx context.Context, sql string, args ...any) ([]Review, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Review
	for rows.Next() {
		v, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// Decide approves or rejects a pending request; approval publishes its course or lesson in the same transaction.
func (r *ReviewRepo) Decide(ctx context.Context, id int, approve bool, adminID int, note string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	status := ReviewRejected
	if approve {
		status = ReviewApproved
	}
	var moduleID int
	var lessonID *int
	err = tx.QueryRow(ctx,
		`UPDATE review_requests SET status = $1, decided_by = $2, decision_note = $3, decided_at = NOW()
		 WHERE id = $4 AND status = 'pending' RETURNING module_id, lesson_id`,
		status, adminID, note, id).Scan(&moduleID, &lessonID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrReviewClosed
	}
	if err != nil {
		return err
	}
	if approve {
		if lessonID != nil {
			_, err = tx.Exec(ctx, `UPDATE lessons SET published = TRUE WHERE id = $1`, *lessonID)
		} else {
			_, err = tx.Exec(ctx, `UPDATE modules SET published = TRUE WHERE id = $1`, moduleID)
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Cancel withdraws a pending request.
func (r *ReviewRepo) Cancel(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE review_requests SET status = 'cancelled', decided_at = NOW() WHERE id = $1 AND status = 'pending'`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrReviewClosed
	}
	return err
}
