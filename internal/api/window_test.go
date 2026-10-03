package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/auth"
	"github.com/gorilla/websocket"
)

// The window relay carries RFB, which has no framing of its own and no escape
// sequences. Anything this code adds to or drops from the stream corrupts the
// student's picture, so the test is about the bytes and nothing else.
func TestWindowRelayPassesBytesBothWays(t *testing.T) {
	h, c := labFixture(t)
	srv := httptest.NewServer(h)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/lessons/102/lab/window"
	header := http.Header{"Cookie": {auth.SessionCookie + "=" + studentToken}}

	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if got := c.sandbox.dialedPort; got != 5900 {
		t.Errorf("dialed port = %d, want 5900 (Qt's VNC default)", got)
	}

	// Sandbox -> browser. The payload is deliberately nasty: a zero byte, a bare
	// CR and a bare LF are exactly what a PTY or a line-oriented relay would
	// mangle, and RFB is full of all three.
	payload := []byte{0x00, 'R', 'F', 'B', 0x0d, 0x0a, 0xff, 0x80, 0x00, 0x1b}
	go func() {
		_, _ = c.sandbox.sandboxWrites.Write(payload)
	}()

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	kind, got, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if kind != websocket.BinaryMessage {
		t.Errorf("frame type = %d, want binary (%d)", kind, websocket.BinaryMessage)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("sandbox -> browser corrupted:\n got %v\nwant %v", got, payload)
	}

	// Browser -> sandbox: a client's keyboard and pointer events go this way.
	up := []byte{0x04, 0x01, 0x00, 0x00, 0xff, 0x0d, 0x0a, 0x00}
	if err := conn.WriteMessage(websocket.BinaryMessage, up); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, len(up))
	if _, err := readFull(c.sandbox.sandboxReads, buf); err != nil {
		t.Fatalf("read from sandbox side: %v", err)
	}
	if !bytes.Equal(buf, up) {
		t.Errorf("browser -> sandbox corrupted:\n got %v\nwant %v", buf, up)
	}
}

func TestWindowRejectsForeignOrigin(t *testing.T) {
	h, _ := labFixture(t)
	srv := httptest.NewServer(h)
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/lessons/102/lab/window"
	_, resp, err := websocket.DefaultDialer.Dial(url, http.Header{
		"Cookie": {auth.SessionCookie + "=" + studentToken},
		"Origin": {"https://evil.test"},
	})
	if err == nil || resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("a cross-origin window must be refused, got %v", err)
	}
}

func readFull(r interface{ Read([]byte) (int, error) }, buf []byte) (int, error) {
	n := 0
	for n < len(buf) {
		m, err := r.Read(buf[n:])
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
