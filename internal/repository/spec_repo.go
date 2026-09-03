package repository

import (
	"context"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpecRepo struct {
	pool *pgxpool.Pool
}

func NewSpecRepo(pool *pgxpool.Pool) *SpecRepo {
	return &SpecRepo{pool: pool}
}

const specCols = `slug, name, icon, description, order_num, cover_image, published, owner_id`

func scanSpec(row pgx.Row) (model.Specialization, error) {
	var s model.Specialization
	err := row.Scan(&s.Slug, &s.Name, &s.Icon, &s.Description, &s.OrderNum, &s.CoverImage, &s.Published, &s.OwnerID)
	return s, err
}

// List returns every section (admin view).
func (r *SpecRepo) List(ctx context.Context) ([]model.Specialization, error) {
	return r.query(ctx, `SELECT `+specCols+` FROM specializations ORDER BY order_num, name`)
}

// ListPublished returns only published sections (student catalogue).
func (r *SpecRepo) ListPublished(ctx context.Context) ([]model.Specialization, error) {
	return r.query(ctx, `SELECT `+specCols+` FROM specializations WHERE published ORDER BY order_num, name`)
}

func (r *SpecRepo) query(ctx context.Context, sql string, args ...any) ([]model.Specialization, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var specs []model.Specialization
	for rows.Next() {
		s, err := scanSpec(rows)
		if err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, rows.Err()
}

func (r *SpecRepo) Get(ctx context.Context, slug string) (*model.Specialization, error) {
	s, err := scanSpec(r.pool.QueryRow(ctx, `SELECT `+specCols+` FROM specializations WHERE slug = $1`, slug))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Upsert inserts or updates a section. On update the owner is preserved (a section
// keeps its original owner); published state and the visible fields are updated.
func (r *SpecRepo) Upsert(ctx context.Context, s model.Specialization) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO specializations (slug, name, icon, description, order_num, cover_image, published, owner_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name, icon=EXCLUDED.icon,
		   description=EXCLUDED.description, order_num=EXCLUDED.order_num,
		   cover_image=COALESCE(NULLIF(EXCLUDED.cover_image,''), specializations.cover_image),
		   published=EXCLUDED.published`,
		s.Slug, s.Name, s.Icon, s.Description, s.OrderNum, s.CoverImage, s.Published, s.OwnerID)
	return err
}

// SetPublished toggles a section's draft/published state.
func (r *SpecRepo) SetPublished(ctx context.Context, slug string, published bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE specializations SET published = $1 WHERE slug = $2`, published, slug)
	return err
}

// Move swaps a section's order_num with its neighbour (dir "up"/"down").
func (r *SpecRepo) Move(ctx context.Context, slug, dir string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var ord int
	if err := tx.QueryRow(ctx, `SELECT order_num FROM specializations WHERE slug=$1`, slug).Scan(&ord); err != nil {
		return err
	}
	q := `SELECT slug, order_num FROM specializations WHERE order_num > $1 ORDER BY order_num ASC LIMIT 1`
	if dir == "up" {
		q = `SELECT slug, order_num FROM specializations WHERE order_num < $1 ORDER BY order_num DESC LIMIT 1`
	}
	var nslug string
	var nord int
	if err := tx.QueryRow(ctx, q, ord).Scan(&nslug, &nord); err != nil {
		if err == pgx.ErrNoRows {
			return tx.Commit(ctx)
		}
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE specializations SET order_num=$1 WHERE slug=$2`, nord, slug); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE specializations SET order_num=$1 WHERE slug=$2`, ord, nslug); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SpecRepo) Delete(ctx context.Context, slug string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM specializations WHERE slug = $1`, slug)
	return err
}
