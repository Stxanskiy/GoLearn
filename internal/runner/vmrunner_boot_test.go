package runner

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// A micro-VM takes 15-25 seconds to come up, and for all of that time alive()
// says no, because sshd is not listening yet. EnsureSession used to read that as
// "dead, rebuild", tear the session down mid-boot and delete the work dir out
// from under the boot script that was still running — so BOTH the request that
// started the boot and the one that interrupted it failed with "No such file or
// directory". Opening the terminal and pressing "Проверить" was enough, and it
// is exactly what production logged.
func TestConcurrentEnsureSessionBootsOnce(t *testing.T) {
	var calls atomic.Int32
	release := make(chan struct{})
	v := &VMRunner{
		enabled:  true,
		sessions: map[string]*vmSession{},
		booting:  map[string]*vmBoot{},
		bringUpHook: func(ctx context.Context, sid string, userID int, key string, spec Spec) (string, error) {
			calls.Add(1)
			<-release // hold the boot open so every caller arrives while it runs
			return "172.31.7.2", nil
		},
	}

	const callers = 8
	ips := make([]string, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ips[i], errs[i] = v.EnsureSession(context.Background(), 1, "l644", Spec{Image: "golearn/sandbox"})
		}(i)
	}
	// Let them pile up on the in-flight boot before it finishes.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("%d concurrent requests started %d boots, want 1", callers, got)
	}
	for i := range ips {
		if errs[i] != nil || ips[i] != "172.31.7.2" {
			t.Errorf("caller %d got ip=%q err=%v", i, ips[i], errs[i])
		}
	}
	if len(v.booting) != 0 {
		t.Errorf("finished boot left %d entries behind", len(v.booting))
	}
}

// A boot that fails must fail for everyone waiting on it, not just the caller
// that happened to start it, and must not leave the sid claimed forever.
func TestWaitersSeeTheBootError(t *testing.T) {
	boom := errors.New("песочница не запустилась")
	release := make(chan struct{})
	v := &VMRunner{
		enabled:  true,
		sessions: map[string]*vmSession{},
		booting:  map[string]*vmBoot{},
		bringUpHook: func(ctx context.Context, sid string, userID int, key string, spec Spec) (string, error) {
			<-release
			return "", boom
		},
	}

	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = v.EnsureSession(context.Background(), 1, "l644", Spec{Image: "golearn/sandbox"})
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	for i, err := range errs {
		if !errors.Is(err, boom) {
			t.Errorf("caller %d got %v, want the boot error", i, err)
		}
	}
	if len(v.booting) != 0 {
		t.Errorf("failed boot left the session claimed: %v", v.booting)
	}
}

// A caller that gives up must not take the boot down with it — the others are
// still waiting on the same VM.
func TestGivingUpDoesNotCancelTheBootForOthers(t *testing.T) {
	release := make(chan struct{})
	v := &VMRunner{
		enabled:  true,
		sessions: map[string]*vmSession{},
		booting:  map[string]*vmBoot{},
		bringUpHook: func(ctx context.Context, sid string, userID int, key string, spec Spec) (string, error) {
			<-release
			return "172.31.7.2", nil
		},
	}

	started := make(chan struct{})
	var stayed error
	var ip string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		close(started)
		ip, stayed = v.EnsureSession(context.Background(), 1, "l644", Spec{Image: "golearn/sandbox"})
	}()
	<-started
	time.Sleep(50 * time.Millisecond)

	gone, cancel := context.WithCancel(context.Background())
	quit := make(chan error, 1)
	go func() {
		_, err := v.EnsureSession(gone, 1, "l644", Spec{Image: "golearn/sandbox"})
		quit <- err
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()

	if err := <-quit; !errors.Is(err, context.Canceled) {
		t.Errorf("the caller that walked away got %v, want context.Canceled", err)
	}
	close(release)
	wg.Wait()
	if stayed != nil || ip != "172.31.7.2" {
		t.Errorf("the caller that waited got ip=%q err=%v", ip, stayed)
	}
}

// Opening a second lesson drops the VM the student already holds. When that VM
// is still booting, cancelling its context was not enough: teardown ran straight
// away, deleting the work dir and the tap out from under a boot script that went
// on using them — and whatever the script created after that was left on the
// host with nothing tracking it, while the slot had already been handed on.
//
// The three tests above all replace bringUp with a hook, so none of them reaches
// this. This one exercises dropOtherSessions directly.
func TestDroppingASessionWaitsForItsBootToStop(t *testing.T) {
	v := &VMRunner{
		enabled:  true,
		sessions: map[string]*vmSession{},
		booting:  map[string]*vmBoot{},
		freeSlot: make([]bool, 4),
		killHook: func(*vmSession) {},
	}

	// The student's first lesson, mid-boot.
	old := &vmSession{sid: "u1-l1", userID: 1, slot: 0, ip: vmIP(0),
		tap: vmTap(0), work: "/opt/fc/sessions/u1-l1-1"}
	bootCtx, cancel := context.WithCancel(context.Background())
	boot := &vmBoot{done: make(chan struct{}), cancel: cancel}
	v.sessions[old.sid] = old
	v.booting[old.sid] = boot
	v.freeSlot[0] = true

	// The boot script: it notices the cancellation and only then returns, which is
	// the ordering teardown has to respect.
	finished := make(chan struct{})
	go func() {
		<-bootCtx.Done()
		time.Sleep(30 * time.Millisecond) // unwinding takes a moment
		close(finished)
		close(boot.done)
	}()

	done := make(chan struct{})
	go func() {
		v.dropOtherSessions(context.Background(), "u1-l2", 1)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("dropOtherSessions did not return — cancel-then-wait deadlocked")
	}

	select {
	case <-finished:
	default:
		t.Error("the session was torn down before its boot had stopped")
	}
	v.mu.Lock()
	_, stillThere := v.sessions[old.sid]
	v.mu.Unlock()
	if stillThere {
		t.Error("the superseded session was not torn down")
	}
}

// Two lessons opened at once by the same user cancel each other. Waiting without
// cancelling first would deadlock; cancelling makes both sides return, so this
// has to finish.
func TestTwoSimultaneousDropsDoNotDeadlock(t *testing.T) {
	v := &VMRunner{
		enabled:  true,
		sessions: map[string]*vmSession{},
		booting:  map[string]*vmBoot{},
		freeSlot: make([]bool, 4),
		killHook: func(*vmSession) {},
	}

	for i, sid := range []string{"u1-l1", "u1-l2"} {
		s := &vmSession{sid: sid, userID: 1, slot: i, ip: vmIP(i),
			tap: vmTap(i), work: "/opt/fc/sessions/" + sid}
		ctx, cancel := context.WithCancel(context.Background())
		b := &vmBoot{done: make(chan struct{}), cancel: cancel}
		v.sessions[sid] = s
		v.booting[sid] = b
		v.freeSlot[i] = true
		go func() {
			<-ctx.Done()
			close(b.done)
		}()
	}

	var wg sync.WaitGroup
	for _, sid := range []string{"u1-l3", "u1-l4"} {
		wg.Add(1)
		go func(sid string) {
			defer wg.Done()
			v.dropOtherSessions(context.Background(), sid, 1)
		}(sid)
	}

	settled := make(chan struct{})
	go func() { wg.Wait(); close(settled) }()
	select {
	case <-settled:
	case <-time.After(5 * time.Second):
		t.Fatal("two concurrent drops deadlocked")
	}

	v.mu.Lock()
	left := len(v.sessions)
	v.mu.Unlock()
	if left != 0 {
		t.Errorf("sessions left behind: %d", left)
	}
}

// A failed bring-up shows the student one line and carries the host's own
// output alongside it. Both halves matter: the student must never see
// "/opt/fc/sessions/…", and an operator must not have to read the pod's log to
// learn whether the VM never booted or a lesson's setup script exited non-zero.
func TestBootErrorSeparatesStudentAndOperator(t *testing.T) {
	err := &BootError{Reason: "setup", Detail: "GLVMERR setup\nE: Unable to locate package postgresql-16"}

	if got := err.Error(); got != errStudentBoot.Error() {
		t.Errorf("student sees %q, want the one safe line", got)
	}
	if strings.Contains(err.Error(), "postgresql") || strings.Contains(err.Error(), "GLVMERR") {
		t.Errorf("host output leaked to the student: %q", err.Error())
	}
	// Anything that compared against the sentinel has to keep working.
	if !errors.Is(err, errStudentBoot) {
		t.Error("errors.Is against the sentinel stopped matching")
	}

	diag := err.Diagnostic()
	if !strings.Contains(diag, "setup") || !strings.Contains(diag, "postgresql-16") {
		t.Errorf("diagnostic lost the reason or the detail: %q", diag)
	}

	// A reason with nothing behind it still says which category it was.
	if got := (&BootError{Reason: "boot-timeout"}).Diagnostic(); got != "boot-timeout" {
		t.Errorf("bare diagnostic = %q", got)
	}
}
