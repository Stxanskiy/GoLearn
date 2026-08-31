package runner

import "context"

// Engine is the lab-sandbox backend the HTTP handlers talk to. Two implementations
// exist:
//
//   - ShellRunner — one ephemeral Docker container per (user, lesson), reached over
//     SSH to a shared sandbox VM. Used by every course today.
//   - VMRunner — one Firecracker micro-VM per (user, lesson), reached over SSH. Gives
//     the Docker/Kubernetes courses a real, isolated kernel (native dockerd/k3s, no
//     dind, no host root) — the model devops404 uses.
//
// The two are interchangeable behind this interface; a dispatcher (see Engines)
// picks one per lesson from the sandbox image name.
type Engine interface {
	Enabled() bool

	// EnsureSession creates (or returns) the session sandbox and gives back an
	// opaque handle to pass to OpenPTY — a container name for ShellRunner, a VM
	// address for VMRunner.
	EnsureSession(ctx context.Context, userID int, key, image, setup string) (string, error)
	OpenPTY(handle string, cols, rows int) (*PTYSession, error)

	Exec(ctx context.Context, userID int, key, image, setup, command string) (string, error)
	Check(ctx context.Context, userID int, key, image, setup, checkScript string) (bool, string, error)
	Preview(ctx context.Context, userID int, key, image, setup string, port int, path string) ([]byte, string, int, error)

	FSList(ctx context.Context, userID int, key, image, setup, dir string) ([]FSEntry, error)
	FSRead(ctx context.Context, userID int, key, image, setup, file string) ([]byte, error)
	FSWrite(ctx context.Context, userID int, key, image, setup, file string, content []byte) error

	Reset(ctx context.Context, userID int, key string) error
}

// Both runners satisfy the interface.
var (
	_ Engine = (*ShellRunner)(nil)
	_ Engine = (*VMRunner)(nil)
)
