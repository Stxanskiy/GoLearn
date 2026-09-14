// Package courseio is the native GoLearn course interchange format: a whole
// course (module + lessons + quiz + tasks, including lab setup/check scripts)
// round-trips losslessly through the Course DTO. The admin import/export uses it
// so a course can be exported to JSON, edited, and re-imported.
package courseio

import (
	"encoding/json"
	"slices"
	"strconv"

	"github.com/backendraz/golearn/internal/lab"
	"github.com/backendraz/golearn/internal/model"
)

// Course is the top-level exchange document.
type Course struct {
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Track       string   `json:"track,omitempty"`
	Difficulty  string   `json:"difficulty,omitempty"`
	Category    string   `json:"category,omitempty"`
	Label       string   `json:"label,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Accent      string   `json:"accent,omitempty"`
	OrderNum    int      `json:"order_num,omitempty"`
	EstMinutes  int      `json:"est_minutes,omitempty"`
	CoverImage  string   `json:"cover_image,omitempty"`
	Lessons     []Lesson `json:"lessons"`
}

// Lesson mirrors a lesson row plus its quiz and tasks.
type Lesson struct {
	Slug       string     `json:"slug"`
	Title      string     `json:"title"`
	Kind       string     `json:"kind,omitempty"`   // theory | quiz | lab | sim
	Format     string     `json:"format,omitempty"` // html | md
	Difficulty string     `json:"difficulty,omitempty"`
	Content    string     `json:"content,omitempty"`
	VMImage    string     `json:"vm_image,omitempty"`
	VMInit     string     `json:"vm_init,omitempty"`
	Quiz       []Question `json:"quiz,omitempty"`
	Tasks      []Task     `json:"tasks,omitempty"`
}

// Question is one quiz question. Correct is 1-based (matches the admin UI and is
// friendlier to hand-edit); it maps to the 0-based correct_index in the DB.
type Question struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	OptionExpl  []string `json:"option_expl,omitempty"`
	Correct     int      `json:"correct"`
	Explanation string   `json:"explanation,omitempty"`
}

// Task mirrors a task row, including the lab fixture scripts.
type Task struct {
	Title        string               `json:"title"`
	Description  string               `json:"description,omitempty"`
	Hints        string               `json:"hints,omitempty"`
	Solution     string               `json:"solution,omitempty"`
	Difficulty   string               `json:"difficulty,omitempty"`
	Kind         string               `json:"kind,omitempty"`   // go | shell
	Format       string               `json:"format,omitempty"` // html | md
	StarterCode  string               `json:"starter_code,omitempty"`
	SandboxImage string               `json:"sandbox_image,omitempty"`
	SetupScript  string               `json:"setup_script,omitempty"`
	CheckScript  string               `json:"check_script,omitempty"`
	Glossary     []model.GlossaryItem `json:"glossary,omitempty"`
	TestCases    []model.TestCase     `json:"test_cases,omitempty"`
}

// FromTree converts a loaded CourseTree into the exchange DTO.
func FromTree(t model.CourseTree) Course {
	m := t.Module
	c := Course{
		Slug: m.Slug, Title: m.Title, Description: m.Description, Track: m.Track,
		Difficulty: m.Difficulty, Category: m.Category, Label: m.Label, Tags: m.Tags,
		Accent: m.Accent, OrderNum: m.OrderNum, EstMinutes: m.EstMinutes, CoverImage: m.CoverImage,
	}
	for _, lb := range t.Lessons {
		l := Lesson{
			Slug: lb.Lesson.Slug, Title: lb.Lesson.Title, Kind: lb.Lesson.Kind,
			Format: lb.Lesson.Format, Difficulty: lb.Lesson.Difficulty, Content: lb.Lesson.Content,
			VMImage: lb.Lesson.VMImage, VMInit: lb.Lesson.VMInit,
		}
		for _, q := range lb.Questions {
			l.Quiz = append(l.Quiz, Question{
				Question: q.Question, Options: q.Options, OptionExpl: q.OptionExpl,
				Correct: q.CorrectIndex + 1, Explanation: q.Explanation,
			})
		}
		for _, tk := range lb.Tasks {
			l.Tasks = append(l.Tasks, Task{
				Title: tk.Title, Description: tk.Description, Hints: tk.Hints, Solution: tk.Solution,
				Difficulty: tk.Difficulty, Kind: tk.Kind, Format: tk.Format, StarterCode: tk.StarterCode,
				SandboxImage: tk.SandboxImage, SetupScript: tk.SetupScript, CheckScript: tk.CheckScript,
				Glossary: tk.Glossary, TestCases: tk.TestCases,
			})
		}
		c.Lessons = append(c.Lessons, l)
	}
	return c
}

// ToTree converts the DTO into a CourseTree with order numbers assigned from
// position. IDs are left zero — the repository resolves them by slug on upsert.
func (c Course) ToTree() model.CourseTree {
	m := model.Module{
		Slug: c.Slug, Title: c.Title, Description: c.Description, Track: c.Track,
		Difficulty: c.Difficulty, Category: c.Category, Label: c.Label, Tags: c.Tags,
		Accent: c.Accent, OrderNum: c.OrderNum, EstMinutes: c.EstMinutes, CoverImage: c.CoverImage,
	}
	t := model.CourseTree{Module: m}
	for li, l := range c.Lessons {
		lb := model.LessonBundle{Lesson: model.Lesson{
			Slug: l.Slug, Title: l.Title, Kind: l.Kind, Format: l.Format,
			Difficulty: l.Difficulty, Content: l.Content, VMImage: l.VMImage, VMInit: l.VMInit,
			OrderNum: li + 1,
		}}
		for qi, q := range l.Quiz {
			ci := q.Correct - 1
			if ci < 0 {
				ci = 0
			}
			lb.Questions = append(lb.Questions, model.QuizQuestion{
				Question: q.Question, Options: q.Options, OptionExpl: q.OptionExpl,
				CorrectIndex: ci, Explanation: q.Explanation, OrderNum: qi + 1,
			})
		}
		for ti, tk := range l.Tasks {
			lb.Tasks = append(lb.Tasks, model.Task{
				Title: tk.Title, Description: tk.Description, Hints: tk.Hints, Solution: tk.Solution,
				Difficulty: tk.Difficulty, Kind: tk.Kind, Format: tk.Format, StarterCode: tk.StarterCode,
				SandboxImage: tk.SandboxImage, SetupScript: tk.SetupScript, CheckScript: tk.CheckScript,
				Glossary: tk.Glossary, TestCases: tk.TestCases, OrderNum: ti + 1,
			})
		}
		t.Lessons = append(t.Lessons, lb)
	}
	return t
}

// Issue is a validation finding. Level is "error" (blocks import) or "warn".
type Issue struct {
	Level      string `json:"level"`
	Code       string `json:"code"`
	LessonSlug string `json:"lesson_slug,omitempty"`
	Path       string `json:"path,omitempty"`
	Msg        string `json:"msg"`
}

func kindOK(k string) bool {
	switch k {
	case "", "theory", "quiz", "lab", "sim", "sql":
		return true
	}
	return false
}

// Validate checks a course for structural problems. "error" issues block the
// import; "warn" issues are advisory. The course slug and title are checked by the caller.
func Validate(c Course) []Issue {
	var out []Issue
	add := func(level, code string, l *Lesson, path, msg string) {
		is := Issue{Level: level, Code: code, Path: path, Msg: msg}
		if l != nil {
			is.LessonSlug = l.Slug
		}
		out = append(out, is)
	}
	seen := map[string]int{}
	if len(c.Lessons) == 0 {
		add("warn", "no_lessons", nil, "/lessons", "в курсе нет ни одной главы")
	}
	for i := range c.Lessons {
		l := &c.Lessons[i]
		n := strconv.Itoa(i + 1)
		lp := "/lessons/" + strconv.Itoa(i)
		switch {
		case l.Slug == "":
			add("error", "lesson_slug_empty", l, lp+"/slug", "глава "+n+": пустой slug")
		case !slugOK(l.Slug):
			add("error", "lesson_slug_invalid", l, lp+"/slug", "глава "+n+": slug «"+l.Slug+"» — только латиница, цифры, дефис")
		default:
			seen[l.Slug]++
			if seen[l.Slug] == 2 {
				add("error", "lesson_slug_duplicate", l, lp+"/slug", "дублирующийся slug главы: "+l.Slug)
			}
		}
		if l.Title == "" {
			add("error", "lesson_title_empty", l, lp+"/title", "глава "+n+": пустой заголовок")
		}
		if !kindOK(l.Kind) {
			add("error", "unknown_kind", l, lp+"/kind", "глава «"+l.Title+"»: неизвестный kind «"+l.Kind+"» (theory/quiz/lab/sim/sql)")
		}
		if l.Format != "" && l.Format != "html" && l.Format != "md" {
			add("error", "unknown_format", l, lp+"/format", "глава «"+l.Title+"»: format «"+l.Format+"» (ожидается html/md)")
		}
		for qi, q := range l.Quiz {
			qp := lp + "/quiz/" + strconv.Itoa(qi)
			if len(q.Options) < 2 {
				add("error", "too_few_options", l, qp+"/options", "глава «"+l.Title+"», вопрос "+strconv.Itoa(qi+1)+": меньше 2 вариантов")
			}
			if q.Correct < 1 || q.Correct > len(q.Options) {
				add("error", "correct_out_of_range", l, qp+"/correct", "глава «"+l.Title+"», вопрос "+strconv.Itoa(qi+1)+": correct вне диапазона 1.."+strconv.Itoa(len(q.Options)))
			}
		}
		for ti, tk := range l.Tasks {
			tp := lp + "/tasks/" + strconv.Itoa(ti)
			if tk.Title == "" {
				add("error", "task_title_empty", l, tp+"/title", "глава «"+l.Title+"», задание "+strconv.Itoa(ti+1)+": пустой заголовок")
			}
			if tk.SandboxImage != "" && !slices.Contains(lab.SandboxImages, tk.SandboxImage) {
				add("error", "unknown_sandbox_image", l, tp+"/sandbox_image", "задание «"+tk.Title+"»: неизвестный sandbox_image «"+tk.SandboxImage+"»")
			}
			if tk.CheckScript != "" && tk.SandboxImage == "" && l.VMImage == "" {
				add("warn", "check_without_image", l, tp+"/check_script", "задание «"+tk.Title+"»: есть check, но не указан sandbox_image (и у главы нет vm_image)")
			}
		}
	}
	return out
}

func slugOK(s string) bool {
	for i, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && ((r != '-' && r != '_') || i == 0) {
			return false
		}
	}
	return len(s) <= 100
}

// HasErrors reports whether any issue is an "error".
func HasErrors(issues []Issue) bool {
	for _, is := range issues {
		if is.Level == "error" {
			return true
		}
	}
	return false
}

// Marshal renders a Course as pretty JSON for download.
func Marshal(c Course) ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// Parse reads a Course from JSON bytes.
func Parse(data []byte) (Course, error) {
	var c Course
	err := json.Unmarshal(data, &c)
	return c, err
}
