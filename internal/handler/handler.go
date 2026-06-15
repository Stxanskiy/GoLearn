package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/backendraz/golearn/internal/repository"
	"github.com/backendraz/golearn/internal/runner"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	moduleRepo     *repository.ModuleRepo
	lessonRepo     *repository.LessonRepo
	progressRepo   *repository.ProgressRepo
	submissionRepo *repository.SubmissionRepo
	userRepo       *repository.UserRepo
	runner         *runner.Runner
	shell          *runner.ShellRunner
	log            *slog.Logger
}

func New(mr *repository.ModuleRepo, lr *repository.LessonRepo, pr *repository.ProgressRepo, sr *repository.SubmissionRepo, ur *repository.UserRepo, log *slog.Logger) *Handler {
	return &Handler{
		moduleRepo:     mr,
		lessonRepo:     lr,
		progressRepo:   pr,
		submissionRepo: sr,
		userRepo:       ur,
		runner:         runner.New(),
		shell:          runner.NewShellRunner(),
		log:            log,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Public routes (no auth)
	r.Get("/login", h.LoginPage)
	r.Post("/login", h.Login)
	r.Get("/register", h.RegisterPage)
	r.Post("/register", h.Register)
	r.Get("/logout", h.Logout)

	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/", h.Dashboard)
		r.Get("/courses", h.CoursesPage)
		r.Get("/roadmap", h.RoadmapPage)
		r.Get("/module/{moduleSlug}", h.ModulePage)
		r.Get("/module/{moduleSlug}/lesson/{lessonSlug}", h.LessonPage)
		r.Get("/module/{moduleSlug}/lesson/{lessonSlug}/quiz", h.QuizPage)
		r.Post("/module/{moduleSlug}/lesson/{lessonSlug}/quiz", h.SubmitQuiz)
		r.Get("/module/{moduleSlug}/lesson/{lessonSlug}/tasks", h.TasksPage)
		r.Post("/api/progress/{lessonID}/status", h.UpdateProgress)
		r.Post("/api/progress/{lessonID}/notes", h.SaveNotes)
		r.Get("/git-trainer", h.GitTrainerPage)
		r.Get("/playground", h.PlaygroundPage)
		r.Post("/api/run", h.RunCode)
		r.Post("/api/run/{taskID}", h.RunTaskCode)
		r.Post("/api/shell/{taskID}/exec", h.ShellExec)
		r.Post("/api/shell/{taskID}/check", h.ShellCheck)
		r.Post("/api/shell/{taskID}/reset", h.ShellReset)
		r.Post("/api/git/exec", h.GitExec)
		r.Post("/api/git/reset", h.GitReset)
	})
}

func (h *Handler) render(w http.ResponseWriter, tmplName string, data any) {
	w.Header().Set("Cache-Control", "no-store")
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
	}).ParseFiles(
		"internal/templates/layouts/base.html",
		"internal/templates/pages/"+tmplName+".html",
	)
	if err != nil {
		h.log.Error("parse templates", "template", tmplName, "error", err)
		http.Error(w, "Template error", 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, tmplName, data); err != nil {
		h.log.Error("execute template", "template", tmplName, "error", err)
		http.Error(w, "Render error", 500)
	}
}
