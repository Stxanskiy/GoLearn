package runner

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

const (
	maxCodeSize   = 64 * 1024 // 64KB
	maxOutputSize = 32 * 1024 // 32KB
)

// langSpec is everything that differs between languages: what the program file
// is called and how it is run. Everything else — writing it into the sandbox,
// feeding stdin, collecting stdout and stderr, comparing with the expected
// value — is the same for all of them (see code.go).
type langSpec struct {
	file string
	run  string
	// gui marks a language whose program opens a window instead of printing and
	// exiting. Those are started detached and watched for a port (see gui.go);
	// they cannot be run to completion, so RunProgram refuses them.
	gui  bool
	port int
}

var langs = map[string]langSpec{
	"go":     {file: "main.go", run: "go run main.go"},
	"python": {file: "main.py", run: "python3 main.py"},
	// Qt's own VNC platform plugin: the student's program becomes the VNC
	// server, so showing a window costs one process and no display stack.
	"python-gui": {
		file: "main.py",
		run:  "QT_QPA_PLATFORM=vnc:size=800x600 python3 main.py",
		gui:  true,
		port: 5900,
	},
}

// Languages the code runner can execute, for validation elsewhere.
func Languages() []string { return []string{"go", "python", "python-gui"} }

// IsWindowed reports whether a program of this kind opens a window instead of
// printing and exiting.
func IsWindowed(lang string) bool { return langs[lang].gui }

// SupportsLang reports whether code of this kind can be run.
func SupportsLang(lang string) bool { _, ok := langs[lang]; return ok }

// Result of running user code.
type Result struct {
	Output   string `json:"output"`
	Errors   string `json:"errors"`
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out"`
}

// TestResult for checking against expected output.
type TestResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
}

// RunResult combines compilation + test results.
type RunResult struct {
	CompileOK   bool         `json:"compile_ok"`
	Result      Result       `json:"result"`
	TestResults []TestResult `json:"test_results,omitempty"`
	AllPassed   bool         `json:"all_passed"`
}

// pythonBlocked are modules that would let one program reach outside its own
// workspace. The container has no network and is wiped per run, but it is
// shared between students while a run is in flight, so the filesystem and
// process modules stay closed.
//
// Courses that genuinely need files (reading and writing, pathlib) will have to
// either get a per-run container or an explicit per-task allowance — decide that
// before shipping the lesson, not by widening this list.
var pythonBlocked = []string{
	"os", "sys", "subprocess", "shutil", "pathlib", "socket", "ctypes",
	"importlib", "multiprocessing", "threading", "signal", "resource",
}

// validatePython rejects the modules above and nothing else. Syntax is not
// pre-checked: for a beginner the interpreter's own SyntaxError, with its line
// number and caret, is more useful than anything we could write here.
//
// The line is split into fields rather than matched by prefix, so "import  os"
// with two spaces cannot slip past — an earlier version of this check did.
func validatePython(code string) error {
	blocked := make(map[string]bool, len(pythonBlocked))
	for _, m := range pythonBlocked {
		blocked[m] = true
	}

	for _, line := range strings.Split(code, "\n") {
		// Cut an end-of-line comment, then split on any run of whitespace.
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		var mods []string
		switch fields[0] {
		case "import":
			// import a, b.c as d  ->  a, b
			for _, part := range strings.Split(strings.Join(fields[1:], " "), ",") {
				name := strings.Fields(part) // drops "as alias"
				if len(name) > 0 {
					mods = append(mods, strings.SplitN(name[0], ".", 2)[0])
				}
			}
		case "from":
			// from a.b import x  ->  a
			mods = append(mods, strings.SplitN(fields[1], ".", 2)[0])
		default:
			continue
		}

		for _, mod := range mods {
			if blocked[mod] {
				return fmt.Errorf("Модуль %s в этой песочнице закрыт. Для заданий курса он не нужен — если кажется, что нужен, перечитай условие.", mod)
			}
		}
	}
	return nil
}

// allowedImports is the whitelist of packages students can use.
var allowedImports = map[string]bool{
	"fmt":            true,
	"strings":        true,
	"strconv":        true,
	"math":           true,
	"math/rand":      true,
	"sort":           true,
	"errors":         true,
	"unicode":        true,
	"unicode/utf8":   true,
	"bytes":          true,
	"regexp":         true,
	"bufio":          true,
	"io":             true,
	"time":           true,
	"slices":         true,
	"maps":           true,
	"cmp":            true,
	"encoding/json":  true,
	"encoding/csv":   true,
	"log":            true,
	"math/big":       true,
	"container/heap": true,
	"container/list": true,
	"os":             true,
	"path/filepath":  true,
	"sync":           true,
	"sync/atomic":    true,
	"context":        true,
}

func validateCode(lang, code string) error {
	if lang == "python" || lang == "python-gui" {
		return validatePython(code)
	}
	if !strings.Contains(code, "package main") {
		return fmt.Errorf("Код должен содержать 'package main'. Каждая исполняемая Go-программа начинается с package main.")
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", code, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("Ошибка синтаксиса: %v. Проверь скобки, кавычки и точки с запятой.", err)
	}

	if f.Name.Name != "main" {
		return fmt.Errorf("Пакет должен быть 'main', а не '%s'. Исполняемая программа всегда в package main.", f.Name.Name)
	}

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if !allowedImports[path] {
			return fmt.Errorf("Запрещённый импорт: %s. Разрешены: fmt, strings, strconv, math, sort, errors, unicode, bytes, regexp, bufio, io, time, slices, encoding/json, sync, context и другие безопасные пакеты.", path)
		}
	}

	if !strings.Contains(code, "func main()") {
		return fmt.Errorf("Код должен содержать 'func main()'. Это точка входа — с неё начинается выполнение программы.")
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "\n... (output truncated)"
	}
	return s
}
