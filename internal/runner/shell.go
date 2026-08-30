package runner

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
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
	// shellSessionTTL reaps a container after this much *idle* time (no exec,
	// no keystrokes) — cleans up abandoned tabs.
	shellSessionTTL = 30 * time.Minute
	// shellSessionMax is a hard cap from container start: even a busy session is
	// torn down after this, so nobody holds a sandbox forever.
	shellSessionMax = 1 * time.Hour
	maxShellOutput  = 64 * 1024
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
	sessions   map[string]time.Time // container -> last activity (idle TTL)
	started    map[string]time.Time // container -> first start (hard cap)
}

func NewShellRunner() *ShellRunner {
	s := &ShellRunner{
		privileged: shellBool("SANDBOX_PRIVILEGED"),
		host:       shellEnv("SANDBOX_SSH_HOST", ""),
		port:       shellEnv("SANDBOX_SSH_PORT", "2222"),
		user:       shellEnv("SANDBOX_SSH_USER", "sandbox"),
		image:      shellEnv("SANDBOX_IMAGE", "golearn/sandbox:latest"),
		sessions:   make(map[string]time.Time),
		started:    make(map[string]time.Time),
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
	out, _, err := s.run(ctx, fmt.Sprintf("docker inspect -f '{{.State.Running}}' %s 2>&1", c))
	if err != nil {
		return "", err
	}
	if dockerUnavailable(out) {
		return "", fmt.Errorf("Docker не запущен — песочница не может создать контейнер. " +
			"Запусти Docker Desktop (или `sudo systemctl start docker` на сервере) и открой лабораторную заново")
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
		if dockerUnavailable(msg) {
			return "", fmt.Errorf("Docker не запущен — песочница не может создать контейнер. " +
				"Запусти Docker Desktop (или `sudo systemctl start docker` на сервере) и открой лабораторную заново")
		}
		if strings.Contains(msg, "Unable to find image") || strings.Contains(msg, "No such image") ||
			strings.Contains(msg, "manifest unknown") || strings.Contains(msg, "pull access denied") {
			return "", fmt.Errorf("образ песочницы %s не собран на этом хосте — "+
				"собери его один раз: docker build -t %s -f deploy/sandbox/Dockerfile deploy/sandbox", image, image)
		}
		return "", fmt.Errorf("sandbox start failed: %s", msg)
	}
	s.touch(c)
	s.mu.Lock()
	s.started[c] = time.Now() // fresh container — start the hard-cap clock
	s.mu.Unlock()
	return c, nil
}

// dockerUnavailable reports whether the failure was the Docker daemon being
// unreachable rather than anything about the lab itself — by far the most
// common cause of "the terminal will not start", and worth saying plainly.
func dockerUnavailable(out string) bool {
	for _, marker := range []string{
		"Cannot connect to the Docker daemon",
		"docker daemon is not running",
		"failed to connect to the docker API",
		"Is the docker daemon running",
		"docker: command not found",
	} {
		if strings.Contains(out, marker) {
			return true
		}
	}
	return false
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

// Preview fetches one HTTP resource from a server the student started INSIDE the
// sandbox. The container has no network of its own (--network none), so we reach
// the server through `docker exec ... curl 127.0.0.1:<port>`. The body is
// base64'd in the container so binary assets (images, fonts) survive the trip,
// and the target URL is passed in base64 too, so a hostile path cannot break out
// of the shell command. Returns body, content-type and HTTP status.
func (s *ShellRunner) Preview(ctx context.Context, userID int, key, image, setup string, port int, path string) ([]byte, string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	c, err := s.ensure(ctx, userID, key, image, setup)
	if err != nil {
		return nil, "", 0, err
	}
	if port <= 0 {
		port = 80
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	urlB64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("http://127.0.0.1:%d%s", port, path)))
	script := fmt.Sprintf(`
url=$(printf %%s '%s' | base64 -d)
if ! command -v curl >/dev/null 2>&1; then echo "GLPREVERR curl-missing"; exit 0; fi
if ! curl -s -m 8 -D /tmp/.glph -o /tmp/.glpb "$url" 2>/dev/null; then echo "GLPREVERR no-server"; exit 0; fi
code=$(head -1 /tmp/.glph 2>/dev/null | tr -d '\r' | awk '{print $2}')
ct=$(grep -i '^content-type:' /tmp/.glph 2>/dev/null | head -1 | tr -d '\r' | cut -d' ' -f2-)
echo "GLPREVIEW ${code:-200} ${ct:-text/html}"
base64 /tmp/.glpb 2>/dev/null
`, urlB64)
	b64 := base64.StdEncoding.EncodeToString([]byte(script))
	remote := fmt.Sprintf(`docker exec %s sh -c 'echo %s | base64 -d | bash' 2>/dev/null`, c, b64)
	out, _, err := s.run(ctx, remote)
	if err != nil {
		return nil, "", 0, err
	}
	s.touch(c)
	nl := strings.IndexByte(out, '\n')
	if nl < 0 {
		return nil, "", 0, fmt.Errorf("preview: пустой ответ песочницы")
	}
	head := strings.TrimSpace(out[:nl])
	bodyB64 := strings.TrimSpace(out[nl+1:])
	if strings.HasPrefix(head, "GLPREVERR") {
		return nil, "", 0, fmt.Errorf("preview: %s", strings.TrimSpace(strings.TrimPrefix(head, "GLPREVERR")))
	}
	fields := strings.SplitN(strings.TrimPrefix(head, "GLPREVIEW "), " ", 2)
	status, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
	ct := "text/html"
	if len(fields) > 1 && strings.TrimSpace(fields[1]) != "" {
		ct = strings.TrimSpace(fields[1])
	}
	body, derr := base64.StdEncoding.DecodeString(strings.ReplaceAll(bodyB64, "\n", ""))
	if derr != nil {
		return nil, "", 0, fmt.Errorf("preview: не удалось раскодировать тело: %w", derr)
	}
	return body, ct, status, nil
}

// ── In-lab file editor (Monaco) backend ──
//
// The editor reads and writes files inside the sandbox over docker exec, the
// same channel the terminal uses (the container has no network). Paths are
// base64'd into the container so a hostile name cannot break out of the shell;
// the handler additionally jails every path under /root before calling here.

// FSEntry is one directory child for the editor's file tree.
type FSEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
}

func (s *ShellRunner) fsExec(ctx context.Context, userID int, key, image, setup, script string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	c, err := s.ensure(ctx, userID, key, image, setup)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(script))
	remote := fmt.Sprintf(`docker exec %s sh -c 'echo %s | base64 -d | bash' 2>/dev/null`, c, b64)
	out, _, err := s.run(ctx, remote)
	if err == nil {
		s.touch(c)
	}
	return out, err
}

// FSList returns the immediate children of dir (dirs first, then files).
func (s *ShellRunner) FSList(ctx context.Context, userID int, key, image, setup, dir string) ([]FSEntry, error) {
	db := base64.StdEncoding.EncodeToString([]byte(dir))
	script := fmt.Sprintf(`d=$(printf %%s '%s' | base64 -d)
find "$d" -maxdepth 1 -mindepth 1 -printf '%%y\t%%f\n' 2>/dev/null | LC_ALL=C sort -t'\t' -k1,1 -k2,2`, db)
	out, err := s.fsExec(ctx, userID, key, image, setup, script)
	if err != nil {
		return nil, err
	}
	var entries []FSEntry
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		entries = append(entries, FSEntry{Name: parts[1], Dir: parts[0] == "d"})
	}
	return entries, nil
}

// FSRead returns up to 512 KB of a file's content.
func (s *ShellRunner) FSRead(ctx context.Context, userID int, key, image, setup, file string) ([]byte, error) {
	fb := base64.StdEncoding.EncodeToString([]byte(file))
	script := fmt.Sprintf(`f=$(printf %%s '%s' | base64 -d)
[ -f "$f" ] && head -c 524288 "$f" | base64`, fb)
	out, err := s.fsExec(ctx, userID, key, image, setup, script)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.TrimSpace(out), "\n", ""))
}

// FSWrite creates/overwrites a file with content (parent dirs are created).
func (s *ShellRunner) FSWrite(ctx context.Context, userID int, key, image, setup, file string, content []byte) error {
	fb := base64.StdEncoding.EncodeToString([]byte(file))
	cb := base64.StdEncoding.EncodeToString(content)
	script := fmt.Sprintf(`f=$(printf %%s '%s' | base64 -d)
mkdir -p "$(dirname "$f")" && printf %%s '%s' | base64 -d > "$f" && echo GLOK`, fb, cb)
	out, err := s.fsExec(ctx, userID, key, image, setup, script)
	if err != nil {
		return err
	}
	if !strings.Contains(out, "GLOK") {
		return fmt.Errorf("запись не удалась: %s", strings.TrimSpace(out))
	}
	return nil
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
	delete(s.started, c)
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
			// Reap on idle timeout OR on the hard cap from first start.
			if now.Sub(last) > shellSessionTTL ||
				(!s.started[c].IsZero() && now.Sub(s.started[c]) > shellSessionMax) {
				stale = append(stale, c)
			}
		}
		for _, c := range stale {
			delete(s.sessions, c)
			delete(s.started, c)
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
