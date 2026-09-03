package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type AdminSimsData struct {
	PageTitle string
	Sims      []model.Simulator
	Error     string
	Editing   string // slug being edited (empty = new)
	EditJSON  string
	EditPub   bool
}

// AdminSims lists simulators and, with ?edit=<slug>, loads one into the JSON editor.
func (h *Handler) AdminSims(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sims, _ := h.simRepo.List(ctx)
	d := &AdminSimsData{PageTitle: "Симуляторы", Sims: sims, Error: r.URL.Query().Get("err"), EditPub: true}
	if ed := r.URL.Query().Get("edit"); ed != "" {
		if s, err := h.simRepo.Get(ctx, ed); err == nil {
			d.Editing = ed
			d.EditJSON = prettyJSON(s.Data)
			d.EditPub = s.Published
		}
	}
	h.render(w, "admin_sims", d)
}

func prettyJSON(s string) string {
	var v any
	if json.Unmarshal([]byte(s), &v) != nil {
		return s
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s
	}
	return string(b)
}

func simBack(w http.ResponseWriter, r *http.Request, msg string) {
	u := "/admin/sims"
	if msg != "" {
		u += "?err=" + url.QueryEscape(msg)
	}
	http.Redirect(w, r, u, http.StatusSeeOther)
}

// AdminSimSave validates the scenario JSON and upserts the simulator by its slug.
func (h *Handler) AdminSimSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_ = r.ParseForm()
	var sc Scenario
	if json.Unmarshal([]byte(r.FormValue("data")), &sc) != nil {
		simBack(w, r, "JSON сценария не разобран")
		return
	}
	if !isValidSlug(sc.Slug) {
		simBack(w, r, "slug: только латиница, цифры, дефис")
		return
	}
	if sc.Title == "" || len(sc.Turns) == 0 {
		simBack(w, r, "нужны title и хотя бы один ход (turns)")
		return
	}
	norm, _ := json.Marshal(&sc)
	sim := model.Simulator{Slug: sc.Slug, Title: sc.Title, Icon: sc.Icon, Role: sc.Role, Data: string(norm)}

	cur, curErr := h.simRepo.Get(ctx, sc.Slug)
	if curErr == nil {
		sim.OrderNum = cur.OrderNum
		sim.OwnerID = cur.OwnerID
		if r.FormValue("has_published") != "" {
			sim.Published = r.FormValue("published") != ""
		} else {
			sim.Published = cur.Published
		}
	} else {
		n, _ := h.simRepo.Count(ctx)
		sim.OrderNum = n + 1
		sim.Published = r.FormValue("has_published") == "" || r.FormValue("published") != ""
		if u := GetUser(ctx); u != nil {
			sim.OwnerID = &u.ID
		}
	}
	if err := h.simRepo.Upsert(ctx, sim); err != nil {
		h.log.Error("admin save simulator", "slug", sc.Slug, "error", err)
		simBack(w, r, "Ошибка сохранения")
		return
	}
	simBack(w, r, "")
}

func (h *Handler) AdminSimDelete(w http.ResponseWriter, r *http.Request) {
	_ = h.simRepo.Delete(r.Context(), chi.URLParam(r, "slug"))
	simBack(w, r, "")
}

func (h *Handler) AdminSimPublish(w http.ResponseWriter, r *http.Request) {
	_ = h.simRepo.SetPublished(r.Context(), chi.URLParam(r, "slug"), r.FormValue("pub") == "1")
	simBack(w, r, "")
}

func (h *Handler) AdminSimMove(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")
	if dir == "" {
		dir = r.FormValue("dir")
	}
	_ = h.simRepo.Move(r.Context(), chi.URLParam(r, "slug"), dir)
	simBack(w, r, "")
}
