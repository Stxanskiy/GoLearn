package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// imageRef matches "repo:tag" and "owner/repo:tag".
var imageRef = regexp.MustCompile(`\b([a-z0-9][a-z0-9._-]*(?:/[a-z0-9][a-z0-9._-]*)?):([a-zA-Z0-9][a-zA-Z0-9._-]*)\b`)

// Every image a lab names has to be inside the sandbox that lab runs in. The
// sandboxes have no network, so an image that is not baked in can never arrive:
// a Docker container never starts and a Pod stays in ImagePullBackOff, while the
// check next to it may well pass on the spec alone.
func TestLabsOnlyUseImagesTheirSandboxCarries(t *testing.T) {
	type problem struct{ module, lesson, image, sandbox string }
	var problems []problem

	for module, lessons := range labFixtures {
		for lesson, spec := range lessons {
			sandbox := spec.Image
			if sandbox == "" {
				sandbox = sandboxImage
			}
			available, known := sandboxImages[sandbox]
			if !known {
				t.Errorf("%s/%s: sandbox %q has no image list in sandbox_images.go", module, lesson, sandbox)
				continue
			}

			scripts := []string{spec.Setup}
			for _, c := range spec.Checks {
				scripts = append(scripts, c)
			}
			// A sandbox without a container runtime can never pull anything, so an
			// image name in one of its lessons is text in a file the student reads or
			// edits: the GitLab CI course writes .gitlab-ci.yml files and Dockerfiles
			// that no pipeline here ever runs, and its checks grep them for the words
			// "docker build". Nothing to verify.
			if len(available) == 0 {
				continue
			}

			seen := map[string]bool{}
			for _, s := range scripts {
				for _, m := range imageRef.FindAllStringSubmatch(s, -1) {
					ref, repo := m[0], m[1]
					if !knownImageRepos[repo] || imageRefExceptions[ref] || seen[ref] {
						continue
					}
					seen[ref] = true
					if !available[ref] {
						problems = append(problems, problem{module, lesson, ref, sandbox})
					}
				}
			}
		}
	}

	sort.Slice(problems, func(i, j int) bool {
		if problems[i].module != problems[j].module {
			return problems[i].module < problems[j].module
		}
		return problems[i].lesson < problems[j].lesson
	})
	for _, p := range problems {
		t.Errorf("%s/%s names %s, which %s does not carry — the lab has no network, so it can never be pulled",
			p.module, p.lesson, p.image, strings.TrimSuffix(strings.TrimPrefix(p.sandbox, "golearn/"), ":latest"))
	}
}

// A task the source types "check" is one the platform is supposed to verify.
// When it reaches the student with no check script it silently turns into a
// "Готово" button, which is how four Docker tasks and one Kubernetes task came
// to be gradeable in a sandbox that had neither docker nor kubectl: the student
// pressed Готово and the platform agreed.
//
// Self-check tasks are a different thing and legitimately have no script.
func TestEveryCheckTaskHasAChecker(t *testing.T) {
	for _, m := range importedModules() {
		for _, l := range m.Lessons {
			for i, task := range l.Tasks {
				if task.SelfCheck || task.Kind != "shell" {
					continue
				}
				key := fmt.Sprintf("%s/%s/%d", m.Slug, l.Slug, i+1)
				if strings.TrimSpace(task.CheckScript) == "" && !uncheckableTasks[key] {
					t.Errorf("%s (%q): typed as a check but has no check script — it reaches the student as a \"Готово\" button. "+
						"Write the check, or add %q to uncheckableTasks with the reason it cannot be checked.",
						key, task.Title, key)
				}
			}
		}
	}
}
