package runner

import (
	"context"
	"strings"
)

// Dispatcher routes lab operations to the live sandbox backend; implements Engine itself.
type Dispatcher struct {
	Shell *ShellRunner
	VM    *VMRunner
}

var _ Engine = (*Dispatcher)(nil)

func NewDispatcher(shell *ShellRunner, vm *VMRunner) *Dispatcher {
	return &Dispatcher{Shell: shell, VM: vm}
}

// pick returns the micro-VM backend when it is enabled, otherwise the container backend.
func (d *Dispatcher) pick() Engine {
	if d.VM == nil || !d.VM.Enabled() {
		return d.Shell
	}
	return d.VM
}

// Enabled reports whether any backend can serve labs at all.
func (d *Dispatcher) Enabled() bool {
	return (d.Shell != nil && d.Shell.Enabled()) || (d.VM != nil && d.VM.Enabled())
}

// engines lists the live backends; a session belongs to exactly one of them.
func (d *Dispatcher) engines() []Engine {
	var out []Engine
	if d.VM != nil && d.VM.Enabled() {
		out = append(out, d.VM)
	}
	if d.Shell != nil && d.Shell.Enabled() {
		out = append(out, d.Shell)
	}
	return out
}

// Session returns the lifetime of whichever backend holds the session.
func (d *Dispatcher) Session(userID int, key string) (SessionInfo, bool) {
	for _, e := range d.engines() {
		if info, ok := e.Session(userID, key); ok {
			return info, true
		}
	}
	return SessionInfo{}, false
}

// Touch marks the session active in whichever backend holds it.
func (d *Dispatcher) Touch(userID int, key string) {
	for _, e := range d.engines() {
		if e.HasSession(userID, key) {
			e.Touch(userID, key)
			return
		}
	}
}

func (d *Dispatcher) HasSession(userID int, key string) bool {
	for _, e := range d.engines() {
		if e.HasSession(userID, key) {
			return true
		}
	}
	return false
}

// EnsureSession creates the sandbox and returns a handle. VM handles are tagged
// with a "vm:" prefix so OpenPTY can route back to the same backend.
func (d *Dispatcher) EnsureSession(ctx context.Context, userID int, key, image, setup string) (string, error) {
	e := d.pick()
	h, err := e.EnsureSession(ctx, userID, key, image, setup)
	if err != nil {
		return "", err
	}
	if _, isVM := e.(*VMRunner); isVM {
		return "vm:" + h, nil
	}
	return h, nil
}

// OpenPTY routes by the handle tag EnsureSession produced.
func (d *Dispatcher) OpenPTY(handle string, cols, rows int) (*PTYSession, error) {
	if strings.HasPrefix(handle, "vm:") {
		return d.VM.OpenPTY(strings.TrimPrefix(handle, "vm:"), cols, rows)
	}
	return d.Shell.OpenPTY(handle, cols, rows)
}

func (d *Dispatcher) Exec(ctx context.Context, userID int, key, image, setup, command string) (string, error) {
	return d.pick().Exec(ctx, userID, key, image, setup, command)
}

func (d *Dispatcher) Check(ctx context.Context, userID int, key, image, setup, checkScript string) (bool, string, error) {
	return d.pick().Check(ctx, userID, key, image, setup, checkScript)
}

func (d *Dispatcher) Preview(ctx context.Context, userID int, key, image, setup string, port int, path string) ([]byte, string, int, error) {
	return d.pick().Preview(ctx, userID, key, image, setup, port, path)
}

func (d *Dispatcher) FSList(ctx context.Context, userID int, key, image, setup, dir string) ([]FSEntry, error) {
	return d.pick().FSList(ctx, userID, key, image, setup, dir)
}

func (d *Dispatcher) FSRead(ctx context.Context, userID int, key, image, setup, file string) ([]byte, error) {
	return d.pick().FSRead(ctx, userID, key, image, setup, file)
}

func (d *Dispatcher) FSWrite(ctx context.Context, userID int, key, image, setup, file string, content []byte) error {
	return d.pick().FSWrite(ctx, userID, key, image, setup, file, content)
}

// Reset clears the session in every live backend; the ones that do not hold it no-op.
func (d *Dispatcher) Reset(ctx context.Context, userID int, key string) error {
	var err error
	for _, e := range d.engines() {
		if re := e.Reset(ctx, userID, key); re != nil && err == nil {
			err = re
		}
	}
	return err
}
