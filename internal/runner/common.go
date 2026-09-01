package runner

import (
	"io"
	"os"
	"strings"
	"time"
)

// Shared sandbox helpers. These used to live in shell.go alongside the container
// ShellRunner; that backend was removed once every course moved to Firecracker
// micro-VMs (see VMRunner), so the pieces the VM runner still needs live here.

const (
	// shellSessionTTL reaps a session after this much *idle* time (no exec, no
	// keystrokes) — cleans up abandoned tabs.
	shellSessionTTL = 30 * time.Minute
	// shellSessionMax is a hard cap from session start: even a busy session is
	// torn down after this, so nobody holds a sandbox forever.
	shellSessionMax = 1 * time.Hour
	maxShellOutput  = 64 * 1024
)

// shellBool reads a boolean-ish environment flag.
func shellBool(k string) bool {
	switch strings.ToLower(os.Getenv(k)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func shellEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// wrap persists the working directory between exec calls, since each exec is a
// fresh process that would otherwise reset cwd.
func wrap(command string) string {
	return `cd "$(cat /root/.gl_cwd 2>/dev/null || echo /root)" 2>/dev/null` + "\n" +
		command + "\n" +
		`__rc=$?` + "\n" +
		`pwd > /root/.gl_cwd 2>/dev/null` + "\n" +
		`exit $__rc` + "\n"
}

// FSEntry is one directory child for the in-lab file editor's tree.
type FSEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
}

// PTYSession is a live interactive shell into a sandbox. The transport is hidden
// behind resize/closer closures.
type PTYSession struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	resize func(rows, cols int)
	closer func()
}

func (p *PTYSession) Resize(rows, cols int) {
	if p.resize != nil {
		p.resize(rows, cols)
	}
}

func (p *PTYSession) Close() {
	if p.closer != nil {
		p.closer()
	}
}
