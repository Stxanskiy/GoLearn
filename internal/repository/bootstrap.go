package repository

import (
	"context"
	"os"
)

// BootstrapAdmin creates the first admin account when the users table is empty,
// so a freshly migrated database is usable without opening self-registration.
// Credentials come from ADMIN_EMAIL / ADMIN_PASSWORD (defaults below).
// Returns the email that was created, or "" if an account already existed.
func (r *UserRepo) BootstrapAdmin(ctx context.Context) (string, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return "", err
	}
	if count > 0 {
		return "", nil
	}
	email := envOr("ADMIN_EMAIL", "admin@golearn.local")
	password := envOr("ADMIN_PASSWORD", "golearn123")
	name := envOr("ADMIN_NAME", "Admin")

	if _, err := r.CreateWithRole(ctx, email, password, name, "admin"); err != nil {
		return "", err
	}
	return email, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
