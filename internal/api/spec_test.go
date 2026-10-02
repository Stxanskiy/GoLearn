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

	api := New(newFakeContent().stores(newFakeUsers()), Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	wildcard := regexp.MustCompile(`/\*$`)
	count := 0
	live := map[string]bool{}
	err = chi.Walk(api.Routes(), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		count++
		live[strings.ToLower(method)+" "+wildcard.ReplaceAllString(strings.TrimSuffix(route, "/"), "/{path}")] = true
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

	// The other direction: an operation documented but never mounted is worse than
	// an undocumented one, because a client generates code against it and only
	// finds out at runtime. "parameters" is a path-level key, not a method.
	for path, ops := range spec.Paths {
		for method := range ops {
			if method == "parameters" {
				continue
			}
			if !live[method+" "+path] {
				t.Errorf("%s %s: in openapi.yaml but not mounted", strings.ToUpper(method), path)
			}
		}
	}
}

// methods are the operation keys in a path item; "parameters" is path-level.
var methods = []string{"get", "post", "put", "patch", "delete"}

type specOp struct {
	Security    *[]map[string][]string `yaml:"security"`
	RequestBody *struct {
		Ref     string               `yaml:"$ref"`
		Content map[string]yaml.Node `yaml:"content"`
	} `yaml:"requestBody"`
	Responses map[string]yaml.Node `yaml:"responses"`
}

type specDoc struct {
	Paths      map[string]map[string]yaml.Node `yaml:"paths"`
	Components struct {
		RequestBodies map[string]struct {
			Content map[string]yaml.Node `yaml:"content"`
		} `yaml:"requestBodies"`
	} `yaml:"components"`
}

func loadSpec(t *testing.T) (specDoc, string) {
	t.Helper()
	raw, err := os.ReadFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var doc specDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc, string(raw)
}

// TestEveryErrorCodeIsDocumented keeps the machine-readable `code` vocabulary
// honest. The frontend localizes by code, so a code the contract never mentions
// is a string the frontend cannot translate — `subscription_required`,
// `internal_error` and `method_not_allowed` were all missing at one point.
func TestEveryErrorCodeIsDocumented(t *testing.T) {
	src, err := os.ReadFile("respond.go")
	if err != nil {
		t.Fatal(err)
	}
	codes := regexp.MustCompile(`code[A-Za-z]+\s*=\s*"([a-z_]+)"`).FindAllStringSubmatch(string(src), -1)
	if len(codes) == 0 {
		t.Fatal("no error code constants found in respond.go")
	}
	_, raw := loadSpec(t)
	for _, m := range codes {
		if !strings.Contains(raw, m[1]) {
			t.Errorf("error code %q is returned by the server but appears nowhere in openapi.yaml", m[1])
		}
	}
}

// TestStatusCodeContract checks the per-operation rules the server actually
// follows. Routing conformance has TestRoutesMatchSpec; before this test the
// status codes had nothing, and the paywall's 402 went undocumented across 21
// operations — a client cannot handle a status the contract does not mention.
//
// 405 is deliberately not checked: it is answered by the router for a known path
// with an unsupported method, so it is stated once in `info.description` rather
// than on every operation.
func TestStatusCodeContract(t *testing.T) {
	doc, _ := loadSpec(t)
	checked := 0
	for path, item := range doc.Paths {
		for _, method := range methods {
			node, ok := item[method]
			if !ok {
				continue
			}
			var op specOp
			if err := node.Decode(&op); err != nil {
				t.Fatalf("%s %s: %v", method, path, err)
			}
			checked++
			has := func(code string) bool { _, ok := op.Responses[code]; return ok }
			fail := func(code, why string) {
				t.Errorf("%s %s: missing %q — %s", strings.ToUpper(method), path, code, why)
			}

			// Every handler can hit a database or a sandbox and give up.
			if !has("500") {
				fail("500", "any operation can fail with internal_error")
			}
			// security: [] marks the operations the router leaves open.
			authed := op.Security == nil || len(*op.Security) > 0
			if authed && !has("401") {
				fail("401", "operation requires a session")
			}
			// csrfGuard runs on every mutating request.
			switch method {
			case "post", "put", "patch", "delete":
				if !has("403") {
					fail("403", "mutating requests are checked for csrf_rejected")
				}
			}
			// decodeJSONLimit rejects any other content type.
			if jsonBody(doc, op) && !has("415") {
				fail("415", "operation takes a JSON body")
			}
			if strings.Contains(path, "{") && !has("404") {
				fail("404", "path parameter may not resolve")
			}
			// visible() runs the paywall for every student-facing operation that
			// loads a lesson, and it loads one from a lesson id, a task id or the
			// course+lesson slug pair. Checked in both directions: a new gated
			// route that forgets 402, and a 402 claimed where no paywall runs.
			if gated := paywalled(path); gated != has("402") {
				if gated {
					fail("402", "operation loads a lesson, so requireCourseAccess runs")
				} else {
					t.Errorf("%s %s: documents 402 but no paywall runs on it",
						strings.ToUpper(method), path)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("no operations checked")
	}
}

func jsonBody(doc specDoc, op specOp) bool {
	if op.RequestBody == nil {
		return false
	}
	content := op.RequestBody.Content
	if ref := op.RequestBody.Ref; ref != "" {
		name := ref[strings.LastIndex(ref, "/")+1:]
		shared, ok := doc.Components.RequestBodies[name]
		if !ok {
			return false
		}
		content = shared.Content
	}
	_, ok := content["application/json"]
	return ok
}

// paywalled reports whether visible() runs for this path. Admin routes load
// lessons through their own editor checks, not the student paywall.
func paywalled(path string) bool {
	if strings.HasPrefix(path, "/admin/") {
		return false
	}
	return strings.Contains(path, "{lessonId}") ||
		strings.Contains(path, "{taskId}") ||
		path == "/courses/{courseSlug}/lessons/{lessonSlug}"
}
