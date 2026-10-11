package runner

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"log/slog"

	"golang.org/x/crypto/ssh"

	"github.com/backendraz/golearn/internal/obs"
)

// VMRunner runs student labs inside per-(user,lesson) Firecracker micro-VMs on a
// dedicated host (the "FC host", e.g. berg), reached over SSH. Each VM has its own
// kernel, so Docker/k3s run natively inside — no dind, no --privileged, no host
// root — and the hypervisor boundary means one student deleting files or crashing
// their VM cannot touch another's.
//
// Layout on the FC host (all under FC_DIR):
//
//	bin/firecracker                    the VMM binary
//	<kernel>                           guest kernel (docker-ready, boots with acpi=off)
//	<rootfs>                           golden read-only ext4 (systemd+root+sshd+dockerd)
//	vmkey                              private key the backend uses to SSH into a VM
//	sessions/<sid>-<attempt>/{boot.ext4,fc.json,fc.pid,fc.log}
//
// Networking: each VM gets a private /30 point-to-point link — host 172.31.<slot>.1,
// VM 172.31.<slot>.2 — on its own tap device (no bridge → VMs can't see each other;
// no NAT → the VM is offline by design). The VM reads its address from the
// kernel cmdline (gl.ip=...) via the baked glnet service.
type VMRunner struct {
	enabled bool

	// SSH transport to the FC host.
	host, port, user, keyFile string

	// Paths on the FC host.
	dir       string // FC_DIR
	bin       string // firecracker binary
	kernel    string // guest kernel
	rootfs    string // golden ext4 (default profile: docker + all CLI tools)
	rootfsK8s string // Kubernetes profile golden (k3s auto-starts inside)
	vmkey     string // key file (on the host) to reach a VM

	vcpus   int
	memMiB  int // docker profile: an engine runs inside the guest
	memLite int // shell-only lessons (Linux, Git, SQL, trainers)
	memK8s  int // Kubernetes VMs need more RAM (k3s)
	// Ceilings for what a lesson may ask for. FC_MAX_VMS of these run at once,
	// so the cap is what keeps one lab definition from exhausting the host.
	maxMemMiB int
	maxVCPUs  int
	maxVMs    int

	mu       sync.Mutex
	sessions map[string]*vmSession // sid -> session
	booting  map[string]*vmBoot    // sid -> boot in flight (see EnsureSession)
	// bringUpHook stands in for bringUp so the coordination in EnsureSession can
	// be tested without a Firecracker host. nil everywhere but in tests.
	bringUpHook func(ctx context.Context, sid string, userID int, key string, spec Spec) (string, error)
	// killHook stands in for the host side of killVM for the same reason. The
	// bookkeeping around it still runs, which is the part worth testing.
	killHook func(*vmSession)
	freeSlot []bool                  // slot in use?
	pool     map[string][]*vmSession // profile -> pre-booted warm VMs
	warmWant map[string]int          // profile -> desired warm count (0 = disabled)
	refill   chan struct{}           // nudge the pool manager to top up
	swept    chan struct{}           // closed once sweepOrphans has finished
}

// vmStartFailed counts bring-ups that did not produce a usable VM. Until this
// existed, "иногда не запускается" could not be answered at all: there was no
// number for how often, and no reason attached to the ones that did fail.
var vmStartFailed = obs.Counter(
	"golearn_vm_start_failures_total",
	"Micro-VM bring-ups that failed, by reason.",
	"reason")

// bootGen numbers boot attempts so each gets its own work directory. killVM
// deletes asynchronously, so without this a teardown scheduled by a failed
// attempt lands on top of the directory the *next* attempt has just created:
// "cp: cannot create '/opt/fc/sessions/warm-docker-0/boot.ext4'". The warm pool
// retries every 400ms, so it span on that for nine attempts before one survived.
var bootGen atomic.Uint64

// vmBoot is one in-flight bring-up, shared by everyone who asks for the same
// session while it is still coming up.
type vmBoot struct {
	done   chan struct{}
	cancel context.CancelFunc // stops a boot the same user has superseded
	ip     string
	err    error
}

type vmSession struct {
	sid     string
	userID  int
	key     string
	image   string
	profile string // "docker" | "k8s" — which golden this VM booted from
	slot    int
	ip      string
	// What the lesson asked for, 0 when it asked for nothing. Kept on the
	// session because bootVM sizes the machine and only sees this struct.
	reqCPUs, reqMemMiB int
	// What the machine actually booted with. A warm VM is sized before any
	// lesson is known, so this is what decides whether it can be handed to a
	// lesson that asked for something particular.
	bootCPUs, bootMemMiB int
	work                 string // host work dir (stored so a re-keyed warm VM tears down correctly)
	tap                  string
	started              time.Time
	last                 time.Time
}

// profileOf returns the pool profile for a sandbox image.
// profileOf maps a lesson's sandbox image onto a VM profile. The profile decides
// the golden rootfs and, just as importantly, the VM size: "lite" lessons (Linux,
// Git, SQL, the trainers) only run shell tools, while "docker" and "k8s" carry an
// engine or a cluster inside the guest.
// clampInt keeps a requested size inside what the host can afford.
func clampInt(v, lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func profileOf(image string) string {
	switch {
	case strings.Contains(image, "sandbox-k8s"):
		return "k8s"
	case strings.Contains(image, "sandbox-docker"), strings.Contains(image, "sandbox-pg"):
		// SQL lessons start a PostgreSQL server inside the guest, so they keep the
		// larger size even though they carry no engine.
		return "docker"
	default:
		return "lite"
	}
}

const (
	// vmMinMemMiB is below what the golden rootfs boots in; asking for less is a
	// typo, and honouring it would hang the lab instead of failing it.
	vmMinMemMiB  = 256
	vmSSHTimeout = 20 * time.Second
	vmBootWait   = 40 * time.Second // budget for boot + sshd + setup
	// vmK8sWait is how long the boot script waits for k3s to register its node
	// Ready. A Kubernetes lab whose terminal opens onto "connection refused" is
	// worse than one that takes longer to open, so the script waits — and the
	// deadline around it has to allow for that wait.
	vmK8sWait = 45 * time.Second
)

func NewVMRunner() *VMRunner {
	v := &VMRunner{
		host:      shellEnv("FC_SSH_HOST", ""),
		port:      shellEnv("FC_SSH_PORT", "22"),
		user:      shellEnv("FC_SSH_USER", "glvm"),
		dir:       shellEnv("FC_DIR", "/opt/fc"),
		kernel:    shellEnv("FC_KERNEL", "vmlinux-6.1.128-tot"),
		rootfs:    shellEnv("FC_ROOTFS", "rootfs-docker.ext4"),
		rootfsK8s: shellEnv("FC_ROOTFS_K8S", "rootfs-k8s.ext4"),
		vmkey:     shellEnv("FC_VMKEY", "vmkey"),
		vcpus:     atoiDefault(shellEnv("FC_VCPUS", "1"), 1),
		memMiB:    atoiDefault(shellEnv("FC_MEM_MIB", "1024"), 1024),
		memLite:   atoiDefault(shellEnv("FC_MEM_LITE", "512"), 512),
		memK8s:    atoiDefault(shellEnv("FC_MEM_K8S", "1536"), 1536),
		maxMemMiB: atoiDefault(shellEnv("FC_MAX_MEM_MIB", "4096"), 4096),
		maxVCPUs:  atoiDefault(shellEnv("FC_MAX_VCPUS", "2"), 2),
		maxVMs:    atoiDefault(shellEnv("FC_MAX_VMS", "8"), 8),
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
	v.booting = make(map[string]*vmBoot)
	v.pool = map[string][]*vmSession{}
	v.warmWant = map[string]int{
		"lite":   atoiDefault(shellEnv("FC_WARM_LITE", "0"), 0),
		"docker": atoiDefault(shellEnv("FC_WARM_DOCKER", "0"), 0),
		"k8s":    atoiDefault(shellEnv("FC_WARM_K8S", "0"), 0),
	}
	v.refill = make(chan struct{}, 1)
	v.swept = make(chan struct{})
	v.enabled = true
	// sweepOrphans wipes sessions/* wholesale. Starting the warm pool alongside it
	// meant the first VM's work directory was deleted mid-boot — production logged
	// exactly that on the first redeploy after the pool was enabled. Nothing boots
	// until the sweep is done.
	go func() {
		v.sweepOrphans()
		close(v.swept)
		if v.warmWant["lite"] > 0 || v.warmWant["docker"] > 0 || v.warmWant["k8s"] > 0 {
			v.poolManager()
		}
	}()
	go v.reaper()
	return v
}

// sweepOrphans clears VMs left behind by a previous run of the app (a redeploy or
// crash loses the in-memory session map, but the Firecracker processes, taps and
// per-session dirs survive on the host). golearn is the only Firecracker user on
// the FC host, so a blanket sweep at startup is safe and prevents leaks. Runs once.
func (v *VMRunner) sweepOrphans() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	script := `pkill -u "$(whoami)" -f 'bin/firecracker --no-api' 2>/dev/null || true
for t in $(ip -br link 2>/dev/null | awk '/^gltap/{print $1}'); do sudo /usr/local/sbin/gl-tap del "$t" 2>/dev/null || true; done
rm -rf ` + v.dir + `/sessions/* 2>/dev/null || true
echo GLSWEEP`
	_, _, _ = v.runHost(ctx, script)
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
	exit, err := classifyExit(ctx, out, err)
	return string(out), exit, err
}

// classifyExit separates "the script ran and said no" from "the script never
// ran". Both used to come back as a plain exit code with a nil error, which is
// how a dead FC host, a killed container or an expired deadline reached the
// student as "task not solved" — and got written to their submissions as a
// failed attempt against a correct answer.
//
// ssh reserves 255 for its own failures. A check script could in principle exit
// 255 too; treating that as a transport error is the safe way round, because
// the cost of the two mistakes is not symmetric: a retry prompt costs a click,
// a false "not solved" costs the student their trust in the checker.
func classifyExit(ctx context.Context, out []byte, err error) (int, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return -1, fmt.Errorf("sandbox command did not finish: %w", ctxErr)
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		// nil, or something that is not an exit status at all (ssh missing, fork
		// failure) — both are ours, not the student's.
		return 0, err
	}
	switch code := ee.ExitCode(); {
	case code < 0:
		return code, fmt.Errorf("sandbox command was killed before it finished")
	case code == 255:
		return code, fmt.Errorf("could not reach the sandbox: %s", firstLine(out))
	default:
		return code, nil
	}
}

// classifyLocalExit is classifyExit for a command run directly rather than over
// ssh. 255 carries no special meaning for a local script, so only a killed
// process and an expired deadline count as failures to run.
func classifyLocalExit(ctx context.Context, err error) (int, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return -1, fmt.Errorf("sandbox command did not finish: %w", ctxErr)
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		return 0, err
	}
	if code := ee.ExitCode(); code < 0 {
		return code, fmt.Errorf("sandbox command was killed before it finished")
	}
	return ee.ExitCode(), nil
}

// firstLine keeps an ssh diagnostic short enough to log.
func firstLine(out []byte) string {
	s := strings.TrimSpace(string(out))
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	if s == "" {
		s = "no output"
	}
	return s
}

// runHostStdin is runHost with stdin streamed to the script (the script is not piped through bash's stdin).
func (v *VMRunner) runHostStdin(ctx context.Context, script string, stdin io.Reader) (string, int, error) {
	b64 := base64.StdEncoding.EncodeToString([]byte(script))
	remote := fmt.Sprintf(`bash -c "$(echo %s | base64 -d)"`, b64)
	cmd := exec.CommandContext(ctx, "ssh",
		"-i", v.keyFile,
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ConnectTimeout=8",
		"-o", "LogLevel=ERROR",
		"-p", v.port,
		v.user+"@"+v.host,
		remote,
	)
	cmd.Stdin = stdin
	out, err := cmd.CombinedOutput()
	exit, err := classifyExit(ctx, out, err)
	return string(out), exit, err
}

// sshIntoStdin is sshInto without -n, so the VM script reads the host command's stdin.
func (v *VMRunner) sshIntoStdin(ip, inner string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(inner))
	return fmt.Sprintf(
		`ssh -i %s/%s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null `+
			`-o ConnectTimeout=5 -o LogLevel=ERROR root@%s `+
			`"$(echo %s | base64 -d)" 2>&1`,
		v.dir, v.vmkey, ip, b64)
}

// sshInto builds the command the FC host runs to reach a VM: ssh with the on-host
// vmkey to root@<ip>. `inner` is base64-decoded and piped to bash inside the VM.
func (v *VMRunner) sshInto(ip, inner string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(inner))
	// -n is essential: this whole command is fed to the host bash over a pipe, and
	// an ssh without -n would read the remaining piped script as its stdin.
	return fmt.Sprintf(
		`ssh -n -i %s/%s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null `+
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

// SessionInfo describes a live sandbox session's lifetime.
type SessionInfo struct {
	Started   time.Time
	ExpiresAt time.Time // earliest of idle timeout and hard limit
}

// Session returns lifetime info of the user's session for key, if one is alive.
func (v *VMRunner) Session(userID int, key string) (SessionInfo, bool) {
	if !v.Enabled() {
		return SessionInfo{}, false
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	s := v.sessions[vmSID(userID, key)]
	if s == nil {
		return SessionInfo{}, false
	}
	exp := s.started.Add(shellSessionMax)
	if idle := s.last.Add(shellSessionTTL); idle.Before(exp) {
		exp = idle
	}
	return SessionInfo{Started: s.started, ExpiresAt: exp}, true
}

// Touch marks the user's session as active (e.g. on terminal keystrokes).
func (v *VMRunner) Touch(userID int, key string) {
	if v.Enabled() {
		v.touch(vmSID(userID, key))
	}
}

// HasSession reports whether a VM for this user+lesson is currently alive, so the
// frontend can auto-reconnect after a page reload instead of losing the environment.
func (v *VMRunner) HasSession(userID int, key string) bool {
	if !v.enabled {
		return false
	}
	sid := vmSID(userID, key)
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.sessions[sid] != nil
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
func (v *VMRunner) EnsureSession(ctx context.Context, userID int, key string, spec Spec) (string, error) {
	if !v.enabled {
		return "", fmt.Errorf("VM-песочница отключена")
	}
	sid := vmSID(userID, key)

	for {
		v.mu.Lock()
		// Someone is already bringing this session up — wait for them.
		//
		// Without this, a second request during the 15-25s a VM takes to boot found
		// the session, asked alive(), got false because sshd was not up yet, and
		// tore the VM down — deleting the work dir out from under the boot script
		// that was still running. Both requests then failed with "No such file or
		// directory", and the student saw "не удалось запустить". Opening the
		// terminal and pressing "Проверить" was enough to trigger it.
		if b := v.booting[sid]; b != nil {
			v.mu.Unlock()
			select {
			case <-b.done:
			case <-ctx.Done():
				return "", ctx.Err()
			}
			return b.ip, b.err
		}
		if s := v.sessions[sid]; s != nil {
			ip := s.ip
			s.last = time.Now()
			v.mu.Unlock()
			// Confirm it is still reachable; if not, drop it and look again — by then
			// another request may have started a boot we should wait for.
			if v.alive(ctx, ip) {
				return ip, nil
			}
			v.teardown(sid)
			continue
		}
		bootCtx, cancel := context.WithCancel(ctx)
		b := &vmBoot{done: make(chan struct{}), cancel: cancel}
		v.booting[sid] = b
		v.mu.Unlock()

		up := v.bringUpHook
		if up == nil {
			up = v.bringUp
		}
		b.ip, b.err = up(bootCtx, sid, userID, key, spec)
		cancel()
		close(b.done)
		v.mu.Lock()
		delete(v.booting, sid)
		v.mu.Unlock()
		return b.ip, b.err
	}
}

// bringUp creates the session: it frees whatever else the user holds, takes a
// warm VM if one matches, and otherwise boots a fresh one. Only ever called with
// this sid claimed in v.booting, so it cannot race another bring-up of its own.
func (v *VMRunner) bringUp(ctx context.Context, sid string, userID int, key string, spec Spec) (string, error) {
	// A student opening a lab in the first seconds after a deploy would otherwise
	// race the startup sweep, which deletes sessions/* wholesale.
	if v.swept != nil {
		select {
		case <-v.swept:
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	profile := profileOf(spec.Image)

	// One-at-a-time: drop any other VM this user holds, then try the warm pool.
	//
	// A session that is still booting is cancelled and then *waited for* before it
	// is torn down. Cancelling alone was not enough: teardown ran while the boot
	// was still mid-flight, so it deleted the work dir and the tap out from under
	// a script that went on using them, and whatever that script created
	// afterwards was left behind with nobody tracking it.
	//
	// Cancel-then-wait cannot deadlock the way plain waiting would: the other side
	// is cancelled first, so its boot returns instead of blocking. The deadline is
	// only a backstop for a boot wedged somewhere uncancellable.
	v.dropOtherSessions(ctx, sid, userID)

	v.mu.Lock()
	// A pre-booted warm VM of the right profile → hand it out instantly, then just
	// apply the lesson setup (no cp, no boot, no k3s wait).
	// A warm VM is pre-booted long before the lesson is known, so it carries the
	// profile's default size. Handing it to a lesson that asked for a different
	// one silently ignored the author's choice: the sandbox size set in the
	// studio did nothing whenever the pool happened to have a VM ready, which is
	// most of the time. Only a warm VM of the right size is handed out now.
	wantCPUs, wantMem := v.sizeFor(profile, spec)
	if warm := v.takeWarmLocked(profile, wantCPUs, wantMem); warm != nil {
		warm.sid, warm.userID, warm.key, warm.image = sid, userID, key, spec.Image
		warm.started, warm.last = time.Now(), time.Now()
		v.sessions[sid] = warm
		v.mu.Unlock()
		v.signalRefill()
		if err := v.applySetup(ctx, warm, spec.Setup); err != nil {
			v.teardown(sid)
			return "", err
		}
		return warm.ip, nil
	}
	slot := v.allocSlot()
	if slot < 0 {
		v.mu.Unlock()
		vmStartFailed("slots")
		return "", fmt.Errorf("все VM-слоты заняты (лимит %d) — попробуй чуть позже", v.maxVMs)
	}
	// Каталог и tap известны до начала загрузки и проставляются здесь, под
	// замком: снос сессии в любой момент после этой строки знает, что убирать.
	// Свой каталог на попытку (bootGen) — чтобы отменённая попытка не удалила
	// каталог следующей.
	sess := &vmSession{
		sid: sid, userID: userID, key: key, image: spec.Image,
		reqCPUs: spec.CPUs, reqMemMiB: spec.MemMiB,
		slot: slot, ip: vmIP(slot), profile: profile,
		tap:     vmTap(slot),
		work:    fmt.Sprintf("%s/sessions/%s-%d", v.dir, sid, bootGen.Add(1)),
		started: time.Now(), last: time.Now(),
	}
	v.sessions[sid] = sess
	v.mu.Unlock()

	if err := v.bootVM(ctx, sess, spec.Setup); err != nil {
		v.mu.Lock()
		v.teardownLocked(sid)
		v.mu.Unlock()
		return "", err
	}
	v.signalRefill()
	return sess.ip, nil
}

// dropOtherSessions releases every VM this user holds except sid.
//
// Split out of bringUp because it has to let go of v.mu while it waits: holding
// the lock across the wait would block every other request on the runner.
func (v *VMRunner) dropOtherSessions(ctx context.Context, sid string, userID int) {
	for {
		v.mu.Lock()
		var other string
		var boot *vmBoot
		for cand, s := range v.sessions {
			if s.userID == userID && cand != sid {
				other, boot = cand, v.booting[cand]
				break
			}
		}
		if other == "" {
			v.mu.Unlock()
			return
		}
		// Nothing in flight: tear it down right here, still under the lock.
		if boot == nil {
			v.teardownLocked(other)
			v.mu.Unlock()
			continue
		}
		if boot.cancel != nil {
			boot.cancel()
		}
		v.mu.Unlock()

		wait, stop := context.WithTimeout(ctx, vmBootWait)
		select {
		case <-boot.done:
		case <-wait.Done():
			// The boot did not return in time. Tearing down anyway is still better
			// than leaking the slot — and the boot script now cleans up after itself
			// on the host, so what it created does not survive either.
			slog.Warn("VM: superseded boot did not finish in time, tearing down anyway",
				"sid", other, "user", userID)
		}
		stop()

		v.mu.Lock()
		v.teardownLocked(other)
		v.mu.Unlock()
	}
}

func vmHostIP(slot int) string { return fmt.Sprintf("172.31.%d.1", slot) }
func vmIP(slot int) string     { return fmt.Sprintf("172.31.%d.2", slot) }
func vmTap(slot int) string    { return fmt.Sprintf("gltap%d", slot) }
func vmMAC(slot int) string    { return fmt.Sprintf("AA:FC:00:00:00:%02x", slot&0xff) }

// bootVM provisions the tap, a per-session rootfs copy and fc.json, launches
// Firecracker, waits for SSH and applies the lesson setup once inside the VM.
func (v *VMRunner) bootVM(ctx context.Context, s *vmSession, setup string) error {
	// The deadline has to be larger than everything the script may legitimately
	// wait for, or it kills a boot that was going to succeed. It was
	// vmBootWait+30s = 70s while a Kubernetes boot can spend 40s waiting for
	// sshd and another 45 for k3s before the setup even starts — so a cold k8s
	// lab could not finish inside its own deadline. It almost never had to: the
	// warm pool answered first, and the bug stayed hidden until a lesson asked
	// for a size the pool did not have.
	budget := vmBootWait + 30*time.Second
	v.mu.Lock()
	isK8s := s.profile == "k8s"
	v.mu.Unlock()
	if isK8s {
		budget += vmK8sWait
	}
	ctx, cancel := context.WithTimeout(ctx, budget)
	defer cancel()

	// work/tap/profile уже заполнены под v.mu тем, кто создал сессию (bringUp или
	// bootWarm), и здесь только читаются.
	//
	// Раньше их писали здесь, без замка, а killVM читал под замком — настоящая
	// гонка по полям vmSession. Хуже того, между созданием сессии и этой строкой
	// существовало окно: снос сессии внутри него видел s.work == "" и вырождался
	// в скрипт без единого действия, поэтому процесс Firecracker оставался жить,
	// а копия rootfs на 5 ГБ — лежать на диске.
	v.mu.Lock()
	tap, work, vmip, profile := s.tap, s.work, s.ip, s.profile
	v.mu.Unlock()
	hostIP := vmHostIP(s.slot)
	mac := vmMAC(s.slot)
	if work == "" || tap == "" {
		return fmt.Errorf("bootVM: сессия %s создана без work/tap", s.sid)
	}

	// Kubernetes lessons get the k3s golden and a bigger VM; everything else uses
	// the default docker+tools golden.
	rootfs, mem, vcpus := v.rootfs, v.memMiB, v.vcpus
	// k8sWait (empty for non-k8s) blocks the boot until k3s registers the node as
	// Ready, so the student's terminal opens onto a working cluster instead of a
	// "connection refused" while k3s (~15s) is still coming up.
	k8sWait := ""
	if profile == "lite" {
		// Shell-only lesson: same golden (it carries the CLI tools) but none of the
		// memory an engine or a cluster would need.
		mem = v.memLite
	}
	if profile == "k8s" {
		// k3s fits comfortably in ~1.5 GB / 1 vCPU (a single node needs about 1 GB); the
		// old 3 GB / 2 vCPU was ~3x too generous.
		rootfs, mem, vcpus = v.rootfsK8s, v.memK8s, 1
		// The same number the deadline above budgets for, so the script cannot
		// outlast the context that is waiting on it.
		k8sWait = fmt.Sprintf(
			`echo "waiting for k3s..."; for i in $(seq 1 %[4]d); do `+
				`ssh -n -i %[1]s/%[2]s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=3 -o LogLevel=ERROR root@%[3]s `+
				`'kubectl get nodes 2>/dev/null | grep -q " Ready"' && break; sleep 1; done`,
			v.dir, v.vmkey, vmip, int(vmK8sWait/time.Second))
	}
	// A lesson may ask for its own size. The profile picked the golden rootfs
	// above and that stays - the image decides what is installed, which cannot be
	// changed at boot - but memory and CPU are per-VM numbers Firecracker takes
	// straight from the machine config, so an author can say what their lab needs
	// instead of living with one of three sizes.
	//
	// sizeFor does the clamping, and the warm pool asks it the same question
	// before handing a VM over: if these two disagreed, a lesson would get a
	// machine of one size while the pool believed it had given another.
	vcpus, mem = v.sizeFor(profile, Spec{CPUs: s.reqCPUs, MemMiB: s.reqMemMiB})
	// Recorded so the pool can tell what this machine actually is.
	v.mu.Lock()
	s.bootCPUs, s.bootMemMiB = vcpus, mem
	v.mu.Unlock()

	// nopv is what keeps an idle VM from burning a full host core.
	//
	// The cause was never ACPI. With KVM paravirtualisation on, the guest never
	// executes HLT at all — measured: halt_exits stays at 0 while the fc_vcpu
	// thread takes 503 ticks of every 500, and the guest's own /proc/stat reports
	// 100% idle. Turning paravirt off drops that to 2 ticks of 500, on this very
	// kernel; a Kubernetes VM idles at 6% instead of 100%, k3s and Docker
	// unaffected. idle=halt alone does nothing, and neither does lapic.
	//
	// What the guest gives up is kvmclock (it falls back to TSC — sleep(5) lands
	// within 2ms), PV EOI, and the paravirt spinlock and TLB paths, which cost a
	// one-vCPU VM nothing.
	//
	// acpi=off stays. Firecracker does publish ACPI tables, but with them enabled
	// virtio_blk and virtio_net fail to probe with -22 and the guest panics on
	// "VFS: Unable to mount root fs" — on 6.1 and on 6.12 alike, with or without
	// swiotlb/iommu/x2apic tweaks. Nothing is lost by leaving it off now that the
	// idle problem is solved elsewhere.
	bootArgs := fmt.Sprintf(
		"console=ttyS0 reboot=k panic=1 acpi=off idle=halt nopv net.ifnames=0 gl.ip=%s/30 root=/dev/vda rw init=/sbin/init",
		vmip)

	script := v.bootScript(setup, rootfs, bootArgs, mac, vcpus, mem, vmip, hostIP, tap, work, k8sWait)

	out, _, err := v.runHost(ctx, script)
	if err != nil {
		// A deadline that fires is not a host fault, and saying "host" sent every
		// slow boot to the wrong place: the metric blamed the FC box and the log
		// line carried an ssh exit status instead of the one fact that mattered.
		// The script can still outlast this budget — a lesson setup does its own
		// waiting (see k8sBoot in cmd/seed), which we cannot see from here — so
		// this is the one honest thing to report when it does.
		reason, detail := "host", err.Error()
		if ctx.Err() != nil {
			reason, detail = "boot-timeout", fmt.Sprintf("boot did not finish in %s", budget)
		}
		vmStartFailed(reason)
		slog.Error("VM boot: host error", "sid", s.sid, "slot", s.slot, "reason", reason, "err", err)
		return &BootError{Reason: reason, Detail: detail}
	}
	if !strings.Contains(out, "GLVMOK") {
		msg := strings.TrimSpace(out)
		reason := "boot"
		switch {
		case strings.Contains(msg, "boot-timeout"):
			reason = "boot-timeout"
		case strings.Contains(msg, "GLVMERR setup"):
			reason = "setup"
		}
		vmStartFailed(reason)
		// Kept out of the student's error on purpose: this is host paths and shell
		// output. It used to be shown verbatim, which is how "/opt/fc/sessions/…"
		// ended up on screen. It rides along on the error so an admin — and only
		// an admin — can be shown it instead of having to read the pod's log.
		slog.Error("VM boot failed", "sid", s.sid, "slot", s.slot, "profile", s.profile, "detail", tail(msg))
		return &BootError{Reason: reason, Detail: tail(msg)}
	}
	return nil
}

// bootScript renders the host script. Split out so its shape can be tested
// without a host: the order of its steps matters and is not visible from the
// outside.
func (v *VMRunner) bootScript(setup, rootfs, bootArgs, mac string, vcpus, mem int, vmip, hostIP, tap, work, k8sWait string) string {
	setupB64 := base64.StdEncoding.EncodeToString([]byte(setup))
	return fmt.Sprintf(`set -e
WORK=%[1]s
# Скрипт убирает за собой сам, если умрёт или будет убит на любом шаге.
#
# exec.CommandContext убивает локальный ssh, но не удалённый bash: тот продолжал
# копировать 5 ГБ и поднимать Firecracker уже после того, как слот отдали
# следующему студенту, и оставлял и процесс, и tap, и копию rootfs. Теперь любой
# выход, кроме успешного, разбирает за собой на той стороне, не полагаясь на
# нашу. Обе операции идемпотентны, так что повторная уборка из killVM безвредна.
cleanup_boot() {
  if [ -f "$WORK/fc.pid" ]; then kill "$(cat "$WORK/fc.pid")" 2>/dev/null || true; fi
  sudo /usr/local/sbin/gl-tap del %[2]s 2>/dev/null || true
  rm -rf "$WORK" 2>/dev/null || true
}
trap cleanup_boot EXIT HUP TERM INT
mkdir -p "$WORK"
# fresh tap via the restricted gl-tap wrapper (the only privileged op we may do)
sudo /usr/local/sbin/gl-tap add %[2]s %[3]s
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
# wait for sshd inside the VM. -n on every ssh: this script is piped into the host
# bash, so an ssh without -n would swallow the rest of the script as its stdin.
for i in $(seq 1 %[11]d); do
  sleep 1
  ssh -n -i %[4]s/%[12]s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=2 -o LogLevel=ERROR root@%[13]s true 2>/dev/null && { UP=1; break; }
done
[ "${UP:-0}" = 1 ] || { echo "GLVMERR boot-timeout"; tail -5 "$WORK/fc.log" 2>/dev/null; exit 0; }
# Ждём кластер ДО setup, а не после. Раньше было наоборот, и для урока по
# Kubernetes это означало, что setup стартовал через пару секунд после sshd —
# когда k3s ещё поднимается. Setup ждал его сам, своей минутой, и если не
# укладывался, вся песочница объявлялась незапустившейся: «k3s did not become
# ready in 60s». Ожидание после setup не помогало никому.
%[15]s
# apply lesson setup once
if [ -n "%[14]s" ]; then
  if ! setup_out=$(ssh -n -i %[4]s/%[12]s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o ConnectTimeout=5 -o LogLevel=ERROR root@%[13]s 'echo %[14]s | base64 -d | bash' 2>&1); then
    echo "GLVMERR setup"; echo "$setup_out" | tail -5; exit 0
  fi
fi
# Дошли до конца — разоружаем уборку, VM остаётся жить.
trap - EXIT HUP TERM INT
echo "GLVMOK %[13]s"
`,
		work,                        // 1
		tap,                         // 2
		hostIP,                      // 3
		v.dir,                       // 4
		rootfs,                      // 5
		v.kernel,                    // 6
		bootArgs,                    // 7
		mac,                         // 8
		vcpus,                       // 9
		mem,                         // 10
		int(vmBootWait/time.Second), // 11
		v.vmkey,                     // 12
		vmip,                        // 13
		setupB64,                    // 14
		k8sWait,                     // 15
	)
}

// BootError is a bring-up that did not produce a usable VM.
//
// Its Error() is the one line a student may see; Diagnostic() is the host's
// own output, which names the failing command and is the difference between
// "try again" and knowing that a setup script exited non-zero. Callers decide
// who is allowed to see which.
type BootError struct {
	// Reason is the coarse category, matching golearn_vm_start_failures_total:
	// host, boot, boot-timeout or setup.
	Reason string
	Detail string
}

func (e *BootError) Error() string { return errStudentBoot.Error() }

// Is makes errors.Is(err, errStudentBoot) keep working for anything that was
// comparing against the sentinel.
func (e *BootError) Is(target error) bool { return target == errStudentBoot }

// Diagnostic reports what actually went wrong, for an operator.
func (e *BootError) Diagnostic() string {
	if e.Detail == "" {
		return e.Reason
	}
	return e.Reason + ": " + e.Detail
}

// errStudentBoot is the one message a failed boot shows; the reason is logged and
// counted in golearn_vm_start_failures_total.
var errStudentBoot = errors.New("песочница не запустилась — попробуй открыть лабораторную заново")

// alive reports whether the VM at ip answers SSH.
func (v *VMRunner) alive(ctx context.Context, ip string) bool {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	out, _, _ := v.runHost(ctx, v.sshInto(ip, "echo GLUP"))
	return strings.Contains(out, "GLUP")
}

// execVM runs a script inside the VM and returns combined output (truncated for display) + exit code.
func (v *VMRunner) execVM(ctx context.Context, ip, script string) (string, int, error) {
	out, exit, err := v.runHost(ctx, v.sshInto(ip, script))
	if len(out) > maxShellOutput {
		out = out[:maxShellOutput] + "\n… (вывод обрезан)"
	}
	return out, exit, err
}

// Exec runs a user command in the session VM and returns combined output.
func (v *VMRunner) Exec(ctx context.Context, userID int, key string, spec Spec, command string) (string, error) {
	ip, err := v.sessionFor(ctx, userID, key, spec)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	out, _, err := v.execVM(ctx, ip, wrap(command))
	v.touch(vmSID(userID, key))
	return out, err
}

// sessionFor gets the session VM with a deadline that fits booting one.
//
// Exec and Check used to put the whole call, boot included, under
// vmSSHTimeout — 20s against a boot budget of vmBootWait plus the wait for
// sshd. The first thing a student does on a lesson is press Check, so the first
// press cut its own VM's boot in half and, through the classification above,
// reported it as a failed attempt. That is the "sometimes it does not start"
// they described.
func (v *VMRunner) sessionFor(ctx context.Context, userID int, key string, spec Spec) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, vmBootWait+vmSSHTimeout)
	defer cancel()
	return v.EnsureSession(ctx, userID, key, spec)
}

// Check runs the task's check script; passed = exit 0.
func (v *VMRunner) Check(ctx context.Context, userID int, key string, spec Spec, checkScript string) (bool, string, error) {
	ip, err := v.sessionFor(ctx, userID, key, spec)
	if err != nil {
		return false, "", err
	}
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
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
func (v *VMRunner) Preview(ctx context.Context, userID int, key string, spec Spec, port int, path string) ([]byte, string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, spec)
	if err != nil {
		return nil, "", 0, err
	}
	body, ct, status, err := v.previewVM(ctx, ip, port, path)
	if err == nil {
		v.touch(vmSID(userID, key))
	}
	return body, ct, status, err
}

// maxPreviewBody caps a proxied preview response.
const maxPreviewBody = 4 << 20

func (v *VMRunner) previewVM(ctx context.Context, ip string, port int, path string) ([]byte, string, int, error) {
	if port <= 0 {
		port = 80
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	script := fmt.Sprintf(`
url=$(printf %%s '%s' | base64 -d)
if ! command -v curl >/dev/null 2>&1; then echo "GLPREVERR curl-missing"; exit 0; fi
if ! curl -s -m 8 --max-filesize %d -D /tmp/.glph -o /tmp/.glpb "$url" 2>/dev/null; then echo "GLPREVERR no-server"; exit 0; fi
code=$(head -1 /tmp/.glph 2>/dev/null | tr -d '\r' | awk '{print $2}')
ct=$(grep -i '^content-type:' /tmp/.glph 2>/dev/null | head -1 | tr -d '\r' | cut -d' ' -f2-)
echo "GLPREVIEW ${code:-200} ${ct:-text/html}"
base64 /tmp/.glpb 2>/dev/null
`, base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("http://127.0.0.1:%d%s", port, path))), maxPreviewBody)
	out, _, err := v.runHost(ctx, v.sshInto(ip, script))
	if err != nil {
		return nil, "", 0, err
	}
	i := strings.Index(out, "GLPREV")
	if i < 0 {
		return nil, "", 0, fmt.Errorf("preview: пустой ответ песочницы")
	}
	out = out[i:]
	nl := strings.IndexByte(out, '\n')
	if nl < 0 {
		nl = len(out)
	}
	head := strings.TrimSpace(out[:nl])
	if strings.HasPrefix(head, "GLPREVERR") {
		return nil, "", 0, fmt.Errorf("preview: %s", strings.TrimSpace(strings.TrimPrefix(head, "GLPREVERR")))
	}
	fields := strings.SplitN(strings.TrimPrefix(head, "GLPREVIEW "), " ", 2)
	status, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
	ct := "text/html"
	if len(fields) > 1 && strings.TrimSpace(fields[1]) != "" {
		ct = strings.TrimSpace(fields[1])
	}
	body, derr := base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.TrimSpace(out[nl:]), "\n", ""))
	if derr != nil {
		return nil, "", 0, fmt.Errorf("preview: не удалось раскодировать тело: %w", derr)
	}
	return body, ct, status, nil
}

// ── In-VM file editor backend (Monaco), mirrors ShellRunner ──

// Sandbox file errors.
var (
	ErrFileNotFound = errors.New("file not found")
	ErrFileTooLarge = errors.New("file too large")
)

// MaxFileSize is the largest file the editor reads or writes.
const MaxFileSize = 2 << 20

func (v *VMRunner) FSList(ctx context.Context, userID int, key string, spec Spec, dir string) ([]FSEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, spec)
	if err != nil {
		return nil, err
	}
	entries, err := v.listDirVM(ctx, ip, dir)
	if err == nil {
		v.touch(vmSID(userID, key))
	}
	return entries, err
}

func (v *VMRunner) FSRead(ctx context.Context, userID int, key string, spec Spec, file string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, spec)
	if err != nil {
		return nil, err
	}
	data, err := v.readFileVM(ctx, ip, file)
	if err == nil {
		v.touch(vmSID(userID, key))
	}
	return data, err
}

func (v *VMRunner) FSWrite(ctx context.Context, userID int, key string, spec Spec, file string, content []byte) error {
	ctx, cancel := context.WithTimeout(ctx, vmSSHTimeout)
	defer cancel()
	ip, err := v.EnsureSession(ctx, userID, key, spec)
	if err != nil {
		return err
	}
	err = v.writeFileVM(ctx, ip, file, content)
	if err == nil {
		v.touch(vmSID(userID, key))
	}
	return err
}

func (v *VMRunner) listDirVM(ctx context.Context, ip, dir string) ([]FSEntry, error) {
	script := fmt.Sprintf(`d=$(printf %%s '%s' | base64 -d)
echo GLLIST
find "$d" -maxdepth 1 -mindepth 1 -printf '%%y\t%%f\n' 2>/dev/null | LC_ALL=C sort`,
		base64.StdEncoding.EncodeToString([]byte(dir)))
	out, _, err := v.execVM(ctx, ip, script)
	if err != nil {
		return nil, err
	}
	i := strings.Index(out, "GLLIST\n")
	if i < 0 {
		return nil, fmt.Errorf("list: %s", strings.TrimSpace(out))
	}
	entries := []FSEntry{}
	for _, line := range strings.Split(strings.TrimSpace(out[i+len("GLLIST\n"):]), "\n") {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		entries = append(entries, FSEntry{Name: parts[1], Dir: parts[0] == "d"})
	}
	return entries, nil
}

func (v *VMRunner) readFileVM(ctx context.Context, ip, file string) ([]byte, error) {
	script := fmt.Sprintf(`f=$(printf %%s '%s' | base64 -d)
[ -f "$f" ] || { echo GLNOFILE; exit 0; }
[ "$(stat -c %%s "$f")" -gt %d ] && { echo GLTOOBIG; exit 0; }
echo GLFILE
base64 "$f"`, base64.StdEncoding.EncodeToString([]byte(file)), MaxFileSize)
	out, _, err := v.runHost(ctx, v.sshInto(ip, script))
	if err != nil {
		return nil, err
	}
	switch {
	case strings.Contains(out, "GLNOFILE"):
		return nil, ErrFileNotFound
	case strings.Contains(out, "GLTOOBIG"):
		return nil, ErrFileTooLarge
	}
	i := strings.Index(out, "GLFILE\n")
	if i < 0 {
		return nil, fmt.Errorf("read: %s", strings.TrimSpace(out))
	}
	return base64.StdEncoding.DecodeString(strings.ReplaceAll(strings.TrimSpace(out[i+len("GLFILE\n"):]), "\n", ""))
}

// writeFileVM streams content through stdin, so its size is not bound by command-line limits.
func (v *VMRunner) writeFileVM(ctx context.Context, ip, file string, content []byte) error {
	if len(content) > MaxFileSize {
		return ErrFileTooLarge
	}
	inner := fmt.Sprintf(`f=$(printf %%s '%s' | base64 -d)
mkdir -p "$(dirname "$f")" && base64 -d > "$f" && echo GLOK`, base64.StdEncoding.EncodeToString([]byte(file)))
	payload := base64.StdEncoding.EncodeToString(content)
	out, _, err := v.runHostStdin(ctx, v.sshIntoStdin(ip, inner), strings.NewReader(payload))
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
	// The slot is freed by killVM once the host has actually removed the tap.
	// Freeing it here let the next attempt take the same slot, create gltap<slot>,
	// and then lose it to the previous attempt's delayed delete — "Cannot find
	// device gltap0".
	v.killVM(s)
	delete(v.sessions, sid)
}

// killVM kills the Firecracker process, removes the tap and workdir (best effort,
// in the background). It does no bookkeeping — the caller frees the slot / map.
func (v *VMRunner) killVM(s *vmSession) {
	tap := s.tap
	if tap == "" {
		tap = vmTap(s.slot)
	}
	// Each boot attempt gets its own directory (see bootGen), so there is nothing
	// to guess. Every session is created with work already set, so this branch is
	// defensive only — but it stays: deriving a path from the sid instead would
	// delete a directory belonging to a *later* attempt.
	rmWork := ""
	if s.work != "" {
		rmWork = "rm -rf " + s.work
	}
	script := fmt.Sprintf(`
[ -n "%[1]s" ] && [ -f %[1]s/fc.pid ] && kill "$(cat %[1]s/fc.pid)" 2>/dev/null || true
sudo /usr/local/sbin/gl-tap del %[2]s 2>/dev/null || true
%[3]s
`, s.work, tap, rmWork)
	slot := s.slot
	// The slot is freed only after the host has actually removed the tap, and
	// always through the lock — killVM is called both with v.mu held
	// (teardownLocked) and without it (bootWarm), so the goroutine is what keeps
	// that discipline the same in both cases.
	host := func(ctx context.Context) {
		if v.killHook != nil {
			v.killHook(s)
			return
		}
		_, _, _ = v.runHost(ctx, script)
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		host(ctx)
		if slot >= 0 {
			v.mu.Lock()
			if slot < len(v.freeSlot) {
				v.freeSlot[slot] = false
			}
			v.mu.Unlock()
		}
	}()
}

// ── Warm pool ──────────────────────────────────────────────────────────────
// Pre-booted VMs (one per profile) so a student attaches to a ready sandbox
// instantly instead of waiting ~30-60s for a cold cp+boot. Gated by FC_WARM_*.

func (v *VMRunner) signalRefill() {
	if v.refill == nil {
		return
	}
	select {
	case v.refill <- struct{}{}:
	default:
	}
}

// takeWarmLocked pops a ready warm VM of the profile whose size is the one
// asked for (caller holds v.mu). A VM of the wrong size is left in the pool:
// the next lesson that wants that size gets it instantly, and this one boots
// its own rather than quietly running smaller or larger than its author said.
func (v *VMRunner) takeWarmLocked(profile string, cpus, mem int) *vmSession {
	list := v.pool[profile]
	for i := len(list) - 1; i >= 0; i-- {
		s := list[i]
		if s.bootCPUs != cpus || s.bootMemMiB != mem {
			continue
		}
		v.pool[profile] = append(list[:i], list[i+1:]...)
		return s
	}
	return nil
}

// sizeFor reports the machine a lesson would get on this profile: the profile's
// own size unless the lesson asked for something, clamped the same way bootVM
// clamps it. Shared so the pool and the boot cannot disagree about what a
// lesson needs.
func (v *VMRunner) sizeFor(profile string, spec Spec) (cpus, mem int) {
	cpus, mem = v.vcpus, v.memMiB
	switch profile {
	case "lite":
		mem = v.memLite
	case "k8s":
		cpus, mem = 1, v.memK8s
	}
	if spec.MemMiB > 0 {
		mem = clampInt(spec.MemMiB, vmMinMemMiB, v.maxMemMiB)
	}
	if spec.CPUs > 0 {
		cpus = clampInt(spec.CPUs, 1, v.maxVCPUs)
	}
	return cpus, mem
}

// applySetup runs the lesson setup on an already-booted VM (used when a warm VM
// is handed out — the cold path applies the setup inside bootVM instead).
func (v *VMRunner) applySetup(ctx context.Context, s *vmSession, setup string) error {
	if setup == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	b64 := base64.StdEncoding.EncodeToString([]byte(setup))
	script := fmt.Sprintf(
		`out=$(ssh -n -i %[1]s/%[2]s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null `+
			`-o ConnectTimeout=5 -o LogLevel=ERROR root@%[3]s 'echo %[4]s | base64 -d | bash' 2>&1) `+
			`&& echo GLVMOK || { echo "GLVMERR setup"; echo "$out" | tail -5; }`,
		v.dir, v.vmkey, s.ip, b64)
	out, _, err := v.runHost(ctx, script)
	if err != nil {
		vmStartFailed("setup-host")
		slog.Error("lab setup: host error", "sid", s.sid, "err", err)
		return errStudentSetup
	}
	if !strings.Contains(out, "GLVMOK") {
		// The script already collected the last lines of the failure; throwing them
		// away here is why "setup failed" used to be the whole story.
		vmStartFailed("setup")
		slog.Error("lab setup failed", "sid", s.sid, "image", s.image, "detail", tail(out))
		return errStudentSetup
	}
	return nil
}

// errStudentSetup is what the student sees. The reason goes to the log instead:
// the raw text is host paths and shell errors, which mean nothing to them and
// disclose the layout of the FC host.
var errStudentSetup = errors.New("не удалось подготовить окружение урока — открой лабораторную заново")

// tail keeps a failure line short enough to log without dumping a whole script.
func tail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 400 {
		s = s[len(s)-400:]
	}
	return s
}

// bootWarm boots one warm VM of the profile and adds it to the pool (synchronous).
func (v *VMRunner) bootWarm(profile string) {
	v.mu.Lock()
	slot := v.allocSlot()
	v.mu.Unlock()
	if slot < 0 {
		return
	}
	img := "golearn/sandbox:latest"
	switch profile {
	case "k8s":
		img = "golearn/sandbox-k8s:latest"
	case "docker":
		img = "golearn/sandbox-docker:latest"
	}
	sid := fmt.Sprintf("warm-%s-%d", profile, slot)
	sess := &vmSession{
		sid: sid, image: img, profile: profile,
		slot: slot, ip: vmIP(slot),
		tap:     vmTap(slot),
		work:    fmt.Sprintf("%s/sessions/%s-%d", v.dir, sid, bootGen.Add(1)),
		started: time.Now(), last: time.Now(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := v.bootVM(ctx, sess, ""); err != nil {
		// killVM frees the slot itself, once the host has actually removed the tap.
		// Freeing it here as well — which this used to do — handed the slot to the
		// next attempt while the previous one's delayed `gl-tap del gltap<slot>`
		// was still queued: the new VM created its tap and then watched it
		// disappear ("Cannot find device gltap0"). teardownLocked carries the same
		// warning; the warm path was the one place that ignored it, and topUp
		// retries without a backoff, so it reproduced on every failed warm boot.
		v.killVM(sess)
		return
	}
	v.mu.Lock()
	v.pool[profile] = append(v.pool[profile], sess)
	v.mu.Unlock()
}

// poolManager keeps each profile's warm pool topped up (one boot at a time so we
// never fire several 5 GB copies at once).
func (v *VMRunner) poolManager() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		v.topUp()
		select {
		case <-t.C:
		case <-v.refill:
		}
	}
}

func (v *VMRunner) topUp() {
	// All three profiles, lite included. It used to warm only docker and k8s, so
	// FC_WARM_LITE was read, started the pool manager, and then did nothing —
	// every Linux, Git and trainer lesson booted cold, which is the majority of
	// them. A profile left at 0 is skipped by the warmWant check below anyway, so
	// listing it here costs nothing when it is off.
	for _, profile := range []string{"lite", "docker", "k8s"} {
		for {
			v.mu.Lock()
			have := len(v.pool[profile])
			total := len(v.sessions)
			for _, p := range v.pool {
				total += len(p)
			}
			freeSlots := 0
			for _, used := range v.freeSlot {
				if !used {
					freeSlots++
				}
			}
			// keep 1 slot in reserve for a real user even while warming
			need := have < v.warmWant[profile] && total < v.maxVMs && freeSlots > 1
			v.mu.Unlock()
			if !need {
				break
			}
			v.bootWarm(profile)
		}
	}
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
	// Nested: FC host -> VM, with a PTY, running an interactive login shell (zsh,
	// styled in the rootfs's /root/.zshrc).
	inner := fmt.Sprintf(
		"ssh -tt -i %s/%s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+
			"-o ConnectTimeout=5 -o LogLevel=ERROR root@%s "+
			"'cd /root; exec env TERM=xterm-256color HOME=/root zsh -l'",
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

// Stats reports what the pool is doing right now. Exposed so the server can turn
// it into metrics without this package knowing about Prometheus: during an
// incident the first question is always how many VMs exist and how many slots
// are left, and until now that could only be answered by counting processes on
// the host by hand.
type Stats struct {
	Enabled   bool
	Sessions  int            // VMs handed out to a student
	Warm      map[string]int // profile -> pre-booted, idle, not handed out
	FreeSlot  int            // slots still available
	SlotsUsed int            // slots taken; a slot is claimed before the VM boots
	MaxVMs    int
}

// A slot is claimed before the VM boots and released only when it is killed, so
// SlotsUsed is always >= Sessions + sum(Warm). A small gap means something is
// booting right now; a gap that stays is a leaked slot, and the pool will run
// out without a single VM to show for it.

func (v *VMRunner) Stats() Stats {
	if v == nil {
		return Stats{}
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	s := Stats{Enabled: v.enabled, Sessions: len(v.sessions), MaxVMs: v.maxVMs, Warm: map[string]int{}}
	for profile, list := range v.pool {
		s.Warm[profile] = len(list)
	}
	for _, used := range v.freeSlot {
		if used {
			s.SlotsUsed++
		} else {
			s.FreeSlot++
		}
	}
	return s
}
