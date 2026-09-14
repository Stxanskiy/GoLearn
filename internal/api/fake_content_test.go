package api

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/jackc/pgx/v5"
)

// fakeContent is in-memory content shared by the fake stores below.
type fakeContent struct {
	modules   []model.Module
	lessons   []model.Lesson
	specs     []model.Specialization
	sims      []model.Simulator
	progress  map[int][]model.Progress // user id → rows
	labPassed map[int]map[int]bool     // user id → lesson id → all tasks passed
	overview  model.ProgressOverview
	cont      *repository.ContinueLesson
	stats     repository.PlatformStats
	questions map[int][]model.QuizQuestion // lesson id → quiz questions
	tasks     map[int][]model.Task         // lesson id → tasks
	passed    map[int]map[int]bool         // user id → task id → passed
	saves     []fakeSubmission
	answers   map[int]map[int]int // user id → question id → selected
	attempts  []fakeAttempt
	sandbox   *fakeSandbox
	code      *fakeCode
}

type fakeSubmission struct {
	userID, taskID int
	code, output   string
	passed         bool
}

type fakeAttempt struct {
	userID, lessonID, score, total int
	answers                        []repository.AttemptAnswer
}

func newFakeContent() *fakeContent {
	return &fakeContent{
		progress: map[int][]model.Progress{}, labPassed: map[int]map[int]bool{},
		questions: map[int][]model.QuizQuestion{}, tasks: map[int][]model.Task{}, answers: map[int]map[int]int{},
		passed: map[int]map[int]bool{}, sandbox: newFakeSandbox(), code: &fakeCode{},
	}
}

func (f *fakeContent) stores(users *fakeUsers) Stores {
	return Stores{
		Users:        users,
		Modules:      fakeModules{f},
		Lessons:      fakeLessons{f},
		Progress:     fakeProgress{f},
		Submissions:  fakeSubmissions{f},
		Specs:        fakeSpecs{f},
		Sims:         fakeSims{f},
		QuizAnswers:  fakeQuizAnswers{f},
		QuizAttempts: fakeQuizAttempts{f},
		Sandbox:      f.sandbox,
		Code:         f.code,
	}
}

type fakeModules struct{ *fakeContent }

func (f fakeModules) GetAll(_ context.Context) ([]model.Module, error) {
	var out []model.Module
	for _, m := range f.modules {
		if m.Published {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OrderNum < out[j].OrderNum })
	return out, nil
}

func (f fakeModules) GetBySlug(_ context.Context, slug string) (*model.Module, error) {
	for _, m := range f.modules {
		if m.Slug == slug {
			cp := m
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeModules) GetByID(_ context.Context, id int) (*model.Module, error) {
	for _, m := range f.modules {
		if m.ID == id {
			cp := m
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeModules) Neighbors(ctx context.Context, m model.Module, tracks []string) (prev, next *model.Module, err error) {
	all, _ := f.GetAll(ctx)
	for i := range all {
		c := all[i]
		if !slices.Contains(tracks, c.Track) || c.ID == m.ID {
			continue
		}
		if c.OrderNum < m.OrderNum {
			prev = &c
		}
		if c.OrderNum > m.OrderNum && next == nil {
			next = &c
		}
	}
	return prev, next, nil
}

func (f fakeModules) Stats(_ context.Context) (repository.PlatformStats, error) { return f.stats, nil }

type fakeLessons struct{ *fakeContent }

func (f fakeLessons) GetByModule(_ context.Context, moduleID int) ([]model.Lesson, error) {
	var out []model.Lesson
	for _, l := range f.lessons {
		if l.ModuleID == moduleID && l.Published {
			out = append(out, l)
		}
	}
	return out, nil
}

func (f fakeLessons) ListPublishedOutline(ctx context.Context) ([]model.Lesson, error) {
	mods, _ := fakeModules{f.fakeContent}.GetAll(ctx)
	published := map[int]bool{}
	for _, m := range mods {
		published[m.ID] = true
	}
	var out []model.Lesson
	for _, l := range f.lessons {
		if l.Published && published[l.ModuleID] {
			out = append(out, l)
		}
	}
	return out, nil
}

func (f fakeLessons) GetBySlug(_ context.Context, moduleID int, slug string) (*model.Lesson, error) {
	for _, l := range f.lessons {
		if l.ModuleID == moduleID && l.Slug == slug {
			cp := l
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeLessons) GetByID(_ context.Context, id int) (*model.Lesson, error) {
	for _, l := range f.lessons {
		if l.ID == id {
			cp := l
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeLessons) GetQuiz(_ context.Context, lessonID int) (*model.Quiz, []model.QuizQuestion, error) {
	qs, ok := f.questions[lessonID]
	if !ok {
		return nil, nil, pgx.ErrNoRows
	}
	return &model.Quiz{ID: lessonID, LessonID: lessonID}, qs, nil
}

func (f fakeLessons) CountsForLesson(_ context.Context, lessonID int) (questions, tasks int) {
	return len(f.questions[lessonID]), len(f.tasks[lessonID])
}

func (f fakeLessons) GetTasks(_ context.Context, lessonID int) ([]model.Task, error) {
	return f.tasks[lessonID], nil
}

func (f fakeLessons) GetTaskByID(_ context.Context, taskID int) (*model.Task, error) {
	for _, ts := range f.tasks {
		for _, t := range ts {
			if t.ID == taskID {
				cp := t
				return &cp, nil
			}
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeLessons) LessonSandbox(_ context.Context, lessonID int) (string, string, error) {
	return "golearn/sandbox:latest", fmt.Sprintf("setup-%d", lessonID), nil
}

type fakeProgress struct{ *fakeContent }

func (f fakeProgress) row(userID, lessonID int) *model.Progress {
	for i := range f.progress[userID] {
		if f.progress[userID][i].LessonID == lessonID {
			return &f.progress[userID][i]
		}
	}
	f.progress[userID] = append(f.progress[userID], model.Progress{LessonID: lessonID, Status: "in_progress"})
	return &f.progress[userID][len(f.progress[userID])-1]
}

func (f fakeProgress) Get(_ context.Context, userID, lessonID int) (*model.Progress, error) {
	for _, p := range f.progress[userID] {
		if p.LessonID == lessonID {
			cp := p
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeProgress) Start(_ context.Context, userID, lessonID int) error {
	f.row(userID, lessonID)
	return nil
}

func (f fakeProgress) Upsert(_ context.Context, userID, lessonID int, status string) error {
	f.row(userID, lessonID).Status = status
	return nil
}

func (f fakeProgress) SaveNotes(_ context.Context, userID, lessonID int, notes string) error {
	f.row(userID, lessonID).Notes = notes
	return nil
}

func (f fakeProgress) ResetLesson(_ context.Context, userID, lessonID int) error {
	p := f.row(userID, lessonID)
	p.Status, p.QuizScore, p.QuizTotal = "not_started", nil, nil
	return nil
}

func (f fakeProgress) SaveQuizResult(_ context.Context, userID, lessonID, score, total int) error {
	p := f.row(userID, lessonID)
	p.QuizScore, p.QuizTotal = &score, &total
	return nil
}

func (f fakeProgress) GetAll(_ context.Context, userID int) ([]model.Progress, error) {
	return f.progress[userID], nil
}

func (f fakeProgress) Overview(_ context.Context, _ int) (*model.ProgressOverview, error) {
	o := f.overview
	return &o, nil
}

func (f fakeProgress) LatestInProgress(_ context.Context, _ int) (*repository.ContinueLesson, error) {
	return f.cont, nil
}

type fakeSubmissions struct{ *fakeContent }

func (f fakeSubmissions) LessonLabStatus(_ context.Context, userID int) (map[int]bool, error) {
	out := map[int]bool{}
	for id, ok := range f.labPassed[userID] {
		out[id] = ok
	}
	for lessonID, ts := range f.tasks {
		all := len(ts) > 0
		for _, t := range ts {
			all = all && f.passed[userID][t.ID]
		}
		if all {
			out[lessonID] = true
		}
	}
	return out, nil
}

func (f fakeSubmissions) PassedTaskIDs(_ context.Context, userID, lessonID int) (map[int]bool, error) {
	out := map[int]bool{}
	for _, t := range f.tasks[lessonID] {
		if f.passed[userID][t.ID] {
			out[t.ID] = true
		}
	}
	return out, nil
}

func (f fakeSubmissions) Save(_ context.Context, userID, taskID int, code, output, _ string, passed bool) error {
	f.fakeContent.saves = append(f.fakeContent.saves, fakeSubmission{userID, taskID, code, output, passed})
	if passed {
		if f.passed[userID] == nil {
			f.passed[userID] = map[int]bool{}
		}
		f.passed[userID][taskID] = true
	}
	return nil
}

func (f fakeSubmissions) ResetLesson(_ context.Context, userID, lessonID int) error {
	for _, t := range f.tasks[lessonID] {
		delete(f.passed[userID], t.ID)
	}
	return nil
}

type fakeSpecs struct{ *fakeContent }

func (f fakeSpecs) ListPublished(_ context.Context) ([]model.Specialization, error) {
	var out []model.Specialization
	for _, s := range f.specs {
		if s.Published {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f fakeSpecs) Get(_ context.Context, slug string) (*model.Specialization, error) {
	for _, s := range f.specs {
		if s.Slug == slug {
			cp := s
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

type fakeSims struct{ *fakeContent }

func (f fakeSims) ListPublished(_ context.Context) ([]model.Simulator, error) {
	var out []model.Simulator
	for _, s := range f.sims {
		if s.Published {
			out = append(out, s)
		}
	}
	return out, nil
}

func (f fakeSims) Get(_ context.Context, slug string) (*model.Simulator, error) {
	for _, s := range f.sims {
		if s.Slug == slug {
			cp := s
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

type fakeQuizAnswers struct{ *fakeContent }

func (f fakeQuizAnswers) Record(_ context.Context, userID, questionID, selected int) (int, bool, error) {
	if f.answers[userID] == nil {
		f.answers[userID] = map[int]int{}
	}
	if stored, ok := f.answers[userID][questionID]; ok {
		return stored, false, nil
	}
	f.answers[userID][questionID] = selected
	return selected, true, nil
}

func (f fakeQuizAnswers) ForLesson(_ context.Context, userID, lessonID int) (map[int]int, error) {
	out := map[int]int{}
	for _, q := range f.questions[lessonID] {
		if sel, ok := f.answers[userID][q.ID]; ok {
			out[q.ID] = sel
		}
	}
	return out, nil
}

func (f fakeQuizAnswers) ResetLesson(_ context.Context, userID, lessonID int) error {
	for _, q := range f.questions[lessonID] {
		delete(f.answers[userID], q.ID)
	}
	return nil
}

type fakeQuizAttempts struct{ *fakeContent }

func (f fakeQuizAttempts) Save(ctx context.Context, userID, lessonID, score, total int, answers []repository.AttemptAnswer) (int, error) {
	f.fakeContent.attempts = append(f.fakeContent.attempts, fakeAttempt{userID, lessonID, score, total, answers})
	return f.Count(ctx, userID, lessonID)
}

func (f fakeQuizAttempts) Count(_ context.Context, userID, lessonID int) (int, error) {
	n := 0
	for _, a := range f.attempts {
		if a.userID == userID && a.lessonID == lessonID {
			n++
		}
	}
	return n, nil
}
