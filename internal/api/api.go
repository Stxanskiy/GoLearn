// Package api serves the JSON API (/api/v1) for the Next.js frontend.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/backendraz/golearn/internal/runner"
	"github.com/backendraz/golearn/internal/storage"
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
	ListPage(ctx context.Context, q string, beforeID, limit int) ([]repository.User, error)
	CreateWithRole(ctx context.Context, email, password, name, role string) (*repository.User, error)
	SetAccess(ctx context.Context, userID int, role string, blocked bool) error
	SetPassword(ctx context.Context, userID int, password string) error
	DeleteSessions(ctx context.Context, userID int) error
	DeleteGuarded(ctx context.Context, userID int) error
}

type moduleStore interface {
	GetAll(ctx context.Context) ([]model.Module, error)
	GetBySlug(ctx context.Context, slug string) (*model.Module, error)
	GetByID(ctx context.Context, id int) (*model.Module, error)
	Neighbors(ctx context.Context, m model.Module, tracks []string) (prev, next *model.Module, err error)
	Stats(ctx context.Context) (repository.PlatformStats, error)
	ListManaged(ctx context.Context, userID int, all bool) ([]repository.CourseRow, error)
	TrackCounts(ctx context.Context) (map[string]int, error)
	TrackUsage(ctx context.Context) (map[string]int, error)
	DraftFor(ctx context.Context, liveID int) (*model.Module, error)
	Drafts(ctx context.Context) (map[int]int, error)
	NextOrder(ctx context.Context) (int, error)
	Create(ctx context.Context, m model.Module) (int, error)
	Update(ctx context.Context, m model.Module) error
	Delete(ctx context.Context, id int) error
	SetPublished(ctx context.Context, id int, published bool) error
	Move(ctx context.Context, id int, dir string) error
	SetCover(ctx context.Context, id int, cover string) error
	SetIcon(ctx context.Context, id int, iconURL string) error
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
	DraftCopyOf(ctx context.Context, draftModuleID, originID int) (int, error)
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

// courseIOStore exports and imports whole courses.
type courseIOStore interface {
	Export(ctx context.Context, moduleID int) (model.CourseTree, error)
	Diff(ctx context.Context, tree model.CourseTree, moduleID int) (repository.CourseDiff, error)
	Upsert(ctx context.Context, tree model.CourseTree, moduleID int) (repository.CourseDiff, error)
}

// draftStore manages drafts of published courses.
type draftStore interface {
	Create(ctx context.Context, liveID int) (int, error)
	Discard(ctx context.Context, liveID int) error
	Changes(ctx context.Context, liveID int) (repository.DraftChanges, error)
}

// reviewStore keeps moderation requests.
type reviewStore interface {
	Create(ctx context.Context, moduleID int, kind string, userID int, note string) (int, error)
	Pending(ctx context.Context, moduleID int) (bool, error)
	Get(ctx context.Context, id int) (*repository.Review, error)
	List(ctx context.Context, f repository.ReviewFilter) ([]repository.Review, error)
	Latest(ctx context.Context, moduleIDs []int) (map[int]repository.Review, error)
	Decide(ctx context.Context, id int, approve bool, adminID int, note string) error
	Cancel(ctx context.Context, id int) error
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
	List(ctx context.Context) ([]model.Specialization, error)
	NextOrder(ctx context.Context) (int, error)
	Upsert(ctx context.Context, s model.Specialization) error
	Delete(ctx context.Context, slug string) error
	SetPublished(ctx context.Context, slug string, published bool) error
	Move(ctx context.Context, slug, dir string) error
	SetCover(ctx context.Context, slug, cover string) error
	SetIcon(ctx context.Context, slug, iconURL string) error
}

type simStore interface {
	ListPublished(ctx context.Context) ([]model.Simulator, error)
	Get(ctx context.Context, slug string) (*model.Simulator, error)
	List(ctx context.Context) ([]model.Simulator, error)
	Count(ctx context.Context) (int, error)
	Upsert(ctx context.Context, s model.Simulator) error
	Delete(ctx context.Context, slug string) error
	SetPublished(ctx context.Context, slug string, published bool) error
	Move(ctx context.Context, slug, dir string) error
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
	CourseIO     courseIOStore
	Reviews      reviewStore
	Drafts       draftStore
	Sandbox      sandbox
	Code         codeRunner
	Images       imageStore
	Billing      billingStore
}

// imageStore puts uploaded images into object storage; nil when it is not configured.
type imageStore interface {
	Put(ctx context.Context, kind storage.Kind, img storage.Image) (string, error)
	Delete(ctx context.Context, publicURL string) error
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
	r.Get("/public/catalog", a.getPublicCatalog)
	r.Get("/public/specializations/{specSlug}", a.getPublicSpecialization)
	r.Get("/public/courses/{courseSlug}", a.getCoursePreview)
	r.Get("/courses/{courseSlug}/cover", a.getCourseCover)
	r.Get("/specializations/{specSlug}/cover", a.getSpecializationCover)

	r.Group(func(r chi.Router) {
		r.Use(requireUser)
		r.Get("/me", a.getMe)
		r.Get("/me/profile", a.getProfile)
		r.Get("/me/subscription", a.getSubscription)
		r.Post("/billing/checkout", a.startCheckout)
		// Stands in for a provider webhook until one exists. Kept behind auth so
		// the stub cannot be poked from outside; a real webhook would not be.
		r.Post("/billing/confirm", a.confirmCheckout)
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
			r.Post("/admin/uploads/images", a.adminUploadImage)
			r.Get("/admin/courses", a.adminListCourses)
			r.Post("/admin/courses", a.adminCreateCourse)
			r.Get("/admin/courses/{courseId}", a.adminGetCourse)
			r.Put("/admin/courses/{courseId}", a.adminUpdateCourse)
			r.Delete("/admin/courses/{courseId}", a.adminDeleteCourse)
			r.Put("/admin/courses/{courseId}/published", a.adminSetCoursePublished)
			r.With(requireAdmin).Post("/admin/courses/{courseId}/move", a.adminMoveCourse)
			// What the subscription is worth is an admin decision, not an author's.
			r.With(requireAdmin).Patch("/admin/courses/{courseId}/access-tier", a.setCourseAccessTier)
			r.Put("/admin/courses/{courseId}/cover", a.adminUploadCourseCover)
			r.Delete("/admin/courses/{courseId}/cover", a.adminDeleteCourseCover)
			r.Put("/admin/courses/{courseId}/icon", a.adminUploadCourseIcon)
			r.Delete("/admin/courses/{courseId}/icon", a.adminDeleteCourseIcon)
			r.Get("/admin/courses/{courseId}/authors", a.adminListCourseAuthors)
			r.Post("/admin/courses/{courseId}/authors", a.adminAddCourseAuthor)
			r.Delete("/admin/courses/{courseId}/authors/{userId}", a.adminRemoveCourseAuthor)
			r.With(requireAdmin).Put("/admin/courses/{courseId}/owner", a.adminSetCourseOwner)
			r.Get("/admin/courses/{courseId}/export", a.adminExportCourse)
			r.Post("/admin/import/preview", a.adminPreviewImport)
			r.Post("/admin/import", a.adminApplyImport)
			r.Get("/admin/specializations", a.adminListSpecializations)
			r.Post("/admin/courses/{courseId}/review", a.adminRequestCourseReview)
			r.Post("/admin/courses/{courseId}/draft", a.adminOpenCourseDraft)
			r.Delete("/admin/courses/{courseId}/draft", a.adminDiscardCourseDraft)
			r.Get("/admin/courses/{courseId}/draft/changes", a.adminGetDraftChanges)
			r.Get("/admin/reviews", a.adminListReviews)
			r.Delete("/admin/reviews/{reviewId}", a.adminCancelReview)
			r.With(requireAdmin).Post("/admin/reviews/{reviewId}/approve", a.adminApproveReview)
			r.With(requireAdmin).Post("/admin/reviews/{reviewId}/reject", a.adminRejectReview)

			r.Group(func(r chi.Router) {
				r.Use(requireAdmin)
				r.Post("/admin/specializations", a.adminCreateSpecialization)
				r.Put("/admin/specializations/{specSlug}", a.adminUpdateSpecialization)
				r.Delete("/admin/specializations/{specSlug}", a.adminDeleteSpecialization)
				r.Put("/admin/specializations/{specSlug}/published", a.adminSetSpecializationPublished)
				r.Post("/admin/specializations/{specSlug}/move", a.adminMoveSpecialization)
				r.Put("/admin/specializations/{specSlug}/cover", a.adminUploadSpecializationCover)
				r.Delete("/admin/specializations/{specSlug}/cover", a.adminDeleteSpecializationCover)
				r.Put("/admin/specializations/{specSlug}/icon", a.adminUploadSpecIcon)
				r.Delete("/admin/specializations/{specSlug}/icon", a.adminDeleteSpecIcon)

				r.Get("/admin/simulators", a.adminListSimulators)
				r.Post("/admin/simulators", a.adminCreateSimulator)
				r.Get("/admin/simulators/{simSlug}", a.adminGetSimulator)
				r.Put("/admin/simulators/{simSlug}", a.adminUpdateSimulator)
				r.Delete("/admin/simulators/{simSlug}", a.adminDeleteSimulator)
				r.Put("/admin/simulators/{simSlug}/published", a.adminSetSimulatorPublished)
				r.Post("/admin/simulators/{simSlug}/move", a.adminMoveSimulator)

				r.Get("/admin/users", a.adminListUsers)
				r.Post("/admin/users", a.adminCreateUser)
				r.Patch("/admin/users/{userId}", a.adminUpdateUser)
				r.Delete("/admin/users/{userId}", a.adminDeleteUser)
				r.Put("/admin/users/{userId}/password", a.adminSetUserPassword)
			})

			r.Post("/admin/courses/{courseId}/lessons", a.adminCreateLesson)
			r.Get("/admin/lessons/{lessonId}", a.adminGetLesson)
			r.Put("/admin/lessons/{lessonId}", a.adminUpdateLesson)
			r.Delete("/admin/lessons/{lessonId}", a.adminDeleteLesson)
			r.Put("/admin/lessons/{lessonId}/published", a.adminSetLessonPublished)
			r.Post("/admin/lessons/{lessonId}/draft", a.adminOpenLessonDraft)
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
