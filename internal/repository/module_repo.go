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
	category, label, tags, cover_image, accent, est_minutes, source, published, owner_id, created_at`

func scanModule(row pgx.Row) (model.Module, error) {
	var m model.Module
	var prereqJSON, tagsJSON []byte
	err := row.Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.OrderNum, &m.Track, &m.Difficulty,
		&prereqJSON, &m.Category, &m.Label, &tagsJSON, &m.CoverImage, &m.Accent, &m.EstMinutes, &m.Source,
		&m.Published, &m.OwnerID, &m.CreatedAt)
	if err != nil {
		return m, err
	}
	_ = json.Unmarshal(prereqJSON, &m.Prerequisites)
	_ = json.Unmarshal(tagsJSON, &m.Tags)
	return m, nil
}

// GetAll returns published modules — the default, student-safe listing. Admin
// screens use GetForAdmin to also see their own drafts.
func (r *ModuleRepo) GetAll(ctx context.Context) ([]model.Module, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+moduleCols+` FROM modules WHERE published ORDER BY order_num`)
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
		`SELECT `+moduleCols+` FROM modules WHERE (track = $1 OR track = 'shared') AND published ORDER BY order_num`, track)
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
		   category, label, tags, cover_image, accent, est_minutes, source, published, owner_id)
		 VALUES ($1,$2,$3,$4,$5,$6,'[]',$7,$8,$9,$10,$11,$12,'admin',$13,$14) RETURNING id`,
		m.Slug, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
		m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes, m.Published, m.OwnerID).Scan(&id)
	return id, err
}

// Update modifies an existing module by id.
func (r *ModuleRepo) Update(ctx context.Context, m model.Module) error {
	tags, _ := json.Marshal(m.Tags)
	_, err := r.pool.Exec(ctx,
		`UPDATE modules SET slug=$1, title=$2, description=$3, order_num=$4, track=$5, difficulty=$6,
		   category=$7, label=$8, tags=$9, cover_image=$10, accent=$11, est_minutes=$12, published=$13 WHERE id=$14`,
		m.Slug, m.Title, m.Description, m.OrderNum, m.Track, m.Difficulty,
		m.Category, m.Label, tags, m.CoverImage, m.Accent, m.EstMinutes, m.Published, m.ID)
	return err
}

// GetForAdmin returns every module an admin may see: all published ones, plus
// their own drafts and ownerless (seed) drafts — a colleague's drafts stay hidden.
func (r *ModuleRepo) GetForAdmin(ctx context.Context, adminID int) ([]model.Module, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+moduleCols+` FROM modules
		 WHERE published OR owner_id = $1 OR owner_id IS NULL
		 ORDER BY order_num`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var mods []model.Module
	for rows.Next() {
		m, err := scanModule(rows)
		if err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, rows.Err()
}

// SetPublished toggles a module's draft/published state.
func (r *ModuleRepo) SetPublished(ctx context.Context, id int, published bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE modules SET published = $1 WHERE id = $2`, published, id)
	return err
}

// Move swaps a module's order_num with its neighbour (dir "up"/"down") within the
// same track — this is the course order in the catalogue and the roadmap.
func (r *ModuleRepo) Move(ctx context.Context, id int, dir string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var track string
	var ord int
	if err := tx.QueryRow(ctx, `SELECT track, order_num FROM modules WHERE id=$1`, id).Scan(&track, &ord); err != nil {
		return err
	}
	q := `SELECT id, order_num FROM modules WHERE track=$1 AND order_num > $2 ORDER BY order_num ASC LIMIT 1`
	if dir == "up" {
		q = `SELECT id, order_num FROM modules WHERE track=$1 AND order_num < $2 ORDER BY order_num DESC LIMIT 1`
	}
	var nid, nord int
	if err := tx.QueryRow(ctx, q, track, ord).Scan(&nid, &nord); err != nil {
		if err == pgx.ErrNoRows {
			return tx.Commit(ctx)
		}
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE modules SET order_num=$1 WHERE id=$2`, nord, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE modules SET order_num=$1 WHERE id=$2`, ord, nid); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
		 WHERE track = ANY($1) AND published AND order_num < $2
		 ORDER BY order_num DESC LIMIT 1`, tracks, m.OrderNum))
	if err == nil {
		prev = &p
	} else if err != pgx.ErrNoRows {
		return nil, nil, err
	}

	n, err := scanModule(r.pool.QueryRow(ctx,
		`SELECT `+moduleCols+` FROM modules
		 WHERE track = ANY($1) AND published AND order_num > $2
		 ORDER BY order_num ASC LIMIT 1`, tracks, m.OrderNum))
	if err == nil {
		next = &n
	} else if err != pgx.ErrNoRows {
		return nil, nil, err
	}
	return prev, next, nil
}

// PlatformStats are the headline numbers shown on the public landing page.
type PlatformStats struct {
	Courses     int
	Lessons     int
	Labs        int
	AutoChecked int
}

// Stats counts what the platform currently offers. Shown to visitors, so the
// numbers come from the database rather than from a hand-written claim.
func (r *ModuleRepo) Stats(ctx context.Context) (PlatformStats, error) {
	var s PlatformStats
	err := r.pool.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM modules),
		       (SELECT count(*) FROM lessons),
		       (SELECT count(*) FROM lessons WHERE kind = 'lab'),
		       (SELECT count(*) FROM tasks  WHERE check_script <> '')`).
		Scan(&s.Courses, &s.Lessons, &s.Labs, &s.AutoChecked)
	return s, err
}
