package runner

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Running a student's program on top of Engine, so code tasks use the same
// sandbox as shell labs: a Firecracker micro-VM in production, a container
// locally. The separate Docker container this used to have never worked in
// production — the app pod has no docker socket and the image has no client —
// which is why it only ever ran on a developer's laptop.
//
// Everything language-specific is in langs (see lang.go). Everything else here
// is the same whatever the language.

// CodeSandbox is the part of Engine a program run needs: put the file in, run
// one command. Taking the narrow interface keeps the API's own sandbox type
// usable here without widening it.
type CodeSandbox interface {
	FSWrite(ctx context.Context, userID int, key string, spec Spec, file string, content []byte) error
	Exec(ctx context.Context, userID int, key string, spec Spec, command string) (string, error)
}

// Both runners satisfy it, through Engine.
var (
	_ CodeSandbox = (*ShellRunner)(nil)
	_ CodeSandbox = (*VMRunner)(nil)
)

// Target identifies whose sandbox to run in. It mirrors the arguments every
// other Engine call takes.
type Target struct {
	UserID  int
	Key     string
	Sandbox Spec
}

const (
	// Files live in the session's home directory. The leading dot keeps them out
	// of the student's way in the file tree.
	codeDir  = "/root"
	inFile   = ".gl_stdin"
	outFile  = ".gl_stdout"
	errFile  = ".gl_stderr"
	codeFile = ".gl_exit"
)

// RunProgram writes code into the session sandbox, runs it and returns stdout,
// stderr and the exit code.
//
// stdout, stderr and the exit status come back in one Exec call, base64-encoded:
// a program's own output cannot then be mistaken for a delimiter, and a VM run
// costs one SSH round trip instead of four.
func RunProgram(ctx context.Context, e CodeSandbox, lang string, t Target, code, stdin string) (*Result, error) {
	spec, ok := langs[lang]
	if !ok {
		return &Result{Errors: fmt.Sprintf("Запуск кода для %q не поддерживается.", lang)}, nil
	}
	if spec.gui {
		// A window program never exits; it is started by StartGUI instead.
		return nil, fmt.Errorf("%q opens a window and cannot be run to completion", lang)
	}
	if len(code) > maxCodeSize {
		return &Result{Errors: "Программа слишком большая (максимум 64 КБ)."}, nil
	}
	if err := validateCode(lang, code); err != nil {
		return &Result{Errors: err.Error()}, nil
	}

	if err := e.FSWrite(ctx, t.UserID, t.Key, t.Sandbox, codeDir+"/"+spec.file, []byte(code)); err != nil {
		return nil, fmt.Errorf("write program: %w", err)
	}
	if err := e.FSWrite(ctx, t.UserID, t.Key, t.Sandbox, codeDir+"/"+inFile, []byte(stdin)); err != nil {
		return nil, fmt.Errorf("write stdin: %w", err)
	}

	// `; ` and not `&& `: the exit code of a program that failed is the point.
	script := fmt.Sprintf(
		"cd %s; %s < %s > %s 2> %s; printf '%%s' $? > %s; "+
			// Each marker on its own line: an empty file makes base64 print nothing
			// at all, not even a newline, so markers would run together as "O:E:..."
			// and the parser would file stderr under stdout.
			"printf 'O:\\n'; base64 %s; printf 'E:\\n'; base64 %s; printf 'C:\\n'; cat %s; printf '\\n'",
		codeDir, spec.run, inFile, outFile, errFile, codeFile,
		outFile, errFile, codeFile)

	raw, err := e.Exec(ctx, t.UserID, t.Key, t.Sandbox, script)
	if err != nil {
		return nil, err
	}
	return parseRun(raw), nil
}

// parseRun pulls the three parts out of the payload. Anything the sandbox wrote
// before the first marker (a shell notice, say) is ignored.
func parseRun(raw string) *Result {
	var out, errs, exit strings.Builder
	cur := (*strings.Builder)(nil)

	for _, line := range strings.Split(raw, "\n") {
		switch {
		case strings.HasPrefix(line, "O:"):
			cur, line = &out, strings.TrimPrefix(line, "O:")
		case strings.HasPrefix(line, "E:"):
			cur, line = &errs, strings.TrimPrefix(line, "E:")
		case strings.HasPrefix(line, "C:"):
			cur, line = &exit, strings.TrimPrefix(line, "C:")
		}
		if cur != nil {
			cur.WriteString(strings.TrimSpace(line))
		}
	}

	res := &Result{
		Output: truncate(unb64(out.String()), maxOutputSize),
		Errors: truncate(unb64(errs.String()), maxOutputSize),
	}
	if n, err := strconv.Atoi(strings.TrimSpace(exit.String())); err == nil {
		res.ExitCode = n
	}
	// A sandbox that returned nothing recognisable is a sandbox problem, not a
	// program that printed nothing: say so rather than showing an empty result.
	if out.Len() == 0 && errs.Len() == 0 && exit.Len() == 0 {
		res.Errors = "Песочница не ответила. Попробуй запустить ещё раз."
	}
	return res
}

func unb64(s string) string {
	// base64 may arrive wrapped, depending on the sandbox's coreutils.
	s = strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(s)
	if s == "" {
		return ""
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(b)
}

// RunProgramWithTests runs the program once per test case and compares its
// stdout with the expected value, the way every beginner course checks work.
func RunProgramWithTests(ctx context.Context, e CodeSandbox, lang string, t Target, code string,
	tests []struct{ Input, Expected string }) (*RunResult, error) {

	result := &RunResult{AllPassed: true}

	if len(tests) == 0 {
		res, err := RunProgram(ctx, e, lang, t, code, "")
		if err != nil {
			return nil, err
		}
		result.Result = *res
		result.CompileOK = res.Errors == "" || res.ExitCode == 0
		return result, nil
	}

	for _, tc := range tests {
		res, err := RunProgram(ctx, e, lang, t, code, tc.Input)
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
