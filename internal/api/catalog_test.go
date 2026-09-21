package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
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
	}
	c.modules = []model.Module{
		{ID: 10, Slug: "linux", Title: "Linux: Старт", Track: "devops", OrderNum: 1, Published: true, Difficulty: "beginner", Tags: []string{"bash"}},
		{ID: 11, Slug: "docker", Title: "Docker", Track: "devops", OrderNum: 2, Published: true, Difficulty: "intermediate", Description: "Containers"},
		{ID: 12, Slug: "k8s-draft", Title: "Kubernetes", Track: "devops", OrderNum: 3, Published: false},
		{ID: 13, Slug: "helm", Title: "Helm", Track: "devops", OrderNum: 4, Published: true},
		{ID: 20, Slug: "pentest", Title: "Пентест", Track: "security-offense", OrderNum: 1, Published: true},
		{ID: 30, Slug: "gym-linux", Title: "Linux тренажёр", Track: "devops", IsTrainer: true, OrderNum: 1, Published: true},
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
	if strings.Join(slugs, ",") != "devops,database" {
		t.Fatalf("specializations = %v (draft security must be hidden)", slugs)
	}

	devops := cat.Specializations[0]
	// Cards come back in progress order: started, then untouched, then completed.
	if len(devops.Courses) != 3 || devops.Courses[0].Slug != "docker" || devops.Courses[1].Slug != "helm" || devops.Courses[2].Slug != "linux" {
		t.Fatalf("devops courses = %+v", devops.Courses)
	}
	linux := devops.Courses[2]
	if linux.Status != "completed" || linux.LessonsCount != 3 || linux.LessonsCompleted != 3 || linux.ProgressPct != 100 {
		t.Errorf("linux card (quiz by score, lab by passed tasks, draft lesson excluded) = %+v", linux)
	}
	if devops.CoursesDone != 1 {
		t.Errorf("courses_done = %d", devops.CoursesDone)
	}
	docker := devops.Courses[0]
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

func TestCoursePreviewIsPublic(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/public/courses/linux", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if got := w.Header().Get("Cache-Control"); got != "public, max-age=300" {
		t.Errorf("Cache-Control = %q", got)
	}
	p := decode[apigen.CoursePreview](t, w)
	if p.Title != "Linux: Старт" || p.CoverURL != "/api/v1/courses/linux/cover" {
		t.Errorf("course = %+v", p)
	}
	if p.LessonsCount != 3 || p.LabsCount != 1 {
		t.Errorf("counts: lessons %d, labs %d", p.LessonsCount, p.LabsCount)
	}
	if len(p.Lessons) != 3 {
		t.Fatalf("lessons = %+v", p.Lessons)
	}
	for _, lesson := range p.Lessons {
		if lesson.Slug == "draft" {
			t.Error("draft lesson is listed in the preview")
		}
	}
	if p.Lessons[0].Kind != apigen.LessonKindTheory || p.Lessons[2].Kind != apigen.LessonKindLab {
		t.Errorf("lesson kinds = %+v", p.Lessons)
	}
	if p.Specialization == nil || p.Specialization.Slug != "devops" {
		t.Errorf("specialization = %+v", p.Specialization)
	}
	if strings.Contains(w.Body.String(), `"tags":null`) {
		t.Error("tags must be an array, got null")
	}
}

func TestCoursePreviewHidesDraftsAndTracks(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	if w := do(h, http.MethodGet, "/public/courses/k8s-draft", ""); w.Code != http.StatusNotFound {
		t.Errorf("draft course: %d", w.Code)
	}
	if w := do(h, http.MethodGet, "/public/courses/nope", ""); w.Code != http.StatusNotFound {
		t.Errorf("missing course: %d", w.Code)
	}

	// A trainer keeps its specialization; a course of an unpublished one loses it.
	w := do(h, http.MethodGet, "/public/courses/gym-linux", "")
	if p := decode[apigen.CoursePreview](t, w); w.Code != http.StatusOK || p.Specialization == nil ||
		p.Specialization.Slug != "devops" {
		t.Errorf("trainer: %d specialization = %+v", w.Code, p.Specialization)
	}
	w = do(h, http.MethodGet, "/public/courses/pentest", "")
	if p := decode[apigen.CoursePreview](t, w); w.Code != http.StatusOK || p.Specialization != nil {
		t.Errorf("hidden specialization: %d %+v", w.Code, p.Specialization)
	}
}

func TestPublicCatalogHasNoProgress(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/public/catalog", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	if got := w.Header().Get("Cache-Control"); got != "public, max-age=300" {
		t.Errorf("Cache-Control = %q", got)
	}
	c := decode[apigen.Catalog](t, w)

	slugs := []string{}
	for _, spec := range c.Specializations {
		if spec.Slug == "security" {
			t.Error("draft specialization is public")
		}
		for _, course := range spec.Courses {
			slugs = append(slugs, course.Slug)
			if course.ProgressPct != 0 || course.LessonsCompleted != 0 {
				t.Errorf("%s leaks progress: %+v", course.Slug, course)
			}
			if course.Status != apigen.ProgressStatus(catalog.StatusNotStarted) {
				t.Errorf("%s status = %s", course.Slug, course.Status)
			}
		}
	}
	if strings.Contains(strings.Join(slugs, " "), "k8s-draft") {
		t.Error("draft course is public")
	}
}

func TestPublicSpecialization(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/public/specializations/devops", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	spec := decode[apigen.SpecializationWithCourses](t, w)
	if spec.Name != "DevOps" || spec.CoursesDone != 0 {
		t.Errorf("spec = %+v", spec)
	}
	if len(spec.Courses) != 3 {
		t.Errorf("courses = %d", len(spec.Courses))
	}

	if w := do(h, http.MethodGet, "/public/specializations/security", ""); w.Code != http.StatusNotFound {
		t.Errorf("draft specialization: %d", w.Code)
	}
	if w := do(h, http.MethodGet, "/public/specializations/nope", ""); w.Code != http.StatusNotFound {
		t.Errorf("missing specialization: %d", w.Code)
	}
}

// landingStorefront adds a second published track so the landing has something to sort.
func landingStorefront() (*fakeUsers, *fakeContent) {
	users, content := storefront()
	content.modules = append(content.modules,
		model.Module{ID: 40, Slug: "sql-basics", Title: "SQL", Track: "database", OrderNum: 1, Published: true},
		model.Module{ID: 41, Slug: "sql-deep", Title: "SQL глубже", Track: "database", OrderNum: 2, Published: true},
	)
	content.lessons = append(content.lessons,
		model.Lesson{ID: 400, ModuleID: 40, Slug: "select", Title: "SELECT", Kind: "theory", OrderNum: 1, Published: true},
	)
	return users, content
}

func trackSlugs(tracks []apigen.LandingTrack) []string {
	slugs := make([]string, 0, len(tracks))
	for _, track := range tracks {
		slugs = append(slugs, track.Slug)
	}
	return slugs
}

func TestLandingCountsAndOrder(t *testing.T) {
	users, content := landingStorefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/public/landing", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	tracks := decode[apigen.Landing](t, w).Tracks
	if got := trackSlugs(tracks); !slices.Equal(got, []string{"devops", "database"}) {
		t.Fatalf("default order = %v", got)
	}
	// devops: linux, docker, helm with 5 published lessons; database: two courses with one.
	if tracks[0].CoursesCount != 3 || tracks[0].LessonsCount != 5 {
		t.Errorf("devops counts = %+v", tracks[0])
	}
	if tracks[1].CoursesCount != 2 || tracks[1].LessonsCount != 1 {
		t.Errorf("database counts = %+v", tracks[1])
	}
}

func TestLandingSortAndLimit(t *testing.T) {
	users, content := landingStorefront()
	h := newTestAPIWith(t, users, content)

	cases := []struct {
		query string
		want  []string
	}{
		{"?sort=lessons", []string{"database", "devops"}},
		{"?sort=lessons&dir=desc", []string{"devops", "database"}},
		{"?sort=courses", []string{"database", "devops"}},
		{"?sort=courses&dir=desc", []string{"devops", "database"}},
		{"?dir=desc", []string{"database", "devops"}},
		{"?sort=lessons&limit=1", []string{"database"}},
		{"?limit=1", []string{"devops"}},
	}
	for _, c := range cases {
		w := do(h, http.MethodGet, "/public/landing"+c.query, "")
		if w.Code != http.StatusOK {
			t.Errorf("%s: status %d", c.query, w.Code)
			continue
		}
		if got := trackSlugs(decode[apigen.Landing](t, w).Tracks); !slices.Equal(got, c.want) {
			t.Errorf("%s: tracks = %v, want %v", c.query, got, c.want)
		}
	}
}

func TestLandingRejectsBadQuery(t *testing.T) {
	users, content := landingStorefront()
	h := newTestAPIWith(t, users, content)

	for _, query := range []string{"?sort=price", "?dir=sideways", "?limit=0", "?limit=51", "?limit=many"} {
		if w := do(h, http.MethodGet, "/public/landing"+query, ""); w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d", query, w.Code)
		}
	}
}

func TestTrainerStandsOutsideTheCoursePath(t *testing.T) {
	users, content := storefront()
	h := newTestAPIWith(t, users, content)

	w := do(h, http.MethodGet, "/courses/gym-linux", "", withCookie(studentToken))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	trainer := decode[apigen.CourseDetail](t, w)
	if !trainer.Course.IsTrainer || trainer.PrevCourse != nil || trainer.NextCourse != nil {
		t.Errorf("trainer = %+v, prev %+v, next %+v", trainer.Course.IsTrainer, trainer.PrevCourse, trainer.NextCourse)
	}
	if len(trainer.Items) != 1 || trainer.Items[0].Status != "completed" {
		t.Errorf("trainer keeps its lessons and progress: %+v", trainer.Items)
	}

	// A regular course never points at a trainer as its neighbour.
	w = do(h, http.MethodGet, "/courses/docker", "", withCookie(studentToken))
	course := decode[apigen.CourseDetail](t, w)
	for _, ref := range []*apigen.LinkRef{course.PrevCourse, course.NextCourse} {
		if ref != nil && ref.Slug == "gym-linux" {
			t.Errorf("neighbour is a trainer: %+v", ref)
		}
	}
}
