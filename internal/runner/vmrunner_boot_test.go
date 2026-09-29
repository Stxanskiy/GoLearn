package runner

import (
	"context"
	"errors"
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
		bringUpHook: func(ctx context.Context, sid string, userID int, key, image, setup string) (string, error) {
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
			ips[i], errs[i] = v.EnsureSession(context.Background(), 1, "l644", "golearn/sandbox", "")
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
		bringUpHook: func(ctx context.Context, sid string, userID int, key, image, setup string) (string, error) {
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
			_, errs[i] = v.EnsureSession(context.Background(), 1, "l644", "golearn/sandbox", "")
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
		bringUpHook: func(ctx context.Context, sid string, userID int, key, image, setup string) (string, error) {
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
		ip, stayed = v.EnsureSession(context.Background(), 1, "l644", "golearn/sandbox", "")
	}()
	<-started
	time.Sleep(50 * time.Millisecond)

	gone, cancel := context.WithCancel(context.Background())
	quit := make(chan error, 1)
	go func() {
		_, err := v.EnsureSession(gone, 1, "l644", "golearn/sandbox", "")
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
