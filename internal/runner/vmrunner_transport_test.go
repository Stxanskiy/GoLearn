package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeSSH is an ssh stand-in that runs the remote command locally, honouring -n, so the two-hop quoting and stdin plumbing are exercised for real.
const fakeSSH = `#!/usr/bin/env bash
nostdin=0
while [ $# -gt 0 ]; do
  case "$1" in
    -i|-o|-p) shift 2 ;;
    -n) nostdin=1; shift ;;
    -*) shift ;;
    *) shift; break ;;
  esac
done
if [ "$nostdin" = 1 ]; then exec bash -c "$*" </dev/null; fi
exec bash -c "$*"
`

func transportRunner(t *testing.T) *VMRunner {
	t.Helper()
	for _, bin := range []string{"bash", "base64", "find", "stat", "curl"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not available", bin)
		}
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ssh"), []byte(fakeSSH), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return &VMRunner{host: "fc", port: "22", user: "glvm", keyFile: "key", dir: dir, vmkey: "vmkey"}
}

func TestExecVMOutputAndExitCode(t *testing.T) {
	v := transportRunner(t)
	out, exit, err := v.execVM(context.Background(), "10.0.0.2", `echo "it's \"quoted\" $((1+2))"; exit 3`)
	if err != nil || exit != 3 || strings.TrimSpace(out) != `it's "quoted" 3` {
		t.Fatalf("out=%q exit=%d err=%v", out, exit, err)
	}
}

func TestFileRoundTripBeyondArgLimits(t *testing.T) {
	v := transportRunner(t)
	ctx := context.Background()
	file := filepath.Join(t.TempDir(), "nested dir", "big.bin")

	content := bytes.Repeat([]byte("line with 'quotes' and $vars\n\x00\xff"), 12000) // ~400 KiB, far above one argv string
	if err := v.writeFileVM(ctx, "10.0.0.2", file, content); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := v.readFileVM(ctx, "10.0.0.2", file)
	if err != nil || !bytes.Equal(got, content) {
		t.Fatalf("read back %d bytes (want %d), err %v", len(got), len(content), err)
	}

	entries, err := v.listDirVM(ctx, "10.0.0.2", filepath.Dir(filepath.Dir(file)))
	if err != nil || len(entries) != 1 || entries[0].Name != "nested dir" || !entries[0].Dir {
		t.Fatalf("list = %+v, err %v", entries, err)
	}

	if _, err := v.readFileVM(ctx, "10.0.0.2", file+".missing"); !errors.Is(err, ErrFileNotFound) {
		t.Errorf("missing file: %v", err)
	}
	if _, err := v.readFileVM(ctx, "10.0.0.2", filepath.Dir(file)); !errors.Is(err, ErrFileNotFound) {
		t.Errorf("directory: %v", err)
	}
	huge := filepath.Join(filepath.Dir(file), "huge")
	if err := os.WriteFile(huge, make([]byte, MaxFileSize+1), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := v.readFileVM(ctx, "10.0.0.2", huge); !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("huge file: %v", err)
	}
	if err := v.writeFileVM(ctx, "10.0.0.2", huge, make([]byte, MaxFileSize+1)); !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("huge write: %v", err)
	}
}

func TestPreviewLargeBody(t *testing.T) {
	v := transportRunner(t)
	body := strings.Repeat("<p>preview</p>", 20000) // ~280 KiB, above the display output cap
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RequestURI() != "/app/page?x=1&y='2'" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, body)
	}))
	defer srv.Close()
	_, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	var port int
	fmt.Sscan(portStr, &port)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	got, ct, status, err := v.previewVM(ctx, "10.0.0.2", port, "/app/page?x=1&y='2'")
	if err != nil || status != 200 || ct != "text/html; charset=utf-8" || string(got) != body {
		t.Fatalf("status=%d ct=%q len=%d err=%v", status, ct, len(got), err)
	}

	srv.Close()
	if _, _, _, err := v.previewVM(ctx, "10.0.0.2", port, "/"); err == nil || !strings.Contains(err.Error(), "no-server") {
		t.Errorf("closed port: %v", err)
	}
}

func TestSessionInfoAndTouch(t *testing.T) {
	now := time.Now()
	v := &VMRunner{enabled: true, sessions: map[string]*vmSession{
		vmSID(1, "l5"): {started: now.Add(-50 * time.Minute), last: now.Add(-1 * time.Minute)},
		vmSID(1, "l6"): {started: now.Add(-5 * time.Minute), last: now.Add(-20 * time.Minute)},
	}}
	info, ok := v.Session(1, "l5")
	if !ok || !info.ExpiresAt.Equal(info.Started.Add(shellSessionMax)) {
		t.Errorf("hard limit expiry: %+v %v", info, ok)
	}
	info, _ = v.Session(1, "l6")
	if got := info.ExpiresAt.Sub(now); got > 11*time.Minute || got < 9*time.Minute {
		t.Errorf("idle expiry in %v, want ~10m", got)
	}
	v.Touch(1, "l6")
	info, _ = v.Session(1, "l6")
	if got := info.ExpiresAt.Sub(now); got < 29*time.Minute {
		t.Errorf("touch did not extend idle expiry: %v", got)
	}
	if _, ok := v.Session(2, "l5"); ok {
		t.Error("other user's session must not be visible")
	}
}
