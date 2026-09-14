package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/backendraz/golearn/internal/runner"
)

// fakeSandbox records sandbox calls; a session exists once EnsureSession ran or running() was set.
type fakeSandbox struct {
	mu        sync.Mutex
	enabled   bool
	sessions  map[string]runner.SessionInfo
	files     map[string][]byte
	execOut   string
	checkPass bool
	checkErr  error
	preview   struct {
		body   []byte
		ct     string
		status int
	}
	execs, resets, touches, ensures []string
	pty                             *fakePTY
}

func newFakeSandbox() *fakeSandbox {
	return &fakeSandbox{enabled: true, sessions: map[string]runner.SessionInfo{}, files: map[string][]byte{}}
}

func sid(userID int, key string) string { return fmt.Sprintf("u%d-%s", userID, key) }

func (f *fakeSandbox) running(userID int, key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	f.sessions[sid(userID, key)] = runner.SessionInfo{Started: now, ExpiresAt: now.Add(30 * time.Minute)}
}

func (f *fakeSandbox) Enabled() bool { return f.enabled }

func (f *fakeSandbox) Session(userID int, key string) (runner.SessionInfo, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	info, ok := f.sessions[sid(userID, key)]
	return info, ok
}

func (f *fakeSandbox) Touch(userID int, key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.touches = append(f.touches, sid(userID, key))
}

func (f *fakeSandbox) EnsureSession(_ context.Context, userID int, key, image, setup string) (string, error) {
	f.mu.Lock()
	f.ensures = append(f.ensures, sid(userID, key)+"|"+image+"|"+setup)
	f.mu.Unlock()
	f.running(userID, key)
	return "10.0.0.2", nil
}

func (f *fakeSandbox) OpenPTY(string, int, int) (*runner.PTYSession, error) {
	f.pty = newFakePTY()
	return &runner.PTYSession{Stdin: f.pty.inW, Stdout: f.pty.outR}, nil
}

func (f *fakeSandbox) Exec(_ context.Context, userID int, key, _, _, command string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.execs = append(f.execs, sid(userID, key)+"|"+command)
	return f.execOut, nil
}

func (f *fakeSandbox) Check(context.Context, int, string, string, string, string) (bool, string, error) {
	return f.checkPass, "check output", f.checkErr
}

func (f *fakeSandbox) Preview(context.Context, int, string, string, string, int, string) ([]byte, string, int, error) {
	if f.preview.body == nil {
		return nil, "", 0, errors.New("preview: no-server")
	}
	return f.preview.body, f.preview.ct, f.preview.status, nil
}

func (f *fakeSandbox) FSList(context.Context, int, string, string, string, string) ([]runner.FSEntry, error) {
	return []runner.FSEntry{{Name: "project", Dir: true}, {Name: "notes.txt"}}, nil
}

func (f *fakeSandbox) FSRead(_ context.Context, _ int, _, _, _, file string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if file == "/root/huge" {
		return nil, runner.ErrFileTooLarge
	}
	data, ok := f.files[file]
	if !ok {
		return nil, runner.ErrFileNotFound
	}
	return data, nil
}

func (f *fakeSandbox) FSWrite(_ context.Context, _ int, _, _, _, file string, content []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.files[file] = content
	return nil
}

func (f *fakeSandbox) Reset(_ context.Context, userID int, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resets = append(f.resets, sid(userID, key))
	delete(f.sessions, sid(userID, key))
	return nil
}

// fakePTY echoes whatever is written to stdin back to stdout.
type fakePTY struct {
	inR  *io.PipeReader
	inW  *io.PipeWriter
	outR *io.PipeReader
	outW *io.PipeWriter
}

func newFakePTY() *fakePTY {
	p := &fakePTY{}
	p.inR, p.inW = io.Pipe()
	p.outR, p.outW = io.Pipe()
	go func() {
		_, _ = io.Copy(p.outW, p.inR)
		_ = p.outW.Close()
	}()
	return p
}

type fakeCode struct {
	lastCode string
}

func (f *fakeCode) Run(_ context.Context, code, stdin string) (*runner.Result, error) {
	f.lastCode = code
	return &runner.Result{Output: "out:" + stdin, ExitCode: 0}, nil
}

func (f *fakeCode) RunWithTests(_ context.Context, code string, tests []struct{ Input, Expected string }) (*runner.RunResult, error) {
	f.lastCode = code
	res := &runner.RunResult{AllPassed: true}
	for _, tc := range tests {
		actual := "42"
		passed := actual == tc.Expected
		res.AllPassed = res.AllPassed && passed
		res.TestResults = append(res.TestResults, runner.TestResult{Input: tc.Input, Expected: tc.Expected, Actual: actual, Passed: passed})
	}
	return res, nil
}
