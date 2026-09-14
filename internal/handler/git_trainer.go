package handler

import (
	"encoding/json"
	"net/http"

	"github.com/backendraz/golearn/internal/lab"
)

// GitTrainerPage serves the git trainer page (real git runtime in the sandbox).
func (h *Handler) GitTrainerPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PageTitle string
		SandboxOK bool
	}{
		PageTitle: "Git Trainer",
		SandboxOK: h.shell.Enabled(),
	}
	h.render(w, "git_trainer", &data)
}

// GitExec runs a real git command in the user's git sandbox repo.
func (h *Handler) GitExec(w http.ResponseWriter, r *http.Request) {
	if !h.shell.Enabled() {
		writeJSON(w, map[string]any{"output": "Песочница недоступна на сервере."})
		return
	}
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	out, err := h.shell.Exec(r.Context(), user.ID, lab.GitTrainerKey, lab.GitTrainerImage, lab.GitTrainerSetup, req.Command)
	if err != nil {
		h.log.Error("git exec", "error", err)
		writeJSON(w, map[string]any{"output": "Ошибка песочницы: " + err.Error()})
		return
	}
	writeJSON(w, map[string]any{"output": out})
}

// GitReset wipes the user's git sandbox so a fresh repo is created next time.
func (h *Handler) GitReset(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	_ = h.shell.Reset(r.Context(), user.ID, lab.GitTrainerKey)
	writeJSON(w, map[string]any{"ok": true})
}
