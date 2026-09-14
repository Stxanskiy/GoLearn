package api

import (
	"context"
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
}

func newFakeContent() *fakeContent {
	return &fakeContent{progress: map[int][]model.Progress{}, labPassed: map[int]map[int]bool{}}
}

func (f *fakeContent) stores(users *fakeUsers) Stores {
	return Stores{
		Users:       users,
		Modules:     fakeModules{f},
		Lessons:     fakeLessons{f},
		Progress:    fakeProgress{f},
		Submissions: fakeSubmissions{f},
		Specs:       fakeSpecs{f},
		Sims:        fakeSims{f},
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

type fakeProgress struct{ *fakeContent }

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
	return f.labPassed[userID], nil
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
