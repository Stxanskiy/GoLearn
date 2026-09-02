package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

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
	specRepo       *repository.SpecRepo
	runner         *runner.Runner
	shell          *runner.VMRunner
	log            *slog.Logger
}

func New(mr *repository.ModuleRepo, lr *repository.LessonRepo, pr *repository.ProgressRepo, sr *repository.SubmissionRepo, ur *repository.UserRepo, spr *repository.SpecRepo, log *slog.Logger) *Handler {
	return &Handler{
		moduleRepo:     mr,
		lessonRepo:     lr,
		progressRepo:   pr,
		submissionRepo: sr,
		userRepo:       ur,
		specRepo:       spr,
		runner:         runner.New(),
		shell:          runner.NewVMRunner(),
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
	// "/" is public: visitors get the landing page, signed-in users the dashboard.
	r.Get("/", h.Home)

	// Protected routes (require auth)
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/profile", h.ProfilePage)
		r.Get("/courses", h.CoursesPage)
		r.Get("/courses/{track}", h.SectionPage)
		r.Get("/api/courses/{slug}/cover", h.CourseCover)
		r.Get("/api/spec/{slug}/cover", h.SpecCover)
		r.Get("/trainers", h.TrainersPage)
		r.Get("/simulators", h.SimulatorsPage)
		r.Get("/simulator/{slug}", h.SimulatorPage)
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
		r.Post("/api/shell/{taskID}/done", h.ShellStepDone)
		r.Post("/api/lab/{lessonID}/reset", h.LabReset)
		r.Post("/api/lab/{lessonID}/retry", h.LabRetry)
		r.Get("/api/lab/{lessonID}/preview/{port}/*", h.LabPreview)
		r.Get("/api/lab/{lessonID}/fs/list", h.LabFSList)
		r.Get("/api/lab/{lessonID}/fs/read", h.LabFSRead)
		r.Put("/api/lab/{lessonID}/fs/write", h.LabFSWrite)
		r.Post("/api/git/exec", h.GitExec)
		r.Post("/api/git/reset", h.GitReset)
		r.Get("/api/term", h.TermWS)

		// Admin (role=admin only)
		r.Group(func(r chi.Router) {
			r.Use(h.AdminMiddleware)
			r.Get("/admin", h.AdminDashboard)
			r.Get("/admin/users", h.AdminUsers)
			r.Post("/admin/users", h.AdminUserCreate)
			r.Post("/admin/preview", h.AdminPreview)
			r.Get("/admin/specs", h.AdminSpecs)
			r.Post("/admin/spec", h.AdminSpecSave)
			r.Post("/admin/spec/{slug}/delete", h.AdminSpecDelete)
			r.Get("/admin/module/new", h.AdminModuleNew)
			r.Post("/admin/module", h.AdminModuleSave)
			r.Get("/admin/module/{id}", h.AdminModuleEdit)
			r.Post("/admin/module/{id}", h.AdminModuleSave)
			r.Post("/admin/module/{id}/delete", h.AdminModuleDelete)
			r.Get("/admin/module/{id}/lesson/new", h.AdminLessonNew)
			r.Post("/admin/lesson", h.AdminLessonSave)
			r.Get("/admin/lesson/{id}", h.AdminLessonEdit)
			r.Post("/admin/lesson/{id}", h.AdminLessonSave)
			r.Post("/admin/lesson/{id}/delete", h.AdminLessonDelete)
			// Quiz questions
			r.Get("/admin/lesson/{id}/question/new", h.AdminQuestionNew)
			r.Post("/admin/question", h.AdminQuestionSave)
			r.Get("/admin/question/{id}", h.AdminQuestionEdit)
			r.Post("/admin/question/{id}", h.AdminQuestionSave)
			r.Post("/admin/question/{id}/delete", h.AdminQuestionDelete)
			// Tasks
			r.Get("/admin/lesson/{id}/task/new", h.AdminTaskNew)
			r.Post("/admin/task", h.AdminTaskSave)
			r.Get("/admin/task/{id}", h.AdminTaskEdit)
			r.Post("/admin/task/{id}", h.AdminTaskSave)
			r.Post("/admin/task/{id}/delete", h.AdminTaskDelete)
		})
	})
}

// templateFuncs is the single FuncMap used by the renderer AND by the template
// parse test, so a helper added here can never be missing at render time.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
		"pct": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a * 100 / b
		},
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"content":  func(format, raw string) template.HTML { return RenderContent(format, raw) },
		"assetVer": func() string { return assetVersion },
	}
}

// assetVersion busts the browser cache for /static assets after a redesign.
// The static files are served without an ETag, so a stale app.css can survive a
// deploy; appending ?v=<version> to the link forces a refetch when it changes.
// Derived once from the app.css modtime, with a build-time fallback.
var assetVersion = func() string {
	if fi, err := os.Stat("internal/static/css/app.css"); err == nil {
		return strconv.FormatInt(fi.ModTime().Unix(), 36)
	}
	return strconv.FormatInt(time.Now().Unix(), 36)
}()

func (h *Handler) render(w http.ResponseWriter, tmplName string, data any) {
	w.Header().Set("Cache-Control", "no-store")
	tmpl, err := template.New("").Funcs(templateFuncs()).ParseFiles(
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
