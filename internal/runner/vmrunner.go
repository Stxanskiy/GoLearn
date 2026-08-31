package runner

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// VMRunner runs student labs inside per-(user,lesson) Firecracker micro-VMs on a
// dedicated host (the "FC host", e.g. berg), reached over SSH. Each VM has its own
// kernel, so Docker/k3s run natively inside — no dind, no --privileged, no host
// root — and the hypervisor boundary means one student deleting files or crashing
// their VM cannot touch another's. This mirrors devops404's lab model.
//
// Layout on the FC host (all under FC_DIR):
//
//	bin/firecracker                    the VMM binary
//	<kernel>                           guest kernel (docker-ready, boots with acpi=off)
//	<rootfs>                           golden read-only ext4 (systemd+root+sshd+dockerd)
//	vmkey                              private key the backend uses to SSH into a VM
//	sessions/<sid>/{boot.ext4,fc.json,fc.pid,fc.log}
//
// Networking: each VM gets a private /30 point-to-point link — host 172.31.<slot>.1,
// VM 172.31.<slot>.2 — on its own tap device (no bridge → VMs can't see each other;
// no NAT → the VM is offline, like devops404). The VM reads its address from the
// kernel cmdline (gl.ip=...) via the baked glnet service.
type VMRunner struct {
	enabled bool

	// SSH transport to the FC host.
	host, port, user, keyFile string

	// Paths on the FC host.
	dir     string // FC_DIR
	bin     string // firecracker binary
	kernel  string // guest kernel
	rootfs  string // golden ext4
	vmkey   string // key file (on the host) to reach a VM

	vcpus  int
	memMiB int
	maxVMs int

	mu       sync.Mutex
	sessions map[string]*vmSession // sid -> session
	freeSlot []bool                // slot in use?
}

type vmSession struct {
	sid     string
	userID  int
	key     string
	slot    int
	ip      string
	started time.Time
	last    time.Time
}

const (
	vmSSHTimeout = 20 * time.Second
	vmBootWait   = 40 * time.Second // budget for boot + sshd + setup
)

func NewVMRunner() *VMRunner {
	v := &VMRunner{
		host:   shellEnv("FC_SSH_HOST", ""),
		port:   shellEnv("FC_SSH_PORT", "22"),
		user:   shellEnv("FC_SSH_USER", "berg"),
		dir:    shellEnv("FC_DIR", "/home/berg/fc-build"),
		kernel: shellEnv("FC_KERNEL", "vmlinux-6.1.128-tot"),
		rootfs: shellEnv("FC_ROOTFS", "rootfs-docker.ext4"),
		vmkey:  shellEnv("FC_VMKEY", "vmkey"),
		vcpus:  atoiDefault(shellEnv("FC_VCPUS", "1"), 1),
		memMiB: atoiDefault(shellEnv("FC_MEM_MIB", "1024"), 1024),
		maxVMs: atoiDefault(shellEnv("FC_MAX_VMS", "8"), 8),
	}
	v.bin = shellEnv("FC_BIN", v.dir+"/bin/firecracker")
	if !shellBool("FC_ENABLED") || v.host == "" {
		return v // disabled
	}
	keySrc := shellEnv("FC_SSH_KEY", "")
	if keySrc == "" {
		return v
	}
	data, err := os.ReadFile(keySrc)
	if err != nil {
		return v
	}
	f, err := os.CreateTemp("", "fckey-*")
	if err != nil {
		return v
	}
	_, _ = f.Write(data)
	_ = f.Chmod(0o600)
	_ = f.Close()
	v.keyFile = f.Name()
	v.freeSlot = make([]bool, v.maxVMs)
	v.sessions = make(map[string]*vmSession)
	v.enabled = true
	go v.reaper()
	return v
}

func (v *VMRunner) Enabled() bool { return v != nil && v.enabled }

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return def
}

func vmSID(userID int, key string) string { return fmt.Sprintf("u%d-%s", userID, key) }

// runHost runs a shell script on the FC host over SSH and returns combined output.
func (v *VMRunner) runHost(ctx context.Context, script string) (string, int, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(script))
	remote := fmt.Sprintf("echo %s | base64 -d | bash", b64)
	args := []string{
		"-i", v.keyFile,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=8",
		"-o", "LogLevel=ERROR",
		"-p", v.port,
		v.user + "@" + v.host,
		remote,
	}
	cmd := exec.CommandContext(ctx, "ssh", args...)
	out, err := cmd.CombinedOutput()
	exit := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exit = ee.ExitCode()
		err = nil
	}
	return string(out), exit, err
}

// sshInto builds the command the FC host runs to reach a VM: ssh with the on-host
// vmkey to root@<ip>. `inner` is base64-decoded and piped to bash inside the VM.
func (v *VMRunner) sshInto(ip, inner string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(inner))
	return fmt.Sprintf(
		`ssh -i %s/%s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null `+
			`-o ConnectTimeout=5 -o LogLevel=ERROR root@%s `+
			`'echo %s | base64 -d | bash' 2>&1`,
		v.dir, v.vmkey, ip, b64)
}

func (v *VMRunner) touch(sid string) {
	v.mu.Lock()
	if s := v.sessions[sid]; s != nil {
		s.last = time.Now()
	}
	v.mu.Unlock()
}

// allocSlot reserves a free VM slot; -1 if the host is at capacity.
func (v *VMRunner) allocSlot() int {
	for i := range v.freeSlot {
		if !v.freeSlot[i] {
			v.freeSlot[i] = true
			return i
		}
	}
	return -1
}

// EnsureSession boots (or returns) the user's VM for this lesson and returns its IP.
// One VM per user at a time: any other VM the user has is torn down first.
func (v *VMRunner) EnsureSession(ctx context.Context, userID int, key, image, setup string) (string, error) {
	if !v.enabled {
		return "", fmt.Errorf("VM-песочница отключена")
	}
	sid := vmSID(userID, key)

	v.mu.Lock()
	if s := v.sessions[sid]; s != nil {
		ip := s.ip
		s.last = time.Now()
		v.mu.Unlock()
		// Confirm it is still reachable; if not, fall through and rebuild.
		if v.alive(ctx, ip) {
			return ip, nil
		}
		v.teardown(sid)
	} else {
		v.mu.Unlock()
	}

	// One-at-a-time: drop any other VM this user holds.
	v.mu.Lock()
	for other, s := range v.sessions {
		if s.userID == userID && other != sid {
			v.teardownLocked(other)
		}
	}
	slot := v.allocSlot()
	if slot < 0 {
		v.mu.Unlock()
		return "", fmt.Errorf("все VM-слоты заняты (лимит %d) — попробуй чуть позже", v.maxVMs)
	}
	sess := &vmSession{sid: sid, userID: userID, key: key, slot: slot, ip: vmIP(slot), started: time.Now(), last: time.Now()}
	v.sessions[sid] = sess
	v.mu.Unlock()

	if err := v.bootVM(ctx, sess, setup); err != nil {
		v.mu.Lock()
		v.teardownLocked(sid)
		v.mu.Unlock()
		return "", err
	}
	return sess.ip, nil
}

func vmHostIP(slot int) string { return fmt.Sprintf("172.31.%d.1", slot) }
func vmIP(slot int) string     { return fmt.Sprintf("172.31.%d.2", slot) }
func vmTap(slot int) string    { return fmt.Sprintf("gltap%d", slot) }
func vmMAC(slot int) string    { return fmt.Sprintf("AA:FC:00:00:00:%02x", slot&0xff) }

// bootVM provisions the tap, a per-session rootfs copy and fc.json, launches
// Firecracker, waits for SSH and applies the lesson setup once inside the VM.
func (v *VMRunner) bootVM(ctx context.Context, s *vmSession, setup string) error {
	ctx, cancel := context.WithTimeout(ctx, vmBootWait+30*time.Second)
	defer cancel()

	tap := vmTap(s.slot)
	hostIP := vmHostIP(s.slot)
	vmip := s.ip
	mac := vmMAC(s.slot)
	work := fmt.Sprintf("%s/sessions/%s", v.dir, s.sid)
	bootArgs := fmt.Sprintf(
		"console=ttyS0 reboot=k panic=1 acpi=off net.ifnames=0 gl.ip=%s/30 root=/dev/vda rw init=/sbin/init",
		vmip)
	setupB64 := base64.StdEncoding.EncodeToString([]byte(setup))

	script := fmt.Sprintf(`set -e
WORK=%[1]s
mkdir -p "$WORK"
# fresh tap owned by the SSH user so Firecracker can open it
sudo ip link del %[2]s 2>/dev/null || true
sudo ip tuntap add %[2]s mode tap user "$(whoami)"
sudo ip addr add %[3]s/30 dev %[2]s
sudo ip link set %[2]s up
# per-session writable rootfs (full copy for now; snapshot/CoW is a later step)
cp -f %[4]s/%[5]s "$WORK/boot.ext4"
# machine config
cat > "$WORK/fc.json" <<JSON
{
  "boot-source": { "kernel_image_path": "%[4]s/%[6]s", "boot_args": "%[7]s" },
  "drives": [ { "drive_id": "rootfs", "path_on_host": "$WORK/boot.ext4", "is_root_device": true, "is_read_only": false } ],
  "network-interfaces": [ { "iface_id": "eth0", "host_dev_name": "%[2]s", "guest_mac": "%[8]s" } ],
  "machine-config": { "vcpu_count": %[9]d, "mem_size_mib": %[10]d }
}
JSON
# launch
setsid %[4]s/bin/firecracker --no-api --config-file "$WORK/fc.json" > "$WORK/fc.log" 2>&1 &
echo $! > "$WORK/fc.pid"
# wait for sshd inside the VM
for i in $(seq 1 %[11]d); do
  sleep 1
  ssh -i %[4]s/%[12]s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=2 -o LogLevel=ERROR root@%[13]s true 2>/dev/null && { UP=1; break; }
done
[ "${UP:-0}" = 1 ] || { echo "GLVMERR boot-timeout"; tail -5 "$WORK/fc.log" 2>/dev/null; exit 0; }
# apply lesson setup once
if [ -n "%[14]s" ]; then
  ssh -i %[4]s/%[12]s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -o LogLevel=ERROR root@%[13]s 'echo %[14]s | base64 -d | bash' >/dev/null 2>&1 || true
fi
echo "GLVMOK %[13]s"
`,
		work,           // 1
		tap,            // 2
		hostIP,         // 3
		v.dir,          // 4
		v.rootfs,       // 5
		v.kernel,       // 6
		bootArgs,       // 7
		mac,            // 8
		v.vcpus,        // 9
		v.memMiB,       // 10
		int(vmBootWait/time.Second), // 11
		v.vmkey,        // 12
		vmip,           // 13
		setupB64,       // 14
	)

	out, _, err := v.runHost(ctx, script)
	if err != nil {
		return fmt.Errorf("VM host error: %w", err)
	}
	if !strings.Contains(out, "GLVMOK") {
		msg := strings.TrimSpace(out)
		if strings.Contains(msg, "boot-timeout") {
			return fmt.Errorf("VM не поднялась вовремя — попробуй открыть лабораторную заново")
		}
		return fmt.Errorf("VM start failed: %s", msg)
	}
	return nil
}

// alive reports whether the VM at ip answers SSH.
func (v *VMRunner) alive(ctx context.Context, ip string) bool {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	out, _, _ := v.runHost(ctx, v.sshInto(ip, "echo GLUP"))
	return strings.Contains(out, "GLUP")
}

// execVM runs a script inside the VM and returns combined output + exit code.
func (v *VMRunner) execVM(ctx context.Context, ip, script string) (string, int, error) {
	out, exit, err := v.runHost(ctx, v.sshInto(ip, script))
	if len(out) > maxShellOutput {
		out = out[:maxShellOutput] + "\n… (вывод обрезан)"
	}
	return out, exit, err
}

// Exec runs a user command in the session VM and returns combined output.
func (v *VMRunner) Exec(ctx context.Context, userID int, key, image, setup, command string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, image, setup)
	if err != nil {
		return "", err
	}
	out, _, err := v.execVM(ctx, ip, wrap(command))
	v.touch(vmSID(userID, key))
	return out, err
}

// Check runs the task's check script; passed = exit 0.
func (v *VMRunner) Check(ctx context.Context, userID int, key, image, setup, checkScript string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, image, setup)
	if err != nil {
		return false, "", err
	}
	// The exit code flows back through the chain: VM command -> inner ssh -> host
	// bash -> runHost, so a single exec gives us both output and verdict.
	out, exit, err := v.execVM(ctx, ip, wrap(checkScript))
	if err != nil {
		return false, out, err
	}
	v.touch(vmSID(userID, key))
	return exit == 0, out, nil
}

// Preview fetches one HTTP resource from a server the student started inside the VM.
func (v *VMRunner) Preview(ctx context.Context, userID int, key, image, setup string, port int, path string) ([]byte, string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, image, setup)
	if err != nil {
		return nil, "", 0, err
	}
	if port <= 0 {
		port = 80
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	script := fmt.Sprintf(`
url="http://127.0.0.1:%d%s"
if ! command -v curl >/dev/null 2>&1; then echo "GLPREVERR curl-missing"; exit 0; fi
if ! curl -s -m 8 -D /tmp/.glph -o /tmp/.glpb "$url" 2>/dev/null; then echo "GLPREVERR no-server"; exit 0; fi
code=$(head -1 /tmp/.glph 2>/dev/null | tr -d '\r' | awk '{print $2}')
ct=$(grep -i '^content-type:' /tmp/.glph 2>/dev/null | head -1 | tr -d '\r' | cut -d' ' -f2-)
echo "GLPREVIEW ${code:-200} ${ct:-text/html}"
base64 /tmp/.glpb 2>/dev/null
`, port, path)
	out, _, err := v.execVM(ctx, ip, script)
	if err != nil {
		return nil, "", 0, err
	}
	v.touch(vmSID(userID, key))
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

// ── In-VM file editor backend (Monaco), mirrors ShellRunner ──

func (v *VMRunner) fsExec(ctx context.Context, userID int, key, image, setup, script string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, image, setup)
	if err != nil {
		return "", err
	}
	out, _, err := v.execVM(ctx, ip, script)
	if err == nil {
		v.touch(vmSID(userID, key))
	}
	return out, err
}

func (v *VMRunner) FSList(ctx context.Context, userID int, key, image, setup, dir string) ([]FSEntry, error) {
	db := base64.StdEncoding.EncodeToString([]byte(dir))
	script := fmt.Sprintf(`d=$(printf %%s '%s' | base64 -d)
find "$d" -maxdepth 1 -mindepth 1 -printf '%%y\t%%f\n' 2>/dev/null | LC_ALL=C sort -t'\t' -k1,1 -k2,2`, db)
	out, err := v.fsExec(ctx, userID, key, image, setup, script)
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

func (v *VMRunner) FSRead(ctx context.Context, userID int, key, image, setup, file string) ([]byte, error) {
	fb := base64.StdEncoding.EncodeToString([]byte(file))
	script := fmt.Sprintf(`f=$(printf %%s '%s' | base64 -d)
[ -f "$f" ] && head -c 524288 "$f" | base64`, fb)
	out, err := v.fsExec(ctx, userID, key, image, setup, script)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.TrimSpace(out), "\n", ""))
}

func (v *VMRunner) FSWrite(ctx context.Context, userID int, key, image, setup, file string, content []byte) error {
	fb := base64.StdEncoding.EncodeToString([]byte(file))
	cb := base64.StdEncoding.EncodeToString(content)
	script := fmt.Sprintf(`f=$(printf %%s '%s' | base64 -d)
mkdir -p "$(dirname "$f")" && printf %%s '%s' | base64 -d > "$f" && echo GLOK`, fb, cb)
	out, err := v.fsExec(ctx, userID, key, image, setup, script)
	if err != nil {
		return err
	}
	if !strings.Contains(out, "GLOK") {
		return fmt.Errorf("запись не удалась: %s", strings.TrimSpace(out))
	}
	return nil
}

// Reset destroys the session VM so the next command starts fresh.
func (v *VMRunner) Reset(ctx context.Context, userID int, key string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.teardownLocked(vmSID(userID, key))
	return nil
}

// teardown tears a session down (caller must hold the lock via teardownLocked; this
// helper locks itself for external callers that already released it).
func (v *VMRunner) teardown(sid string) {
	v.mu.Lock()
	v.teardownLocked(sid)
	v.mu.Unlock()
}

// teardownLocked kills the VM, removes its tap and workdir, and frees the slot.
// The caller must hold v.mu.
func (v *VMRunner) teardownLocked(sid string) {
	s := v.sessions[sid]
	if s == nil {
		return
	}
	tap := vmTap(s.slot)
	work := fmt.Sprintf("%s/sessions/%s", v.dir, sid)
	script := fmt.Sprintf(`
[ -f %[1]s/fc.pid ] && kill "$(cat %[1]s/fc.pid)" 2>/dev/null || true
sudo ip link del %[2]s 2>/dev/null || true
rm -rf %[1]s
`, work, tap)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_, _, _ = v.runHost(ctx, script)
	}()
	if s.slot >= 0 && s.slot < len(v.freeSlot) {
		v.freeSlot[s.slot] = false
	}
	delete(v.sessions, sid)
}

func (v *VMRunner) reaper() {
	t := time.NewTicker(5 * time.Minute)
	defer t.Stop()
	for range t.C {
		v.mu.Lock()
		now := time.Now()
		var stale []string
		for sid, s := range v.sessions {
			if now.Sub(s.last) > shellSessionTTL || now.Sub(s.started) > shellSessionMax {
				stale = append(stale, sid)
			}
		}
		for _, sid := range stale {
			v.teardownLocked(sid)
		}
		v.mu.Unlock()
	}
}

// ── Interactive PTY into the VM (xterm.js over WebSocket) ──

// OpenPTY opens an interactive bash inside the VM (handle = the VM IP). The
// backend SSHes to the FC host, which SSHes into the VM with a PTY.
func (v *VMRunner) OpenPTY(handle string, cols, rows int) (*PTYSession, error) {
	if !v.enabled {
		return nil, fmt.Errorf("VM-песочница отключена")
	}
	sg, err := v.signer()
	if err != nil {
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            v.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(sg)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	client, err := ssh.Dial("tcp", v.host+":"+v.port, cfg)
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
	// Nested: FC host -> VM, with a PTY, running an interactive login shell.
	inner := fmt.Sprintf(
		"ssh -tt -i %s/%s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+
			"-o ConnectTimeout=5 -o LogLevel=ERROR root@%s "+
			"'cd /root; exec env TERM=xterm-256color HOME=/root bash -l'",
		v.dir, v.vmkey, handle)
	if err := session.Start(inner); err != nil {
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

func (v *VMRunner) signer() (ssh.Signer, error) {
	data, err := os.ReadFile(v.keyFile)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(data)
}
