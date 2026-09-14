// Package api serves the JSON API (/api/v1) for the Next.js frontend.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/backendraz/golearn/internal/runner"
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
	GetByID(ctx context.Context, id int) (*repository.User, error)
}

type moduleStore interface {
	GetAll(ctx context.Context) ([]model.Module, error)
	GetBySlug(ctx context.Context, slug string) (*model.Module, error)
	GetByID(ctx context.Context, id int) (*model.Module, error)
	Neighbors(ctx context.Context, m model.Module, tracks []string) (prev, next *model.Module, err error)
	Stats(ctx context.Context) (repository.PlatformStats, error)
	ListManaged(ctx context.Context, userID int, all bool) ([]repository.CourseRow, error)
	NextOrder(ctx context.Context) (int, error)
	Create(ctx context.Context, m model.Module) (int, error)
	Update(ctx context.Context, m model.Module) error
	Delete(ctx context.Context, id int) error
	SetPublished(ctx context.Context, id int, published bool) error
	Move(ctx context.Context, id int, dir string) error
	SetCover(ctx context.Context, id int, cover string) error
}

type lessonStore interface {
	GetByModule(ctx context.Context, moduleID int) ([]model.Lesson, error)
	ListPublishedOutline(ctx context.Context) ([]model.Lesson, error)
	GetBySlug(ctx context.Context, moduleID int, slug string) (*model.Lesson, error)
	GetByID(ctx context.Context, id int) (*model.Lesson, error)
	GetQuiz(ctx context.Context, lessonID int) (*model.Quiz, []model.QuizQuestion, error)
	CountsForLesson(ctx context.Context, lessonID int) (questions, tasks int)
	GetTasks(ctx context.Context, lessonID int) ([]model.Task, error)
	GetTaskByID(ctx context.Context, taskID int) (*model.Task, error)
	LessonSandbox(ctx context.Context, lessonID int) (image, setup string, err error)
	GetByModuleAll(ctx context.Context, moduleID int) ([]model.Lesson, error)
	NextOrder(ctx context.Context, moduleID int) (int, error)
	Create(ctx context.Context, l model.Lesson) (int, error)
	Update(ctx context.Context, l model.Lesson) error
	Delete(ctx context.Context, id int) error
	SetPublished(ctx context.Context, id int, published bool) error
	MoveLesson(ctx context.Context, id int, dir string) error
	DuplicateLesson(ctx context.Context, id int) (int, error)
	EnsureQuiz(ctx context.Context, lessonID int, title string) (int, error)
	AddQuestion(ctx context.Context, quizID int, q model.QuizQuestion) (int, error)
	UpdateQuestion(ctx context.Context, q model.QuizQuestion) error
	DeleteQuestion(ctx context.Context, id int) error
	GetQuestionByID(ctx context.Context, id int) (*model.QuizQuestion, error)
	LessonIDForQuiz(ctx context.Context, quizID int) (int, error)
	CreateTask(ctx context.Context, t model.Task) (int, error)
	UpdateTask(ctx context.Context, t model.Task) error
	DeleteTask(ctx context.Context, id int) error
}

// authorStore keeps course co-authors.
type authorStore interface {
	IsCoauthor(ctx context.Context, moduleID, userID int) (bool, error)
	List(ctx context.Context, moduleID int) ([]repository.User, error)
	Add(ctx context.Context, moduleID, userID, addedBy int) (bool, error)
	Remove(ctx context.Context, moduleID, userID int) (bool, error)
	SetOwner(ctx context.Context, moduleID, userID, byID int) error
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
	ResetLesson(ctx context.Context, userID, lessonID int) error
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
	PassedTaskIDs(ctx context.Context, userID, lessonID int) (map[int]bool, error)
	Save(ctx context.Context, userID, taskID int, code, output, errors string, passed bool) error
	ResetLesson(ctx context.Context, userID, lessonID int) error
}

// sandbox is the per-user lab VM runner.
type sandbox interface {
	Enabled() bool
	Session(userID int, key string) (runner.SessionInfo, bool)
	Touch(userID int, key string)
	EnsureSession(ctx context.Context, userID int, key, image, setup string) (string, error)
	OpenPTY(handle string, cols, rows int) (*runner.PTYSession, error)
	Exec(ctx context.Context, userID int, key, image, setup, command string) (string, error)
	Check(ctx context.Context, userID int, key, image, setup, checkScript string) (bool, string, error)
	Preview(ctx context.Context, userID int, key, image, setup string, port int, path string) ([]byte, string, int, error)
	FSList(ctx context.Context, userID int, key, image, setup, dir string) ([]runner.FSEntry, error)
	FSRead(ctx context.Context, userID int, key, image, setup, file string) ([]byte, error)
	FSWrite(ctx context.Context, userID int, key, image, setup, file string, content []byte) error
	Reset(ctx context.Context, userID int, key string) error
}

// codeRunner runs Go playground and code-task programs.
type codeRunner interface {
	Run(ctx context.Context, code, stdin string) (*runner.Result, error)
	RunWithTests(ctx context.Context, code string, tests []struct{ Input, Expected string }) (*runner.RunResult, error)
}

type specStore interface {
	ListPublished(ctx context.Context) ([]model.Specialization, error)
	Get(ctx context.Context, slug string) (*model.Specialization, error)
}

type simStore interface {
	ListPublished(ctx context.Context) ([]model.Simulator, error)
	Get(ctx context.Context, slug string) (*model.Simulator, error)
}

// Stores are the repositories and runners the API uses.
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
	Authors      authorStore
	Sandbox      sandbox
	Code         codeRunner
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

		r.Get("/lessons/{lessonId}/lab", a.getLab)
		r.Get("/lessons/{lessonId}/lab/session", a.getLabSession)
		r.Delete("/lessons/{lessonId}/lab/session", a.stopLabSession)
		r.Post("/lessons/{lessonId}/lab/retry", a.retryLab)
		r.Get("/lessons/{lessonId}/lab/terminal", a.openLabTerminal)
		r.Get("/lessons/{lessonId}/lab/git-graph", a.getLabGitGraph)
		r.Get("/lessons/{lessonId}/lab/fs/entries", a.listLabFiles)
		r.Get("/lessons/{lessonId}/lab/fs/content", a.readLabFile)
		r.Put("/lessons/{lessonId}/lab/fs/content", a.writeLabFile)
		r.Get("/lessons/{lessonId}/lab/preview/{port}/*", a.previewLab)
		r.Post("/tasks/{taskId}/check", a.checkTask)
		r.Post("/tasks/{taskId}/done", a.markTaskDone)
		r.Post("/tasks/{taskId}/run", a.runTaskCode)

		r.Post("/playground/run", a.runPlayground)
		r.Get("/git-trainer/session", a.getGitTrainerSession)
		r.Delete("/git-trainer/session", a.resetGitTrainer)
		r.Get("/git-trainer/terminal", a.openGitTrainerTerminal)
		r.Get("/git-trainer/git-graph", a.getGitTrainerGraph)

		r.Group(func(r chi.Router) {
			r.Use(requireAuthor)
			r.Post("/admin/content/preview", a.adminPreviewContent)
			r.Get("/admin/courses", a.adminListCourses)
			r.Post("/admin/courses", a.adminCreateCourse)
			r.Get("/admin/courses/{courseId}", a.adminGetCourse)
			r.Put("/admin/courses/{courseId}", a.adminUpdateCourse)
			r.Delete("/admin/courses/{courseId}", a.adminDeleteCourse)
			r.Put("/admin/courses/{courseId}/published", a.adminSetCoursePublished)
			r.With(requireAdmin).Post("/admin/courses/{courseId}/move", a.adminMoveCourse)
			r.Put("/admin/courses/{courseId}/cover", a.adminUploadCourseCover)
			r.Delete("/admin/courses/{courseId}/cover", a.adminDeleteCourseCover)
			r.Get("/admin/courses/{courseId}/authors", a.adminListCourseAuthors)
			r.Post("/admin/courses/{courseId}/authors", a.adminAddCourseAuthor)
			r.Delete("/admin/courses/{courseId}/authors/{userId}", a.adminRemoveCourseAuthor)
			r.With(requireAdmin).Put("/admin/courses/{courseId}/owner", a.adminSetCourseOwner)

			r.Post("/admin/courses/{courseId}/lessons", a.adminCreateLesson)
			r.Get("/admin/lessons/{lessonId}", a.adminGetLesson)
			r.Put("/admin/lessons/{lessonId}", a.adminUpdateLesson)
			r.Delete("/admin/lessons/{lessonId}", a.adminDeleteLesson)
			r.Put("/admin/lessons/{lessonId}/published", a.adminSetLessonPublished)
			r.Post("/admin/lessons/{lessonId}/move", a.adminMoveLesson)
			r.Post("/admin/lessons/{lessonId}/duplicate", a.adminDuplicateLesson)
			r.Post("/admin/lessons/{lessonId}/questions", a.adminCreateQuestion)
			r.Put("/admin/questions/{questionId}", a.adminUpdateQuestion)
			r.Delete("/admin/questions/{questionId}", a.adminDeleteQuestion)
			r.Post("/admin/lessons/{lessonId}/tasks", a.adminCreateTask)
			r.Put("/admin/tasks/{taskId}", a.adminUpdateTask)
			r.Delete("/admin/tasks/{taskId}", a.adminDeleteTask)
		})
	})
	return r
}
