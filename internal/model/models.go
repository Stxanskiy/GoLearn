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
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
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
	Kind         string `json:"kind" db:"kind"`                   // go | shell
	SandboxImage string `json:"sandbox_image" db:"sandbox_image"` // shell tasks
	SetupScript  string `json:"setup_script" db:"setup_script"`
	CheckScript  string `json:"check_script" db:"check_script"`
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
