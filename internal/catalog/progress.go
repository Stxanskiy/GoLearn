package catalog

import (
	"sort"
	"time"

	"github.com/backendraz/golearn/internal/model"
)

// Progress statuses.
const (
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
)

// UserProgress is one user's raw lesson progress and passed labs.
type UserProgress struct {
	Lessons   map[int]model.Progress // lesson id → progress row
	LabPassed map[int]bool           // lesson id → all tasks passed
}

// NewUserProgress indexes progress rows by lesson id.
func NewUserProgress(rows []model.Progress, labPassed map[int]bool) UserProgress {
	up := UserProgress{Lessons: make(map[int]model.Progress, len(rows)), LabPassed: labPassed}
	for _, p := range rows {
		up.Lessons[p.LessonID] = p
	}
	if up.LabPassed == nil {
		up.LabPassed = map[int]bool{}
	}
	return up
}

// LessonStatus is the single source of truth for a lesson's status: quizzes complete by score, labs by passed tasks.
func (up UserProgress) LessonStatus(l model.Lesson) string {
	p, ok := up.Lessons[l.ID]
	raw := StatusNotStarted
	if ok && p.Status != "" {
		raw = p.Status
	}
	var done bool
	switch l.Kind {
	case "quiz":
		done = ok && p.QuizScore != nil
	case "lab":
		done = up.LabPassed[l.ID]
	default:
		return raw
	}
	switch {
	case done:
		return StatusCompleted
	case raw != StatusNotStarted:
		return StatusInProgress
	default:
		return StatusNotStarted
	}
}

// LessonState is a lesson with its derived status.
type LessonState struct {
	Lesson model.Lesson
	Status string
}

// Course is a module with its published lessons and the user's derived progress.
type Course struct {
	Module     model.Module
	Spec       string
	Category   string
	Label      string
	EstMinutes int
	Lessons    []LessonState
	Completed  int
	Pct        int
	Status     string
	// LastActivity is the newest lesson progress timestamp; zero when untouched.
	LastActivity time.Time
}

// BuildCourse derives course progress from its published lessons (in order).
func BuildCourse(m model.Module, lessons []model.Lesson, up UserProgress) Course {
	c := Course{
		Module:     m,
		Spec:       SpecForTrack(m.Track),
		Category:   Category(m),
		Label:      LabelCode(m),
		EstMinutes: EstMinutes(m, len(lessons)),
		Lessons:    make([]LessonState, 0, len(lessons)),
		Status:     StatusNotStarted,
	}
	started := false
	for _, l := range lessons {
		s := up.LessonStatus(l)
		c.Lessons = append(c.Lessons, LessonState{Lesson: l, Status: s})
		if s == StatusCompleted {
			c.Completed++
		}
		if s != StatusNotStarted {
			started = true
		}
		if p, ok := up.Lessons[l.ID]; ok && p.UpdatedAt.After(c.LastActivity) {
			c.LastActivity = p.UpdatedAt
		}
	}
	if len(lessons) > 0 {
		c.Pct = c.Completed * 100 / len(lessons)
	}
	switch {
	case len(lessons) > 0 && c.Completed == len(lessons):
		c.Status = StatusCompleted
	case started:
		c.Status = StatusInProgress
	}
	return c
}

// NextLesson returns the index of the first not completed lesson, or -1 when all are done.
func (c Course) NextLesson() int {
	for i, l := range c.Lessons {
		if l.Status != StatusCompleted {
			return i
		}
	}
	return -1
}

// statusRank orders the catalogue: what is being learned now, then what is left, then what is done.
func statusRank(status string) int {
	switch status {
	case StatusInProgress:
		return 0
	case StatusCompleted:
		return 2
	default:
		return 1
	}
}

// SortByProgress puts started courses first (newest activity first), then untouched, then completed.
// Courses of the same rank keep their curriculum order.
func SortByProgress(courses []Course) {
	sort.SliceStable(courses, func(i, j int) bool {
		a, b := courses[i], courses[j]
		if ra, rb := statusRank(a.Status), statusRank(b.Status); ra != rb {
			return ra < rb
		}
		if a.Status == StatusInProgress && !a.LastActivity.Equal(b.LastActivity) {
			return a.LastActivity.After(b.LastActivity)
		}
		return false
	})
}
