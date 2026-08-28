package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type TasksPageData struct {
	PageTitle     string
	Module        *model.Module
	Lesson        *model.Lesson
	ContentHTML   string
	Tasks         []model.Task
	IsShellLab    bool
	IsGitLab      bool         // a Git-course lab: show the live commit-graph panel
	PassedIDs     map[int]bool // steps already solved by this user
	DoneCount     int
	HasQuiz       bool
	QuestionsJSON template.JS
	PassedJSON    template.JS
}

func (h *Handler) TasksPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")
	lessonSlug := chi.URLParam(r, "lessonSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lesson, err := h.lessonRepo.GetBySlug(ctx, mod.ID, lessonSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	tasks, err := h.lessonRepo.GetTasks(ctx, lesson.ID)
	if err != nil {
		h.log.Error("get tasks", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	data := TasksPageData{
		PageTitle:   "Лаборатория: " + lesson.Title,
		Module:      mod,
		Lesson:      lesson,
		ContentHTML: lesson.Content,
		Tasks:       tasks,
	}
	if len(tasks) > 0 && tasks[0].Kind == "shell" {
		data.IsShellLab = true
	}
	// Git-course labs get a live commit-graph beside the terminal.
	data.IsGitLab = strings.EqualFold(mod.Category, "Git") ||
		strings.Contains(strings.ToLower(mod.Slug), "git")

	// Solved steps are restored from the database, so reloading the page (or
	// coming back a week later) keeps the lab's completed steps marked.
	passed, err := h.submissionRepo.PassedTaskIDs(ctx, currentUserID(ctx), lesson.ID)
	if err != nil {
		h.log.Error("passed task ids", "error", err)
		passed = map[int]bool{}
	}
	data.PassedIDs = passed
	data.DoneCount = len(passed)
	passedList := make([]int, 0, len(passed))
	for _, t := range tasks {
		if passed[t.ID] {
			passedList = append(passedList, t.ID)
		}
	}
	pjson, _ := json.Marshal(passedList)
	data.PassedJSON = template.JS(pjson)

	// Inline quiz (lab chapters from the import carry quiz questions too).
	if _, questions, qerr := h.lessonRepo.GetQuiz(ctx, lesson.ID); qerr == nil && len(questions) > 0 {
		type qp struct {
			Q  string   `json:"q"`
			O  []string `json:"o"`
			OE []string `json:"oe"`
			C  int      `json:"c"`
			E  string   `json:"e"`
		}
		payload := make([]qp, len(questions))
		for i, q := range questions {
			payload[i] = qp{Q: q.Question, O: q.Options, OE: q.OptionExpl, C: q.CorrectIndex, E: q.Explanation}
		}
		qjson, _ := json.Marshal(payload)
		data.HasQuiz = true
		data.QuestionsJSON = template.JS(qjson)
	}

	h.render(w, "tasks", &data)
}
