package main

// Fixtures + auto-checks for "Git: основы" (git-basics).
//
// Every lesson builds the exact repository state its tasks assume: a repo with
// history, a conflicting pair of branches, a bare "remote" under /srv/git, and
// so on. The bare repo makes clone/push/fetch work without any network, which
// the lab containers do not have.

// gitInit prepares a repo with an identity the log-filtering lesson can search for.
const gitInit = `git init -q /root/project
git -C /root/project config user.name student
git -C /root/project config user.email student@golearn.local
git -C /root/project symbolic-ref HEAD refs/heads/main`

var gitBasicsLabs = map[string]labSpec{
	// ── Lab 1: первый репозиторий ──
	"ch-git-lab1": {
		Setup: `set -e
rm -rf /root/project
mkdir -p /root/project`,
		Checks: map[int]string{
			1: check(`[ -d /root/project/.git ]`,
				"репозиторий инициализирован",
				"cd /root/project && git init"),
			2: check(`grep -q '# My Project' /root/project/README.md 2>/dev/null && (git -C /root/project diff --cached --name-only | grep -qx 'README.md' || git -C /root/project ls-files | grep -qx 'README.md')`,
				"README.md создан и добавлен в staging",
				"echo '# My Project' > README.md && git add README.md"),
			3: check(`git -C /root/project log --oneline 2>/dev/null | grep -q 'init'`,
				"первый коммит сделан",
				"git commit -m init (если git ругается на identity — git config user.name student)"),
			4: check(`git -C /root/project log --oneline 2>/dev/null | grep -q 'add app' && git -C /root/project show HEAD:app.py 2>/dev/null | grep -q 'print(\"hello\")'`,
				"app.py закоммичен",
				"echo 'print(\"hello\")' > app.py && git add app.py && git commit -m 'add app'"),
			5: check(`[ -z "$(git -C /root/project status --porcelain 2>/dev/null)" ]`,
				"рабочее дерево чистое",
				"git status показывает незакоммиченные изменения — добавь и закоммить их: git add -A && git commit -m wip"),
		},
	},

	// ── Lab 2: изменения и staging ──
	"ch-git-lab2": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
printf '# My Project\n'    > /root/project/README.md
printf 'print("hello")\n'  > /root/project/app.py
printf 'legacy=true\n'     > /root/project/old_config.txt
git -C /root/project add -A
git -C /root/project commit -qm "init"`,
		Checks: map[int]string{
			1: check(`grep -q 'version = "1.0"' /root/project/app.py 2>/dev/null`,
				"строка добавлена, изменение видно в git diff",
				"echo 'version = \"1.0\"' >> /root/project/app.py — затем посмотри: git diff"),
			2: check(`! grep -q 'version = "1.0"' /root/project/app.py 2>/dev/null && ! git -C /root/project status --porcelain app.py | grep -q .`,
				"изменение отменено, файл как в последнем коммите",
				"git restore app.py (вернёт файл к версии из последнего коммита)"),
			3: check(`grep -q 'author = "student"' /root/project/app.py 2>/dev/null && git -C /root/project diff --cached --name-only | grep -qx 'app.py'`,
				"изменение добавлено в staging",
				"echo 'author = \"student\"' >> app.py && git add app.py — затем git diff --staged"),
			4: check(`! git -C /root/project diff --cached --name-only | grep -qx 'app.py' && grep -q 'author = "student"' /root/project/app.py`,
				"файл убран из staging, изменения на месте",
				"git restore --staged app.py (изменения в файле должны остаться)"),
			5: check(`git -C /root/project ls-files | grep -qx 'readme.md' && ! git -C /root/project ls-files | grep -qx 'README.md'`,
				"файл переименован через git",
				"git mv README.md readme.md"),
			6: check(`[ ! -e /root/project/old_config.txt ] && ! git -C /root/project ls-files | grep -qx 'old_config.txt'`,
				"файл удалён из репозитория и рабочей директории",
				"git rm old_config.txt && git commit -m 'remove old config'"),
		},
	},

	// ── Lab 3: ветки и слияние ──
	"ch-git-lab3": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
printf '# My Project\n' > /root/project/README.md
git -C /root/project add -A
git -C /root/project commit -qm "init"`,
		Checks: map[int]string{
			1: check(`git -C /root/project branch --list 'feature/login' | grep -q feature/login`,
				"ветка feature/login создана",
				"git branch feature/login"),
			2: check(`[ "$(git -C /root/project rev-parse --abbrev-ref HEAD)" = feature/login ]`,
				"переключение на feature/login выполнено",
				"git switch feature/login — сейчас ты на ветке $(git -C /root/project rev-parse --abbrev-ref HEAD)"),
			3: check(`git -C /root/project log feature/login --oneline | grep -q 'add login' && git -C /root/project show feature/login:login.py >/dev/null 2>&1`,
				"коммит в ветке feature/login сделан",
				"echo '# login module' > login.py && git add login.py && git commit -m 'add login'"),
			4: check(`git -C /root/project branch --list 'feature/api' | grep -q feature/api`,
				"ветка feature/api создана",
				"git switch -c feature/api (создаёт ветку и сразу переключается)"),
			5: check(`[ "$(git -C /root/project rev-parse --abbrev-ref HEAD)" = main ] && [ -f /root/project/login.py ] && git -C /root/project branch --merged main | grep -q feature/login`,
				"feature/login слита в main",
				"git switch main && git merge feature/login"),
			6: check(`! git -C /root/project branch --list 'feature/login' | grep -q feature/login`,
				"слитая ветка удалена",
				"git branch -d feature/login"),
		},
	},

	// ── Lab 4: конфликты ──
	"ch-git-lab4": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
printf 'Main version\n' > /root/project/app.txt
git -C /root/project add -A
git -C /root/project commit -qm "init"
git -C /root/project switch -qc branch-b
printf 'Branch B version\n' > /root/project/app.txt
git -C /root/project commit -qam "branch b change"
git -C /root/project switch -q main
printf 'Main updated version\n' > /root/project/app.txt
git -C /root/project commit -qam "main change"`,
		Checks: map[int]string{
			1: check(`[ -f /root/project/.git/MERGE_HEAD ] || git -C /root/project status --porcelain | grep -q '^UU'`,
				"конфликт возник, конфликтующий файл виден в git status",
				"git merge branch-b — затем посмотри git status (файлы в состоянии both modified)"),
			2: check(`grep -q 'Resolved version' /root/project/app.txt 2>/dev/null && ! grep -qE '^(<<<<<<<|=======|>>>>>>>)' /root/project/app.txt`,
				"конфликт разрешён, маркеры удалены",
				"Открой app.txt, удали строки <<<<<<< ======= >>>>>>> и оставь: Resolved version"),
			3: check(`[ -z "$(git -C /root/project ls-files -u)" ] && git -C /root/project diff --cached --name-only | grep -qx 'app.txt'`,
				"файл отмечен как разрешённый",
				"git add app.txt"),
			4: check(`git -C /root/project log --oneline | grep -q 'resolve conflict' && [ ! -f /root/project/.git/MERGE_HEAD ]`,
				"merge завершён коммитом",
				"git commit -m 'resolve conflict'"),
			5: check(`[ "$(git -C /root/project rev-list --parents -n1 HEAD | wc -w)" -ge 3 ] && git -C /root/project log -1 --pretty=%s | grep -q 'resolve conflict'`,
				"в истории есть merge commit с двумя родителями",
				"Проверь: git log --oneline --graph — последний коммит должен быть merge-коммитом"),
		},
	},

	// ── Lab 5: удалённый репозиторий (bare repo вместо сети) ──
	"ch-git-lab5": {
		Setup: `set -e
rm -rf /root/project /srv/git /tmp/seedrepo
mkdir -p /srv/git
git init -q --bare /srv/git/project.git
git init -q /tmp/seedrepo
git -C /tmp/seedrepo config user.name student
git -C /tmp/seedrepo config user.email student@golearn.local
git -C /tmp/seedrepo symbolic-ref HEAD refs/heads/main
printf '# Shared project\n' > /tmp/seedrepo/README.md
git -C /tmp/seedrepo add -A
git -C /tmp/seedrepo commit -qm "init"
git -C /tmp/seedrepo push -q /srv/git/project.git main
rm -rf /tmp/seedrepo`,
		Checks: map[int]string{
			1: check(`[ -d /root/project/.git ] && [ -f /root/project/README.md ] && git -C /root/project remote -v | grep -q '/srv/git/project.git'`,
				"репозиторий склонирован",
				"git clone /srv/git/project.git /root/project"),
			2: check(`git --git-dir=/srv/git/project.git log --oneline --all 2>/dev/null | grep -q 'add feature'`,
				"коммит запушен в origin",
				"cd /root/project && echo x > feature.txt && git add feature.txt && git commit -m 'add feature' && git push origin main"),
			3: check(`git --git-dir=/srv/git/project.git show --name-only --pretty=format: HEAD 2>/dev/null | grep -q 'feature.txt'`,
				"feature.txt виден в bare-репозитории",
				"Проверь удалённый репозиторий: git --git-dir=/srv/git/project.git log --oneline --stat -1"),
			4: check(`git -C /root/project branch -a 2>/dev/null | grep -q 'remotes/origin'`,
				"remote-tracking ветки получены",
				"git fetch origin затем git branch -a (нужен префикс remotes/origin)"),
			5: check(`git --git-dir=/srv/git/project.git rev-parse --verify dev >/dev/null 2>&1 && git --git-dir=/srv/git/project.git show dev:dev.txt 2>/dev/null | grep -q dev`,
				"ветка dev запушена в origin",
				"git switch -c dev && echo dev > dev.txt && git add dev.txt && git commit -m 'add dev' && git push origin dev"),
		},
	},

	// ── Lab 6: теги, stash, revert ──
	"ch-git-lab6": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
printf '# My Project\n' > /root/project/README.md
git -C /root/project add -A
git -C /root/project commit -qm "init"
printf 'feature v1\n' > /root/project/feature.txt
git -C /root/project add -A
git -C /root/project commit -qm "add feature"
printf 'broken\n' > /root/project/bad.txt
git -C /root/project add -A
git -C /root/project commit -qm "bad commit"
printf 'feature v2 (не закоммичено)\n' >> /root/project/feature.txt`,
		Checks: map[int]string{
			1: check(`[ "$(git -C /root/project cat-file -t v1.0.0 2>/dev/null)" = tag ] && git -C /root/project tag -n1 v1.0.0 | grep -q 'First release'`,
				"аннотированный тег v1.0.0 создан",
				"git tag -a v1.0.0 -m 'First release' (без -a получится лёгкий тег — проверка его не примет)"),
			2: check(`[ -z "$(git -C /root/project status --porcelain 2>/dev/null)" ] && git -C /root/project stash list | grep -q .`,
				"изменения убраны в stash, дерево чистое",
				"git stash — затем git status должен быть чистым, а git stash list непустым"),
			3: check(`[ -z "$(git -C /root/project stash list 2>/dev/null)" ] && git -C /root/project status --porcelain | grep -q .`,
				"изменения возвращены из stash",
				"git stash pop"),
			4: check(`git -C /root/project log --oneline | grep -qi 'revert' && [ ! -e /root/project/bad.txt ]`,
				"плохой коммит отменён через revert",
				"Найди хеш: git log --oneline | grep 'bad commit', затем git revert --no-edit <hash>"),
			5: check(`git -C /root/project log --oneline | grep -qi 'revert' && git -C /root/project log --oneline | grep -q 'bad commit'`,
				"история не переписана: остались и плохой коммит, и revert",
				"git log --oneline — в истории должны быть обе записи: bad commit и Revert ..."),
		},
	},

	// ── Lab 7: история и cherry-pick ──
	"ch-git-lab7": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
printf '# My Project\n' > /root/project/README.md
git -C /root/project add -A
git -C /root/project commit -qm "init readme"
printf 'вторая строка\n' >> /root/project/README.md
git -C /root/project commit -qam "update readme"
git -C /root/project switch -qc hotfix
printf 'hotfix applied\n' > /root/project/fix.txt
git -C /root/project add -A
git -C /root/project commit -qm "hotfix: critical fix"
git -C /root/project switch -q main
git -C /root/project switch -qc develop
printf 'develop only\n' > /root/project/develop.txt
git -C /root/project add -A
git -C /root/project commit -qm "add develop file"
git -C /root/project switch -q main
rm -f /root/student_commits.txt /root/readme_history.txt /root/branch_diff.txt`,
		Checks: map[int]string{
			1: check(`[ -s /root/student_commits.txt ] && grep -qE '[0-9a-f]{7}' /root/student_commits.txt`,
				"коммиты автора student выгружены",
				"git -C /root/project log --author=student --oneline > /root/student_commits.txt"),
			2: check(`[ -s /root/readme_history.txt ] && grep -q 'readme' /root/readme_history.txt`,
				"история README.md сохранена",
				"git -C /root/project log --oneline -- README.md > /root/readme_history.txt"),
			3: check(`[ -f /root/project/fix.txt ] && [ "$(git -C /root/project rev-parse --abbrev-ref HEAD)" = main ] && git -C /root/project log main --oneline | grep -q 'critical fix'`,
				"коммит из hotfix перенесён в main",
				"Находясь на main: git cherry-pick $(git rev-parse hotfix)"),
			4: check(`grep -q 'develop.txt' /root/branch_diff.txt 2>/dev/null`,
				"разница между ветками сохранена",
				"git -C /root/project diff --name-only main develop > /root/branch_diff.txt"),
		},
	},
}
