package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// Same-origin app behind nginx basic-auth; the page is already authenticated.
	CheckOrigin: func(r *http.Request) bool { return true },
}

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

	var image, setup string
	var taskID int
	if r.URL.Query().Get("kind") == "git" {
		taskID, image, setup = gitTaskID, gitImage, gitSetup
	} else {
		id, err := strconv.Atoi(r.URL.Query().Get("task"))
		if err != nil {
			http.Error(w, "bad task id", 400)
			return
		}
		task, err := h.lessonRepo.GetTaskByID(r.Context(), id)
		if err != nil {
			http.Error(w, "task not found", 404)
			return
		}
		taskID, image, setup = id, task.SandboxImage, task.SetupScript
	}

	container, err := h.shell.EnsureSession(r.Context(), user.ID, taskID, image, setup)
	if err != nil {
		h.log.Error("term ensure session", "error", err)
		http.Error(w, "sandbox start failed", 500)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	pty, err := h.shell.OpenPTY(container, 100, 28)
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
