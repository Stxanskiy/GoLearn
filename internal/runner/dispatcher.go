package runner

import (
	"context"
	"strings"
)

// Dispatcher routes each lab operation to the right sandbox engine based on the
// lesson's image, so both backends can run side by side during the migration:
//
//   - lessons whose image is the Firecracker Docker profile (golearn/sandbox-docker)
//     go to the VMRunner when it is enabled (FC_ENABLED);
//   - everything else stays on the container ShellRunner.
//
// It implements Engine itself, so the handlers keep a single field and never learn
// which backend served a given request.
type Dispatcher struct {
	Shell *ShellRunner
	VM    *VMRunner
}

var _ Engine = (*Dispatcher)(nil)

func NewDispatcher(shell *ShellRunner, vm *VMRunner) *Dispatcher {
	return &Dispatcher{Shell: shell, VM: vm}
}

// pick chooses the engine for an image. The Docker course is the pilot: its
// lessons carry golearn/sandbox-docker and route to the micro-VM when it is on.
func (d *Dispatcher) pick(image string) Engine {
	if d.VM != nil && d.VM.Enabled() && strings.Contains(image, "sandbox-docker") {
		return d.VM
	}
	return d.Shell
}

// Enabled reports whether any backend can serve labs at all.
func (d *Dispatcher) Enabled() bool {
	return (d.Shell != nil && d.Shell.Enabled()) || (d.VM != nil && d.VM.Enabled())
}

// EnsureSession creates the sandbox and returns a handle. VM handles are tagged
// with a "vm:" prefix so OpenPTY can route back to the same backend.
func (d *Dispatcher) EnsureSession(ctx context.Context, userID int, key, image, setup string) (string, error) {
	e := d.pick(image)
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
	return d.pick(image).Exec(ctx, userID, key, image, setup, command)
}

func (d *Dispatcher) Check(ctx context.Context, userID int, key, image, setup, checkScript string) (bool, string, error) {
	return d.pick(image).Check(ctx, userID, key, image, setup, checkScript)
}

func (d *Dispatcher) Preview(ctx context.Context, userID int, key, image, setup string, port int, path string) ([]byte, string, int, error) {
	return d.pick(image).Preview(ctx, userID, key, image, setup, port, path)
}

func (d *Dispatcher) FSList(ctx context.Context, userID int, key, image, setup, dir string) ([]FSEntry, error) {
	return d.pick(image).FSList(ctx, userID, key, image, setup, dir)
}

func (d *Dispatcher) FSRead(ctx context.Context, userID int, key, image, setup, file string) ([]byte, error) {
	return d.pick(image).FSRead(ctx, userID, key, image, setup, file)
}

func (d *Dispatcher) FSWrite(ctx context.Context, userID int, key, image, setup, file string, content []byte) error {
	return d.pick(image).FSWrite(ctx, userID, key, image, setup, file, content)
}

// Reset has no image to route on, so it clears the session in both backends; the
// one that does not hold it simply no-ops.
func (d *Dispatcher) Reset(ctx context.Context, userID int, key string) error {
	var err error
	if d.Shell != nil {
		err = d.Shell.Reset(ctx, userID, key)
	}
	if d.VM != nil && d.VM.Enabled() {
		if e := d.VM.Reset(ctx, userID, key); e != nil && err == nil {
			err = e
		}
	}
	return err
}
