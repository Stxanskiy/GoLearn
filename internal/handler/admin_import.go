package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/backendraz/golearn/internal/courseio"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

// AdminImportData powers the import/export screen.
type AdminImportData struct {
	PageTitle string
	Modules   []model.Module
}

func (h *Handler) AdminImportPage(w http.ResponseWriter, r *http.Request) {
	adminID := 0
	if u := GetUser(r.Context()); u != nil {
		adminID = u.ID
	}
	mods, _ := h.moduleRepo.GetForAdmin(r.Context(), adminID)
	h.render(w, "admin_import", &AdminImportData{PageTitle: "Импорт / Экспорт курса", Modules: mods})
}

// AdminModuleExport streams a module as a native course.json download.
func (h *Handler) AdminModuleExport(w http.ResponseWriter, r *http.Request) {
	id := atoiDefault(chi.URLParam(r, "id"), 0)
	tree, err := h.courseRepo.Export(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data, err := courseio.Marshal(courseio.FromTree(tree))
	if err != nil {
		http.Error(w, "Ошибка сериализации", 500)
		return
	}
	name := tree.Module.Slug
	if name == "" {
		name = "course"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`.course.json"`)
	_, _ = w.Write(data)
}

// importPayload reads the course JSON from an uploaded file or the textarea.
func (h *Handler) importPayload(r *http.Request) string {
	_ = r.ParseMultipartForm(16 << 20)
	if f, _, err := r.FormFile("file"); err == nil {
		defer f.Close()
		if data, err := io.ReadAll(io.LimitReader(f, 16<<20)); err == nil && len(data) > 0 {
			return string(data)
		}
	}
	return r.FormValue("json")
}

type importPreviewResp struct {
	OK       bool             `json:"ok"`
	Error    string           `json:"error,omitempty"`
	Slug     string           `json:"slug,omitempty"`
	Title    string           `json:"title,omitempty"`
	Exists   bool             `json:"exists"`
	Lessons  int              `json:"lessons"`
	New      []string         `json:"new,omitempty"`
	Updated  []string         `json:"updated,omitempty"`
	Removed  []string         `json:"removed,omitempty"`
	NewCount int              `json:"new_count"`
	UpdCount int              `json:"upd_count"`
	DelCount int              `json:"del_count"`
	Issues   []courseio.Issue `json:"issues,omitempty"`
	Blocked  bool             `json:"blocked"` // has error-level issues -> apply disabled
}

// AdminImportPreview parses the pasted/uploaded JSON and returns the diff (no write).
func (h *Handler) AdminImportPreview(w http.ResponseWriter, r *http.Request) {
	c, err := courseio.Parse([]byte(h.importPayload(r)))
	if err != nil {
		writeJSON(w, importPreviewResp{Error: "JSON не разобран: " + err.Error()})
		return
	}
	if c.Slug == "" || c.Title == "" {
		writeJSON(w, importPreviewResp{Error: "в курсе обязательны slug и title"})
		return
	}
	if !isValidSlug(c.Slug) {
		writeJSON(w, importPreviewResp{Error: "slug курса: только латиница, цифры, дефис"})
		return
	}
	d, err := h.courseRepo.Diff(r.Context(), c.ToTree())
	if err != nil {
		writeJSON(w, importPreviewResp{Error: "Ошибка сверки: " + err.Error()})
		return
	}
	issues := courseio.Validate(c)
	writeJSON(w, importPreviewResp{
		OK: true, Slug: d.Slug, Title: d.Title, Exists: d.Exists, Lessons: len(c.Lessons),
		New: d.New, Updated: d.Updated, Removed: d.Removed,
		NewCount: d.NewCount, UpdCount: d.UpdCount, DelCount: d.DelCount,
		Issues: issues, Blocked: courseio.HasErrors(issues),
	})
}

// AdminImportApply parses the JSON and upserts the whole course in one transaction.
func (h *Handler) AdminImportApply(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, err := courseio.Parse([]byte(h.importPayload(r)))
	if err != nil {
		http.Error(w, "JSON не разобран: "+err.Error(), 400)
		return
	}
	if c.Slug == "" || c.Title == "" {
		http.Error(w, "в курсе обязательны slug и title", 400)
		return
	}
	if !isValidSlug(c.Slug) {
		http.Error(w, "slug курса: только латиница, цифры, дефис", 400)
		return
	}
	if issues := courseio.Validate(c); courseio.HasErrors(issues) {
		msg := "Импорт заблокирован — исправь ошибки:"
		for _, is := range issues {
			if is.Level == "error" {
				msg += "\n• " + is.Msg
			}
		}
		http.Error(w, msg, 400)
		return
	}
	if _, err := h.courseRepo.Upsert(ctx, c.ToTree()); err != nil {
		h.log.Error("admin import course", "slug", c.Slug, "error", err)
		http.Error(w, "Ошибка импорта: "+err.Error(), 500)
		return
	}
	// land on the imported course's builder page
	if m, err := h.moduleRepo.GetBySlug(ctx, c.Slug); err == nil {
		http.Redirect(w, r, "/admin/module/"+strconv.Itoa(m.ID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
