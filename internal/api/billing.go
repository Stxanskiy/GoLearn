package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/backendraz/golearn/internal/repository"
)

// billingStore is the subscription side of the database.
type billingStore interface {
	Current(ctx context.Context, userID int) (*repository.Subscription, error)
	HasAccess(ctx context.Context, userID int) (bool, error)
	StartPayment(ctx context.Context, userID, months int, amountMinor int64, currency, provider, ref string) (*repository.Payment, error)
	Confirm(ctx context.Context, provider, ref string) (*repository.Subscription, error)
	SetCourseTier(ctx context.Context, moduleID int, tier string) error
}

// Prices live here until a real provider and a plan table exist. One month, one
// price: the product is a platform subscription, not a catalogue of tiers.
const (
	planMonths      = 1
	planAmountMinor = 49000 // 490.00 RUB
	planCurrency    = "RUB"
)

type subscriptionResponse struct {
	Active    bool       `json:"active"`
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Provider  string     `json:"provider,omitempty"`
}

// getSubscription reports the caller's subscription. A user who never subscribed
// and one whose subscription lapsed are different answers, so "none" and
// "expired" are both reported rather than collapsed into a bare false.
func (a *API) getSubscription(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	sub, err := a.Billing.Current(r.Context(), user.ID)
	if err != nil {
		a.internalError(w, "load subscription", err)
		return
	}
	if sub == nil {
		writeJSON(w, http.StatusOK, subscriptionResponse{Status: "none"})
		return
	}
	writeJSON(w, http.StatusOK, subscriptionResponse{
		Active:    sub.IsActive(),
		Status:    sub.Status,
		ExpiresAt: &sub.ExpiresAt,
		Provider:  sub.Provider,
	})
}

type checkoutResponse struct {
	PaymentID   int    `json:"payment_id"`
	ProviderRef string `json:"provider_ref"`
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
	Months      int    `json:"months"`
	// ConfirmURL is where the caller sends the user to pay. With the stub there
	// is nothing to pay, so it points back at our own confirm endpoint.
	ConfirmURL string `json:"confirm_url"`
}

// startCheckout opens a payment. There is no provider yet: the stub records a
// pending payment and hands back the reference that confirms it, so the whole
// flow — pending, confirm, access — is exercisable before any money moves.
func (a *API) startCheckout(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())
	ref := fmt.Sprintf("stub-%d-%d", user.ID, time.Now().UnixNano())
	p, err := a.Billing.StartPayment(r.Context(), user.ID, planMonths, planAmountMinor, planCurrency, "stub", ref)
	if err != nil {
		a.internalError(w, "start payment", err)
		return
	}
	writeJSON(w, http.StatusCreated, checkoutResponse{
		PaymentID:   p.ID,
		ProviderRef: p.ProviderRef,
		AmountMinor: p.AmountMinor,
		Currency:    p.Currency,
		Months:      p.Months,
		ConfirmURL:  "/api/v1/billing/confirm?ref=" + p.ProviderRef,
	})
}

// confirmCheckout stands in for the provider's webhook. It is deliberately
// idempotent in the store, so calling it twice cannot buy a second month.
func (a *API) confirmCheckout(w http.ResponseWriter, r *http.Request) {
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		writeError(w, http.StatusBadRequest, codeValidationFailed, "ref is required")
		return
	}
	sub, err := a.Billing.Confirm(r.Context(), "stub", ref)
	if errors.Is(err, repository.ErrPaymentNotFound) {
		writeError(w, http.StatusNotFound, codeNotFound, "payment not found")
		return
	}
	if err != nil {
		a.internalError(w, "confirm payment", err)
		return
	}
	writeJSON(w, http.StatusOK, subscriptionResponse{
		Active:    sub.IsActive(),
		Status:    sub.Status,
		ExpiresAt: &sub.ExpiresAt,
		Provider:  sub.Provider,
	})
}

// requireCourseAccess gates content that the subscription covers.
//
// Admins pass so the catalogue stays inspectable, and free courses are never
// gated — locking the free tier by accident is the failure that loses trust
// fastest. 402 rather than 403: the caller is not forbidden, they have not paid.
func (a *API) requireCourseAccess(w http.ResponseWriter, r *http.Request, tier string) bool {
	if tier != "subscription" {
		return true
	}
	user := userFrom(r.Context())
	if user.IsAdmin() {
		return true
	}
	ok, err := a.Billing.HasAccess(r.Context(), user.ID)
	if err != nil {
		a.internalError(w, "check subscription", err)
		return false
	}
	if !ok {
		writeError(w, http.StatusPaymentRequired, codeSubscriptionNeeded,
			"this course is part of the subscription")
		return false
	}
	return true
}

type accessTierRequest struct {
	AccessTier string `json:"access_tier"`
}

// setCourseAccessTier moves a course between free and subscription-only. Admin
// only: this is the control that decides what the subscription is worth.
func (a *API) setCourseAccessTier(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "courseId", "course not found")
	if !ok {
		return
	}
	var req accessTierRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	err := a.Billing.SetCourseTier(r.Context(), id, req.AccessTier)
	switch {
	case errors.Is(err, repository.ErrBadTier):
		writeError(w, http.StatusBadRequest, codeValidationFailed, err.Error())
	case errors.Is(err, repository.ErrCourseNotFound):
		writeError(w, http.StatusNotFound, codeNotFound, "course not found")
	case err != nil:
		a.internalError(w, "set access tier", err)
	default:
		writeJSON(w, http.StatusOK, accessTierRequest{AccessTier: req.AccessTier})
	}
}
