package repository

import (
	"context"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedDetachRepo moves content out from under the seeder.
//
// The seeder owns any lesson or module whose edited_at is NULL and rewrites it
// on every deploy from the JSON in cmd/seed/content. Stamping edited_at hands
// ownership to the database — the same thing that happens the moment an author
// saves in the studio, done to everything at once.
//
// It is one-way by design. After this the files no longer reach the site, and
// the studio is where courses are written.
type SeedDetachRepo struct {
	pool *pgxpool.Pool
}

func NewSeedDetachRepo(pool *pgxpool.Pool) *SeedDetachRepo {
	return &SeedDetachRepo{pool: pool}
}

// MarkdownLessons lists the lessons still stored as Markdown, with enough of
// each to render it. Only id, format and content are read: this runs over every
// lesson in the database, and the rest would be carried for nothing.
func (r *SeedDetachRepo) MarkdownLessons(ctx context.Context) ([]model.Lesson, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, format, content FROM lessons WHERE format = 'md' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Lesson
	for rows.Next() {
		var l model.Lesson
		if err := rows.Scan(&l.ID, &l.Format, &l.Content); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// SetLessonContent replaces a lesson's body and format.
//
// It does not touch edited_at: the caller stamps everything in one statement
// afterwards, and doing it here as well would make a partial run look like a
// finished one.
func (r *SeedDetachRepo) SetLessonContent(ctx context.Context, id int, format, content string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lessons SET format = $2, content = $3 WHERE id = $1`, id, format, content)
	return err
}

// Detach stamps edited_at on every lesson and module that still carries none,
// and reports how many of each that was.
//
// A lesson the seeder skips takes its quiz and its tasks with it — the seeder
// moves on to the next lesson rather than descending into one it does not own —
// so this is enough to detach a course completely.
func (r *SeedDetachRepo) Detach(ctx context.Context) (lessons, modules int64, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	lt, err := tx.Exec(ctx, `UPDATE lessons SET edited_at = now() WHERE edited_at IS NULL`)
	if err != nil {
		return 0, 0, err
	}
	mt, err := tx.Exec(ctx, `UPDATE modules SET edited_at = now() WHERE edited_at IS NULL`)
	if err != nil {
		return 0, 0, err
	}
	return lt.RowsAffected(), mt.RowsAffected(), tx.Commit(ctx)
}
