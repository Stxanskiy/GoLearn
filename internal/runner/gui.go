package runner

import (
	"context"
	"fmt"
	"strings"
)

// Running a windowed program for a student to look at.
//
// The program serves its own VNC: Qt ships a VNC platform plugin
// (plugins/platforms/libqvnc.so, present in the PyQt6 wheel), so
// QT_QPA_PLATFORM=vnc makes the student's process listen on 5900 and speak RFB
// itself. No Xvfb, no x11vnc, no desktop — one process, measured at ~44 MB.
//
// The lifecycle differs from a text task: a window program never exits, so the
// run cannot be "execute and collect output". It is started detached, we wait
// for the port, and the browser then attaches over the WebSocket relay.

// GUIResult is what a student gets back from starting a windowed program.
type GUIResult struct {
	// Running is false when the program failed before it could show a window.
	Running bool `json:"running"`
	// Errors is the program's stderr. For a window program this is the only
	// place a traceback can appear, so it has to be reported even on success:
	// a warning does not stop the window from opening.
	Errors string `json:"errors"`
	Output string `json:"output"`
	// Port is where the program's VNC server is listening inside the sandbox.
	Port int `json:"port"`
}

const (
	guiOutFile = ".gl_gui_out"
	guiErrFile = ".gl_gui_err"
	guiPIDFile = ".gl_gui_pid"
	// How long to wait for the window to come up. Qt starts in well under a
	// second; the budget is for a cold sandbox.
	guiWaitTries = 40
)

// StartGUI writes the program into the sandbox and starts it detached, then
// waits for its VNC port. A previous run is killed first, so pressing Run twice
// does not leave two windows fighting over the port.
func StartGUI(ctx context.Context, e CodeSandbox, lang string, t Target, code string) (*GUIResult, error) {
	spec, ok := langs[lang]
	if !ok || !spec.gui {
		return nil, fmt.Errorf("%q is not a windowed language", lang)
	}
	if len(code) > maxCodeSize {
		return &GUIResult{Errors: "Программа слишком большая (максимум 64 КБ)."}, nil
	}
	if err := validateCode(lang, code); err != nil {
		return &GUIResult{Errors: err.Error()}, nil
	}

	if err := e.FSWrite(ctx, t.UserID, t.Key, t.Sandbox, codeDir+"/"+spec.file, []byte(code)); err != nil {
		return nil, fmt.Errorf("write program: %w", err)
	}

	// The port check uses python rather than bash's /dev/tcp: python is the one
	// thing a Python sandbox is guaranteed to have.
	script := fmt.Sprintf(`cd %s
if [ -f %s ]; then kill $(cat %s) 2>/dev/null; fi
rm -f %s %s %s
%s > %s 2> %s &
echo $! > %s
for i in $(seq 1 %d); do
  if python3 -c "import socket,sys; socket.create_connection(('127.0.0.1',%d),0.3)" 2>/dev/null; then
    printf 'S:ready\n'; break
  fi
  if ! kill -0 $(cat %s) 2>/dev/null; then printf 'S:died\n'; break; fi
  sleep 0.25
done
printf 'O:\n'; base64 %s 2>/dev/null; printf 'E:\n'; base64 %s 2>/dev/null; printf '\n'`,
		codeDir,
		guiPIDFile, guiPIDFile,
		guiOutFile, guiErrFile, guiPIDFile,
		spec.run, guiOutFile, guiErrFile,
		guiPIDFile,
		guiWaitTries, spec.port,
		guiPIDFile,
		guiOutFile, guiErrFile)

	raw, err := e.Exec(ctx, t.UserID, t.Key, t.Sandbox, script)
	if err != nil {
		return nil, err
	}

	res := parseGUI(raw)
	res.Port = spec.port
	return res, nil
}

func parseGUI(raw string) *GUIResult {
	res := &GUIResult{}
	var out, errs strings.Builder
	cur := (*strings.Builder)(nil)

	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "S:"):
			res.Running = strings.TrimSpace(strings.TrimPrefix(line, "S:")) == "ready"
			cur = nil
			continue
		case strings.HasPrefix(line, "O:"):
			cur, line = &out, strings.TrimPrefix(line, "O:")
		case strings.HasPrefix(line, "E:"):
			cur, line = &errs, strings.TrimPrefix(line, "E:")
		}
		if cur != nil {
			cur.WriteString(strings.TrimSpace(line))
		}
	}

	res.Output = truncate(unb64(out.String()), maxOutputSize)
	res.Errors = truncate(dropQtNoise(unb64(errs.String())), maxOutputSize)

	// No window and nothing on stderr: the wait ran out. Say which it was,
	// because "nothing happened" is the least useful thing to show a student.
	if !res.Running && res.Errors == "" {
		res.Errors = "Окно не открылось, и программа не сообщила об ошибке. " +
			"Проверь, что в конце есть app.exec() — без него программа завершается сразу."
	}
	return res
}

// dropQtNoise removes what Qt itself says on startup. "QVncServer created on
// port 5900" is the library announcing our own plumbing; showing it to a
// student under the heading "Errors" would teach them to ignore that panel,
// which is the one place their traceback appears.
func dropQtNoise(s string) string {
	if s == "" {
		return ""
	}
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "QVncServer created on port") {
			continue
		}
		kept = append(kept, line)
	}
	out := strings.Join(kept, "\n")
	if strings.TrimSpace(out) == "" {
		return ""
	}
	return out
}

// StopGUI kills a running window program. Called when the student leaves the
// lesson, so a forgotten window does not hold memory for an hour.
func StopGUI(ctx context.Context, e CodeSandbox, t Target) error {
	script := fmt.Sprintf("cd %s; if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi; printf 'ok'",
		codeDir, guiPIDFile, guiPIDFile, guiPIDFile)
	_, err := e.Exec(ctx, t.UserID, t.Key, t.Sandbox, script)
	return err
}
