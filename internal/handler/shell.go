package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type shellExecReq struct {
	Command string `json:"command"`
}

// labKey names the sandbox session shared by a whole lab. Every step of a
// lesson — the interactive terminal and each step's check — must run in the
// SAME container, otherwise a check looks at an empty filesystem and can never
// pass.
func labKey(lessonID int) string { return fmt.Sprintf("l%d", lessonID) }

// labSandbox resolves the image and combined setup script for a lesson's lab.
func (h *Handler) labSandbox(r *http.Request, lessonID int) (image, setup string) {
	image, setup, err := h.lessonRepo.LessonSandbox(r.Context(), lessonID)
	if err != nil {
		h.log.Error("lab sandbox", "lesson", lessonID, "error", err)
	}
	return image, setup
}

// ShellExec handles POST /api/shell/{taskID}/exec — runs one command in the
// lab's sandbox container and returns combined output.
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
	image, setup := h.labSandbox(r, task.LessonID)
	out, err := h.shell.Exec(r.Context(), user.ID, labKey(task.LessonID), image, setup, req.Command)
	if err != nil {
		h.log.Error("shell exec", "error", err)
		writeJSON(w, map[string]any{"output": "Ошибка песочницы: " + err.Error()})
		return
	}
	writeJSON(w, map[string]any{"output": out})
}

// ShellCheck handles POST /api/shell/{taskID}/check — runs the task's check
// script in the lab session and, on success, records the submission and
// advances lesson progress.
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
	if task.CheckScript == "" {
		writeJSON(w, map[string]any{"passed": false, "output": "У этого шага нет автопроверки."})
		return
	}
	image, setup := h.labSandbox(r, task.LessonID)
	passed, out, err := h.shell.Check(r.Context(), user.ID, labKey(task.LessonID), image, setup, task.CheckScript)
	if err != nil {
		h.log.Error("shell check", "error", err)
		writeJSON(w, map[string]any{"passed": false, "output": "Ошибка проверки: " + err.Error()})
		return
	}
	if passed {
		_ = h.submissionRepo.Save(r.Context(), user.ID, taskID, "[shell]", out, "", true)
		if allDone, e := h.submissionRepo.AllLessonTasksPassed(r.Context(), user.ID, task.LessonID); e == nil && allDone {
			_ = h.progressRepo.Upsert(r.Context(), user.ID, task.LessonID, "completed")
		} else {
			_ = h.progressRepo.Upsert(r.Context(), user.ID, task.LessonID, "in_progress")
		}
	}
	writeJSON(w, map[string]any{"passed": passed, "output": out})
}

// ShellStepDone handles POST /api/shell/{taskID}/done — marks a step without an
// auto-check as completed, so manual steps survive a page reload like checked
// ones do.
func (h *Handler) ShellStepDone(w http.ResponseWriter, r *http.Request) {
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
	_ = h.submissionRepo.Save(r.Context(), user.ID, taskID, "[manual]", "", "", true)
	if allDone, e := h.submissionRepo.AllLessonTasksPassed(r.Context(), user.ID, task.LessonID); e == nil && allDone {
		_ = h.progressRepo.Upsert(r.Context(), user.ID, task.LessonID, "completed")
	} else {
		_ = h.progressRepo.Upsert(r.Context(), user.ID, task.LessonID, "in_progress")
	}
	writeJSON(w, map[string]any{"ok": true})
}

// LabReset handles POST /api/lab/{lessonID}/reset — throws away the sandbox
// container so the next command starts from a clean environment. Progress is
// untouched.
func (h *Handler) LabReset(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	lessonID, err := strconv.Atoi(chi.URLParam(r, "lessonID"))
	if err != nil {
		http.Error(w, "bad lesson id", 400)
		return
	}
	_ = h.shell.Reset(r.Context(), user.ID, labKey(lessonID))
	writeJSON(w, map[string]any{"ok": true})
}

// LabRetry handles POST /api/lab/{lessonID}/retry — "Пройти заново": clears the
// user's solved steps and quiz score for the lesson AND recycles the sandbox,
// so the lab can be taken again from scratch.
func (h *Handler) LabRetry(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	lessonID, err := strconv.Atoi(chi.URLParam(r, "lessonID"))
	if err != nil {
		http.Error(w, "bad lesson id", 400)
		return
	}
	if err := h.submissionRepo.ResetLesson(r.Context(), user.ID, lessonID); err != nil {
		h.log.Error("lab retry: reset submissions", "error", err)
		http.Error(w, "reset failed", 500)
		return
	}
	if err := h.progressRepo.ResetLesson(r.Context(), user.ID, lessonID); err != nil {
		h.log.Error("lab retry: reset progress", "error", err)
		http.Error(w, "reset failed", 500)
		return
	}
	_ = h.shell.Reset(r.Context(), user.ID, labKey(lessonID))
	writeJSON(w, map[string]any{"ok": true})
}
