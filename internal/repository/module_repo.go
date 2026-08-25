package repository

import (
	"context"
	"encoding/json"

	"github.com/backendraz/golearn/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ModuleRepo struct {
	pool *pgxpool.Pool
}

func NewModuleRepo(pool *pgxpool.Pool) *ModuleRepo {
	return &ModuleRepo{pool: pool}
}

const moduleCols = `id, slug, title, description, order_num, track, difficulty, prerequisites,
	category, label, tags, cover_image, accent, est_minutes, source, created_at`

func scanModule(row pgx.Row) (model.Module, error) {
	var m model.Module
	var prereqJSON, tagsJSON []byte
	err := row.Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.OrderNum, &m.Track, &m.Difficulty,
		&prereqJSON, &m.Category, &m.Label, &tagsJSON, &m.CoverImage, &m.Accent, &m.EstMinutes, &m.Source, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	_ = json.Unmarshal(prereqJSON, &m.Prerequisites)
	_ = json.Unmarshal(tagsJSON, &m.Tags)
	return m, nil
}

func (r *ModuleRepo) GetAll(ctx context.Context) ([]model.Module, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+moduleCols+` FROM modules ORDER BY order_num`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []model.Module
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

func (r *ModuleRepo) GetByTrack(ctx context.Context, track string) ([]model.Module, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+moduleCols+` FROM modules WHERE track = $1 OR track = 'shared' ORDER BY order_num`, track)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []model.Module
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

func (r *ModuleRepo) GetBySlug(ctx context.Context, slug string) (*model.Module, error) {
	m, err := scanModule(r.pool.QueryRow(ctx, `SELECT `+moduleCols+` FROM modules WHERE slug = $1`, slug))
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create inserts an admin-authored module and returns its id.
func (r *ModuleRepo) Create(ctx context.Context, m model.Module) (int, error) {
	tags, _ := json.Marshal(m.Tags)
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites,
		   category, label, tags, cover_image, accent, est_minutes, source)
		 VALUES ($1,$2,$3,$4,$5,$6,'[]',$7,$8,$9,$10,$11,$12,'admin') RETURNING id`,
		m.Slug, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
		m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes).Scan(&id)
	return id, err
}

// Update modifies an existing module by id.
func (r *ModuleRepo) Update(ctx context.Context, m model.Module) error {
	tags, _ := json.Marshal(m.Tags)
	_, err := r.pool.Exec(ctx,
		`UPDATE modules SET slug=$1, title=$2, description=$3, order_num=$4, track=$5, difficulty=$6,
		   category=$7, label=$8, tags=$9, cover_image=$10, accent=$11, est_minutes=$12 WHERE id=$13`,
		m.Slug, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
		m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes, m.ID)
	return err
}

func (r *ModuleRepo) Delete(ctx context.Context, id int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM modules WHERE id = $1`, id)
	return err
}

func (r *ModuleRepo) GetByID(ctx context.Context, id int) (*model.Module, error) {
	m, err := scanModule(r.pool.QueryRow(ctx, `SELECT `+moduleCols+` FROM modules WHERE id = $1`, id))
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Neighbors returns the previous and next course inside the same
// specialization, following the curriculum order (order_num). Either may be
// nil at the ends of the path.
func (r *ModuleRepo) Neighbors(ctx context.Context, m model.Module, tracks []string) (prev, next *model.Module, err error) {
	p, err := scanModule(r.pool.QueryRow(ctx,
		`SELECT `+moduleCols+` FROM modules
		 WHERE track = ANY($1) AND order_num < $2
		 ORDER BY order_num DESC LIMIT 1`, tracks, m.OrderNum))
	if err == nil {
		prev = &p
	} else if err != pgx.ErrNoRows {
		return nil, nil, err
	}

	n, err := scanModule(r.pool.QueryRow(ctx,
		`SELECT `+moduleCols+` FROM modules
		 WHERE track = ANY($1) AND order_num > $2
		 ORDER BY order_num ASC LIMIT 1`, tracks, m.OrderNum))
	if err == nil {
		next = &n
	} else if err != pgx.ErrNoRows {
		return nil, nil, err
	}
	return prev, next, nil
}
