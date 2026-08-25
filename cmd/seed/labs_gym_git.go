package main

// Fixtures + auto-checks for the Git trainer (gym-git).
// Unlike the Git course, these labs start from an already-populated repository:
// the drill is the operation itself (squash, cherry-pick, stash, revert), so the
// history the task talks about is prepared here.

var gymGitLabs = map[string]labSpec{
	// ── Lab 1: clone → commit → push → branch → merge ──
	"ch-git-gym-lab1": {
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
			1: check(`[ -d /root/project/.git ] && [ -f /root/project/README.md ]`,
				"репозиторий склонирован",
				"git clone /srv/git/project.git /root/project"),
			2: check(`git -C /root/project log --oneline | grep -q 'Add feature' && git -C /root/project show HEAD:feature.txt 2>/dev/null | grep -q 'new feature'`,
				"feature.txt закоммичен",
				"echo 'new feature' > feature.txt && git add feature.txt && git commit -m 'Add feature'"),
			3: check(`git --git-dir=/srv/git/project.git log --oneline --all 2>/dev/null | grep -q 'Add feature'`,
				"коммит доехал до bare-репозитория",
				"git push origin main — проверь: git --git-dir=/srv/git/project.git log --oneline"),
			4: check(`[ "$(git -C /root/project rev-parse --abbrev-ref HEAD)" = dev ]`,
				"ветка dev создана, переключение выполнено",
				"git switch -c dev — сейчас ты на ветке $(git -C /root/project rev-parse --abbrev-ref HEAD)"),
			5: check(`git -C /root/project log main --oneline | grep -q "$(git -C /root/project log dev --oneline -1 | cut -d' ' -f1)" && git -C /root/project show main:dev.txt >/dev/null 2>&1`,
				"ветка dev слита в main",
				"В dev: echo x > dev.txt && git add dev.txt && git commit -m 'add dev'; затем git switch main && git merge dev"),
		},
	},

	// ── Lab 2: перепись истории ──
	"ch-git-gym-lab2": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
for i in 1 2 3 4 5; do
  printf "change %s\n" "$i" >> /root/project/feature.txt
  git -C /root/project add -A
  git -C /root/project commit -qm "step $i"
done
rm -f /root/reflog.txt /root/gitlog.txt`,
		Checks: map[int]string{
			1: check(`[ "$(git -C /root/project rev-list --count HEAD)" = 1 ] && git -C /root/project log -1 --pretty=%s | grep -q 'Complete feature'`,
				"пять коммитов схлопнуты в один",
				"git rebase -i --root — оставь первый как pick, остальные пометь squash, итоговое сообщение: Complete feature"),
			2: check(`[ -f /root/project/hotfix.txt ] && [ "$(git -C /root/project rev-parse --abbrev-ref HEAD)" = main ] && git -C /root/project log main --oneline | grep -qi 'hotfix'`,
				"коммит из hotfix перенесён в main",
				"git switch -c hotfix; echo fix > hotfix.txt; git add -A; git commit -m hotfix; git switch main; git cherry-pick hotfix"),
			3: check(`[ -s /root/reflog.txt ] && grep -qE 'HEAD@\{[0-9]+\}' /root/reflog.txt`,
				"reflog сохранён",
				"git -C /root/project reflog > /root/reflog.txt"),
			4: check(`[ -s /root/gitlog.txt ] && grep -qE '[0-9a-f]{7}' /root/gitlog.txt`,
				"граф веток сохранён",
				"git -C /root/project log --oneline --graph --all > /root/gitlog.txt"),
			5: check(`git -C /root/project diff --cached --name-only | grep -q . && [ -z "$(git -C /root/project log --oneline -1 --pretty=%s | grep -i hotfix)" ]`,
				"последний коммит откачен, изменения остались в staging",
				"git reset HEAD~1 --soft — после этого git status покажет файлы в staging"),
		},
	},

	// ── Lab 3: теги, stash, revert ──
	"ch-git-gym-lab3": {
		Setup: `set -e
rm -rf /root/project
` + gitInit + `
printf 'feature v1\n' > /root/project/feature.txt
git -C /root/project add -A
git -C /root/project commit -qm "add feature"
printf 'debug=1\n' > /root/project/bug.txt
git -C /root/project add -A
git -C /root/project commit -qm "bad commit: debug code"`,
		Checks: map[int]string{
			1: check(`[ "$(git -C /root/project cat-file -t v1.0.0 2>/dev/null)" = commit ]`,
				"lightweight-тег v1.0.0 создан",
				"git tag v1.0.0 (без -a — это и есть lightweight-тег)"),
			2: check(`[ "$(git -C /root/project cat-file -t v2.0.0 2>/dev/null)" = tag ] && git -C /root/project tag -n1 v2.0.0 | grep -q 'Release 2.0'`,
				"annotated-тег v2.0.0 создан",
				"git tag -a v2.0.0 -m 'Release 2.0'"),
			3: check(`[ -z "$(git -C /root/project status --porcelain 2>/dev/null)" ] && git -C /root/project stash list | grep -q . && git -C /root/project stash show -p 2>/dev/null | grep -q wip`,
				"изменение спрятано в stash, дерево чистое",
				"echo wip >> feature.txt && git stash — в stash должна попасть именно строка wip"),
			4: check(`grep -q '^wip$' /root/project/feature.txt 2>/dev/null && [ -z "$(git -C /root/project stash list 2>/dev/null)" ]`,
				"изменение возвращено из stash",
				"git stash pop — строка wip должна вернуться в feature.txt"),
			5: check(`[ ! -e /root/project/bug.txt ] && git -C /root/project log --oneline | grep -qi 'revert' && git -C /root/project log --oneline | grep -q 'bad commit'`,
				"плохой коммит отменён, история сохранена",
				"git revert --no-edit HEAD — bug.txt исчезнет, но оба коммита останутся в истории"),
		},
	},
}
