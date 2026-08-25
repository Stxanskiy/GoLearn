package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Git — с визуальным тренажёром
// ════════════════════════════════════════════════════════════════

func mod_git() M {
	return M{
		Slug:          "git",
		Title:         "Git — контроль версий",
		Description:   "Визуальный тренажёр: коммиты, ветки, merge, rebase, конфликты. Интерактивный граф прямо в браузере.",
		Order:         9, // после пакетов, перед http
		Track:         "shared",
		Difficulty:    "intermediate",
		Prerequisites: []string{"packages"},
		Lessons: []L{
			lesson_git_basics(),
			lesson_git_branches(),
			lesson_git_merge_rebase(),
			lesson_git_remote(),
			lesson_git_advanced(),
		},
	}
}

func lesson_git_basics() L {
	return L{
		Slug: "git-basics", Title: "Git: основы — коммиты и история", Order: 1,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Git — машина времени для кода</h1>

<h2>Что такое Git?</h2>
<p><strong>Git</strong> — система контроля версий. Она сохраняет историю изменений твоего кода. Ты можешь:</p>
<ul>
<li>Вернуться к любой прежней версии</li>
<li>Работать над разными фичами параллельно (ветки)</li>
<li>Объединять работу нескольких людей</li>
<li>Видеть кто, когда и что изменил</li>
</ul>

<h2>Три зоны Git</h2>
<p>У Git есть три "зоны" где живут файлы:</p>

<pre><code>Working Directory    →    Staging Area    →    Repository
(рабочая папка)         (подготовка)          (история)
                git add              git commit</code></pre>

<ol>
<li><strong>Working Directory</strong> — твоя папка с файлами, где ты пишешь код</li>
<li><strong>Staging Area</strong> (index) — "сцена", куда ты выбираешь что войдёт в следующий коммит</li>
<li><strong>Repository</strong> — хранилище всех коммитов (история)</li>
</ol>

<h2>Основные команды</h2>
<pre><code># Создать новый репозиторий
git init

# Посмотреть статус (что изменено, что готово к коммиту)
git status

# Добавить файл в staging area
git add main.go           # конкретный файл
git add .                 # все изменённые файлы

# Создать коммит (снимок)
git commit -m "Add user login"

# Посмотреть историю коммитов
git log --oneline
# a1b2c3d Add user login
# e4f5g6h Initial commit</code></pre>

<h2>Что такое коммит?</h2>
<p><strong>Коммит</strong> — снимок состояния всех файлов в определённый момент времени. Каждый коммит имеет:</p>
<ul>
<li><strong>Hash</strong> (ID) — уникальный идентификатор: <code>a1b2c3d</code></li>
<li><strong>Сообщение</strong> — описание: "Add user login"</li>
<li><strong>Автор</strong> — кто сделал коммит</li>
<li><strong>Родитель</strong> — ссылка на предыдущий коммит (цепочка)</li>
</ul>

<div class="git-trainer" data-exercise="basics-1"></div>

<h2>Хорошие сообщения коммитов</h2>
<pre><code># ПЛОХО
git commit -m "fix"
git commit -m "update"
git commit -m "changes"

# ХОРОШО — императив, конкретно
git commit -m "Fix login redirect for expired sessions"
git commit -m "Add rate limiting to API endpoints"
git commit -m "Remove unused UserService dependency"</code></pre>

<h2>.gitignore — что НЕ хранить в Git</h2>
<pre><code># .gitignore
*.exe          # скомпилированные файлы
.env           # секреты (пароли, ключи API)
vendor/        # зависимости (скачиваются через go mod)
node_modules/  # зависимости JavaScript
.idea/         # настройки IDE</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что делает git add?",
				Options:     []string{"Создаёт коммит", "Перемещает файлы в staging area (подготовка к коммиту)", "Отправляет код на сервер", "Скачивает файл"},
				Correct:     1,
				Explanation: "git add перемещает изменения из рабочей директории в staging area. Это как положить вещи в коробку перед отправкой. git commit — запечатать и отправить.",
			},
			{
				Question:    "Что такое коммит?",
				Options:     []string{"Файл с кодом", "Снимок состояния всех файлов в определённый момент + метаданные (автор, сообщение, hash)", "Ветка", "Удалённый сервер"},
				Correct:     1,
				Explanation: "Коммит = snapshot + metadata. Каждый коммит знает своего родителя, образуя цепочку истории.",
			},
			{
				Question:    "Почему .env файл добавляют в .gitignore?",
				Options:     []string{"Он слишком большой", "Он содержит секреты (пароли, API-ключи), которые нельзя хранить в репозитории", "Git не поддерживает .env", "Для скорости"},
				Correct:     1,
				Explanation: "Секреты в Git = утечка. Любой с доступом к репо увидит пароли. Даже если удалить позже — они остаются в истории. .gitignore предотвращает случайный коммит.",
			},
			{
				Question:    "Из чего состоит SHA-1 хеш коммита?",
				Options:     []string{"Только из кода", "Из содержимого файлов, дерева, автора, даты, сообщения и хеша родителя", "Из имени ветки", "Из номера коммита"},
				Correct:     1,
				Explanation: "SHA-1 хеш — криптографический отпечаток всех данных коммита. Изменение любого бита (даже пробел в сообщении) даёт другой хеш. Поэтому историю Git нельзя подделать.",
			},
			{
				Question:    "Чем staging area (index) отличается от рабочей директории?",
				Options:     []string{"Ничем", "Staging — промежуточная область: только добавленные через git add изменения попадут в следующий коммит", "Staging — это remote", "Staging хранит бинарные файлы"},
				Correct:     1,
				Explanation: "Три области Git: working dir (твои файлы) → staging/index (git add) → repository (git commit). Staging позволяет выборочно коммитить — не обязательно всё сразу.",
			},
		},
		Tasks: []T{
			{
				Title:      "Парсер git log",
				Difficulty: "easy",
				Glossary: []GlossaryItem{
					{Term: "strings.SplitN(s, sep, n)", Definition: "Разбивает строку макс. на n частей. SplitN(\"abc def ghi\", \" \", 2) → [\"abc\", \"def ghi\"]"},
					{Term: "bufio.NewScanner(os.Stdin)", Definition: "Создаёт сканер для построчного чтения из stdin."},
				},
				Description: `<p>На вход подаётся вывод <code>git log --oneline</code> — по одной строке: <code>hash сообщение</code>.</p>
<p>Для каждой строки выведи: <code>[hash] сообщение</code></p>
<p><em>Пример входа:</em> <code>a1b2c3d Add user login</code></p>
<p><em>Выход:</em> <code>[a1b2c3d] Add user login</code></p>`,
				Hints: `<p>Используй <code>strings.SplitN(line, " ", 2)</code> — разобьёт на hash и остаток сообщения.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		fmt.Printf("[%s] %s\n", parts[0], parts[1])
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: разбей line на hash и message
		// Используй strings.SplitN(line, " ", 2)
		// Выведи в формате [hash] message
		_ = line
	}
}`,
				TestCases: []TestCase{
					{Input: "a1b2c3d Add user login\ne4f5g6h Initial commit\n1234abc Fix redirect bug", ExpectedOutput: "[a1b2c3d] Add user login\n[e4f5g6h] Initial commit\n[1234abc] Fix redirect bug"},
				},
			},
			{
				Title:      "Фильтр .gitignore",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "strings.HasSuffix(s, suffix)", Definition: "Проверяет заканчивается ли строка суффиксом. Для матчинга расширений вроде *.exe."},
					{Term: "strings.HasPrefix(s, prefix)", Definition: "Проверяет начинается ли строка с prefix. Для матчинга директорий вроде vendor/."},
					{Term: "strings.TrimSuffix(s, suffix)", Definition: "Убирает суффикс из строки если он есть."},
				},
				Description: `<p>Реализуй упрощённый .gitignore.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — число N (количество паттернов)</li>
<li>Следующие N строк — паттерны gitignore</li>
<li>Остальные строки — пути файлов для проверки</li>
</ol>
<p><strong>Правила матчинга (упрощённые):</strong></p>
<ul>
<li><code>*.ext</code> — файлы с расширением .ext</li>
<li><code>dirname/</code> — директория (совпадает если путь начинается с dirname/)</li>
<li><code>filename</code> — точное имя файла (последний сегмент пути)</li>
</ul>
<p>Для каждого файла выведи <code>IGNORED</code> или <code>TRACKED</code>.</p>`,
				Hints: `<p>Для каждого паттерна проверь тип: если начинается с * — расширение, если заканчивается / — директория, иначе — точное имя. Для файлов используй последний сегмент пути (после последнего /).</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	patterns := make([]string, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		patterns[i] = scanner.Text()
	}

	for scanner.Scan() {
		file := scanner.Text()
		ignored := false
		// Extract filename (last segment)
		parts := strings.Split(file, "/")
		filename := parts[len(parts)-1]

		for _, p := range patterns {
			if strings.HasPrefix(p, "*") {
				ext := p[1:] // e.g. ".exe"
				if strings.HasSuffix(filename, ext) {
					ignored = true
				}
			} else if strings.HasSuffix(p, "/") {
				dir := strings.TrimSuffix(p, "/")
				if strings.HasPrefix(file, dir+"/") || parts[0] == dir {
					ignored = true
				}
			} else {
				if filename == p {
					ignored = true
				}
			}
		}
		if ignored {
			fmt.Println("IGNORED")
		} else {
			fmt.Println("TRACKED")
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Считай количество паттернов
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	// Считай паттерны
	patterns := make([]string, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		patterns[i] = scanner.Text()
	}

	// Проверь каждый файл
	for scanner.Scan() {
		file := scanner.Text()
		parts := strings.Split(file, "/")
		filename := parts[len(parts)-1]

		ignored := false
		// TODO: проверь file против каждого паттерна
		// *.ext → strings.HasSuffix(filename, ext)
		// dir/ → strings.HasPrefix(file, dir+"/")
		// name → filename == name

		_ = filename
		_ = patterns

		if ignored {
			fmt.Println("IGNORED")
		} else {
			fmt.Println("TRACKED")
		}
	}
}`,
				TestCases: []TestCase{
					{Input: "3\n*.exe\nvendor/\n.env\nmain.go\napp.exe\nvendor/chi/chi.go\n.env\nREADME.md", ExpectedOutput: "TRACKED\nIGNORED\nIGNORED\nIGNORED\nTRACKED"},
					{Input: "2\n*.log\n.idea/\nserver.log\nsrc/main.go\n.idea/workspace.xml\napp.log", ExpectedOutput: "IGNORED\nTRACKED\nIGNORED\nIGNORED"},
				},
			},
			{
				Title:      "Генератор commit message",
				Difficulty: "easy",
				Description: `<p>По типу изменения и описанию сгенерируй commit message по конвенции Conventional Commits:</p>
<p>Ввод:</p><pre><code>3
feat add user login
fix null pointer in handler
docs update README</code></pre>
<p>Вывод:</p><pre><code>feat: add user login
fix: null pointer in handler
docs: update README</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Conventional Commits", Definition: "Стандарт: type: description. Типы: feat, fix, docs, refactor, test, chore."},
				},
				TestCases: []TestCase{
					{Input: "3\nfeat add user login\nfix null pointer in handler\ndocs update README", ExpectedOutput: "feat: add user login\nfix: null pointer in handler\ndocs: update README"},
				},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        sc.Scan(); parts := strings.SplitN(sc.Text(), " ", 2)
        fmt.Printf("%s: %s\n", parts[0], parts[1])
    }
}`,
				Hints: `<p><code>strings.SplitN(line, " ", 2)</code> — разбить на тип и описание.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); p := strings.SplitN(sc.Text(), " ", 2); fmt.Printf("%s: %s\n", p[0], p[1]) }
}</code></pre>`,
			},
			{
				Title:      "Сокращение SHA хеша",
				Difficulty: "easy",
				Description: `<p>Git показывает короткие хеши (7 символов). Напиши функцию которая сокращает полный SHA до 7 символов и проверяет уникальность:</p>
<p>Ввод:</p><pre><code>3
abc1234567890abcdef1234567890abcdef123456
abc1234999990abcdef1234567890abcdef123456
def5678567890abcdef1234567890abcdef123456</code></pre>
<p>Вывод:</p><pre><code>abc1234 (collision!)
abc1234 (collision!)
def5678</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Short SHA", Definition: "Первые 7 символов полного SHA-1 (40 символов). Обычно уникальны в рамках проекта."},
				},
				TestCases: []TestCase{
					{Input: "3\nabc1234567890abcdef1234567890abcdef123456\nabc1234999990abcdef1234567890abcdef123456\ndef5678567890abcdef1234567890abcdef123456", ExpectedOutput: "abc1234 (collision!)\nabc1234 (collision!)\ndef5678"},
				},
				StarterCode: `package main
import ("bufio";"fmt";"os")
func main() {
    sc := bufio.NewScanner(os.Stdin); var n int; fmt.Scan(&n)
    seen := map[string]int{}; shorts := []string{}
    for i := 0; i < n; i++ { sc.Scan(); s := sc.Text()[:7]; seen[s]++; shorts = append(shorts, s) }
    for _, s := range shorts {
        if seen[s] > 1 { fmt.Println(s, "(collision!)") } else { fmt.Println(s) }
    }
}`,
				Hints: `<p><code>s[:7]</code> — первые 7 символов. Map для подсчёта коллизий.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os")
func main() {
    sc := bufio.NewScanner(os.Stdin); var n int; fmt.Scan(&n)
    seen := map[string]int{}; ss := []string{}
    for i := 0; i < n; i++ { sc.Scan(); s := sc.Text()[:7]; seen[s]++; ss = append(ss, s) }
    for _, s := range ss { if seen[s]>1{fmt.Println(s,"(collision!)")}else{fmt.Println(s)} }
}</code></pre>`,
			},
			{
				Title:      "Симулятор staging area",
				Difficulty: "medium",
				Description: `<p>Эмулируй три области Git: working dir, staging, committed. Команды: <code>edit FILE</code>, <code>add FILE</code>, <code>commit</code>, <code>status</code>:</p>
<p>Ввод:</p><pre><code>6
edit main.go
edit README.md
add main.go
status
commit
status</code></pre>
<p>Вывод:</p><pre><code>Modified: [README.md main.go]
Staged: [main.go]
Committed: []
Committed: [main.go]</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Working → Staging → Committed", Definition: "edit → modified. add → staged. commit → все staged файлы переходят в committed."},
				},
				TestCases: []TestCase{
					{Input: "6\nedit main.go\nedit README.md\nadd main.go\nstatus\ncommit\nstatus", ExpectedOutput: "Modified: [README.md main.go]\nStaged: [main.go]\nCommitted: []\nCommitted: [main.go]"},
				},
				StarterCode: `package main
import ("bufio";"fmt";"os";"sort";"strings")

func sorted(m map[string]bool) []string {
    var r []string; for k := range m { r = append(r, k) }; sort.Strings(r); return r
}

func main() {
    sc := bufio.NewScanner(os.Stdin); var n int; fmt.Scan(&n)
    modified := map[string]bool{}; staged := map[string]bool{}; committed := map[string]bool{}
    for i := 0; i < n; i++ {
        sc.Scan(); parts := strings.Fields(sc.Text())
        switch parts[0] {
        case "edit": modified[parts[1]] = true
        case "add":
            if modified[parts[1]] { staged[parts[1]] = true; delete(modified, parts[1]) }
        case "commit":
            for f := range staged { committed[f] = true }; staged = map[string]bool{}
        case "status":
            fmt.Printf("Modified: %v\nStaged: %v\n", sorted(modified), sorted(staged))
        }
    }
    fmt.Printf("Committed: %v\n", sorted(committed))
}`,
				Hints: `<p>Три map: modified, staged, committed. edit → modified. add → staged (из modified). commit → all staged → committed.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"sort";"strings")
func s(m map[string]bool) []string { var r []string; for k := range m{r=append(r,k)}; sort.Strings(r); return r }
func main() {
    sc:=bufio.NewScanner(os.Stdin); var n int; fmt.Scan(&n)
    mod:=map[string]bool{}; stg:=map[string]bool{}; com:=map[string]bool{}
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text())
        switch p[0]{case "edit":mod[p[1]]=true;case "add":if mod[p[1]]{stg[p[1]]=true;delete(mod,p[1])}
        case "commit":for f:=range stg{com[f]=true};stg=map[string]bool{}
        case "status":fmt.Printf("Modified: %v\nStaged: %v\n",s(mod),s(stg))}}
    fmt.Printf("Committed: %v\n",s(com))
}</code></pre>`,
			},
		},
	}
}

func lesson_git_branches() L {
	return L{
		Slug: "git-branches", Title: "Ветки — параллельная работа", Order: 2,
		Difficulty: "intermediate", Track: "shared",
		Content: `<h1>Ветки — параллельная работа</h1>

<h2>Что такое ветка?</h2>
<p><strong>Ветка</strong> (branch) — это независимая линия разработки. Представь дерево: ствол — это <code>main</code>, а ветки — фичи, над которыми ты работаешь параллельно.</p>

<div class="git-trainer" data-exercise="branches-1"></div>

<h2>Основные команды</h2>
<pre><code># Создать ветку
git branch feature/login

# Переключиться на ветку
git checkout feature/login
# или (Go 1.22+)
git switch feature/login

# Создать и сразу переключиться
git checkout -b feature/login
git switch -c feature/login

# Посмотреть все ветки
git branch
# * main
#   feature/login

# Удалить ветку (после merge)
git branch -d feature/login</code></pre>

<h2>Типичный workflow</h2>
<pre><code># 1. Создать ветку для фичи
git switch -c feature/add-rooms

# 2. Работать, коммитить
git add .
git commit -m "Add Room struct and handler"
git commit -m "Add room creation API"

# 3. Вернуться на main
git switch main

# 4. Слить ветку в main
git merge feature/add-rooms

# 5. Удалить ветку
git branch -d feature/add-rooms</code></pre>

<h2>Удалённые репозитории</h2>
<pre><code># Связать с GitHub
git remote add origin https://github.com/user/repo.git

# Отправить на GitHub
git push origin main
git push -u origin feature/login  # -u запоминает связь

# Скачать изменения
git pull origin main

# Клонировать чужой репозиторий
git clone https://github.com/user/repo.git</code></pre>

<div class="git-trainer" data-exercise="branches-2"></div>

<h2>Naming conventions для веток</h2>
<pre><code>feature/add-rooms    # новая фича
fix/login-redirect   # исправление бага
hotfix/security-fix  # срочное исправление
refactor/clean-api   # рефакторинг
docs/update-readme   # документация</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что делает git switch -c feature/x?",
				Options:     []string{"Удаляет ветку", "Создаёт новую ветку feature/x и переключается на неё", "Коммитит изменения", "Скачивает ветку"},
				Correct:     1,
				Explanation: "switch -c (create) = создать + переключиться. Эквивалент старого checkout -b. switch безопаснее — он только для веток.",
			},
			{
				Question:    "Зачем нужны ветки?",
				Options:     []string{"Для красоты", "Для параллельной работы: каждая фича в своей ветке, не мешает другим", "Для ускорения", "Git не работает без веток"},
				Correct:     1,
				Explanation: "Ветки изолируют работу. Ты можешь делать фичу в feature/x, коллега — в feature/y. Когда готово — сливаете в main. Без конфликтов (обычно).",
			},
			{
				Question:    "Что такое HEAD в Git?",
				Options:     []string{"Первый коммит", "Указатель на текущий коммит/ветку — то где ты сейчас находишься", "Remote сервер", "Тег версии"},
				Correct:     1,
				Explanation: "HEAD — символическая ссылка на текущую позицию. Обычно HEAD → ветка → коммит. При detached HEAD → напрямую на коммит.",
			},
			{
				Question:    "Чем git switch отличается от git checkout?",
				Options:     []string{"Ничем", "switch только для веток (безопаснее), checkout ещё и для файлов (может потерять данные)", "switch быстрее", "checkout устарел полностью"},
				Correct:     1,
				Explanation: "checkout перегружен: ветки + восстановление файлов. switch (2.23+) — только ветки. restore — только файлы. Разделение предотвращает случайную потерю данных.",
			},
			{
				Question:    "Что означает fast-forward при merge?",
				Options:     []string{"Ускоренная загрузка", "Целевая ветка просто передвигает указатель вперёд — нет merge-коммита, линейная история", "Конфликт", "Принудительный push"},
				Correct:     1,
				Explanation: "Fast-forward возможен когда main не менялась после ответвления. Git просто двигает указатель main на последний коммит фичи. Без лишнего merge-коммита.",
			},
		},
		Tasks: []T{
			{
				Title:      "Валидатор имени ветки",
				Difficulty: "easy",
				Glossary: []GlossaryItem{
					{Term: "strings.HasPrefix(s, p)", Definition: "Проверяет начинается ли строка s с префикса p."},
					{Term: "strings.ContainsAny(s, chars)", Definition: "Проверяет содержит ли строка любой из символов chars."},
				},
				Description: `<p>Проверь имена Git-веток на соответствие конвенциям.</p>
<p><strong>Правила:</strong></p>
<ul>
<li>Должно начинаться с одного из префиксов: <code>feature/</code>, <code>fix/</code>, <code>hotfix/</code>, <code>refactor/</code>, <code>docs/</code></li>
<li>После префикса должно быть хотя бы 1 символ</li>
<li>Не должно содержать пробелы</li>
</ul>
<p>Для каждой ветки выведи <code>OK</code> или <code>BAD: причина</code>.</p>`,
				Hints: `<p>Проверь HasPrefix для каждого допустимого префикса. Потом проверь пробелы через strings.Contains. Потом длину после префикса.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	prefixes := []string{"feature/", "fix/", "hotfix/", "refactor/", "docs/"}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()
		if strings.Contains(name, " ") {
			fmt.Println("BAD: пробелы запрещены")
			continue
		}
		found := false
		for _, p := range prefixes {
			if strings.HasPrefix(name, p) {
				if len(name) > len(p) {
					found = true
				} else {
					fmt.Println("BAD: пустое имя после префикса")
					found = true
				}
				break
			}
		}
		if !found {
			fmt.Println("BAD: неизвестный префикс")
		} else if len(name) > len("feature/") || len(name) > len("fix/") {
			// check passed already
		}
		if found && !strings.Contains(name, " ") {
			hasContent := false
			for _, p := range prefixes {
				if strings.HasPrefix(name, p) && len(name) > len(p) {
					hasContent = true
					break
				}
			}
			if hasContent {
				fmt.Println("OK")
			}
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	prefixes := []string{"feature/", "fix/", "hotfix/", "refactor/", "docs/"}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()

		// TODO: проверь правила:
		// 1. Нет пробелов → "BAD: пробелы запрещены"
		// 2. Начинается с одного из prefixes → иначе "BAD: неизвестный префикс"
		// 3. После префикса есть символы → иначе "BAD: пустое имя после префикса"
		// Если всё ок → "OK"
		_ = name
		_ = prefixes
	}
}`,
				TestCases: []TestCase{
					{Input: "feature/add-rooms\nfix/login-bug\nmain\nfeature/\nmy branch", ExpectedOutput: "OK\nOK\nBAD: неизвестный префикс\nBAD: пустое имя после префикса\nBAD: пробелы запрещены"},
				},
			},
			{
				Title:      "Симулятор коммитов",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "map[string][]string", Definition: "Словарь: ключ — имя ветки, значение — слайс хешей коммитов этой ветки."},
					{Term: "fmt.Sprintf(\"%07x\", n)", Definition: "Форматирует число как 7-значный hex. Для генерации фейковых commit hash."},
				},
				Description: `<p>Симулируй Git-операции. На вход — команды:</p>
<ul>
<li><code>commit &lt;message&gt;</code> — создать коммит в текущей ветке (hash = порядковый номер в hex, 7 символов)</li>
<li><code>branch &lt;name&gt;</code> — создать ветку от текущей (копирует все коммиты)</li>
<li><code>switch &lt;name&gt;</code> — переключиться на ветку</li>
<li><code>log</code> — вывести коммиты текущей ветки (от новых к старым, формат: <code>hash message</code>)</li>
</ul>
<p>Начальная ветка: <code>main</code>, счётчик с 1.</p>`,
				Hints: `<p>branches := map[string][]Commit. При branch — копируй слайс текущей ветки. При commit — append. При log — обратный порядок.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Commit struct {
	Hash, Message string
}

func main() {
	branches := map[string][]Commit{"main": {}}
	current := "main"
	counter := 1

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]

		switch cmd {
		case "commit":
			hash := fmt.Sprintf("%07x", counter)
			counter++
			branches[current] = append(branches[current], Commit{hash, parts[1]})
		case "branch":
			name := parts[1]
			src := branches[current]
			cp := make([]Commit, len(src))
			copy(cp, src)
			branches[name] = cp
		case "switch":
			current = parts[1]
		case "log":
			commits := branches[current]
			for i := len(commits) - 1; i >= 0; i-- {
				fmt.Printf("%s %s\n", commits[i].Hash, commits[i].Message)
			}
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Commit struct {
	Hash, Message string
}

func main() {
	branches := map[string][]Commit{"main": {}}
	current := "main"
	counter := 1

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]

		switch cmd {
		case "commit":
			// TODO: hash = fmt.Sprintf("%07x", counter), counter++
			// Добавь Commit{hash, parts[1]} в branches[current]
		case "branch":
			// TODO: скопируй коммиты текущей ветки в новую
			// Важно: copy, а не присваивание (иначе общий слайс!)
			_ = parts[1]
		case "switch":
			// TODO: переключись на ветку parts[1]
		case "log":
			// TODO: выведи коммиты в обратном порядке
			// Формат: "hash message"
		}
	}
}`,
				TestCases: []TestCase{
					{Input: "commit Initial commit\ncommit Add main.go\nbranch feature/login\nswitch feature/login\ncommit Add login handler\nlog", ExpectedOutput: "0000003 Add login handler\n0000002 Add main.go\n0000001 Initial commit"},
					{Input: "commit Init\nbranch dev\nswitch dev\ncommit Dev work\nswitch main\nlog", ExpectedOutput: "0000001 Init"},
				},
			},
			{
				Title:      "Визуализация веток",
				Difficulty: "easy",
				Description: `<p>Покажи какие ветки указывают на один коммит (одинаковый последний хеш). На вход: имя ветки и хеш последнего коммита:</p>
<p>Ввод:</p><pre><code>4
main abc1234
dev abc1234
feature/x def5678
hotfix/y abc1234</code></pre>
<p>Вывод:</p><pre><code>abc1234: main, dev, hotfix/y
def5678: feature/x</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "branch = pointer", Definition: "Ветка — это просто указатель на коммит. Несколько веток могут указывать на один коммит."},
				},
				TestCases: []TestCase{
					{Input: "4\nmain abc1234\ndev abc1234\nfeature/x def5678\nhotfix/y abc1234", ExpectedOutput: "abc1234: main, dev, hotfix/y\ndef5678: feature/x"},
				},
				StarterCode: `package main
import ("fmt";"sort";"strings")
func main() {
    var n int; fmt.Scan(&n)
    groups := map[string][]string{}; order := []string{}
    for i := 0; i < n; i++ {
        var name, hash string; fmt.Scan(&name, &hash)
        if _, ok := groups[hash]; !ok { order = append(order, hash) }
        groups[hash] = append(groups[hash], name)
    }
    for _, h := range order { fmt.Printf("%s: %s\n", h, strings.Join(groups[h], ", ")) }
    _ = sort.Strings
}`,
				Hints: `<p>Map hash → []branches. Сохраняй порядок первого появления хеша.</p>`,
				Solution: `<pre><code>package main
import ("fmt";"strings")
func main() {
    var n int; fmt.Scan(&n); g := map[string][]string{}; var o []string
    for i := 0; i < n; i++ { var nm, h string; fmt.Scan(&nm, &h); if _,ok:=g[h];!ok{o=append(o,h)}; g[h]=append(g[h],nm) }
    for _, h := range o { fmt.Printf("%s: %s\n", h, strings.Join(g[h], ", ")) }
}</code></pre>`,
			},
			{
				Title:      "Детектор stale веток",
				Difficulty: "medium",
				Description: `<p>Найди "заброшенные" ветки — те, в которых последний коммит был более N дней назад:</p>
<p>Ввод:</p><pre><code>30
4
feature/login 5
feature/old-api 45
hotfix/urgent 2
dev 31</code></pre>
<p>Вывод:</p><pre><code>Stale branches (>30 days):
  feature/old-api (45 days)
  dev (31 days)</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "stale branch", Definition: "Ветка без активности. На практике: git branch --sort=-committerdate показывает возраст."},
				},
				TestCases: []TestCase{
					{Input: "30\n4\nfeature/login 5\nfeature/old-api 45\nhotfix/urgent 2\ndev 31", ExpectedOutput: "Stale branches (>30 days):\n  feature/old-api (45 days)\n  dev (31 days)"},
				},
				StarterCode: `package main
import "fmt"
func main() {
    var threshold, n int; fmt.Scan(&threshold, &n)
    fmt.Printf("Stale branches (>%d days):\n", threshold)
    for i := 0; i < n; i++ {
        var name string; var days int; fmt.Scan(&name, &days)
        if days > threshold { fmt.Printf("  %s (%d days)\n", name, days) }
    }
}`,
				Hints: `<p>Просто сравни days > threshold для каждой ветки.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func main() { var t,n int; fmt.Scan(&t,&n); fmt.Printf("Stale branches (>%d days):\n",t)
    for i:=0;i<n;i++{var nm string;var d int;fmt.Scan(&nm,&d);if d>t{fmt.Printf("  %s (%d days)\n",nm,d)}} }</code></pre>`,
			},
			{
				Title:      "Fast-forward checker",
				Difficulty: "hard",
				Description: `<p>Определи можно ли сделать fast-forward merge. FF возможен если все коммиты main есть в начале feature (feature — продолжение main):</p>
<p>Ввод:</p><pre><code>main: A B C
feature: A B C D E</code></pre>
<p>Вывод: <code>Fast-forward: YES (2 new commits)</code></p>
<p>Ввод:</p><pre><code>main: A B C X
feature: A B C D E</code></pre>
<p>Вывод: <code>Fast-forward: NO (main has diverged)</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "fast-forward merge", Definition: "Возможен когда main — префикс feature. Git двигает указатель main вперёд без merge-коммита."},
				},
				TestCases: []TestCase{
					{Input: "main: A B C\nfeature: A B C D E", ExpectedOutput: "Fast-forward: YES (2 new commits)"},
					{Input: "main: A B C X\nfeature: A B C D E", ExpectedOutput: "Fast-forward: NO (main has diverged)"},
				},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); mainLine := strings.TrimPrefix(sc.Text(), "main: "); mainCommits := strings.Fields(mainLine)
    sc.Scan(); featLine := strings.TrimPrefix(sc.Text(), "feature: "); featCommits := strings.Fields(featLine)
    if len(mainCommits) > len(featCommits) { fmt.Println("Fast-forward: NO (main has diverged)"); return }
    for i, c := range mainCommits {
        if featCommits[i] != c { fmt.Println("Fast-forward: NO (main has diverged)"); return }
    }
    fmt.Printf("Fast-forward: YES (%d new commits)\n", len(featCommits)-len(mainCommits))
}`,
				Hints: `<p>FF возможен если main — префикс feature. Проверь каждый коммит main есть в начале feature.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() {
    sc:=bufio.NewScanner(os.Stdin); sc.Scan(); mc:=strings.Fields(strings.TrimPrefix(sc.Text(),"main: "))
    sc.Scan(); fc:=strings.Fields(strings.TrimPrefix(sc.Text(),"feature: "))
    if len(mc)>len(fc){fmt.Println("Fast-forward: NO (main has diverged)");return}
    for i,c:=range mc{if fc[i]!=c{fmt.Println("Fast-forward: NO (main has diverged)");return}}
    fmt.Printf("Fast-forward: YES (%d new commits)\n",len(fc)-len(mc))
}</code></pre>`,
			},
		},
	}
}

func lesson_git_merge_rebase() L {
	return L{
		Slug: "git-merge-rebase", Title: "Merge, Rebase и конфликты", Order: 3,
		Difficulty: "intermediate", Track: "shared",
		Content: `<h1>Merge, Rebase и конфликты</h1>

<h2>Merge — объединение веток</h2>
<p><strong>Merge</strong> берёт изменения из одной ветки и объединяет с другой:</p>

<pre><code># Находимся на main
git merge feature/login
# Git создаёт "merge commit" — точку слияния</code></pre>

<div class="git-trainer" data-exercise="merge-1"></div>

<h2>Fast-forward vs Merge commit</h2>
<ul>
<li><strong>Fast-forward</strong> — если main не менялся, Git просто сдвигает указатель вперёд (без merge commit)</li>
<li><strong>Merge commit</strong> — если обе ветки имеют новые коммиты, Git создаёт специальный коммит с двумя родителями</li>
</ul>

<h2>Rebase — перебазирование</h2>
<p><strong>Rebase</strong> переносит коммиты ветки на вершину другой ветки, создавая линейную историю:</p>

<pre><code># Находимся на feature/login
git rebase main
# Коммиты feature/login "переигрываются" поверх main</code></pre>

<div class="git-trainer" data-exercise="rebase-1"></div>

<p><strong>Merge vs Rebase:</strong></p>
<ul>
<li><strong>Merge</strong> — сохраняет полную историю, создаёт merge commit. Безопасен.</li>
<li><strong>Rebase</strong> — создаёт линейную историю (чище). Опасен для общих веток!</li>
</ul>

<p><strong>Золотое правило:</strong> никогда не делай rebase на публичных ветках (main, develop), которые используют другие люди.</p>

<h2>Конфликты</h2>
<p>Конфликт возникает когда <strong>две ветки изменили одну и ту же строку</strong>. Git не знает какую версию выбрать:</p>

<pre><code><<<<<<< HEAD
fmt.Println("Hello from main")
=======
fmt.Println("Hello from feature")
>>>>>>> feature/login</code></pre>

<p>Как решить:</p>
<ol>
<li>Открой файл с конфликтом</li>
<li>Удали маркеры (<code>&lt;&lt;&lt;&lt;&lt;&lt;&lt;</code>, <code>=======</code>, <code>&gt;&gt;&gt;&gt;&gt;&gt;&gt;</code>)</li>
<li>Оставь правильный вариант (или объедини оба)</li>
<li><code>git add файл</code> + <code>git commit</code></li>
</ol>

<h2>Полезные команды</h2>
<pre><code># Отменить последний коммит (сохранить изменения)
git reset --soft HEAD~1

# Посмотреть разницу
git diff                 # рабочая директория vs staging
git diff --staged        # staging vs последний коммит

# Спрятать изменения на время
git stash                # спрятать
git stash pop            # достать обратно

# Интерактивный лог с графом
git log --oneline --graph --all</code></pre>`,

		Quiz: []Q{
			{
				Question:    "В чём разница между merge и rebase?",
				Options:     []string{"Ничем", "Merge создаёт merge commit и сохраняет историю; rebase переписывает историю в линейную", "Merge быстрее", "Rebase безопаснее"},
				Correct:     1,
				Explanation: "Merge — безопасное слияние с сохранением полной истории. Rebase — переписывает коммиты поверх другой ветки, создавая чистую линейную историю.",
			},
			{
				Question:    "Когда возникает конфликт?",
				Options:     []string{"При каждом merge", "Когда две ветки изменили одну и ту же строку в одном файле", "Когда файл слишком большой", "При push"},
				Correct:     1,
				Explanation: "Конфликт = два изменения в одном месте. Git не знает какое оставить и просит разрешить вручную.",
			},
			{
				Question:    "Что делает git stash?",
				Options:     []string{"Удаляет файлы", "Временно прячет незакоммиченные изменения (можно достать обратно через stash pop)", "Создаёт ветку", "Отменяет коммит"},
				Correct:     1,
				Explanation: "stash — 'карман'. Спрятал изменения, переключился на другую ветку, сделал дела, вернулся, достал обратно (stash pop).",
			},
			{
				Question:    "Почему нельзя rebase публичные ветки (main)?",
				Options:     []string{"Git запрещает", "Rebase переписывает хеши коммитов — у коллег будут конфликты, т.к. их копия истории не совпадёт", "Rebase медленнее", "Можно без проблем"},
				Correct:     1,
				Explanation: "Rebase = новые хеши. Если кто-то уже скачал старые коммиты — push --force сломает их историю. Золотое правило: rebase только свои локальные ветки.",
			},
			{
				Question:    "Что означают маркеры <<<<<<< и >>>>>>> в файле при конфликте?",
				Options:     []string{"Комментарии", "<<<<<<< — начало твоих изменений, ======= — разделитель, >>>>>>> — изменения другой ветки", "Ошибки синтаксиса", "Метаданные Git"},
				Correct:     1,
				Explanation: "Git вставляет маркеры в файл: между <<<<<<< и ======= — твои изменения (HEAD), между ======= и >>>>>>> — изменения другой ветки. Нужно выбрать что оставить и удалить маркеры.",
			},
		},
		Tasks: []T{
			{
				Title:      "Парсер конфликтов",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "<<<<<<< HEAD", Definition: "Маркер начала конфликта. После него — версия текущей ветки (ours)."},
					{Term: "=======", Definition: "Разделитель между ours и theirs версиями в конфликте."},
					{Term: ">>>>>>> branch", Definition: "Маркер конца конфликта. После >>>>>>> — имя ветки с чужими изменениями."},
				},
				Description: `<p>На вход подаётся файл с Git-конфликтами. Извлеки обе версии.</p>
<p>Для каждого конфликта выведи:</p>
<pre><code>CONFLICT:
OURS: &lt;текст текущей ветки&gt;
THEIRS: &lt;текст другой ветки&gt;</code></pre>
<p>Строки вне конфликтов — пропускай.</p>`,
				Hints: `<p>Используй флаг состояния: 0 = обычный текст, 1 = читаем OURS (после &lt;&lt;&lt;&lt;&lt;&lt;&lt;), 2 = читаем THEIRS (после =======). При >>>>>>> — выводи результат.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	state := 0 // 0=normal, 1=ours, 2=theirs
	var ours, theirs []string

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "<<<<<<<"):
			state = 1
			ours = nil
			theirs = nil
		case line == "=======" && state == 1:
			state = 2
		case strings.HasPrefix(line, ">>>>>>>"):
			fmt.Println("CONFLICT:")
			fmt.Println("OURS: " + strings.Join(ours, "; "))
			fmt.Println("THEIRS: " + strings.Join(theirs, "; "))
			state = 0
		case state == 1:
			ours = append(ours, line)
		case state == 2:
			theirs = append(theirs, line)
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	state := 0 // 0=normal, 1=ours, 2=theirs
	var ours, theirs []string

	for scanner.Scan() {
		line := scanner.Text()

		// TODO: определи что за строка:
		// strings.HasPrefix(line, "<<<<<<<") → начало конфликта, state=1
		// line == "=======" && state==1  → переход к theirs, state=2
		// strings.HasPrefix(line, ">>>>>>>") → конец, выведи результат
		// state==1 → добавь в ours
		// state==2 → добавь в theirs
		//
		// Формат вывода:
		// CONFLICT:
		// OURS: строка1; строка2
		// THEIRS: строка1; строка2
		_ = line
		_ = ours
		_ = theirs
	}
}`,
				TestCases: []TestCase{
					{Input: "package main\n<<<<<<< HEAD\nfmt.Println(\"Hello from main\")\n=======\nfmt.Println(\"Hello from feature\")\n>>>>>>> feature/login", ExpectedOutput: "CONFLICT:\nOURS: fmt.Println(\"Hello from main\")\nTHEIRS: fmt.Println(\"Hello from feature\")"},
					{Input: "line1\n<<<<<<< HEAD\nA\nB\n=======\nC\n>>>>>>> dev\nline2\n<<<<<<< HEAD\nX\n=======\nY\nZ\n>>>>>>> dev", ExpectedOutput: "CONFLICT:\nOURS: A; B\nTHEIRS: C\nCONFLICT:\nOURS: X\nTHEIRS: Y; Z"},
				},
			},
			{
				Title:      "Граф коммитов: найди общего предка",
				Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "map[string]string", Definition: "Словарь parent: ключ — hash коммита, значение — hash родителя."},
					{Term: "Merge base", Definition: "Общий предок двух веток. Git использует его для 3-way merge: base vs ours vs theirs."},
				},
				Description: `<p>При merge Git ищет <strong>общего предка</strong> (merge base) двух веток.</p>
<p><strong>Вход:</strong></p>
<ul>
<li>Строки формата <code>hash parent</code> (коммит и его родитель). ROOT-коммит: <code>hash ROOT</code></li>
<li>Последняя строка: <code>? hash1 hash2</code> — найди общего предка</li>
</ul>
<p>Выведи hash общего предка.</p>
<p><em>Алгоритм:</em> собери всех предков hash1 в set. Потом иди от hash2 вверх пока не встретишь узел из set.</p>`,
				Hints: `<p>1) Построй map[string]string (child→parent). 2) От hash1 иди к ROOT, каждый hash добавляй в set. 3) От hash2 иди к ROOT, первый hash который уже в set — ответ.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	parent := map[string]string{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "? ") {
			parts := strings.Split(line[2:], " ")
			h1, h2 := parts[0], parts[1]

			// Collect ancestors of h1
			ancestors := map[string]bool{}
			for cur := h1; cur != "" && cur != "ROOT"; cur = parent[cur] {
				ancestors[cur] = true
			}

			// Walk from h2 up, find first common
			for cur := h2; cur != "" && cur != "ROOT"; cur = parent[cur] {
				if ancestors[cur] {
					fmt.Println(cur)
					return
				}
			}
		} else {
			parts := strings.Split(line, " ")
			parent[parts[0]] = parts[1]
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	parent := map[string]string{}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "? ") {
			// TODO: последняя строка — запрос
			// parts := strings.Split(line[2:], " ")
			// h1, h2 := parts[0], parts[1]
			//
			// 1. Собери всех предков h1 в set (map[string]bool)
			//    Иди от h1 вверх по parent пока != "ROOT"
			// 2. Иди от h2 вверх — первый попавший в set = ответ
			_ = parent
		} else {
			// Строка "hash parent" — запомни связь
			parts := strings.Split(line, " ")
			parent[parts[0]] = parts[1]
		}
	}
}`,
				TestCases: []TestCase{
					{Input: "aaa ROOT\nbbb aaa\nccc bbb\nddd bbb\neee ddd\n? ccc eee", ExpectedOutput: "bbb"},
					{Input: "a1 ROOT\na2 a1\na3 a2\nb1 a2\nb2 b1\n? a3 b2", ExpectedOutput: "a2"},
				},
			},
			{
				Title:      "Resolve конфликт",
				Difficulty: "easy",
				Description: `<p>Удали маркеры конфликта и оставь обе версии (ours + theirs). На вход текст с конфликтом:</p>
<p>Ввод:</p><pre><code>Hello
<<<<<<< HEAD
world from main
=======
world from feature
>>>>>>> feature
End</code></pre>
<p>Вывод:</p><pre><code>Hello
world from main
world from feature
End</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "conflict markers", Definition: "<<<<<<< HEAD, =======, >>>>>>> branch — Git вставляет при конфликте. Удали маркеры, оставь нужный код."},
				},
				TestCases: []TestCase{
					{Input: "Hello\n<<<<<<< HEAD\nworld from main\n=======\nworld from feature\n>>>>>>> feature\nEnd", ExpectedOutput: "Hello\nworld from main\nworld from feature\nEnd"},
				},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() {
    sc := bufio.NewScanner(os.Stdin)
    for sc.Scan() {
        line := sc.Text()
        if strings.HasPrefix(line, "<<<<<<<") || line == "=======" || strings.HasPrefix(line, ">>>>>>>") {
            continue
        }
        fmt.Println(line)
    }
}`,
				Hints: `<p>Пропусти строки начинающиеся с <code>&lt;&lt;&lt;&lt;&lt;&lt;&lt;</code>, <code>=======</code>, <code>&gt;&gt;&gt;&gt;&gt;&gt;&gt;</code>.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin)
    for sc.Scan(){l:=sc.Text();if strings.HasPrefix(l,"<<<<<<<")||l=="======="||strings.HasPrefix(l,">>>>>>>"){continue};fmt.Println(l)} }</code></pre>`,
			},
			{
				Title:      "Merge vs Rebase selector",
				Difficulty: "easy",
				Description: `<p>По ситуации посоветуй merge или rebase:</p>
<p>Ввод:</p><pre><code>3
public-branch
local-feature
shared-with-team</code></pre>
<p>Вывод:</p><pre><code>public-branch: MERGE (never rebase public branches)
local-feature: REBASE (clean linear history)
shared-with-team: MERGE (others may have pulled)</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Golden rule", Definition: "Никогда не rebase ветку которую кто-то ещё использует. Rebase только свои локальные ветки."},
				},
				TestCases: []TestCase{
					{Input: "3\npublic-branch\nlocal-feature\nshared-with-team", ExpectedOutput: "public-branch: MERGE (never rebase public branches)\nlocal-feature: REBASE (clean linear history)\nshared-with-team: MERGE (others may have pulled)"},
				},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        sc.Scan(); name := sc.Text()
        if strings.Contains(name, "local") { fmt.Printf("%s: REBASE (clean linear history)\n", name)
        } else { fmt.Printf("%s: MERGE (never rebase public branches)\n", name) }
    }
}`,
				Hints: `<p>"local" → REBASE, "shared" → MERGE (others may have pulled), остальное → MERGE.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() { var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();nm:=sc.Text()
        if strings.Contains(nm,"local"){fmt.Printf("%s: REBASE (clean linear history)\n",nm)
        }else if strings.Contains(nm,"shared"){fmt.Printf("%s: MERGE (others may have pulled)\n",nm)
        }else{fmt.Printf("%s: MERGE (never rebase public branches)\n",nm)}} }</code></pre>`,
			},
			{
				Title:      "Three-way merge симулятор",
				Difficulty: "hard",
				Description: `<p>Симулируй 3-way merge: base, ours, theirs. Если строка одинакова в обоих — берём. Если отличается в одном — берём изменённую. Если в обоих — CONFLICT:</p>
<p>Ввод:</p><pre><code>3
hello hello hello
world WORLD world
foo bar baz</code></pre>
<p>Вывод:</p><pre><code>hello
WORLD
CONFLICT: bar vs baz</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "3-way merge", Definition: "Сравниваем base с ours и theirs. Если одна сторона изменила — берём изменение. Обе — конфликт."},
				},
				TestCases: []TestCase{
					{Input: "3\nhello hello hello\nworld WORLD world\nfoo bar baz", ExpectedOutput: "hello\nWORLD\nCONFLICT: bar vs baz"},
				},
				StarterCode: `package main
import "fmt"
func main() {
    var n int; fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var base, ours, theirs string; fmt.Scan(&base, &ours, &theirs)
        if ours == theirs { fmt.Println(ours)
        } else if ours == base { fmt.Println(theirs)
        } else if theirs == base { fmt.Println(ours)
        } else { fmt.Printf("CONFLICT: %s vs %s\n", ours, theirs) }
    }
}`,
				Hints: `<p>4 случая: обе одинаковы, только ours изменился, только theirs, оба изменились → конфликт.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func main() { var n int;fmt.Scan(&n);for i:=0;i<n;i++{var b,o,t string;fmt.Scan(&b,&o,&t)
    if o==t{fmt.Println(o)}else if o==b{fmt.Println(t)}else if t==b{fmt.Println(o)}else{fmt.Printf("CONFLICT: %s vs %s\n",o,t)}} }</code></pre>`,
			},
		},
	}
}
