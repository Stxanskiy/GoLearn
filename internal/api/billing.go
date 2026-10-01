package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/backendraz/golearn/internal/billing"
	"github.com/backendraz/golearn/internal/repository"
)

// billingStore is the subscription side of the database.
type billingStore interface {
	Current(ctx context.Context, userID int) (*repository.Subscription, error)
	HasAccess(ctx context.Context, userID int) (bool, error)
	StartPayment(ctx context.Context, userID, months int, amountMinor int64, currency, provider, ref string) (*repository.Payment, error)
	SetProviderRef(ctx context.Context, paymentID int, ref string) error
	Confirm(ctx context.Context, provider, ref string) (*repository.Subscription, error)
	SetCourseTier(ctx context.Context, moduleID int, tier string) error
}

// Prices live here until a real provider and a plan table exist. One month, one
// price: the product is a platform subscription, not a catalogue of tiers.
// providerRobokassa names the provider in the payments table; the stub keeps
// its own name so the two can never confirm each other's rows.
const providerRobokassa = "robokassa"

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

// startCheckout opens a payment. With Robokassa configured the student is sent
// to its page; without it the old stub stays, so development needs no merchant
// account.
func (a *API) startCheckout(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r.Context())

	if a.Robokassa.Configured() {
		// The payment row is created first: Robokassa's invoice number has to be
		// something we can look up when the callback arrives, and the row's own id
		// is the only identifier that is unique and ours.
		p, err := a.Billing.StartPayment(r.Context(), user.ID, planMonths, planAmountMinor, planCurrency, providerRobokassa, "")
		if err != nil {
			a.internalError(w, "start payment", err)
			return
		}
		ref := strconv.Itoa(p.ID)
		if err := a.Billing.SetProviderRef(r.Context(), p.ID, ref); err != nil {
			a.internalError(w, "set provider ref", err)
			return
		}
		shp := url.Values{}
		shp.Set("Shp_user", strconv.Itoa(user.ID))
		link, err := a.Robokassa.PayLink(int64(p.ID), p.AmountMinor, "Подписка TOT на "+strconv.Itoa(p.Months)+" мес.", shp)
		if err != nil {
			a.internalError(w, "robokassa link", err)
			return
		}
		writeJSON(w, http.StatusCreated, checkoutResponse{
			PaymentID:   p.ID,
			ProviderRef: ref,
			AmountMinor: p.AmountMinor,
			Currency:    p.Currency,
			Months:      p.Months,
			ConfirmURL:  link,
		})
		return
	}

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

// robokassaResult is the server-to-server callback. It is the only thing that
// may grant a subscription: the student's browser never sees it, so unlike the
// Success redirect it cannot be replayed or forged by them.
//
// Robokassa expects the body "OK<InvId>" and retries until it gets it, so a
// callback that arrives twice must be harmless — Confirm is idempotent in the
// store, and a second call cannot buy a second month.
func (a *API) robokassaResult(w http.ResponseWriter, r *http.Request) {
	// Plain text throughout: Robokassa is not a JSON client. It looks at the
	// status and at whether the body is OK<InvId>, so an error carries no more
	// than a line of text.
	fail := func(status int, msg string) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintln(w, msg)
	}
	if !a.Robokassa.Configured() {
		fail(http.StatusNotFound, "robokassa is not configured")
		return
	}
	if err := r.ParseForm(); err != nil {
		fail(http.StatusBadRequest, "malformed form")
		return
	}
	// Robokassa sends these in the query for GET and the body for POST; ParseForm
	// merges both, so one read covers either.
	outSum := r.Form.Get("OutSum")
	invID, err := strconv.ParseInt(r.Form.Get("InvId"), 10, 64)
	if err != nil || invID <= 0 {
		fail(http.StatusBadRequest, "bad InvId")
		return
	}
	if !a.Robokassa.VerifyResult(outSum, invID, r.Form.Get("SignatureValue"), billing.ShpValues(r.Form)) {
		// Deliberately terse: an attacker probing signatures learns nothing from
		// this, and the detail goes to the log instead.
		a.log.Warn("robokassa: bad callback signature", "inv", invID, "sum", outSum)
		fail(http.StatusForbidden, "bad signature")
		return
	}
	// The amount is signed, but a signature only proves Robokassa sent it — not
	// that it is the amount we asked for. Check it against the plan before
	// granting anything.
	if outSum != billing.Amount(planAmountMinor) {
		a.log.Warn("robokassa: unexpected amount", "inv", invID, "sum", outSum)
		fail(http.StatusBadRequest, "unexpected amount")
		return
	}

	if _, err := a.Billing.Confirm(r.Context(), providerRobokassa, strconv.FormatInt(invID, 10)); err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			// 500, not 404: Robokassa retries a non-200 for hours, and this is the
			// one case where retrying is what we want. A signature verified against
			// our own password means the invoice is ours, so a missing row is our
			// problem — most likely the callback overtook the row's commit.
			a.log.Error("robokassa: confirmed payment has no row", "inv", invID)
			fail(http.StatusInternalServerError, "payment not found, retry")
			return
		}
		a.log.Error("robokassa: confirm failed", "inv", invID, "error", err)
		fail(http.StatusInternalServerError, "confirm failed, retry")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "OK%d", invID)
}

// confirmCheckout stands in for the provider's webhook. It is deliberately
// idempotent in the store, so calling it twice cannot buy a second month.
func (a *API) confirmCheckout(w http.ResponseWriter, r *http.Request) {
	// With a real provider this endpoint must not exist: a signed-in student can
	// call it with their own reference, and it would hand them a subscription for
	// nothing. It stays only while the stub is the provider.
	if a.Robokassa.Configured() {
		writeError(w, http.StatusNotFound, codeNotFound, "route not found")
		return
	}
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
