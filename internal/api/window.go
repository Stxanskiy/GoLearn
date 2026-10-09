package api

import (
	"net/http"
	"strconv"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/lab"
	"github.com/backendraz/golearn/internal/runner"
	"github.com/gorilla/websocket"
)

// Windowed labs: the student writes a PyQt program, presses Run, and sees the
// window it opens.
//
// The program is its own VNC server (QT_QPA_PLATFORM=vnc — Qt's VNC plugin
// ships inside the PyQt6 wheel), so nothing extra runs in the sandbox: no
// Xvfb, no x11vnc, no desktop. This file is the two ends of that: start the
// program, then relay its VNC bytes to the browser over a WebSocket, which is
// what noVNC expects on the other side.

// startWindow runs a windowed task's program and reports whether a window came
// up. It does not wait for the program to finish — a window program never does.
func (a *API) startWindow(w http.ResponseWriter, r *http.Request, t taskRun) {
	res, err := runner.StartGUI(r.Context(), a.Sandbox, t.kind, t.target, t.code)
	if err != nil {
		a.sandboxError(w, "start window", err)
		return
	}
	writeJSON(w, http.StatusOK, apigen.RunResult{
		Running: &res.Running,
		Errors:  res.Errors,
		Output:  res.Output,
	})
}

// taskRun is what both run paths need: which language, whose sandbox, what code.
type taskRun struct {
	kind   string
	target runner.Target
	code   string
}

// serveWindow bridges a WebSocket to the VNC port of the program running in the
// student's session. Binary frames both ways and nothing else: noVNC speaks RFB
// over the socket, and anything this code added to the stream would corrupt it.
func (a *API) serveWindow(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.sandboxLab(w, r)
	if !ok {
		return
	}
	user := userFrom(r.Context())

	if origin := r.Header.Get("Origin"); origin != "" && !a.originAllowed(r, origin) {
		writeError(w, http.StatusForbidden, codeCSRFRejected, "cross-origin window rejected")
		return
	}
	if !terminalLimiter.Allow(strconv.Itoa(user.ID)) {
		writeRateLimited(w, terminalLimiter)
		return
	}

	handle, err := a.Sandbox.EnsureSession(r.Context(), user.ID, ref.key(), ref.spec)
	if err != nil {
		a.sandboxError(w, "window: ensure session", err)
		return
	}

	// Upgrade only once the sandbox answered: a failure before this point can
	// still be reported as a normal HTTP error the client can read.
	upgrader := websocket.Upgrader{
		ReadBufferSize:  32 << 10,
		WriteBufferSize: 32 << 10,
		Subprotocols:    []string{"binary"}, // what noVNC asks for
		CheckOrigin:     func(*http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	port, err := a.Sandbox.DialPort(handle, lab.WindowPort)
	if err != nil {
		a.log.Error("window: dial port", "error", err)
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "окно не запущено"))
		return
	}
	defer port.Close()

	done := make(chan struct{}, 2)

	// Sandbox -> browser.
	go func() {
		defer func() { done <- struct{}{} }()
		buf := make([]byte, 32<<10)
		for {
			n, err := port.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
				a.Sandbox.Touch(user.ID, ref.key())
			}
			if err != nil {
				return
			}
		}
	}()

	// Browser -> sandbox.
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if _, werr := port.Write(data); werr != nil {
				return
			}
		}
	}()

	<-done
}
