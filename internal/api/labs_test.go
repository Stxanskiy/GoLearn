package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/auth"
	"github.com/backendraz/golearn/internal/model"
	"github.com/gorilla/websocket"
)

// labFixture gives lesson 102 (linux lab) an auto-checked, a manual and a hidden-solution task, and lesson 111 (docker) a Go task.
func labFixture(t *testing.T) (http.Handler, *fakeContent) {
	t.Helper()
	h, c := lessonsFixture(t)
	c.progress[1] = nil
	c.labPassed[1] = nil
	c.tasks[102] = []model.Task{
		{ID: 900, LessonID: 102, Kind: "shell", Title: "Create dir", Format: "md", Description: "Run **mkdir**", Hints: "use -p",
			CheckScript: "test -d /root/x SECRET_CHECK", SetupScript: "SECRET_SETUP", Solution: "SECRET_SOLUTION",
			Glossary: []model.GlossaryItem{{Term: "mkdir", Definition: "make dir"}}},
		{ID: 901, LessonID: 102, Kind: "shell", Title: "Look around", Format: "html", Description: "<p>ls</p>"},
	}
	c.tasks[111] = []model.Task{
		{ID: 950, LessonID: 111, Kind: "go", Title: "Answer", StarterCode: "package main",
			TestCases: []model.TestCase{{Input: "", ExpectedOutput: "SECRET_EXPECTED"}}},
	}
	checkLimiter = auth.NewLimiter(30, time.Minute)
	runLimiter = auth.NewLimiter(30, time.Minute)
	terminalLimiter = auth.NewLimiter(20, time.Minute)
	return h, c
}

func TestGetLab(t *testing.T) {
	h, c := labFixture(t)
	c.passed[1] = map[int]bool{901: true}

	w := do(h, http.MethodGet, "/lessons/102/lab", "", withCookie(studentToken))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	for _, leak := range []string{"SECRET_CHECK", "SECRET_SETUP", "SECRET_SOLUTION"} {
		if strings.Contains(w.Body.String(), leak) {
			t.Errorf("lab leaks %s", leak)
		}
	}
	l := decode[apigen.Lab](t, w)
	if l.Mode != "shell" || l.IsGitLab || l.Sandbox.Running || !l.Sandbox.Enabled || len(l.Tasks) != 2 || l.Nav.Position != 3 {
		t.Fatalf("lab = %+v", l)
	}
	auto, manual := l.Tasks[0], l.Tasks[1]
	if auto.CheckMode != "auto" || auto.Passed || !strings.Contains(auto.DescriptionHTML, "<strong>mkdir</strong>") || auto.HintsHTML == nil || len(auto.Glossary) != 1 || auto.StarterCode != nil {
		t.Errorf("auto task = %+v", auto)
	}
	if manual.CheckMode != "manual" || !manual.Passed || manual.HintsHTML != nil {
		t.Errorf("manual task = %+v", manual)
	}

	w = do(h, http.MethodGet, "/lessons/111/lab", "", withCookie(studentToken))
	if strings.Contains(w.Body.String(), "SECRET_EXPECTED") {
		t.Error("go lab leaks expected output")
	}
	if g := decode[apigen.Lab](t, w); g.Mode != "code" || g.Tasks[0].CheckMode != "tests" || g.Tasks[0].StarterCode == nil {
		t.Errorf("go lab = %+v", g)
	}

	for _, id := range []string{"100", "103", "999"} {
		if w := do(h, http.MethodGet, "/lessons/"+id+"/lab", "", withCookie(studentToken)); w.Code != http.StatusNotFound {
			t.Errorf("lesson %s lab: %d", id, w.Code)
		}
	}
}

func TestLabSessionAndRetry(t *testing.T) {
	h, c := labFixture(t)
	sb := c.sandbox

	if s := decode[apigen.SandboxSession](t, do(h, http.MethodGet, "/lessons/102/lab/session", "", withCookie(studentToken))); s.Running || s.ExpiresAt != nil {
		t.Errorf("idle session = %+v", s)
	}
	sb.running(1, "l102")
	if s := decode[apigen.SandboxSession](t, do(h, http.MethodGet, "/lessons/102/lab/session", "", withCookie(studentToken))); !s.Running || s.ExpiresAt == nil || s.StartedAt == nil {
		t.Errorf("running session = %+v", s)
	}
	if w := do(h, http.MethodDelete, "/lessons/102/lab/session", "", withCookie(studentToken)); w.Code != http.StatusNoContent || len(sb.resets) != 1 {
		t.Fatalf("stop: %d resets %v", w.Code, sb.resets)
	}

	c.passed[1] = map[int]bool{900: true, 901: true}
	c.answers[1] = map[int]int{1001: 1}
	c.questions[102] = c.questions[101]
	score := 1
	c.progress[1] = []model.Progress{{LessonID: 102, Status: "completed", QuizScore: &score, Notes: "keep"}}
	if w := do(h, http.MethodPost, "/lessons/102/lab/retry", "", withCookie(studentToken)); w.Code != http.StatusNoContent {
		t.Fatalf("retry: %d %s", w.Code, w.Body)
	}
	if len(c.passed[1]) != 0 || len(c.answers[1]) != 0 || c.progress[1][0].Status != "not_started" || c.progress[1][0].Notes != "keep" || len(sb.resets) != 2 {
		t.Errorf("after retry passed=%v answers=%v progress=%+v resets=%v", c.passed[1], c.answers[1], c.progress[1], sb.resets)
	}
}

func TestCheckAndDone(t *testing.T) {
	h, c := labFixture(t)
	sb := c.sandbox
	post := func(path string) *httptest.ResponseRecorder {
		return do(h, http.MethodPost, path, "", withCookie(studentToken))
	}

	sb.checkPass = false
	r := decode[apigen.TaskCheckResult](t, post("/tasks/900/check"))
	if r.Passed || r.Output == nil || *r.Output != "check output" || r.LessonStatus != "not_started" || c.passed[1][900] {
		t.Errorf("failed check = %+v", r)
	}

	sb.checkPass = true
	r = decode[apigen.TaskCheckResult](t, post("/tasks/900/check"))
	if !r.Passed || r.LessonStatus != "in_progress" {
		t.Errorf("passed check = %+v", r)
	}

	if w := post("/tasks/900/done"); w.Code != http.StatusConflict || errorCode(t, w) != codeTaskAutoChecked {
		t.Errorf("done on auto-checked task: %d %s", w.Code, w.Body)
	}
	if w := post("/tasks/901/check"); w.Code != http.StatusConflict || errorCode(t, w) != codeTaskNotAutoChecked {
		t.Errorf("check on manual task: %d", w.Code)
	}

	r = decode[apigen.TaskCheckResult](t, post("/tasks/901/done"))
	if !r.Passed || r.Output != nil || r.LessonStatus != "completed" {
		t.Errorf("manual done = %+v", r)
	}
	if len(c.saves) != 3 || c.saves[2].code != "[manual]" {
		t.Errorf("submissions = %+v", c.saves)
	}

	sb.checkErr = errors.New("VM slots busy")
	if w := post("/tasks/900/check"); w.Code != http.StatusBadGateway || errorCode(t, w) != codeSandboxError {
		t.Errorf("sandbox error: %d", w.Code)
	}
	sb.checkErr = nil
	checkLimiter = auth.NewLimiter(1, time.Minute)
	post("/tasks/900/check")
	if w := post("/tasks/900/check"); w.Code != http.StatusTooManyRequests {
		t.Errorf("rate limit: %d", w.Code)
	}
	sb.enabled = false
	if w := post("/tasks/900/check"); w.Code != http.StatusServiceUnavailable {
		t.Errorf("disabled sandbox: %d", w.Code)
	}
	if w := post("/tasks/12345/check"); w.Code != http.StatusNotFound {
		t.Errorf("unknown task: %d", w.Code)
	}
}

func TestLabFiles(t *testing.T) {
	h, c := labFixture(t)
	sb := c.sandbox
	get := func(path string) *httptest.ResponseRecorder {
		return do(h, http.MethodGet, path, "", withCookie(studentToken))
	}

	if w := get("/lessons/102/lab/fs/entries?path=/root"); w.Code != http.StatusConflict || errorCode(t, w) != codeSandboxNotRunning {
		t.Fatalf("not running: %d %s", w.Code, w.Body)
	}
	sb.running(1, "l102")

	l := decode[apigen.LabDirListing](t, get("/lessons/102/lab/fs/entries"))
	if l.Path != "/root" || len(l.Entries) != 2 || !l.Entries[0].IsDir {
		t.Errorf("listing = %+v", l)
	}
	if w := get("/lessons/102/lab/fs/entries?path=/etc"); w.Code != http.StatusBadRequest || errorCode(t, w) != codePathOutsideJail {
		t.Errorf("jail: %d", w.Code)
	}

	content := strings.Repeat("big file line\n", 10000)
	r := httptest.NewRequest(http.MethodPut, "/lessons/102/lab/fs/content?path=project/app.yml", strings.NewReader(content))
	r.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: studentToken})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || string(sb.files["/root/project/app.yml"]) != content {
		t.Fatalf("write: %d %s", w.Code, w.Body)
	}
	w = get("/lessons/102/lab/fs/content?path=/root/project/app.yml")
	if w.Code != http.StatusOK || w.Body.String() != content || w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("read: %d %q", w.Code, w.Header().Get("Content-Type"))
	}
	if w := get("/lessons/102/lab/fs/content?path=missing"); w.Code != http.StatusNotFound {
		t.Errorf("missing: %d", w.Code)
	}
	if w := get("/lessons/102/lab/fs/content?path=huge"); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("huge: %d", w.Code)
	}

	r = httptest.NewRequest(http.MethodPut, "/lessons/102/lab/fs/content?path=x", strings.NewReader(strings.Repeat("x", 3<<20)))
	r.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: studentToken})
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized write: %d", w.Code)
	}
}

func TestLabPreviewAndGitGraph(t *testing.T) {
	h, c := labFixture(t)
	sb := c.sandbox
	get := func(path string) *httptest.ResponseRecorder {
		return do(h, http.MethodGet, path, "", withCookie(studentToken))
	}

	w := get("/lessons/102/lab/preview/8080/index.html")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "127.0.0.1:8080") || len(sb.ensures) != 0 {
		t.Fatalf("preview without session must not boot: %d ensures=%v", w.Code, sb.ensures)
	}
	sb.running(1, "l102")
	sb.preview.body, sb.preview.ct, sb.preview.status = []byte("<html><head></head>hi"), "text/html", 404
	w = get("/lessons/102/lab/preview/8080/app/?q=1")
	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `<base href="/api/v1/lessons/102/lab/preview/8080/">`) {
		t.Errorf("html preview: %d %s", w.Code, w.Body)
	}
	sb.preview.body, sb.preview.ct, sb.preview.status = []byte("{}"), "application/json", 200
	if w := get("/lessons/102/lab/preview/8080/api"); strings.Contains(w.Body.String(), "<base") || w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("json preview: %s", w.Body)
	}
	if w := get("/lessons/102/lab/preview/99999/"); w.Code != http.StatusNotFound {
		t.Errorf("bad port: %d", w.Code)
	}

	delete(sb.sessions, sid(1, "l102"))
	if g := decode[apigen.GitGraph](t, get("/lessons/102/lab/git-graph")); len(g.Commits) != 0 || len(sb.execs) != 0 {
		t.Errorf("graph without session = %+v execs=%v", g, sb.execs)
	}
	sb.running(1, "l102")
	sb.execOut = "abc\x1f\x1fHEAD -> refs/heads/main, refs/remotes/origin/main\x1finit\n"
	g := decode[apigen.GitGraph](t, get("/lessons/102/lab/git-graph"))
	if len(g.Commits) != 1 || len(g.Commits[0].Refs) != 3 || g.Commits[0].Refs[2].Type != "remote" || !strings.Contains(sb.execs[0], "/root/project") {
		t.Errorf("graph = %+v execs=%v", g, sb.execs)
	}

	sb.running(1, "git")
	if g := decode[apigen.GitGraph](t, get("/git-trainer/git-graph")); len(g.Commits) != 1 || !strings.Contains(sb.execs[1], "/root/repo") {
		t.Errorf("trainer graph = %+v", g)
	}
	if s := decode[apigen.SandboxSession](t, get("/git-trainer/session")); !s.Running {
		t.Errorf("trainer session = %+v", s)
	}
	if w := do(h, http.MethodDelete, "/git-trainer/session", "", withCookie(studentToken)); w.Code != http.StatusNoContent || sb.resets[len(sb.resets)-1] != "u1-git" {
		t.Errorf("trainer reset: %d %v", w.Code, sb.resets)
	}
}

func TestRunCode(t *testing.T) {
	h, c := labFixture(t)
	post := func(path, body string) *httptest.ResponseRecorder {
		return do(h, http.MethodPost, path, body, withCookie(studentToken))
	}

	w := post("/tasks/950/run", `{"code":"package main"}`)
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "SECRET_EXPECTED") || strings.Contains(w.Body.String(), "expected") {
		t.Fatalf("run tests: %d %s", w.Code, w.Body)
	}
	res := decode[apigen.RunResult](t, w)
	if res.AllPassed == nil || *res.AllPassed || res.TestResults == nil || len(*res.TestResults) != 1 || (*res.TestResults)[0].Actual != "42" || *res.LessonStatus != "not_started" {
		t.Errorf("run result = %+v", res)
	}
	if len(c.saves) != 1 || c.saves[0].passed {
		t.Errorf("submissions = %+v", c.saves)
	}

	if w := post("/tasks/900/run", `{"code":"x"}`); w.Code != http.StatusConflict {
		t.Errorf("shell task run: %d", w.Code)
	}
	if w := post("/playground/run", `{"code":"  "}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty code: %d", w.Code)
	}
	if r := decode[apigen.RunResult](t, post("/playground/run", `{"code":"package main","stdin":"hi"}`)); r.Output != "out:hi" || r.TestResults != nil {
		t.Errorf("playground = %+v", r)
	}
}

func TestLabTerminal(t *testing.T) {
	h, c := labFixture(t)
	sb := c.sandbox
	srv := httptest.NewServer(h)
	defer srv.Close()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/lessons/102/lab/terminal?cols=120&rows=30"
	header := http.Header{"Cookie": {auth.SessionCookie + "=" + studentToken}}

	if _, resp, err := websocket.DefaultDialer.Dial(url, http.Header{"Cookie": header["Cookie"], "Origin": {"https://evil.test"}}); err == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("foreign origin must be rejected: %v", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var greeting strings.Builder
	for i := 0; i < 2; i++ {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read greeting: %v", err)
		}
		greeting.Write(msg)
	}
	if !strings.Contains(greeting.String(), "TOT lab environment") {
		t.Fatalf("greeting %q", greeting.String())
	}

	_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"resize":[100,40]}`))
	_ = conn.WriteMessage(websocket.BinaryMessage, []byte("ls\r"))
	mt, msg, err := conn.ReadMessage()
	if err != nil || mt != websocket.BinaryMessage || string(msg) != "ls\r" {
		t.Fatalf("echo: %v %d %q (resize must not reach stdin)", err, mt, msg)
	}
	sb.mu.Lock()
	if len(sb.touches) != 1 || len(sb.ensures) != 1 || !strings.Contains(sb.ensures[0], "u1-l102|golearn/sandbox:latest|setup-102") {
		t.Errorf("touches %v ensures %v", sb.touches, sb.ensures)
	}
	sb.mu.Unlock()

	sb.enabled = false
	if _, resp, err := websocket.DefaultDialer.Dial(url, header); err == nil || resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("disabled sandbox: %v", err)
	}
}
