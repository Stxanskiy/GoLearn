package runner

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestShellRunnerLocal exercises the container backend end to end on this machine's
// Docker. It needs the golearn/sandbox image, so it only runs with SANDBOX_LOCAL set.
func TestShellRunnerLocal(t *testing.T) {
	if !shellBool("SANDBOX_LOCAL") {
		t.Skip("set SANDBOX_LOCAL=1 (and build golearn/sandbox) to run the container sandbox test")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skip("docker is not available")
	}

	s := NewShellRunner()
	if !s.Enabled() {
		t.Fatal("ShellRunner disabled with SANDBOX_LOCAL set")
	}
	const key = "selftest"
	uid := os.Getpid() // keep parallel runs from sharing a container
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	defer func() { _ = s.Reset(context.Background(), uid, key) }()

	if _, err := s.EnsureSession(ctx, uid, key, "golearn/sandbox:latest", "echo hi > /root/setup.txt"); err != nil {
		t.Fatalf("EnsureSession: %v", err)
	}
	if !s.HasSession(uid, key) {
		t.Error("HasSession = false right after EnsureSession")
	}
	info, ok := s.Session(uid, key)
	if !ok || !info.ExpiresAt.After(time.Now()) {
		t.Errorf("Session = %+v, %v", info, ok)
	}

	out, err := s.Exec(ctx, uid, key, "golearn/sandbox:latest", "", "cat /root/setup.txt")
	if err != nil || !strings.Contains(out, "hi") {
		t.Errorf("setup script did not run: out=%q err=%v", out, err)
	}

	passed, _, err := s.Check(ctx, uid, key, "golearn/sandbox:latest", "", "test -f /root/setup.txt")
	if err != nil || !passed {
		t.Errorf("Check = %v, %v", passed, err)
	}

	if err := s.FSWrite(ctx, uid, key, "golearn/sandbox:latest", "", "/root/note.txt", []byte("written")); err != nil {
		t.Fatalf("FSWrite: %v", err)
	}
	data, err := s.FSRead(ctx, uid, key, "golearn/sandbox:latest", "", "/root/note.txt")
	if err != nil || string(data) != "written" {
		t.Errorf("FSRead = %q, %v", data, err)
	}
	entries, err := s.FSList(ctx, uid, key, "golearn/sandbox:latest", "", "/root")
	if err != nil {
		t.Fatalf("FSList: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.Name == "note.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("FSList(/root) missing note.txt: %+v", entries)
	}

	// The interactive terminal is what a student actually opens.
	handle, err := s.EnsureSession(ctx, uid, key, "golearn/sandbox:latest", "")
	if err != nil {
		t.Fatalf("EnsureSession (pty): %v", err)
	}
	pty, err := s.OpenPTY(handle, 80, 24)
	if err != nil {
		t.Fatalf("OpenPTY: %v", err)
	}
	defer pty.Close()
	if _, err := pty.Stdin.Write([]byte("echo GLPTY_OK\n")); err != nil {
		t.Fatalf("pty write: %v", err)
	}
	buf := make([]byte, 4096)
	deadline := time.Now().Add(20 * time.Second)
	var seen string
	for time.Now().Before(deadline) && !strings.Contains(seen, "GLPTY_OK") {
		n, err := pty.Stdout.Read(buf)
		seen += string(buf[:n])
		if err != nil {
			break
		}
	}
	if !strings.Contains(seen, "GLPTY_OK") {
		t.Errorf("pty did not echo the command; got %q", seen)
	}
}
