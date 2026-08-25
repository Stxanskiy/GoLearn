package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type ModulePageData struct {
	PageTitle   string
	Module      *model.Module
	Category    string
	Icon        string
	Cover       string
	Items       []CourseItem
	DoneCount   int
	TotalCount  int
	Pct         int
	ContinueURL string
	PrevCourse  *model.Module // previous course in this specialization
	NextCourse  *model.Module // next course — where to go once this one is done
}

// CourseItem is one row in the "Содержание курса" list: a lesson, quiz, or lab.
type CourseItem struct {
	Kind    string // display: Урок | Тест | Лаб. работа
	KindKey string // css/icon key: lesson | quiz | lab
	Title   string
	URL     string
	Status  string // completed | in_progress | not_started
	Num     int
	IsNext  bool
}

type LessonPageData struct {
	PageTitle      string
	Module         *model.Module
	Lesson         *model.Lesson
	ContentHTML    string
	Progress       *model.Progress
	Kind           string
	HasQuiz        bool
	HasTasks       bool
	QuestionsJSON  template.JS
	PrevLesson     *model.Lesson
	NextLesson     *model.Lesson
	Chapters       []LessonWithProgress // all lessons in the module (sidebar TOC)
	CurrentIndex   int                  // 1-based position of current lesson
	ModuleProgress int                  // % of module completed
}

// lessonKindLabel maps a lesson.kind to its display label and css/icon key.
func lessonKindLabel(kind string) (label, key string) {
	switch kind {
	case "quiz":
		return "Тест", "quiz"
	case "lab":
		return "Лаб. работа", "lab"
	case "sim":
		return "Симулятор", "sim"
	case "sql":
		return "SQL", "sql"
	default:
		return "Урок", "lesson"
	}
}

func (h *Handler) ModulePage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lessons, err := h.lessonRepo.GetByModule(ctx, mod.ID)
	if err != nil {
		h.log.Error("get lessons", "error", err)
		http.Error(w, "Internal error", 500)
		return
	}

	uid := currentUserID(ctx)
	allProgress, _ := h.progressRepo.GetAll(ctx, uid)
	progMap := make(map[int]model.Progress)
	for _, p := range allProgress {
		progMap[p.LessonID] = p
	}
	labStatus, _ := h.submissionRepo.LessonLabStatus(ctx, uid)

	cat := mod.Category
	if cat == "" {
		cat = categorize(mod.Track, mod.Title, mod.Slug)
	}
	cover := "/api/courses/" + mod.Slug + "/cover"

	data := ModulePageData{
		PageTitle: mod.Title,
		Module:    mod,
		Category:  cat,
		Icon:      categoryIcon(cat),
		Cover:     cover,
	}

	base := "/module/" + mod.Slug + "/lesson/"
	for i, l := range lessons {
		p := progMap[l.ID]
		kind, key := lessonKindLabel(l.Kind)

		status := p.Status
		if status == "" {
			status = "not_started"
		}
		switch l.Kind {
		case "quiz":
			if p.QuizScore != nil {
				status = "completed"
			} else if status == "completed" {
				status = "in_progress"
			}
		case "lab":
			if labStatus[l.ID] {
				status = "completed"
			} else if status == "completed" {
				status = "in_progress"
			}
		}

		url := base + l.Slug
		if l.Kind == "lab" {
			url += "/tasks" // labs open the two-pane terminal workspace directly
		}
		data.Items = append(data.Items, CourseItem{
			Kind: kind, KindKey: key, Title: l.Title,
			URL: url, Status: status, Num: i + 1,
		})
	}

	data.TotalCount = len(data.Items)
	nextSet := false
	for i := range data.Items {
		if data.Items[i].Status == "completed" {
			data.DoneCount++
		} else if !nextSet {
			data.Items[i].IsNext = true
			data.ContinueURL = data.Items[i].URL
			nextSet = true
		}
	}
	if data.TotalCount > 0 {
		data.Pct = data.DoneCount * 100 / data.TotalCount
	}
	if data.ContinueURL == "" && len(data.Items) > 0 {
		data.ContinueURL = data.Items[0].URL // all done -> jump to start
	}

	// Courses are meant to be taken in order, so each one points at its
	// neighbours in the same specialization.
	if prev, next, err := h.moduleRepo.Neighbors(ctx, *mod, specTracks(specForTrack(mod.Track))); err == nil {
		data.PrevCourse, data.NextCourse = prev, next
	} else {
		h.log.Error("course neighbors", "module", mod.Slug, "error", err)
	}

	h.render(w, "module", &data)
}

func (h *Handler) LessonPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")
	lessonSlug := chi.URLParam(r, "lessonSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lesson, err := h.lessonRepo.GetBySlug(ctx, mod.ID, lessonSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Labs live in the two-pane terminal workspace, not the article view.
	if lesson.Kind == "lab" {
		http.Redirect(w, r, "/module/"+mod.Slug+"/lesson/"+lesson.Slug+"/tasks", http.StatusSeeOther)
		return
	}

	uid := currentUserID(ctx)
	// Mark as in_progress if not started
	progress, _ := h.progressRepo.Get(ctx, uid, lesson.ID)
	if progress == nil {
		_ = h.progressRepo.Upsert(ctx, uid, lesson.ID, "in_progress")
		progress, _ = h.progressRepo.Get(ctx, uid, lesson.ID)
	}

	// Check if quiz/tasks exist
	_, questions, _ := h.lessonRepo.GetQuiz(ctx, lesson.ID)
	tasks, _ := h.lessonRepo.GetTasks(ctx, lesson.ID)

	// Inline quiz payload (one-at-a-time client-side checking).
	var qjson []byte
	if len(questions) > 0 {
		type qp struct {
			Q  string   `json:"q"`
			O  []string `json:"o"`
			OE []string `json:"oe"`
			C  int      `json:"c"`
			E  string   `json:"e"`
		}
		payload := make([]qp, len(questions))
		for i, q := range questions {
			payload[i] = qp{Q: q.Question, O: q.Options, OE: q.OptionExpl, C: q.CorrectIndex, E: q.Explanation}
		}
		qjson, _ = json.Marshal(payload)
	}

	// Find prev/next lessons + build sidebar TOC with progress
	allLessons, _ := h.lessonRepo.GetByModule(ctx, mod.ID)
	allProgress, _ := h.progressRepo.GetAll(ctx, uid)
	labStatusMap, _ := h.submissionRepo.LessonLabStatus(ctx, uid)
	pmap := make(map[int]model.Progress)
	for _, p := range allProgress {
		pmap[p.LessonID] = p
	}

	var prev, next *model.Lesson
	var chapters []LessonWithProgress
	curIndex, completed := 0, 0
	for i, l := range allLessons {
		p := pmap[l.ID]
		status := p.Status
		if status == "" {
			status = "not_started"
		}
		// quiz/lab lessons complete via score / passed tasks, not the raw status
		switch l.Kind {
		case "quiz":
			if p.QuizScore != nil {
				status = "completed"
			}
		case "lab":
			if labStatusMap[l.ID] {
				status = "completed"
			}
		}
		if status == "completed" {
			completed++
		}
		chapters = append(chapters, LessonWithProgress{
			ID: l.ID, Slug: l.Slug, Title: l.Title, OrderNum: l.OrderNum, Status: status,
		})
		if l.ID == lesson.ID {
			curIndex = i + 1
			if i > 0 {
				prev = &allLessons[i-1]
			}
			if i < len(allLessons)-1 {
				next = &allLessons[i+1]
			}
		}
	}
	modProgress := 0
	if len(allLessons) > 0 {
		modProgress = completed * 100 / len(allLessons)
	}

	data := LessonPageData{
		PageTitle:      lesson.Title,
		Module:         mod,
		Lesson:         lesson,
		ContentHTML:    lesson.Content,
		Progress:       progress,
		Kind:           lesson.Kind,
		HasQuiz:        len(questions) > 0,
		HasTasks:       len(tasks) > 0,
		QuestionsJSON:  template.JS(qjson),
		PrevLesson:     prev,
		NextLesson:     next,
		Chapters:       chapters,
		CurrentIndex:   curIndex,
		ModuleProgress: modProgress,
	}

	h.render(w, "lesson", &data)
}
