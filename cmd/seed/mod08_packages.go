package main

func mod08_packages() M {
	return M{
		Slug: "packages", Title: "Пакеты и модули", Order: 8,
		Description:   "go mod, internal/, видимость, циклические зависимости, vendor, replace.",
		Track:         "shared",
		Difficulty:    "intermediate",
		Prerequisites: []string{"errors"},
		Lessons: []L{
			{
				Slug: "go-modules", Title: "Модули, пакеты и видимость", Order: 1,
				Content: `<h1>Организация кода в Go</h1>

<h2>Под капотом: как Go находит пакеты</h2>
<p>Когда ты пишешь <code>import "github.com/go-chi/chi/v5"</code>, Go делает:</p>
<ol>
<li>Смотрит в <code>go.mod</code> — какая версия нужна</li>
<li>Ищет в локальном кеше <code>$GOPATH/pkg/mod/</code></li>
<li>Если нет — скачивает через прокси (<code>proxy.golang.org</code>)</li>
<li>Проверяет хеш в <code>go.sum</code> (целостность)</li>
</ol>

<h2>go.mod — сердце модуля</h2>
<pre><code>module github.com/backendraz/watchtogether  // путь модуля

go 1.22  // минимальная версия Go

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/jackc/pgx/v5 v5.5.5
)

// Для локальной разработки связанных проектов:
replace github.com/mylib => ../mylib</code></pre>

<h2>Видимость: заглавная = экспортировано</h2>
<pre><code>package video

func ProcessVideo() {} // ✅ видно из других пакетов
func processVideo() {} // ❌ только внутри пакета video

type Video struct {
    Title    string  // ✅ экспортировано
    filePath string  // ❌ приватное поле
}</code></pre>

<p><strong>Под капотом:</strong> компилятор буквально проверяет первую букву имени через <code>unicode.IsUpper()</code>. Это не конвенция — это правило языка.</p>

<h2>internal/ — жёсткая граница видимости</h2>
<pre><code>project/
├── cmd/server/main.go     // может импортировать internal/*
├── internal/
│   ├── handler/           // НЕ может быть импортирован извне модуля
│   ├── service/
│   └── model/
├── pkg/
│   └── videoscanner/      // МОЖЕТ быть импортирован кем угодно</code></pre>

<p><strong>Компилятор запрещает</strong> импорт <code>internal/</code> из-за пределов родительского каталога. Это не конвенция — это enforcement на уровне go build.</p>

<h2>Циклические зависимости</h2>
<pre><code>// ОШИБКА КОМПИЛЯЦИИ: import cycle
// package handler imports package service
// package service imports package handler

// Решение: интерфейсы
// handler определяет интерфейс
// service реализует его
// зависимость идёт в одну сторону</code></pre>

<h2>Стандартная структура проекта</h2>
<pre><code>project/
├── cmd/                   // точки входа (main пакеты)
│   └── server/main.go
├── internal/              // приватный код проекта
│   ├── handler/           // HTTP обработчики
│   ├── service/           // бизнес-логика
│   ├── repository/        // работа с БД
│   ├── model/             // модели данных
│   ├── config/            // конфигурация
│   └── middleware/        // HTTP middleware
├── pkg/                   // публичные библиотеки
├── migrations/            // SQL миграции
├── Dockerfile
├── docker-compose.yml
├── Makefile               // команды сборки
├── go.mod
└── go.sum</code></pre>

<h2>Полезные команды</h2>
<pre><code>go mod tidy       # убрать лишнее, добавить нужное
go mod download   # скачать все зависимости
go mod vendor     # скопировать зависимости в vendor/
go mod why github.com/lib  # почему нужна эта зависимость
go list -m all    # все зависимости</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: один каталог = один пакет
// Нельзя иметь package handler и package service в одной папке

// ОШИБКА: import _ "пакет" без причины
// blank import нужен ТОЛЬКО для side-effects (драйверы БД, init())

// ОШИБКА: пакеты-утилиты (utils/, helpers/, common/)
// Каждый пакет должен иметь ОДНУ ответственность с осмысленным именем</code></pre>`,

				Quiz: []Q{
					{Question: "Что делает go.sum?", Options: []string{"Список зависимостей", "Хеши зависимостей для проверки целостности — защита от подмены", "Суммарный размер", "Документация"}, Correct: 1, Explanation: "go.sum содержит криптографические хеши каждой зависимости. Go проверяет хеш при сборке — если кто-то подменил пакет, сборка упадёт."},
					{Question: "Можно ли два пакета в одной директории?", Options: []string{"Да", "Нет — один каталог = один пакет (кроме _test)", "Зависит от настроек", "Только с go:generate"}, Correct: 1, Explanation: "Go требует один пакет на директорию. Исключение: файлы _test.go могут использовать package xxx_test (black-box тестирование)."},
					{Question: "Почему utils/ — плохое имя для пакета?", Options: []string{"Слишком короткое", "Не описывает что пакет делает — нарушает принцип единой ответственности", "Зарезервировано Go", "Слишком длинное"}, Correct: 1, Explanation: "Хороший пакет имеет ясное назначение: auth, cache, validator. utils — свалка без границ. Каждая функция из utils обычно принадлежит конкретному пакету."},
					{Question: "Что произойдёт при циклическом импорте?", Options: []string{"Предупреждение", "Ошибка компиляции — Go запрещает циклические зависимости", "Бесконечный цикл", "Go разрешит автоматически"}, Correct: 1, Explanation: "Go не компилирует код с циклическими зависимостями. Решение: вынести общие типы в отдельный пакет или использовать интерфейсы."},
					{Question: "Зачем нужен replace в go.mod?", Options: []string{"Заменить Go версию", "Указать локальный путь к зависимости для разработки связанных модулей", "Удалить пакет", "Переименовать модуль"}, Correct: 1, Explanation: "replace позволяет подменить зависимость на локальную копию. Полезно при разработке двух связанных модулей одновременно."},
				},
				Tasks: []T{
					{
						Title:      "Видимость: экспорт или нет?",
						Difficulty: "easy",
						Glossary: []GlossaryItem{
							{Term: "unicode.IsUpper(r)", Definition: "Проверяет, является ли руна заглавной буквой. Именно так Go определяет экспортируемость."},
							{Term: "[]rune(s)[0]", Definition: "Преобразует строку в слайс рун и берёт первую. Нужно для юникода (кириллица, эмодзи)."},
							{Term: "bufio.Scanner", Definition: "Читает ввод построчно. scanner.Scan() → true пока есть строки, scanner.Text() → текст строки."},
						},
						Description: `<p>На вход подаются имена Go-идентификаторов (по одному на строку). Для каждого выведи:</p>
<ul>
<li><code>ЭКСПОРТ</code> — если начинается с заглавной буквы (экспортировано за пределы пакета)</li>
<li><code>ПРИВАТНОЕ</code> — если начинается со строчной</li>
</ul>
<p><strong>Правило Go:</strong> компилятор проверяет первую букву через <code>unicode.IsUpper()</code>. Это не конвенция — это правило языка.</p>`,
						Hints: `<p>Используй <code>unicode.IsUpper([]rune(name)[0])</code> — это именно то, что делает компилятор Go.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()
		if unicode.IsUpper([]rune(name)[0]) {
			fmt.Println("ЭКСПОРТ")
		} else {
			fmt.Println("ПРИВАТНОЕ")
		}
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		name := scanner.Text()
		// TODO: проверь первую руну name через unicode.IsUpper
		// Выведи "ЭКСПОРТ" или "ПРИВАТНОЕ"
		_ = name
	}
}`,
						TestCases: []TestCase{
							{Input: "ProcessVideo\nfilePath\nTitle\nhandleRequest\nNewServer\ninit", ExpectedOutput: "ЭКСПОРТ\nПРИВАТНОЕ\nЭКСПОРТ\nПРИВАТНОЕ\nЭКСПОРТ\nПРИВАТНОЕ"},
							{Input: "ID\nctx\nErrNotFound\nmutex\nHTTPClient\nclose", ExpectedOutput: "ЭКСПОРТ\nПРИВАТНОЕ\nЭКСПОРТ\nПРИВАТНОЕ\nЭКСПОРТ\nПРИВАТНОЕ"},
						},
					},
					{
						Title:      "Разбор import-пути",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "strings.Split(s, sep)", Definition: "Разбивает строку на слайс по разделителю. \"a/b/c\" → [\"a\",\"b\",\"c\"]"},
							{Term: "strings.HasPrefix(s, prefix)", Definition: "Проверяет начинается ли строка с prefix. Для определения стандартной vs внешней библиотеки."},
							{Term: "strings.Contains(s, substr)", Definition: "Проверяет содержит ли строка подстроку. Полезно для поиска версии (v2, v3)."},
						},
						Description: `<p>На вход подаются Go import-пути (по одному на строку). Для каждого выведи три строки:</p>
<ol>
<li><strong>Тип:</strong> <code>STD</code> (стандартная библиотека — нет точки в первом сегменте) или <code>EXT</code> (внешняя)</li>
<li><strong>Имя пакета</strong> (последний сегмент пути, без версии): <code>chi</code>, <code>fmt</code>, <code>pgx</code></li>
<li><strong>Версия:</strong> если есть сегмент вида <code>v2</code>, <code>v5</code> — вывести его, иначе <code>v1</code></li>
</ol>
<p><em>Пример:</em> <code>github.com/go-chi/chi/v5</code> → <code>EXT chi v5</code></p>
<p><em>Пример:</em> <code>fmt</code> → <code>STD fmt v1</code></p>`,
						Hints: `<p>Разбей путь через <code>strings.Split(path, "/")</code>. Стандартный пакет — первый сегмент без точки. Версия — последний сегмент если начинается с "v" и дальше цифра.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := scanner.Text()
		parts := strings.Split(path, "/")

		typ := "STD"
		if strings.Contains(parts[0], ".") {
			typ = "EXT"
		}

		version := "v1"
		name := parts[len(parts)-1]
		if len(name) >= 2 && name[0] == 'v' && unicode.IsDigit(rune(name[1])) {
			version = name
			name = parts[len(parts)-2]
		}

		fmt.Printf("%s %s %s\n", typ, name, version)
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
		path := scanner.Text()
		parts := strings.Split(path, "/")

		// TODO: определи тип (STD или EXT)
		// Подсказка: стандартный пакет не содержит точку в первом сегменте
		typ := "?"

		// TODO: найди имя пакета (последний сегмент, но не версия)
		// и версию (vN если есть, иначе "v1")
		name := "?"
		version := "v1"

		_ = parts

		fmt.Printf("%s %s %s\n", typ, name, version)
	}
}`,
						TestCases: []TestCase{
							{Input: "fmt\ngithub.com/go-chi/chi/v5\nstrings\ngithub.com/jackc/pgx/v5\nio", ExpectedOutput: "STD fmt v1\nEXT chi v5\nSTD strings v1\nEXT pgx v5\nSTD io v1"},
							{Input: "encoding/json\ngithub.com/gorilla/mux\nnet/http\ngithub.com/redis/go-redis/v9", ExpectedOutput: "STD json v1\nEXT mux v1\nSTD http v1\nEXT go-redis v9"},
						},
					},
					{
						Title:      "Детектор циклических зависимостей",
						Difficulty: "hard",
						Glossary: []GlossaryItem{
							{Term: "map[string][]string", Definition: "Граф смежности: ключ — пакет, значение — слайс пакетов от которых он зависит."},
							{Term: "DFS (поиск в глубину)", Definition: "Обход графа. Для поиска циклов нужно отслеживать три состояния: не посещён, в процессе, завершён."},
							{Term: "strings.Split(s, \" -> \")", Definition: "Разбить строку \"handler -> service\" на два пакета."},
						},
						Description: `<p>Go запрещает циклические зависимости между пакетами — код не скомпилируется.</p>
<p>На вход подаются зависимости (по одной на строку) в формате <code>пакет -> зависимость</code>. Определи, есть ли цикл.</p>
<p>Выведи <code>CYCLE: пакет1 -> пакет2 -> ... -> пакет1</code> если цикл найден, или <code>OK</code> если циклов нет.</p>
<p><em>Пример входа:</em></p>
<pre><code>handler -> service
service -> repository
repository -> model</code></pre>
<p><em>Выход:</em> <code>OK</code></p>`,
						Hints: `<p>Используй DFS с тремя состояниями: 0 — не посещён, 1 — в процессе обхода (на стеке), 2 — завершён. Если при обходе встречаешь узел в состоянии 1 — нашёл цикл. Восстанови путь через отдельный слайс.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	graph := map[string][]string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " -> ")
		from, to := parts[0], parts[1]
		graph[from] = append(graph[from], to)
	}

	state := map[string]int{} // 0=white, 1=gray, 2=black
	var path []string

	var dfs func(node string) bool
	dfs = func(node string) bool {
		state[node] = 1
		path = append(path, node)
		for _, dep := range graph[node] {
			if state[dep] == 1 {
				// Found cycle — extract it
				start := 0
				for i, p := range path {
					if p == dep {
						start = i
						break
					}
				}
				cycle := append(path[start:], dep)
				fmt.Println("CYCLE: " + strings.Join(cycle, " -> "))
				return true
			}
			if state[dep] == 0 {
				if dfs(dep) {
					return true
				}
			}
		}
		path = path[:len(path)-1]
		state[node] = 2
		return false
	}

	for node := range graph {
		if state[node] == 0 {
			if dfs(node) {
				return
			}
		}
	}
	fmt.Println("OK")
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Считай зависимости в граф
	graph := map[string][]string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " -> ")
		from, to := parts[0], parts[1]
		graph[from] = append(graph[from], to)
	}

	// TODO: реализуй DFS для поиска цикла
	// state: 0 = не посещён, 1 = в процессе (серый), 2 = завершён (чёрный)
	// Если встречаешь серый узел — цикл!
	// Выведи "CYCLE: a -> b -> ... -> a" или "OK"

	_ = graph
	fmt.Println("OK")
}`,
						TestCases: []TestCase{
							{Input: "handler -> service\nservice -> repository\nrepository -> model", ExpectedOutput: "OK"},
							{Input: "handler -> service\nservice -> handler", ExpectedOutput: "CYCLE: handler -> service -> handler"},
							{Input: "handler -> service\nservice -> repository\nrepository -> handler", ExpectedOutput: "CYCLE: handler -> service -> repository -> handler"},
						},
					},
					{
						Title:      "go.mod парсер",
						Difficulty: "easy",
						Description: `<p>Прочитай упрощённый go.mod и выведи: имя модуля, версию Go и количество зависимостей:</p>
<p>Ввод:</p><pre><code>module github.com/user/myapp
go 1.22
require github.com/go-chi/chi v5.0.12
require github.com/jackc/pgx/v5 v5.5.0
require github.com/redis/go-redis/v9 v9.4.0</code></pre>
<p>Вывод:</p><pre><code>Module: github.com/user/myapp
Go: 1.22
Dependencies: 3</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "go.mod", Definition: "Файл описания модуля Go: имя, версия Go, зависимости. Аналог package.json (Node) или requirements.txt (Python)."},
						},
						TestCases: []TestCase{
							{Input: "module github.com/user/myapp\ngo 1.22\nrequire github.com/go-chi/chi v5.0.12\nrequire github.com/jackc/pgx/v5 v5.5.0\nrequire github.com/redis/go-redis/v9 v9.4.0", ExpectedOutput: "Module: github.com/user/myapp\nGo: 1.22\nDependencies: 3"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    var modName, goVer string
    deps := 0
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "module ") {
            modName = strings.TrimPrefix(line, "module ")
        } else if strings.HasPrefix(line, "go ") {
            goVer = strings.TrimPrefix(line, "go ")
        } else if strings.HasPrefix(line, "require ") {
            deps++
        }
    }
    fmt.Printf("Module: %s\nGo: %s\nDependencies: %d\n", modName, goVer, deps)
}`,
						Hints: `<p><code>strings.HasPrefix(line, "require ")</code> — считай все строки с require.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    var mod, ver string; deps := 0
    for scanner.Scan() {
        l := scanner.Text()
        if strings.HasPrefix(l, "module ") { mod = l[7:] } else if strings.HasPrefix(l, "go ") { ver = l[3:] } else if strings.HasPrefix(l, "require ") { deps++ }
    }
    fmt.Printf("Module: %s\nGo: %s\nDependencies: %d\n", mod, ver, deps)
}</code></pre>`,
					},
					{
						Title:      "Сортировка импортов",
						Difficulty: "medium",
						Description: `<p>Go-конвенция: импорты в 3 группах: 1) стандартные, 2) внешние, 3) внутренние (свой модуль). Рассортируй список импортов:</p>
<p>Ввод:</p><pre><code>github.com/user/myapp
fmt
github.com/go-chi/chi
net/http
github.com/user/myapp/internal/handler
strings
github.com/jackc/pgx</code></pre>
<p>Вывод:</p><pre><code>STD:
  fmt
  net/http
  strings
EXT:
  github.com/go-chi/chi
  github.com/jackc/pgx
INT:
  github.com/user/myapp
  github.com/user/myapp/internal/handler</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "goimports", Definition: "Утилита Go — автоматически сортирует импорты по группам. Встроена в IDE."},
						},
						TestCases: []TestCase{
							{Input: "github.com/user/myapp\nfmt\ngithub.com/go-chi/chi\nnet/http\ngithub.com/user/myapp/internal/handler\nstrings\ngithub.com/jackc/pgx", ExpectedOutput: "STD:\n  fmt\n  net/http\n  strings\nEXT:\n  github.com/go-chi/chi\n  github.com/jackc/pgx\nINT:\n  github.com/user/myapp\n  github.com/user/myapp/internal/handler"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    // Первая строка с точкой — наш модуль (самый короткий внешний путь, являющийся префиксом для INT)
    var std, ext, internal []string
    myModule := ""
    var all []string
    for scanner.Scan() { all = append(all, scanner.Text()) }
    // Найди самый короткий путь с точкой — это наш модуль
    for _, p := range all {
        if strings.Contains(p, ".") {
            if myModule == "" || len(p) < len(myModule) { myModule = p }
        }
    }
    for _, p := range all {
        if !strings.Contains(p, ".") { std = append(std, p) } else if strings.HasPrefix(p, myModule) { internal = append(internal, p) } else { ext = append(ext, p) }
    }
    sort.Strings(std); sort.Strings(ext); sort.Strings(internal)
    fmt.Println("STD:"); for _, s := range std { fmt.Println("  " + s) }
    fmt.Println("EXT:"); for _, s := range ext { fmt.Println("  " + s) }
    fmt.Println("INT:"); for _, s := range internal { fmt.Println("  " + s) }
}`,
						Hints: `<p>STD: нет точки в первом сегменте. INT: начинается с пути модуля. EXT: остальное.</p>`,
						Solution: `<pre><code>package main

import ("bufio";"fmt";"os";"sort";"strings")

func main() {
    sc := bufio.NewScanner(os.Stdin)
    var std, ext, intl, all []string
    for sc.Scan() { all = append(all, sc.Text()) }
    myMod := ""
    for _, p := range all {
        if strings.Contains(p, ".") && (myMod == "" || len(p) < len(myMod)) { myMod = p }
    }
    for _, p := range all {
        if !strings.Contains(p, ".") { std = append(std, p) } else if strings.HasPrefix(p, myMod) { intl = append(intl, p) } else { ext = append(ext, p) }
    }
    sort.Strings(std); sort.Strings(ext); sort.Strings(intl)
    fmt.Println("STD:"); for _, s := range std { fmt.Println("  "+s) }
    fmt.Println("EXT:"); for _, s := range ext { fmt.Println("  "+s) }
    fmt.Println("INT:"); for _, s := range intl { fmt.Println("  "+s) }
}</code></pre>`,
					},
				},
			},
		},
	}
}
