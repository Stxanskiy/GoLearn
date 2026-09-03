package model

import "time"

// Module represents a learning module (group of lessons).
type Module struct {
	ID            int       `json:"id" db:"id"`
	Slug          string    `json:"slug" db:"slug"`
	Title         string    `json:"title" db:"title"`
	Description   string    `json:"description" db:"description"`
	OrderNum      int       `json:"order_num" db:"order_num"`
	Track         string    `json:"track" db:"track"`               // backend | devops | shared
	Difficulty    string    `json:"difficulty" db:"difficulty"`       // beginner | intermediate | advanced | expert
	Prerequisites []string  `json:"prerequisites" db:"prerequisites"` // module slugs
	Category      string    `json:"category" db:"category"`           // explicit catalog tag; empty -> derived
	Label         string    `json:"label" db:"label"`                 // Старт | Практика | Вызов; empty -> derived
	Tags          []string  `json:"tags"`                             // topic chips (JSON array in DB)
	CoverImage    string    `json:"cover_image" db:"cover_image"`     // real photo; empty -> generated SVG
	Accent        string    `json:"accent" db:"accent"`               // gradient key; empty -> by category
	EstMinutes    int       `json:"est_minutes" db:"est_minutes"`     // 0 -> derived from lesson count
	Source        string    `json:"source" db:"source"`               // seed | admin
	Published     bool      `json:"published" db:"published"`          // false -> draft, hidden from students
	OwnerID       *int      `json:"owner_id" db:"owner_id"`            // admin who owns it; nil -> system/shared
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Lesson represents a single lesson within a module.
type Lesson struct {
	ID         int       `json:"id" db:"id"`
	ModuleID   int       `json:"module_id" db:"module_id"`
	Slug       string    `json:"slug" db:"slug"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	OrderNum   int       `json:"order_num" db:"order_num"`
	Difficulty string    `json:"difficulty" db:"difficulty"` // beginner | intermediate | advanced | expert
	Track      string    `json:"track" db:"track"`           // backend | devops | shared
	Kind       string    `json:"kind" db:"kind"`             // theory | quiz | lab | sim
	Format     string    `json:"format" db:"format"`         // html | md
	VMImage    string    `json:"vm_image" db:"vm_image"`     // lab terminal image
	VMInit     string    `json:"vm_init" db:"vm_init"`       // lab setup reference/script
	Source     string    `json:"source" db:"source"`         // seed | admin
	Published  bool      `json:"published" db:"published"`   // false -> draft, hidden from students
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Specialization is a top-level section (admin-managed).
type Specialization struct {
	Slug        string `json:"slug" db:"slug"`
	Name        string `json:"name" db:"name"`
	Icon        string `json:"icon" db:"icon"`
	Description string `json:"description" db:"description"`
	OrderNum    int    `json:"order_num" db:"order_num"`
	CoverImage  string `json:"cover_image" db:"cover_image"`
	Published   bool   `json:"published" db:"published"` // false -> draft, hidden from students
	OwnerID     *int   `json:"owner_id" db:"owner_id"`   // admin who owns it; nil -> system/shared
}

// Simulator is an admin-managed turn-based scenario. Data holds the full scenario
// JSON (metrics/turns/choices); the other fields mirror it for cheap listing.
type Simulator struct {
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Icon      string `json:"icon"`
	Role      string `json:"role"`
	OrderNum  int    `json:"order_num"`
	Published bool   `json:"published"`
	OwnerID   *int   `json:"owner_id"`
	Data      string `json:"data"` // full Scenario JSON
}

// Quiz represents a quiz attached to a lesson.
type Quiz struct {
	ID       int    `json:"id" db:"id"`
	LessonID int    `json:"lesson_id" db:"lesson_id"`
	Title    string `json:"title" db:"title"`
}

// QuizQuestion is a single question in a quiz.
type QuizQuestion struct {
	ID            int      `json:"id" db:"id"`
	QuizID        int      `json:"quiz_id" db:"quiz_id"`
	Question      string   `json:"question" db:"question"`
	Options       []string `json:"options"`         // JSON array in DB
	OptionExpl    []string `json:"option_expl"`     // per-option explanation (JSON array in DB)
	CorrectIndex  int      `json:"correct_index" db:"correct_index"`
	Explanation   string   `json:"explanation" db:"explanation"`
	OrderNum      int      `json:"order_num" db:"order_num"`
}

// GlossaryItem explains a function or concept used in a task.
type GlossaryItem struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

// TestCase for automatic code checking.
type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}

// Task represents a coding task/assignment for a lesson.
type Task struct {
	ID          int             `json:"id" db:"id"`
	LessonID    int             `json:"lesson_id" db:"lesson_id"`
	Title       string          `json:"title" db:"title"`
	Description string          `json:"description" db:"description"`
	Hints       string          `json:"hints" db:"hints"`
	Solution    string          `json:"solution" db:"solution"`
	OrderNum    int             `json:"order_num" db:"order_num"`
	Difficulty  string          `json:"difficulty" db:"difficulty"`     // easy | medium | hard
	Glossary    []GlossaryItem  `json:"glossary"`
	TestCases   []TestCase      `json:"test_cases"`
	StarterCode string          `json:"starter_code" db:"starter_code"`
	Format       string `json:"format" db:"format"`               // html | md
	Kind         string `json:"kind" db:"kind"`                   // go | shell
	SandboxImage string `json:"sandbox_image" db:"sandbox_image"` // shell tasks
	SetupScript  string `json:"setup_script" db:"setup_script"`
	CheckScript  string `json:"check_script" db:"check_script"`
}

// LessonBundle is a lesson together with its quiz questions and tasks — the unit
// the course import/export moves around.
type LessonBundle struct {
	Lesson    Lesson
	Questions []QuizQuestion
	Tasks     []Task
}

// CourseTree is a whole course (module + its lessons with quiz/tasks). It is the
// aggregate the admin import/export round-trips through.
type CourseTree struct {
	Module  Module
	Lessons []LessonBundle
}

// Submission tracks a user's code attempt.
type Submission struct {
	ID        int       `json:"id" db:"id"`
	TaskID    int       `json:"task_id" db:"task_id"`
	Code      string    `json:"code" db:"code"`
	Output    string    `json:"output" db:"output"`
	Errors    string    `json:"errors" db:"errors"`
	Passed    bool      `json:"passed" db:"passed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Progress tracks user's learning progress.
type Progress struct {
	ID           int       `json:"id" db:"id"`
	LessonID     int       `json:"lesson_id" db:"lesson_id"`
	Status       string    `json:"status" db:"status"` // "not_started", "in_progress", "completed"
	QuizScore    *int      `json:"quiz_score,omitempty" db:"quiz_score"`
	QuizTotal    *int      `json:"quiz_total,omitempty" db:"quiz_total"`
	Notes        string    `json:"notes" db:"notes"` // Personal notes
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// DashboardStats for the main page.
type DashboardStats struct {
	TotalLessons    int     `json:"total_lessons"`
	CompletedCount  int     `json:"completed_count"`
	InProgressCount int     `json:"in_progress_count"`
	AvgQuizScore    float64 `json:"avg_quiz_score"`
	CurrentStreak   int     `json:"current_streak"`
}

// ProgressOverview powers the "Мой прогресс" dashboard (devops404 style).
type ProgressOverview struct {
	Streak        int
	ActiveDays    int
	TasksSolved   int
	LabsDone      int
	LabsTotal     int
	ArticlesRead  int
	ArticlesTotal int
	TestsPassed   int
	Simulators    int
	SimulatorsTot int
	Trainers      int
	TrainersTot   int
	Activity      map[string]int // "YYYY-MM-DD" -> activity count
}
