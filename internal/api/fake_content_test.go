package api

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/backendraz/golearn/internal/runner"
	"github.com/jackc/pgx/v5"
)

// fakeContent is in-memory content shared by the fake stores below.
type fakeContent struct {
	modules    []model.Module
	lessons    []model.Lesson
	specs      []model.Specialization
	sims       []model.Simulator
	progress   map[int][]model.Progress // user id → rows
	labPassed  map[int]map[int]bool     // user id → lesson id → all tasks passed
	overview   model.ProgressOverview
	cont       *repository.ContinueLesson
	stats      repository.PlatformStats
	questions  map[int][]model.QuizQuestion // lesson id → quiz questions
	tasks      map[int][]model.Task         // lesson id → tasks
	passed     map[int]map[int]bool         // user id → task id → passed
	saves      []fakeSubmission
	answers    map[int]map[int]int // user id → question id → selected
	attempts   []fakeAttempt
	sandbox    *fakeSandbox
	images     *fakeImages
	coauthors  map[int][]int // module id → co-author user ids
	reviews    []repository.Review
	origins    map[int]int  // draft row id → live row id
	subscribed map[int]bool // user id → has an active subscription
	launches   map[int]int  // user id → sandboxes started inside the free window
	detached   int          // how many times content was detached from the seeder
	nextID     int
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
		passed: map[int]map[int]bool{}, sandbox: newFakeSandbox(),
		coauthors: map[int][]int{}, nextID: 5000, origins: map[int]int{},
		subscribed: map[int]bool{},
		launches:   map[int]int{},
	}
}

func (f *fakeContent) stores(users *fakeUsers) Stores {
	s := Stores{
		Users:        users,
		Modules:      fakeModules{f},
		Lessons:      fakeLessons{f},
		Progress:     fakeProgress{f},
		Submissions:  fakeSubmissions{f},
		Specs:        fakeSpecs{f},
		Sims:         fakeSims{f},
		QuizAnswers:  fakeQuizAnswers{f},
		QuizAttempts: fakeQuizAttempts{f},
		Authors:      fakeAuthors{f},
		CourseIO:     fakeCourseIO{f},
		Reviews:      fakeReviews{f},
		Drafts:       fakeDrafts{f},
		Sandbox:      f.sandbox,
		Billing:      fakeBilling{f},
		SeedDetach:   fakeSeedDetach{f},
	}
	// A typed nil in the interface would look configured, so only set it when present.
	if f.images != nil {
		s.Images = f.images
	}
	return s
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
		if !slices.Contains(tracks, c.Track) || c.ID == m.ID || c.IsTrainer {
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

func (f fakeLessons) LessonSandbox(_ context.Context, lessonID int) (runner.Spec, error) {
	return runner.Spec{Image: "golearn/sandbox:latest", Setup: fmt.Sprintf("setup-%d", lessonID)}, nil
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

func (f fakeQuizAttempts) Latest(_ context.Context, userID, lessonID int) (*repository.Attempt, error) {
	var last *fakeAttempt
	n := 0
	for i := range f.attempts {
		if f.attempts[i].userID == userID && f.attempts[i].lessonID == lessonID {
			last = &f.attempts[i]
			n++
		}
	}
	if last == nil {
		return nil, nil
	}
	return &repository.Attempt{Number: n, Score: last.score, Total: last.total, Answers: last.answers}, nil
}

// fakeBilling is enough to exercise the subscription gate. Only HasAccess is
// reachable from the read paths; the rest satisfies the interface.
type fakeBilling struct{ *fakeContent }

// Current has to agree with HasAccess: the handler reads the subscription once
// and derives everything from it, so a double that says "no subscription" while
// HasAccess says yes tests a state that cannot happen.
func (f fakeBilling) Current(_ context.Context, userID int) (*repository.Subscription, error) {
	if !f.subscribed[userID] {
		return nil, nil
	}
	until := time.Now().Add(30 * 24 * time.Hour)
	return &repository.Subscription{
		ID: userID, Status: "active", Plan: "month",
		StartedAt: time.Now().Add(-24 * time.Hour), ExpiresAt: &until, Provider: "stub",
	}, nil
}

func (f fakeBilling) HasAccess(_ context.Context, userID int) (bool, error) {
	return f.subscribed[userID], nil
}

func (f fakeBilling) StartPayment(context.Context, int, string, int, int64, string, string, string) (*repository.Payment, error) {
	return nil, errors.New("not used in tests")
}

func (f fakeBilling) SetProviderRef(context.Context, int, string) error { return nil }

func (f fakeBilling) Confirm(context.Context, string, string, int64) (*repository.Subscription, error) {
	return nil, repository.ErrPaymentNotFound
}

// The sandbox allowance: launches is a count per user, which is all the gate
// reads. Tests that never start a sandbox leave it at zero and pass the gate.
func (f fakeBilling) RecordLaunch(_ context.Context, userID int, _ string) error {
	if f.launches == nil {
		return nil
	}
	f.launches[userID]++
	return nil
}

func (f fakeBilling) CountLaunches(_ context.Context, userID int, _ time.Time) (int, error) {
	return f.launches[userID], nil
}

func (f fakeBilling) OldestLaunchSince(_ context.Context, userID int, _ time.Time) (time.Time, bool, error) {
	if f.launches[userID] == 0 {
		return time.Time{}, false, nil
	}
	return time.Now().Add(-time.Hour), true, nil
}

func (f fakeBilling) SetCourseTier(_ context.Context, moduleID int, tier string) error {
	for i := range f.modules {
		if f.modules[i].ID == moduleID {
			f.modules[i].AccessTier = tier
			return nil
		}
	}
	return repository.ErrCourseNotFound
}

// fakeSeedDetach works on the fixture's own lessons, so a test can see the
// conversion land rather than only count it.
type fakeSeedDetach struct{ *fakeContent }

func (f fakeSeedDetach) MarkdownLessons(context.Context) ([]model.Lesson, error) {
	var out []model.Lesson
	for _, l := range f.lessons {
		if l.Format == "md" {
			out = append(out, l)
		}
	}
	return out, nil
}

func (f fakeSeedDetach) SetLessonContent(_ context.Context, id int, format, body string) error {
	for i := range f.lessons {
		if f.lessons[i].ID == id {
			f.lessons[i].Format, f.lessons[i].Content = format, body
			return nil
		}
	}
	return repository.ErrCourseNotFound
}

func (f fakeSeedDetach) Detach(context.Context) (int64, int64, error) {
	f.detached++
	return int64(len(f.lessons)), int64(len(f.modules)), nil
}
