package runner

import (
	"testing"
	"time"
)

func TestRunningSince(t *testing.T) {
	at, running := runningSince("true 2026-09-19T10:11:12.5Z\n")
	if !running || !at.Equal(time.Date(2026, 9, 19, 10, 11, 12, 500000000, time.UTC)) {
		t.Fatalf("at = %v, running = %v", at, running)
	}

	before := time.Now()
	at, running = runningSince("true")
	if !running || at.Before(before) {
		t.Fatalf("missing StartedAt: at = %v, running = %v", at, running)
	}

	for _, out := range []string{"false 2026-09-19T10:11:12Z", "", "Error: No such object: gl-s-u1-l2"} {
		if _, running := runningSince(out); running {
			t.Fatalf("%q must not count as running", out)
		}
	}
}
