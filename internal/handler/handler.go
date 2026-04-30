package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/backendraz/golearn/internal/repository"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	moduleRepo   *repository.ModuleRepo
	lessonRepo   *repository.LessonRepo
	progressRepo *repository.ProgressRepo
	log          *slog.Logger
}

func New(mr *repository.ModuleRepo, lr *repository.LessonRepo, pr *repository.ProgressRepo, log *slog.Logger) *Handler {
	// We'll parse templates from disk for hot-reload in dev
	return &Handler{
		moduleRepo:   mr,
		lessonRepo:   lr,
		progressRepo: pr,
		log:          log,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.Dashboard)
	r.Get("/module/{moduleSlug}", h.ModulePage)
	r.Get("/module/{moduleSlug}/lesson/{lessonSlug}", h.LessonPage)
	r.Get("/module/{moduleSlug}/lesson/{lessonSlug}/quiz", h.QuizPage)
	r.Post("/module/{moduleSlug}/lesson/{lessonSlug}/quiz", h.SubmitQuiz)
	r.Get("/module/{moduleSlug}/lesson/{lessonSlug}/tasks", h.TasksPage)
	r.Post("/api/progress/{lessonID}/status", h.UpdateProgress)
	r.Post("/api/progress/{lessonID}/notes", h.SaveNotes)
}

func (h *Handler) render(w http.ResponseWriter, tmplName string, data any) {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"pct": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a * 100 / b
		},
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).ParseGlob("internal/templates/layouts/*.html")
	if err != nil {
		h.log.Error("parse layout templates", "error", err)
		http.Error(w, "Template error", 500)
		return
	}

	tmpl, err = tmpl.ParseGlob("internal/templates/pages/*.html")
	if err != nil {
		h.log.Error("parse page templates", "error", err)
		http.Error(w, "Template error", 500)
		return
	}

	tmpl, err = tmpl.ParseGlob("internal/templates/components/*.html")
	if err != nil {
		h.log.Error("parse component templates", "error", err)
		http.Error(w, "Template error", 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		h.log.Error("execute template", "template", tmplName, "error", err)
		http.Error(w, "Render error", 500)
	}
}
