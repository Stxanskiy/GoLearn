package repository

import (
	"context"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SimRepo struct {
	pool *pgxpool.Pool
}

func NewSimRepo(pool *pgxpool.Pool) *SimRepo {
	return &SimRepo{pool: pool}
}

const simCols = `slug, title, icon, role, order_num, published, owner_id, data`

func scanSim(row pgx.Row) (model.Simulator, error) {
	var s model.Simulator
	var data []byte
	err := row.Scan(&s.Slug, &s.Title, &s.Icon, &s.Role, &s.OrderNum, &s.Published, &s.OwnerID, &data)
	s.Data = string(data)
	return s, err
}

func (r *SimRepo) query(ctx context.Context, sql string, args ...any) ([]model.Simulator, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sims []model.Simulator
	for rows.Next() {
		s, err := scanSim(rows)
		if err != nil {
			return nil, err
		}
		sims = append(sims, s)
	}
	return sims, rows.Err()
}

// List returns every simulator (admin).
func (r *SimRepo) List(ctx context.Context) ([]model.Simulator, error) {
	return r.query(ctx, `SELECT `+simCols+` FROM simulators ORDER BY order_num, title`)
}

// ListPublished returns only published simulators (students).
func (r *SimRepo) ListPublished(ctx context.Context) ([]model.Simulator, error) {
	return r.query(ctx, `SELECT `+simCols+` FROM simulators WHERE published ORDER BY order_num, title`)
}

func (r *SimRepo) Get(ctx context.Context, slug string) (*model.Simulator, error) {
	s, err := scanSim(r.pool.QueryRow(ctx, `SELECT `+simCols+` FROM simulators WHERE slug = $1`, slug))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SimRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM simulators`).Scan(&n)
	return n, err
}

// Upsert inserts or updates a simulator by slug (owner preserved on update).
func (r *SimRepo) Upsert(ctx context.Context, s model.Simulator) error {
	data := s.Data
	if data == "" {
		data = "{}"
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO simulators (slug, title, icon, role, order_num, published, owner_id, data)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		 ON CONFLICT (slug) DO UPDATE SET title=EXCLUDED.title, icon=EXCLUDED.icon, role=EXCLUDED.role,
		   order_num=EXCLUDED.order_num, published=EXCLUDED.published, data=EXCLUDED.data`,
		s.Slug, s.Title, s.Icon, s.Role, s.OrderNum, s.Published, s.OwnerID, data)
	return err
}

func (r *SimRepo) Delete(ctx context.Context, slug string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM simulators WHERE slug = $1`, slug)
	return err
}

func (r *SimRepo) SetPublished(ctx context.Context, slug string, published bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE simulators SET published = $1 WHERE slug = $2`, published, slug)
	return err
}

// Move swaps a simulator's order_num with its neighbour (dir "up"/"down").
func (r *SimRepo) Move(ctx context.Context, slug, dir string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var ord int
	if err := tx.QueryRow(ctx, `SELECT order_num FROM simulators WHERE slug=$1`, slug).Scan(&ord); err != nil {
		return err
	}
	q := `SELECT slug, order_num FROM simulators WHERE order_num > $1 ORDER BY order_num ASC LIMIT 1`
	if dir == "up" {
		q = `SELECT slug, order_num FROM simulators WHERE order_num < $1 ORDER BY order_num DESC LIMIT 1`
	}
	var nslug string
	var nord int
	if err := tx.QueryRow(ctx, q, ord).Scan(&nslug, &nord); err != nil {
		if err == pgx.ErrNoRows {
			return tx.Commit(ctx)
		}
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE simulators SET order_num=$1 WHERE slug=$2`, nord, slug); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE simulators SET order_num=$1 WHERE slug=$2`, ord, nslug); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
