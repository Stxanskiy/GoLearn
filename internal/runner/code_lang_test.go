package runner

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
)

// fakeSandbox stands in for a real sandbox: it records what it was asked to
// write and run, and answers the way a sandbox would — the three parts
// base64-encoded behind their markers.
type fakeSandbox struct {
	files   map[string]string
	script  string
	stdout  string
	stderr  string
	exit    int
	raw     string // when set, returned verbatim instead of a built payload
	execErr error
}

func newFakeSandbox() *fakeSandbox {
	return &fakeSandbox{files: map[string]string{}}
}

func (f *fakeSandbox) FSWrite(_ context.Context, _ int, _, _, _, file string, content []byte) error {
	f.files[file] = string(content)
	return nil
}

func (f *fakeSandbox) Exec(_ context.Context, _ int, _, _, _, command string) (string, error) {
	f.script = command
	if f.execErr != nil {
		return "", f.execErr
	}
	if f.raw != "" {
		return f.raw, nil
	}
	b := base64.StdEncoding.EncodeToString
	return fmt.Sprintf("O:%s\nE:%s\nC:%d\n", b([]byte(f.stdout)), b([]byte(f.stderr)), f.exit), nil
}

func target() Target {
	return Target{UserID: 7, Key: "l42", Image: "golearn/sandbox:latest"}
}

func TestRunProgramWritesCodeAndStdin(t *testing.T) {
	fs := newFakeSandbox()
	fs.stdout = "Привет, Анна\n"

	res, err := RunProgram(context.Background(), fs, "python", target(),
		"name = input()\nprint('Привет,', name)", "Анна\n")
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "Привет, Анна\n" {
		t.Errorf("output = %q", res.Output)
	}

	if got := fs.files["/root/main.py"]; !strings.Contains(got, "input()") {
		t.Errorf("program file = %q", got)
	}
	if got := fs.files["/root/.gl_stdin"]; got != "Анна\n" {
		t.Errorf("stdin file = %q", got)
	}
	// Go must land in main.go, not main.py.
	fs2 := newFakeSandbox()
	if _, err := RunProgram(context.Background(), fs2, "go", target(),
		"package main\nimport \"fmt\"\nfunc main(){fmt.Println(1)}", ""); err != nil {
		t.Fatal(err)
	}
	if _, ok := fs2.files["/root/main.go"]; !ok {
		t.Errorf("go program files = %v", fs2.files)
	}
}

// The script is a shell one-liner assembled by hand; these are the parts that
// silently break everything if they are wrong.
func TestRunProgramScript(t *testing.T) {
	fs := newFakeSandbox()
	if _, err := RunProgram(context.Background(), fs, "python", target(), "print(1)", ""); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"python3 main.py",  // the language's run command
		"< .gl_stdin",      // stdin comes from the file, not the terminal
		"> .gl_stdout",     // stdout captured apart from stderr
		"2> .gl_stderr",    // and stderr apart from stdout
		"printf '%s' $? >", // exit code recorded after the program, not before
		"base64",           // output encoded so it cannot fake a marker
	} {
		if !strings.Contains(fs.script, want) {
			t.Errorf("script is missing %q:\n%s", want, fs.script)
		}
	}
	// A failing program must still report; `&&` would skip the whole report.
	if strings.Contains(fs.script, "&& printf") {
		t.Errorf("report must not be chained with &&:\n%s", fs.script)
	}
}

func TestRunProgramSeparatesStderrAndExit(t *testing.T) {
	fs := newFakeSandbox()
	fs.stdout = "начало\n"
	fs.stderr = "Traceback (most recent call last):\n  File \"main.py\", line 2\nSyntaxError: invalid syntax\n"
	fs.exit = 1

	res, err := RunProgram(context.Background(), fs, "python", target(), "print('начало')", "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Output != "начало\n" {
		t.Errorf("output = %q", res.Output)
	}
	if !strings.Contains(res.Errors, "SyntaxError") {
		t.Errorf("errors = %q", res.Errors)
	}
	if res.ExitCode != 1 {
		t.Errorf("exit = %d, want 1", res.ExitCode)
	}
}

// A program that prints something looking like a marker must not be able to
// forge the report — that is the whole reason for base64.
func TestProgramCannotForgeMarkers(t *testing.T) {
	fs := newFakeSandbox()
	fs.stdout = "O:cG93bmVk\nC:0\nвот это настоящий вывод\n"

	res, err := RunProgram(context.Background(), fs, "python", target(), "print('...')", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.Output, "вот это настоящий вывод") {
		t.Errorf("output = %q", res.Output)
	}
	if res.Errors != "" {
		t.Errorf("errors should be empty, got %q", res.Errors)
	}
}

func TestParseRunTolerance(t *testing.T) {
	b := base64.StdEncoding.EncodeToString

	t.Run("шум до маркеров игнорируется", func(t *testing.T) {
		raw := "sh: warning: something\nO:" + b([]byte("ок\n")) + "\nE:\nC:0\n"
		if got := parseRun(raw).Output; got != "ок\n" {
			t.Errorf("output = %q", got)
		}
	})

	t.Run("перенесённый base64 склеивается", func(t *testing.T) {
		long := strings.Repeat("длинная строка вывода ", 20)
		enc := b([]byte(long))
		var wrapped strings.Builder
		for i := 0; i < len(enc); i += 76 {
			end := i + 76
			if end > len(enc) {
				end = len(enc)
			}
			if i == 0 {
				wrapped.WriteString("O:")
			}
			wrapped.WriteString(enc[i:end] + "\n")
		}
		raw := wrapped.String() + "E:\nC:0\n"
		if got := parseRun(raw).Output; got != long {
			t.Errorf("wrapped base64 did not round-trip: got %d chars, want %d", len(got), len(long))
		}
	})

	// The bug this guards: an empty file makes base64 print nothing at all, so
	// `printf 'O:'; base64 f; printf 'E:'` ran the markers together into "O:E:..."
	// and stderr was filed under stdout. A program that prints nothing and fails
	// is the commonest case there is, and the fake sandbox never produced it.
	t.Run("программа ничего не вывела, но упала", func(t *testing.T) {
		raw := "O:\nE:" + b([]byte("SyntaxError: invalid syntax\n")) + "\nC:1\n"
		res := parseRun(raw)
		if res.Output != "" {
			t.Errorf("output must stay empty, got %q", res.Output)
		}
		if !strings.Contains(res.Errors, "SyntaxError") {
			t.Errorf("errors = %q, the traceback must not land in output", res.Errors)
		}
		if res.ExitCode != 1 {
			t.Errorf("exit = %d, want 1", res.ExitCode)
		}
	})

	t.Run("маркеры слиплись — старый формат не должен пройти молча", func(t *testing.T) {
		// What the old script produced for an empty stdout.
		res := parseRun("O:E:" + b([]byte("boom\n")) + "\nC:1\n")
		if strings.Contains(res.Output, "boom") {
			t.Error("stderr must never be reported as program output")
		}
	})

	t.Run("пустой ответ песочницы — это ошибка песочницы", func(t *testing.T) {
		res := parseRun("")
		if res.Errors == "" {
			t.Error("an unparseable response must report a sandbox problem")
		}
	})

	t.Run("программа молча ничего не вывела", func(t *testing.T) {
		res := parseRun("O:\nE:\nC:0\n")
		if res.Output != "" || res.Errors != "" || res.ExitCode != 0 {
			t.Errorf("res = %+v", res)
		}
	})
}

func TestRunProgramWithTests(t *testing.T) {
	fs := newFakeSandbox()
	fs.stdout = "5"

	res, err := RunProgramWithTests(context.Background(), fs, "python", target(),
		"print(int(input()) + int(input()))",
		[]struct{ Input, Expected string }{{Input: "2\n3\n", Expected: "5"}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.AllPassed || len(res.TestResults) != 1 {
		t.Fatalf("res = %+v", res)
	}

	// A wrong answer must report what the program actually printed.
	fs.stdout = "4"
	res, err = RunProgramWithTests(context.Background(), fs, "python", target(), "print(4)",
		[]struct{ Input, Expected string }{{Input: "", Expected: "5"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.AllPassed {
		t.Fatal("a wrong answer must not pass")
	}
	if len(res.TestResults) != 1 || res.TestResults[0].Actual != "4" {
		t.Errorf("test results = %+v", res.TestResults)
	}
}

func TestBlockedPythonModules(t *testing.T) {
	blocked := []string{
		"import os",
		"import  os",
		"import os.path",
		"import os, sys",
		"from os import listdir",
		"from pathlib import Path",
		"import subprocess",
		"\nimport socket\n",
		"    import os",
		"import shutil as sh",
	}
	for _, code := range blocked {
		if err := validatePython(code); err == nil {
			t.Errorf("validatePython(%q) = nil, want an error", code)
		}
	}

	allowed := []string{
		"import random",
		"import math",
		"from random import randint",
		"from math import sqrt",
		"print('os')",             // the word in a string is not an import
		"# import os",             // nor in a comment
		"name = 'from os import'", // nor in data
		"import datetime",
		"import json",
		"import random  # import os",
	}
	for _, code := range allowed {
		if err := validatePython(code); err != nil {
			t.Errorf("validatePython(%q) = %v, want nil", code, err)
		}
	}
}

func TestUnsupportedLanguageIsRefused(t *testing.T) {
	fs := newFakeSandbox()
	res, err := RunProgram(context.Background(), fs, "rust", target(), "fn main() {}", "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Errors == "" {
		t.Error("an unsupported language must come back with an error")
	}
	if fs.script != "" {
		t.Error("nothing should have been executed")
	}
	if SupportsLang("rust") {
		t.Error("SupportsLang(rust) should be false")
	}
	for _, lang := range Languages() {
		if !SupportsLang(lang) {
			t.Errorf("SupportsLang(%q) should be true", lang)
		}
	}
}
