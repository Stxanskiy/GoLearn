package repository

import (
	"context"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ModuleRepo struct {
	pool *pgxpool.Pool
}

func NewModuleRepo(pool *pgxpool.Pool) *ModuleRepo {
	return &ModuleRepo{pool: pool}
}

func (r *ModuleRepo) GetAll(ctx context.Context) ([]model.Module, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, title, description, order_num, created_at
		FROM modules ORDER BY order_num`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []model.Module
	for rows.Next() {
		var m model.Module
		if err := rows.Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.OrderNum, &m.CreatedAt); err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

func (r *ModuleRepo) GetBySlug(ctx context.Context, slug string) (*model.Module, error) {
	var m model.Module
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, title, description, order_num, created_at
		FROM modules WHERE slug = $1`, slug).
		Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.OrderNum, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
