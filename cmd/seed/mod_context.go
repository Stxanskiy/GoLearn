package main

// ════════════════════════════════════════════════════════════════
// context.Context — основа асинхронного Go
// Между пакетами и файлами/HTTP
// ════════════════════════════════════════════════════════════════

func mod_context() M {
	return M{
		Slug:          "context",
		Title:         "context.Context — управление жизненным циклом",
		Description:   "Отмена, таймауты, передача значений между горутинами. Используется в каждом HTTP-обработчике и БД-запросе.",
		Order:         100, // will be overridden
		Track:         "shared",
		Difficulty:    "intermediate",
		Prerequisites: []string{"packages"},
		Lessons: []L{
			{
				Slug: "context-basics", Title: "Зачем нужен context и как он работает", Order: 1,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>context.Context — управление жизненным циклом</h1>

<h2>Зачем нужен context?</h2>
<p>Представь: пользователь отправил HTTP-запрос, но закрыл вкладку через 2 секунды. Твой сервер всё ещё делает тяжёлый запрос к БД. <strong>Context</strong> позволяет <em>отменить</em> всю цепочку операций:</p>

<pre><code>// HTTP запрос → handler → service → repository → БД
// Пользователь закрыл вкладку → context отменяется
// → БД-запрос прерывается → горутины завершаются → ресурсы освобождаются</code></pre>

<p><strong>Context</strong> — это объект, который несёт:</p>
<ul>
<li><strong>Сигнал отмены</strong> — "прекрати работу"</li>
<li><strong>Дедлайн</strong> — "у тебя N секунд"</li>
<li><strong>Значения</strong> — request-scoped данные (request ID, user ID)</li>
</ul>

<h2>Под капотом</h2>
<pre><code>type Context interface {
    Deadline() (deadline time.Time, ok bool) // когда истекает
    Done() <-chan struct{}                   // канал: закроется при отмене
    Err() error                             // причина: Canceled или DeadlineExceeded
    Value(key any) any                      // request-scoped значения
}</code></pre>

<p><strong>Done()</strong> возвращает канал, который <em>закрывается</em> при отмене. Это позволяет использовать <code>select</code> для реакции на отмену.</p>

<h2>context.Background() и context.TODO()</h2>
<pre><code>// Background — корневой context. Используй в main(), тестах, top-level
ctx := context.Background()

// TODO — когда ещё не знаешь какой context передать
// Используй временно, потом замени на правильный
ctx := context.TODO()</code></pre>

<h2>WithCancel — ручная отмена</h2>
<pre><code>ctx, cancel := context.WithCancel(context.Background())
defer cancel() // ВСЕГДА defer cancel! Иначе утечка горутин

go func() {
    // Долгая работа...
    select {
    case <-ctx.Done():
        fmt.Println("Отменено:", ctx.Err())
        return
    case result := <-doWork():
        fmt.Println("Готово:", result)
    }
}()

// Через 2 секунды отменяем
time.Sleep(2 * time.Second)
cancel() // закрывает ctx.Done() → горутина завершается</code></pre>

<h2>WithTimeout — автоматическая отмена по времени</h2>
<pre><code>// Дай 3 секунды на операцию
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// pgx, net/http — все принимают context
rows, err := pool.Query(ctx, "SELECT * FROM videos")
// Если запрос займёт > 3 сек → ctx отменится → запрос прервётся

if ctx.Err() == context.DeadlineExceeded {
    fmt.Println("Таймаут!")
}</code></pre>

<h2>WithValue — передача данных</h2>
<pre><code>type ctxKey string

const requestIDKey ctxKey = "requestID"

// Middleware добавляет request ID
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), requestIDKey, uuid.New().String())
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Handler читает request ID
func handler(w http.ResponseWriter, r *http.Request) {
    reqID := r.Context().Value(requestIDKey).(string)
    log.Printf("[%s] handling request", reqID)
}</code></pre>

<p><strong>Правила Value:</strong></p>
<ul>
<li>Используй для request-scoped данных (request ID, user ID, trace ID)</li>
<li>НЕ используй для передачи зависимостей (используй DI через конструктор)</li>
<li>Ключ — свой тип (type ctxKey string), чтобы избежать коллизий</li>
</ul>

<h2>Правило #1: context — всегда первый параметр</h2>
<pre><code>// ✅ Правильно
func GetUser(ctx context.Context, id int) (*User, error) { ... }
func (r *Repo) Save(ctx context.Context, v *Video) error { ... }

// ❌ Неправильно
func GetUser(id int, ctx context.Context) (*User, error) { ... }
func (r *Repo) Save(v *Video) error { ... } // нет context!</code></pre>

<h2>Паттерн: propagation через всю цепочку</h2>
<pre><code>// HTTP handler → Service → Repository → БД
// Context пробрасывается через все слои

func (h *Handler) GetVideos(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // из HTTP-запроса

    videos, err := h.service.ListVideos(ctx) // → service
    // ...
}

func (s *Service) ListVideos(ctx context.Context) ([]Video, error) {
    return s.repo.GetAll(ctx) // → repository
}

func (r *Repo) GetAll(ctx context.Context) ([]Video, error) {
    rows, err := r.pool.Query(ctx, "SELECT ...") // → БД
    // Если пользователь закрыл вкладку → ctx отменён
    // → SQL-запрос прерывается автоматически!
}</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: забыл defer cancel → утечка горутин
ctx, cancel := context.WithTimeout(bg, 5*time.Second)
// cancel() никогда не вызван → таймер не освобождён!

// ОШИБКА: хранить context в структуре
type Server struct {
    ctx context.Context // ❌ НИКОГДА! Context — per-request
}

// ОШИБКА: передать context.Background() вместо реального
func handler(w http.ResponseWriter, r *http.Request) {
    data := fetchData(context.Background()) // ❌ не отменится!
    data := fetchData(r.Context())          // ✅ связан с запросом
}</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Зачем нужен context.Context?",
						Options:     []string{"Для хранения конфигурации", "Для отмены операций, таймаутов и передачи request-scoped данных через цепочку вызовов", "Для логирования", "Для тестирования"},
						Correct:     1,
						Explanation: "Context — механизм отмены и дедлайнов. Когда пользователь закрыл вкладку, context.Done() срабатывает и все операции в цепочке могут прерваться.",
					},
					{
						Question:    "Что происходит когда ctx.Done() закрывается?",
						Options:     []string{"Программа завершается", "Все select-case, слушающие ctx.Done(), срабатывают — можно корректно завершить работу", "Ничего", "Паника"},
						Correct:     1,
						Explanation: "Done() возвращает канал. При отмене канал закрывается → все горутины, ждущие <-ctx.Done(), продолжают выполнение и могут сделать cleanup.",
					},
					{
						Question:    "Почему нельзя хранить context в структуре?",
						Options:     []string{"Это не компилируется", "Context — per-request, а структура живёт дольше. Хранение → утечки и неправильная отмена", "Можно, это рекомендация", "Занимает много памяти"},
						Correct:     1,
						Explanation: "Context привязан к одному запросу/операции. Хранение в struct привяжет все операции к одному context, и отмена одного запроса отменит все остальные.",
					},
					{
						Question:    "Где должен стоять context в сигнатуре функции?",
						Options:     []string{"Последним параметром", "Первым параметром — ctx context.Context", "В возвращаемых значениях", "В любом месте"},
						Correct:     1,
						Explanation: "Конвенция Go: context всегда первый параметр, имя ctx. Это стандарт, которому следуют stdlib, pgx, chi и все популярные библиотеки.",
					},
					{
						Question:    "Чем context.WithTimeout отличается от context.WithDeadline?",
						Options:     []string{"Ничем", "WithTimeout принимает длительность (5s), WithDeadline — конкретный момент времени. Оба отменяют context.", "WithTimeout быстрее", "WithDeadline для БД"},
						Correct:     1,
						Explanation: "WithTimeout(ctx, 5*time.Second) = WithDeadline(ctx, time.Now().Add(5s)). WithTimeout удобнее для 'через N секунд', WithDeadline — для 'до конкретного момента'.",
					},
				},
				Tasks: []T{
					{
						Title:      "Таймаут операции",
						Difficulty: "easy",
						Glossary: []GlossaryItem{
							{Term: "context.WithTimeout(parent, d)", Definition: "Создаёт context, который автоматически отменяется через d. Возвращает (ctx, cancel)."},
							{Term: "ctx.Err()", Definition: "Возвращает причину отмены: context.Canceled или context.DeadlineExceeded."},
						},
						Description: `<p>Симулируй операцию с таймаутом. На вход: таймаут (мс) и время операции (мс).</p>
<p>Если операция успевает — выведи <code>OK: done in Nms</code>, иначе — <code>TIMEOUT: deadline exceeded</code>.</p>
<p>Ввод: <code>100 50</code> (таймаут 100мс, операция 50мс)</p>
<p>Вывод: <code>OK: done in 50ms</code></p>`,
						Hints: `<p>Используй <code>context.WithTimeout</code> + <code>select</code> с <code>time.After(operationTime)</code> и <code>ctx.Done()</code>.</p>`,
						Solution: `<pre><code>package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	var timeoutMs, operationMs int
	fmt.Scan(&timeoutMs, &operationMs)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(time.Duration(operationMs) * time.Millisecond):
		fmt.Printf("OK: done in %dms\n", operationMs)
	case <-ctx.Done():
		fmt.Println("TIMEOUT: deadline exceeded")
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	var timeoutMs, operationMs int
	fmt.Scan(&timeoutMs, &operationMs)

	// TODO: создай context с таймаутом timeoutMs
	// ctx, cancel := context.WithTimeout(...)
	// defer cancel()

	// TODO: select между операцией и отменой
	// case <-time.After(operationTime): → OK
	// case <-ctx.Done(): → TIMEOUT

	_ = timeoutMs
	_ = operationMs
}`,
						TestCases: []TestCase{
							{Input: "100 50", ExpectedOutput: "OK: done in 50ms"},
							{Input: "50 100", ExpectedOutput: "TIMEOUT: deadline exceeded"},
						},
					},
					{
						Title:      "Context value — request ID",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "context.WithValue(parent, key, val)", Definition: "Создаёт дочерний context с key=val. Используй свой тип для ключа."},
							{Term: "ctx.Value(key)", Definition: "Извлекает значение по ключу. Возвращает any — нужен type assertion."},
						},
						Description: `<p>Симулируй цепочку вызовов с request ID через context.</p>
<p>Ввод: request ID на каждой строке. Для каждого выведи:</p>
<pre><code>[REQ-ID] handler → service → repository</code></pre>
<p>Ввод: <code>abc-123</code></p>
<p>Вывод: <code>[abc-123] handler → service → repository</code></p>`,
						Hints: `<p>Создай context.WithValue с request ID. Пробрось через handler→service→repo. В каждом слое читай ctx.Value(key).</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type ctxKey string

const reqIDKey ctxKey = "requestID"

func handler(ctx context.Context) string {
	reqID := ctx.Value(reqIDKey).(string)
	result := service(ctx)
	return fmt.Sprintf("[%s] handler → %s", reqID, result)
}

func service(ctx context.Context) string {
	result := repository(ctx)
	return fmt.Sprintf("service → %s", result)
}

func repository(ctx context.Context) string {
	return "repository"
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		reqID := scanner.Text()
		ctx := context.WithValue(context.Background(), reqIDKey, reqID)
		fmt.Println(handler(ctx))
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
)

type ctxKey string
const reqIDKey ctxKey = "requestID"

func handler(ctx context.Context) string {
	// TODO: извлеки reqID из ctx через ctx.Value(reqIDKey).(string)
	// Вызови service(ctx) и верни "[reqID] handler → ..."
	return ""
}

func service(ctx context.Context) string {
	// Вызови repository(ctx) и верни "service → ..."
	result := repository(ctx)
	return fmt.Sprintf("service → %s", result)
}

func repository(ctx context.Context) string {
	return "repository"
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		reqID := scanner.Text()
		ctx := context.WithValue(context.Background(), reqIDKey, reqID)
		fmt.Println(handler(ctx))
	}
}`,
						TestCases: []TestCase{
							{Input: "abc-123", ExpectedOutput: "[abc-123] handler → service → repository"},
							{Input: "req-42\nreq-99", ExpectedOutput: "[req-42] handler → service → repository\n[req-99] handler → service → repository"},
						},
					},
					{
						Title:      "Отмена context вручную",
						Difficulty: "easy",
						Description: `<p>Создай context с cancel, запусти "работу" в горутине, отмени через cancel() и покажи что горутина завершилась:</p>
<p>Ввод: (нет)</p>
<p>Вывод:</p><pre><code>Worker started
Context cancelled
Worker stopped</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "context.WithCancel", Definition: "Возвращает ctx и cancel(). Вызов cancel() закрывает ctx.Done() канал."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "Worker started\nContext cancelled\nWorker stopped"},
						},
						StarterCode: `package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context, done chan struct{}) {
    fmt.Println("Worker started")
    <-ctx.Done()
    fmt.Println("Worker stopped")
    close(done)
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})
    go worker(ctx, done)
    time.Sleep(10 * time.Millisecond)
    cancel()
    fmt.Println("Context cancelled")
    <-done
}`,
						Hints:    `<p><code>cancel()</code> закрывает <code>ctx.Done()</code> → горутина выходит из <code><-ctx.Done()</code>.</p>`,
						Solution: `<pre><code>package main

import (
    "context"
    "fmt"
    "time"
)

func worker(ctx context.Context, done chan struct{}) {
    fmt.Println("Worker started")
    <-ctx.Done()
    fmt.Println("Worker stopped")
    close(done)
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    done := make(chan struct{})
    go worker(ctx, done)
    time.Sleep(10 * time.Millisecond)
    cancel()
    fmt.Println("Context cancelled")
    <-done
}</code></pre>`,
					},
					{
						Title:      "Таймаут с fallback",
						Difficulty: "medium",
						Description: `<p>Функция <code>fetchData</code> эмулирует медленный запрос. Если не успела за таймаут — вернуть fallback:</p>
<p>Ввод: <code>50 100</code> (задержка_ms таймаут_ms)</p>
<p>Вывод: <code>data: result</code></p>
<p>Ввод: <code>200 50</code></p>
<p>Вывод: <code>data: fallback (timeout)</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "select + ctx.Done()", Definition: "select { case result := <-ch: ... case <-ctx.Done(): return fallback } — стандартный паттерн таймаута."},
						},
						TestCases: []TestCase{
							{Input: "50 100", ExpectedOutput: "data: result"},
							{Input: "200 50", ExpectedOutput: "data: fallback (timeout)"},
						},
						StarterCode: `package main

import (
    "context"
    "fmt"
    "time"
)

func fetchData(ctx context.Context, delay time.Duration) string {
    ch := make(chan string, 1)
    go func() {
        time.Sleep(delay)
        ch <- "result"
    }()
    select {
    case res := <-ch:
        return res
    case <-ctx.Done():
        return "fallback (timeout)"
    }
}

func main() {
    var delayMs, timeoutMs int
    fmt.Scan(&delayMs, &timeoutMs)
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
    defer cancel()
    fmt.Println("data:", fetchData(ctx, time.Duration(delayMs)*time.Millisecond))
}`,
						Hints:    `<p>select ждёт первый готовый канал: либо результат, либо ctx.Done() (таймаут).</p>`,
						Solution: `<pre><code>package main

import ("context";"fmt";"time")

func fetchData(ctx context.Context, delay time.Duration) string {
    ch := make(chan string, 1)
    go func() { time.Sleep(delay); ch <- "result" }()
    select {
    case r := <-ch: return r
    case <-ctx.Done(): return "fallback (timeout)"
    }
}

func main() {
    var d, t int; fmt.Scan(&d, &t)
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(t)*time.Millisecond)
    defer cancel()
    fmt.Println("data:", fetchData(ctx, time.Duration(d)*time.Millisecond))
}</code></pre>`,
					},
					{
						Title:      "Context с несколькими значениями",
						Difficulty: "hard",
						Description: `<p>Создай middleware-цепочку: каждый слой добавляет значение в context. Финальный handler читает все:</p>
<p>Ввод: <code>user-123 admin req-abc</code> (userID role requestID)</p>
<p>Вывод: <code>Request req-abc by user-123 (role: admin)</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "context.WithValue", Definition: "Добавляет key-value в context. Каждый вызов создаёт новый слой (linked list)."},
							{Term: "type contextKey string", Definition: "Кастомный тип для ключей — предотвращает коллизии между пакетами."},
						},
						TestCases: []TestCase{
							{Input: "user-123 admin req-abc", ExpectedOutput: "Request req-abc by user-123 (role: admin)"},
						},
						StarterCode: `package main

import (
    "context"
    "fmt"
)

type contextKey string

const (
    userIDKey    contextKey = "userID"
    roleKey      contextKey = "role"
    requestIDKey contextKey = "requestID"
)

func withUserID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, userIDKey, id)
}

func withRole(ctx context.Context, role string) context.Context {
    return context.WithValue(ctx, roleKey, role)
}

func withRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func handler(ctx context.Context) {
    uid := ctx.Value(userIDKey).(string)
    role := ctx.Value(roleKey).(string)
    rid := ctx.Value(requestIDKey).(string)
    fmt.Printf("Request %s by %s (role: %s)\n", rid, uid, role)
}

func main() {
    var userID, role, reqID string
    fmt.Scan(&userID, &role, &reqID)
    ctx := context.Background()
    ctx = withUserID(ctx, userID)
    ctx = withRole(ctx, role)
    ctx = withRequestID(ctx, reqID)
    handler(ctx)
}`,
						Hints:    `<p>Каждый <code>context.WithValue</code> оборачивает предыдущий context. Все значения доступны через <code>ctx.Value(key)</code>.</p>`,
						Solution: `<pre><code>package main

import ("context";"fmt")

type contextKey string
const (userIDKey contextKey = "userID"; roleKey contextKey = "role"; requestIDKey contextKey = "requestID")

func main() {
    var uid, role, rid string; fmt.Scan(&uid, &role, &rid)
    ctx := context.WithValue(context.WithValue(context.WithValue(context.Background(), userIDKey, uid), roleKey, role), requestIDKey, rid)
    fmt.Printf("Request %s by %s (role: %s)\n", ctx.Value(requestIDKey), ctx.Value(userIDKey), ctx.Value(roleKey))
}</code></pre>`,
					},
				},
			},
		},
	}
}
