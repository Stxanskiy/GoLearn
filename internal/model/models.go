package model

import "time"

// Module represents a learning module (group of lessons).
type Module struct {
	ID          int       `json:"id" db:"id"`
	Slug        string    `json:"slug" db:"slug"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	OrderNum    int       `json:"order_num" db:"order_num"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Lesson represents a single lesson within a module.
type Lesson struct {
	ID        int       `json:"id" db:"id"`
	ModuleID  int       `json:"module_id" db:"module_id"`
	Slug      string    `json:"slug" db:"slug"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`       // Markdown content
	OrderNum  int       `json:"order_num" db:"order_num"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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

// Task represents a coding task/assignment for a lesson.
type Task struct {
	ID          int    `json:"id" db:"id"`
	LessonID    int    `json:"lesson_id" db:"lesson_id"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"` // Markdown
	Hints       string `json:"hints" db:"hints"`             // Markdown
	Solution    string `json:"solution" db:"solution"`       // Markdown
	OrderNum    int    `json:"order_num" db:"order_num"`
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
