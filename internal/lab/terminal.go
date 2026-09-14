package lab

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/backendraz/golearn/internal/runner"
	"github.com/gorilla/websocket"
)

// Banner is shown when a sandbox terminal opens (ANSI cyan, CRLF endings).
const Banner = "\x1b[1;36m" +
	"\r\n ████████   ██████   ████████" +
	"\r\n    ██     ██    ██     ██   " +
	"\r\n    ██     ██    ██     ██   " +
	"\r\n    ██     ██    ██     ██   " +
	"\r\n    ██      ██████      ██   " +
	"\x1b[0m\r\n\x1b[90m Welcome to your TOT lab environment!\x1b[0m\r\n\r\n"

const (
	pingInterval  = 30 * time.Second
	writeWait     = 10 * time.Second
	touchInterval = 30 * time.Second
)

// Pump bridges a WebSocket and a PTY until either side closes; touch is called at most every 30s while the user types.
func Pump(conn *websocket.Conn, pty *runner.PTYSession, touch func()) {
	var mu sync.Mutex
	write := func(mt int, data []byte) error {
		mu.Lock()
		defer mu.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(mt, data)
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		t := time.NewTicker(pingInterval)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				if write(websocket.PingMessage, nil) != nil {
					return
				}
			}
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pty.Stdout.Read(buf)
			if n > 0 && write(websocket.BinaryMessage, buf[:n]) != nil {
				return
			}
			if err != nil {
				_ = write(websocket.TextMessage, []byte("\r\n\x1b[90m[сессия завершена]\x1b[0m\r\n"))
				_ = conn.Close()
				return
			}
		}
	}()

	var lastTouch time.Time
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
				pty.Resize(ctl.Resize[1], ctl.Resize[0])
				continue
			}
		}
		if touch != nil && time.Since(lastTouch) > touchInterval {
			touch()
			lastTouch = time.Now()
		}
		if _, err := pty.Stdin.Write(data); err != nil {
			return
		}
	}
}
