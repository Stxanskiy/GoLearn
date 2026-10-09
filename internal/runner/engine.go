package runner

import "context"

// Engine is the lab-sandbox backend the HTTP handlers talk to. Two implementations
// exist:
//
//   - ShellRunner — one ephemeral Docker container per (user, lesson), reached over
//     SSH to a shared sandbox VM. Used by every course today.
//   - VMRunner — one Firecracker micro-VM per (user, lesson), reached over SSH. Gives
//     the Docker/Kubernetes courses a real, isolated kernel (native dockerd/k3s, no
//     dind, no host root) — the model this kind of platform uses.
//
// The two are interchangeable behind this interface; a dispatcher (see Engines)
// picks one per lesson from the sandbox image name.
type Engine interface {
	Enabled() bool

	// Session, Touch and HasSession expose the session's lifetime to the API, which
	// shows the remaining time and reconnects a reloaded page to a live sandbox.
	Session(userID int, key string) (SessionInfo, bool)
	Touch(userID int, key string)
	HasSession(userID int, key string) bool

	// EnsureSession creates (or returns) the session sandbox and gives back an
	// opaque handle to pass to OpenPTY — a container name for ShellRunner, a VM
	// address for VMRunner.
	EnsureSession(ctx context.Context, userID int, key string, spec Spec) (string, error)
	OpenPTY(handle string, cols, rows int) (*PTYSession, error)

	// DialPort opens a byte pipe to a TCP port inside the session: how a
	// windowed program's VNC reaches the browser. No PTY — the stream is binary.
	DialPort(handle string, port int) (*PortConn, error)

	Exec(ctx context.Context, userID int, key string, spec Spec, command string) (string, error)
	Check(ctx context.Context, userID int, key string, spec Spec, checkScript string) (bool, string, error)
	Preview(ctx context.Context, userID int, key string, spec Spec, port int, path string) ([]byte, string, int, error)

	FSList(ctx context.Context, userID int, key string, spec Spec, dir string) ([]FSEntry, error)
	FSRead(ctx context.Context, userID int, key string, spec Spec, file string) ([]byte, error)
	FSWrite(ctx context.Context, userID int, key string, spec Spec, file string, content []byte) error

	Reset(ctx context.Context, userID int, key string) error
}

// Both runners satisfy the interface.
var (
	_ Engine = (*ShellRunner)(nil)
	_ Engine = (*VMRunner)(nil)
)

// Spec is what a lesson asks its sandbox to be.
//
// It replaced a pair of `image, setup string` parameters threaded through every
// method of this interface. That pair could not grow: the VM's size was decided
// by profileOf() from the image name, so three hard-coded sizes were the only
// ones a lesson could have, and an author had no way to say "this one needs
// more memory" short of inventing an image.
//
// Zero fields mean "whatever the profile gives", so a lesson that asks for
// nothing behaves exactly as it did before.
type Spec struct {
	Image  string // sandbox image; picks the golden rootfs and the default size
	Setup  string // script run inside the sandbox once it is up
	CPUs   int    // 0 → the profile's default
	MemMiB int    // 0 → the profile's default
}
