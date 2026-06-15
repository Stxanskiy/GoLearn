package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type shellExecReq struct {
	Command string `json:"command"`
}

// ShellExec handles POST /api/shell/{taskID}/exec — runs one command in the
// student's per-task sandbox container and returns combined output.
func (h *Handler) ShellExec(w http.ResponseWriter, r *http.Request) {
	if !h.shell.Enabled() {
		writeJSON(w, map[string]any{"output": "Песочница недоступна (sandbox не настроен на сервере)."})
		return
	}
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "bad task id", 400)
		return
	}
	var req shellExecReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	task, err := h.lessonRepo.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, "task not found", 404)
		return
	}
	out, err := h.shell.Exec(r.Context(), user.ID, taskID, task.SandboxImage, task.SetupScript, req.Command)
	if err != nil {
		h.log.Error("shell exec", "error", err)
		writeJSON(w, map[string]any{"output": "Ошибка песочницы: " + err.Error()})
		return
	}
	writeJSON(w, map[string]any{"output": out})
}

// ShellCheck handles POST /api/shell/{taskID}/check — runs the task's check
// script; on success records the submission and advances lesson progress.
func (h *Handler) ShellCheck(w http.ResponseWriter, r *http.Request) {
	if !h.shell.Enabled() {
		writeJSON(w, map[string]any{"passed": false, "output": "Песочница недоступна."})
		return
	}
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "bad task id", 400)
		return
	}
	task, err := h.lessonRepo.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, "task not found", 404)
		return
	}
	passed, out, err := h.shell.Check(r.Context(), user.ID, taskID, task.SandboxImage, task.SetupScript, task.CheckScript)
	if err != nil {
		h.log.Error("shell check", "error", err)
		writeJSON(w, map[string]any{"passed": false, "output": "Ошибка проверки: " + err.Error()})
		return
	}
	if passed {
		_ = h.submissionRepo.Save(r.Context(), taskID, "[shell]", out, "", true)
		if allDone, e := h.submissionRepo.AllLessonTasksPassed(r.Context(), task.LessonID); e == nil && allDone {
			_ = h.progressRepo.Upsert(r.Context(), task.LessonID, "completed")
		} else {
			_ = h.progressRepo.Upsert(r.Context(), task.LessonID, "in_progress")
		}
	}
	writeJSON(w, map[string]any{"passed": passed, "output": out})
}

// ShellReset handles POST /api/shell/{taskID}/reset — wipes the sandbox session.
func (h *Handler) ShellReset(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	taskID, err := strconv.Atoi(chi.URLParam(r, "taskID"))
	if err != nil {
		http.Error(w, "bad task id", 400)
		return
	}
	_ = h.shell.Reset(r.Context(), user.ID, taskID)
	writeJSON(w, map[string]any{"ok": true})
}
