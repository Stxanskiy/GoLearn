package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/billing"
)

func TestListPlansIsPublic(t *testing.T) {
	users, c := storefront()
	h := newTestAPIWith(t, users, c)

	// No cookie: the pricing page has to render for someone deciding whether to
	// sign up at all.
	w := do(h, http.MethodGet, "/billing/plans", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", w.Code, w.Body)
	}
	plans := decode[[]apigen.Plan](t, w)
	if len(plans) != 3 {
		t.Fatalf("got %d plans, want 3: %+v", len(plans), plans)
	}

	byID := map[string]apigen.Plan{}
	for _, p := range plans {
		byID[string(p.ID)] = p
	}
	if free := byID["free"]; free.AmountMinor != 0 || free.LaunchesPerWeek != billing.FreeLaunchesPerWeek {
		t.Errorf("free plan = %+v", free)
	}
	if month := byID["month"]; month.AmountMinor <= 0 || month.Months != 1 || month.Lifetime {
		t.Errorf("month plan = %+v", month)
	}
	// The lifetime plan is the one a bug would make free or make expire.
	life := byID["lifetime"]
	if life.AmountMinor <= 0 || !life.Lifetime || life.Months != 0 {
		t.Errorf("lifetime plan = %+v", life)
	}
	if life.LaunchesPerWeek != 0 {
		t.Errorf("a paid plan must not be capped: %+v", life)
	}
}

// Checkout decides what to charge, so an unknown or unbuyable plan has to be
// refused rather than falling back to something.
func TestCheckoutRefusesPlansThatCannotBeBought(t *testing.T) {
	users, c := storefront()
	h := newTestAPIWith(t, users, c)

	for _, plan := range []string{`"nonsense"`, `"free"`} {
		w := do(h, http.MethodPost, "/billing/checkout", `{"plan":`+plan+`}`, withCookie(studentToken))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("plan %s: status = %d, want 400 (body %s)", plan, w.Code, w.Body)
		}
		if got := errorCode(t, w); got != "validation_failed" {
			t.Errorf("plan %s: code = %q", plan, got)
		}
	}
}

func TestSubscriptionReportsTheFreeAllowance(t *testing.T) {
	users, c := storefront()
	h := newTestAPIWith(t, users, c)

	sub := decode[apigen.Subscription](t, do(h, http.MethodGet, "/me/subscription", "", withCookie(studentToken)))
	if sub.Active || string(sub.Status) != "none" {
		t.Fatalf("fresh student = %+v", sub)
	}
	if sub.Quota == nil {
		t.Fatal("no quota reported; the client has nothing to show the free tier")
	}
	if sub.Quota.Limit != billing.FreeLaunchesPerWeek || sub.Quota.Left != billing.FreeLaunchesPerWeek {
		t.Fatalf("quota = %+v", *sub.Quota)
	}
	if sub.Quota.ResetsAt != nil {
		t.Errorf("nothing used, so nothing resets: %v", *sub.Quota.ResetsAt)
	}

	// Spend the allowance and it must count down, not go negative.
	c.launches[1] = billing.FreeLaunchesPerWeek + 3
	sub = decode[apigen.Subscription](t, do(h, http.MethodGet, "/me/subscription", "", withCookie(studentToken)))
	if sub.Quota.Left != 0 {
		t.Errorf("left = %d, want 0", sub.Quota.Left)
	}
	if sub.Quota.ResetsAt == nil {
		t.Error("with launches spent the student must be told when one comes back")
	}
}

// A subscriber has no sandbox cap, and says so with limit 0 rather than by
// omitting the object — the client renders one shape either way.
func TestSubscriberHasNoLaunchCap(t *testing.T) {
	users, c := storefront()
	c.subscribed[1] = true
	h := newTestAPIWith(t, users, c)

	sub := decode[apigen.Subscription](t, do(h, http.MethodGet, "/me/subscription", "", withCookie(studentToken)))
	if sub.Quota == nil || sub.Quota.Limit != 0 {
		t.Fatalf("quota = %+v", sub.Quota)
	}
}

// The return pages must never claim a payment succeeded on a signature they
// could not check. Without Robokassa configured nothing verifies, so the
// student is told it is pending rather than paid.
func TestRobokassaReturnDoesNotClaimAnUncheckedPayment(t *testing.T) {
	users, c := storefront()
	h := newTestAPIWith(t, users, c)

	w := do(h, http.MethodGet, "/billing/robokassa/success?InvId=42&OutSum=4000.00&SignatureValue=deadbeef", "")
	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", w.Code)
	}
	loc := w.Header().Get("Location")
	if !strings.Contains(loc, "status=pending") {
		t.Errorf("an unverified return claimed %q", loc)
	}
	if !strings.Contains(loc, "invoice=42") {
		t.Errorf("the invoice must survive the redirect: %q", loc)
	}

	// The fail page needs no signature; it grants nothing and says nothing.
	w = do(h, http.MethodGet, "/billing/robokassa/fail?InvId=42", "")
	if w.Code != http.StatusFound || !strings.Contains(w.Header().Get("Location"), "status=fail") {
		t.Errorf("fail return = %d %q", w.Code, w.Header().Get("Location"))
	}
}

// The spec says the checkout body is optional, and it has to actually be
// optional: a caller written before plans existed sends nothing, and must
// still get the monthly plan rather than a 400.
func TestCheckoutWithoutABodyBuysTheMonth(t *testing.T) {
	users, c := storefront()
	h := newTestAPIWith(t, users, c)

	// The fake store refuses to open a payment, so the furthest this can get is
	// the 500 from that — which is proof it passed validation. A rejected plan
	// would have stopped at 400.
	w := do(h, http.MethodPost, "/billing/checkout", "", withCookie(studentToken))
	if w.Code == http.StatusBadRequest || w.Code == http.StatusUnsupportedMediaType {
		t.Fatalf("an empty body was refused: %d %s", w.Code, w.Body)
	}
}
