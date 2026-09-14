// Package lab holds sandbox lab rules shared by the HTML handlers and the API: session keys, path jail, preview and git helpers.
package lab

import (
	"fmt"
	"path"
	"strings"

	"github.com/backendraz/golearn/internal/model"
)

// Key names the sandbox session shared by a lesson's terminal, steps and checks.
func Key(lessonID int) string { return fmt.Sprintf("l%d", lessonID) }

// Git trainer sandbox: one repo per user.
const (
	GitTrainerKey   = "git"
	GitTrainerImage = "golearn/git:latest"
	GitTrainerRepo  = "/root/repo"
	GitTrainerSetup = `git config --global user.name "Student"
git config --global user.email "student@golearn.local"
git config --global init.defaultBranch main
git config --global advice.detachedHead false
rm -rf /root/repo && mkdir -p /root/repo && cd /root/repo && git init -q
printf '# My Project\n' > README.md && git add README.md && git commit -qm "Initial commit"
echo /root/repo > /root/.gl_cwd`
)

// SandboxImages are the lab images a shell task may use.
var SandboxImages = []string{
	"golearn/sandbox:latest",
	"golearn/sandbox-pg:latest",
	"golearn/sandbox-docker:latest",
	"golearn/sandbox-k8s:latest",
}

// LabGitRepo is the repository git labs work in.
const LabGitRepo = "/root/project"

// IsGitLab reports whether a course's labs show the commit graph.
func IsGitLab(m model.Module) bool {
	return strings.EqualFold(m.Category, "Git") || strings.Contains(strings.ToLower(m.Slug), "git")
}

// EditRoot is the directory the in-lab editor may browse and edit.
const EditRoot = "/root"

// JailPath cleans p (absolute or relative to EditRoot) and reports whether it stays inside EditRoot.
func JailPath(p string) (string, bool) {
	if strings.TrimSpace(p) == "" {
		p = EditRoot
	}
	if !strings.HasPrefix(p, "/") {
		p = EditRoot + "/" + p
	}
	p = path.Clean(p)
	if p != EditRoot && !strings.HasPrefix(p, EditRoot+"/") {
		return "", false
	}
	return p, true
}

// InjectBase inserts <base href> so relative links on a previewed page resolve through the proxy prefix.
func InjectBase(body []byte, base string) []byte {
	tag := []byte(`<base href="` + base + `">`)
	lower := strings.ToLower(string(body))
	if i := strings.Index(lower, "<head>"); i >= 0 {
		i += len("<head>")
		return append(append(append([]byte{}, body[:i]...), tag...), body[i:]...)
	}
	return append(tag, body...)
}

// PreviewErrorPage is the iframe placeholder shown when nothing answers on the preview port.
func PreviewErrorPage(port int, err error) string {
	reason := "на этом порту пока никто не отвечает"
	if strings.Contains(err.Error(), "curl-missing") {
		reason = "в этой песочнице нет curl"
	}
	return fmt.Sprintf(`<!doctype html><html lang="ru"><head><meta charset="utf-8">`+
		`<style>body{margin:0;font-family:system-ui,sans-serif;background:#22201b;color:#d6c9b6;`+
		`display:flex;align-items:center;justify-content:center;height:100vh;text-align:center}`+
		`.c{max-width:420px;padding:24px}h2{margin:0 0 8px;font-size:1.1em;color:#e8dcc8}`+
		`code{background:#000;padding:2px 7px;border-radius:5px;color:#7fd6c2}p{line-height:1.6;font-size:.9em;color:#a89a85}</style>`+
		`</head><body><div class="c"><h2>Пока нечего показать</h2>`+
		`<p>Web Preview обращается к <code>127.0.0.1:%d</code> внутри песочницы — %s.<br><br>`+
		`Запусти сервер в терминале (например <code>python3 -m http.server %d</code> или свой контейнер), затем нажми «Обновить».</p></div></body></html>`,
		port, reason, port)
}
