package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RunRequest struct {
	Code  string `json:"code"`
	Stdin string `json:"stdin"`
}

type RunResponse struct {
	Output      string       `json:"output"`
	Errors      string       `json:"errors"`
	ExitCode    int          `json:"exit_code"`
	TimedOut    bool         `json:"timed_out"`
	TestResults []TestResult `json:"test_results,omitempty"`
	AllPassed   bool         `json:"all_passed"`
}

type TestResult struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
}

// RunCode handles POST /api/run — runs user code in sandbox.
func (h *Handler) RunCode(w http.ResponseWriter, r *http.Request) {
	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if req.Code == "" {
		writeJSON(w, RunResponse{Errors: "No code provided"})
		return
	}

	result, err := h.runner.Run(r.Context(), req.Code, req.Stdin)
	if err != nil {
		h.log.Error("run code", "error", err)
		http.Error(w, "Runner error", 500)
		return
	}

	writeJSON(w, RunResponse{
		Output:   result.Output,
		Errors:   result.Errors,
		ExitCode: result.ExitCode,
		TimedOut: result.TimedOut,
	})
}

// RunTaskCode handles POST /api/run/{taskID} — runs code and checks against task test cases.
func (h *Handler) RunTaskCode(w http.ResponseWriter, r *http.Request) {
	taskIDStr := chi.URLParam(r, "taskID")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", 400)
		return
	}

	var req RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	if req.Code == "" {
		writeJSON(w, RunResponse{Errors: "No code provided"})
		return
	}

	// Get task to find test cases
	task, err := h.lessonRepo.GetTaskByID(r.Context(), taskID)
	if err != nil {
		http.Error(w, "Task not found", 404)
		return
	}

	if len(task.TestCases) == 0 {
		// No tests — just run
		result, err := h.runner.Run(r.Context(), req.Code, req.Stdin)
		if err != nil {
			h.log.Error("run code", "error", err)
			http.Error(w, "Runner error", 500)
			return
		}
		writeJSON(w, RunResponse{
			Output:   result.Output,
			Errors:   result.Errors,
			ExitCode: result.ExitCode,
			TimedOut: result.TimedOut,
		})
		return
	}

	// Run with test cases
	var tests []struct{ Input, Expected string }
	for _, tc := range task.TestCases {
		tests = append(tests, struct{ Input, Expected string }{tc.Input, tc.ExpectedOutput})
	}

	runResult, err := h.runner.RunWithTests(r.Context(), req.Code, tests)
	if err != nil {
		h.log.Error("run code with tests", "error", err)
		http.Error(w, "Runner error", 500)
		return
	}

	resp := RunResponse{
		Output:    runResult.Result.Output,
		Errors:    runResult.Result.Errors,
		ExitCode:  runResult.Result.ExitCode,
		TimedOut:  runResult.Result.TimedOut,
		AllPassed: runResult.AllPassed,
	}
	for _, tr := range runResult.TestResults {
		resp.TestResults = append(resp.TestResults, TestResult{
			Input:    tr.Input,
			Expected: tr.Expected,
			Actual:   tr.Actual,
			Passed:   tr.Passed,
		})
	}

	// Save submission
	if err := h.submissionRepo.Save(r.Context(), taskID, req.Code, runResult.Result.Output, runResult.Result.Errors, runResult.AllPassed); err != nil {
		h.log.Error("save submission", "error", err)
	}

	// Auto-progress: update lesson status when tests pass
	if runResult.AllPassed {
		allDone, err := h.submissionRepo.AllLessonTasksPassed(r.Context(), task.LessonID)
		if err != nil {
			h.log.Error("check lesson progress", "error", err)
		} else if allDone {
			_ = h.progressRepo.Upsert(r.Context(), task.LessonID, "completed")
		} else {
			_ = h.progressRepo.Upsert(r.Context(), task.LessonID, "in_progress")
		}
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
