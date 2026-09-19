package repository

import (
	"context"
	"encoding/json"
	"strings"

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
	category, label, tags, cover_image, icon_url, accent, est_minutes, source, published, owner_id, created_at, draft_of, access_tier`

func scanModule(row pgx.Row) (model.Module, error) {
	var m model.Module
	var prereqJSON, tagsJSON []byte
	err := row.Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.OrderNum, &m.Track, &m.Difficulty,
		&prereqJSON, &m.Category, &m.Label, &tagsJSON, &m.CoverImage, &m.IconURL, &m.Accent, &m.EstMinutes, &m.Source,
		&m.Published, &m.OwnerID, &m.CreatedAt, &m.DraftOf, &m.AccessTier)
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
		 WHERE draft_of IS NULL AND (published OR owner_id = $1 OR owner_id IS NULL)
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

// CourseRow is a module with lesson counts for the course manager; CoverImage is "data:" for uploaded covers.
type CourseRow struct {
	model.Module
	Lessons int
	Labs    int
}

// ListManaged returns all courses when all is true, otherwise the ones the user owns or co-authors.
func (r *ModuleRepo) ListManaged(ctx context.Context, userID int, all bool) ([]CourseRow, error) {
	cols := strings.Replace(moduleCols, "cover_image", "CASE WHEN cover_image LIKE 'data:%' THEN 'data:' ELSE cover_image END", 1)
	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+`,
		   (SELECT count(*) FROM lessons l WHERE l.module_id = modules.id),
		   (SELECT count(*) FROM lessons l WHERE l.module_id = modules.id AND l.kind = 'lab')
		 FROM modules
		 WHERE draft_of IS NULL
		   AND ($2 OR owner_id = $1 OR EXISTS (SELECT 1 FROM course_authors ca WHERE ca.module_id = modules.id AND ca.user_id = $1))
		 ORDER BY order_num, id`, userID, all)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CourseRow
	for rows.Next() {
		var c CourseRow
		var prereqJSON, tagsJSON []byte
		m := &c.Module
		if err := rows.Scan(&m.ID, &m.Slug, &m.Title, &m.Description, &m.OrderNum, &m.Track, &m.Difficulty,
			&prereqJSON, &m.Category, &m.Label, &tagsJSON, &m.CoverImage, &m.IconURL, &m.Accent, &m.EstMinutes, &m.Source,
			&m.Published, &m.OwnerID, &m.CreatedAt, &m.DraftOf, &m.AccessTier, &c.Lessons, &c.Labs); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(prereqJSON, &m.Prerequisites)
		_ = json.Unmarshal(tagsJSON, &m.Tags)
		out = append(out, c)
	}
	return out, rows.Err()
}

// TrackCounts returns the number of courses per track.
func (r *ModuleRepo) TrackCounts(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT track, count(*) FROM modules WHERE draft_of IS NULL GROUP BY track`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var track string
		var n int
		if err := rows.Scan(&track, &n); err != nil {
			return nil, err
		}
		out[track] = n
	}
	return out, rows.Err()
}

// DraftFor returns the draft copy of a live course.
func (r *ModuleRepo) DraftFor(ctx context.Context, liveID int) (*model.Module, error) {
	m, err := scanModule(r.pool.QueryRow(ctx, `SELECT `+moduleCols+` FROM modules WHERE draft_of = $1`, liveID))
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Drafts maps live course ids to the ids of their drafts.
func (r *ModuleRepo) Drafts(ctx context.Context) (map[int]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT draft_of, id FROM modules WHERE draft_of IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int]int{}
	for rows.Next() {
		var live, draft int
		if err := rows.Scan(&live, &draft); err != nil {
			return nil, err
		}
		out[live] = draft
	}
	return out, rows.Err()
}

// NextOrder returns an order_num after every existing course.
func (r *ModuleRepo) NextOrder(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(order_num), 0) + 1 FROM modules`).Scan(&n)
	return n, err
}

// SetIcon replaces the course icon URL; empty falls back to the derived icon.
func (r *ModuleRepo) SetIcon(ctx context.Context, id int, iconURL string) error {
	_, err := r.pool.Exec(ctx, `UPDATE modules SET icon_url = $1 WHERE id = $2`, iconURL, id)
	return err
}

// SetCover replaces the cover (data URI, URL or empty for the generated one).
func (r *ModuleRepo) SetCover(ctx context.Context, id int, cover string) error {
	_, err := r.pool.Exec(ctx, `UPDATE modules SET cover_image = $1 WHERE id = $2`, cover, id)
	return err
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
	q := `SELECT id, order_num FROM modules WHERE track=$1 AND draft_of IS NULL AND order_num > $2 ORDER BY order_num ASC LIMIT 1`
	if dir == "up" {
		q = `SELECT id, order_num FROM modules WHERE track=$1 AND draft_of IS NULL AND order_num < $2 ORDER BY order_num DESC LIMIT 1`
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

// Stats counts published content; trainer (gym) modules are not counted as courses.
func (r *ModuleRepo) Stats(ctx context.Context) (PlatformStats, error) {
	var s PlatformStats
	err := r.pool.QueryRow(ctx, `
		WITH pub AS (
			SELECT l.id, l.kind FROM lessons l JOIN modules m ON m.id = l.module_id
			WHERE l.published AND m.published
		)
		SELECT (SELECT count(*) FROM modules WHERE published AND track <> 'gym'),
		       (SELECT count(*) FROM pub),
		       (SELECT count(*) FROM pub WHERE kind = 'lab'),
		       (SELECT count(*) FROM tasks t JOIN pub ON pub.id = t.lesson_id WHERE t.check_script <> '')`).
		Scan(&s.Courses, &s.Lessons, &s.Labs, &s.AutoChecked)
	return s, err
}
