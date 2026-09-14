package api

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
)

// TestRoutesMatchSpec fails when a mounted route is missing from api/openapi.yaml.
func TestRoutesMatchSpec(t *testing.T) {
	raw, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}

	api := New(newFakeUsers(), Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	wildcard := regexp.MustCompile(`/\*$`)
	count := 0
	err = chi.Walk(api.Routes(), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		count++
		path := wildcard.ReplaceAllString(strings.TrimSuffix(route, "/"), "/{path}")
		ops, ok := spec.Paths[path]
		if !ok {
			t.Errorf("%s %s: path not in openapi.yaml", method, path)
			return nil
		}
		if _, ok := ops[strings.ToLower(method)]; !ok {
			t.Errorf("%s %s: method not in openapi.yaml", method, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("no routes walked")
	}
}
