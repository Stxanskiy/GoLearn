package main

// ════════════════════════════════════════════════════════════════
// Git — дополнительные уроки (4-5)
// ════════════════════════════════════════════════════════════════

func lesson_git_remote() L {
	return L{
		Slug: "git-remote", Title: "Remote: GitHub и командная работа", Order: 4,
		Difficulty: "intermediate", Track: "shared",
		Content: `<h1>Remote — работа с GitHub</h1>

<h2>Remote = удалённый репозиторий</h2>
<pre><code># origin — стандартное имя для основного remote
git remote -v                    # показать remotes
git remote add origin URL        # подключить remote

# Клонирование — скачать репозиторий
git clone https://github.com/user/repo.git
cd repo</code></pre>

<h2>Push и Pull</h2>
<pre><code># Push — отправить коммиты в remote
git push origin main             # отправить ветку main
git push -u origin feature       # -u запоминает upstream
git push                         # после -u можно без аргументов

# Pull — получить и объединить
git pull                         # = fetch + merge
git pull --rebase                # = fetch + rebase (чище история)

# Fetch — только скачать, без merge
git fetch origin                 # скачать все изменения
git log origin/main..main        # что у меня есть, чего нет в remote</code></pre>

<h2>Pull Request (GitHub)</h2>
<pre><code># 1. Создай ветку
git checkout -b feature/add-auth

# 2. Коммить и пуш
git add . && git commit -m "feat: add JWT auth"
git push -u origin feature/add-auth

# 3. На GitHub: "Compare & Pull Request"
# 4. Код-ревью → Approve → Merge
# 5. Удалить ветку после merge:
git checkout main && git pull
git branch -d feature/add-auth</code></pre>

<h2>Fork и Contribute</h2>
<pre><code># 1. Fork — копия чужого репозитория в твой аккаунт
# 2. Clone свой fork
git clone https://github.com/YOU/repo.git
# 3. Добавь upstream (оригинал)
git remote add upstream https://github.com/ORIGINAL/repo.git
# 4. Синхронизация
git fetch upstream
git merge upstream/main</code></pre>

<h2>Частые проблемы</h2>
<pre><code># Rejected push (кто-то запушил раньше)
git pull --rebase && git push

# Удалить remote ветку
git push origin --delete old-branch

# Посмотреть что изменилось в remote
git fetch && git log ..origin/main</code></pre>`,

		Quiz: []Q{
			{Question: "Что делает git push -u origin feature?", Options: []string{"Удаляет ветку", "Пушит ветку и запоминает upstream — потом можно git push без аргументов", "Создаёт PR", "Мержит"}, Correct: 1, Explanation: "-u (--set-upstream) связывает локальную ветку с remote. После этого git push/pull работают без указания remote и branch."},
			{Question: "git pull vs git fetch?", Options: []string{"Одно и то же", "fetch только скачивает. pull = fetch + merge (или rebase)", "pull быстрее", "fetch удаляет"}, Correct: 1, Explanation: "fetch безопасен — только скачивает, ничего не меняет. pull сразу мержит. Для контроля: fetch → просмотр → merge вручную."},
			{Question: "Что такое Fork на GitHub?", Options: []string{"Ветка", "Полная копия чужого репозитория в твой аккаунт — для contribute без прав на оригинал", "Pull Request", "Clone"}, Correct: 1, Explanation: "Fork = твоя копия. Работаешь в ней, пушишь в неё. Потом PR из форка в оригинальный репо. Стандартный flow для open source."},
			{Question: "Почему git push rejected?", Options: []string{"Баг Git", "Remote содержит коммиты, которых нет локально — сначала pull, потом push", "Нет интернета", "Ветка защищена"}, Correct: 1, Explanation: "Non-fast-forward: remote ушёл вперёд. Решение: git pull --rebase (ставит твои коммиты поверх remote) → git push."},
			{Question: "Как синхронизировать fork с оригиналом?", Options: []string{"Удалить и заново fork", "git fetch upstream && git merge upstream/main", "git pull origin", "GitHub кнопка Sync"}, Correct: 1, Explanation: "upstream = оригинальный репо. fetch скачивает его изменения. merge/rebase вливает в твою ветку. Потом push в свой fork."},
		},
		Tasks: []T{
			{Title: "Git remote команды", Difficulty: "easy", Description: `<p>Сгенерируй последовательность команд для: клонировать, создать ветку, запушить:</p><p>Ввод: <code>https://github.com/user/repo.git feature/auth</code></p><p>Вывод:</p><pre><code>git clone https://github.com/user/repo.git
cd repo
git checkout -b feature/auth
git push -u origin feature/auth</code></pre>`, Glossary: []GlossaryItem{{Term: "git clone URL", Definition: "Скачать весь репозиторий с историей. Автоматически настраивает origin."}}, TestCases: []TestCase{{Input: "https://github.com/user/repo.git feature/auth", ExpectedOutput: "git clone https://github.com/user/repo.git\ncd repo\ngit checkout -b feature/auth\ngit push -u origin feature/auth"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var url, branch string; fmt.Scan(&url, &branch)
    parts := strings.Split(url, "/"); repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
    fmt.Printf("git clone %s\ncd %s\ngit checkout -b %s\ngit push -u origin %s\n", url, repo, branch, branch)
}`, Hints: `<p>Извлеки имя репо из URL. Последовательность: clone → cd → checkout -b → push -u.</p>`, Solution: `<pre><code>package main
import ("fmt"; "strings")
func main() { var u, b string; fmt.Scan(&u, &b); p := strings.Split(u, "/"); r := strings.TrimSuffix(p[len(p)-1], ".git"); fmt.Printf("git clone %s\ncd %s\ngit checkout -b %s\ngit push -u origin %s\n", u, r, b, b) }</code></pre>`},
			{Title: "PR workflow", Difficulty: "easy", Description: `<p>Сгенерируй полный PR workflow:</p><p>Ввод: <code>fix/login-bug</code></p><p>Вывод:</p><pre><code>git checkout -b fix/login-bug
git add .
git commit -m "fix: login bug"
git push -u origin fix/login-bug</code></pre>`, Glossary: []GlossaryItem{{Term: "PR workflow", Definition: "branch → commit → push → PR → review → merge → delete branch."}}, TestCases: []TestCase{{Input: "fix/login-bug", ExpectedOutput: "git checkout -b fix/login-bug\ngit add .\ngit commit -m \"fix: login bug\"\ngit push -u origin fix/login-bug"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var branch string; fmt.Scan(&branch)
    parts := strings.SplitN(branch, "/", 2); prefix := parts[0]; msg := strings.ReplaceAll(parts[1], "-", " ")
    fmt.Printf("git checkout -b %s\ngit add .\ngit commit -m \"%s: %s\"\ngit push -u origin %s\n", branch, prefix, msg, branch)
}`, Hints: `<p>Из branch name извлеки тип (fix/feat) и описание для commit message.</p>`, Solution: `<pre><code>package main
import ("fmt"; "strings")
func main() { var b string; fmt.Scan(&b); p := strings.SplitN(b, "/", 2); fmt.Printf("git checkout -b %s\ngit add .\ngit commit -m \"%s: %s\"\ngit push -u origin %s\n", b, p[0], strings.ReplaceAll(p[1], "-", " "), b) }</code></pre>`},
			{Title: "Fork sync", Difficulty: "medium", Description: `<p>Сгенерируй команды для синхронизации fork с upstream:</p><p>Ввод: <code>https://github.com/original/repo.git main</code></p><p>Вывод:</p><pre><code>git remote add upstream https://github.com/original/repo.git
git fetch upstream
git merge upstream/main
git push origin main</code></pre>`, Glossary: []GlossaryItem{{Term: "upstream", Definition: "Оригинальный репозиторий (от которого ты сделал fork)."}}, TestCases: []TestCase{{Input: "https://github.com/original/repo.git main", ExpectedOutput: "git remote add upstream https://github.com/original/repo.git\ngit fetch upstream\ngit merge upstream/main\ngit push origin main"}},
				StarterCode: `package main
import "fmt"
func main() {
    var url, branch string; fmt.Scan(&url, &branch)
    fmt.Printf("git remote add upstream %s\ngit fetch upstream\ngit merge upstream/%s\ngit push origin %s\n", url, branch, branch)
}`, Hints: `<p>upstream = оригинал. fetch → merge → push в свой fork.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var u, b string; fmt.Scan(&u, &b); fmt.Printf("git remote add upstream %s\ngit fetch upstream\ngit merge upstream/%s\ngit push origin %s\n", u, b, b) }</code></pre>`},
			{Title: "Resolve rejected push", Difficulty: "medium", Description: `<p>Push отклонён. Сгенерируй fix:</p><p>Ввод: <code>main</code></p><p>Вывод:</p><pre><code>git pull --rebase origin main
git push origin main</code></pre>`, Glossary: []GlossaryItem{{Term: "--rebase", Definition: "Ставит твои коммиты поверх remote. Чище чем merge commit."}}, TestCases: []TestCase{{Input: "main", ExpectedOutput: "git pull --rebase origin main\ngit push origin main"}},
				StarterCode: `package main
import "fmt"
func main() { var branch string; fmt.Scan(&branch); fmt.Printf("git pull --rebase origin %s\ngit push origin %s\n", branch, branch) }`, Hints: `<p>pull --rebase → push. Ребейс ставит твои коммиты поверх чужих.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var b string; fmt.Scan(&b); fmt.Printf("git pull --rebase origin %s\ngit push origin %s\n", b, b) }</code></pre>`},
			{Title: "Commit message parser", Difficulty: "hard", Description: `<p>Парси Conventional Commits и валидируй:</p><p>Ввод:</p><pre><code>3
feat: add user authentication
fixed login bug
docs: update README</code></pre><p>Вывод:</p><pre><code>feat: add user authentication - VALID
fixed login bug - INVALID (no type prefix)
docs: update README - VALID</code></pre>`, Glossary: []GlossaryItem{{Term: "Conventional Commits", Definition: "type: description. Types: feat, fix, docs, chore, refactor, test."}}, TestCases: []TestCase{{Input: "3\nfeat: add user authentication\nfixed login bug\ndocs: update README", ExpectedOutput: "feat: add user authentication - VALID\nfixed login bug - INVALID (no type prefix)\ndocs: update README - VALID"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    types := map[string]bool{"feat":true,"fix":true,"docs":true,"chore":true,"refactor":true,"test":true,"style":true,"perf":true}
    for i := 0; i < n; i++ { scanner.Scan(); msg := scanner.Text()
        parts := strings.SplitN(msg, ": ", 2)
        if len(parts) == 2 && types[parts[0]] { fmt.Printf("%s - VALID\n", msg) } else { fmt.Printf("%s - INVALID (no type prefix)\n", msg) }
    }
}`, Hints: `<p>Split по ": ". Проверяй что первая часть — валидный тип (feat, fix, docs...).</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin); ts := map[string]bool{"feat":true,"fix":true,"docs":true,"chore":true,"refactor":true,"test":true,"style":true,"perf":true}
    for i := 0; i < n; i++ { sc.Scan(); m := sc.Text(); p := strings.SplitN(m, ": ", 2)
        if len(p) == 2 && ts[p[0]] { fmt.Printf("%s - VALID\n", m) } else { fmt.Printf("%s - INVALID (no type prefix)\n", m) } } }</code></pre>`},
		},
	}
}

func lesson_git_advanced() L {
	return L{
		Slug: "git-advanced", Title: "Git Flow и продвинутые техники", Order: 5,
		Difficulty: "advanced", Track: "shared",
		Content: `<h1>Git Flow и продвинутые техники</h1>

<h2>Git Flow — стратегия ветвления</h2>
<pre><code># main — продакшен (всегда стабильный)
# develop — интеграционная ветка
# feature/* — новые фичи (от develop)
# release/* — подготовка релиза
# hotfix/* — срочные фиксы (от main)

# Workflow:
# 1. feature/auth от develop → работа → PR в develop
# 2. develop стабилен → release/1.0 → тестирование → merge в main + develop
# 3. Баг на проде → hotfix/fix-login от main → merge в main + develop</code></pre>

<h2>Trunk-Based Development (альтернатива)</h2>
<pre><code># Все работают в main (или short-lived branches <2 дней)
# Feature flags вместо долгих веток
# CI/CD деплоит каждый merge в main
# Подходит для: маленькие команды, быстрый деплой</code></pre>

<h2>Interactive Rebase — чистая история</h2>
<pre><code># Объединить последние 3 коммита в один:
git rebase -i HEAD~3
# pick abc123 first commit
# squash def456 WIP                ← объединить с предыдущим
# squash ghi789 fix typo           ← объединить

# Переименовать коммит:
# reword abc123 old message</code></pre>

<h2>Stash — временное хранилище</h2>
<pre><code>git stash              # спрятать незакоммиченные изменения
git stash list         # список stash-ей
git stash pop          # достать последний
git stash drop         # удалить последний

# Ситуация: работаешь над фичей, нужно срочно в другую ветку
git stash && git checkout hotfix && ... && git checkout feature && git stash pop</code></pre>

<h2>Cherry-pick — взять один коммит</h2>
<pre><code># Скопировать конкретный коммит из другой ветки
git cherry-pick abc123

# Ситуация: hotfix сделан в main, нужен и в develop
git checkout develop
git cherry-pick <hotfix-commit-hash></code></pre>

<h2>Bisect — найти баг</h2>
<pre><code># Бинарный поиск коммита, который сломал тесты
git bisect start
git bisect bad              # текущий — сломан
git bisect good v1.0        # v1.0 — работал
# Git переключает на середину → тестируешь → good/bad
# За log2(N) шагов находишь точный коммит</code></pre>

<h2>.gitignore и .gitattributes</h2>
<pre><code># .gitignore
.env            # секреты
*.exe           # бинарники
vendor/         # зависимости (если не vendoring)
.idea/          # IDE
tmp/            # временные файлы

# .gitattributes — нормализация line endings
* text=auto
*.go text eol=lf
*.sh text eol=lf</code></pre>`,

		Quiz: []Q{
			{Question: "Git Flow vs Trunk-Based — когда что?", Options: []string{"Git Flow всегда", "Git Flow для больших команд с релизными циклами. Trunk-Based для быстрого деплоя маленькими командами", "Trunk-Based устарел", "Нет разницы"}, Correct: 1, Explanation: "Git Flow: предсказуемые релизы, много разработчиков, QA-процесс. Trunk-Based: CI/CD, feature flags, 1-2 дня на фичу, деплой каждый merge."},
			{Question: "Что делает git stash?", Options: []string{"Удаляет изменения", "Временно прячет незакоммиченные изменения — можно переключить ветку и вернуть позже", "Коммитит", "Пушит"}, Correct: 1, Explanation: "Stash = карман. Спрятал → переключился → сделал дело → вернулся → stash pop. Без коммита грязных изменений."},
			{Question: "git cherry-pick — зачем?", Options: []string{"Удалить коммит", "Скопировать один конкретный коммит из другой ветки в текущую", "Создать ветку", "Отменить merge"}, Correct: 1, Explanation: "Cherry-pick = 'возьми только этот один коммит'. Hotfix из main нужен и в develop? cherry-pick. Не нужен целый merge всей ветки."},
			{Question: "git bisect — что это?", Options: []string{"Разделить ветку", "Бинарный поиск коммита, который сломал код — за log2(N) шагов", "Удалить половину коммитов", "Merge конфликт"}, Correct: 1, Explanation: "100 коммитов, где-то сломалось. Bisect: 7 шагов (log2(100)) вместо 100. Git переключает на середину, ты говоришь good/bad."},
			{Question: "Почему .env в .gitignore?", Options: []string{"Файл большой", "Содержит секреты (пароли, ключи) — нельзя коммитить в репозиторий", "Не нужен", "Git не поддерживает"}, Correct: 1, Explanation: "Секреты в git = утечка. Даже если удалишь потом — останется в истории. .env.example с пустыми значениями — коммитить. .env — в .gitignore."},
		},
		Tasks: []T{
			{Title: "Git Flow commands", Difficulty: "easy", Description: `<p>Сгенерируй Git Flow для новой фичи:</p><p>Ввод: <code>auth</code></p><p>Вывод:</p><pre><code>git checkout develop
git checkout -b feature/auth
git add .
git commit -m "feat: auth"
git push -u origin feature/auth</code></pre>`, Glossary: []GlossaryItem{{Term: "feature/*", Definition: "Ветка фичи. Создаётся от develop, мержится в develop."}}, TestCases: []TestCase{{Input: "auth", ExpectedOutput: "git checkout develop\ngit checkout -b feature/auth\ngit add .\ngit commit -m \"feat: auth\"\ngit push -u origin feature/auth"}},
				StarterCode: `package main
import "fmt"
func main() { var name string; fmt.Scan(&name); fmt.Printf("git checkout develop\ngit checkout -b feature/%s\ngit add .\ngit commit -m \"feat: %s\"\ngit push -u origin feature/%s\n", name, name, name) }`, Hints: `<p>Шаблон Git Flow: develop → feature/ → commit → push.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; fmt.Scan(&n); fmt.Printf("git checkout develop\ngit checkout -b feature/%s\ngit add .\ngit commit -m \"feat: %s\"\ngit push -u origin feature/%s\n", n, n, n) }</code></pre>`},
			{Title: "Stash workflow", Difficulty: "easy", Description: `<p>Сгенерируй stash workflow:</p><p>Ввод: <code>hotfix/urgent</code></p><p>Вывод:</p><pre><code>git stash
git checkout hotfix/urgent
git checkout -
git stash pop</code></pre>`, Glossary: []GlossaryItem{{Term: "git stash", Definition: "Прячет changes. pop — достаёт обратно. git checkout - = предыдущая ветка."}}, TestCases: []TestCase{{Input: "hotfix/urgent", ExpectedOutput: "git stash\ngit checkout hotfix/urgent\ngit checkout -\ngit stash pop"}},
				StarterCode: `package main
import "fmt"
func main() { var branch string; fmt.Scan(&branch); fmt.Printf("git stash\ngit checkout %s\ngit checkout -\ngit stash pop\n", branch) }`, Hints: `<p>stash → switch → work → switch back → pop.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var b string; fmt.Scan(&b); fmt.Printf("git stash\ngit checkout %s\ngit checkout -\ngit stash pop\n", b) }</code></pre>`},
			{Title: ".gitignore generator", Difficulty: "medium", Description: `<p>Сгенерируй .gitignore по языку:</p><p>Ввод: <code>go</code></p><p>Вывод:</p><pre><code>.env
*.exe
/vendor/
.idea/
tmp/
*.test
coverage.out</code></pre>`, Glossary: []GlossaryItem{{Term: ".gitignore", Definition: "Шаблоны файлов которые Git игнорирует."}}, TestCases: []TestCase{{Input: "go", ExpectedOutput: ".env\n*.exe\n/vendor/\n.idea/\ntmp/\n*.test\ncoverage.out"}},
				StarterCode: `package main
import "fmt"
func main() {
    var lang string; fmt.Scan(&lang)
    ignores := map[string][]string{
        "go": {".env", "*.exe", "/vendor/", ".idea/", "tmp/", "*.test", "coverage.out"},
    }
    for _, ig := range ignores[lang] { fmt.Println(ig) }
}`, Hints: `<p>Map с шаблонами по языку.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var l string; fmt.Scan(&l); for _, ig := range []string{".env", "*.exe", "/vendor/", ".idea/", "tmp/", "*.test", "coverage.out"} { fmt.Println(ig) } }</code></pre>`},
			{Title: "Hotfix workflow", Difficulty: "medium", Description: `<p>Полный hotfix: от main, fix, merge в main и develop:</p><p>Ввод: <code>fix-crash v1.2.1</code></p><p>Вывод:</p><pre><code>git checkout main
git checkout -b hotfix/fix-crash
git add .
git commit -m "fix: fix-crash"
git checkout main
git merge hotfix/fix-crash
git tag v1.2.1
git checkout develop
git merge hotfix/fix-crash
git branch -d hotfix/fix-crash</code></pre>`, Glossary: []GlossaryItem{{Term: "hotfix/*", Definition: "Срочный фикс от main. Мержится и в main и в develop."}}, TestCases: []TestCase{{Input: "fix-crash v1.2.1", ExpectedOutput: "git checkout main\ngit checkout -b hotfix/fix-crash\ngit add .\ngit commit -m \"fix: fix-crash\"\ngit checkout main\ngit merge hotfix/fix-crash\ngit tag v1.2.1\ngit checkout develop\ngit merge hotfix/fix-crash\ngit branch -d hotfix/fix-crash"}},
				StarterCode: `package main
import "fmt"
func main() {
    var name, tag string; fmt.Scan(&name, &tag)
    fmt.Printf("git checkout main\ngit checkout -b hotfix/%s\ngit add .\ngit commit -m \"fix: %s\"\ngit checkout main\ngit merge hotfix/%s\ngit tag %s\ngit checkout develop\ngit merge hotfix/%s\ngit branch -d hotfix/%s\n", name, name, name, tag, name, name)
}`, Hints: `<p>hotfix от main → fix → merge в main (+ tag) → merge в develop → delete branch.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n, t string; fmt.Scan(&n, &t); fmt.Printf("git checkout main\ngit checkout -b hotfix/%s\ngit add .\ngit commit -m \"fix: %s\"\ngit checkout main\ngit merge hotfix/%s\ngit tag %s\ngit checkout develop\ngit merge hotfix/%s\ngit branch -d hotfix/%s\n", n, n, n, t, n, n) }</code></pre>`},
			{Title: "Branch strategy analyzer", Difficulty: "hard", Description: `<p>Определи стратегию по описанию команды:</p><p>Ввод: <code>10 devs, 2-week sprints, QA team, scheduled releases</code></p><p>Вывод: <code>Recommended: Git Flow (large team, release cycles)</code></p><p>Ввод: <code>3 devs, deploy daily, feature flags, CI/CD</code></p><p>Вывод: <code>Recommended: Trunk-Based (small team, continuous deploy)</code></p>`, Glossary: []GlossaryItem{{Term: "Branch strategy", Definition: "Git Flow = release cycles. Trunk-Based = continuous deploy."}}, TestCases: []TestCase{{Input: "10 devs, 2-week sprints, QA team, scheduled releases", ExpectedOutput: "Recommended: Git Flow (large team, release cycles)"}, {Input: "3 devs, deploy daily, feature flags, CI/CD", ExpectedOutput: "Recommended: Trunk-Based (small team, continuous deploy)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); desc := strings.ToLower(scanner.Text())
    if strings.Contains(desc, "daily") || strings.Contains(desc, "trunk") || strings.Contains(desc, "3 dev") || strings.Contains(desc, "feature flag") {
        fmt.Println("Recommended: Trunk-Based (small team, continuous deploy)")
    } else {
        fmt.Println("Recommended: Git Flow (large team, release cycles)")
    }
}`, Hints: `<p>Ключевые слова: daily/feature flags → Trunk. Sprints/QA/scheduled → Git Flow.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { sc := bufio.NewScanner(os.Stdin); sc.Scan(); d := strings.ToLower(sc.Text())
    if strings.Contains(d, "daily") || strings.Contains(d, "feature flag") || strings.Contains(d, "3 dev") { fmt.Println("Recommended: Trunk-Based (small team, continuous deploy)") } else { fmt.Println("Recommended: Git Flow (large team, release cycles)") } }</code></pre>`},
		},
	}
}
