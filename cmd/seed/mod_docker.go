package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Docker — расширенный (5 уроков)
// Заменяет Docker-часть mod15_devops()
// ════════════════════════════════════════════════════════════════

func mod_docker_full() M {
	return M{
		Slug: "docker", Title: "Docker — контейнеризация", Order: 15,
		Description: "Контейнеры под капотом, Dockerfile, multi-stage, compose, networking, volumes, отладка.",
		Track:       "devops", Difficulty: "intermediate", Prerequisites: []string{"packages"},
		Lessons: []L{
			{
				Slug: "docker-basics", Title: "Что такое Docker и зачем он нужен", Order: 1,
				Difficulty: "beginner", Track: "devops",
				Content: `<h1>Docker — основы контейнеризации</h1>

<h2>Проблема: "у меня работает!"</h2>
<p>Разработчик: "На моей машине всё запускается". Сервер: ошибка. Причина: другая версия Go, нет PostgreSQL, не тот Linux. <strong>Docker решает это</strong> — приложение упаковано вместе с окружением.</p>

<h2>Контейнер vs Виртуальная машина</h2>
<pre><code>VM:        [App] [OS целиком] [Hypervisor]  → тяжёлая (GB)
Container: [App] [Libs только нужные]        → лёгкий (MB)

// Контейнер — это НЕ VM! Это процесс с изоляцией.
// Ядро Linux общее с хостом.
// Изоляция через: namespaces (PID, сеть, FS) + cgroups (лимиты CPU/RAM)</code></pre>

<h2>Основные команды</h2>
<pre><code># Скачать и запустить контейнер
docker run -d --name myapp -p 8080:8080 golang:1.22-alpine

# Посмотреть запущенные
docker ps

# Логи
docker logs myapp
docker logs myapp --tail 50 -f  # последние 50 + follow

# Войти внутрь
docker exec -it myapp sh

# Остановить и удалить
docker stop myapp
docker rm myapp

# Образы
docker images          # список
docker pull nginx      # скачать
docker rmi nginx       # удалить</code></pre>

<h2>Жизненный цикл</h2>
<pre><code># Image → Container → Running → Stopped → Removed
#
# Image = "рецепт" (read-only шаблон, слои FS)
# Container = "блюдо" (запущенный экземпляр образа)
#
# Можно создать 10 контейнеров из одного образа</code></pre>

<h2>Порты: -p host:container</h2>
<pre><code># -p 8080:3000 = порт 8080 на хосте → порт 3000 в контейнере
docker run -p 8080:3000 myapp
# Теперь curl http://localhost:8080 попадёт в контейнер на порт 3000

# -p 5433:5432 = PostgreSQL на нестандартном порту хоста
docker run -p 5433:5432 postgres:16</code></pre>

<h2>Читать глубже</h2>
<ul>
<li><a href="https://habr.com/ru/articles/310460/" target="_blank">Хабр: Docker для начинающих</a></li>
<li><a href="https://metanit.com/go/docker/" target="_blank">Metanit: Docker с Go</a></li>
</ul>`,

				Quiz: []Q{
					{Question: "Контейнер — это виртуальная машина?", Options: []string{"Да", "Нет — изолированный процесс с общим ядром ОС", "Зависит от ОС", "Микросервис"}, Correct: 1, Explanation: "Контейнер использует ядро хоста + namespaces. VM имеет своё ядро. Контейнер легче и быстрее."},
					{Question: "Чем Image отличается от Container?", Options: []string{"Ничем", "Image — read-only шаблон (рецепт). Container — запущенный экземпляр (блюдо)", "Image больше", "Container — на диске"}, Correct: 1, Explanation: "Image = слои FS, не запущен. Container = Image + writable layer + процесс. Из одного Image можно создать много Containers."},
					{Question: "docker run -p 8080:3000 — что значит?", Options: []string{"Открыть порт 8080", "Порт 8080 хоста маппится на порт 3000 контейнера", "Запустить на порту 80803000", "Открыть 2 порта"}, Correct: 1, Explanation: "-p host:container. Запрос на localhost:8080 → контейнер:3000. Нужно когда в контейнере app слушает 3000, а ты хочешь обращаться на 8080."},
					{Question: "Как посмотреть логи контейнера?", Options: []string{"cat /var/log/docker", "docker logs <name> — основная команда для отладки", "docker status", "docker inspect"}, Correct: 1, Explanation: "docker logs = stdout/stderr контейнера. --tail 50 — последние 50 строк. -f — follow (как tail -f). Первое что делаешь при проблемах."},
					{Question: "Зачем Docker если можно go build и скопировать бинарник?", Options: []string{"Мода", "Воспроизводимость: одинаковое окружение everywhere + зависимости (postgres, redis) + orchestration", "Быстрее", "Безопаснее"}, Correct: 1, Explanation: "Бинарник — ок для Go. Но: БД, Redis, миграции, env переменные, сети, TLS, healthchecks — всё это Docker решает декларативно. Одна команда — вся инфраструктура."},
				},
				Tasks: []T{
					{Title: "Docker команды", Difficulty: "easy", Description: `<p>Напиши какой командой:</p><ol><li>Запустить nginx на порту 9090</li><li>Посмотреть логи</li><li>Войти внутрь</li><li>Остановить</li></ol><p>Ввод: <code>nginx 9090</code></p><p>Вывод:</p><pre><code>docker run -d --name nginx -p 9090:80 nginx
docker logs nginx
docker exec -it nginx sh
docker stop nginx</code></pre>`, Glossary: []GlossaryItem{{Term: "-d", Definition: "Detached mode — в фоне. Без -d логи идут в терминал."}}, TestCases: []TestCase{{Input: "nginx 9090", ExpectedOutput: "docker run -d --name nginx -p 9090:80 nginx\ndocker logs nginx\ndocker exec -it nginx sh\ndocker stop nginx"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name string; var port int
    fmt.Scan(&name, &port)
    fmt.Printf("docker run -d --name %s -p %d:80 %s\n", name, port, name)
    fmt.Printf("docker logs %s\n", name)
    fmt.Printf("docker exec -it %s sh\n", name)
    fmt.Printf("docker stop %s\n", name)
}`, Hints: `<p>fmt.Printf с подстановкой имени и порта.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var name string; var port int; fmt.Scan(&name, &port); fmt.Printf("docker run -d --name %s -p %d:80 %s\ndocker logs %s\ndocker exec -it %s sh\ndocker stop %s\n", name, port, name, name, name, name) }</code></pre>`},
					{Title: "Порт маппинг", Difficulty: "easy", Description: `<p>По входным данным сгенерируй docker run с правильным маппингом:</p><p>Ввод: <code>postgres 5433 5432</code></p><p>Вывод: <code>docker run -d -p 5433:5432 postgres:16-alpine</code></p>`, Glossary: []GlossaryItem{{Term: "-p host:container", Definition: "Маппинг порта хоста на порт контейнера."}}, TestCases: []TestCase{{Input: "postgres 5433 5432", ExpectedOutput: "docker run -d -p 5433:5432 postgres:16-alpine"}, {Input: "redis 6380 6379", ExpectedOutput: "docker run -d -p 6380:6379 redis:7-alpine"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name string; var hostPort, containerPort int
    fmt.Scan(&name, &hostPort, &containerPort)
    var image string
    switch name {
    case "postgres": image = "postgres:16-alpine"
    case "redis": image = "redis:7-alpine"
    default: image = name + ":latest"
    }
    fmt.Printf("docker run -d -p %d:%d %s\n", hostPort, containerPort, image)
}`, Hints: `<p>switch для image name, fmt.Printf для команды.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; var hp, cp int; fmt.Scan(&n, &hp, &cp); img := n+":latest"; switch n { case "postgres": img = "postgres:16-alpine"; case "redis": img = "redis:7-alpine" }; fmt.Printf("docker run -d -p %d:%d %s\n", hp, cp, img) }</code></pre>`},
					{Title: "Парсинг docker ps", Difficulty: "medium", Description: `<p>Парсинг вывода docker ps — извлеки имя и статус:</p><p>Ввод:</p><pre><code>2
abc123 myapp Up 2 hours
def456 postgres Exited (1) 5 min ago</code></pre><p>Вывод:</p><pre><code>myapp: running
postgres: stopped</code></pre>`, Glossary: []GlossaryItem{{Term: "docker ps", Definition: "Список запущенных контейнеров. -a — включая остановленные."}}, TestCases: []TestCase{{Input: "2\nabc123 myapp Up 2 hours\ndef456 postgres Exited (1) 5 min ago", ExpectedOutput: "myapp: running\npostgres: stopped"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); parts := strings.Fields(scanner.Text())
        name := parts[1]; status := "stopped"
        if strings.Contains(scanner.Text(), "Up") { status = "running" }
        fmt.Printf("%s: %s\n", name, status)
    }
}`, Hints: `<p>strings.Contains(line, \"Up\") → running, иначе stopped.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); p := strings.Fields(scanner.Text()); s := "stopped"; if strings.Contains(scanner.Text(), "Up") { s = "running" }; fmt.Printf("%s: %s\n", p[1], s) } }</code></pre>`},
					{Title: "Генератор .dockerignore", Difficulty: "medium", Description: `<p>Сгенерируй .dockerignore для Go-проекта:</p><p>Ввод: <code>go</code></p><p>Вывод:</p><pre><code>.git
.idea
*.md
tmp/
vendor/</code></pre>`, Glossary: []GlossaryItem{{Term: ".dockerignore", Definition: "Файлы которые НЕ копируются в образ при COPY. Уменьшает размер контекста сборки."}}, TestCases: []TestCase{{Input: "go", ExpectedOutput: ".git\n.idea\n*.md\ntmp/\nvendor/"}},
						StarterCode: `package main
import "fmt"
func main() {
    var lang string; fmt.Scan(&lang)
    ignores := []string{".git", ".idea", "*.md", "tmp/", "vendor/"}
    for _, ig := range ignores { fmt.Println(ig) }
}`, Hints: `<p>Список стандартных игнорируемых файлов для Go.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var l string; fmt.Scan(&l); for _, ig := range []string{".git", ".idea", "*.md", "tmp/", "vendor/"} { fmt.Println(ig) } }</code></pre>`},
					{Title: "Image size analyzer", Difficulty: "hard", Description: `<p>Проанализируй слои образа и предложи оптимизации:</p><p>Ввод:</p><pre><code>3
FROM golang:1.22 800
COPY . . 50
RUN go build 200</code></pre><p>Вывод:</p><pre><code>Total: 1050 MB
Suggestion: use multi-stage build (final ~20MB)</code></pre>`, Glossary: []GlossaryItem{{Term: "docker image history", Definition: "Показывает размер каждого слоя. Помогает найти что раздувает образ."}}, TestCases: []TestCase{{Input: "3\nFROM golang:1.22 800\nCOPY . . 50\nRUN go build 200", ExpectedOutput: "Total: 1050 MB\nSuggestion: use multi-stage build (final ~20MB)"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); total := 0; scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); parts := strings.Fields(scanner.Text()); var size int; fmt.Sscan(parts[len(parts)-1], &size); total += size }
    fmt.Printf("Total: %d MB\n", total)
    if total > 100 { fmt.Println("Suggestion: use multi-stage build (final ~20MB)") }
}`, Hints: `<p>Суммируй размеры слоёв. Если >100MB — предложи multi-stage.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { var n int; fmt.Scan(&n); total := 0; scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); p := strings.Fields(scanner.Text()); var s int; fmt.Sscan(p[len(p)-1], &s); total += s }
    fmt.Printf("Total: %d MB\n", total); if total > 100 { fmt.Println("Suggestion: use multi-stage build (final ~20MB)") } }</code></pre>`},
				},
			},
			{
				Slug: "dockerfile", Title: "Dockerfile — сборка образа", Order: 2,
				Difficulty: "intermediate", Track: "devops",
				Content: `<h1>Dockerfile — рецепт образа</h1>

<h2>Структура</h2>
<pre><code>FROM golang:1.22-alpine    # базовый образ
WORKDIR /app               # рабочая директория
COPY go.mod go.sum ./      # копировать файлы
RUN go mod download        # выполнить команду
COPY . .                   # остальной код
RUN go build -o server     # собрать
EXPOSE 8080                # документация порта
CMD ["./server"]           # команда запуска</code></pre>

<h2>Multi-stage build</h2>
<pre><code># Stage 1: сборка (800MB)
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: runtime (15MB)
FROM alpine:3.19
RUN adduser -D appuser
COPY --from=builder /server /server
USER appuser
CMD ["/server"]</code></pre>

<h2>Layer caching</h2>
<pre><code># ПРАВИЛЬНО: deps кешируются отдельно от кода
COPY go.mod go.sum ./   # меняется редко → кеш
RUN go mod download     # кешируется
COPY . .                # меняется часто
RUN go build

# ПЛОХО: любое изменение кода → пересобрать всё
COPY . .
RUN go mod download && go build</code></pre>

<h2>Best practices</h2>
<ul>
<li><code>USER appuser</code> — не запускай от root</li>
<li>Явные версии: <code>golang:1.22-alpine</code>, не <code>:latest</code></li>
<li><code>-ldflags="-s -w"</code> — уменьшить бинарник на 30%</li>
<li><code>.dockerignore</code> — исключить .git, .idea, tmp/</li>
</ul>`,

				Quiz: []Q{
					{Question: "Зачем multi-stage build?", Options: []string{"Красота", "Финальный образ содержит ТОЛЬКО бинарник (~15MB), без компилятора и исходников (~800MB)", "Быстрее собирается", "Для безопасности"}, Correct: 1, Explanation: "Stage 1 = сборка (нужен компилятор). Stage 2 = runtime (только бинарник + alpine). Меньше образ = быстрее deploy + меньше attack surface."},
					{Question: "Зачем CGO_ENABLED=0?", Options: []string{"Ускоряет", "Статический бинарник без libc — работает в scratch/alpine без зависимостей", "Уменьшает", "Безопасность"}, Correct: 1, Explanation: "С CGO бинарник зависит от libc. Alpine имеет musl (не glibc). CGO_ENABLED=0 = полностью static = работает везде."},
					{Question: "Почему COPY go.mod перед COPY . .?", Options: []string{"Обязательно", "Docker кеширует слои — deps меняются редко и кешируются отдельно от кода", "Быстрее", "Go требует"}, Correct: 1, Explanation: "Docker инвалидирует кеш при изменении COPY. go.mod редко меняется → go mod download кешируется. Код часто → отдельный слой."},
					{Question: "Что делает USER appuser?", Options: []string{"Создаёт пользователя", "Переключает процесс на non-root — если контейнер скомпрометирован, злоумышленник не root", "Логин", "Ничего"}, Correct: 1, Explanation: "Root в контейнере = root на хосте (при escape). USER переключает на непривилегированного пользователя. Обязательно для production."},
					{Question: "CMD vs ENTRYPOINT?", Options: []string{"Одно и то же", "CMD — команда по умолчанию (можно переопределить). ENTRYPOINT — фиксированная (аргументы добавляются)", "CMD быстрее", "ENTRYPOINT устарел"}, Correct: 1, Explanation: "CMD [\"./server\"] — при docker run image другая_команда заменяется. ENTRYPOINT [\"./server\"] — аргументы добавляются: docker run image --port=9090."},
				},
				Tasks: []T{
					{Title: "Dockerfile generator", Difficulty: "easy", Description: `<p>Сгенерируй базовый Dockerfile для Go:</p><p>Ввод: <code>1.22 8080 ./cmd/server</code></p><p>Вывод:</p><pre><code>FROM golang:1.22-alpine
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server
EXPOSE 8080
CMD ["./server"]</code></pre>`, Glossary: []GlossaryItem{{Term: "Dockerfile", Definition: "Текстовый рецепт сборки образа. FROM → WORKDIR → COPY → RUN → CMD."}}, TestCases: []TestCase{{Input: "1.22 8080 ./cmd/server", ExpectedOutput: "FROM golang:1.22-alpine\nWORKDIR /app\nCOPY . .\nRUN go build -o server ./cmd/server\nEXPOSE 8080\nCMD [\"./server\"]"}},
						StarterCode: `package main
import "fmt"
func main() {
    var ver string; var port int; var path string
    fmt.Scan(&ver, &port, &path)
    fmt.Printf("FROM golang:%s-alpine\nWORKDIR /app\nCOPY . .\nRUN go build -o server %s\nEXPOSE %d\nCMD [\"./server\"]\n", ver, path, port)
}`, Hints: `<p>fmt.Printf с подстановкой версии, пути и порта.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var v string; var p int; var path string; fmt.Scan(&v, &p, &path); fmt.Printf("FROM golang:%s-alpine\nWORKDIR /app\nCOPY . .\nRUN go build -o server %s\nEXPOSE %d\nCMD [\"./server\"]\n", v, path, p) }</code></pre>`},
					{Title: "Multi-stage generator", Difficulty: "medium", Description: `<p>Сгенерируй multi-stage Dockerfile:</p><p>Ввод: <code>1.22 ./cmd/server</code></p><p>Вывод:</p><pre><code>FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server
FROM alpine:3.19
COPY --from=builder /server /server
CMD ["/server"]</code></pre>`, Glossary: []GlossaryItem{{Term: "AS builder", Definition: "Именованный stage. COPY --from=builder копирует из него."}}, TestCases: []TestCase{{Input: "1.22 ./cmd/server", ExpectedOutput: "FROM golang:1.22-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 go build -ldflags=\"-s -w\" -o /server ./cmd/server\nFROM alpine:3.19\nCOPY --from=builder /server /server\nCMD [\"/server\"]"}},
						StarterCode: `package main
import "fmt"
func main() {
    var ver, path string; fmt.Scan(&ver, &path)
    fmt.Printf("FROM golang:%s-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 go build -ldflags=\"-s -w\" -o /server %s\nFROM alpine:3.19\nCOPY --from=builder /server /server\nCMD [\"/server\"]\n", ver, path)
}`, Hints: `<p>Два FROM — два stage. COPY --from=builder берёт из первого.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var v, p string; fmt.Scan(&v, &p); fmt.Printf("FROM golang:%s-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN CGO_ENABLED=0 go build -ldflags=\"-s -w\" -o /server %s\nFROM alpine:3.19\nCOPY --from=builder /server /server\nCMD [\"/server\"]\n", v, p) }</code></pre>`},
					{Title: "Layer cache validator", Difficulty: "medium", Description: `<p>Проверь Dockerfile на cache invalidation:</p><p>Ввод: <code>COPY . .\nRUN go mod download\nRUN go build</code></p><p>Вывод: <code>WARN: COPY all before deps</code></p><p>Ввод: <code>COPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN go build</code></p><p>Вывод: <code>OK: good layer order</code></p>`, Glossary: []GlossaryItem{{Term: "Layer cache", Definition: "Docker кеширует каждый слой. Если COPY изменился — все последующие слои пересобираются."}}, TestCases: []TestCase{{Input: "COPY . .\nRUN go mod download\nRUN go build", ExpectedOutput: "WARN: COPY all before deps"}, {Input: "COPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN go build", ExpectedOutput: "OK: good layer order"}},
						StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var lines string; fmt.Scanln(&lines)
    // Split by \n and check order
    parts := strings.Split(lines, "\\n")
    copyAllBeforeDeps := false
    for i, p := range parts {
        if p == "COPY . ." {
            for _, next := range parts[i+1:] {
                if strings.Contains(next, "go mod download") { copyAllBeforeDeps = true }
            }
        }
    }
    if copyAllBeforeDeps { fmt.Println("WARN: COPY all before deps") } else { fmt.Println("OK: good layer order") }
}`, Hints: `<p>Если "COPY . ." появляется перед "go mod download" — плохо.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); line := scanner.Text()
    parts := strings.Split(line, "\\n"); bad := false
    for i, p := range parts { if p == "COPY . ." { for _, next := range parts[i+1:] { if strings.Contains(next, "go mod download") { bad = true } } } }
    if bad { fmt.Println("WARN: COPY all before deps") } else { fmt.Println("OK: good layer order") }
}</code></pre>`},
					{Title: "Dockerfile security checker", Difficulty: "hard", Description: `<p>Проверь Dockerfile на безопасность:</p><p>Ввод: <code>FROM golang:latest\nCOPY . .\nRUN go build\nCMD ["./app"]</code></p><p>Вывод:</p><pre><code>WARN: using :latest tag
WARN: no USER directive
WARN: no multi-stage</code></pre>`, Glossary: []GlossaryItem{{Term: "Security best practices", Definition: "Не :latest, есть USER, multi-stage, нет секретов в ENV."}}, TestCases: []TestCase{{Input: "FROM golang:latest\nCOPY . .\nRUN go build\nCMD [\"./app\"]", ExpectedOutput: "WARN: using :latest tag\nWARN: no USER directive\nWARN: no multi-stage"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    scanner := bufio.NewScanner(os.Stdin); scanner.Scan()
    lines := strings.Split(scanner.Text(), "\\n")
    hasUser := false; hasLatest := false; fromCount := 0
    for _, l := range lines {
        if strings.Contains(l, ":latest") { hasLatest = true }
        if strings.HasPrefix(l, "USER") { hasUser = true }
        if strings.HasPrefix(l, "FROM") { fromCount++ }
    }
    if hasLatest { fmt.Println("WARN: using :latest tag") }
    if !hasUser { fmt.Println("WARN: no USER directive") }
    if fromCount < 2 { fmt.Println("WARN: no multi-stage") }
}`, Hints: `<p>Проверяй :latest, USER, количество FROM.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); lines := strings.Split(scanner.Text(), "\\n")
    hu := false; hl := false; fc := 0
    for _, l := range lines { if strings.Contains(l, ":latest") { hl = true }; if strings.HasPrefix(l, "USER") { hu = true }; if strings.HasPrefix(l, "FROM") { fc++ } }
    if hl { fmt.Println("WARN: using :latest tag") }; if !hu { fmt.Println("WARN: no USER directive") }; if fc < 2 { fmt.Println("WARN: no multi-stage") } }</code></pre>`},
					{Title: "Build size estimator", Difficulty: "hard", Description: `<p>Посчитай размер образа с и без multi-stage:</p><p>Ввод: <code>800 50</code> (base image MB, app MB)</p><p>Вывод:</p><pre><code>Without multi-stage: 850 MB
With multi-stage: 55 MB (alpine 5MB + app 50MB)
Savings: 94%</code></pre>`, Glossary: []GlossaryItem{{Term: "Alpine", Definition: "Минимальный Linux дистрибутив ~5MB. Идеален для финального stage."}}, TestCases: []TestCase{{Input: "800 50", ExpectedOutput: "Without multi-stage: 850 MB\nWith multi-stage: 55 MB (alpine 5MB + app 50MB)\nSavings: 94%"}},
						StarterCode: `package main
import "fmt"
func main() {
    var base, app int; fmt.Scan(&base, &app)
    without := base + app
    with := 5 + app
    savings := 100 - (with*100)/without
    fmt.Printf("Without multi-stage: %d MB\n", without)
    fmt.Printf("With multi-stage: %d MB (alpine 5MB + app %dMB)\n", with, app)
    fmt.Printf("Savings: %d%%\n", savings)
}`, Hints: `<p>Alpine ~5MB. Multi-stage = alpine + бинарник. Savings = (1 - with/without) * 100.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var b, a int; fmt.Scan(&b, &a); w := b+a; ms := 5+a; fmt.Printf("Without multi-stage: %d MB\nWith multi-stage: %d MB (alpine 5MB + app %dMB)\nSavings: %d%%\n", w, ms, a, 100-(ms*100)/w) }</code></pre>`},
				},
			},
			{
				Slug: "docker-compose", Title: "Docker Compose — оркестрация сервисов", Order: 3,
				Difficulty: "intermediate", Track: "devops",
				Content: `<h1>Docker Compose</h1>

<h2>Зачем?</h2>
<p>Реальное приложение = app + DB + Redis + Nginx. Запускать каждый docker run вручную — ад. Compose описывает всё в одном YAML:</p>

<pre><code># docker-compose.yml
services:
  app:
    build: .
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/app
    depends_on:
      db: { condition: service_healthy }

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 5s

volumes:
  pgdata:</code></pre>

<h2>Команды</h2>
<pre><code>docker compose up -d       # запустить всё в фоне
docker compose down        # остановить и удалить
docker compose logs -f app # логи одного сервиса
docker compose ps          # статус
docker compose build       # пересобрать образы
docker compose exec app sh # войти в контейнер</code></pre>

<h2>Networking</h2>
<pre><code># Compose создаёт сеть автоматически
# Сервисы видят друг друга по ИМЕНИ: app → db:5432, app → redis:6379
# Не нужен IP — DNS по имени сервиса</code></pre>

<h2>Volumes</h2>
<pre><code># Named volume — данные сохраняются между перезапусками
volumes:
  pgdata:  # PostgreSQL data

# Bind mount — монтировать хост-папку
volumes:
  - ./migrations:/docker-entrypoint-initdb.d  # SQL при первом запуске
  - ./uploads:/app/uploads                     # файлы пользователей</code></pre>

<h2>Healthchecks</h2>
<pre><code>healthcheck:
  test: ["CMD-SHELL", "pg_isready -U user"]
  interval: 5s    # проверять каждые 5 сек
  timeout: 5s     # таймаут проверки
  retries: 5      # сколько fail до unhealthy

# depends_on + condition: service_healthy
# = app не запустится пока db не станет healthy</code></pre>`,

				Quiz: []Q{
					{Question: "Как сервисы видят друг друга в Compose?", Options: []string{"По IP", "По имени сервиса через встроенный DNS (app→db:5432)", "Через localhost", "Через volumes"}, Correct: 1, Explanation: "Compose создаёт Docker network. Встроенный DNS: имя сервиса = hostname. db:5432 из app — работает без указания IP."},
					{Question: "Зачем healthcheck + depends_on?", Options: []string{"Красота", "App не запустится пока DB реально не готова принимать соединения (а не просто контейнер запущен)", "Обязательно", "Для логов"}, Correct: 1, Explanation: "Без healthcheck depends_on ждёт только запуск контейнера. Postgres может запускаться 5-10 секунд. App пытается подключиться → ошибка. Healthcheck гарантирует readiness."},
					{Question: "Named volume vs bind mount?", Options: []string{"Одно и то же", "Named volume управляется Docker (pgdata). Bind mount монтирует папку хоста (./uploads:/app/uploads)", "Named быстрее", "Bind безопаснее"}, Correct: 1, Explanation: "Named volume: Docker управляет (pgdata сохраняется между docker compose down/up). Bind mount: папка на хосте, удобно для development (hot reload)."},
					{Question: "docker compose down vs stop?", Options: []string{"Одно и то же", "stop — останавливает. down — останавливает И удаляет контейнеры, сети. Volumes сохраняются", "down опаснее", "stop удаляет"}, Correct: 1, Explanation: "down = stop + rm. Контейнеры удаляются, но named volumes и images сохраняются. down -v — удаляет и volumes (осторожно!)."},
					{Question: "environment vs env_file?", Options: []string{"Нет разницы", "environment — inline в yml. env_file — из файла .env. env_file удобнее для секретов (.gitignore)", "env_file устарел", "environment безопаснее"}, Correct: 1, Explanation: "environment видно в docker-compose.yml (может попасть в git). env_file: .env — в .gitignore. Для продакшена: docker secrets или vault."},
				},
				Tasks: []T{
					{Title: "Compose generator", Difficulty: "easy", Description: `<p>Сгенерируй docker-compose.yml для app + postgres:</p><p>Ввод: <code>myapp 8080 mydb user pass</code></p><p>Вывод: services с app и db</p>`, Glossary: []GlossaryItem{{Term: "docker-compose.yml", Definition: "Декларативное описание всех сервисов, сетей, volumes."}}, TestCases: []TestCase{{Input: "myapp 8080 mydb user pass", ExpectedOutput: "services:\n  myapp:\n    build: .\n    ports: [\"8080:8080\"]\n  db:\n    image: postgres:16-alpine\n    environment:\n      POSTGRES_DB: mydb\n      POSTGRES_USER: user\n      POSTGRES_PASSWORD: pass"}},
						StarterCode: `package main
import "fmt"
func main() {
    var app string; var port int; var db, user, pass string
    fmt.Scan(&app, &port, &db, &user, &pass)
    fmt.Printf("services:\n  %s:\n    build: .\n    ports: [\"%d:%d\"]\n  db:\n    image: postgres:16-alpine\n    environment:\n      POSTGRES_DB: %s\n      POSTGRES_USER: %s\n      POSTGRES_PASSWORD: %s\n", app, port, port, db, user, pass)
}`, Hints: `<p>fmt.Printf с YAML-форматированием.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var a string; var p int; var d, u, pw string; fmt.Scan(&a, &p, &d, &u, &pw); fmt.Printf("services:\n  %s:\n    build: .\n    ports: [\"%d:%d\"]\n  db:\n    image: postgres:16-alpine\n    environment:\n      POSTGRES_DB: %s\n      POSTGRES_USER: %s\n      POSTGRES_PASSWORD: %s\n", a, p, p, d, u, pw) }</code></pre>`},
					{Title: "DSN builder из compose env", Difficulty: "easy", Description: `<p>Из переменных compose собери DATABASE_URL:</p><p>Ввод: <code>user pass db 5432 app</code></p><p>Вывод: <code>postgres://user:pass@db:5432/app?sslmode=disable</code></p>`, Glossary: []GlossaryItem{{Term: "DATABASE_URL", Definition: "Стандартная env-переменная для подключения к БД в контейнерах."}}, TestCases: []TestCase{{Input: "user pass db 5432 app", ExpectedOutput: "postgres://user:pass@db:5432/app?sslmode=disable"}},
						StarterCode: `package main
import "fmt"
func main() {
    var user, pass, host string; var port int; var dbname string
    fmt.Scan(&user, &pass, &host, &port, &dbname)
    fmt.Printf("postgres://%s:%s@%s:%d/%s?sslmode=disable\n", user, pass, host, port, dbname)
}`, Hints: `<p>fmt.Sprintf для формирования DSN.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var u,p,h string; var port int; var d string; fmt.Scan(&u,&p,&h,&port,&d); fmt.Printf("postgres://%s:%s@%s:%d/%s?sslmode=disable\n",u,p,h,port,d) }</code></pre>`},
					{Title: "Service dependency resolver", Difficulty: "medium", Description: `<p>Определи порядок запуска по depends_on:</p><p>Ввод:</p><pre><code>3
app db,redis
worker db
db</code></pre><p>Вывод (порядок запуска):</p><pre><code>1: db
2: redis worker
3: app</code></pre>`, Glossary: []GlossaryItem{{Term: "depends_on", Definition: "Порядок запуска. Сервис запускается после зависимостей."}}, TestCases: []TestCase{{Input: "3\napp db,redis\nworker db\ndb", ExpectedOutput: "1: db\n2: redis worker\n3: app"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "sort"; "strings")
func main() {
    var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    deps := make(map[string][]string)
    for i := 0; i < n; i++ { scanner.Scan(); parts := strings.Fields(scanner.Text()); name := parts[0]
        if len(parts) > 1 { deps[name] = strings.Split(parts[1], ",") } else { deps[name] = nil } }
    // Topological sort (simple BFS)
    resolved := make(map[string]int)
    level := 1
    for len(resolved) < len(deps) {
        var current []string
        for name, d := range deps { if _, done := resolved[name]; done { continue }
            allResolved := true; for _, dep := range d { if _, ok := resolved[dep]; !ok { allResolved = false } }
            if allResolved { current = append(current, name) } }
        sort.Strings(current)
        for _, name := range current { resolved[name] = level }
        fmt.Printf("%d: %s\n", level, strings.Join(current, " ")); level++
    }
}`, Hints: `<p>Топологическая сортировка: сервисы без зависимостей → первые. Потом зависящие от них.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "sort"; "strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin); deps := make(map[string][]string)
    for i := 0; i < n; i++ { sc.Scan(); p := strings.Fields(sc.Text()); if len(p) > 1 { deps[p[0]] = strings.Split(p[1], ",") } else { deps[p[0]] = nil } }
    resolved := map[string]int{}; level := 1
    for len(resolved) < len(deps) { var cur []string
        for name, d := range deps { if _, ok := resolved[name]; ok { continue }; ok := true; for _, dep := range d { if _, r := resolved[dep]; !r { ok = false } }; if ok { cur = append(cur, name) } }
        sort.Strings(cur); for _, c := range cur { resolved[c] = level }; fmt.Printf("%d: %s\n", level, strings.Join(cur, " ")); level++ } }</code></pre>`},
					{Title: "Healthcheck config", Difficulty: "medium", Description: `<p>Сгенерируй healthcheck секцию по типу сервиса:</p><p>Ввод: <code>postgres user</code></p><p>Вывод:</p><pre><code>healthcheck:
  test: ["CMD-SHELL", "pg_isready -U user"]
  interval: 5s
  timeout: 5s
  retries: 5</code></pre>`, Glossary: []GlossaryItem{{Term: "healthcheck", Definition: "Проверка готовности сервиса. Docker переводит в healthy/unhealthy."}}, TestCases: []TestCase{{Input: "postgres user", ExpectedOutput: "healthcheck:\n  test: [\"CMD-SHELL\", \"pg_isready -U user\"]\n  interval: 5s\n  timeout: 5s\n  retries: 5"}, {Input: "redis", ExpectedOutput: "healthcheck:\n  test: [\"CMD\", \"redis-cli\", \"ping\"]\n  interval: 5s\n  timeout: 5s\n  retries: 5"}},
						StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var parts string; fmt.Scanln(&parts); args := strings.Fields(parts)
    svc := args[0]; var test string
    switch svc {
    case "postgres": test = fmt.Sprintf("[\"CMD-SHELL\", \"pg_isready -U %s\"]", args[1])
    case "redis": test = "[\"CMD\", \"redis-cli\", \"ping\"]"
    }
    fmt.Printf("healthcheck:\n  test: %s\n  interval: 5s\n  timeout: 5s\n  retries: 5\n", test)
}`, Hints: `<p>switch по типу сервиса: postgres → pg_isready, redis → redis-cli ping.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { sc := bufio.NewScanner(os.Stdin); sc.Scan(); args := strings.Fields(sc.Text()); var t string
    switch args[0] { case "postgres": t = fmt.Sprintf("[\"CMD-SHELL\", \"pg_isready -U %s\"]", args[1]); case "redis": t = "[\"CMD\", \"redis-cli\", \"ping\"]" }
    fmt.Printf("healthcheck:\n  test: %s\n  interval: 5s\n  timeout: 5s\n  retries: 5\n", t) }</code></pre>`},
					{Title: "Full compose для WatchTogether", Difficulty: "hard", Description: `<p>Сгенерируй полный docker-compose с app, db, redis, volumes, healthchecks:</p><p>Ввод: <code>watchtogether 8080</code></p><p>Вывод: полный docker-compose.yml</p>`, Glossary: []GlossaryItem{{Term: "Full stack compose", Definition: "app + db + cache + volumes + healthchecks + depends_on — production-ready."}}, TestCases: []TestCase{{Input: "watchtogether 8080", ExpectedOutput: "services:\n  app:\n    build: .\n    ports: [\"8080:8080\"]\n    depends_on: [db, redis]\n  db:\n    image: postgres:16-alpine\n    volumes: [pgdata:/var/lib/postgresql/data]\n  redis:\n    image: redis:7-alpine\nvolumes:\n  pgdata:"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name string; var port int; fmt.Scan(&name, &port)
    fmt.Printf("services:\n  app:\n    build: .\n    ports: [\"%d:%d\"]\n    depends_on: [db, redis]\n  db:\n    image: postgres:16-alpine\n    volumes: [pgdata:/var/lib/postgresql/data]\n  redis:\n    image: redis:7-alpine\nvolumes:\n  pgdata:\n", port, port)
}`, Hints: `<p>Шаблон с подстановкой name и port. depends_on для порядка запуска.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; var p int; fmt.Scan(&n, &p); fmt.Printf("services:\n  app:\n    build: .\n    ports: [\"%d:%d\"]\n    depends_on: [db, redis]\n  db:\n    image: postgres:16-alpine\n    volumes: [pgdata:/var/lib/postgresql/data]\n  redis:\n    image: redis:7-alpine\nvolumes:\n  pgdata:\n", p, p) }</code></pre>`},
				},
			},
			{
				Slug: "docker-networking", Title: "Docker Networking и Volumes", Order: 4,
				Difficulty: "intermediate", Track: "devops",
				Content: `<h1>Docker Networking</h1>

<h2>Типы сетей</h2>
<pre><code># bridge (по умолчанию) — изолированная сеть контейнеров
# host — контейнер использует сеть хоста напрямую
# none — без сети
# overlay — для Docker Swarm (multi-host)

docker network ls
docker network create mynet
docker run --network mynet --name app myapp
docker run --network mynet --name db postgres</code></pre>

<h2>DNS в Docker</h2>
<pre><code># В одной сети контейнеры видят друг друга по имени:
# app может подключиться к db:5432 без указания IP

# Compose автоматически создаёт сеть: projectname_default
# Все сервисы в одной сети по умолчанию</code></pre>

<h2>Port publishing</h2>
<pre><code># -p hostPort:containerPort
# -p 8080:3000  — снаружи на 8080, внутри на 3000
# -p 127.0.0.1:8080:3000 — только localhost
# Без -p — порт доступен только другим контейнерам в сети</code></pre>

<h2>Volumes — персистентность данных</h2>
<pre><code># Проблема: контейнер удалён → данные потеряны
# Решение: volumes

# Named volume (Docker управляет)
docker volume create pgdata
docker run -v pgdata:/var/lib/postgresql/data postgres

# Bind mount (папка хоста)
docker run -v $(pwd)/uploads:/app/uploads myapp

# tmpfs (только в RAM, для секретов)
docker run --tmpfs /tmp myapp</code></pre>

<h2>Debugging networking</h2>
<pre><code>docker exec app ping db         # проверить связность
docker exec app nslookup db     # DNS resolution
docker network inspect bridge   # детали сети
docker port app                 # маппинг портов</code></pre>`,

				Quiz: []Q{
					{Question: "Как контейнеры видят друг друга в Docker network?", Options: []string{"По IP", "По имени контейнера через встроенный DNS", "Через localhost", "Никак"}, Correct: 1, Explanation: "Docker запускает DNS. В одной network контейнер db доступен как db:5432. Не нужен IP."},
					{Question: "Что произойдёт с данными PostgreSQL если удалить контейнер без volume?", Options: []string{"Сохранятся", "Потеряются — данные в writable layer контейнера удаляются с ним", "Переместятся", "Зависит от настроек"}, Correct: 1, Explanation: "Контейнер = ephemeral. Удалил → всё что было в нём пропало. Volume = данные вне контейнера, переживают docker rm."},
					{Question: "Named volume vs bind mount — когда что?", Options: []string{"Всегда named", "Named для production (DB data). Bind mount для development (code hot-reload)", "Всегда bind", "Нет разницы"}, Correct: 1, Explanation: "Named: Docker управляет, portable, для persistent data. Bind: папка на хосте, для dev (монтируешь код → видишь изменения без rebuild)."},
					{Question: "-p 127.0.0.1:8080:3000 vs -p 8080:3000?", Options: []string{"Одно и то же", "127.0.0.1 = только с localhost. Без IP = доступен извне (0.0.0.0)", "Первый быстрее", "Второй безопаснее"}, Correct: 1, Explanation: "-p 8080:3000 слушает на 0.0.0.0 = доступен с любого IP. -p 127.0.0.1:8080:3000 = только localhost. Для безопасности: БД только 127.0.0.1."},
					{Question: "Как отладить если app не может подключиться к db?", Options: []string{"Перезапустить", "docker exec app ping db + проверить network + DATABASE_URL", "docker restart", "Удалить и создать заново"}, Correct: 1, Explanation: "1) docker exec app ping db — есть ли связь. 2) docker network inspect — в одной ли сети. 3) Проверить DATABASE_URL (host=db, не localhost!)."},
				},
				Tasks: []T{
					{Title: "Network commands", Difficulty: "easy", Description: `<p>Создай сеть и запусти два контейнера в ней:</p><p>Ввод: <code>mynet app db</code></p><p>Вывод:</p><pre><code>docker network create mynet
docker run -d --network mynet --name app myapp
docker run -d --network mynet --name db postgres</code></pre>`, Glossary: []GlossaryItem{{Term: "--network", Definition: "Подключить контейнер к конкретной сети."}}, TestCases: []TestCase{{Input: "mynet app db", ExpectedOutput: "docker network create mynet\ndocker run -d --network mynet --name app myapp\ndocker run -d --network mynet --name db postgres"}},
						StarterCode: `package main
import "fmt"
func main() {
    var net, app, db string; fmt.Scan(&net, &app, &db)
    fmt.Printf("docker network create %s\n", net)
    fmt.Printf("docker run -d --network %s --name %s myapp\n", net, app)
    fmt.Printf("docker run -d --network %s --name %s postgres\n", net, db)
}`, Hints: `<p>--network для подключения к сети. --name для имени.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n, a, d string; fmt.Scan(&n, &a, &d); fmt.Printf("docker network create %s\ndocker run -d --network %s --name %s myapp\ndocker run -d --network %s --name %s postgres\n", n, n, a, n, d) }</code></pre>`},
					{Title: "Volume mount generator", Difficulty: "easy", Description: `<p>Сгенерируй docker run с volumes:</p><p>Ввод: <code>postgres pgdata /var/lib/postgresql/data</code></p><p>Вывод: <code>docker run -d -v pgdata:/var/lib/postgresql/data postgres</code></p>`, Glossary: []GlossaryItem{{Term: "-v name:path", Definition: "Монтировать named volume в path внутри контейнера."}}, TestCases: []TestCase{{Input: "postgres pgdata /var/lib/postgresql/data", ExpectedOutput: "docker run -d -v pgdata:/var/lib/postgresql/data postgres"}},
						StarterCode: `package main
import "fmt"
func main() { var img, vol, path string; fmt.Scan(&img, &vol, &path); fmt.Printf("docker run -d -v %s:%s %s\n", vol, path, img) }`, Hints: `<p>-v volume_name:container_path</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var i, v, p string; fmt.Scan(&i, &v, &p); fmt.Printf("docker run -d -v %s:%s %s\n", v, p, i) }</code></pre>`},
					{Title: "Debug connectivity", Difficulty: "medium", Description: `<p>По симптому выведи команду отладки:</p><p>Ввод: <code>cannot_connect db 5432</code></p><p>Вывод:</p><pre><code>docker exec app ping db
docker exec app nslookup db
docker network inspect bridge</code></pre>`, Glossary: []GlossaryItem{{Term: "docker exec", Definition: "Выполнить команду внутри контейнера. -it для интерактивного."}}, TestCases: []TestCase{{Input: "cannot_connect db 5432", ExpectedOutput: "docker exec app ping db\ndocker exec app nslookup db\ndocker network inspect bridge"}},
						StarterCode: `package main
import "fmt"
func main() {
    var symptom, target string; var port int; fmt.Scan(&symptom, &target, &port)
    fmt.Printf("docker exec app ping %s\n", target)
    fmt.Printf("docker exec app nslookup %s\n", target)
    fmt.Println("docker network inspect bridge")
}`, Hints: `<p>ping для связности, nslookup для DNS, inspect для деталей сети.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var s, t string; var p int; fmt.Scan(&s, &t, &p); fmt.Printf("docker exec app ping %s\ndocker exec app nslookup %s\ndocker network inspect bridge\n", t, t) }</code></pre>`},
					{Title: "Secure port binding", Difficulty: "medium", Description: `<p>Определи безопасный binding для сервиса:</p><p>Ввод: <code>postgres 5432</code> (DB не должна быть доступна извне)</p><p>Вывод: <code>-p 127.0.0.1:5432:5432</code></p><p>Ввод: <code>nginx 80</code> (Web должен быть доступен)</p><p>Вывод: <code>-p 80:80</code></p>`, Glossary: []GlossaryItem{{Term: "127.0.0.1:port:port", Definition: "Доступ только с localhost. Для internal сервисов (DB, Redis)."}}, TestCases: []TestCase{{Input: "postgres 5432", ExpectedOutput: "-p 127.0.0.1:5432:5432"}, {Input: "nginx 80", ExpectedOutput: "-p 80:80"}},
						StarterCode: `package main
import "fmt"
func main() {
    var svc string; var port int; fmt.Scan(&svc, &port)
    internal := map[string]bool{"postgres": true, "redis": true, "mysql": true}
    if internal[svc] { fmt.Printf("-p 127.0.0.1:%d:%d\n", port, port) } else { fmt.Printf("-p %d:%d\n", port, port) }
}`, Hints: `<p>DB/Redis → 127.0.0.1 (только localhost). Web/API → 0.0.0.0 (доступен извне).</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var s string; var p int; fmt.Scan(&s, &p); if s == "postgres" || s == "redis" || s == "mysql" { fmt.Printf("-p 127.0.0.1:%d:%d\n", p, p) } else { fmt.Printf("-p %d:%d\n", p, p) } }</code></pre>`},
					{Title: "Volume backup strategy", Difficulty: "hard", Description: `<p>Сгенерируй команду backup named volume в tar:</p><p>Ввод: <code>pgdata /backups</code></p><p>Вывод: <code>docker run --rm -v pgdata:/data -v /backups:/backup alpine tar czf /backup/pgdata.tar.gz /data</code></p>`, Glossary: []GlossaryItem{{Term: "Volume backup", Definition: "Монтируем volume + backup-папку в temp контейнер → tar архив."}}, TestCases: []TestCase{{Input: "pgdata /backups", ExpectedOutput: "docker run --rm -v pgdata:/data -v /backups:/backup alpine tar czf /backup/pgdata.tar.gz /data"}},
						StarterCode: `package main
import "fmt"
func main() {
    var vol, backupDir string; fmt.Scan(&vol, &backupDir)
    fmt.Printf("docker run --rm -v %s:/data -v %s:/backup alpine tar czf /backup/%s.tar.gz /data\n", vol, backupDir, vol)
}`, Hints: `<p>Temp контейнер с двумя volumes: данные + backup. tar создаёт архив.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var v, b string; fmt.Scan(&v, &b); fmt.Printf("docker run --rm -v %s:/data -v %s:/backup alpine tar czf /backup/%s.tar.gz /data\n", v, b, v) }</code></pre>`},
				},
			},
			{
				Slug: "docker-production", Title: "Docker в production", Order: 5,
				Difficulty: "advanced", Track: "devops",
				Content: `<h1>Docker в production</h1>

<h2>Security best practices</h2>
<pre><code># 1. Non-root user
RUN adduser -D appuser
USER appuser

# 2. Minimal base image
FROM scratch    # 0 bytes — только бинарник
FROM alpine     # 5MB — shell + apk
FROM distroless # Google: без shell, без пакетного менеджера

# 3. No secrets in image
# ПЛОХО: ENV JWT_SECRET=xxx (видно в docker history)
# ХОРОШО: передать при запуске: docker run -e JWT_SECRET=xxx

# 4. Read-only filesystem
docker run --read-only --tmpfs /tmp myapp

# 5. Resource limits
docker run --memory=256m --cpus=0.5 myapp</code></pre>

<h2>Logging</h2>
<pre><code># Приложение пишет в stdout/stderr
# Docker собирает и хранит логи
docker logs app --since 1h    # за последний час
docker logs app 2>&1 | grep ERROR

# Для production: log driver
docker run --log-driver=json-file --log-opt max-size=10m --log-opt max-file=3 myapp</code></pre>

<h2>Health monitoring</h2>
<pre><code>HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Docker показывает статус:
# healthy / unhealthy / starting</code></pre>

<h2>Restart policies</h2>
<pre><code>docker run --restart=always myapp       # всегда перезапускать
docker run --restart=on-failure:5 myapp  # макс 5 перезапусков при ошибке
docker run --restart=unless-stopped      # кроме ручного docker stop</code></pre>

<h2>Multi-platform builds</h2>
<pre><code># Собрать для ARM и x86:
docker buildx build --platform linux/amd64,linux/arm64 -t myapp .</code></pre>`,

				Quiz: []Q{
					{Question: "Зачем scratch/distroless вместо ubuntu?", Options: []string{"Мода", "Минимум attack surface: нет shell, нет пакетного менеджера, нет лишних утилит", "Быстрее", "Меньше памяти"}, Correct: 1, Explanation: "Ubuntu = 70MB + shell + утилиты = если злоумышленник попал в контейнер, у него есть инструменты. Scratch = только бинарник = нечего использовать."},
					{Question: "Почему --read-only?", Options: []string{"Быстрее", "Контейнер не может записывать в FS — если скомпрометирован, не может сохранить malware", "Экономия диска", "Docker требует"}, Correct: 1, Explanation: "Принцип least privilege. Приложение не должно писать в FS (кроме /tmp, логов). --read-only + --tmpfs /tmp = безопасно."},
					{Question: "--restart=always vs unless-stopped?", Options: []string{"Одно и то же", "always перезапускает даже после docker stop + reboot. unless-stopped не перезапускает после ручного stop", "unless-stopped устарел", "always опасен"}, Correct: 1, Explanation: "always = при любом останове (включая reboot). unless-stopped = не перезапускает если вы сами сделали docker stop. Для prod обычно always."},
					{Question: "Куда приложение должно писать логи в Docker?", Options: []string{"В файл /var/log", "В stdout/stderr — Docker собирает и управляет ими через log driver", "В базу данных", "В volume"}, Correct: 1, Explanation: "12-factor app: logs = event stream → stdout. Docker перехватывает, хранит, ротирует. docker logs показывает. Log driver отправляет в ELK/Loki."},
					{Question: "Зачем --memory=256m?", Options: []string{"Ускоряет", "Лимит RAM: если app утекает — OOM kill контейнера, не всего хоста", "Экономия", "Docker требует"}, Correct: 1, Explanation: "Без лимитов memory leak в одном контейнере убьёт весь хост. --memory ограничивает: превысил → OOM Killed. Остальные сервисы живут."},
				},
				Tasks: []T{
					{Title: "Production Dockerfile", Difficulty: "easy", Description: `<p>Сгенерируй production-ready Dockerfile с security best practices:</p><p>Ввод: <code>1.22 ./cmd/server</code></p><p>Вывод: multi-stage + USER + scratch</p>`, Glossary: []GlossaryItem{{Term: "FROM scratch", Definition: "Пустой образ. Только бинарник. 0 attack surface."}}, TestCases: []TestCase{{Input: "1.22 ./cmd/server", ExpectedOutput: "FROM golang:1.22-alpine AS builder\nWORKDIR /app\nCOPY . .\nRUN CGO_ENABLED=0 go build -o /server ./cmd/server\nFROM scratch\nCOPY --from=builder /server /server\nCMD [\"/server\"]"}},
						StarterCode: `package main
import "fmt"
func main() {
    var ver, path string; fmt.Scan(&ver, &path)
    fmt.Printf("FROM golang:%s-alpine AS builder\nWORKDIR /app\nCOPY . .\nRUN CGO_ENABLED=0 go build -o /server %s\nFROM scratch\nCOPY --from=builder /server /server\nCMD [\"/server\"]\n", ver, path)
}`, Hints: `<p>scratch + COPY --from=builder = минимальный безопасный образ.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var v, p string; fmt.Scan(&v, &p); fmt.Printf("FROM golang:%s-alpine AS builder\nWORKDIR /app\nCOPY . .\nRUN CGO_ENABLED=0 go build -o /server %s\nFROM scratch\nCOPY --from=builder /server /server\nCMD [\"/server\"]\n", v, p) }</code></pre>`},
					{Title: "Resource limits", Difficulty: "easy", Description: `<p>Сгенерируй docker run с ресурсными лимитами:</p><p>Ввод: <code>myapp 256 0.5</code> (memory MB, CPU cores)</p><p>Вывод: <code>docker run -d --memory=256m --cpus=0.5 --restart=always myapp</code></p>`, Glossary: []GlossaryItem{{Term: "--memory/--cpus", Definition: "Лимиты ресурсов. Превышение RAM → OOM kill."}}, TestCases: []TestCase{{Input: "myapp 256 0.5", ExpectedOutput: "docker run -d --memory=256m --cpus=0.5 --restart=always myapp"}},
						StarterCode: `package main
import "fmt"
func main() { var name string; var mem int; var cpu float64; fmt.Scan(&name, &mem, &cpu); fmt.Printf("docker run -d --memory=%dm --cpus=%.1f --restart=always %s\n", mem, cpu, name) }`, Hints: `<p>--memory=Nm --cpus=X --restart=always</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; var m int; var c float64; fmt.Scan(&n, &m, &c); fmt.Printf("docker run -d --memory=%dm --cpus=%.1f --restart=always %s\n", m, c, n) }</code></pre>`},
					{Title: "HEALTHCHECK generator", Difficulty: "medium", Description: `<p>Сгенерируй HEALTHCHECK для Dockerfile:</p><p>Ввод: <code>/health 8080 30 3</code> (path, port, interval_sec, retries)</p><p>Вывод:</p><pre><code>HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1</code></pre>`, Glossary: []GlossaryItem{{Term: "HEALTHCHECK", Definition: "Инструкция Dockerfile. Docker периодически проверяет endpoint. healthy/unhealthy."}}, TestCases: []TestCase{{Input: "/health 8080 30 3", ExpectedOutput: "HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1"}},
						StarterCode: `package main
import "fmt"
func main() {
    var path string; var port, interval, retries int
    fmt.Scan(&path, &port, &interval, &retries)
    fmt.Printf("HEALTHCHECK --interval=%ds --timeout=3s --retries=%d CMD wget --quiet --tries=1 --spider http://localhost:%d%s || exit 1\n", interval, retries, port, path)
}`, Hints: `<p>wget --spider проверяет что endpoint отвечает. || exit 1 при ошибке.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var p string; var port, i, r int; fmt.Scan(&p, &port, &i, &r); fmt.Printf("HEALTHCHECK --interval=%ds --timeout=3s --retries=%d CMD wget --quiet --tries=1 --spider http://localhost:%d%s || exit 1\n", i, r, port, p) }</code></pre>`},
					{Title: "Security audit", Difficulty: "hard", Description: `<p>Проверь Docker конфигурацию на безопасность:</p><p>Ввод:</p><pre><code>root latest no-healthcheck 0</code></pre><p>(user, tag, healthcheck, readonly)</p><p>Вывод:</p><pre><code>FAIL: running as root
FAIL: using :latest
FAIL: no healthcheck
FAIL: filesystem is writable</code></pre>`, Glossary: []GlossaryItem{{Term: "Security audit", Definition: "Проверка: non-root, pinned versions, healthcheck, read-only FS, resource limits."}}, TestCases: []TestCase{{Input: "root latest no-healthcheck 0", ExpectedOutput: "FAIL: running as root\nFAIL: using :latest\nFAIL: no healthcheck\nFAIL: filesystem is writable"}, {Input: "appuser 1.22-alpine has-healthcheck 1", ExpectedOutput: "PASS: all checks passed"}},
						StarterCode: `package main
import "fmt"
func main() {
    var user, tag, hc string; var readonly int
    fmt.Scan(&user, &tag, &hc, &readonly)
    fails := 0
    if user == "root" { fmt.Println("FAIL: running as root"); fails++ }
    if tag == "latest" { fmt.Println("FAIL: using :latest"); fails++ }
    if hc == "no-healthcheck" { fmt.Println("FAIL: no healthcheck"); fails++ }
    if readonly == 0 { fmt.Println("FAIL: filesystem is writable"); fails++ }
    if fails == 0 { fmt.Println("PASS: all checks passed") }
}`, Hints: `<p>Проверяй каждый параметр. Если всё ок → PASS.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var u, t, h string; var r int; fmt.Scan(&u, &t, &h, &r); f := 0
    if u == "root" { fmt.Println("FAIL: running as root"); f++ }
    if t == "latest" { fmt.Println("FAIL: using :latest"); f++ }
    if h == "no-healthcheck" { fmt.Println("FAIL: no healthcheck"); f++ }
    if r == 0 { fmt.Println("FAIL: filesystem is writable"); f++ }
    if f == 0 { fmt.Println("PASS: all checks passed") } }</code></pre>`},
					{Title: "Deployment checklist", Difficulty: "hard", Description: `<p>Сгенерируй pre-deploy checklist для Docker production:</p><p>Ввод: <code>myapp</code></p><p>Вывод:</p><pre><code>[x] Multi-stage build
[x] Non-root user
[x] Pinned versions
[x] Healthcheck
[x] Resource limits
[x] .dockerignore
[x] No secrets in image</code></pre>`, Glossary: []GlossaryItem{{Term: "Pre-deploy checklist", Definition: "Список обязательных проверок перед выкаткой в prod."}}, TestCases: []TestCase{{Input: "myapp", ExpectedOutput: "[x] Multi-stage build\n[x] Non-root user\n[x] Pinned versions\n[x] Healthcheck\n[x] Resource limits\n[x] .dockerignore\n[x] No secrets in image"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name string; fmt.Scan(&name)
    checks := []string{"Multi-stage build", "Non-root user", "Pinned versions", "Healthcheck", "Resource limits", ".dockerignore", "No secrets in image"}
    for _, c := range checks { fmt.Printf("[x] %s\n", c) }
}`, Hints: `<p>Список best practices для production Docker.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; fmt.Scan(&n); for _, c := range []string{"Multi-stage build", "Non-root user", "Pinned versions", "Healthcheck", "Resource limits", ".dockerignore", "No secrets in image"} { fmt.Printf("[x] %s\n", c) } }</code></pre>`},
				},
			},
		},
	}
}
