package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CourseAuthorRepo stores course co-authors.
type CourseAuthorRepo struct {
	pool *pgxpool.Pool
}

func NewCourseAuthorRepo(pool *pgxpool.Pool) *CourseAuthorRepo {
	return &CourseAuthorRepo{pool: pool}
}

// IsCoauthor reports whether the user co-authors the course.
func (r *CourseAuthorRepo) IsCoauthor(ctx context.Context, moduleID, userID int) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM course_authors WHERE module_id = $1 AND user_id = $2)`,
		moduleID, userID).Scan(&ok)
	return ok, err
}

// List returns the co-authors of a course in the order they were added.
func (r *CourseAuthorRepo) List(ctx context.Context, moduleID int) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.email, u.name, u.role FROM course_authors ca JOIN users u ON u.id = ca.user_id
		 WHERE ca.module_id = $1 ORDER BY ca.created_at, u.id`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// Add makes the user a co-author; created is false when they already were one.
func (r *CourseAuthorRepo) Add(ctx context.Context, moduleID, userID, addedBy int) (created bool, err error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO course_authors (module_id, user_id, added_by) VALUES ($1, $2, $3)
		 ON CONFLICT DO NOTHING`, moduleID, userID, addedBy)
	return tag.RowsAffected() == 1, err
}

// Remove deletes a co-author; removed is false when the user was not one.
func (r *CourseAuthorRepo) Remove(ctx context.Context, moduleID, userID int) (removed bool, err error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM course_authors WHERE module_id = $1 AND user_id = $2`, moduleID, userID)
	return tag.RowsAffected() == 1, err
}

// SetOwner transfers ownership; the previous owner becomes a co-author.
func (r *CourseAuthorRepo) SetOwner(ctx context.Context, moduleID, userID, byID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var prev *int
	if err := tx.QueryRow(ctx, `SELECT owner_id FROM modules WHERE id = $1 FOR UPDATE`, moduleID).Scan(&prev); err != nil {
		return err
	}
	if prev != nil && *prev != userID {
		if _, err := tx.Exec(ctx,
			`INSERT INTO course_authors (module_id, user_id, added_by) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			moduleID, *prev, byID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM course_authors WHERE module_id = $1 AND user_id = $2`, moduleID, userID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE modules SET owner_id = $1 WHERE id = $2`, userID, moduleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return tx.Commit(ctx)
}
