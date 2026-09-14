package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/lab"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

// taskByID loads a task whose lesson is visible to the user, writing 404/500 itself when it returns false.
func (a *API) taskByID(w http.ResponseWriter, r *http.Request) (*model.Task, lessonRef, bool) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "taskId"))
	if err != nil {
		writeError(w, http.StatusNotFound, codeNotFound, "task not found")
		return nil, lessonRef{}, false
	}
	t, err := a.Lessons.GetTaskByID(ctx, id)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "task not found")
		return nil, lessonRef{}, false
	}
	if err != nil {
		a.internalError(w, "load task", err)
		return nil, lessonRef{}, false
	}
	l, err := a.Lessons.GetByID(ctx, t.LessonID)
	if err != nil {
		a.lessonLoadError(w, err)
		return nil, lessonRef{}, false
	}
	m, err := a.Modules.GetByID(ctx, l.ModuleID)
	if err != nil {
		a.lessonLoadError(w, err)
		return nil, lessonRef{}, false
	}
	ref, ok := a.visible(w, r, lessonRef{lesson: l, module: m})
	return t, ref, ok
}

func (a *API) checkTask(w http.ResponseWriter, r *http.Request) {
	t, ref, ok := a.taskByID(w, r)
	if !ok {
		return
	}
	if checkMode(*t) != "auto" {
		writeError(w, http.StatusConflict, codeTaskNotAutoChecked, "task has no automatic check")
		return
	}
	if !a.requireSandbox(w) {
		return
	}
	ctx := r.Context()
	uid := userFrom(ctx).ID
	if !checkLimiter.Allow(strconv.Itoa(uid)) {
		writeRateLimited(w, checkLimiter)
		return
	}
	image, setup, err := a.Lessons.LessonSandbox(ctx, t.LessonID)
	if err != nil {
		a.internalError(w, "check: sandbox config", err)
		return
	}
	passed, output, err := a.Sandbox.Check(ctx, uid, lab.Key(t.LessonID), image, setup, t.CheckScript)
	if err != nil {
		a.sandboxError(w, "check", err)
		return
	}
	a.finishTask(w, r, *t, *ref.lesson, passed, "[shell]", output, "", &output)
}

func (a *API) markTaskDone(w http.ResponseWriter, r *http.Request) {
	t, ref, ok := a.taskByID(w, r)
	if !ok {
		return
	}
	if checkMode(*t) != "manual" {
		writeError(w, http.StatusConflict, codeTaskAutoChecked, "task must pass its check")
		return
	}
	a.finishTask(w, r, *t, *ref.lesson, true, "[manual]", "", "", nil)
}

// finishTask stores a submission, updates lesson progress and responds with the result.
func (a *API) finishTask(w http.ResponseWriter, r *http.Request, t model.Task, l model.Lesson, passed bool, code, output, errs string, shownOutput *string) {
	ctx := r.Context()
	uid := userFrom(ctx).ID
	status, err := a.recordSubmission(ctx, uid, t, l, passed, code, output, errs)
	if err != nil {
		a.internalError(w, "task: record", err)
		return
	}
	writeJSON(w, http.StatusOK, apigen.TaskCheckResult{TaskID: t.ID, Passed: passed, Output: shownOutput, LessonStatus: apigen.ProgressStatus(status)})
}

// recordSubmission saves an attempt, marks the lesson completed once every task passed, and returns the derived lesson status.
func (a *API) recordSubmission(ctx context.Context, uid int, t model.Task, l model.Lesson, passed bool, code, output, errs string) (string, error) {
	if err := a.Submissions.Save(ctx, uid, t.ID, code, output, errs, passed); err != nil {
		return "", err
	}
	if passed {
		tasks, err := a.Lessons.GetTasks(ctx, l.ID)
		if err != nil {
			return "", err
		}
		done, err := a.Submissions.PassedTaskIDs(ctx, uid, l.ID)
		if err != nil {
			return "", err
		}
		status := catalog.StatusCompleted
		for _, task := range tasks {
			if !done[task.ID] {
				status = catalog.StatusInProgress
				break
			}
		}
		if err := a.Progress.Upsert(ctx, uid, l.ID, status); err != nil {
			return "", err
		}
	}
	up, err := a.userProgress(ctx, uid)
	if err != nil {
		return "", err
	}
	return up.LessonStatus(l), nil
}

func (a *API) runTaskCode(w http.ResponseWriter, r *http.Request) {
	t, ref, ok := a.taskByID(w, r)
	if !ok {
		return
	}
	if t.Kind != "go" {
		writeError(w, http.StatusConflict, codeTaskNotCode, "task is not a code task")
		return
	}
	body, ok := a.runRequest(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	if len(t.TestCases) == 0 {
		a.runCode(w, r, body)
		return
	}
	tests := make([]struct{ Input, Expected string }, 0, len(t.TestCases))
	for _, tc := range t.TestCases {
		tests = append(tests, struct{ Input, Expected string }{tc.Input, tc.ExpectedOutput})
	}
	res, err := a.Code.RunWithTests(ctx, body.Code, tests)
	if err != nil {
		a.sandboxError(w, "run tests", err)
		return
	}
	status, err := a.recordSubmission(ctx, userFrom(ctx).ID, *t, *ref.lesson, res.AllPassed, body.Code, res.Result.Output, res.Result.Errors)
	if err != nil {
		a.internalError(w, "run tests: record", err)
		return
	}
	results := make([]apigen.RunTestResult, 0, len(res.TestResults))
	for i, tr := range res.TestResults {
		results = append(results, apigen.RunTestResult{Index: i, Passed: tr.Passed, Input: tr.Input, Actual: tr.Actual})
	}
	lessonStatus := apigen.ProgressStatus(status)
	writeJSON(w, http.StatusOK, apigen.RunResult{
		Output:       res.Result.Output,
		Errors:       res.Result.Errors,
		ExitCode:     res.Result.ExitCode,
		TimedOut:     res.Result.TimedOut,
		AllPassed:    &res.AllPassed,
		TestResults:  &results,
		LessonStatus: &lessonStatus,
	})
}

func (a *API) runPlayground(w http.ResponseWriter, r *http.Request) {
	if body, ok := a.runRequest(w, r); ok {
		a.runCode(w, r, body)
	}
}

// runRequest decodes and validates a code run request, applying the per-user rate limit.
func (a *API) runRequest(w http.ResponseWriter, r *http.Request) (apigen.RunRequest, bool) {
	var body apigen.RunRequest
	if !decodeJSON(w, r, &body) {
		return body, false
	}
	switch {
	case strings.TrimSpace(body.Code) == "":
		writeValidation(w, map[string]string{"code": fieldRequired})
		return body, false
	case len(body.Code) > 64<<10:
		writeValidation(w, map[string]string{"code": fieldTooLong})
		return body, false
	}
	if !runLimiter.Allow(strconv.Itoa(userFrom(r.Context()).ID)) {
		writeRateLimited(w, runLimiter)
		return body, false
	}
	return body, true
}

func (a *API) runCode(w http.ResponseWriter, r *http.Request, body apigen.RunRequest) {
	stdin := ""
	if body.Stdin != nil {
		stdin = *body.Stdin
	}
	res, err := a.Code.Run(r.Context(), body.Code, stdin)
	if err != nil {
		a.sandboxError(w, "run", err)
		return
	}
	writeJSON(w, http.StatusOK, apigen.RunResult{Output: res.Output, Errors: res.Errors, ExitCode: res.ExitCode, TimedOut: res.TimedOut})
}
