package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

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

// LabPreview handles GET /api/lab/{lessonID}/preview/{port}/* — a tiny reverse
// proxy that shows, inside an iframe, whatever HTTP server the student started
// in their sandbox (e.g. `python3 -m http.server` or a container they ran). The
// port is part of the path so relative asset URLs on the previewed page resolve
// back through this same proxy.
func (h *Handler) LabPreview(w http.ResponseWriter, r *http.Request) {
	if !h.shell.Enabled() {
		http.Error(w, "sandbox disabled", 503)
		return
	}
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
	port, _ := strconv.Atoi(chi.URLParam(r, "port"))
	if port <= 0 {
		port = 80
	}
	path := "/" + chi.URLParam(r, "*")
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	image, setup := h.labSandbox(r, lessonID)
	body, ct, status, err := h.shell.Preview(r.Context(), user.ID, labKey(lessonID), image, setup, port, path)
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(previewErrPage(port, err)))
		return
	}
	if strings.HasPrefix(ct, "text/html") {
		body = injectBase(body, fmt.Sprintf("/api/lab/%d/preview/%d/", lessonID, port))
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// injectBase inserts a <base> tag so relative links on the previewed page point
// back through the proxy prefix instead of the LMS origin.
func injectBase(body []byte, base string) []byte {
	tag := []byte(`<base href="` + base + `">`)
	lower := strings.ToLower(string(body))
	if i := strings.Index(lower, "<head>"); i >= 0 {
		i += len("<head>")
		return append(append(append([]byte{}, body[:i]...), tag...), body[i:]...)
	}
	return append(tag, body...)
}

// fsEditRoot is what the in-lab editor may browse and edit.
const fsEditRoot = "/root"

// jailPath cleans p and confirms it stays under /root — the editor must not read
// or write anywhere else in the container.
func jailPath(p string) (string, bool) {
	if strings.TrimSpace(p) == "" {
		p = fsEditRoot
	}
	if !strings.HasPrefix(p, "/") {
		p = fsEditRoot + "/" + p
	}
	p = path.Clean(p)
	if p != fsEditRoot && !strings.HasPrefix(p, fsEditRoot+"/") {
		return "", false
	}
	return p, true
}

// fsLessonID resolves the lesson id and jailed path shared by the FS handlers.
func (h *Handler) fsCommon(w http.ResponseWriter, r *http.Request) (userID, lessonID int, p, image, setup string, ok bool) {
	if !h.shell.Enabled() {
		http.Error(w, "sandbox disabled", 503)
		return
	}
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
	p, jok := jailPath(r.URL.Query().Get("path"))
	if !jok {
		http.Error(w, "path outside /root", 400)
		return
	}
	image, setup = h.labSandbox(r, lessonID)
	return user.ID, lessonID, p, image, setup, true
}

// LabFSList — GET /api/lab/{lessonID}/fs/list?path=/root — file tree children.
func (h *Handler) LabFSList(w http.ResponseWriter, r *http.Request) {
	userID, lessonID, p, image, setup, ok := h.fsCommon(w, r)
	if !ok {
		return
	}
	entries, err := h.shell.FSList(r.Context(), userID, labKey(lessonID), image, setup, p)
	if err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"path": p, "entries": entries})
}

// LabFSRead — GET /api/lab/{lessonID}/fs/read?path=/root/project/app.yml.
func (h *Handler) LabFSRead(w http.ResponseWriter, r *http.Request) {
	userID, lessonID, p, image, setup, ok := h.fsCommon(w, r)
	if !ok {
		return
	}
	data, err := h.shell.FSRead(r.Context(), userID, labKey(lessonID), image, setup, p)
	if err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"path": p, "content": string(data)})
}

// LabFSWrite — PUT /api/lab/{lessonID}/fs/write?path=/root/project/app.yml, body = content.
func (h *Handler) LabFSWrite(w http.ResponseWriter, r *http.Request) {
	userID, lessonID, p, image, setup, ok := h.fsCommon(w, r)
	if !ok {
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20)) // 2 MB cap
	if err != nil {
		http.Error(w, "read body", 400)
		return
	}
	if err := h.shell.FSWrite(r.Context(), userID, labKey(lessonID), image, setup, p, body); err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": p})
}

func previewErrPage(port int, err error) string {
	reason := "на этом порту пока никто не отвечает"
	if strings.Contains(err.Error(), "curl-missing") {
		reason = "в этой песочнице нет curl"
	}
	return fmt.Sprintf(`<!doctype html><html lang="ru"><head><meta charset="utf-8">`+
		`<style>body{margin:0;font-family:system-ui,sans-serif;background:#22201b;color:#d6c9b6;`+
		`display:flex;align-items:center;justify-content:center;height:100vh;text-align:center}`+
		`.c{max-width:420px;padding:24px}h2{margin:0 0 8px;font-size:1.1em;color:#e8dcc8}`+
		`code{background:#000;padding:2px 7px;border-radius:5px;color:#7fd6c2}p{line-height:1.6;font-size:.9em;color:#a89a85}</style>`+
		`</head><body><div class="c"><h2>Пока нечего показать</h2>`+
		`<p>Web Preview обращается к <code>127.0.0.1:%d</code> внутри песочницы — %s.<br><br>`+
		`Запусти сервер в терминале (например <code>python3 -m http.server %d</code> или свой контейнер), затем нажми «Обновить».</p></div></body></html>`,
		port, reason, port)
}
