// Package migrate applies the SQL files in migrations/ exactly once each,
// tracking what ran in a schema_migrations table. Both cmd/server and cmd/seed
// call Up on startup, so a fresh database (or a stale volume missing late
// columns) is brought up to date without any manual psql step.
package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultDir is the migrations directory relative to the repo root.
const DefaultDir = "migrations"

// Up applies every pending *.up.sql file in dir, in filename order. Each file
// runs inside its own transaction together with the bookkeeping insert, so a
// failing migration leaves no partial record behind.
func Up(ctx context.Context, pool *pgxpool.Pool, dir string) ([]string, error) {
	if dir == "" {
		dir = DefaultDir
	}
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := make(map[string]bool)
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return nil, err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	var ran []string
	for _, f := range files {
		version := strings.TrimSuffix(filepath.Base(f), ".up.sql")
		if applied[version] {
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return ran, fmt.Errorf("read %s: %w", f, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return ran, err
		}
		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return ran, fmt.Errorf("apply %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING`, version); err != nil {
			_ = tx.Rollback(ctx)
			return ran, err
		}
		if err := tx.Commit(ctx); err != nil {
			return ran, err
		}
		ran = append(ran, version)
	}
	return ran, nil
}
