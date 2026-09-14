package api

import (
	"encoding/json"
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

func (a *API) listSimulators(w http.ResponseWriter, r *http.Request) {
	sims, err := a.Sims.ListPublished(r.Context())
	if err != nil {
		a.internalError(w, "simulators: list", err)
		return
	}
	out := []apigen.SimulatorSummary{}
	for _, s := range sims {
		sc, ok := parseScenario(s)
		if !ok {
			a.log.Warn("simulators: invalid scenario", "slug", s.Slug)
			continue
		}
		out = append(out, apigen.SimulatorSummary{Slug: sc.Slug, Title: sc.Title, Icon: sc.Icon, Role: sc.Role, TurnsCount: len(sc.Turns)})
	}
	writeJSON(w, http.StatusOK, out)
}

func (a *API) getSimulator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sim, err := a.Sims.Get(ctx, chi.URLParam(r, "simSlug"))
	if err != nil && !isNotFound(err) {
		a.internalError(w, "simulator: get", err)
		return
	}
	var sc apigen.Scenario
	ok := err == nil && (sim.Published || userFrom(ctx).IsAdmin())
	if ok {
		sc, ok = parseScenario(*sim)
	}
	if !ok {
		writeError(w, http.StatusNotFound, codeNotFound, "simulator not found")
		return
	}
	writeJSON(w, http.StatusOK, sc)
}

// parseScenario decodes the stored scenario JSON, normalizing nil slices and maps.
func parseScenario(s model.Simulator) (apigen.Scenario, bool) {
	var sc apigen.Scenario
	if json.Unmarshal([]byte(s.Data), &sc) != nil || sc.Slug == "" {
		return sc, false
	}
	if sc.Metrics == nil {
		sc.Metrics = []apigen.ScenarioMetric{}
	}
	if sc.Turns == nil {
		sc.Turns = []apigen.ScenarioTurn{}
	}
	for i := range sc.Turns {
		if sc.Turns[i].Choices == nil {
			sc.Turns[i].Choices = []apigen.ScenarioChoice{}
		}
		for j := range sc.Turns[i].Choices {
			if sc.Turns[i].Choices[j].Effects == nil {
				sc.Turns[i].Choices[j].Effects = map[string]int{}
			}
		}
	}
	return sc, true
}
