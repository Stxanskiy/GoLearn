package runner

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestClassifyExitSeparatesTransportFromVerdict(t *testing.T) {
	ctx := context.Background()

	// A check script that says "no" is a verdict, not a failure.
	err := exec.Command("bash", "-c", "exit 1").Run()
	if code, cErr := classifyExit(ctx, nil, err); code != 1 || cErr != nil {
		t.Errorf("exit 1: code=%d err=%v, want 1/nil", code, cErr)
	}

	// ssh's own failure code must not be reported as a wrong answer.
	err = exec.Command("bash", "-c", "exit 255").Run()
	if code, cErr := classifyExit(ctx, []byte("ssh: connect to host: refused"), err); code != 255 || cErr == nil {
		t.Errorf("exit 255: code=%d err=%v, want 255 and an error", code, cErr)
	}
	// But for a local script 255 carries no such meaning.
	if code, cErr := classifyLocalExit(ctx, err); code != 255 || cErr != nil {
		t.Errorf("local exit 255: code=%d err=%v, want 255/nil", code, cErr)
	}

	// An expired deadline is ours, whatever the process did.
	dead, cancel := context.WithTimeout(ctx, time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(dead, "bash", "-c", "sleep 2")
	runErr := cmd.Run()
	if code, cErr := classifyExit(dead, nil, runErr); cErr == nil || code != -1 {
		t.Errorf("deadline: code=%d err=%v, want -1 and an error", code, cErr)
	}
	if _, cErr := classifyLocalExit(dead, runErr); cErr == nil {
		t.Error("deadline, local: want an error")
	}
}
