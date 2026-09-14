package api

import (
	"net/http"

	"github.com/backendraz/golearn/internal/lab"
)

// requireSandbox writes 503 and returns false when the sandbox is not configured.
func (a *API) requireSandbox(w http.ResponseWriter) bool {
	if !a.Sandbox.Enabled() {
		writeError(w, http.StatusServiceUnavailable, codeSandboxDisabled, "sandbox is not configured")
		return false
	}
	return true
}

func (a *API) getGitTrainerSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.sandboxSession(userFrom(r.Context()).ID, lab.GitTrainerKey))
}

func (a *API) resetGitTrainer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := a.Sandbox.Reset(ctx, userFrom(ctx).ID, lab.GitTrainerKey); err != nil {
		a.internalError(w, "git trainer: reset", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) openGitTrainerTerminal(w http.ResponseWriter, r *http.Request) {
	if a.requireSandbox(w) {
		a.serveTerminal(w, r, lab.GitTrainerKey, lab.GitTrainerImage, lab.GitTrainerSetup)
	}
}

func (a *API) getGitTrainerGraph(w http.ResponseWriter, r *http.Request) {
	if a.requireSandbox(w) {
		a.writeGitGraph(w, r, lab.GitTrainerKey, lab.GitTrainerImage, lab.GitTrainerSetup, lab.GitTrainerRepo)
	}
}
