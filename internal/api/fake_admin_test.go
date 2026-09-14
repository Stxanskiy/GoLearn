package api

import (
	"context"
	"slices"
	"strings"

	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var errUnique = &pgconn.PgError{Code: "23505"}

func (f *fakeContent) id() int {
	f.nextID++
	return f.nextID
}

func (f *fakeUsers) GetByID(_ context.Context, id int) (*repository.User, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			cp := *u
			return &cp, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (f fakeModules) module(id int) *model.Module {
	for i := range f.modules {
		if f.modules[i].ID == id {
			return &f.modules[i]
		}
	}
	return nil
}

func (f fakeModules) ListManaged(_ context.Context, userID int, all bool) ([]repository.CourseRow, error) {
	var out []repository.CourseRow
	for _, m := range f.modules {
		owned := m.OwnerID != nil && *m.OwnerID == userID
		if !all && !owned && !slices.Contains(f.coauthors[m.ID], userID) {
			continue
		}
		row := repository.CourseRow{Module: m}
		for _, l := range f.lessons {
			if l.ModuleID == m.ID {
				row.Lessons++
				if l.Kind == "lab" {
					row.Labs++
				}
			}
		}
		out = append(out, row)
	}
	return out, nil
}

func (f fakeModules) NextOrder(_ context.Context) (int, error) {
	n := 0
	for _, m := range f.modules {
		n = max(n, m.OrderNum)
	}
	return n + 1, nil
}

func (f fakeModules) Create(_ context.Context, m model.Module) (int, error) {
	if _, err := f.GetBySlug(context.Background(), m.Slug); err == nil {
		return 0, errUnique
	}
	m.ID, m.Source = f.id(), "admin"
	f.modules = append(f.modules, m)
	return m.ID, nil
}

func (f fakeModules) Update(_ context.Context, m model.Module) error {
	for _, o := range f.modules {
		if o.Slug == m.Slug && o.ID != m.ID {
			return errUnique
		}
	}
	if cur := f.module(m.ID); cur != nil {
		*cur = m
	}
	return nil
}

func (f fakeModules) Delete(_ context.Context, id int) error {
	f.modules = slices.DeleteFunc(f.modules, func(m model.Module) bool { return m.ID == id })
	f.lessons = slices.DeleteFunc(f.lessons, func(l model.Lesson) bool { return l.ModuleID == id })
	return nil
}

func (f fakeModules) SetPublished(_ context.Context, id int, published bool) error {
	if m := f.module(id); m != nil {
		m.Published = published
	}
	return nil
}

func (f fakeModules) Move(_ context.Context, id int, dir string) error {
	if m := f.module(id); m != nil {
		if dir == "up" {
			m.OrderNum--
		} else {
			m.OrderNum++
		}
	}
	return nil
}

func (f fakeModules) SetCover(_ context.Context, id int, cover string) error {
	if m := f.module(id); m != nil {
		m.CoverImage = cover
	}
	return nil
}

func (f fakeLessons) lesson(id int) *model.Lesson {
	for i := range f.lessons {
		if f.lessons[i].ID == id {
			return &f.lessons[i]
		}
	}
	return nil
}

func (f fakeLessons) GetByModuleAll(_ context.Context, moduleID int) ([]model.Lesson, error) {
	var out []model.Lesson
	for _, l := range f.lessons {
		if l.ModuleID == moduleID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (f fakeLessons) NextOrder(_ context.Context, moduleID int) (int, error) {
	n := 0
	for _, l := range f.lessons {
		if l.ModuleID == moduleID {
			n = max(n, l.OrderNum)
		}
	}
	return n + 1, nil
}

func (f fakeLessons) slugTaken(l model.Lesson) bool {
	for _, o := range f.lessons {
		if o.ModuleID == l.ModuleID && o.Slug == l.Slug && o.ID != l.ID {
			return true
		}
	}
	return false
}

func (f fakeLessons) Create(_ context.Context, l model.Lesson) (int, error) {
	if f.slugTaken(l) {
		return 0, errUnique
	}
	l.ID, l.Source = f.id(), "admin"
	f.lessons = append(f.lessons, l)
	return l.ID, nil
}

func (f fakeLessons) Update(_ context.Context, l model.Lesson) error {
	if f.slugTaken(l) {
		return errUnique
	}
	if cur := f.lesson(l.ID); cur != nil {
		l.Source, l.CreatedAt = cur.Source, cur.CreatedAt
		*cur = l
	}
	return nil
}

func (f fakeLessons) Delete(_ context.Context, id int) error {
	f.lessons = slices.DeleteFunc(f.lessons, func(l model.Lesson) bool { return l.ID == id })
	delete(f.questions, id)
	delete(f.tasks, id)
	return nil
}

func (f fakeLessons) SetPublished(_ context.Context, id int, published bool) error {
	if l := f.lesson(id); l != nil {
		l.Published = published
	}
	return nil
}

func (f fakeLessons) MoveLesson(_ context.Context, id int, dir string) error {
	if l := f.lesson(id); l != nil {
		if dir == "up" {
			l.OrderNum--
		} else {
			l.OrderNum++
		}
	}
	return nil
}

func (f fakeLessons) DuplicateLesson(_ context.Context, id int) (int, error) {
	l := f.lesson(id)
	if l == nil {
		return 0, pgx.ErrNoRows
	}
	cp := *l
	cp.ID, cp.Slug, cp.Published = f.id(), l.Slug+"-copy", false
	f.lessons = append(f.lessons, cp)
	return cp.ID, nil
}

func (f fakeLessons) EnsureQuiz(_ context.Context, lessonID int, _ string) (int, error) {
	if _, ok := f.questions[lessonID]; !ok {
		f.questions[lessonID] = []model.QuizQuestion{}
	}
	return lessonID, nil
}

func (f fakeLessons) AddQuestion(_ context.Context, quizID int, q model.QuizQuestion) (int, error) {
	q.ID, q.QuizID, q.OrderNum = f.id(), quizID, len(f.questions[quizID])+1
	f.questions[quizID] = append(f.questions[quizID], q)
	return q.ID, nil
}

func (f fakeLessons) question(id int) *model.QuizQuestion {
	for lid := range f.questions {
		for i := range f.questions[lid] {
			if f.questions[lid][i].ID == id {
				return &f.questions[lid][i]
			}
		}
	}
	return nil
}

func (f fakeLessons) UpdateQuestion(_ context.Context, q model.QuizQuestion) error {
	if cur := f.question(q.ID); cur != nil {
		q.QuizID, q.OrderNum = cur.QuizID, cur.OrderNum
		*cur = q
	}
	return nil
}

func (f fakeLessons) DeleteQuestion(_ context.Context, id int) error {
	for lid := range f.questions {
		f.questions[lid] = slices.DeleteFunc(f.questions[lid], func(q model.QuizQuestion) bool { return q.ID == id })
	}
	return nil
}

func (f fakeLessons) GetQuestionByID(_ context.Context, id int) (*model.QuizQuestion, error) {
	if q := f.question(id); q != nil {
		cp := *q
		if cp.QuizID == 0 {
			for lid, qs := range f.questions {
				if slices.ContainsFunc(qs, func(o model.QuizQuestion) bool { return o.ID == id }) {
					cp.QuizID = lid
				}
			}
		}
		return &cp, nil
	}
	return nil, pgx.ErrNoRows
}

func (f fakeLessons) LessonIDForQuiz(_ context.Context, quizID int) (int, error) { return quizID, nil }

func (f fakeLessons) CreateTask(_ context.Context, t model.Task) (int, error) {
	t.ID, t.OrderNum = f.id(), len(f.tasks[t.LessonID])+1
	f.tasks[t.LessonID] = append(f.tasks[t.LessonID], t)
	return t.ID, nil
}

func (f fakeLessons) UpdateTask(_ context.Context, t model.Task) error {
	for i, cur := range f.tasks[t.LessonID] {
		if cur.ID == t.ID {
			t.OrderNum = cur.OrderNum
			f.tasks[t.LessonID][i] = t
		}
	}
	return nil
}

func (f fakeLessons) DeleteTask(_ context.Context, id int) error {
	for lid := range f.tasks {
		f.tasks[lid] = slices.DeleteFunc(f.tasks[lid], func(t model.Task) bool { return t.ID == id })
	}
	return nil
}

type fakeAuthors struct{ *fakeContent }

func (f fakeAuthors) IsCoauthor(_ context.Context, moduleID, userID int) (bool, error) {
	return slices.Contains(f.coauthors[moduleID], userID), nil
}

func (f fakeAuthors) List(_ context.Context, moduleID int) ([]repository.User, error) {
	var out []repository.User
	for _, id := range f.coauthors[moduleID] {
		out = append(out, repository.User{ID: id, Name: "user", Email: "u@example.com"})
	}
	return out, nil
}

func (f fakeAuthors) Add(_ context.Context, moduleID, userID, _ int) (bool, error) {
	if slices.Contains(f.coauthors[moduleID], userID) {
		return false, nil
	}
	f.coauthors[moduleID] = append(f.coauthors[moduleID], userID)
	return true, nil
}

func (f fakeAuthors) Remove(_ context.Context, moduleID, userID int) (bool, error) {
	before := len(f.coauthors[moduleID])
	f.coauthors[moduleID] = slices.DeleteFunc(f.coauthors[moduleID], func(id int) bool { return id == userID })
	return len(f.coauthors[moduleID]) < before, nil
}

func (f fakeAuthors) SetOwner(_ context.Context, moduleID, userID, _ int) error {
	m := fakeModules{f.fakeContent}.module(moduleID)
	if m == nil {
		return pgx.ErrNoRows
	}
	if m.OwnerID != nil && *m.OwnerID != userID && !slices.Contains(f.coauthors[moduleID], *m.OwnerID) {
		f.coauthors[moduleID] = append(f.coauthors[moduleID], *m.OwnerID)
	}
	f.coauthors[moduleID] = slices.DeleteFunc(f.coauthors[moduleID], func(id int) bool { return id == userID })
	m.OwnerID = &userID
	return nil
}

type fakeCourseIO struct{ *fakeContent }

func (f fakeCourseIO) Export(_ context.Context, moduleID int) (model.CourseTree, error) {
	m := fakeModules{f.fakeContent}.module(moduleID)
	if m == nil {
		return model.CourseTree{}, pgx.ErrNoRows
	}
	tree := model.CourseTree{Module: *m}
	for _, l := range f.lessons {
		if l.ModuleID == moduleID {
			tree.Lessons = append(tree.Lessons, model.LessonBundle{Lesson: l, Questions: f.questions[l.ID], Tasks: f.tasks[l.ID]})
		}
	}
	return tree, nil
}

func (f fakeCourseIO) Diff(ctx context.Context, tree model.CourseTree) (repository.CourseDiff, error) {
	d := repository.CourseDiff{Slug: tree.Module.Slug, Title: tree.Module.Title}
	m, err := fakeModules{f.fakeContent}.GetBySlug(ctx, tree.Module.Slug)
	if err == nil {
		d.Exists, d.ModuleID = true, m.ID
	}
	incoming := map[string]bool{}
	for _, lb := range tree.Lessons {
		incoming[lb.Lesson.Slug] = true
		ref := repository.LessonRef{Slug: lb.Lesson.Slug, Title: lb.Lesson.Title}
		if cur, err := (fakeLessons{f.fakeContent}).GetBySlug(ctx, d.ModuleID, lb.Lesson.Slug); err == nil && d.Exists && cur != nil {
			d.Updated = append(d.Updated, ref)
		} else {
			d.New = append(d.New, ref)
		}
	}
	for _, l := range f.lessons {
		if d.Exists && l.ModuleID == d.ModuleID && !incoming[l.Slug] {
			d.Removed = append(d.Removed, repository.LessonRef{Slug: l.Slug, Title: l.Title})
			d.LostSubmissions += len(f.tasks[l.ID])
		}
	}
	return d, nil
}

func (f fakeCourseIO) Upsert(ctx context.Context, tree model.CourseTree) (repository.CourseDiff, error) {
	d, _ := f.Diff(ctx, tree)
	mods, lessons := fakeModules{f.fakeContent}, fakeLessons{f.fakeContent}
	id := d.ModuleID
	if !d.Exists {
		id, _ = mods.Create(ctx, tree.Module)
	} else {
		m := mods.module(id)
		published, owner, cover := m.Published, m.OwnerID, m.CoverImage
		*m = tree.Module
		m.ID, m.Published, m.OwnerID = id, published, owner
		if m.CoverImage == "" {
			m.CoverImage = cover
		}
	}
	keep := map[string]bool{}
	for _, lb := range tree.Lessons {
		keep[lb.Lesson.Slug] = true
		l := lb.Lesson
		l.ModuleID = id
		if cur, err := lessons.GetBySlug(ctx, id, l.Slug); err == nil {
			l.ID, l.Published = cur.ID, cur.Published
			_ = lessons.Update(ctx, l)
		} else {
			l.ID, _ = lessons.Create(ctx, l)
		}
		f.questions[l.ID], f.tasks[l.ID] = lb.Questions, lb.Tasks
	}
	f.lessons = slices.DeleteFunc(f.lessons, func(l model.Lesson) bool { return l.ModuleID == id && !keep[l.Slug] })
	return d, nil
}

func (f fakeModules) TrackCounts(_ context.Context) (map[string]int, error) {
	out := map[string]int{}
	for _, m := range f.modules {
		out[m.Track]++
	}
	return out, nil
}

func (f fakeSpecs) spec(slug string) *model.Specialization {
	for i := range f.specs {
		if f.specs[i].Slug == slug {
			return &f.specs[i]
		}
	}
	return nil
}

func (f fakeSpecs) List(_ context.Context) ([]model.Specialization, error) { return f.specs, nil }

func (f fakeSpecs) NextOrder(_ context.Context) (int, error) { return len(f.specs) + 1, nil }

func (f fakeSpecs) Upsert(_ context.Context, s model.Specialization) error {
	cur := f.spec(s.Slug)
	if cur == nil {
		f.specs = append(f.specs, s)
		return nil
	}
	owner, cover := cur.OwnerID, cur.CoverImage
	*cur = s
	cur.OwnerID = owner
	if s.CoverImage == "" {
		cur.CoverImage = cover
	}
	return nil
}

func (f fakeSpecs) Delete(_ context.Context, slug string) error {
	f.specs = slices.DeleteFunc(f.specs, func(s model.Specialization) bool { return s.Slug == slug })
	return nil
}

func (f fakeSpecs) SetPublished(_ context.Context, slug string, published bool) error {
	if s := f.spec(slug); s != nil {
		s.Published = published
	}
	return nil
}

func (f fakeSpecs) Move(_ context.Context, slug, dir string) error {
	if s := f.spec(slug); s != nil && dir == "up" {
		s.OrderNum--
	}
	return nil
}

func (f fakeSpecs) SetCover(_ context.Context, slug, cover string) error {
	if s := f.spec(slug); s != nil {
		s.CoverImage = cover
	}
	return nil
}

func (f fakeSims) sim(slug string) *model.Simulator {
	for i := range f.sims {
		if f.sims[i].Slug == slug {
			return &f.sims[i]
		}
	}
	return nil
}

func (f fakeSims) List(_ context.Context) ([]model.Simulator, error) { return f.sims, nil }

func (f fakeSims) Count(_ context.Context) (int, error) { return len(f.sims), nil }

func (f fakeSims) Upsert(_ context.Context, s model.Simulator) error {
	if cur := f.sim(s.Slug); cur != nil {
		owner := cur.OwnerID
		*cur = s
		cur.OwnerID = owner
		return nil
	}
	f.sims = append(f.sims, s)
	return nil
}

func (f fakeSims) Delete(_ context.Context, slug string) error {
	f.sims = slices.DeleteFunc(f.sims, func(s model.Simulator) bool { return s.Slug == slug })
	return nil
}

func (f fakeSims) SetPublished(_ context.Context, slug string, published bool) error {
	if s := f.sim(slug); s != nil {
		s.Published = published
	}
	return nil
}

func (f fakeSims) Move(_ context.Context, slug, dir string) error {
	if s := f.sim(slug); s != nil && dir == "up" {
		s.OrderNum--
	}
	return nil
}

func (f *fakeUsers) ListPage(_ context.Context, q string, beforeID, limit int) ([]repository.User, error) {
	var out []repository.User
	for _, u := range f.byEmail {
		if (beforeID == 0 || u.ID < beforeID) && (q == "" || strings.Contains(u.Name, q) || strings.Contains(u.Email, q)) {
			out = append(out, *u)
		}
	}
	slices.SortFunc(out, func(a, b repository.User) int { return b.ID - a.ID })
	return out[:min(limit, len(out))], nil
}

func (f *fakeUsers) CreateWithRole(ctx context.Context, email, password, name, role string) (*repository.User, error) {
	u, err := f.Create(ctx, email, password, name)
	if err != nil {
		return nil, err
	}
	u.Role = role
	return u, nil
}

func (f *fakeUsers) byID(id int) *repository.User {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u
		}
	}
	return nil
}

func (f *fakeUsers) activeAdmins() int {
	n := 0
	for _, u := range f.byEmail {
		if u.Role == repository.RoleAdmin && !u.Blocked {
			n++
		}
	}
	return n
}

func (f *fakeUsers) SetAccess(ctx context.Context, userID int, role string, blocked bool) error {
	u := f.byID(userID)
	if u == nil {
		return pgx.ErrNoRows
	}
	prevRole, prevBlocked := u.Role, u.Blocked
	u.Role, u.Blocked = role, blocked
	if f.activeAdmins() == 0 {
		u.Role, u.Blocked = prevRole, prevBlocked
		return repository.ErrLastAdmin
	}
	if blocked {
		return f.DeleteSessions(ctx, userID)
	}
	return nil
}

func (f *fakeUsers) SetPassword(_ context.Context, userID int, password string) error {
	if u := f.byID(userID); u != nil {
		u.PasswordHash = password
	}
	return nil
}

func (f *fakeUsers) DeleteSessions(_ context.Context, userID int) error {
	for token, id := range f.sessions {
		if id == userID {
			delete(f.sessions, token)
		}
	}
	return nil
}

func (f *fakeUsers) DeleteGuarded(ctx context.Context, userID int) error {
	u := f.byID(userID)
	if u == nil {
		return nil
	}
	delete(f.byEmail, u.Email)
	if f.activeAdmins() == 0 {
		f.byEmail[u.Email] = u
		return repository.ErrLastAdmin
	}
	return f.DeleteSessions(ctx, userID)
}
