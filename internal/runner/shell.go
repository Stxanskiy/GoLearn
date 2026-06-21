package runner

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"golang.org/x/crypto/ssh"
)

const (
	shellExecTimeout = 20 * time.Second
	shellSessionTTL  = 30 * time.Minute
	maxShellOutput   = 64 * 1024
)

// ShellRunner runs student shell commands inside per-(user,task) ephemeral
// containers that live in a dedicated sandbox VM, reached over SSH. The VM
// boundary isolates arbitrary shell execution from the host and the relay.
type ShellRunner struct {
	host, port, user, image, keyFile string
	enabled                          bool
	local                            bool // run docker on the host directly (no SSH/VM)
	mu                               sync.Mutex
	sessions                         map[string]time.Time
}

func NewShellRunner() *ShellRunner {
	s := &ShellRunner{
		host:     shellEnv("SANDBOX_SSH_HOST", ""),
		port:     shellEnv("SANDBOX_SSH_PORT", "2222"),
		user:     shellEnv("SANDBOX_SSH_USER", "sandbox"),
		image:    shellEnv("SANDBOX_IMAGE", "ubuntu:24.04"),
		sessions: make(map[string]time.Time),
	}
	// Local mode: execute docker on the host directly. Suited to a single-user
	// local install where a dedicated sandbox VM is overkill.
	if v := strings.ToLower(shellEnv("SANDBOX_LOCAL", "")); v == "1" || v == "true" || v == "yes" {
		s.local = true
		s.enabled = true
		go s.reaper()
		return s
	}
	keySrc := shellEnv("SANDBOX_SSH_KEY", "")
	if s.host == "" || keySrc == "" {
		return s // disabled — Enabled() reports false
	}
	// ssh refuses world-readable private keys, so copy to a 0600 file.
	data, err := os.ReadFile(keySrc)
	if err != nil {
		return s
	}
	f, err := os.CreateTemp("", "sbkey-*")
	if err != nil {
		return s
	}
	_, _ = f.Write(data)
	_ = f.Chmod(0o600)
	_ = f.Close()
	s.keyFile = f.Name()
	s.enabled = true
	go s.reaper()
	return s
}

func (s *ShellRunner) Enabled() bool { return s != nil && s.enabled }

// run executes a shell script via the active transport: locally (bash on the
// host) in local mode, or over SSH to the sandbox VM otherwise.
func (s *ShellRunner) run(ctx context.Context, script string) (string, int, error) {
	if s.local {
		cmd := exec.CommandContext(ctx, "bash", "-c", script)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		exit := 0
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
			err = nil
		}
		return out.String(), exit, err
	}
	return s.runSSH(ctx, script)
}

func (s *ShellRunner) runSSH(ctx context.Context, remote string) (string, int, error) {
	args := []string{
		"-i", s.keyFile,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=8",
		"-o", "LogLevel=ERROR",
		"-p", s.port,
		s.user + "@" + s.host,
		remote,
	}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	exit := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exit = ee.ExitCode()
		err = nil
	}
	return out.String(), exit, err
}

func sessionName(userID, taskID int) string {
	return fmt.Sprintf("gl-s-u%d-t%d", userID, taskID)
}

func (s *ShellRunner) touch(c string) {
	s.mu.Lock()
	s.sessions[c] = time.Now()
	s.mu.Unlock()
}

// ensure creates the per-(user,task) container if it is not running, applying
// the task setup script once. Returns the container name.
func (s *ShellRunner) ensure(ctx context.Context, userID, taskID int, image, setup string) (string, error) {
	if image == "" {
		image = s.image
	}
	c := sessionName(userID, taskID)
	out, _, err := s.run(ctx, fmt.Sprintf("docker inspect -f '{{.State.Running}}' %s 2>/dev/null", c))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "true" {
		s.touch(c)
		return c, nil
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(setup))
	create := fmt.Sprintf(
		`docker rm -f %s >/dev/null 2>&1; `+
			`docker run -d --name %s --hostname sandbox --network none --memory 512m --cpus 1 --pids-limit 256 %s sleep infinity >/dev/null 2>&1 || exit 1; `+
			`if [ -n "%s" ]; then docker exec %s sh -c 'echo %s | base64 -d | bash' >/dev/null 2>&1; fi; `+
			`echo OK`,
		c, c, image, b64, c, b64,
	)
	out, _, err = s.run(ctx, create)
	if err != nil {
		return "", err
	}
	if !strings.Contains(out, "OK") {
		return "", fmt.Errorf("sandbox start failed: %s", strings.TrimSpace(out))
	}
	s.touch(c)
	return c, nil
}

// wrap persists the working directory between exec calls, since each
// `docker exec` is a fresh process that would otherwise reset cwd.
func wrap(command string) string {
	return `cd "$(cat /root/.gl_cwd 2>/dev/null || echo /root)" 2>/dev/null` + "\n" +
		command + "\n" +
		`__rc=$?` + "\n" +
		`pwd > /root/.gl_cwd 2>/dev/null` + "\n" +
		`exit $__rc` + "\n"
}

func (s *ShellRunner) execIn(ctx context.Context, container, script string) (string, int, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(script))
	remote := fmt.Sprintf(`docker exec %s sh -c 'echo %s | base64 -d | bash' 2>&1`, container, b64)
	out, exit, err := s.run(ctx, remote)
	if len(out) > maxShellOutput {
		out = out[:maxShellOutput] + "\n… (вывод обрезан)"
	}
	return out, exit, err
}

// Exec runs a user command in the session container and returns combined output.
func (s *ShellRunner) Exec(ctx context.Context, userID, taskID int, image, setup, command string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, shellExecTimeout)
	defer cancel()
	c, err := s.ensure(ctx, userID, taskID, image, setup)
	if err != nil {
		return "", err
	}
	out, _, err := s.execIn(ctx, c, wrap(command))
	return out, err
}

// Check runs the task's check script in the session; passed = exit code 0.
func (s *ShellRunner) Check(ctx context.Context, userID, taskID int, image, setup, checkScript string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, shellExecTimeout)
	defer cancel()
	c, err := s.ensure(ctx, userID, taskID, image, setup)
	if err != nil {
		return false, "", err
	}
	out, exit, err := s.execIn(ctx, c, wrap(checkScript))
	if err != nil {
		return false, out, err
	}
	return exit == 0, out, nil
}

// Reset destroys the session container so the next command starts fresh.
func (s *ShellRunner) Reset(ctx context.Context, userID, taskID int) error {
	ctx, cancel := context.WithTimeout(ctx, shellExecTimeout)
	defer cancel()
	c := sessionName(userID, taskID)
	_, _, err := s.run(ctx, fmt.Sprintf("docker rm -f %s >/dev/null 2>&1; echo OK", c))
	s.mu.Lock()
	delete(s.sessions, c)
	s.mu.Unlock()
	return err
}

func (s *ShellRunner) reaper() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		s.mu.Lock()
		now := time.Now()
		var stale []string
		for c, last := range s.sessions {
			if now.Sub(last) > shellSessionTTL {
				stale = append(stale, c)
			}
		}
		for _, c := range stale {
			delete(s.sessions, c)
		}
		s.mu.Unlock()
		for _, c := range stale {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, _, _ = s.run(ctx, fmt.Sprintf("docker rm -f %s >/dev/null 2>&1", c))
			cancel()
		}
	}
}

func shellEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ── Interactive PTY terminal (xterm.js over WebSocket) ──

// EnsureSession is a public wrapper to create/get the session container.
func (s *ShellRunner) EnsureSession(ctx context.Context, userID, taskID int, image, setup string) (string, error) {
	return s.ensure(ctx, userID, taskID, image, setup)
}

func (s *ShellRunner) signer() (ssh.Signer, error) {
	data, err := os.ReadFile(s.keyFile)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(data)
}

// PTYSession is a live interactive shell into a sandbox container. The
// transport (SSH or local docker) is hidden behind resize/closer closures.
type PTYSession struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	resize func(rows, cols int)
	closer func()
}

func (p *PTYSession) Resize(rows, cols int) {
	if p.resize != nil {
		p.resize(rows, cols)
	}
}

func (p *PTYSession) Close() {
	if p.closer != nil {
		p.closer()
	}
}

// OpenPTY opens an interactive bash (with a real TTY) inside the container.
func (s *ShellRunner) OpenPTY(container string, cols, rows int) (*PTYSession, error) {
	if !s.enabled {
		return nil, fmt.Errorf("sandbox disabled")
	}
	if s.local {
		return s.openPTYLocal(container, cols, rows)
	}
	return s.openPTYSSH(container, cols, rows)
}

// openPTYLocal attaches to an interactive docker exec via a host PTY.
func (s *ShellRunner) openPTYLocal(container string, cols, rows int) (*PTYSession, error) {
	cmd := exec.Command("docker", "exec", "-it", container, "env", "TERM=xterm-256color", "HOME=/root", "bash")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	_ = pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
	return &PTYSession{
		Stdin:  ptmx,
		Stdout: ptmx,
		resize: func(r, c int) { _ = pty.Setsize(ptmx, &pty.Winsize{Rows: uint16(r), Cols: uint16(c)}) },
		closer: func() {
			_ = ptmx.Close()
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = cmd.Wait()
		},
	}, nil
}

func (s *ShellRunner) openPTYSSH(container string, cols, rows int) (*PTYSession, error) {
	sg, err := s.signer()
	if err != nil {
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            s.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(sg)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	client, err := ssh.Dial("tcp", s.host+":"+s.port, cfg)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, err
	}
	modes := ssh.TerminalModes{ssh.ECHO: 1, ssh.TTY_OP_ISPEED: 14400, ssh.TTY_OP_OSPEED: 14400}
	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, err
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, err
	}
	cmd := fmt.Sprintf("docker exec -it %s env TERM=xterm-256color HOME=/root bash", container)
	if err := session.Start(cmd); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, err
	}
	return &PTYSession{
		Stdin:  stdin,
		Stdout: stdout,
		resize: func(r, c int) { _ = session.WindowChange(r, c) },
		closer: func() { _ = session.Close(); _ = client.Close() },
	}, nil
}
