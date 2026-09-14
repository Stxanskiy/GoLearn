// Package api serves the JSON API (/api/v1) for the Next.js frontend.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/go-chi/chi/v5"
)

type userStore interface {
	GetByEmail(ctx context.Context, email string) (*repository.User, error)
	CheckPassword(user *repository.User, password string) bool
	Create(ctx context.Context, email, password, name string) (*repository.User, error)
	CreateSession(ctx context.Context, userID int) (string, error)
	GetUserBySession(ctx context.Context, token string) (*repository.User, error)
	DeleteSession(ctx context.Context, token string) error
	SetRole(ctx context.Context, userID int, role string) error
}

type moduleStore interface {
	GetAll(ctx context.Context) ([]model.Module, error)
	GetBySlug(ctx context.Context, slug string) (*model.Module, error)
	GetByID(ctx context.Context, id int) (*model.Module, error)
	Neighbors(ctx context.Context, m model.Module, tracks []string) (prev, next *model.Module, err error)
	Stats(ctx context.Context) (repository.PlatformStats, error)
}

type lessonStore interface {
	GetByModule(ctx context.Context, moduleID int) ([]model.Lesson, error)
	ListPublishedOutline(ctx context.Context) ([]model.Lesson, error)
	GetBySlug(ctx context.Context, moduleID int, slug string) (*model.Lesson, error)
	GetByID(ctx context.Context, id int) (*model.Lesson, error)
	GetQuiz(ctx context.Context, lessonID int) (*model.Quiz, []model.QuizQuestion, error)
	CountsForLesson(ctx context.Context, lessonID int) (questions, tasks int)
}

type progressStore interface {
	GetAll(ctx context.Context, userID int) ([]model.Progress, error)
	Overview(ctx context.Context, userID int) (*model.ProgressOverview, error)
	LatestInProgress(ctx context.Context, userID int) (*repository.ContinueLesson, error)
	Get(ctx context.Context, userID, lessonID int) (*model.Progress, error)
	Start(ctx context.Context, userID, lessonID int) error
	Upsert(ctx context.Context, userID, lessonID int, status string) error
	SaveNotes(ctx context.Context, userID, lessonID int, notes string) error
	SaveQuizResult(ctx context.Context, userID, lessonID, score, total int) error
}

type quizAttemptStore interface {
	Save(ctx context.Context, userID, lessonID, score, total int, answers []repository.AttemptAnswer) (int, error)
	Count(ctx context.Context, userID, lessonID int) (int, error)
}

type quizAnswerStore interface {
	Record(ctx context.Context, userID, questionID, selected int) (stored int, created bool, err error)
	ForLesson(ctx context.Context, userID, lessonID int) (map[int]int, error)
	ResetLesson(ctx context.Context, userID, lessonID int) error
}

type submissionStore interface {
	LessonLabStatus(ctx context.Context, userID int) (map[int]bool, error)
}

type specStore interface {
	ListPublished(ctx context.Context) ([]model.Specialization, error)
	Get(ctx context.Context, slug string) (*model.Specialization, error)
}

type simStore interface {
	ListPublished(ctx context.Context) ([]model.Simulator, error)
	Get(ctx context.Context, slug string) (*model.Simulator, error)
}

// Stores are the repositories the API reads and writes.
type Stores struct {
	Users        userStore
	Modules      moduleStore
	Lessons      lessonStore
	Progress     progressStore
	Submissions  submissionStore
	Specs        specStore
	Sims         simStore
	QuizAnswers  quizAnswerStore
	QuizAttempts quizAttemptStore
}

// Config holds API settings.
type Config struct {
	AllowedOrigins []string // extra origins accepted for mutating requests, e.g. http://localhost:3000
}

// API wires handlers to their dependencies.
type API struct {
	Stores
	cfg Config
	log *slog.Logger
}

func New(stores Stores, cfg Config, log *slog.Logger) *API {
	return &API{Stores: stores, cfg: cfg, log: log}
}

// Routes returns the router to mount at /api/v1.
func (a *API) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(noStore, a.csrfGuard, a.sessionUser)
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, codeNotFound, "route not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "method not allowed")
	})

	r.Get("/auth/config", a.getAuthConfig)
	r.Post("/auth/login", a.login)
	r.Post("/auth/register", a.register)
	r.Post("/auth/logout", a.logout)

	r.Get("/public/landing", a.getLanding)
	r.Get("/courses/{courseSlug}/cover", a.getCourseCover)
	r.Get("/specializations/{specSlug}/cover", a.getSpecializationCover)

	r.Group(func(r chi.Router) {
		r.Use(requireUser)
		r.Get("/me", a.getMe)
		r.Get("/me/profile", a.getProfile)
		r.Get("/me/dashboard", a.getDashboard)
		r.Get("/catalog", a.getCatalog)
		r.Get("/specializations/{specSlug}", a.getSpecialization)
		r.Get("/courses/{courseSlug}", a.getCourse)
		r.Get("/simulators", a.listSimulators)
		r.Get("/simulators/{simSlug}", a.getSimulator)

		r.Get("/courses/{courseSlug}/lessons/{lessonSlug}", a.getLesson)
		r.Post("/lessons/{lessonId}/visit", a.visitLesson)
		r.Post("/lessons/{lessonId}/complete", a.completeLesson)
		r.Put("/lessons/{lessonId}/notes", a.saveLessonNotes)
		r.Post("/lessons/{lessonId}/quiz/answers", a.answerQuizQuestion)
		r.Delete("/lessons/{lessonId}/quiz/answers", a.resetQuizAttempt)
		r.Post("/lessons/{lessonId}/quiz/submit", a.submitQuiz)
	})
	return r
}
