package runner

import (
	"bytes"
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	maxCodeSize    = 64 * 1024 // 64KB
	runTimeout     = 15 * time.Second
	maxOutputSize  = 32 * 1024 // 32KB
	containerImage = "golang:1.22-alpine"
	containerName  = "golearn-runner"
)

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

// Runner executes Go code in a Docker sandbox with warm build cache.
type Runner struct {
	useDocker   bool
	containerUp bool
	mu          sync.Mutex
}

// New creates a Runner. Uses persistent Docker container for fast execution.
func New() *Runner {
	r := &Runner{}

	// Check if Docker is available
	cmd := exec.Command("docker", "info")
	if cmd.Run() == nil {
		r.useDocker = true
		// Start persistent container in background
		go r.ensureContainer()
	}

	return r
}

// ensureContainer creates or starts the persistent runner container with warm build cache.
func (r *Runner) ensureContainer() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.containerUp {
		return
	}

	// Check if container already exists
	check := exec.Command("docker", "inspect", containerName)
	if check.Run() == nil {
		// Container exists — make sure it's running
		start := exec.Command("docker", "start", containerName)
		if start.Run() == nil {
			r.containerUp = true
			return
		}
		// If start failed, remove and recreate
		exec.Command("docker", "rm", "-f", containerName).Run()
	}

	// Create persistent container with security constraints
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--network=none",
		"--memory=256m",
		"--cpus=1",
		"--pids-limit=128",
		containerImage,
		"sh", "-c", "mkdir -p /workspace && sleep infinity",
	}

	cmd := exec.Command("docker", args...)
	if err := cmd.Run(); err != nil {
		return
	}

	// Warm up the Go build cache with a dummy program
	warmup := `package main
import "fmt"
func main() { fmt.Println("warm") }`

	warmCmd := exec.Command("docker", "exec", containerName,
		"sh", "-c", fmt.Sprintf("echo '%s' > /workspace/warmup.go && go run /workspace/warmup.go && rm /workspace/warmup.go", warmup))
	warmCmd.Run() // ignore error, best effort

	r.containerUp = true
}

// Close stops the persistent container.
func (r *Runner) Close() {
	if r.useDocker {
		exec.Command("docker", "rm", "-f", containerName).Run()
	}
}

// Run executes Go code and returns the result.
func (r *Runner) Run(ctx context.Context, code string, stdin string) (*Result, error) {
	if len(code) > maxCodeSize {
		return &Result{Errors: "Code too large (max 64KB)"}, nil
	}

	// Basic security checks
	if err := validateCode(code); err != nil {
		return &Result{Errors: err.Error()}, nil
	}

	if !r.useDocker {
		// Never execute untrusted code on the host — require the Docker sandbox.
		return &Result{Errors: "Песочница недоступна (Docker не запущен). Запуск кода временно отключён."}, nil
	}
	return r.runInDocker(ctx, code, stdin)
}

// RunWithTests executes code and checks against test cases.
func (r *Runner) RunWithTests(ctx context.Context, code string, tests []struct{ Input, Expected string }) (*RunResult, error) {
	result := &RunResult{AllPassed: true}

	if len(tests) == 0 {
		res, err := r.Run(ctx, code, "")
		if err != nil {
			return nil, err
		}
		result.Result = *res
		result.CompileOK = res.Errors == "" || res.ExitCode == 0
		return result, nil
	}

	for _, tc := range tests {
		res, err := r.Run(ctx, code, tc.Input)
		if err != nil {
			return nil, err
		}

		actual := strings.TrimSpace(res.Output)
		expected := strings.TrimSpace(tc.Expected)
		passed := actual == expected

		if !passed {
			result.AllPassed = false
		}

		result.TestResults = append(result.TestResults, TestResult{
			Input:    tc.Input,
			Expected: expected,
			Actual:   actual,
			Passed:   passed,
		})
		result.Result = *res
	}

	result.CompileOK = result.Result.Errors == "" || result.Result.ExitCode == 0
	return result, nil
}

func (r *Runner) runInDocker(ctx context.Context, code string, stdin string) (*Result, error) {
	// Ensure container is running
	r.ensureContainer()

	if !r.containerUp {
		return &Result{Errors: "Песочница недоступна — не удалось запустить контейнер. Попробуй позже."}, nil
	}

	// Create a unique workspace path inside the container
	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	workDir := "/workspace/" + runID

	// Write code into container
	setupCmd := exec.Command("docker", "exec", containerName,
		"sh", "-c", fmt.Sprintf("mkdir -p %s", workDir))
	if err := setupCmd.Run(); err != nil {
		// Fail closed: never fall back to running untrusted code on the host.
		return &Result{Errors: "Песочница временно недоступна. Попробуй ещё раз через минуту."}, nil
	}

	// Copy code into container via docker exec
	copyCmd := exec.Command("docker", "exec", "-i", containerName,
		"sh", "-c", fmt.Sprintf("cat > %s/main.go", workDir))
	copyCmd.Stdin = strings.NewReader(code)
	if err := copyCmd.Run(); err != nil {
		// Fail closed: never fall back to running untrusted code on the host.
		return &Result{Errors: "Песочница временно недоступна. Попробуй ещё раз через минуту."}, nil
	}

	// Execute code with timeout
	ctx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	var execArgs []string
	if stdin != "" {
		execArgs = []string{"exec", "-i", containerName,
			"sh", "-c", fmt.Sprintf("cd %s && go run main.go", workDir)}
	} else {
		execArgs = []string{"exec", containerName,
			"sh", "-c", fmt.Sprintf("cd %s && go run main.go", workDir)}
	}

	cmd := exec.CommandContext(ctx, "docker", execArgs...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Cleanup workspace
	go exec.Command("docker", "exec", containerName, "rm", "-rf", workDir).Run()

	result := &Result{
		Output: truncate(stdout.String(), maxOutputSize),
		Errors: truncate(stderr.String(), maxOutputSize),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Errors = "Timeout: program took too long (max 15 seconds)"
		return result, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err != nil {
		result.Errors = err.Error()
		result.ExitCode = 1
	}

	return result, nil
}

func (r *Runner) runLocal(ctx context.Context, code string, stdin string) (*Result, error) {
	tmpDir, err := os.MkdirTemp("", "golearn-run-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	mainFile := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainFile, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("write code: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", mainFile)

	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	result := &Result{
		Output: truncate(stdout.String(), maxOutputSize),
		Errors: truncate(stderr.String(), maxOutputSize),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Errors = "Timeout: program took too long (max 15 seconds)"
		return result, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err != nil {
		result.Errors = err.Error()
		result.ExitCode = 1
	}

	return result, nil
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

// validateCode parses the code as Go AST and checks for forbidden imports.
func validateCode(code string) error {
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
