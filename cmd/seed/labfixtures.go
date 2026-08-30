package main

// Lab fixtures and auto-checks for imported courses.
//
// The devops404 export ships task text but no environment and no validator, so
// a task like "выведи ошибки из /opt/devops/lab5/server.log" had nothing to
// read and no way to pass. Each lesson below gets:
//   - Setup: the files the lesson's tasks talk about, created once per lab
//     container (the whole lesson shares one sandbox);
//   - Checks: a script per task, keyed by the task's 1-based position; exit 0
//     means solved. Checks print a short hint on failure, so a red result tells
//     the student what is still missing.
import "strings"

type labSpec struct {
	// Image overrides the sandbox image for this lesson (empty -> the default
	// golearn/sandbox). The SQL course needs a PostgreSQL server, for example.
	Image  string
	Setup  string
	Checks map[int]string
	// Descs replaces a task's description. Used where the imported text is
	// unusable (the export mangled it), keyed like Checks by task position.
	Descs map[int]string
}

// labFixtures maps module slug -> lesson slug -> lab definition.
var labFixtures = map[string]map[string]labSpec{
	"linux-start":            linuxStartLabs,
	"linux-core":             linuxCoreLabs,
	"linux-advanced":         linuxAdvancedLabs,
	"git-basics":             gitBasicsLabs,
	"gitlab-ci":              gitlabCILabs,
	"ansible":                ansibleLabs,
	"express-devops":         expressDevopsLabs,
	"gym-git":                gymGitLabs,
	"sql-express":            sqlExpressLabs,
	"docker-basics":          dockerBasicsLabs,
	"docker-compose":         dockerComposeLabs,
	"k8s-intro":              k8sIntroLabs,
	"gym-linux-troubleshoot": gymTroubleshootLabs,
	"gym-linux-start": rekey(linuxStartLabs, map[string]string{
		"ch-lnav-lab1": "gym-lstart-lab1",
		"ch-lnav-lab2": "gym-lstart-lab2",
		"ch-lnav-lab3": "gym-lstart-lab3",
		"ch-lnav-lab4": "gym-lstart-lab4",
		"ch-lnav-lab5": "gym-lstart-lab5",
		"ch-lnav-lab6": "gym-lstart-lab6",
		"ch-lnav-lab7": "gym-lstart-lab7",
		"ch-lnav-lab8": "gym-lstart-lab8",
	}),
}

// rekey re-uses one course's lab definitions under different lesson slugs.
// The Linux trainer ("тренажёр") repeats the Linux course labs verbatim, so it
// shares their fixtures and checks instead of duplicating them.
func rekey(src map[string]labSpec, slugs map[string]string) map[string]labSpec {
	out := make(map[string]labSpec, len(slugs))
	for from, to := range slugs {
		if spec, ok := src[from]; ok {
			out[to] = spec
		}
	}
	return out
}

// applyLabFixtures attaches the setup script and per-task checks for one lesson.
// Tasks without an entry keep the manual "Готово" button.
func applyLabFixtures(moduleSlug string, l *L) {
	lessons, ok := labFixtures[moduleSlug]
	if !ok {
		return
	}
	spec, ok := lessons[l.Slug]
	if !ok {
		return
	}
	for i := range l.Tasks {
		if spec.Image != "" {
			l.Tasks[i].SandboxImage = spec.Image
		}
		if i == 0 {
			l.Tasks[i].SetupScript = spec.Setup
		}
		if chk, ok := spec.Checks[i+1]; ok {
			l.Tasks[i].CheckScript = chk
		}
		if d, ok := spec.Descs[i+1]; ok {
			l.Tasks[i].Description = d
		}
	}
}

// ok/fail keep every check's output in the same shape. Hints are echoed inside
// double quotes so $(...) in a hint still reports live state, which means any
// literal quote in the text has to be escaped or it would break the script.
func ok(msg string) string   { return `echo "✓ ` + escapeQuotes(msg) + `"` }
func fail(msg string) string { return `echo "✗ ` + escapeQuotes(msg) + `"; exit 1` }

func escapeQuotes(s string) string { return strings.ReplaceAll(s, `"`, `\"`) }

// check builds "if <cond>; then ✓; else ✗; fi".
func check(cond, good, bad string) string {
	return "if " + cond + "; then " + ok(good) + "; else " + fail(bad) + "; fi"
}
