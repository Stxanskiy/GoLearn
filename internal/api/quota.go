package api

import (
	"net/http"
	"time"

	"github.com/backendraz/golearn/internal/billing"
)

// requireLaunchAllowance gates starting a *new* sandbox against the free tier's
// weekly allowance. It answers true when the caller may go ahead, and has
// already written the error when it answers false.
//
// Only a new micro-VM is counted. Reopening a terminal on a sandbox that is
// already running costs nothing, and charging for it would punish a student for
// a dropped connection — the one thing they cannot control.
//
// This is where the subscription is actually worth something: a micro-VM is the
// most expensive thing on the host by a wide margin, so it is the resource the
// free tier is limited on rather than lessons read or quizzes answered.
func (a *API) requireLaunchAllowance(w http.ResponseWriter, r *http.Request, key string) bool {
	user := userFrom(r.Context())
	if user.IsAdmin() {
		return true
	}
	// Already running: not a launch.
	if _, live := a.Sandbox.Session(user.ID, key); live {
		return true
	}
	subscribed, err := a.Billing.HasAccess(r.Context(), user.ID)
	if err != nil {
		a.internalError(w, "check subscription", err)
		return false
	}
	if subscribed {
		return true
	}

	used, err := a.Billing.CountLaunches(r.Context(), user.ID, time.Now().Add(-freeWindow))
	if err != nil {
		a.internalError(w, "count sandbox launches", err)
		return false
	}
	if used < billing.FreeLaunchesPerWeek {
		return true
	}
	// 402, matching the subscription gate: the caller is not forbidden, they
	// have run out of what the free tier includes.
	writeError(w, http.StatusPaymentRequired, codeLaunchQuota,
		"the free plan covers 10 sandbox starts a week")
	return false
}

// noteLaunch records a sandbox that actually started.
//
// After the fact, not before: a launch that failed to boot has cost the student
// nothing they can use, and spending an allowance on it would make a bad day
// worse. The trade is that a burst of parallel requests can overshoot the
// allowance slightly, which is the cheaper mistake.
func (a *API) noteLaunch(r *http.Request, key string) {
	user := userFrom(r.Context())
	if user.IsAdmin() {
		return
	}
	if err := a.Billing.RecordLaunch(r.Context(), user.ID, key); err != nil {
		// Not fatal: the sandbox is up and the student should get to use it. The
		// cost of losing the row is one uncounted launch.
		a.log.Error("record sandbox launch", "user", user.ID, "key", key, "error", err)
	}
}
