package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestDumpLab writes one lesson's sandbox the way the seeder would store it —
// image, the lesson's setup, the checks in task order — so
// scripts/labcheck/lab.sh can verify a lab without a seeded database:
//
//	GL_DUMP_LAB=<module>/<lesson> GL_DUMP_OUT=<file> go test ./cmd/seed -run TestDumpLab
//
// A normal test run skips it.
func TestDumpLab(t *testing.T) {
	target, out := os.Getenv("GL_DUMP_LAB"), os.Getenv("GL_DUMP_OUT")
	if target == "" || out == "" {
		t.Skip("set GL_DUMP_LAB=<module>/<lesson> and GL_DUMP_OUT=<file>")
	}
	module, lesson, _ := strings.Cut(target, "/")
	type dump struct {
		Image  string   `json:"image"`
		Setup  string   `json:"setup"`
		Checks []string `json:"checks"`
		Titles []string `json:"titles"`
	}
	for _, m := range importedModules() {
		if m.Slug != module {
			continue
		}
		for _, l := range m.Lessons {
			if l.Slug != lesson {
				continue
			}
			var d dump
			seen := map[string]bool{}
			for _, task := range l.Tasks {
				if d.Image == "" && task.Kind == "shell" {
					d.Image = task.SandboxImage
				}
				// Same rule as repository.LessonSandbox: each distinct setup once, in task order.
				if s := task.SetupScript; s != "" && !seen[s] {
					seen[s] = true
					d.Setup += s + "\n"
				}
				if strings.TrimSpace(task.CheckScript) != "" {
					d.Checks = append(d.Checks, task.CheckScript)
					d.Titles = append(d.Titles, task.Title)
				}
			}
			b, err := json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(out, b, 0o644); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("lesson %s not found", target)
}
