package main

// ════════════════════════════════════════════════════════════════
// MILESTONE PROJECTS — итоговые проекты между уровнями
// ════════════════════════════════════════════════════════════════

func mod_milestone_beginner() M {
	return M{
		Slug:          "project-cli",
		Title:         "Проект: CLI-утилита WatchTogether",
		Description:   "Итоговый проект уровня Beginner. Объедини всё что знаешь: переменные, функции, структуры, ввод/вывод, ошибки.",
		Order:         100, // will be overridden
		Track:         "shared",
		Difficulty:    "beginner",
		Prerequisites: []string{"errors"},
		Lessons: []L{
			{
				Slug: "project-cli-tool", Title: "Проект: CLI менеджер видео", Order: 1,
				Difficulty: "beginner", Track: "shared",
				Content: `<h1>Итоговый проект: CLI менеджер видео</h1>

<h2>Что строим</h2>
<p>Консольное приложение для управления каталогом видео WatchTogether. Объединяет всё что ты изучил:</p>
<ul>
<li><strong>Переменные и типы</strong> — структура Video</li>
<li><strong>Функции</strong> — add, list, search, delete</li>
<li><strong>Слайсы и map</strong> — хранение каталога</li>
<li><strong>Ошибки</strong> — валидация ввода</li>
<li><strong>Указатели</strong> — изменение данных</li>
<li><strong>Циклы и условия</strong> — REPL-интерфейс</li>
</ul>

<h2>Требования</h2>
<pre><code>WatchTogether Video Manager
Commands: add, list, search, delete, stats, quit

> add
Title: Matrix
Year: 1999
Genre: sci-fi
Duration (min): 136
Added: Matrix (1999) [sci-fi] 2h16m

> list
1. Matrix (1999) [sci-fi] 2h16m
2. Inception (2010) [sci-fi] 2h28m

> search sci-fi
Found 2 videos:
1. Matrix (1999) [sci-fi]
2. Inception (2010) [sci-fi]

> stats
Total: 2 videos
Genres: sci-fi (2)
Total duration: 4h44m

> delete 1
Deleted: Matrix

> quit
Bye!</code></pre>

<h2>Подсказки по архитектуре</h2>
<pre><code>type Video struct {
    Title    string
    Year     int
    Genre    string
    Duration int // minutes
}

type Catalog struct {
    videos []Video
}

func (c *Catalog) Add(v Video) { ... }
func (c *Catalog) Delete(index int) error { ... }
func (c *Catalog) Search(query string) []Video { ... }
func (c Catalog) Stats() string { ... }
func (c Catalog) List() string { ... }

func main() {
    catalog := &Catalog{}
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("> ")
        scanner.Scan()
        cmd := scanner.Text()
        switch cmd {
        case "add": ...
        case "list": ...
        case "quit": return
        }
    }
}</code></pre>

<h2>Критерии оценки</h2>
<ol>
<li>Программа компилируется и запускается</li>
<li>Все команды работают корректно</li>
<li>Ошибки обрабатываются (некорректный год, пустое название)</li>
<li>Код организован: структуры, методы, отдельные функции</li>
<li>Бонус: сохранение каталога в JSON-файл</li>
</ol>`,

				Quiz: []Q{},
				Tasks: []T{
					{
						Title:      "CLI менеджер видео — базовый",
						Difficulty: "medium",
						Description: `<p>Реализуй минимальную версию: add, list, quit. Читай ввод через <code>bufio.Scanner</code>.</p>
<p>Ввод:</p>
<pre><code>add
Matrix 1999 136
list
quit</code></pre>
<p>Вывод:</p>
<pre><code>Added: Matrix (1999) 2h16m
1. Matrix (1999) 2h16m
Bye!</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "bufio.NewScanner(os.Stdin)", Definition: "Создаёт сканер для чтения из stdin построчно. scanner.Scan() читает строку, scanner.Text() возвращает её."},
							{Term: "REPL", Definition: "Read-Eval-Print Loop — бесконечный цикл: читай команду → выполни → напечатай результат → повтори."},
						},
						TestCases: []TestCase{
							{Input: "add\nMatrix 1999 136\nlist\nquit", ExpectedOutput: "Added: Matrix (1999) 2h16m\n1. Matrix (1999) 2h16m\nBye!"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
)

type Video struct {
    Title    string
    Year     int
    Duration int
}

func (v Video) String() string {
    return fmt.Sprintf("%s (%d) %dh%02dm", v.Title, v.Year, v.Duration/60, v.Duration%60)
}

func main() {
    var videos []Video
    scanner := bufio.NewScanner(os.Stdin)

    for {
        scanner.Scan()
        cmd := scanner.Text()

        switch cmd {
        case "add":
            // Читай Title Year Duration из следующей строки
        case "list":
            // Выведи все видео
        case "quit":
            fmt.Println("Bye!")
            return
        }
    }
}`,
						Hints:    `<p>Для add: <code>scanner.Scan(); fmt.Sscanf(scanner.Text(), "%s %d %d", &v.Title, &v.Year, &v.Duration)</code>. Для list: <code>for i, v := range videos</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
)

type Video struct {
    Title    string
    Year     int
    Duration int
}

func (v Video) String() string {
    return fmt.Sprintf("%s (%d) %dh%02dm", v.Title, v.Year, v.Duration/60, v.Duration%60)
}

func main() {
    var videos []Video
    scanner := bufio.NewScanner(os.Stdin)

    for {
        scanner.Scan()
        cmd := scanner.Text()

        switch cmd {
        case "add":
            scanner.Scan()
            var v Video
            fmt.Sscanf(scanner.Text(), "%s %d %d", &v.Title, &v.Year, &v.Duration)
            videos = append(videos, v)
            fmt.Printf("Added: %s\n", v)
        case "list":
            for i, v := range videos {
                fmt.Printf("%d. %s\n", i+1, v)
            }
        case "quit":
            fmt.Println("Bye!")
            return
        }
    }
}</code></pre>`,
					},
					{
						Title:      "CLI менеджер — полная версия",
						Difficulty: "hard",
						Description: `<p>Добавь к базовой версии:</p>
<ul>
<li><code>search &lt;query&gt;</code> — поиск по названию (регистронезависимый)</li>
<li><code>delete &lt;номер&gt;</code> — удаление по номеру из списка</li>
<li><code>stats</code> — общее количество видео и суммарная длительность</li>
<li>Обработку ошибок (некорректный номер, пустой список)</li>
</ul>`,
						Glossary: []GlossaryItem{
							{Term: "strings.Contains + strings.ToLower", Definition: "Регистронезависимый поиск: strings.Contains(strings.ToLower(s), strings.ToLower(query))."},
							{Term: "strconv.Atoi(s)", Definition: "Конвертирует строку в число. Для парсинга номера из команды delete."},
						},
						TestCases: []TestCase{},
						StarterCode: "",
						Hints:       `<p>search: цикл по videos, фильтруй по strings.Contains. delete: парси номер через strconv.Atoi, проверь диапазон, удали через append(s[:i], s[i+1:]...).</p>`,
						Solution:    `<p>Расширь базовую версию. Ключевое: проверка индексов, strings.ToLower для поиска, fmt.Sprintf для stats.</p>`,
					},
				},
			},
		},
	}
}

func mod_milestone_intermediate() M {
	return M{
		Slug:          "project-api",
		Title:         "Проект: REST API WatchTogether",
		Description:   "Итоговый проект уровня Intermediate. HTTP сервер + PostgreSQL + JSON API.",
		Order:         100,
		Track:         "backend",
		Difficulty:    "intermediate",
		Prerequisites: []string{"testing"},
		Lessons: []L{
			{
				Slug: "project-rest-api", Title: "Проект: REST API с нуля", Order: 1,
				Difficulty: "intermediate", Track: "backend",
				Content: `<h1>Итоговый проект: REST API WatchTogether</h1>

<h2>Что строим</h2>
<p>Полноценный REST API для WatchTogether. Объединяет модули HTTP, PostgreSQL, пакеты:</p>

<h2>Endpoints</h2>
<pre><code>GET    /api/videos          — список видео (с пагинацией)
GET    /api/videos/:id      — одно видео
POST   /api/videos          — создать видео
PUT    /api/videos/:id      — обновить видео
DELETE /api/videos/:id      — удалить видео
GET    /api/videos/search?q= — поиск по названию

GET    /api/rooms            — список комнат
POST   /api/rooms            — создать комнату (привязать видео)

GET    /api/health           — healthcheck</code></pre>

<h2>Структура проекта</h2>
<pre><code>watchtogether-api/
├── cmd/server/main.go       — точка входа, сборка зависимостей
├── internal/
│   ├── handler/             — HTTP handlers (chi router)
│   ├── model/               — структуры данных
│   ├── repository/          — SQL запросы (pgx)
│   └── config/              — конфигурация из ENV
├── migrations/              — SQL миграции
├── docker-compose.yml       — PostgreSQL для разработки
├── Makefile                 — команды: run, test, migrate
└── go.mod</code></pre>

<h2>Пошаговый план</h2>
<ol>
<li>Создай go.mod, docker-compose.yml с PostgreSQL</li>
<li>Напиши миграции: таблицы videos и rooms</li>
<li>Реализуй model/ — структуры Video, Room</li>
<li>Реализуй repository/ — VideoRepo с методами CRUD</li>
<li>Реализуй handler/ — HTTP handlers с JSON I/O</li>
<li>Собери всё в main.go</li>
<li>Протестируй через curl или httpie</li>
</ol>

<h2>Критерии</h2>
<ul>
<li>API работает, все endpoints возвращают JSON</li>
<li>Параметризованные SQL-запросы (без SQL-инъекций)</li>
<li>Правильные HTTP-коды: 200, 201, 400, 404, 500</li>
<li>Graceful shutdown</li>
<li>Пагинация через ?page=&limit=</li>
</ul>

<h2>Тестирование</h2>
<pre><code># Создать видео
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{"title":"Matrix","year":1999,"duration":136}'

# Список
curl http://localhost:8080/api/videos

# Поиск
curl "http://localhost:8080/api/videos/search?q=matrix"

# Health check
curl http://localhost:8080/api/health</code></pre>`,

				Quiz: []Q{},
				Tasks: []T{
					{
						Title:      "REST API — полная реализация",
						Difficulty: "hard",
						Description: `<p>Построй API по спецификации выше. Это объёмная задача — разбей на шаги:</p>
<ol>
<li>docker-compose.yml + миграции</li>
<li>model + repository (pgx)</li>
<li>handler (chi) + main.go</li>
<li>Тестирование через curl</li>
</ol>`,
						Glossary: []GlossaryItem{
							{Term: "chi.URLParam(r, \"id\")", Definition: "Извлекает параметр из URL. /videos/{id} → chi.URLParam(r, \"id\") вернёт значение."},
							{Term: "json.NewEncoder(w).Encode(v)", Definition: "Пишет JSON прямо в HTTP response. Эффективнее чем Marshal + Write."},
							{Term: "w.WriteHeader(http.StatusCreated)", Definition: "Устанавливает HTTP-код ответа. 201 Created для POST, 204 No Content для DELETE."},
						},
						TestCases:   []TestCase{},
						StarterCode: "",
						Hints:       `<p>Начни с health endpoint. Потом GET /videos (select all). Потом POST. Тестируй каждый шаг через curl.</p>`,
						Solution:    `<p>Паттерн из модулей HTTP + PostgreSQL. handler → repository → pool. main.go собирает зависимости.</p>`,
					},
				},
			},
		},
	}
}
