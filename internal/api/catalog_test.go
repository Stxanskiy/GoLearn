package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
	"github.com/backendraz/golearn/internal/repository"
)

const studentToken, adminToken = "student-token", "admin-token"

// storefront builds a small catalog: devops (published) and security (draft) specs, three courses, a trainer and a draft course.
func storefront() (*fakeUsers, *fakeContent) {
	users := newFakeUsers(
		&repository.User{ID: 1, Email: "s@example.com", Name: "Student", Role: "student"},
		&repository.User{ID: 2, Email: "a@example.com", Name: "Admin", Role: "admin"},
	)
	users.sessions[studentToken] = 1
	users.sessions[adminToken] = 2

	score := 4
	c := newFakeContent()
	c.specs = []model.Specialization{
		{Slug: "devops", Name: "DevOps", Icon: "♾️", Published: true},
		{Slug: "database", Name: "Базы данных", Published: true},
		{Slug: "security", Name: "Security", Published: false},
		{Slug: "gym", Name: "Тренажёры", Published: true},
	}
	c.modules = []model.Module{
		{ID: 10, Slug: "linux", Title: "Linux: Старт", Track: "devops", OrderNum: 1, Published: true, Difficulty: "beginner", Tags: []string{"bash"}},
		{ID: 11, Slug: "docker", Title: "Docker", Track: "devops", OrderNum: 2, Published: true, Difficulty: "intermediate", Description: "Containers"},
		{ID: 12, Slug: "k8s-draft", Title: "Kubernetes", Track: "devops", OrderNum: 3, Published: false},
		{ID: 13, Slug: "helm", Title: "Helm", Track: "devops", OrderNum: 4, Published: true},
		{ID: 20, Slug: "pentest", Title: "Пентест", Track: "security-offense", OrderNum: 1, Published: true},
		{ID: 30, Slug: "gym-linux", Title: "Linux тренажёр", Track: "gym", OrderNum: 1, Published: true},
	}
	c.lessons = []model.Lesson{
		{ID: 100, ModuleID: 10, Slug: "intro", Title: "Intro", Kind: "theory", OrderNum: 1, Published: true},
		{ID: 101, ModuleID: 10, Slug: "quiz", Title: "Quiz", Kind: "quiz", OrderNum: 2, Published: true},
		{ID: 102, ModuleID: 10, Slug: "lab", Title: "Lab", Kind: "lab", OrderNum: 3, Published: true},
		{ID: 103, ModuleID: 10, Slug: "draft", Title: "Draft", Kind: "theory", OrderNum: 4, Published: false},
		{ID: 110, ModuleID: 11, Slug: "images", Title: "Images", Kind: "", OrderNum: 1, Published: true},
		{ID: 111, ModuleID: 11, Slug: "compose", Title: "Compose", Kind: "lab", OrderNum: 2, Published: true},
		{ID: 120, ModuleID: 12, Slug: "pods", Title: "Pods", Kind: "theory", OrderNum: 1, Published: true},
		{ID: 300, ModuleID: 30, Slug: "gym-1", Title: "Gym 1", Kind: "lab", OrderNum: 1, Published: true},
	}
	c.progress[1] = []model.Progress{
		{LessonID: 100, Status: "completed"},
		{LessonID: 101, Status: "in_progress", QuizScore: &score},
		{LessonID: 110, Status: "in_progress"},
		{LessonID: 111, Status: "completed"}, // lab marked completed but tasks not passed
	}
	c.labPassed[1] = map[int]bool{102: true, 300: true}
	return users, c
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %T from %q: %v", v, w.Body.String(), err)
	}
	return v
}

func TestStorefrontRequiresSession(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)
	for _, path := range []string{"/catalog", "/me/profile", "/me/dashboard", "/courses/linux", "/specializations/devops", "/simulators"} {
		if w := do(h, http.MethodGet, path, ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s anonymous: %d", path, w.Code)
		}
	}
}

func TestCatalog(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)
	w := do(h, http.MethodGet, "/catalog", "", withCookie(studentToken))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if strings.Contains(w.Body.String(), `"tags":null`) {
		t.Error("tags must be an array, got null")
	}
	cat := decode[apigen.Catalog](t, w)

	var slugs []string
	for _, s := range cat.Specializations {
		slugs = append(slugs, s.Slug)
	}
	if strings.Join(slugs, ",") != "devops,database,gym" {
		t.Fatalf("specializations = %v (draft security must be hidden)", slugs)
	}

	devops := cat.Specializations[0]
	if len(devops.Courses) != 3 || devops.Courses[0].Slug != "linux" || devops.Courses[1].Slug != "docker" || devops.Courses[2].Slug != "helm" {
		t.Fatalf("devops courses = %+v", devops.Courses)
	}
	linux := devops.Courses[0]
	if linux.Status != "completed" || linux.LessonsCount != 3 || linux.LessonsCompleted != 3 || linux.ProgressPct != 100 {
		t.Errorf("linux card (quiz by score, lab by passed tasks, draft lesson excluded) = %+v", linux)
	}
	if devops.CoursesDone != 1 {
		t.Errorf("courses_done = %d", devops.CoursesDone)
	}
	docker := devops.Courses[1]
	if docker.Status != "in_progress" || docker.LessonsCompleted != 0 || docker.Label != "practice" || docker.Category != "Docker" || docker.EstMinutes != 20 {
		t.Errorf("docker card = %+v", docker)
	}
	if docker.CoverURL != "/api/v1/courses/docker/cover" || devops.CoverURL != "/api/v1/specializations/devops/cover" {
		t.Errorf("cover urls: %q %q", docker.CoverURL, devops.CoverURL)
	}
	if len(cat.Specializations[1].Courses) != 0 {
		t.Errorf("database courses = %+v", cat.Specializations[1].Courses)
	}
	if len(cat.Trainers) != 1 || cat.Trainers[0].Slug != "gym-linux" || cat.Trainers[0].Status != "completed" {
		t.Errorf("trainers = %+v", cat.Trainers)
	}
}

func TestGetSpecialization(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/specializations/devops", "", withCookie(studentToken))
	if w.Code != http.StatusOK || len(decode[apigen.SpecializationWithCourses](t, w).Courses) != 3 {
		t.Fatalf("devops: %d %s", w.Code, w.Body)
	}
	if w := do(h, http.MethodGet, "/specializations/security", "", withCookie(studentToken)); w.Code != http.StatusNotFound {
		t.Errorf("draft spec for student: %d", w.Code)
	}
	w = do(h, http.MethodGet, "/specializations/security", "", withCookie(adminToken))
	if w.Code != http.StatusOK || len(decode[apigen.SpecializationWithCourses](t, w).Courses) != 1 {
		t.Errorf("draft spec for admin: %d %s", w.Code, w.Body)
	}
	if w := do(h, http.MethodGet, "/specializations/nope", "", withCookie(studentToken)); w.Code != http.StatusNotFound {
		t.Errorf("unknown spec: %d", w.Code)
	}
}

func TestGetCourse(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/courses/docker", "", withCookie(studentToken))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	d := decode[apigen.CourseDetail](t, w)
	if d.Course.Description == nil || *d.Course.Description != "Containers" {
		t.Errorf("description = %v", d.Course.Description)
	}
	if len(d.Items) != 2 || d.CompletedCount != 0 || d.ProgressPct != 0 {
		t.Fatalf("detail = %+v", d)
	}
	first, lab := d.Items[0], d.Items[1]
	if first.Kind != apigen.LessonKindTheory || first.Status != "in_progress" || !first.IsNext {
		t.Errorf("first item = %+v", first)
	}
	if lab.Kind != "lab" || lab.Status != "in_progress" || lab.IsNext {
		t.Errorf("lab item (completed flag without passed tasks) = %+v", lab)
	}
	if d.PrevCourse == nil || d.PrevCourse.Slug != "linux" || d.NextCourse == nil || d.NextCourse.Slug != "helm" {
		t.Errorf("neighbours = %+v %+v", d.PrevCourse, d.NextCourse)
	}

	d = decode[apigen.CourseDetail](t, do(h, http.MethodGet, "/courses/linux", "", withCookie(studentToken)))
	for _, it := range d.Items {
		if it.IsNext {
			t.Errorf("completed course must have no next item: %+v", it)
		}
	}

	if w := do(h, http.MethodGet, "/courses/k8s-draft", "", withCookie(studentToken)); w.Code != http.StatusNotFound {
		t.Errorf("draft course for student: %d", w.Code)
	}
	if w := do(h, http.MethodGet, "/courses/k8s-draft", "", withCookie(adminToken)); w.Code != http.StatusOK {
		t.Errorf("draft course for admin: %d", w.Code)
	}
	if w := do(h, http.MethodGet, "/courses/nope", "", withCookie(studentToken)); w.Code != http.StatusNotFound {
		t.Errorf("unknown course: %d", w.Code)
	}
}

func TestLanding(t *testing.T) {
	users, content := storefront()
	content.stats = repository.PlatformStats{Courses: 4, Lessons: 7, Labs: 3, AutoChecked: 9}
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/public/landing", "")
	if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "public, max-age=300" {
		t.Fatalf("status %d cache %q", w.Code, w.Header().Get("Cache-Control"))
	}
	l := decode[apigen.Landing](t, w)
	if l.Stats.Courses != 4 || l.Stats.AutoCheckedTasks != 9 {
		t.Errorf("stats = %+v", l.Stats)
	}
	if len(l.Tracks) != 1 || l.Tracks[0].Slug != "devops" || l.Tracks[0].CoursesCount != 3 {
		t.Errorf("tracks (no empty, draft or gym) = %+v", l.Tracks)
	}
}

func TestCovers(t *testing.T) {
	users, content := storefront()
	content.modules[0].CoverImage = "data:image/png;base64,aGk="
	content.modules[1].CoverImage = "https://cdn.example.com/docker.jpg"
	content.modules[3].CoverImage = "data:text/html;base64,PHNjcmlwdD4="
	h := newTestAPIWith(t, users, content)

	tests := []struct {
		path, wantType string
		wantStatus     int
	}{
		{"/courses/linux/cover", "image/png", http.StatusOK},
		{"/courses/docker/cover", "", http.StatusFound},
		{"/courses/helm/cover", "image/svg+xml; charset=utf-8", http.StatusOK},
		{"/courses/pentest/cover", "image/svg+xml; charset=utf-8", http.StatusOK},
		{"/specializations/devops/cover", "image/svg+xml; charset=utf-8", http.StatusOK},
		{"/courses/nope/cover", "application/json", http.StatusNotFound},
	}
	for _, tt := range tests {
		w := do(h, http.MethodGet, tt.path, "")
		if w.Code != tt.wantStatus || (tt.wantType != "" && w.Header().Get("Content-Type") != tt.wantType) {
			t.Errorf("%s: %d %q", tt.path, w.Code, w.Header().Get("Content-Type"))
		}
	}
}

func TestProfile(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)
	p := decode[apigen.Profile](t, do(h, http.MethodGet, "/me/profile", "", withCookie(studentToken)))
	// completed: linux intro+quiz+lab, gym-1; started adds docker images+compose
	if p.User.ID != 1 || p.LessonsCompleted != 4 || p.LessonsStarted != 6 {
		t.Errorf("profile = %+v", p)
	}
}

func TestDashboard(t *testing.T) {
	users, content := storefront()
	today := time.Now().Format(time.DateOnly)
	content.overview = model.ProgressOverview{
		Streak: 2, TasksSolved: 5, ArticlesRead: 3, ArticlesTotal: 8,
		Activity: map[string]int{today: 3, "2001-01-01": 7},
	}
	content.cont = &repository.ContinueLesson{LessonID: 110, LessonSlug: "images", LessonTitle: "Images", LessonKind: "", ModuleSlug: "docker", ModuleTitle: "Docker"}
	content.sims = []model.Simulator{{Slug: "a", Published: true}, {Slug: "b", Published: false}}
	h := newTestAPIWith(t, users, content)

	d := decode[apigen.Dashboard](t, do(h, http.MethodGet, "/me/dashboard", "", withCookie(studentToken)))
	if d.Overview.Streak != 2 || d.Overview.LessonsTotal != 8 || d.Overview.SimulatorsTotal != 1 || d.Overview.TrainersTotal != 1 || d.Overview.TrainersDone != 1 {
		t.Errorf("overview = %+v", d.Overview)
	}
	if len(d.Activity) != 1 || d.Activity[today] != 3 {
		t.Errorf("activity = %v", d.Activity)
	}
	if d.Continue == nil || d.Continue.Course.Slug != "docker" || d.Continue.Kind != apigen.LessonKindTheory {
		t.Errorf("continue = %+v", d.Continue)
	}

	content.cont = nil
	w := do(h, http.MethodGet, "/me/dashboard", "", withCookie(studentToken))
	if strings.Contains(w.Body.String(), `"continue"`) {
		t.Errorf("continue must be omitted: %s", w.Body)
	}
}

func TestSimulators(t *testing.T) {
	users, content := storefront()
	content.sims = []model.Simulator{
		{Slug: "junior", Published: true, Data: `{"slug":"junior","title":"Junior","role":"Dev","icon":"🚀","intro":"Hi","metrics":[{"key":"budget","label":"B","unit":"k","start":10,"higher":true}],"turns":[{"title":"T1","situation":"S","choices":[{"text":"A","result":"R"}]}]}`},
		{Slug: "broken", Published: true, Data: `{not json`},
		{Slug: "draft", Published: false, Data: `{"slug":"draft","title":"Draft","turns":[]}`},
	}
	h := newTestAPIWith(t, users, content)

	list := decode[[]apigen.SimulatorSummary](t, do(h, http.MethodGet, "/simulators", "", withCookie(studentToken)))
	if len(list) != 1 || list[0].Slug != "junior" || list[0].TurnsCount != 1 {
		t.Fatalf("list = %+v", list)
	}

	w := do(h, http.MethodGet, "/simulators/junior", "", withCookie(studentToken))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"effects":{}`) {
		t.Fatalf("detail: %d %s", w.Code, w.Body)
	}
	for _, tt := range []struct {
		slug, token string
		want        int
	}{
		{"broken", adminToken, http.StatusNotFound},
		{"draft", studentToken, http.StatusNotFound},
		{"draft", adminToken, http.StatusOK},
		{"nope", studentToken, http.StatusNotFound},
	} {
		if w := do(h, http.MethodGet, "/simulators/"+tt.slug, "", withCookie(tt.token)); w.Code != tt.want {
			t.Errorf("%s as %s: %d want %d", tt.slug, tt.token, w.Code, tt.want)
		}
	}
}
