package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/catalog"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *API) adminListSpecializations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	specs, err := a.Specs.List(ctx)
	if err != nil {
		a.internalError(w, "admin: list specializations", err)
		return
	}
	counts, err := a.specCourseCounts(r)
	if err != nil {
		a.internalError(w, "admin: specialization courses", err)
		return
	}
	out := make([]apigen.AdminSpecialization, 0, len(specs))
	for _, s := range specs {
		out = append(out, toAdminSpecialization(s, counts[s.Slug]))
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) adminCreateSpecialization(w http.ResponseWriter, r *http.Request) {
	var body apigen.AdminCreateSpecializationJSONRequestBody
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	s, fields := specFromInput(apigen.AdminSpecializationInput{
		Name: body.Name, Icon: body.Icon, Description: body.Description, CoverURL: body.CoverURL,
	})
	s.Slug = strings.TrimSpace(body.Slug)
	checkSlug(fields, "slug", s.Slug)
	if len(s.Slug) > 40 {
		fields["slug"] = fieldTooLong
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	if _, err := a.Specs.Get(ctx, s.Slug); err == nil {
		writeError(w, http.StatusConflict, codeSlugTaken, "specialization slug is taken")
		return
	} else if !isNotFound(err) {
		a.internalError(w, "admin: find specialization", err)
		return
	}
	var err error
	if s.OrderNum, err = a.Specs.NextOrder(ctx); err != nil {
		a.internalError(w, "admin: specialization order", err)
		return
	}
	s.Published = deref(body.Published)
	s.OwnerID = &userFrom(ctx).ID
	if err := a.Specs.Upsert(ctx, s); err != nil {
		a.internalError(w, "admin: create specialization", err)
		return
	}
	a.writeAdminSpecialization(w, r, http.StatusCreated, s.Slug)
}

func (a *API) adminUpdateSpecialization(w http.ResponseWriter, r *http.Request) {
	cur, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	var body apigen.AdminSpecializationInput
	if !decodeJSON(w, r, &body) {
		return
	}
	s, fields := specFromInput(body)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	s.Slug, s.OrderNum, s.OwnerID, s.Published = cur.Slug, cur.OrderNum, cur.OwnerID, cur.Published
	if body.Published != nil {
		s.Published = *body.Published
	}
	if err := a.Specs.Upsert(r.Context(), s); err != nil {
		a.internalError(w, "admin: update specialization", err)
		return
	}
	a.writeAdminSpecialization(w, r, http.StatusOK, s.Slug)
}

func (a *API) adminDeleteSpecialization(w http.ResponseWriter, r *http.Request) {
	s, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	counts, err := a.specTrackUsage(r)
	if err != nil {
		a.internalError(w, "admin: specialization courses", err)
		return
	}
	if counts[s.Slug] > 0 {
		writeError(w, http.StatusConflict, codeSpecNotEmpty, "specialization still has courses")
		return
	}
	if err := a.Specs.Delete(r.Context(), s.Slug); err != nil {
		a.internalError(w, "admin: delete specialization", err)
		return
	}
	a.dropStored(r, s.IconURL)
	a.dropStored(r, s.CoverImage)
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminSetSpecializationPublished(w http.ResponseWriter, r *http.Request) {
	s, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	var body apigen.Published
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := a.Specs.SetPublished(r.Context(), s.Slug, body.Published); err != nil {
		a.internalError(w, "admin: publish specialization", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminMoveSpecialization(w http.ResponseWriter, r *http.Request) {
	s, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	dir, ok := decodeMove(w, r)
	if !ok {
		return
	}
	if err := a.Specs.Move(r.Context(), s.Slug, dir); err != nil {
		a.internalError(w, "admin: move specialization", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminUploadSpecializationCover(w http.ResponseWriter, r *http.Request) {
	s, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	cover, ok := a.storeCover(w, r)
	if !ok {
		return
	}
	a.dropStored(r, s.CoverImage)
	if err := a.Specs.SetCover(r.Context(), s.Slug, cover); err != nil {
		a.internalError(w, "admin: upload specialization cover", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminDeleteSpecializationCover(w http.ResponseWriter, r *http.Request) {
	s, ok := a.specBySlug(w, r)
	if !ok {
		return
	}
	a.dropStored(r, s.CoverImage)
	if err := a.Specs.SetCover(r.Context(), s.Slug, ""); err != nil {
		a.internalError(w, "admin: delete specialization cover", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminListSimulators(w http.ResponseWriter, r *http.Request) {
	sims, err := a.Sims.List(r.Context())
	if err != nil {
		a.internalError(w, "admin: list simulators", err)
		return
	}
	out := make([]apigen.AdminSimulatorRow, 0, len(sims))
	for _, s := range sims {
		out = append(out, toAdminSimulatorRow(s))
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) adminGetSimulator(w http.ResponseWriter, r *http.Request) {
	s, ok := a.simBySlug(w, r)
	if !ok {
		return
	}
	a.writeAdminSimulator(w, http.StatusOK, *s)
}

func (a *API) adminCreateSimulator(w http.ResponseWriter, r *http.Request) {
	var body apigen.AdminSimulatorInput
	if !decodeJSON(w, r, &body) {
		return
	}
	sim, fields := simFromInput(body.Scenario)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	ctx := r.Context()
	if _, err := a.Sims.Get(ctx, sim.Slug); err == nil {
		writeError(w, http.StatusConflict, codeSlugTaken, "simulator slug is taken")
		return
	} else if !isNotFound(err) {
		a.internalError(w, "admin: find simulator", err)
		return
	}
	n, err := a.Sims.Count(ctx)
	if err != nil {
		a.internalError(w, "admin: count simulators", err)
		return
	}
	sim.OrderNum, sim.Published, sim.OwnerID = n+1, deref(body.Published), &userFrom(ctx).ID
	if err := a.Sims.Upsert(ctx, sim); err != nil {
		a.internalError(w, "admin: create simulator", err)
		return
	}
	a.writeAdminSimulator(w, http.StatusCreated, sim)
}

func (a *API) adminUpdateSimulator(w http.ResponseWriter, r *http.Request) {
	cur, ok := a.simBySlug(w, r)
	if !ok {
		return
	}
	var body apigen.AdminSimulatorInput
	if !decodeJSON(w, r, &body) {
		return
	}
	sim, fields := simFromInput(body.Scenario)
	if sim.Slug != cur.Slug && fields["scenario.slug"] == "" {
		fields["scenario.slug"] = fieldInvalidValue
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	sim.OrderNum, sim.OwnerID, sim.Published = cur.OrderNum, cur.OwnerID, cur.Published
	if body.Published != nil {
		sim.Published = *body.Published
	}
	if err := a.Sims.Upsert(r.Context(), sim); err != nil {
		a.internalError(w, "admin: update simulator", err)
		return
	}
	a.writeAdminSimulator(w, http.StatusOK, sim)
}

func (a *API) adminDeleteSimulator(w http.ResponseWriter, r *http.Request) {
	s, ok := a.simBySlug(w, r)
	if !ok {
		return
	}
	if err := a.Sims.Delete(r.Context(), s.Slug); err != nil {
		a.internalError(w, "admin: delete simulator", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminSetSimulatorPublished(w http.ResponseWriter, r *http.Request) {
	s, ok := a.simBySlug(w, r)
	if !ok {
		return
	}
	var body apigen.Published
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := a.Sims.SetPublished(r.Context(), s.Slug, body.Published); err != nil {
		a.internalError(w, "admin: publish simulator", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) adminMoveSimulator(w http.ResponseWriter, r *http.Request) {
	s, ok := a.simBySlug(w, r)
	if !ok {
		return
	}
	dir, ok := decodeMove(w, r)
	if !ok {
		return
	}
	if err := a.Sims.Move(r.Context(), s.Slug, dir); err != nil {
		a.internalError(w, "admin: move simulator", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// specBySlug loads the specSlug specialization, writing 404/500 itself when it returns false.
func (a *API) specBySlug(w http.ResponseWriter, r *http.Request) (*model.Specialization, bool) {
	s, err := a.Specs.Get(r.Context(), chi.URLParam(r, "specSlug"))
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "specialization not found")
		return nil, false
	}
	if err != nil {
		a.internalError(w, "admin: load specialization", err)
		return nil, false
	}
	return s, true
}

// simBySlug loads the simSlug simulator, writing 404/500 itself when it returns false.
func (a *API) simBySlug(w http.ResponseWriter, r *http.Request) (*model.Simulator, bool) {
	s, err := a.Sims.Get(r.Context(), chi.URLParam(r, "simSlug"))
	if isNotFound(err) {
		writeError(w, http.StatusNotFound, codeNotFound, "simulator not found")
		return nil, false
	}
	if err != nil {
		a.internalError(w, "admin: load simulator", err)
		return nil, false
	}
	return s, true
}

// specCourseCounts returns the number of courses per specialization slug; trainers are not courses.
func (a *API) specCourseCounts(r *http.Request) (map[string]int, error) {
	tracks, err := a.Modules.TrackCounts(r.Context())
	if err != nil {
		return nil, err
	}
	return bySpec(tracks), nil
}

// specTrackUsage returns the number of modules per specialization slug, trainers included;
// it guards the deletion of a specialization against leaving them orphaned.
func (a *API) specTrackUsage(r *http.Request) (map[string]int, error) {
	tracks, err := a.Modules.TrackUsage(r.Context())
	if err != nil {
		return nil, err
	}
	return bySpec(tracks), nil
}

// bySpec folds per-track counts onto their specialization slug.
func bySpec(tracks map[string]int) map[string]int {
	out := map[string]int{}
	for track, n := range tracks {
		out[catalog.SpecForTrack(track)] += n
	}
	return out
}

func (a *API) writeAdminSpecialization(w http.ResponseWriter, r *http.Request, status int, slug string) {
	s, err := a.Specs.Get(r.Context(), slug)
	if err != nil {
		a.internalError(w, "admin: reload specialization", err)
		return
	}
	counts, err := a.specCourseCounts(r)
	if err != nil {
		a.internalError(w, "admin: specialization courses", err)
		return
	}
	writeJSON(w, status, toAdminSpecialization(*s, counts[slug]))
}

func (a *API) writeAdminSimulator(w http.ResponseWriter, status int, s model.Simulator) {
	sc, ok := parseScenario(s)
	if !ok {
		a.internalError(w, "admin: parse simulator", nil)
		return
	}
	writeJSON(w, status, apigen.AdminSimulator{
		Slug: s.Slug, Title: s.Title, Icon: s.Icon, Role: s.Role, OrderNum: s.OrderNum,
		Published: s.Published, OwnerID: s.OwnerID, Scenario: sc,
	})
}

func specFromInput(in apigen.AdminSpecializationInput) (model.Specialization, map[string]string) {
	fields := map[string]string{}
	s := model.Specialization{
		Name:        strings.TrimSpace(in.Name),
		Icon:        strings.TrimSpace(deref(in.Icon)),
		Description: deref(in.Description),
		CoverImage:  deref(in.CoverURL),
	}
	checkText(fields, "name", s.Name, 80, true)
	checkText(fields, "icon", s.Icon, 16, false)
	checkText(fields, "description", s.Description, 5000, false)
	if in.CoverURL != nil && !isHTTPURL(*in.CoverURL) {
		fields["cover_url"] = fieldInvalidFormat
	}
	return s, fields
}

// simFromInput validates a scenario and returns the simulator row it is stored as.
func simFromInput(sc apigen.Scenario) (model.Simulator, map[string]string) {
	fields := map[string]string{}
	sc.Slug, sc.Title = strings.TrimSpace(sc.Slug), strings.TrimSpace(sc.Title)
	checkSlug(fields, "scenario.slug", sc.Slug)
	if len(sc.Slug) > 60 {
		fields["scenario.slug"] = fieldTooLong
	}
	checkText(fields, "scenario.title", sc.Title, 200, true)
	checkText(fields, "scenario.role", sc.Role, 160, false)
	checkText(fields, "scenario.icon", sc.Icon, 16, false)
	metrics := map[string]bool{}
	for _, m := range sc.Metrics {
		if m.Key == "" || metrics[m.Key] {
			fields["scenario.metrics"] = fieldInvalidValue
		}
		metrics[m.Key] = true
	}
	switch {
	case len(sc.Turns) == 0:
		fields["scenario.turns"] = fieldTooShort
	case len(sc.Turns) > 100:
		fields["scenario.turns"] = fieldTooLong
	}
	for _, t := range sc.Turns {
		if len(t.Choices) == 0 {
			fields["scenario.turns"] = fieldInvalidValue
		}
		for _, c := range t.Choices {
			for key := range c.Effects {
				if !metrics[key] {
					fields["scenario.turns"] = fieldInvalidValue
				}
			}
		}
	}
	data, _ := json.Marshal(sc)
	return model.Simulator{Slug: sc.Slug, Title: sc.Title, Icon: sc.Icon, Role: sc.Role, Data: string(data)}, fields
}

func toAdminSpecialization(s model.Specialization, courses int) apigen.AdminSpecialization {
	out := apigen.AdminSpecialization{
		Slug: s.Slug, Name: s.Name, Icon: s.Icon, IconURL: s.IconURL, Description: s.Description, Published: s.Published,
		OrderNum: s.OrderNum, OwnerID: s.OwnerID, CoursesCount: courses,
		HasCustomCover:  s.CoverImage != "",
		CoverPreviewURL: "/api/v1/specializations/" + s.Slug + "/cover",
	}
	if s.CoverImage != "" && !strings.HasPrefix(s.CoverImage, "data:") {
		out.CoverURL = &s.CoverImage
	}
	return out
}

func toAdminSimulatorRow(s model.Simulator) apigen.AdminSimulatorRow {
	return apigen.AdminSimulatorRow{
		Slug: s.Slug, Title: s.Title, Icon: s.Icon, Role: s.Role, OrderNum: s.OrderNum,
		Published: s.Published, OwnerID: s.OwnerID,
	}
}
