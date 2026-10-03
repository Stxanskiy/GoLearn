package runner

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Reaching a TCP port inside a session sandbox.
//
// A windowed program serves VNC on a port of its own (see gui.go) and the
// browser has to talk to it. Sandboxes publish no ports — labs run with
// --network none, deliberately — so the bytes travel the way everything else
// does: through a command in the session, one pipe each way.
//
// No PTY anywhere in here. OpenPTY asks for one because a terminal wants line
// discipline; a VNC stream is binary, and a PTY would translate CR and LF in
// the middle of pixel data.
//
// The relay is a Python one-liner rather than netcat for the same reason the
// port check in gui.go is: nc may or may not be in an image, python is always
// in one that can run a Python lesson.

// PortConn is a byte pipe to a port inside the sandbox. Deliberately nothing
// more than an io.ReadWriteCloser: the caller copies bytes and knows nothing
// about how they get there.
type PortConn struct {
	r      io.Reader
	w      io.WriteCloser
	closer func()
}

func (p *PortConn) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *PortConn) Write(b []byte) (int, error) { return p.w.Write(b) }

// NewPortConn builds a PortConn from any pair of pipes. Exists so tests and
// fakes can stand in for a sandbox without a container.
func NewPortConn(r io.Reader, w io.WriteCloser, closer func()) *PortConn {
	return &PortConn{r: r, w: w, closer: closer}
}

func (p *PortConn) Close() error {
	if p.closer != nil {
		p.closer()
	}
	return nil
}

// portRelay copies stdin into the port and the port back to stdout. read1 and
// an explicit flush, not buffered reads: a VNC stream has no framing, and
// waiting for a full buffer would hold the first frame until something else
// happened to arrive.
const portRelay = `import socket,sys,threading
s=socket.create_connection(('127.0.0.1',%d),10)
def up():
    try:
        while True:
            d=sys.stdin.buffer.read1(65536)
            if not d: break
            s.sendall(d)
    except Exception: pass
    try: s.shutdown(socket.SHUT_WR)
    except Exception: pass
threading.Thread(target=up,daemon=True).start()
try:
    while True:
        d=s.recv(65536)
        if not d: break
        sys.stdout.buffer.write(d); sys.stdout.buffer.flush()
except Exception: pass
`

func relayScript(port int) string { return fmt.Sprintf(portRelay, port) }

// shellQuote wraps a script so a remote shell passes it through unchanged.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// DialPort opens a byte pipe to a TCP port inside a running container session.
func (s *ShellRunner) DialPort(container string, port int) (*PortConn, error) {
	if !s.enabled {
		return nil, fmt.Errorf("sandbox disabled")
	}
	inner := fmt.Sprintf("docker exec -i %s python3 -c %s", container, shellQuote(relayScript(port)))

	var cmd *exec.Cmd
	if s.local {
		cmd = exec.Command("bash", "-c", inner)
	} else {
		cmd = exec.Command("ssh",
			"-i", s.keyFile,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ConnectTimeout=8",
			"-o", "LogLevel=ERROR",
			"-p", s.port,
			s.user+"@"+s.host,
			inner)
	}
	return startRelay(cmd)
}

// DialPort opens a byte pipe to a TCP port inside a running micro-VM. The hop
// is nested the same way every other VM call is: into the Firecracker host over
// SSH, then from there into the VM.
func (v *VMRunner) DialPort(handle string, port int) (*PortConn, error) {
	if !v.enabled {
		return nil, fmt.Errorf("VM-песочница отключена")
	}
	sg, err := v.signer()
	if err != nil {
		return nil, err
	}
	client, err := ssh.Dial("tcp", v.host+":"+v.port, &ssh.ClientConfig{
		User:            v.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(sg)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
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
	inner := fmt.Sprintf(
		"ssh -i %s/%s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "+
			"-o ConnectTimeout=5 -o LogLevel=ERROR root@%s python3 -c %s",
		v.dir, v.vmkey, handle, shellQuote(relayScript(port)))
	if err := session.Start(inner); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, err
	}
	return &PortConn{
		r: stdout,
		w: stdin,
		closer: func() {
			_ = session.Close()
			_ = client.Close()
		},
	}, nil
}

func startRelay(cmd *exec.Cmd) (*PortConn, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &PortConn{
		r: stdout,
		w: stdin,
		closer: func() {
			_ = stdin.Close()
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			_ = cmd.Wait()
		},
	}, nil
}
