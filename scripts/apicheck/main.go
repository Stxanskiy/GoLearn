// Command apicheck probes a running deployment against api/openapi.yaml.
//
//	go run ./scripts/apicheck                                  # https://tot.prod-factory.ru
//	go run ./scripts/apicheck -base http://localhost:8080
//	GOLEARN_SESSION=<cookie value> go run ./scripts/apicheck   # also check signed-in GETs
//
// It only ever sends GET requests without path parameters, so it is safe to
// point at production. What it proves:
//
//   - the site really reaches this service under /api/v1, and not the
//     frontend's own pages: a protected operation must answer 401 with the API's
//     Error JSON, never an HTML page;
//   - every public operation answers 200 with JSON;
//   - the deployed contract (/api/v1/openapi.yaml) is the one in this checkout,
//     so a stale deploy shows up;
//   - with a session, every signed-in GET answers with a status the spec lists.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/backendraz/golearn/internal/auth"
)

type operation struct {
	Security  *[]map[string][]string `yaml:"security"`
	Responses map[string]any         `yaml:"responses"`
	Tags      []string               `yaml:"tags"`
}

type result struct {
	name, want, got string
	ok              bool
}

func main() {
	base := flag.String("base", "https://tot.prod-factory.ru", "site origin; the API is under <base>/api/v1")
	specPath := flag.String("spec", "api/openapi.yaml", "contract to check against")
	flag.Parse()

	raw, err := os.ReadFile(*specPath)
	if err != nil {
		fail("read spec: %v (run from the repository root)", err)
	}
	var spec struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		fail("parse spec: %v", err)
	}

	session := os.Getenv("GOLEARN_SESSION")
	client := &http.Client{Timeout: 20 * time.Second}
	api := strings.TrimRight(*base, "/") + "/api/v1"
	var results []result
	skipped := 0

	// The deployed contract must be this one; otherwise the rest of the report
	// compares the site against a spec it does not run.
	status, ctype, body, err := get(client, api+"/openapi.yaml", "")
	switch {
	case err != nil:
		results = append(results, result{"GET /openapi.yaml", "200, same as checkout", err.Error(), false})
	case status != 200:
		results = append(results, result{"GET /openapi.yaml", "200, same as checkout", fmt.Sprintf("%d %s", status, ctype), false})
	default:
		same := bytes.Equal(body, raw)
		got := "200, same as checkout"
		if !same {
			got = "200, DIFFERENT from checkout (deploy is behind or ahead)"
		}
		results = append(results, result{"GET /openapi.yaml", "200, same as checkout", got, same})
	}
	status, ctype, _, err = get(client, api+"/docs", "")
	results = append(results, check("GET /docs", status, ctype, err, []string{"200"}, "text/html"))

	paths := make([]string, 0, len(spec.Paths))
	for p := range spec.Paths {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, path := range paths {
		for method, node := range spec.Paths[path] {
			if method != "get" {
				continue
			}
			var op operation
			if err := node.Decode(&op); err != nil {
				fail("decode GET %s: %v", path, err)
			}
			if _, ws := op.Responses["101"]; ws || strings.Contains(path, "{") {
				skipped++ // WebSocket upgrades, and operations that need an id or slug
				continue
			}
			public := op.Security != nil && len(*op.Security) == 0
			name := "GET " + path
			switch {
			case public:
				status, ctype, _, err := get(client, api+path, "")
				results = append(results, check(name, status, ctype, err, []string{"200"}, "application/json"))
			case session == "":
				status, ctype, _, err := get(client, api+path, "")
				results = append(results, check(name+" (no session)", status, ctype, err, []string{"401"}, "application/json"))
			default:
				// With a session a 401 means the cookie is stale, not that the
				// operation works, so it is never an accepted answer here.
				status, ctype, _, err := get(client, api+path, session)
				results = append(results, check(name, status, ctype, err, listed(op, "401"), "application/json"))
			}
		}
	}

	failed := 0
	for _, r := range results {
		mark := "OK  "
		if !r.ok {
			mark = "FAIL"
			failed++
		}
		fmt.Printf("%s  %-44s want %-28s got %s\n", mark, r.name, r.want, r.got)
	}
	fmt.Printf("\n%s: %d checked, %d failed, %d skipped (path parameters or WebSocket)\n", api, len(results), failed, skipped)
	if session == "" {
		fmt.Println("Signed-in GETs were checked for 401 only; set GOLEARN_SESSION to check their real responses.")
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func get(c *http.Client, url, session string) (int, string, []byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, "", nil, err
	}
	req.Header.Set("Accept", "application/json")
	if session != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: session})
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err == nil && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") && !json.Valid(body) {
		err = fmt.Errorf("%d with invalid JSON", resp.StatusCode)
	}
	return resp.StatusCode, resp.Header.Get("Content-Type"), body, err
}

func check(name string, status int, ctype string, err error, codes []string, wantType string) result {
	want := strings.Join(codes, "/") + " " + wantType
	if err != nil {
		return result{name, want, err.Error(), false}
	}
	got := fmt.Sprintf("%d %s", status, strings.Split(ctype, ";")[0])
	code := fmt.Sprint(status)
	for _, c := range codes {
		if c == code && strings.HasPrefix(ctype, wantType) {
			return result{name, want, got, true}
		}
	}
	return result{name, want, got, false}
}

// listed returns the status codes the spec documents for an operation, minus except.
func listed(op operation, except string) []string {
	codes := make([]string, 0, len(op.Responses))
	for c := range op.Responses {
		if c != except {
			codes = append(codes, c)
		}
	}
	sort.Strings(codes)
	return codes
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "apicheck: "+format+"\n", args...)
	os.Exit(2)
}
