package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	Email        string
	PasswordHash string
	Name         string
	Role         string
	Blocked      bool
	CreatedAt    time.Time
}

// User roles.
const (
	RoleStudent = "student"
	RoleAuthor  = "author"
	RoleAdmin   = "admin"
)

func (u *User) IsAdmin() bool { return u != nil && u.Role == RoleAdmin }

// CanAuthor reports whether the user may manage courses.
func (u *User) CanAuthor() bool { return u != nil && (u.Role == RoleAuthor || u.Role == RoleAdmin) }

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, email, password, name string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user User
	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, name, role, created_at`,
		email, string(hash), name,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	return &user, err
}

// CreateWithRole creates a user with an explicit role (used by the admin panel).
func (r *UserRepo) CreateWithRole(ctx context.Context, email, password, name, role string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	var user User
	err = r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, role) VALUES ($1, $2, $3, $4)
		 RETURNING id, email, password_hash, name, role, created_at`,
		email, string(hash), name, role,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	return &user, err
}

// List returns all users (newest first) for the admin panel; password hashes omitted.
func (r *UserRepo) List(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, name, role, blocked, created_at FROM users ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Blocked, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ListPage returns up to limit users with id below beforeID (0 = newest) whose name or email contains q.
func (r *UserRepo) ListPage(ctx context.Context, q string, beforeID, limit int) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, email, name, role, blocked, created_at FROM users
		 WHERE ($1 = 0 OR id < $1)
		   AND ($2 = '' OR name ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%')
		 ORDER BY id DESC LIMIT $3`, beforeID, escapeLike(q), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Blocked, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// ErrLastAdmin means a change would leave the platform without an active admin.
var ErrLastAdmin = errors.New("no active admin would remain")

// adminGuardLock serializes changes that can remove admins.
const adminGuardLock = 7427001

// SetAccess sets role and blocked flag, revoking sessions of a blocked user; returns ErrLastAdmin instead of removing the last active admin.
func (r *UserRepo) SetAccess(ctx context.Context, userID int, role string, blocked bool) error {
	return r.withAdminGuard(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE users SET role = $1, blocked = $2 WHERE id = $3`, role, blocked, userID)
		if err == nil && tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		if err == nil && blocked {
			_, err = tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
		}
		return err
	})
}

// DeleteGuarded removes a user unless they are the last active admin.
func (r *UserRepo) DeleteGuarded(ctx context.Context, userID int) error {
	return r.withAdminGuard(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		return err
	})
}

func (r *UserRepo) withAdminGuard(ctx context.Context, change func(tx pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, adminGuardLock); err != nil {
		return err
	}
	if err := change(tx); err != nil {
		return err
	}
	var admins int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE role = 'admin' AND NOT blocked`).Scan(&admins); err != nil {
		return err
	}
	if admins == 0 {
		return ErrLastAdmin
	}
	return tx.Commit(ctx)
}

// DeleteSessions signs the user out everywhere.
func (r *UserRepo) DeleteSessions(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// SetPassword updates a user's password (admin reset).
func (r *UserRepo) SetPassword(ctx context.Context, userID int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, string(hash), userID)
	return err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, blocked, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.Blocked, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, name, role, blocked, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.Blocked, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) SetRole(ctx context.Context, userID int, role string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, userID)
	return err
}

func (r *UserRepo) CheckPassword(user *User, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}

// Session management

func (r *UserRepo) CreateSession(ctx context.Context, userID int) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days

	_, err = r.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	return token, err
}

func (r *UserRepo) GetUserBySession(ctx context.Context, token string) (*User, error) {
	var user User
	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password_hash, u.name, u.role, u.created_at
		 FROM users u JOIN sessions s ON s.user_id = u.id
		 WHERE s.token = $1 AND s.expires_at > NOW() AND u.blocked = FALSE`,
		token,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) DeleteSession(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

func (r *UserRepo) CleanExpiredSessions(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
