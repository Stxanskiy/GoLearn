package handler

import (
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

// AdminMiddleware allows only users with role=admin.
func (h *Handler) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := GetUser(r.Context())
		if u == nil || !u.IsAdmin() {
			http.Error(w, "403 — доступ только для администратора", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func atoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return v
	}
	return def
}

func splitTags(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// coverFromForm returns a cover value from an uploaded file (as data URI) or a
// pasted URL; empty string means "leave unchanged".
func (h *Handler) coverFromForm(r *http.Request) string {
	if url := strings.TrimSpace(r.FormValue("cover_url")); url != "" {
		return url
	}
	file, hdr, err := r.FormFile("cover_file")
	if err != nil {
		return ""
	}
	defer file.Close()
	const maxCover = 4 << 20 // 4MB
	data, err := io.ReadAll(io.LimitReader(file, maxCover+1))
	if err != nil || len(data) == 0 || len(data) > maxCover {
		return ""
	}
	ct := hdr.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		ct = "image/png"
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(data)
}

// ── Dashboard ──

type AdminModuleRow struct {
	model.Module
	Lessons int
	Source  string
}

type AdminDashData struct {
	PageTitle string
	Specs     []model.Specialization
	Modules   []AdminModuleRow
}

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	specs, _ := h.specRepo.List(ctx)
	mods, _ := h.moduleRepo.GetAll(ctx)
	var rows []AdminModuleRow
	for _, m := range mods {
		lessons, _ := h.lessonRepo.GetByModule(ctx, m.ID)
		rows = append(rows, AdminModuleRow{Module: m, Lessons: len(lessons)})
	}
	h.render(w, "admin_dashboard", &AdminDashData{PageTitle: "Админка — TOT", Specs: specs, Modules: rows})
}

// ── Module form ──

type AdminModuleData struct {
	PageTitle string
	Module    *model.Module
	Specs     []model.Specialization
	Lessons   []model.Lesson
	IsNew     bool
}

func (h *Handler) AdminModuleNew(w http.ResponseWriter, r *http.Request) {
	specs, _ := h.specRepo.List(r.Context())
	h.render(w, "admin_module", &AdminModuleData{PageTitle: "Новый курс", Module: &model.Module{Difficulty: "beginner"}, Specs: specs, IsNew: true})
}

func (h *Handler) AdminModuleEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	mod, err := h.moduleRepo.GetByID(ctx, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	specs, _ := h.specRepo.List(ctx)
	lessons, _ := h.lessonRepo.GetByModule(ctx, id)
	h.render(w, "admin_module", &AdminModuleData{PageTitle: "Курс: " + mod.Title, Module: mod, Specs: specs, Lessons: lessons})
}

func (h *Handler) AdminModuleSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseMultipartForm(8 << 20)
	id := atoiDefault(chi.URLParam(r, "id"), 0)

	m := model.Module{
		ID:          id,
		Slug:        strings.TrimSpace(r.FormValue("slug")),
		Title:       strings.TrimSpace(r.FormValue("title")),
		Description: r.FormValue("description"),
		Track:       r.FormValue("track"),
		Difficulty:  r.FormValue("difficulty"),
		Category:    strings.TrimSpace(r.FormValue("category")),
		Label:       strings.TrimSpace(r.FormValue("label")),
		Tags:        splitTags(r.FormValue("tags")),
		Accent:      strings.TrimSpace(r.FormValue("accent")),
		OrderNum:    atoiDefault(r.FormValue("order_num"), 0),
		EstMinutes:  atoiDefault(r.FormValue("est_minutes"), 0),
		CoverImage:  h.coverFromForm(r),
	}
	if m.Slug == "" || m.Title == "" {
		http.Error(w, "slug и title обязательны", 400)
		return
	}

	if id == 0 {
		if _, err := h.moduleRepo.Create(ctx, m); err != nil {
			h.log.Error("admin create module", "error", err)
			http.Error(w, "Ошибка создания: "+err.Error(), 500)
			return
		}
	} else {
		if m.CoverImage == "" { // keep existing cover when none provided
			if cur, err := h.moduleRepo.GetByID(ctx, id); err == nil {
				m.CoverImage = cur.CoverImage
			}
		}
		if err := h.moduleRepo.Update(ctx, m); err != nil {
			h.log.Error("admin update module", "error", err)
			http.Error(w, "Ошибка сохранения: "+err.Error(), 500)
			return
		}
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) AdminModuleDelete(w http.ResponseWriter, r *http.Request) {
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	if err := h.moduleRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, "Ошибка удаления", 500)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// ── Lesson form ──

type AdminLessonData struct {
	PageTitle string
	Lesson    *model.Lesson
	ModuleID  int
	IsNew     bool
	Questions []model.QuizQuestion
	Tasks     []model.Task
}

func (h *Handler) AdminLessonNew(w http.ResponseWriter, r *http.Request) {
	mid := atoiDefault(chi.URLParam(r, "id"), 0)
	h.render(w, "admin_lesson", &AdminLessonData{PageTitle: "Новая глава", Lesson: &model.Lesson{Kind: "theory", Difficulty: "beginner", Format: "md"}, ModuleID: mid, IsNew: true})
}

// AdminPreview renders markdown/html content the same way the lesson page does.
func (h *Handler) AdminPreview(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	html := RenderContent(r.FormValue("format"), r.FormValue("content"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (h *Handler) AdminLessonEdit(w http.ResponseWriter, r *http.Request) {
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	l, err := h.lessonRepo.GetByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, questions, _ := h.lessonRepo.GetQuiz(r.Context(), id)
	tasks, _ := h.lessonRepo.GetTasks(r.Context(), id)
	h.render(w, "admin_lesson", &AdminLessonData{
		PageTitle: "Глава: " + l.Title, Lesson: l, ModuleID: l.ModuleID,
		Questions: questions, Tasks: tasks,
	})
}

func (h *Handler) AdminLessonSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	mid := atoiDefault(r.FormValue("module_id"), 0)

	l := model.Lesson{
		ID:         id,
		ModuleID:   mid,
		Slug:       strings.TrimSpace(r.FormValue("slug")),
		Title:      strings.TrimSpace(r.FormValue("title")),
		Content:    r.FormValue("content"),
		Kind:       r.FormValue("kind"),
		Format:     r.FormValue("format"),
		Difficulty: r.FormValue("difficulty"),
		OrderNum:   atoiDefault(r.FormValue("order_num"), 0),
		VMImage:    strings.TrimSpace(r.FormValue("vm_image")),
		VMInit:     r.FormValue("vm_init"),
	}
	if l.Format == "" {
		l.Format = "html"
	}
	if l.Slug == "" || l.Title == "" {
		http.Error(w, "slug и title обязательны", 400)
		return
	}

	var redirectID int
	if id == 0 {
		if _, err := h.lessonRepo.Create(ctx, l); err != nil {
			http.Error(w, "Ошибка создания: "+err.Error(), 500)
			return
		}
		redirectID = mid
	} else {
		if err := h.lessonRepo.Update(ctx, l); err != nil {
			http.Error(w, "Ошибка сохранения: "+err.Error(), 500)
			return
		}
		redirectID = l.ModuleID
	}
	http.Redirect(w, r, "/admin/module/"+strconv.Itoa(redirectID), http.StatusSeeOther)
}

func (h *Handler) AdminLessonDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	l, _ := h.lessonRepo.GetByID(ctx, id)
	_ = h.lessonRepo.Delete(ctx, id)
	mid := 0
	if l != nil {
		mid = l.ModuleID
	}
	http.Redirect(w, r, "/admin/module/"+strconv.Itoa(mid), http.StatusSeeOther)
}

// ── Specializations ──

type AdminSpecsData struct {
	PageTitle string
	Specs     []model.Specialization
}

func (h *Handler) AdminSpecs(w http.ResponseWriter, r *http.Request) {
	specs, _ := h.specRepo.List(r.Context())
	h.render(w, "admin_specs", &AdminSpecsData{PageTitle: "Разделы", Specs: specs})
}

func (h *Handler) AdminSpecSave(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	s := model.Specialization{
		Slug:        strings.TrimSpace(r.FormValue("slug")),
		Name:        strings.TrimSpace(r.FormValue("name")),
		Icon:        strings.TrimSpace(r.FormValue("icon")),
		Description: r.FormValue("description"),
		OrderNum:    atoiDefault(r.FormValue("order_num"), 0),
	}
	if s.Slug == "" || s.Name == "" {
		http.Error(w, "slug и name обязательны", 400)
		return
	}
	if err := h.specRepo.Upsert(r.Context(), s); err != nil {
		http.Error(w, "Ошибка: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/admin/specs", http.StatusSeeOther)
}

func (h *Handler) AdminSpecDelete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := h.specRepo.Delete(r.Context(), slug); err != nil {
		http.Error(w, "Ошибка удаления", 500)
		return
	}
	http.Redirect(w, r, "/admin/specs", http.StatusSeeOther)
}

// ── Quiz questions ──

func linesToSlice(s string) []string {
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if ln = strings.TrimRight(ln, "\r"); strings.TrimSpace(ln) != "" {
			out = append(out, strings.TrimSpace(ln))
		}
	}
	return out
}

type AdminQuestionData struct {
	PageTitle string
	Question  *model.QuizQuestion
	LessonID  int
	IsNew     bool
}

func (h *Handler) AdminQuestionNew(w http.ResponseWriter, r *http.Request) {
	lid := atoiDefault(chi.URLParam(r, "id"), 0)
	h.render(w, "admin_question", &AdminQuestionData{PageTitle: "Новый вопрос", Question: &model.QuizQuestion{}, LessonID: lid, IsNew: true})
}

func (h *Handler) AdminQuestionEdit(w http.ResponseWriter, r *http.Request) {
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	q, err := h.lessonRepo.GetQuestionByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	lid, _ := h.lessonRepo.LessonIDForQuiz(r.Context(), q.QuizID)
	h.render(w, "admin_question", &AdminQuestionData{PageTitle: "Вопрос", Question: q, LessonID: lid})
}

func (h *Handler) AdminQuestionSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	lid := atoiDefault(r.FormValue("lesson_id"), 0)
	q := model.QuizQuestion{
		ID:           id,
		Question:     r.FormValue("question"),
		Options:      linesToSlice(r.FormValue("options")),
		CorrectIndex: atoiDefault(r.FormValue("correct"), 1) - 1, // 1-based in UI
		Explanation:  r.FormValue("explanation"),
	}
	if q.Question == "" || len(q.Options) < 2 {
		http.Error(w, "нужен вопрос и минимум 2 варианта", 400)
		return
	}
	if id == 0 {
		quizID, err := h.lessonRepo.EnsureQuiz(ctx, lid, "Квиз")
		if err != nil {
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}
		if err := h.lessonRepo.AddQuestion(ctx, quizID, q); err != nil {
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}
	} else {
		if err := h.lessonRepo.UpdateQuestion(ctx, q); err != nil {
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}
	}
	http.Redirect(w, r, "/admin/lesson/"+strconv.Itoa(lid), http.StatusSeeOther)
}

func (h *Handler) AdminQuestionDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	lid := atoiDefault(r.FormValue("lesson_id"), 0)
	_ = h.lessonRepo.DeleteQuestion(ctx, id)
	http.Redirect(w, r, "/admin/lesson/"+strconv.Itoa(lid), http.StatusSeeOther)
}

// ── Tasks ──

type AdminTaskData struct {
	PageTitle string
	Task      *model.Task
	LessonID  int
	IsNew     bool
}

func (h *Handler) AdminTaskNew(w http.ResponseWriter, r *http.Request) {
	lid := atoiDefault(chi.URLParam(r, "id"), 0)
	h.render(w, "admin_task", &AdminTaskData{PageTitle: "Новое задание", Task: &model.Task{Kind: "shell", Difficulty: "medium", Format: "md"}, LessonID: lid, IsNew: true})
}

func (h *Handler) AdminTaskEdit(w http.ResponseWriter, r *http.Request) {
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	t, err := h.lessonRepo.GetTaskByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "admin_task", &AdminTaskData{PageTitle: "Задание: " + t.Title, Task: t, LessonID: t.LessonID})
}

func (h *Handler) AdminTaskSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	lid := atoiDefault(r.FormValue("lesson_id"), 0)
	t := model.Task{
		ID:           id,
		LessonID:     lid,
		Title:        strings.TrimSpace(r.FormValue("title")),
		Description:  r.FormValue("description"),
		Hints:        r.FormValue("hints"),
		Solution:     r.FormValue("solution"),
		Difficulty:   r.FormValue("difficulty"),
		Kind:         r.FormValue("kind"),
		Format:       r.FormValue("format"),
		StarterCode:  r.FormValue("starter_code"),
		SandboxImage: strings.TrimSpace(r.FormValue("sandbox_image")),
		SetupScript:  r.FormValue("setup_script"),
		CheckScript:  r.FormValue("check_script"),
	}
	if t.Title == "" {
		http.Error(w, "нужен заголовок задания", 400)
		return
	}
	if id == 0 {
		if _, err := h.lessonRepo.CreateTask(ctx, t); err != nil {
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}
	} else {
		if err := h.lessonRepo.UpdateTask(ctx, t); err != nil {
			http.Error(w, "Ошибка: "+err.Error(), 500)
			return
		}
	}
	http.Redirect(w, r, "/admin/lesson/"+strconv.Itoa(lid), http.StatusSeeOther)
}

func (h *Handler) AdminTaskDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	lid := atoiDefault(r.FormValue("lesson_id"), 0)
	_ = h.lessonRepo.DeleteTask(ctx, id)
	http.Redirect(w, r, "/admin/lesson/"+strconv.Itoa(lid), http.StatusSeeOther)
}
