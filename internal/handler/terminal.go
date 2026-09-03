package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// Reject cross-site WebSocket connections (the session cookie is sent cross-site).
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // non-browser client (no Origin header)
		}
		u, err := url.Parse(origin)
		return err == nil && u.Host == r.Host
	},
}

// labBanner is shown when a sandbox terminal opens (ANSI cyan, CRLF endings).
const labBanner = "\x1b[1;36m" +
	"\r\n ████████   ██████   ████████" +
	"\r\n    ██     ██    ██     ██   " +
	"\r\n    ██     ██    ██     ██   " +
	"\r\n    ██     ██    ██     ██   " +
	"\r\n    ██      ██████      ██   " +
	"\x1b[0m\r\n\x1b[90m Welcome to your TOT lab environment!\x1b[0m\r\n\r\n"

// TermWS upgrades to a WebSocket and attaches a real interactive PTY shell to
// the user's sandbox container. Query: ?kind=git  OR  ?task=<id>.
func (h *Handler) TermWS(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", 401)
		return
	}
	if !h.shell.Enabled() {
		http.Error(w, "sandbox disabled", 503)
		return
	}

	// The session is keyed by lesson (?lesson=) so the terminal, every step and
	// every check share one container. ?task= is still accepted and resolved to
	// its lesson for older links.
	var image, setup, key string
	q := r.URL.Query()
	switch {
	case q.Get("kind") == "git":
		key, image, setup = gitKey, gitImage, gitSetup
	case q.Get("lesson") != "":
		lessonID, err := strconv.Atoi(q.Get("lesson"))
		if err != nil {
			http.Error(w, "bad lesson id", 400)
			return
		}
		key = labKey(lessonID)
		image, setup = h.labSandbox(r, lessonID)
	default:
		id, err := strconv.Atoi(q.Get("task"))
		if err != nil {
			http.Error(w, "bad task id", 400)
			return
		}
		task, err := h.lessonRepo.GetTaskByID(r.Context(), id)
		if err != nil {
			http.Error(w, "task not found", 404)
			return
		}
		key = labKey(task.LessonID)
		image, setup = h.labSandbox(r, task.LessonID)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = conn.WriteMessage(websocket.TextMessage, []byte(labBanner))

	// The sandbox is started AFTER the upgrade so that a failure can be
	// explained inside the terminal. Failing before it leaves the browser with
	// nothing but a closed socket — the "terminal just does not start" case.
	_ = conn.WriteMessage(websocket.TextMessage, []byte("\x1b[90mЗапускаю песочницу…\x1b[0m\r\n"))
	container, err := h.shell.EnsureSession(r.Context(), user.ID, key, image, setup)
	if err != nil {
		h.log.Error("term ensure session", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage,
			[]byte("\r\n\x1b[31m"+err.Error()+"\x1b[0m\r\n"))
		return
	}

	// Open the PTY at the size the browser measured (passed in the query) so the
	// shell draws its very first prompt at the right column count — no initial
	// mismatch window where input looks "eaten".
	cols := atoiDefault(q.Get("cols"), 100)
	rows := atoiDefault(q.Get("rows"), 28)
	if cols < 20 || cols > 500 {
		cols = 100
	}
	if rows < 5 || rows > 200 {
		rows = 28
	}
	pty, err := h.shell.OpenPTY(container, cols, rows)
	if err != nil {
		h.log.Error("term open pty", "error", err)
		_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[31mНе удалось запустить терминал: "+err.Error()+"\x1b[0m\r\n"))
		return
	}
	defer pty.Close()

	// PTY output -> WebSocket
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pty.Stdout.Read(buf)
			if n > 0 {
				if e := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); e != nil {
					return
				}
			}
			if err != nil {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\n\x1b[90m[сессия завершена]\x1b[0m\r\n"))
				_ = conn.Close()
				return
			}
		}
	}()

	// WebSocket input -> PTY (text frames may carry a resize control message)
	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if mt == websocket.TextMessage {
			var ctl struct {
				Resize []int `json:"resize"`
			}
			if json.Unmarshal(data, &ctl) == nil && len(ctl.Resize) == 2 {
				pty.Resize(ctl.Resize[1], ctl.Resize[0]) // rows, cols
				continue
			}
		}
		if _, err := pty.Stdin.Write(data); err != nil {
			return
		}
	}
}
