package repository

import (
	"context"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SpecRepo struct {
	pool *pgxpool.Pool
}

func NewSpecRepo(pool *pgxpool.Pool) *SpecRepo {
	return &SpecRepo{pool: pool}
}

func (r *SpecRepo) List(ctx context.Context) ([]model.Specialization, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT slug, name, icon, description, order_num FROM specializations ORDER BY order_num, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var specs []model.Specialization
	for rows.Next() {
		var s model.Specialization
		if err := rows.Scan(&s.Slug, &s.Name, &s.Icon, &s.Description, &s.OrderNum); err != nil {
			return nil, err
		}
		specs = append(specs, s)
	}
	return specs, rows.Err()
}

func (r *SpecRepo) Get(ctx context.Context, slug string) (*model.Specialization, error) {
	var s model.Specialization
	err := r.pool.QueryRow(ctx,
		`SELECT slug, name, icon, description, order_num FROM specializations WHERE slug = $1`, slug).
		Scan(&s.Slug, &s.Name, &s.Icon, &s.Description, &s.OrderNum)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SpecRepo) Upsert(ctx context.Context, s model.Specialization) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO specializations (slug, name, icon, description, order_num)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (slug) DO UPDATE SET name=EXCLUDED.name, icon=EXCLUDED.icon,
		   description=EXCLUDED.description, order_num=EXCLUDED.order_num`,
		s.Slug, s.Name, s.Icon, s.Description, s.OrderNum)
	return err
}

func (r *SpecRepo) Delete(ctx context.Context, slug string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM specializations WHERE slug = $1`, slug)
	return err
}
