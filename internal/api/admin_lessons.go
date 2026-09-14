package api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/lab"
	"github.com/backendraz/golearn/internal/model"
)

func (a *API) adminCreateLesson(w http.ResponseWriter, r *http.Request) {
	course, ok := a.managedCourseParam(w, r, needEdit)
	if !ok {
		return
	}
	var body apigen.AdminLessonInput
	if !decodeJSON(w, r, &body) {
		return
	}
	l, fields := lessonFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	if deref(body.Published) && !checkPublishedChange(w, course.level, true, true, course.module.Published) {
		return
	}
	ctx := r.Context()
	var err error
	if l.OrderNum, err = a.Lessons.NextOrder(ctx, course.module.ID); err != nil {
		a.internalError(w, "admin: lesson order", err)
		return
	}
	l.ModuleID, l.Track, l.Published = course.module.ID, course.module.Track, deref(body.Published)
	id, err := a.Lessons.Create(ctx, l)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeSlugTaken, "lesson slug is taken in this course")
		return
	}
	if err != nil {
		a.internalError(w, "admin: create lesson", err)
		return
	}
	a.writeAdminLesson(w, r, http.StatusCreated, id)
}

func (a *API) adminGetLesson(w http.ResponseWriter, r *http.Request) {
	l, _, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	ctx := r.Context()
	reviews, err := a.latestReviews(r, l.ModuleID)
	if err != nil {
		a.internalError(w, "admin: lesson reviews", err)
		return
	}
	questions, err := a.quizQuestions(ctx, l.ID)
	if err != nil {
		a.internalError(w, "admin: lesson questions", err)
		return
	}
	tasks, err := a.Lessons.GetTasks(ctx, l.ID)
	if err != nil {
		a.internalError(w, "admin: lesson tasks", err)
		return
	}
	out := apigen.AdminLessonDetail{
		Lesson:    toAdminLesson(*l),
		Review:    openReview(reviews[l.ModuleID], l.ID),
		Questions: make([]apigen.AdminQuestion, 0, len(questions)),
		Tasks:     make([]apigen.AdminTask, 0, len(tasks)),
	}
	for _, q := range questions {
		out.Questions = append(out.Questions, toAdminQuestion(q, l.ID))
	}
	for _, t := range tasks {
		out.Tasks = append(out.Tasks, toAdminTask(t))
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) adminUpdateLesson(w http.ResponseWriter, r *http.Request) {
	cur, course, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	var body apigen.AdminLessonInput
	if !decodeJSON(w, r, &body) {
		return
	}
	l, fields := lessonFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	l.ID, l.ModuleID, l.Track, l.OrderNum, l.Published = cur.ID, cur.ModuleID, cur.Track, cur.OrderNum, cur.Published
	if body.Published != nil && *body.Published != cur.Published {
		if !checkPublishedChange(w, course.level, *body.Published, true, course.module.Published) {
			return
		}
		l.Published = *body.Published
	}
	err := a.Lessons.Update(r.Context(), l)
	if isUniqueViolation(err) {
		writeError(w, http.StatusConflict, codeSlugTaken, "lesson slug is taken in this course")
		return
	}
	if err != nil {
		a.internalError(w, "admin: update lesson", err)
		return
	}
	a.writeAdminLesson(w, r, http.StatusOK, l.ID)
}

func (a *API) adminDeleteLesson(w http.ResponseWriter, r *http.Request) {
	l, course, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	if l.Published && course.level < needOwner {
		writeError(w, http.StatusForbidden, codeForbidden, "only the owner can delete published lessons")
		return
	}
	if err := a.Lessons.Delete(r.Context(), l.ID); err != nil {
		a.internalError(w, "admin: delete lesson", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminSetLessonPublished(w http.ResponseWriter, r *http.Request) {
	l, course, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	var body apigen.Published
	if !decodeJSON(w, r, &body) || !checkPublishedChange(w, course.level, body.Published, true, course.module.Published) {
		return
	}
	if err := a.Lessons.SetPublished(r.Context(), l.ID, body.Published); err != nil {
		a.internalError(w, "admin: publish lesson", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminMoveLesson(w http.ResponseWriter, r *http.Request) {
	l, _, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	dir, ok := decodeMove(w, r)
	if !ok {
		return
	}
	if err := a.Lessons.MoveLesson(r.Context(), l.ID, dir); err != nil {
		a.internalError(w, "admin: move lesson", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminDuplicateLesson(w http.ResponseWriter, r *http.Request) {
	l, _, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	id, err := a.Lessons.DuplicateLesson(r.Context(), l.ID)
	if err != nil {
		a.internalError(w, "admin: duplicate lesson", err)
		return
	}
	a.writeAdminLesson(w, r, http.StatusCreated, id)
}

func (a *API) adminCreateQuestion(w http.ResponseWriter, r *http.Request) {
	l, _, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	var body apigen.AdminQuestionInput
	if !decodeJSON(w, r, &body) {
		return
	}
	q, fields := questionFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	ctx := r.Context()
	quizID, err := a.Lessons.EnsureQuiz(ctx, l.ID, "Квиз")
	if err != nil {
		a.internalError(w, "admin: ensure quiz", err)
		return
	}
	id, err := a.Lessons.AddQuestion(ctx, quizID, q)
	if err != nil {
		a.internalError(w, "admin: create question", err)
		return
	}
	a.writeAdminQuestion(w, r, http.StatusCreated, id, l.ID)
}

func (a *API) adminUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	cur, lessonID, ok := a.managedQuestion(w, r)
	if !ok {
		return
	}
	var body apigen.AdminQuestionInput
	if !decodeJSON(w, r, &body) {
		return
	}
	q, fields := questionFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	q.ID = cur.ID
	if err := a.Lessons.UpdateQuestion(r.Context(), q); err != nil {
		a.internalError(w, "admin: update question", err)
		return
	}
	a.writeAdminQuestion(w, r, http.StatusOK, q.ID, lessonID)
}

func (a *API) adminDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	q, _, ok := a.managedQuestion(w, r)
	if !ok {
		return
	}
	if err := a.Lessons.DeleteQuestion(r.Context(), q.ID); err != nil {
		a.internalError(w, "admin: delete question", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminCreateTask(w http.ResponseWriter, r *http.Request) {
	l, _, ok := a.managedLesson(w, r, needEdit)
	if !ok {
		return
	}
	var body apigen.AdminTaskInput
	if !decodeJSON(w, r, &body) {
		return
	}
	t, fields := taskFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	t.LessonID = l.ID
	id, err := a.Lessons.CreateTask(r.Context(), t)
	if err != nil {
		a.internalError(w, "admin: create task", err)
		return
	}
	a.writeAdminTask(w, r, http.StatusCreated, id)
}

func (a *API) adminUpdateTask(w http.ResponseWriter, r *http.Request) {
	cur, ok := a.managedTask(w, r)
	if !ok {
		return
	}
	var body apigen.AdminTaskInput
	if !decodeJSON(w, r, &body) {
		return
	}
	t, fields := taskFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	t.ID, t.LessonID = cur.ID, cur.LessonID
	if err := a.Lessons.UpdateTask(r.Context(), t); err != nil {
		a.internalError(w, "admin: update task", err)
		return
	}
	a.writeAdminTask(w, r, http.StatusOK, t.ID)
}

func (a *API) adminDeleteTask(w http.ResponseWriter, r *http.Request) {
	t, ok := a.managedTask(w, r)
	if !ok {
		return
	}
	if err := a.Lessons.DeleteTask(r.Context(), t.ID); err != nil {
		a.internalError(w, "admin: delete task", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// managedQuestion loads the questionId question and checks edit access to its course.
func (a *API) managedQuestion(w http.ResponseWriter, r *http.Request) (*model.QuizQuestion, int, bool) {
	id, ok := pathID(w, r, "questionId", "question not found")
	if !ok {
		return nil, 0, false
	}
	ctx := r.Context()
	q, err := a.Lessons.GetQuestionByID(ctx, id)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "question not found")
		return nil, 0, false
	}
	if err != nil {
		a.internalError(w, "admin: load question", err)
		return nil, 0, false
	}
	lessonID, err := a.Lessons.LessonIDForQuiz(ctx, q.QuizID)
	if err != nil {
		a.internalError(w, "admin: question lesson", err)
		return nil, 0, false
	}
	if _, _, ok := a.managedLessonByID(w, r, lessonID, needEdit); !ok {
		return nil, 0, false
	}
	return q, lessonID, true
}

// managedTask loads the taskId task and checks edit access to its course.
func (a *API) managedTask(w http.ResponseWriter, r *http.Request) (*model.Task, bool) {
	id, ok := pathID(w, r, "taskId", "task not found")
	if !ok {
		return nil, false
	}
	t, err := a.Lessons.GetTaskByID(r.Context(), id)
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "task not found")
		return nil, false
	}
	if err != nil {
		a.internalError(w, "admin: load task", err)
		return nil, false
	}
	if _, _, ok := a.managedLessonByID(w, r, t.LessonID, needEdit); !ok {
		return nil, false
	}
	return t, true
}

func (a *API) writeAdminLesson(w http.ResponseWriter, r *http.Request, status, id int) {
	l, err := a.Lessons.GetByID(r.Context(), id)
	if err != nil {
		a.internalError(w, "admin: reload lesson", err)
		return
	}
	writeJSON(w, status, toAdminLesson(*l))
}

func (a *API) writeAdminQuestion(w http.ResponseWriter, r *http.Request, status, id, lessonID int) {
	q, err := a.Lessons.GetQuestionByID(r.Context(), id)
	if err != nil {
		a.internalError(w, "admin: reload question", err)
		return
	}
	writeJSON(w, status, toAdminQuestion(*q, lessonID))
}

func (a *API) writeAdminTask(w http.ResponseWriter, r *http.Request, status, id int) {
	t, err := a.Lessons.GetTaskByID(r.Context(), id)
	if err != nil {
		a.internalError(w, "admin: reload task", err)
		return
	}
	writeJSON(w, status, toAdminTask(*t))
}

func lessonFromInput(in apigen.AdminLessonInput) (model.Lesson, map[string]string) {
	fields := map[string]string{}
	l := model.Lesson{
		Slug:       strings.TrimSpace(in.Slug),
		Title:      strings.TrimSpace(in.Title),
		Kind:       string(in.Kind),
		Format:     string(in.Format),
		Difficulty: string(in.Difficulty),
		Content:    deref(in.Content),
		VMImage:    strings.TrimSpace(deref(in.VMImage)),
		VMInit:     deref(in.VMInit),
	}
	checkSlug(fields, "slug", l.Slug)
	checkText(fields, "title", l.Title, 200, true)
	checkText(fields, "content", l.Content, 500000, false)
	checkText(fields, "vm_image", l.VMImage, 40, false)
	checkText(fields, "vm_init", l.VMInit, 64000, false)
	checkEnum(fields, "kind", in.Kind.Valid())
	checkEnum(fields, "format", in.Format.Valid())
	checkEnum(fields, "difficulty", in.Difficulty.Valid())
	return l, fields
}

func questionFromInput(in apigen.AdminQuestionInput) (model.QuizQuestion, map[string]string) {
	fields := map[string]string{}
	q := model.QuizQuestion{
		Question:     strings.TrimSpace(in.Question),
		Options:      in.Options,
		OptionExpl:   deref(in.OptionExplanations),
		CorrectIndex: in.CorrectIndex,
		Explanation:  deref(in.Explanation),
	}
	checkText(fields, "question", q.Question, 5000, true)
	checkText(fields, "explanation", q.Explanation, 5000, false)
	switch {
	case len(q.Options) < 2:
		fields["options"] = fieldTooShort
	case len(q.Options) > 10:
		fields["options"] = fieldTooLong
	}
	for i, o := range q.Options {
		q.Options[i] = strings.TrimSpace(o)
		checkText(fields, "options", q.Options[i], 2000, true)
	}
	if len(q.OptionExpl) != 0 && len(q.OptionExpl) != len(q.Options) {
		fields["option_explanations"] = fieldInvalidValue
	}
	for _, e := range q.OptionExpl {
		checkText(fields, "option_explanations", e, 5000, false)
	}
	if q.CorrectIndex < 0 || q.CorrectIndex >= len(q.Options) {
		fields["correct_index"] = fieldInvalidValue
	}
	return q, fields
}

func taskFromInput(in apigen.AdminTaskInput) (model.Task, map[string]string) {
	fields := map[string]string{}
	t := model.Task{
		Title:        strings.TrimSpace(in.Title),
		Kind:         string(in.Kind),
		Format:       string(in.Format),
		Difficulty:   string(in.Difficulty),
		Description:  deref(in.Description),
		Hints:        deref(in.Hints),
		Solution:     deref(in.Solution),
		StarterCode:  deref(in.StarterCode),
		SandboxImage: strings.TrimSpace(deref(in.SandboxImage)),
		SetupScript:  deref(in.SetupScript),
		CheckScript:  deref(in.CheckScript),
		Glossary:     []model.GlossaryItem{},
		TestCases:    []model.TestCase{},
	}
	checkText(fields, "title", t.Title, 200, true)
	checkText(fields, "description", t.Description, 100000, false)
	checkText(fields, "hints", t.Hints, 20000, false)
	checkText(fields, "solution", t.Solution, 100000, false)
	checkText(fields, "starter_code", t.StarterCode, 64000, false)
	checkText(fields, "setup_script", t.SetupScript, 64000, false)
	checkText(fields, "check_script", t.CheckScript, 64000, false)
	checkEnum(fields, "kind", in.Kind.Valid())
	checkEnum(fields, "format", in.Format.Valid())
	checkEnum(fields, "difficulty", in.Difficulty.Valid())
	if t.SandboxImage != "" && !slices.Contains(lab.SandboxImages, t.SandboxImage) {
		fields["sandbox_image"] = fieldInvalidValue
	}
	for _, g := range deref(in.Glossary) {
		t.Glossary = append(t.Glossary, model.GlossaryItem{Term: g.Term, Definition: g.Definition})
	}
	for _, tc := range deref(in.TestCases) {
		t.TestCases = append(t.TestCases, model.TestCase{Input: tc.Input, ExpectedOutput: tc.ExpectedOutput})
	}
	if len(t.Glossary) > 100 {
		fields["glossary"] = fieldTooLong
	}
	if len(t.TestCases) > 100 {
		fields["test_cases"] = fieldTooLong
	}
	return t, fields
}

func checkEnum(fields map[string]string, name string, valid bool) {
	if !valid {
		fields[name] = fieldInvalidValue
	}
}

func toAdminLesson(l model.Lesson) apigen.AdminLesson {
	return apigen.AdminLesson{
		ID: l.ID, CourseID: l.ModuleID, Slug: l.Slug, Title: l.Title, Kind: lessonKind(l.Kind),
		Format: apigen.ContentFormat(l.Format), Difficulty: apigen.Difficulty(l.Difficulty),
		Content: l.Content, VMImage: l.VMImage, VMInit: l.VMInit, Published: l.Published,
		OrderNum: l.OrderNum, Source: apigen.AdminLessonSource(l.Source), CreatedAt: l.CreatedAt,
	}
}

func toAdminQuestion(q model.QuizQuestion, lessonID int) apigen.AdminQuestion {
	return apigen.AdminQuestion{
		ID: q.ID, LessonID: lessonID, OrderNum: q.OrderNum, Question: q.Question,
		Options: nonNil(q.Options), OptionExplanations: nonNil(q.OptionExpl),
		CorrectIndex: q.CorrectIndex, Explanation: q.Explanation,
	}
}

func toAdminTask(t model.Task) apigen.AdminTask {
	out := apigen.AdminTask{
		ID: t.ID, LessonID: t.LessonID, OrderNum: t.OrderNum, Title: t.Title,
		Kind: apigen.TaskKind(t.Kind), Format: apigen.ContentFormat(t.Format),
		Difficulty: apigen.TaskDifficulty(t.Difficulty), Description: t.Description, Hints: t.Hints,
		Solution: t.Solution, StarterCode: t.StarterCode, SandboxImage: t.SandboxImage,
		SetupScript: t.SetupScript, CheckScript: t.CheckScript,
		Glossary: make([]apigen.GlossaryItem, 0, len(t.Glossary)), TestCases: make([]apigen.TestCase, 0, len(t.TestCases)),
	}
	for _, g := range t.Glossary {
		out.Glossary = append(out.Glossary, apigen.GlossaryItem{Term: g.Term, Definition: g.Definition})
	}
	for _, tc := range t.TestCases {
		out.TestCases = append(out.TestCases, apigen.TestCase{Input: tc.Input, ExpectedOutput: tc.ExpectedOutput})
	}
	return out
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
