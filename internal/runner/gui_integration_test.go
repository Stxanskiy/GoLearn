package runner

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// End-to-end check of a windowed lab: a real container, a real PyQt6 program,
// its own VNC server, a real RFB conversation. Skipped unless the Qt sandbox
// image is present, because that is the only thing it cannot do without.
//
// This is the test that justifies the whole design: if Qt's VNC plugin did not
// work, showing a student their window would need Xvfb and x11vnc in every
// sandbox.
func qtSandboxOrSkip(t *testing.T) {
	t.Helper()
	if exec.Command("docker", "info").Run() != nil {
		t.Skip("docker is not available")
	}
	out, err := exec.Command("docker", "images", "-q", qtImage).Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		t.Skipf("image %s is not built (deploy/sandbox-qt/Dockerfile)", qtImage)
	}
}

const qtImage = "golearn/sandbox-qt:latest"

func localShell(t *testing.T) *ShellRunner {
	t.Helper()
	t.Setenv("SANDBOX_LOCAL", "1")
	t.Setenv("SANDBOX_IMAGE", qtImage)
	s := NewShellRunner()
	if !s.Enabled() {
		t.Skip("shell runner is not enabled")
	}
	return s
}

const guiProgram = `from PyQt6.QtWidgets import QApplication, QWidget, QVBoxLayout, QLabel, QPushButton

app = QApplication([])
w = QWidget()
w.setWindowTitle("Справочник студентов")
w.resize(420, 240)

lay = QVBoxLayout(w)
lay.addWidget(QLabel("Это окно открылось в браузере"))
lay.addWidget(QPushButton("Добавить запись"))

w.show()
app.exec()
`

func TestStartGUIOpensWindow(t *testing.T) {
	qtSandboxOrSkip(t)
	s := localShell(t)

	target := Target{UserID: 4242, Key: "gui-test", Image: qtImage}
	t.Cleanup(func() {
		_ = StopGUI(context.Background(), s, target)
		_ = s.Reset(context.Background(), target.UserID, target.Key)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	res, err := StartGUI(ctx, s, "python-gui", target, guiProgram)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Running {
		t.Fatalf("window did not open: errors = %q, output = %q", res.Errors, res.Output)
	}
	if res.Port != 5900 {
		t.Errorf("port = %d, want 5900", res.Port)
	}
	t.Logf("window is up on port %d; stderr = %q", res.Port, res.Errors)
}

// A program that cannot run must report its traceback, not an empty window.
// This is the case a beginner hits constantly, and "nothing happened" would be
// the least useful thing to show them.
func TestStartGUIReportsBrokenProgram(t *testing.T) {
	qtSandboxOrSkip(t)
	s := localShell(t)

	target := Target{UserID: 4243, Key: "gui-broken", Image: qtImage}
	t.Cleanup(func() { _ = s.Reset(context.Background(), target.UserID, target.Key) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	res, err := StartGUI(ctx, s, "python-gui", target, "from PyQt6.QtWidgets import QApplication\nQApplication([]\n")
	if err != nil {
		t.Fatal(err)
	}
	if res.Running {
		t.Error("a program with a syntax error must not count as running")
	}
	if !strings.Contains(res.Errors, "SyntaxError") {
		t.Errorf("the student must see the real message, got %q", res.Errors)
	}
}

// Forgetting app.exec() is the other classic: the program exits immediately and
// the window flashes out of existence. The message has to name the cause.
func TestStartGUIExplainsMissingExec(t *testing.T) {
	qtSandboxOrSkip(t)
	s := localShell(t)

	target := Target{UserID: 4244, Key: "gui-noexec", Image: qtImage}
	t.Cleanup(func() { _ = s.Reset(context.Background(), target.UserID, target.Key) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	code := "from PyQt6.QtWidgets import QApplication, QWidget\napp = QApplication([])\nw = QWidget()\nw.show()\n"
	res, err := StartGUI(ctx, s, "python-gui", target, code)
	if err != nil {
		t.Fatal(err)
	}
	if res.Running {
		t.Error("without app.exec() the program exits, so nothing should be running")
	}
	if !strings.Contains(res.Errors, "app.exec()") {
		t.Errorf("the message should name the cause, got %q", res.Errors)
	}
}

// Speaks RFB to the student's program through the sandbox and saves the window
// as a PNG, so the rendering can be looked at rather than taken on trust.
func TestGUIFramebufferIsReal(t *testing.T) {
	qtSandboxOrSkip(t)
	s := localShell(t)

	target := Target{UserID: 4245, Key: "gui-fb", Image: qtImage}
	t.Cleanup(func() {
		_ = StopGUI(context.Background(), s, target)
		_ = s.Reset(context.Background(), target.UserID, target.Key)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if res, err := StartGUI(ctx, s, "python-gui", target, guiProgram); err != nil {
		t.Fatal(err)
	} else if !res.Running {
		t.Fatalf("window did not open: %q", res.Errors)
	}

	// The RFB conversation runs inside the sandbox — the port is not published —
	// and the resulting PNG comes back base64-encoded.
	grab := `cd /root && cat > grab.py <<'GLEOF'` + "\n" + rfbGrabber + "\nGLEOF\npython3 grab.py && printf 'P:' && base64 window.png"

	out, err := s.Exec(ctx, target.UserID, target.Key, target.Image, target.Setup, grab)
	if err != nil {
		t.Fatal(err)
	}
	i := strings.Index(out, "P:")
	if i < 0 {
		t.Fatalf("no image came back:\n%s", out)
	}
	png, err := base64.StdEncoding.DecodeString(
		strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(out[i+2:]))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	if len(png) < 1000 || string(png[1:4]) != "PNG" {
		t.Fatalf("not a PNG: %d bytes, head %q", len(png), png[:min(16, len(png))])
	}

	path := os.Getenv("GUI_SHOT")
	if path == "" {
		path = "/tmp/gui-window.png"
	}
	if err := os.WriteFile(path, png, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("window captured through RFB: %d bytes -> %s", len(png), path)
}

// rfbGrabber talks RFB 3.3 — the version Qt's VNC plugin announces — asks for
// the whole framebuffer in Raw encoding and writes it out as a PNG.
const rfbGrabber = `
import socket, struct, zlib, time

s = socket.create_connection(("127.0.0.1", 5900), timeout=15)
s.recv(12)
s.sendall(b"RFB 003.003\n")          # the server dictates the version in 3.3
struct.unpack(">I", s.recv(4))[0]    # security type: 1 = none
s.sendall(bytes([1]))                # shared

init = s.recv(24)
W, H = struct.unpack(">HH", init[:4])
nlen = struct.unpack(">I", init[20:24])[0]
s.recv(nlen)

# Pin the pixel format instead of guessing the server's native one: without
# this the rows are parsed at the wrong stride and the image comes out striped.
# 32 bpp, depth 24, little-endian, true colour, R<<16 G<<8 B.
s.sendall(struct.pack(">BBBB BBBB HHH BBB BBB", 0, 0, 0, 0,
                      32, 24, 0, 1, 255, 255, 255, 16, 8, 0, 0, 0, 0))
s.sendall(struct.pack(">BBHi", 2, 0, 1, 0))             # SetEncodings: Raw
s.sendall(struct.pack(">BBHHHH", 3, 0, 0, 0, W, H))     # FramebufferUpdateRequest

def recvall(sock, n):
    buf = b""
    while len(buf) < n:
        b = sock.recv(min(1 << 16, n - len(buf)))
        if not b:
            raise RuntimeError("connection closed with %d of %d bytes" % (len(buf), n))
        buf += b
    return buf

hdr = recvall(s, 4)
assert hdr[0] == 0, "expected FramebufferUpdate, got %d" % hdr[0]
nrect = struct.unpack(">H", hdr[2:4])[0]

pix = bytearray(W * H * 4)
for _ in range(nrect):
    x, y, w, h, enc = struct.unpack(">HHHHi", recvall(s, 12))
    assert enc == 0, "only Raw is handled here, got %d" % enc
    data = recvall(s, w * h * 4)
    for row in range(h):
        src = row * w * 4
        dst = ((y + row) * W + x) * 4
        pix[dst:dst + w * 4] = data[src:src + w * 4]

# BGRX from Qt -> RGB rows for the PNG
rows = bytearray()
for row in range(H):
    rows += b"\x00"
    o = row * W * 4
    for col in range(W):
        p = o + col * 4
        rows += bytes((pix[p + 2], pix[p + 1], pix[p + 0]))

def chunk(tag, body):
    return (struct.pack(">I", len(body)) + tag + body
            + struct.pack(">I", zlib.crc32(tag + body) & 0xFFFFFFFF))

png = (b"\x89PNG\r\n\x1a\n"
       + chunk(b"IHDR", struct.pack(">IIBBBBB", W, H, 8, 2, 0, 0, 0))
       + chunk(b"IDAT", zlib.compress(bytes(rows), 6))
       + chunk(b"IEND", b""))
open("window.png", "wb").write(png)
print("ok %dx%d" % (W, H))
`

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
