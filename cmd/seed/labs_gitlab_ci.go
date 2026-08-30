package main

// Fixtures + auto-checks for "GitLab CI/CD" (gitlab-ci).
//
// The devops404 original ran these labs against a live GitLab stand: students
// pushed to a real GitLab, watched pipelines run, clicked jobs in the UI, used
// the container registry and a deploy VM. None of that exists in our
// --network none sandbox. So the labs are reframed as *offline authoring* labs:
// the student writes .gitlab-ci.yml in /root/gitlab-ci-lab and each check
// validates the file's structure (grep-based — the base image has no YAML
// parser and no network). Steps that were "push and watch in the UI" become
// "commit the file"; UI/registry/deploy-only steps are reframed to the YAML the
// student would write. Descs override the imported prompts (several of which the
// export also mangled into "$25"-style placeholders).

// ciInit builds the working repository the CI labs assume at /root/gitlab-ci-lab.
const ciInit = `set -e
rm -rf /root/gitlab-ci-lab
git init -q /root/gitlab-ci-lab
git -C /root/gitlab-ci-lab config user.name student
git -C /root/gitlab-ci-lab config user.email student@golearn.local
git -C /root/gitlab-ci-lab symbolic-ref HEAD refs/heads/main
printf '# CI demo project\n' > /root/gitlab-ci-lab/README.md
mkdir -p /root/gitlab-ci-lab/scripts
printf '#!/bin/sh\necho "tests ok"\n' > /root/gitlab-ci-lab/scripts/test.sh
chmod +x /root/gitlab-ci-lab/scripts/test.sh
git -C /root/gitlab-ci-lab add -A
git -C /root/gitlab-ci-lab commit -qm init`

// ci greps the lab's .gitlab-ci.yml (most checks look for keys/jobs in it).
func ci(pattern string) string {
	return `grep -Eq '` + pattern + `' /root/gitlab-ci-lab/.gitlab-ci.yml 2>/dev/null`
}

var gitlabCILabs = map[string]labSpec{
	// ── Lab 1: первый pipeline ──
	"ch-gitlab-ci-lab1": {
		Setup: ciInit,
		Descs: map[int]string{
			1: `<p>В этой песочнице нет живого GitLab — мы учимся <b>писать</b> <code>.gitlab-ci.yml</code>, а проверка смотрит на структуру файла.</p>
<p>В корне репозитория <code>/root/gitlab-ci-lab</code> создай файл <code>.gitlab-ci.yml</code> с одной job, у которой есть <code>script:</code>. Минимум:</p>
<pre>hello:
  script:
    - echo "Hello, CI"</pre>`,
			2: `<p>В настоящем GitLab pipeline запускается после <code>git push</code>. Здесь сети нет, поэтому «запуск» мы имитируем коммитом: закоммить свой <code>.gitlab-ci.yml</code>.</p>
<pre>cd /root/gitlab-ci-lab
git add .gitlab-ci.yml
git commit -m "Add pipeline"</pre>`,
			3: `<p>Добавь вторую job <code>shell-basics</code> с несколькими bash-командами в <code>script</code>:</p>
<pre>shell-basics:
  script:
    - echo "step 1"
    - date
    - ls -la</pre>`,
			4: `<p>Теперь перепиши <code>script</code> в <code>shell-basics</code> на <b>многострочный блок</b> через <code>- |</code>. В YAML это значит «один shell-блок, записанный в несколько строк»:</p>
<pre>shell-basics:
  script:
    - |
      echo "начало"
      for i in 1 2 3; do echo "шаг $i"; done
      echo "конец"</pre>`,
		},
		Checks: map[int]string{
			1: check(`[ -f /root/gitlab-ci-lab/.gitlab-ci.yml ] && `+ci(`^[[:space:]]*script:`),
				".gitlab-ci.yml создан и содержит job со script",
				"создай /root/gitlab-ci-lab/.gitlab-ci.yml с job, у которой есть блок script:"),
			2: check(`git -C /root/gitlab-ci-lab log --oneline -- .gitlab-ci.yml 2>/dev/null | grep -q .`,
				".gitlab-ci.yml закоммичен",
				"cd /root/gitlab-ci-lab && git add .gitlab-ci.yml && git commit -m 'Add pipeline'"),
			3: check(ci(`^shell-basics:`),
				"job shell-basics добавлена",
				"добавь в .gitlab-ci.yml job 'shell-basics:' со script из нескольких команд"),
			4: check(`awk '/^shell-basics:/{f=1} f&&/script:/{s=1} s&&/^[[:space:]]*-[[:space:]]*\|/{print;exit}' /root/gitlab-ci-lab/.gitlab-ci.yml 2>/dev/null | grep -q .`,
				"многострочный блок '- |' есть в script",
				"внутри script у shell-basics используй '- |' и напиши несколько строк под ним"),
		},
	},
}
