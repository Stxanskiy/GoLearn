package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BillingRepo stores platform subscriptions and the payments behind them.
//
// Access is a platform-wide subscription, not a per-course purchase: a course is
// either free or part of the subscription, and an admin decides which.
type BillingRepo struct {
	pool *pgxpool.Pool
}

func NewBillingRepo(pool *pgxpool.Pool) *BillingRepo {
	return &BillingRepo{pool: pool}
}

// Subscription is one purchased period. Active is derived from status and expiry
// together, so a row left active past its expiry never grants access.
type Subscription struct {
	ID        int       `json:"id"`
	Status    string    `json:"status"`
	Plan      string    `json:"plan"`
	StartedAt time.Time `json:"started_at"`
	// ExpiresAt is nil for a lifetime subscription. Nothing else in the schema
	// can express "never", and a date far in the future would quietly lapse.
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Provider  string     `json:"provider"`
}

func (s *Subscription) IsActive() bool {
	if s == nil || s.Status != "active" {
		return false
	}
	return s.ExpiresAt == nil || s.ExpiresAt.After(time.Now())
}

// Payment is one attempt to buy a period, successful or not.
type Payment struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	AmountMinor int64      `json:"amount_minor"`
	Currency    string     `json:"currency"`
	Plan        string     `json:"plan"`
	Months      int        `json:"months"`
	Provider    string     `json:"provider"`
	ProviderRef string     `json:"provider_ref"`
	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

// Current returns the user's active subscription, or nil when there is none.
// Expired rows are reported as inactive rather than hidden, so a caller can tell
// "never subscribed" from "lapsed".
func (r *BillingRepo) Current(ctx context.Context, userID int) (*Subscription, error) {
	var s Subscription
	// NULLs sort first under DESC in PostgreSQL, so a lifetime row wins over any
	// dated one — which is the answer a holder of both should get.
	err := r.pool.QueryRow(ctx, `
		SELECT id, status, plan, started_at, expires_at, provider
		  FROM subscriptions
		 WHERE user_id = $1
		 ORDER BY expires_at DESC
		 LIMIT 1`, userID).Scan(&s.ID, &s.Status, &s.Plan, &s.StartedAt, &s.ExpiresAt, &s.Provider)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// A row can outlive its expiry if nothing swept it; report the truth.
	if s.Status == "active" && s.ExpiresAt != nil && !s.ExpiresAt.After(time.Now()) {
		s.Status = "expired"
	}
	return &s, nil
}

// HasAccess reports whether the user may open paid content right now.
func (r *BillingRepo) HasAccess(ctx context.Context, userID int) (bool, error) {
	sub, err := r.Current(ctx, userID)
	if err != nil {
		return false, err
	}
	return sub.IsActive(), nil
}

// StartPayment records a pending attempt. The provider reference is filled in by
// the caller once the provider hands one out; with the stub it is generated here.
func (r *BillingRepo) StartPayment(ctx context.Context, userID int, plan string, months int, amountMinor int64, currency, provider, ref string) (*Payment, error) {
	var p Payment
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (user_id, status, amount_minor, currency, plan, months, provider, provider_ref)
		VALUES ($1, 'pending', $2, $3, $4, $5, $6, $7)
		RETURNING id, status, amount_minor, currency, plan, months, provider, provider_ref, created_at, paid_at`,
		userID, amountMinor, currency, plan, months, provider, ref).
		Scan(&p.ID, &p.Status, &p.AmountMinor, &p.Currency, &p.Plan, &p.Months, &p.Provider, &p.ProviderRef, &p.CreatedAt, &p.PaidAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SetProviderRef records the identifier the provider will quote back.
//
// Robokassa's invoice number has to be a positive integer unique to the shop,
// and the payment row's own id is the only such number we have — but it does
// not exist until the row is inserted. So the row is created without a
// reference and given one here.
func (r *BillingRepo) SetProviderRef(ctx context.Context, paymentID int, ref string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE payments SET provider_ref = $2 WHERE id = $1 AND status = 'pending'`,
		paymentID, ref)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

// AnyAmount tells Confirm not to check what was paid. Only the stub provider
// passes it: with a real provider the amount is the thing most worth checking,
// because a signature proves who sent the callback and nothing about the sum.
const AnyAmount int64 = -1

// Confirm marks a payment paid and extends the subscription by its months.
//
// paidMinor is what the provider says was paid; it must match the amount the
// payment was opened for, or nothing is granted. Pass AnyAmount to skip that.
//
// It is safe to call twice with the same reference: a payment already marked
// paid returns its existing subscription instead of granting another period,
// which is what keeps a replayed provider webhook from stacking free months.
func (r *BillingRepo) Confirm(ctx context.Context, provider, ref string, paidMinor int64) (*Subscription, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var paymentID, userID, months int
	var status, plan string
	var wantMinor int64
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, plan, months, status, amount_minor FROM payments
		 WHERE provider = $1 AND provider_ref = $2
		 FOR UPDATE`, provider, ref).Scan(&paymentID, &userID, &plan, &months, &status, &wantMinor)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	// Checked before the already-handled branch: a replay carrying a different
	// sum is not a replay, it is someone editing the callback.
	if paidMinor != AnyAmount && paidMinor != wantMinor {
		return nil, ErrAmountMismatch
	}

	if status != "pending" {
		// Already handled; hand back what the user has rather than granting more.
		var s Subscription
		err = tx.QueryRow(ctx, `
			SELECT id, status, plan, started_at, expires_at, provider FROM subscriptions
			 WHERE user_id = $1 ORDER BY expires_at DESC LIMIT 1`, userID).
			Scan(&s.ID, &s.Status, &s.Plan, &s.StartedAt, &s.ExpiresAt, &s.Provider)
		if err != nil {
			return nil, err
		}
		return &s, tx.Commit(ctx)
	}

	// Extend an active subscription from its own expiry, so paying early never
	// costs the buyer the days they already hold. months = 0 is the lifetime
	// plan: it has no expiry, and once held it cannot be downgraded by a later
	// monthly purchase — paying more must never take access away.
	var s Subscription
	err = tx.QueryRow(ctx, `
		INSERT INTO subscriptions (user_id, status, plan, started_at, expires_at, provider, provider_ref)
		VALUES ($1, 'active', $5, now(),
		        CASE WHEN $2 = 0 THEN NULL ELSE now() + make_interval(months => $2) END,
		        $3, $4)
		ON CONFLICT (user_id) WHERE status = 'active'
		DO UPDATE SET expires_at = CASE
		                  WHEN $2 = 0 THEN NULL
		                  WHEN subscriptions.expires_at IS NULL THEN NULL
		                  ELSE subscriptions.expires_at + make_interval(months => $2)
		              END,
		              plan = CASE
		                  WHEN $2 = 0 OR subscriptions.expires_at IS NULL THEN 'lifetime'
		                  ELSE EXCLUDED.plan
		              END,
		              -- Also the provider: extending a subscription left it naming
		              -- whoever was paid the first time, so a renewal through a real
		              -- provider still reported the one before it.
		              provider = EXCLUDED.provider,
		              provider_ref = EXCLUDED.provider_ref,
		              updated_at = now()
		RETURNING id, status, plan, started_at, expires_at, provider`,
		userID, months, provider, ref, plan).
		Scan(&s.ID, &s.Status, &s.Plan, &s.StartedAt, &s.ExpiresAt, &s.Provider)
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `
		UPDATE payments SET status = 'paid', paid_at = now(), subscription_id = $2
		 WHERE id = $1`, paymentID, s.ID); err != nil {
		return nil, err
	}
	return &s, tx.Commit(ctx)
}

// SweepExpired marks subscriptions whose period has run out.
//
// Access never depended on this: HasAccess checks the expiry, not the status, so
// a row left 'active' past its date already grants nothing. What it fixes is the
// stored status, which anything reading the table directly — a report, a support
// query, a future dunning job — would otherwise read as a lie.
//
// It also frees the partial unique index on active subscriptions, so a returning
// customer opens a fresh row instead of extending a dead one.
func (r *BillingRepo) SweepExpired(ctx context.Context) (int64, error) {
	ct, err := r.pool.Exec(ctx, `
		UPDATE subscriptions SET status = 'expired', updated_at = now()
		 WHERE status = 'active' AND expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return ct.RowsAffected(), nil
}

// RecordLaunch notes that a sandbox was started, for the free tier's weekly
// allowance. It is called after the sandbox is up, so a launch that failed does
// not spend an allowance the student never got the benefit of.
func (r *BillingRepo) RecordLaunch(ctx context.Context, userID int, key string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sandbox_launches (user_id, sandbox_key) VALUES ($1, $2)`, userID, key)
	return err
}

// CountLaunches counts the sandboxes a user started since a moment — a rolling
// window, not a calendar week: "10 per week" resetting every Monday means a
// student who joins on Sunday gets one day's worth.
func (r *BillingRepo) CountLaunches(ctx context.Context, userID int, since time.Time) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM sandbox_launches WHERE user_id = $1 AND created_at >= $2`,
		userID, since).Scan(&n)
	return n, err
}

// OldestLaunchSince is when the earliest launch still inside the window
// happened, so the caller can say when the next allowance frees up. The second
// result is false when nothing is inside the window.
func (r *BillingRepo) OldestLaunchSince(ctx context.Context, userID int, since time.Time) (time.Time, bool, error) {
	var t time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT created_at FROM sandbox_launches
		 WHERE user_id = $1 AND created_at >= $2
		 ORDER BY created_at ASC LIMIT 1`, userID, since).Scan(&t)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, err
	}
	return t, true, nil
}

// SetCourseTier moves a course between free and subscription-only.
func (r *BillingRepo) SetCourseTier(ctx context.Context, moduleID int, tier string) error {
	if tier != "free" && tier != "subscription" {
		return ErrBadTier
	}
	ct, err := r.pool.Exec(ctx, `UPDATE modules SET access_tier = $2 WHERE id = $1`, moduleID, tier)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrCourseNotFound
	}
	return nil
}

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrCourseNotFound  = errors.New("course not found")
	ErrBadTier         = errors.New("access tier must be free or subscription")
	ErrAmountMismatch  = errors.New("the amount paid is not the amount owed")
)
