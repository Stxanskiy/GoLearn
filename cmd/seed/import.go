package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
)

// rscFlight matches the start of a Next.js React-Server-Components flight node
// (e.g. ["$","div",null,  or  ["$","$L24",null,). The scrape sometimes appends
// the page's flight payload after the real lesson HTML — we cut it off there.
var rscFlight = regexp.MustCompile(`\["\$","[^"]+",null,`)

// trailingChunkID strips a leftover flight chunk id like "\nf:" or "\n28:".
var trailingChunkID = regexp.MustCompile(`[0-9A-Za-z]{1,5}:\s*$`)

// imgTag / pictureBlock strip lesson illustrations and cover art. The course
// images are being removed (to be replaced later); the embedded data-URI covers
// also bloat every lesson by hundreds of KB. Migration 014 does the same to
// rows already in the database.
var imgTag = regexp.MustCompile(`(?is)<img\b[^>]*>`)
var pictureBlock = regexp.MustCompile(`(?is)<picture\b[^>]*>.*?</picture>`)
var emptyMediaWrap = regexp.MustCompile(`(?is)<(figure|figcaption)\b[^>]*>\s*</(figure|figcaption)>`)

func cleanContent(h string) string {
	if loc := rscFlight.FindStringIndex(h); loc != nil {
		h = h[:loc[0]]
		h = strings.TrimRight(h, " \t\r\n")
		h = trailingChunkID.ReplaceAllString(h, "")
	}
	h = pictureBlock.ReplaceAllString(h, "")
	h = imgTag.ReplaceAllString(h, "")
	h = emptyMediaWrap.ReplaceAllString(h, "")
	return strings.TrimSpace(h)
}

//go:embed all:content
var contentFS embed.FS

// ── parser JSON schema ──

type pkCourse struct {
	ID string `json:"id"`
	// Some exports put the course title at the top level and leave "meta" empty
	// (crs_gitlab_ci does); others fill "meta". Read both, prefer meta.
	Title       string `json:"title"`
	Description string `json:"description"`
	Meta        struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Image       string `json:"image"` // data URI (base64 webp)
	} `json:"meta"`
	ChapterList []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position int    `json:"position"`
	} `json:"chapter_list"`
}

type pkOption struct {
	Text        string `json:"text"`
	Correct     bool   `json:"correct"`
	Explanation string `json:"explanation"`
}

type pkTask struct {
	Type         string     `json:"type"` // quiz | check | self | code
	Title        string     `json:"title"`
	TaskID       string     `json:"taskid"`
	Hint         string     `json:"hint"`
	QuestionHTML string     `json:"question_html"`
	Options      []pkOption `json:"options"`

	// Code tasks: solved in the editor on the page rather than in a terminal.
	// Lang is the task kind the runner dispatches on ("python", "python-gui",
	// "go"); the window kind starts a program that opens a window instead of
	// printing. Image is needed because a Qt lesson cannot run in the base
	// sandbox.
	Lang        string   `json:"lang"`
	StarterCode string   `json:"starter_code"`
	Image       string   `json:"image"`
	Tests       []pkTest `json:"tests"`
}

// pkTest is one input/expected pair: the student's program runs with Input on
// stdin and its stdout is compared with Expected.
type pkTest struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type pkChapter struct {
	N           int      `json:"n"`
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Kind        string   `json:"kind"` // lesson | tasks | other
	ContentHTML string   `json:"content_html"`
	Tasks       []pkTask `json:"tasks"`
	VMImage     string   `json:"vm_image"`
	VMInit      string   `json:"vm_init"`
	Format      string   `json:"format"`
}

// importSpec describes how an embedded parser course maps into a GoLearn module.
type importSpec struct {
	Dir        string // folder under content/
	Slug       string
	Track      string
	Difficulty string
	Category   string
	Trainer    bool // practice-only course, listed under trainers
	// Draft seeds the course unpublished, so content still being written reaches
	// production hidden; an admin publishes it when it is ready.
	Draft bool
	Image string // force this sandbox image on all shell tasks (empty = auto)
}

// importedModules builds GoLearn modules from embedded parser content.
func importedModules() []M {
	specs := []importSpec{
		// ── DevOps section (all from the imported export) ──
		{Dir: "crs_express_devops", Slug: "express-devops", Track: "devops", Difficulty: "beginner", Category: "DevOps"},
		{Dir: "module_devops", Slug: "devops-foundations", Track: "devops", Difficulty: "beginner", Category: "DevOps"},
		{Dir: "module_linux_start", Slug: "linux-start", Track: "devops", Difficulty: "beginner", Category: "Linux"},
		{Dir: "module_linux_core", Slug: "linux-core", Track: "devops", Difficulty: "intermediate", Category: "Linux"},
		{Dir: "module_linux_advanced", Slug: "linux-advanced", Track: "devops", Difficulty: "advanced", Category: "Linux"},
		{Dir: "module_git_basics", Slug: "git-basics", Track: "devops", Difficulty: "beginner", Category: "Git"},
		{Dir: "crs_gitlab_ci", Slug: "gitlab-ci", Track: "devops", Difficulty: "intermediate", Category: "CI/CD"},
		{Dir: "module_ansible", Slug: "ansible", Track: "devops", Difficulty: "intermediate", Category: "Ansible"},
		{Dir: "module_docker_basics", Slug: "docker-basics", Track: "devops", Difficulty: "intermediate", Category: "Docker"},
		{Dir: "module_docker_compose", Slug: "docker-compose", Track: "devops", Difficulty: "intermediate", Category: "Docker"},
		{Dir: "module_k8s_intro", Slug: "k8s-intro", Track: "devops", Difficulty: "intermediate", Category: "Kubernetes"},
		{Dir: "module_k8s_ckad", Slug: "k8s-ckad", Track: "devops", Difficulty: "advanced", Category: "Kubernetes", Image: sandboxImageK8s},
		{Dir: "module_helm", Slug: "helm", Track: "devops", Difficulty: "intermediate", Category: "Kubernetes", Image: sandboxImageK8s},
		// ── Python section: с нуля до демонстрационного экзамена ──
		// Курсы 1–5 идут в базовом образе: в нём уже есть python3 и python3-venv.
		// Курсам 6–9 нужен образ с PostgreSQL и колёсами PyQt6 — его ещё нет,
		// поэтому Image им пока не задан (см. tmp/python-demo-exam-plan.md).
		{Dir: "module_python_first_steps", Slug: "python-first-steps", Track: "python", Difficulty: "beginner", Category: "Python", Draft: true},
		{Dir: "module_python_data", Slug: "python-data", Track: "python", Difficulty: "beginner", Category: "Python", Draft: true},
		{Dir: "module_python_functions", Slug: "python-functions", Track: "python", Difficulty: "beginner", Category: "Python", Draft: true},
		{Dir: "module_python_oop", Slug: "python-oop", Track: "python", Difficulty: "intermediate", Category: "Python", Draft: true},
		{Dir: "module_python_quality", Slug: "python-quality", Track: "python", Difficulty: "intermediate", Category: "Python", Draft: true},
		{Dir: "module_postgres_psycopg", Slug: "postgres-psycopg", Track: "python", Difficulty: "intermediate", Category: "PostgreSQL", Draft: true},
		{Dir: "module_pyqt6_start", Slug: "pyqt6-start", Track: "python", Difficulty: "intermediate", Category: "PyQt6", Draft: true},
		{Dir: "module_pyqt6_data", Slug: "pyqt6-data", Track: "python", Difficulty: "advanced", Category: "PyQt6", Draft: true},
		{Dir: "module_demo_exam", Slug: "demo-exam", Track: "python", Difficulty: "advanced", Category: "Python", Draft: true},
		// ── Database section ──
		{Dir: "course_pg_start", Slug: "pg-start", Track: "database", Difficulty: "beginner", Category: "PostgreSQL"},
		{Dir: "course_pg_sql", Slug: "pg-sql", Track: "database", Difficulty: "beginner", Category: "PostgreSQL", Draft: true},
		{Dir: "course_pg_ops", Slug: "pg-ops", Track: "database", Difficulty: "intermediate", Category: "PostgreSQL", Draft: true},
		// ── Trainers (gyms) — practice-only, shown on /trainers, not in /courses ──
		{Dir: "gym_linux_start", Slug: "gym-linux-start", Track: "devops", Trainer: true, Difficulty: "beginner", Category: "Linux"},
		{Dir: "gym_linux_troubleshoot", Slug: "gym-linux-troubleshoot", Track: "devops", Trainer: true, Difficulty: "intermediate", Category: "Linux"},
		{Dir: "gym_git", Slug: "gym-git", Track: "devops", Trainer: true, Difficulty: "beginner", Category: "Git"},
		{Dir: "gym_sql", Slug: "gym-sql", Track: "database", Trainer: true, Difficulty: "beginner", Category: "PostgreSQL", Draft: true},
	}
	var mods []M
	for _, s := range specs {
		m, err := buildModule(s)
		if err != nil {
			fmt.Printf("import %s: %v\n", s.Dir, err)
			continue
		}
		mods = append(mods, m)
	}
	return mods
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func buildModule(s importSpec) (M, error) {
	base := path.Join("content", s.Dir)

	var course pkCourse
	if data, err := contentFS.ReadFile(path.Join(base, "_course.json")); err == nil {
		_ = json.Unmarshal(data, &course)
	}

	// Map chapter id -> filename (files are "NN_<id>.json"; NN is scrape order, NOT
	// learning order). Learning order is the ORDER OF THE chapter_list ARRAY — the
	// "position" field in it is decorative and read by nothing, so a chapter
	// appended to the end lands at the end no matter what position it claims.
	entries, err := contentFS.ReadDir(base)
	if err != nil {
		return M{}, err
	}
	byID := make(map[string]string)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || name == "_course.json" || !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if i := strings.Index(id, "_"); i >= 0 {
			id = id[i+1:] // strip "NN_" prefix
		}
		byID[id] = name
	}

	// Ordered chapter ids: prefer chapter_list; fall back to sorted filenames.
	ordered := course.ChapterList
	if len(ordered) == 0 {
		var ids []string
		for id := range byID {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			ordered = append(ordered, struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Position int    `json:"position"`
			}{ID: id})
		}
	}

	m := M{
		Slug:        s.Slug,
		Title:       firstNonEmpty(course.Meta.Title, course.Title),
		Description: firstNonEmpty(course.Meta.Description, course.Description),
		Track:       s.Track,
		Difficulty:  s.Difficulty,
		Category:    s.Category,
		Trainer:     s.Trainer,
		Draft:       s.Draft,
		// CoverImage intentionally empty: covers are SVG placeholders until the
		// admin sets real ones (user request).
	}
	if m.Title == "" {
		m.Title = s.Slug
	}

	for idx, entry := range ordered {
		fname, ok := byID[entry.ID]
		if !ok {
			continue
		}
		var ch pkChapter
		data, err := contentFS.ReadFile(path.Join(base, fname))
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &ch); err != nil {
			continue
		}
		i := idx

		l := L{
			Slug:    slugify(ch.ID, i+1),
			Title:   ch.Title,
			Content: cleanContent(ch.ContentHTML),
			Order:   i + 1,
			Track:   s.Track,
			VMImage: ch.VMImage,
			VMInit:  ch.VMInit,
			Format:  ch.Format,
		}

		// "self" tasks are practice steps the platform never validated; they are
		// still work the student does, so they belong in the lesson.
		hasCheck := false
		for _, t := range ch.Tasks {
			if t.Type == "check" || t.Type == "self" {
				hasCheck = true
				break
			}
		}

		switch {
		case ch.Kind == "other":
			l.Kind = "sim"
		case ch.Kind == "tasks" && hasCheck:
			l.Kind = "lab"
		case ch.Kind == "tasks":
			l.Kind = "quiz"
		default:
			l.Kind = "theory"
		}

		// The source keeps questions and tasks in ONE ordered list, and the author
		// interleaved them on purpose: a task, then the question that checks it,
		// then the next task. Splitting that list into l.Quiz and l.Tasks loses
		// the sequence - each is then numbered from one, independently, and the
		// lesson reads as "all the questions, then all the work". Carry the
		// position in the source list through both so the order survives.
		for pos, t := range ch.Tasks {
			at := pos + 1
			switch t.Type {
			case "quiz":
				if q, ok := toQuiz(t); ok {
					q.Pos = at
					l.Quiz = append(l.Quiz, q)
				}
			case "check":
				task := toShellTask(t, ch.VMImage)
				task.Pos = at
				l.Tasks = append(l.Tasks, task)
			case "self":
				task := toShellTask(t, ch.VMImage)
				task.SelfCheck = true
				task.Pos = at
				l.Tasks = append(l.Tasks, task)
			case "code":
				task := toCodeTask(t)
				task.Pos = at
				l.Tasks = append(l.Tasks, task)
			}
		}

		// Attach fixtures + auto-checks where this course has them authored.
		applyLabFixtures(s.Slug, &l)

		// Force the course's sandbox image where the spec pins one (helm and
		// k8s-ckad need the k3s image, not the base shell image).
		if s.Image != "" {
			for i := range l.Tasks {
				if l.Tasks[i].Kind == "shell" {
					l.Tasks[i].SandboxImage = s.Image
				}
			}
		}

		m.Lessons = append(m.Lessons, l)
	}
	return m, nil
}

func toQuiz(t pkTask) (Q, bool) {
	if len(t.Options) == 0 {
		return Q{}, false
	}
	q := Q{Question: strings.TrimSpace(t.QuestionHTML)}
	correctExpl := ""
	for i, o := range t.Options {
		q.Options = append(q.Options, o.Text)
		q.OptionExpl = append(q.OptionExpl, o.Explanation)
		if o.Correct {
			q.Correct = i
			correctExpl = o.Explanation
		}
	}
	q.Explanation = correctExpl
	return q, true
}

func toShellTask(t pkTask, vmImage string) T {
	// The export's vm_image values are platform-side labels ("base", "docker",
	// "kuber", "golearn/linux:latest"), not images that exist here — a container
	// could not even start with them. Anything we do not actually ship falls
	// back to the standard sandbox image.
	img := vmImage
	switch img {
	case sandboxImage, sandboxImagePG, sandboxImageDocker, sandboxImageK8s:
	default:
		img = sandboxImage
	}
	return T{
		SourceKey:    strings.TrimSpace(t.TaskID),
		Title:        t.Title,
		Description:  t.QuestionHTML,
		Hints:        t.Hint,
		Difficulty:   "medium",
		Kind:         "shell",
		SandboxImage: img,
		// check_script intentionally empty: platform-side validator is not in the
		// static export. Lab runs in a real terminal; auto-check added per task later.
	}
}

// toCodeTask builds a task solved in the editor. Unlike a shell task it carries
// no check script: a code task is checked by running it against its test cases,
// and a windowed one is not auto-checked at all — the student looks at the
// window.
func toCodeTask(t pkTask) T {
	lang := t.Lang
	if lang == "" {
		lang = "python"
	}
	tests := make([]TestCase, 0, len(t.Tests))
	for _, tc := range t.Tests {
		tests = append(tests, TestCase{Input: tc.Input, ExpectedOutput: tc.Expected})
	}
	return T{
		SourceKey:    strings.TrimSpace(t.TaskID),
		Title:        t.Title,
		Description:  t.QuestionHTML,
		Hints:        t.Hint,
		Difficulty:   "medium",
		Kind:         lang,
		StarterCode:  t.StarterCode,
		SandboxImage: t.Image,
		TestCases:    tests,
		// A window task has nothing to compare, so it is the student's own call
		// that it is done — the same contract as a "self" task.
		SelfCheck: lang == "python-gui" || len(tests) == 0,
	}
}

func slugify(id string, n int) string {
	s := strings.ToLower(id)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == ' ' || r == '/':
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = fmt.Sprintf("ch-%d", n)
	}
	return out
}
