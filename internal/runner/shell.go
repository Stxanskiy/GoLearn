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
	// Docker labs boot an engine and load images on first use, which is far
	// slower than a shell command.
	dockerExecTimeout = 3 * time.Minute
	shellSessionTTL   = 30 * time.Minute
	maxShellOutput    = 64 * 1024
)

// ShellRunner runs student shell commands inside per-(user,task) ephemeral
// containers that live in a dedicated sandbox VM, reached over SSH. The VM
// boundary isolates arbitrary shell execution from the host and the relay.
type ShellRunner struct {
	host, port, user, image, keyFile string
	enabled                          bool
	local                            bool // run docker on the host directly (no SSH/VM)
	// privileged enables the Docker/Kubernetes courses, whose labs need a
	// container runtime of their own and therefore --privileged. That is root
	// on the sandbox host for anyone who can open a lab, so it is OFF unless
	// SANDBOX_PRIVILEGED is set explicitly.
	privileged bool
	mu         sync.Mutex
	sessions   map[string]time.Time
}

func NewShellRunner() *ShellRunner {
	s := &ShellRunner{
		privileged: shellBool("SANDBOX_PRIVILEGED"),
		host:       shellEnv("SANDBOX_SSH_HOST", ""),
		port:       shellEnv("SANDBOX_SSH_PORT", "2222"),
		user:       shellEnv("SANDBOX_SSH_USER", "sandbox"),
		image:      shellEnv("SANDBOX_IMAGE", "golearn/sandbox:latest"),
		sessions:   make(map[string]time.Time),
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

// sessionName maps a (user, session key) pair to a container name. The key is
// the *lesson* ("l42") for course labs and a fixed word for standalone
// trainers ("git") — one container per lab, so every step of a lab and its
// checks share the same filesystem as the terminal the student is typing in.
func sessionName(userID int, key string) string {
	return fmt.Sprintf("gl-s-u%d-%s", userID, key)
}

// dindVolume names the per-session volume that backs the inner Docker daemon.
func dindVolume(userID int, key string) string {
	return fmt.Sprintf("gl-dind-u%d-%s", userID, key)
}

// needsDocker reports whether an image ships a container runtime of its own —
// Docker Engine or a k3s cluster. Those containers need --privileged, and their
// runtime state must live on a volume: on the overlay rootfs the fast storage
// drivers refuse to start.
func needsDocker(image string) bool {
	return strings.Contains(image, "sandbox-docker") || strings.Contains(image, "sandbox-k8s")
}

// runOpts returns the docker run flags for one kind of sandbox.
func runOpts(userID int, key, image string) string {
	switch {
	case strings.Contains(image, "sandbox-k8s"):
		// k3s keeps its state (containerd, images, etcd) under /var/lib/rancher
		// and wants a writable /run.
		return fmt.Sprintf(
			"--privileged --memory 4g --cpus 3 --pids-limit 4096 --tmpfs /run -v %s:/var/lib/rancher",
			dindVolume(userID, key))
	case strings.Contains(image, "sandbox-docker"):
		return fmt.Sprintf(
			"--privileged --memory 2g --cpus 2 --pids-limit 2048 -v %s:/var/lib/docker",
			dindVolume(userID, key))
	default:
		return "--memory 512m --cpus 1 --pids-limit 256"
	}
}

func (s *ShellRunner) touch(c string) {
	s.mu.Lock()
	s.sessions[c] = time.Now()
	s.mu.Unlock()
}

// ensure creates the per-(user,task) container if it is not running, applying
// the task setup script once. Returns the container name.
func (s *ShellRunner) ensure(ctx context.Context, userID int, key, image, setup string) (string, error) {
	if image == "" {
		image = s.image
	}
	if needsDocker(image) && !s.privileged {
		return "", fmt.Errorf("курсы Docker и Kubernetes на этом сервере отключены: " +
			"их лаборатории требуют привилегированной песочницы (SANDBOX_PRIVILEGED). " +
			"Проходить их можно в локальной установке")
	}
	c := sessionName(userID, key)
	out, _, err := s.run(ctx, fmt.Sprintf("docker inspect -f '{{.State.Running}}' %s 2>/dev/null", c))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "true" {
		s.touch(c)
		return c, nil
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(setup))
	// Courses with their own runtime (Docker, Kubernetes) need elevated
	// privileges and more headroom than a shell-only lab.
	opts := runOpts(userID, key, image)
	create := fmt.Sprintf(
		`docker rm -f %s >/dev/null 2>&1; `+
			`docker run -d --init --name %s --hostname sandbox --network none %s %s sleep infinity >/dev/null || exit 1; `+
			`if [ -n "%s" ]; then docker exec %s sh -c 'echo %s | base64 -d | bash' >/dev/null 2>&1; fi; `+
			`echo OK`,
		c, c, opts, image, b64, c, b64,
	)
	out, _, err = s.run(ctx, create)
	if err != nil {
		return "", err
	}
	if !strings.Contains(out, "OK") {
		msg := strings.TrimSpace(out)
		if strings.Contains(msg, "Unable to find image") || strings.Contains(msg, "No such image") ||
			strings.Contains(msg, "manifest unknown") || strings.Contains(msg, "pull access denied") {
			return "", fmt.Errorf("образ песочницы %s не собран на этом хосте — "+
				"собери его один раз: docker build -t %s -f deploy/sandbox/Dockerfile deploy/sandbox", image, image)
		}
		return "", fmt.Errorf("sandbox start failed: %s", msg)
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
func (s *ShellRunner) Exec(ctx context.Context, userID int, key, image, setup, command string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, execTimeout(image))
	defer cancel()
	c, err := s.ensure(ctx, userID, key, image, setup)
	if err != nil {
		return "", err
	}
	out, _, err := s.execIn(ctx, c, wrap(command))
	return out, err
}

// Check runs the task's check script in the session; passed = exit code 0.
func (s *ShellRunner) Check(ctx context.Context, userID int, key, image, setup, checkScript string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, execTimeout(image))
	defer cancel()
	c, err := s.ensure(ctx, userID, key, image, setup)
	if err != nil {
		return false, "", err
	}
	out, exit, err := s.execIn(ctx, c, wrap(checkScript))
	if err != nil {
		return false, out, err
	}
	return exit == 0, out, nil
}

// execTimeout picks the deadline for one command in this kind of sandbox.
func execTimeout(image string) time.Duration {
	if needsDocker(image) {
		return dockerExecTimeout
	}
	return shellExecTimeout
}

// Reset destroys the session container so the next command starts fresh.
func (s *ShellRunner) Reset(ctx context.Context, userID int, key string) error {
	ctx, cancel := context.WithTimeout(ctx, dockerExecTimeout)
	defer cancel()
	c := sessionName(userID, key)
	// The dind volume holds the inner daemon's images and containers, so a reset
	// has to drop it too — otherwise the "clean environment" still has the
	// student's old containers in it.
	_, _, err := s.run(ctx, fmt.Sprintf(
		"docker rm -f %s >/dev/null 2>&1; docker volume rm -f %s >/dev/null 2>&1; echo OK",
		c, dindVolume(userID, key)))
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
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			// The matching dind volume (if any) is named after the container.
			vol := strings.Replace(c, "gl-s-", "gl-dind-", 1)
			_, _, _ = s.run(ctx, fmt.Sprintf(
				"docker rm -f %s >/dev/null 2>&1; docker volume rm -f %s >/dev/null 2>&1", c, vol))
			cancel()
		}
	}
}

// shellBool reads a boolean-ish environment flag.
func shellBool(k string) bool {
	switch strings.ToLower(os.Getenv(k)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func shellEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ── Interactive PTY terminal (xterm.js over WebSocket) ──

// EnsureSession is a public wrapper to create/get the session container.
func (s *ShellRunner) EnsureSession(ctx context.Context, userID int, key, image, setup string) (string, error) {
	return s.ensure(ctx, userID, key, image, setup)
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
