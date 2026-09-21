package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Billing is the one place where wrong SQL either costs a user money or gives
// the content away, so these run against a real database. They skip when none is
// configured, which keeps `go test ./...` green without one.
func billingPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("no TEST_DATABASE_URL or DATABASE_URL set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Skipf("database unreachable: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Skipf("database unreachable: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testUser(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	ctx := context.Background()
	var id int
	email := "billing-test-" + time.Now().Format("150405.000000000") + "@example.test"
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, name, role) VALUES ($1, 'x', 'billing test', 'student') RETURNING id`,
		email).Scan(&id)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func TestConfirmGrantsThenIsIdempotent(t *testing.T) {
	pool := billingPool(t)
	repo := NewBillingRepo(pool)
	ctx := context.Background()
	user := testUser(t, pool)

	if ok, err := repo.HasAccess(ctx, user); err != nil || ok {
		t.Fatalf("fresh user has access: %v %v", ok, err)
	}

	if _, err := repo.StartPayment(ctx, user, 1, 49000, "RUB", "stub", "ref-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	// A pending payment must not open the content.
	if ok, _ := repo.HasAccess(ctx, user); ok {
		t.Fatal("pending payment granted access")
	}

	sub, err := repo.Confirm(ctx, "stub", "ref-1")
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	first := sub.ExpiresAt
	if !sub.IsActive() {
		t.Fatalf("subscription not active after payment: %+v", sub)
	}

	// A provider replaying its webhook must not stack free months.
	again, err := repo.Confirm(ctx, "stub", "ref-1")
	if err != nil {
		t.Fatalf("confirm twice: %v", err)
	}
	if !again.ExpiresAt.Equal(first) {
		t.Fatalf("replayed webhook extended the subscription: %v -> %v", first, again.ExpiresAt)
	}

	// A genuinely new payment does extend it.
	if _, err := repo.StartPayment(ctx, user, 1, 49000, "RUB", "stub", "ref-2"); err != nil {
		t.Fatalf("start second: %v", err)
	}
	extended, err := repo.Confirm(ctx, "stub", "ref-2")
	if err != nil {
		t.Fatalf("confirm second: %v", err)
	}
	if !extended.ExpiresAt.After(first) {
		t.Fatalf("second payment did not extend: %v -> %v", first, extended.ExpiresAt)
	}
}

func TestExpiredSubscriptionGrantsNothing(t *testing.T) {
	pool := billingPool(t)
	repo := NewBillingRepo(pool)
	ctx := context.Background()
	user := testUser(t, pool)

	// A row left 'active' past its expiry is what a missing sweep leaves behind;
	// it must not keep the content open.
	if _, err := pool.Exec(ctx, `
		INSERT INTO subscriptions (user_id, status, started_at, expires_at)
		VALUES ($1, 'active', now() - interval '2 months', now() - interval '1 day')`, user); err != nil {
		t.Fatalf("seed: %v", err)
	}
	ok, err := repo.HasAccess(ctx, user)
	if err != nil {
		t.Fatalf("access: %v", err)
	}
	if ok {
		t.Fatal("expired subscription still grants access")
	}
	cur, err := repo.Current(ctx, user)
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if cur.Status != "expired" {
		t.Fatalf("status %q, want expired", cur.Status)
	}
}

func TestSweepExpiredMarksOnlyLapsed(t *testing.T) {
	pool := billingPool(t)
	repo := NewBillingRepo(pool)
	ctx := context.Background()
	lapsed := testUser(t, pool)
	current := testUser(t, pool)

	if _, err := pool.Exec(ctx, `
		INSERT INTO subscriptions (user_id, status, started_at, expires_at)
		VALUES ($1, 'active', now() - interval '2 months', now() - interval '1 day'),
		       ($2, 'active', now(), now() + interval '1 month')`, lapsed, current); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := repo.SweepExpired(ctx); err != nil {
		t.Fatalf("sweep: %v", err)
	}

	gone, err := repo.Current(ctx, lapsed)
	if err != nil || gone.Status != "expired" {
		t.Fatalf("lapsed subscription is %v (err %v), want expired", gone, err)
	}
	// The sweep must not touch a subscription that is still running: doing so
	// would lock out a paying customer.
	live, err := repo.Current(ctx, current)
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if !live.IsActive() {
		t.Fatalf("sweep expired a live subscription: %+v", live)
	}
}

func TestSetCourseTierRejectsUnknown(t *testing.T) {
	repo := NewBillingRepo(nil)
	if err := repo.SetCourseTier(context.Background(), 1, "premium"); err != ErrBadTier {
		t.Fatalf("got %v, want ErrBadTier", err)
	}
}
