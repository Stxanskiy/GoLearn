package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/content"
	"github.com/backendraz/golearn/internal/lab"
	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/runner"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// Per-user limits for sandbox-heavy actions.
var (
	terminalLimiter = auth.NewLimiter(20, time.Minute)
	checkLimiter    = auth.NewLimiter(30, time.Minute)
	runLimiter      = auth.NewLimiter(30, time.Minute)
)

// labRef is a visible lesson with its tasks and sandbox configuration.
type labRef struct {
	lessonRef
	tasks []model.Task
	spec  runner.Spec
}

func (l labRef) key() string { return lab.Key(l.lesson.ID) }

// labByID loads a visible lesson that has tasks, writing 404/500 itself when it returns false.
func (a *API) labByID(w http.ResponseWriter, r *http.Request) (labRef, bool) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return labRef{}, false
	}
	ctx := r.Context()
	tasks, err := a.Lessons.GetTasks(ctx, ref.lesson.ID)
	if err != nil {
		a.internalError(w, "lab: tasks", err)
		return labRef{}, false
	}
	if len(tasks) == 0 {
		writeError(w, http.StatusNotFound, codeNotFound, "lesson has no lab")
		return labRef{}, false
	}
	spec, err := a.Lessons.LessonSandbox(ctx, ref.lesson.ID)
	if err != nil {
		a.internalError(w, "lab: sandbox config", err)
		return labRef{}, false
	}
	return labRef{lessonRef: ref, tasks: tasks, spec: spec}, true
}

// sandboxLab is labByID for endpoints that need the sandbox enabled.
func (a *API) sandboxLab(w http.ResponseWriter, r *http.Request) (labRef, bool) {
	ref, ok := a.labByID(w, r)
	if !ok || !a.requireSandbox(w) {
		return labRef{}, false
	}
	return ref, true
}

func (a *API) sandboxSession(userID int, key string) apigen.SandboxSession {
	out := apigen.SandboxSession{Enabled: a.Sandbox.Enabled()}
	if info, ok := a.Sandbox.Session(userID, key); ok {
		out.Running = true
		out.StartedAt, out.ExpiresAt = &info.Started, &info.ExpiresAt
	}
	return out
}

func (a *API) getLab(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.labByID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	uid := userFrom(ctx).ID
	l, m := ref.lesson, ref.module

	passed, err := a.Submissions.PassedTaskIDs(ctx, uid, l.ID)
	if err != nil {
		a.internalError(w, "lab: passed tasks", err)
		return
	}
	siblings, err := a.Lessons.GetByModule(ctx, m.ID)
	if err != nil {
		a.internalError(w, "lab: siblings", err)
		return
	}
	up, err := a.userProgress(ctx, uid)
	if err != nil {
		a.internalError(w, "lab: progress", err)
		return
	}
	quiz, err := a.lessonQuiz(ctx, uid, l.ID)
	if err != nil {
		a.internalError(w, "lab: quiz", err)
		return
	}
	attempts, err := a.QuizAttempts.Count(ctx, uid, l.ID)
	if err != nil {
		a.internalError(w, "lab: attempts", err)
		return
	}

	out := apigen.Lab{
		LessonID:   l.ID,
		Slug:       l.Slug,
		Title:      l.Title,
		Mode:       apigen.LabMode("code"),
		IsGitLab:   lab.IsGitLab(*m),
		TheoryHTML: content.Render(l.Format, l.Content),
		Tasks:      make([]apigen.LabTask, 0, len(ref.tasks)),
		Quiz:       quiz,
		Nav:        lessonNav(*m, *l, siblings, up),
		Progress:   lessonProgress(*l, up, attempts),
		Sandbox:    a.sandboxSession(uid, ref.key()),
	}
	if ref.tasks[0].Kind == "shell" {
		out.Mode = "shell"
	}
	for _, t := range ref.tasks {
		lt := apigen.LabTask{
			ID:              t.ID,
			Step:            t.OrderNum,
			Title:           t.Title,
			Kind:            apigen.TaskKind(t.Kind),
			DescriptionHTML: content.Render(t.Format, t.Description),
			Glossary:        make([]apigen.GlossaryItem, 0, len(t.Glossary)),
			CheckMode:       apigen.LabTaskCheckMode(checkMode(t)),
			Passed:          passed[t.ID],
		}
		if t.Hints != "" {
			hints := content.Render(t.Format, t.Hints)
			lt.HintsHTML = &hints
		}
		if runner.SupportsLang(t.Kind) && t.StarterCode != "" {
			code := t.StarterCode
			lt.StarterCode = &code
		}
		for _, g := range t.Glossary {
			lt.Glossary = append(lt.Glossary, apigen.GlossaryItem{Term: g.Term, Definition: g.Definition})
		}
		out.Tasks = append(out.Tasks, lt)
	}
	writeJSON(w, http.StatusOK, out)
}

// checkMode tells how a task is completed: auto (check script), tests (Go test cases) or manual.
func checkMode(t model.Task) string {
	switch {
	case t.Kind == "shell" && t.CheckScript != "":
		return "auto"
	case runner.SupportsLang(t.Kind) && len(t.TestCases) > 0:
		return "tests"
	default:
		return "manual"
	}
}

func (a *API) getLabSession(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.labByID(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, a.sandboxSession(userFrom(r.Context()).ID, ref.key()))
}

func (a *API) stopLabSession(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.labByID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	if err := a.Sandbox.Reset(ctx, userFrom(ctx).ID, ref.key()); err != nil {
		a.internalError(w, "lab: stop session", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) retryLab(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.labByID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	uid := userFrom(ctx).ID
	steps := []struct {
		op  string
		run func() error
	}{
		{"submissions", func() error { return a.Submissions.ResetLesson(ctx, uid, ref.lesson.ID) }},
		{"quiz answers", func() error { return a.QuizAnswers.ResetLesson(ctx, uid, ref.lesson.ID) }},
		{"progress", func() error { return a.Progress.ResetLesson(ctx, uid, ref.lesson.ID) }},
		{"sandbox", func() error { return a.Sandbox.Reset(ctx, uid, ref.key()) }},
	}
	for _, s := range steps {
		if err := s.run(); err != nil {
			a.internalError(w, "lab retry: "+s.op, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) openLabTerminal(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.sandboxLab(w, r)
	if !ok {
		return
	}
	a.serveTerminal(w, r, ref.key(), ref.spec)
}

// serveTerminal upgrades to a WebSocket and bridges it to a PTY in the user's sandbox session.
func (a *API) serveTerminal(w http.ResponseWriter, r *http.Request, key string, spec runner.Spec) {
	user := userFrom(r.Context())
	if origin := r.Header.Get("Origin"); origin != "" && !a.originAllowed(r, origin) {
		writeError(w, http.StatusForbidden, codeCSRFRejected, "cross-origin terminal rejected")
		return
	}
	if !terminalLimiter.Allow(strconv.Itoa(user.ID)) {
		writeRateLimited(w, terminalLimiter)
		return
	}
	cols := boundedQueryInt(r, "cols", 100, 20, 500)
	rows := boundedQueryInt(r, "rows", 28, 5, 200)

	upgrader := websocket.Upgrader{ReadBufferSize: 4096, WriteBufferSize: 4096, CheckOrigin: func(*http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	say := func(text string) { _ = conn.WriteMessage(websocket.TextMessage, []byte(text)) }
	say(lab.Banner)
	say("\x1b[90mЗапускаю песочницу…\x1b[0m\r\n")
	handle, err := a.Sandbox.EnsureSession(r.Context(), user.ID, key, spec)
	if err != nil {
		a.log.Error("terminal: ensure session", "error", err)
		say("\r\n\x1b[31m" + err.Error() + "\x1b[0m\r\n")
		return
	}
	pty, err := a.Sandbox.OpenPTY(handle, cols, rows)
	if err != nil {
		a.log.Error("terminal: open pty", "error", err)
		say("\r\n\x1b[31mНе удалось запустить терминал: " + err.Error() + "\x1b[0m\r\n")
		return
	}
	defer pty.Close()
	lab.Pump(conn, pty, func() { a.Sandbox.Touch(user.ID, key) })
}

func boundedQueryInt(r *http.Request, name string, def, min, max int) int {
	n, err := strconv.Atoi(r.URL.Query().Get(name))
	if err != nil || n < min || n > max {
		return def
	}
	return n
}

func (a *API) getLabGitGraph(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.sandboxLab(w, r)
	if !ok {
		return
	}
	a.writeGitGraph(w, r, ref.key(), ref.spec, lab.LabGitRepo)
}

// writeGitGraph returns the parsed commit graph of repo, or an empty graph when no session is running.
func (a *API) writeGitGraph(w http.ResponseWriter, r *http.Request, key string, spec runner.Spec, repo string) {
	ctx := r.Context()
	uid := userFrom(ctx).ID
	out := apigen.GitGraph{Commits: []apigen.GitCommit{}}
	if _, running := a.Sandbox.Session(uid, key); !running {
		writeJSON(w, http.StatusOK, out)
		return
	}
	log, err := a.Sandbox.Exec(ctx, uid, key, spec, lab.GitLogCommand(repo))
	if err != nil {
		a.sandboxError(w, "git graph", err)
		return
	}
	for _, c := range lab.ParseGitLog(log) {
		gc := apigen.GitCommit{Hash: c.Hash, Parents: c.Parents, Subject: c.Subject, Refs: make([]apigen.GitRef, 0, len(c.Refs))}
		for _, ref := range c.Refs {
			gc.Refs = append(gc.Refs, apigen.GitRef{Name: ref.Name, Type: apigen.GitRefType(ref.Type)})
		}
		out.Commits = append(out.Commits, gc)
	}
	writeJSON(w, http.StatusOK, out)
}

// fsLab resolves a running lab sandbox and the jailed ?path for editor endpoints.
func (a *API) fsLab(w http.ResponseWriter, r *http.Request) (labRef, string, bool) {
	ref, ok := a.sandboxLab(w, r)
	if !ok {
		return labRef{}, "", false
	}
	p, ok := lab.JailPath(r.URL.Query().Get("path"))
	if !ok {
		writeError(w, http.StatusBadRequest, codePathOutsideJail, "path must stay inside /root")
		return labRef{}, "", false
	}
	if _, running := a.Sandbox.Session(userFrom(r.Context()).ID, ref.key()); !running {
		writeError(w, http.StatusConflict, codeSandboxNotRunning, "sandbox is not running")
		return labRef{}, "", false
	}
	return ref, p, true
}

func (a *API) listLabFiles(w http.ResponseWriter, r *http.Request) {
	ref, p, ok := a.fsLab(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	entries, err := a.Sandbox.FSList(ctx, userFrom(ctx).ID, ref.key(), ref.spec, p)
	if err != nil {
		a.sandboxError(w, "fs list", err)
		return
	}
	out := apigen.LabDirListing{Path: p, Entries: make([]apigen.LabFileEntry, 0, len(entries))}
	for _, e := range entries {
		out.Entries = append(out.Entries, apigen.LabFileEntry{Name: e.Name, IsDir: e.Dir})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) readLabFile(w http.ResponseWriter, r *http.Request) {
	ref, p, ok := a.fsLab(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	data, err := a.Sandbox.FSRead(ctx, userFrom(ctx).ID, ref.key(), ref.spec, p)
	if err != nil {
		a.sandboxError(w, "fs read", err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(data)
}

func (a *API) writeLabFile(w http.ResponseWriter, r *http.Request) {
	ref, p, ok := a.fsLab(w, r)
	if !ok {
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, runner.MaxFileSize))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, codeFileTooLarge, "file exceeds 2 MiB")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, codeMalformedJSON, "cannot read body")
		return
	}
	ctx := r.Context()
	if err := a.Sandbox.FSWrite(ctx, userFrom(ctx).ID, ref.key(), ref.spec, p, data); err != nil {
		a.sandboxError(w, "fs write", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) previewLab(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.sandboxLab(w, r)
	if !ok {
		return
	}
	port, err := strconv.Atoi(chi.URLParam(r, "port"))
	if err != nil || port < 1 || port > 65535 {
		writeError(w, http.StatusNotFound, codeNotFound, "invalid port")
		return
	}
	ctx := r.Context()
	uid := userFrom(ctx).ID

	// The service-wide policy is "default-src 'none'; frame-ancestors 'none'",
	// which is right for a JSON API and fatal here: this is the one endpoint that
	// returns a page meant to be framed, and frame-ancestors wins over
	// X-Frame-Options, so the browser refused to show the preview at all. The
	// same policy also blocked the page's own assets and the <base href> injected
	// below.
	//
	// Relaxing it is safe because the isolation is the iframe, not the header: the
	// frontend mounts it with sandbox="allow-scripts allow-forms" and no
	// allow-same-origin, so the student's page runs in an opaque origin and cannot
	// reach our cookies. Framing stays restricted to us.
	//
	// The pane also offers "open in a new tab", and there the sandbox is gone: the
	// student's own HTML then runs on our origin. That is deliberate — it is the
	// only way their JavaScript works in the standalone view, which is the point
	// of the feature — and it grants nothing, because the session cookie is
	// HttpOnly and the preview always resolves to the VM of whoever is asking. A
	// student cannot serve their page to anyone else, so the worst case is running
	// their own code against their own account.
	previewCSP := func() {
		w.Header().Set("Content-Security-Policy", "default-src * data: blob: 'unsafe-inline' 'unsafe-eval'; frame-ancestors 'self'")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	}
	placeholder := func(err error) {
		previewCSP()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(lab.PreviewErrorPage(port, err)))
	}
	if _, running := a.Sandbox.Session(uid, ref.key()); !running {
		placeholder(errors.New("sandbox not running"))
		return
	}
	target := "/" + chi.URLParam(r, "*")
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	body, ct, status, err := a.Sandbox.Preview(ctx, uid, ref.key(), ref.spec, port, target)
	if err != nil {
		placeholder(err)
		return
	}
	if strings.HasPrefix(ct, "text/html") {
		body = lab.InjectBase(body, fmt.Sprintf("/api/v1/lessons/%d/lab/preview/%d/", ref.lesson.ID, port))
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	if status == 0 {
		status = http.StatusOK
	}
	previewCSP()
	w.Header().Set("Content-Type", ct)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// sandboxError maps runner errors to API errors.
func (a *API) sandboxError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, runner.ErrFileNotFound):
		writeError(w, http.StatusNotFound, codeNotFound, "file not found")
	case errors.Is(err, runner.ErrFileTooLarge):
		writeError(w, http.StatusRequestEntityTooLarge, codeFileTooLarge, "file exceeds 2 MiB")
	case errors.Is(err, context.Canceled):
		// Still write a status. Returning without one makes net/http send an
		// implicit 200 with an empty body, which a client that did *not* go away —
		// a cancelled inner context, say — reads as "the file is empty" or "there
		// are no commits".
		a.log.Info(op+": client went away", "error", err)
		writeError(w, http.StatusServiceUnavailable, codeSandboxError, "request cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		a.log.Warn(op+": timed out", "error", err)
		writeError(w, http.StatusGatewayTimeout, codeSandboxError,
			"песочница не ответила вовремя — попробуй ещё раз")
	default:
		a.log.Error(op, "error", err)
		writeError(w, http.StatusBadGateway, codeSandboxError, err.Error())
	}
}
