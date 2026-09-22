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
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Provider  string    `json:"provider"`
}

func (s *Subscription) IsActive() bool {
	return s != nil && s.Status == "active" && s.ExpiresAt.After(time.Now())
}

// Payment is one attempt to buy a period, successful or not.
type Payment struct {
	ID          int        `json:"id"`
	Status      string     `json:"status"`
	AmountMinor int64      `json:"amount_minor"`
	Currency    string     `json:"currency"`
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
	err := r.pool.QueryRow(ctx, `
		SELECT id, status, started_at, expires_at, provider
		  FROM subscriptions
		 WHERE user_id = $1
		 ORDER BY expires_at DESC
		 LIMIT 1`, userID).Scan(&s.ID, &s.Status, &s.StartedAt, &s.ExpiresAt, &s.Provider)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// A row can outlive its expiry if nothing swept it; report the truth.
	if s.Status == "active" && !s.ExpiresAt.After(time.Now()) {
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
func (r *BillingRepo) StartPayment(ctx context.Context, userID, months int, amountMinor int64, currency, provider, ref string) (*Payment, error) {
	var p Payment
	err := r.pool.QueryRow(ctx, `
		INSERT INTO payments (user_id, status, amount_minor, currency, months, provider, provider_ref)
		VALUES ($1, 'pending', $2, $3, $4, $5, $6)
		RETURNING id, status, amount_minor, currency, months, provider, provider_ref, created_at, paid_at`,
		userID, amountMinor, currency, months, provider, ref).
		Scan(&p.ID, &p.Status, &p.AmountMinor, &p.Currency, &p.Months, &p.Provider, &p.ProviderRef, &p.CreatedAt, &p.PaidAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Confirm marks a payment paid and extends the subscription by its months.
//
// It is safe to call twice with the same reference: a payment already marked
// paid returns its existing subscription instead of granting another period,
// which is what keeps a replayed provider webhook from stacking free months.
func (r *BillingRepo) Confirm(ctx context.Context, provider, ref string) (*Subscription, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var paymentID, userID, months int
	var status string
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, months, status FROM payments
		 WHERE provider = $1 AND provider_ref = $2
		 FOR UPDATE`, provider, ref).Scan(&paymentID, &userID, &months, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	if status != "pending" {
		// Already handled; hand back what the user has rather than granting more.
		var s Subscription
		err = tx.QueryRow(ctx, `
			SELECT id, status, started_at, expires_at, provider FROM subscriptions
			 WHERE user_id = $1 ORDER BY expires_at DESC LIMIT 1`, userID).
			Scan(&s.ID, &s.Status, &s.StartedAt, &s.ExpiresAt, &s.Provider)
		if err != nil {
			return nil, err
		}
		return &s, tx.Commit(ctx)
	}

	// Extend an active subscription from its own expiry, so paying early never
	// costs the buyer the days they already hold.
	var s Subscription
	err = tx.QueryRow(ctx, `
		INSERT INTO subscriptions (user_id, status, started_at, expires_at, provider, provider_ref)
		VALUES ($1, 'active', now(), now() + make_interval(months => $2), $3, $4)
		ON CONFLICT (user_id) WHERE status = 'active'
		DO UPDATE SET expires_at = subscriptions.expires_at + make_interval(months => $2),
		              updated_at = now()
		RETURNING id, status, started_at, expires_at, provider`,
		userID, months, provider, ref).
		Scan(&s.ID, &s.Status, &s.StartedAt, &s.ExpiresAt, &s.Provider)
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
)
