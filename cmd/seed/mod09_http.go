package main

func mod09_http() M {
	return M{
		Slug: "http", Title: "HTTP Сервер", Order: 9,
		Description:   "net/http под капотом, chi router, middleware, JSON API, таймауты, graceful shutdown.",
		Track:         "backend",
		Difficulty:    "intermediate",
		Prerequisites: []string{"packages"},
		Lessons: []L{
			{
				Slug: "net-http-deep", Title: "net/http: как работает HTTP сервер", Order: 1,
				Content: `<h1>HTTP сервер — под капотом</h1>

<h2>Что происходит при http.ListenAndServe(":8080", handler)</h2>
<ol>
<li>Go открывает TCP сокет на порту 8080</li>
<li>Входящее TCP соединение → Go создаёт <strong>горутину</strong> для каждого соединения</li>
<li>Горутина читает HTTP запрос, парсит заголовки</li>
<li>Вызывает твой handler.ServeHTTP(w, r)</li>
<li>После ответа — соединение остаётся открытым (keep-alive) или закрывается</li>
</ol>

<p><strong>Ключевой момент:</strong> каждый запрос обрабатывается в своей горутине. Это значит:
<ul>
<li>Твой handler вызывается <strong>конкурентно</strong> из множества горутин</li>
<li>Если handler читает/пишет общие данные — нужна синхронизация (mutex)</li>
<li>Go может обрабатывать тысячи запросов одновременно</li>
</ul></p>

<h2>http.Handler — единственный интерфейс</h2>
<pre><code>type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// http.HandlerFunc — адаптер (функция → Handler)
type HandlerFunc func(ResponseWriter, *Request)
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }

// Поэтому работает:
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "OK")
})</code></pre>

<h2>http.ResponseWriter — порядок критичен!</h2>
<pre><code>// Внутри ResponseWriter буферизирует заголовки
// При ПЕРВОМ вызове Write() автоматически вызывается WriteHeader(200)

// ПРАВИЛЬНЫЙ порядок:
w.Header().Set("Content-Type", "application/json")  // 1. заголовки
w.WriteHeader(http.StatusCreated)                    // 2. статус
w.Write(data)                                        // 3. тело

// НЕПРАВИЛЬНО — заголовок не отправится:
w.Write(data)                                        // WriteHeader(200) вызван автоматически!
w.Header().Set("X-Custom", "value")                  // ПОЗДНО — заголовки уже ушли</code></pre>

<h2>http.Request — всё о запросе</h2>
<pre><code>r.Method                        // "GET", "POST", "PUT", "DELETE"
r.URL.Path                      // "/api/videos/42"
r.URL.Query().Get("page")       // query параметр ?page=2
r.Header.Get("Authorization")   // заголовок
r.Body                          // io.ReadCloser — тело запроса
r.Context()                     // context (отмена, таймаут, значения)
r.RemoteAddr                    // IP клиента "192.168.1.1:54321"</code></pre>

<h2>Безопасный JSON API</h2>
<pre><code>func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(data); err != nil {
        slog.Error("encode response", "error", err)
    }
}

func decodeJSON(r *http.Request, dst any) error {
    // Лимит тела — защита от abuse
    r.Body = http.MaxBytesReader(nil, r.Body, 1_048_576) // 1МБ

    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields() // отклонить неизвестные поля

    if err := dec.Decode(dst); err != nil {
        return fmt.Errorf("decode json: %w", err)
    }
    return nil
}

func respondError(w http.ResponseWriter, status int, msg string) {
    respondJSON(w, status, map[string]string{"error": msg})
}</code></pre>

<h2>Таймауты — ОБЯЗАТЕЛЬНЫ в проде</h2>
<pre><code>srv := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  15 * time.Second,  // макс. время чтения запроса
    WriteTimeout: 15 * time.Second,  // макс. время записи ответа
    IdleTimeout:  60 * time.Second,  // keep-alive таймаут
}

// Без таймаутов:
// - Slowloris атака: клиент отправляет запрос по байту в секунду → горутина висит вечно
// - Утечка горутин → сервер падает</code></pre>

<h2>Graceful Shutdown</h2>
<pre><code>// Запускаем сервер в горутине
go func() {
    if err := srv.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// Ждём сигнал остановки (Ctrl+C или kill)
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Graceful: дождаться завершения текущих запросов
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
srv.Shutdown(ctx) // новые запросы не принимает, текущие дорабатывают</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: http.ListenAndServe без таймаутов
http.ListenAndServe(":8080", handler) // уязвим к Slowloris!

// ОШИБКА: забыть return после отправки ошибки
func handler(w http.ResponseWriter, r *http.Request) {
    if err != nil {
        http.Error(w, "error", 500)
        // ЗАБЫЛ return! Код ниже продолжит выполняться
    }
    // ... этот код выполнится даже при ошибке
}

// ОШИБКА: паника в handler без Recoverer
// Одна паника убьёт ТОЛЬКО эту горутину, но соединение оборвётся
// middleware.Recoverer ловит panic и возвращает 500</code></pre>`,

				Quiz: []Q{
					{Question: "Сколько горутин создаёт HTTP сервер на 1000 одновременных запросов?", Options: []string{"1", "Зависит от CPU", "1000 — по горутине на каждый запрос", "10"}, Correct: 2, Explanation: "Go создаёт отдельную горутину для каждого HTTP соединения. 1000 запросов = 1000 горутин. Это нормально — горутины легковесные (~2КБ)."},
					{Question: "Почему нужен MaxBytesReader при декодировании JSON?", Options: []string{"Для скорости", "Ограничивает размер тела — без него атакующий исчерпает память сервера", "Для сжатия", "Не нужен"}, Correct: 1, Explanation: "Без лимита клиент может отправить тело в гигабайтах. MaxBytesReader прерывает чтение после указанного лимита."},
					{Question: "Что делает srv.Shutdown(ctx)?", Options: []string{"Мгновенно убивает сервер", "Прекращает приём новых запросов и ждёт завершения текущих", "Перезапускает сервер", "Ничего"}, Correct: 1, Explanation: "Graceful shutdown: новые соединения отклоняются, текущие запросы дорабатывают. Через timeout из context — принудительное завершение."},
					{Question: "Что будет если написать w.Write() а потом w.Header().Set()?", Options: []string{"Нормально", "Заголовок не отправится — Write() уже отправил заголовки", "Ошибка компиляции", "Panic"}, Correct: 1, Explanation: "Первый вызов Write() отправляет заголовки клиенту (включая автоматический WriteHeader(200)). После этого менять заголовки бесполезно."},
					{Question: "Зачем return после http.Error()?", Options: []string{"Для красоты", "Без return выполнение продолжится — может записать данные после ошибки", "http.Error требует return", "Не нужен"}, Correct: 1, Explanation: "http.Error НЕ останавливает функцию. Без return код продолжит выполняться и может записать данные поверх ошибки — двойная запись в ResponseWriter."},
				},
				Tasks: []T{
					{
						Title:      "HTTP роутер на map",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "map[string]func()", Definition: "Маршрутизация: ключ — путь, значение — обработчик. Основа любого HTTP-роутера."},
							{Term: "strings.SplitN(s, \" \", 2)", Definition: "Разбить строку на метод и путь: \"GET /api/health\" → [\"GET\", \"/api/health\"]"},
						},
						Description: `<p>Симулируй HTTP-роутер. На вход — запросы (<code>METHOD /path</code>). Выведи ответ согласно таблице маршрутов:</p>
<ul>
<li><code>GET /health</code> → <code>200 OK</code></li>
<li><code>GET /api/videos</code> → <code>200 [video list]</code></li>
<li><code>POST /api/videos</code> → <code>201 created</code></li>
<li>Остальное → <code>404 not found</code></li>
</ul>`,
						Hints: `<p>Используй map[string]string где ключ "METHOD /path". Проверяй через _, ok := routes[key].</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	routes := map[string]string{
		"GET /health":     "200 OK",
		"GET /api/videos": "200 [video list]",
		"POST /api/videos": "201 created",
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		req := scanner.Text()
		if resp, ok := routes[req]; ok {
			fmt.Println(resp)
		} else {
			fmt.Println("404 not found")
		}
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	routes := map[string]string{
		"GET /health":      "200 OK",
		"GET /api/videos":  "200 [video list]",
		"POST /api/videos": "201 created",
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		req := scanner.Text()
		// TODO: найди req в routes
		// Если есть → выведи ответ, иначе → "404 not found"
		_ = req
		_ = routes
	}
}`,
						TestCases: []TestCase{
							{Input: "GET /health\nGET /api/videos\nPOST /api/videos\nDELETE /api/videos\nGET /unknown", ExpectedOutput: "200 OK\n200 [video list]\n201 created\n404 not found\n404 not found"},
						},
					},
					{
						Title:      "JSON request/response симулятор",
						Difficulty: "hard",
						Glossary: []GlossaryItem{
							{Term: "json.NewDecoder(r).Decode(&v)", Definition: "Потоковый парсинг JSON из Reader. Используется для HTTP body."},
							{Term: "json.NewEncoder(w).Encode(v)", Definition: "Потоковая запись JSON в Writer. Используется для HTTP response."},
						},
						Description: `<p>Симулируй JSON API. Каждая строка — JSON-запрос с полем "action":</p>
<ul>
<li><code>{"action":"create","title":"Matrix","year":1999}</code> → добавь видео, выведи <code>{"id":N,"status":"created"}</code></li>
<li><code>{"action":"list"}</code> → выведи <code>{"count":N}</code></li>
<li><code>{"action":"get","id":1}</code> → выведи <code>{"title":"...","year":N}</code> или <code>{"error":"not found"}</code></li>
</ul>`,
						Hints: `<p>Декодируй в map[string]any. Проверяй action. Храни видео в map[int]Video.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Video struct {
	Title string
	Year  int
}

func main() {
	videos := map[int]Video{}
	nextID := 1
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req map[string]any
		json.Unmarshal([]byte(scanner.Text()), &req)

		switch req["action"] {
		case "create":
			title := req["title"].(string)
			year := int(req["year"].(float64))
			videos[nextID] = Video{title, year}
			resp, _ := json.Marshal(map[string]any{"id": nextID, "status": "created"})
			fmt.Println(string(resp))
			nextID++
		case "list":
			resp, _ := json.Marshal(map[string]int{"count": len(videos)})
			fmt.Println(string(resp))
		case "get":
			id := int(req["id"].(float64))
			if v, ok := videos[id]; ok {
				resp, _ := json.Marshal(map[string]any{"title": v.Title, "year": v.Year})
				fmt.Println(string(resp))
			} else {
				fmt.Println("{\"error\":\"not found\"}")
			}
		}
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Video struct {
	Title string
	Year  int
}

func main() {
	videos := map[int]Video{}
	nextID := 1
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var req map[string]any
		json.Unmarshal([]byte(scanner.Text()), &req)

		switch req["action"] {
		case "create":
			// TODO: извлеки title и year, сохрани видео
			// Выведи {"id":N,"status":"created"}
			_ = nextID
		case "list":
			// TODO: выведи {"count":N}
		case "get":
			// TODO: найди видео по id, выведи или {"error":"not found"}
		}
	}
}`,
						TestCases: []TestCase{
							{Input: "{\"action\":\"create\",\"title\":\"Matrix\",\"year\":1999}\n{\"action\":\"create\",\"title\":\"Inception\",\"year\":2010}\n{\"action\":\"list\"}\n{\"action\":\"get\",\"id\":1}\n{\"action\":\"get\",\"id\":5}", ExpectedOutput: "{\"id\":1,\"status\":\"created\"}\n{\"id\":2,\"status\":\"created\"}\n{\"count\":2}\n{\"title\":\"Matrix\",\"year\":1999}\n{\"error\":\"not found\"}"},
						},
					},
					{
						Title: "HTTP Status Code справочник", Difficulty: "easy",
						Description: `<p>По коду статуса выведи категорию и описание:</p>
<p>Ввод:</p><pre><code>5
200 301 404 500 201</code></pre>
<p>Вывод:</p><pre><code>200: 2xx Success - OK
301: 3xx Redirect - Moved Permanently
404: 4xx Client Error - Not Found
500: 5xx Server Error - Internal Server Error
201: 2xx Success - Created</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "HTTP Status Codes", Definition: "1xx Info, 2xx Success, 3xx Redirect, 4xx Client Error, 5xx Server Error."}},
						TestCases: []TestCase{{Input: "5\n200 301 404 500 201", ExpectedOutput: "200: 2xx Success - OK\n301: 3xx Redirect - Moved Permanently\n404: 4xx Client Error - Not Found\n500: 5xx Server Error - Internal Server Error\n201: 2xx Success - Created"}},
						StarterCode: `package main
import "fmt"
func main() {
    codes := map[int]string{200:"2xx Success - OK",201:"2xx Success - Created",301:"3xx Redirect - Moved Permanently",400:"4xx Client Error - Bad Request",404:"4xx Client Error - Not Found",500:"5xx Server Error - Internal Server Error"}
    var n int; fmt.Scan(&n)
    for i := 0; i < n; i++ { var c int; fmt.Scan(&c); fmt.Printf("%d: %s\n", c, codes[c]) }
}`,
						Hints: `<p>Map code → description.</p>`,
						Solution: `<pre><code>package main
import "fmt"
func main() { c:=map[int]string{200:"2xx Success - OK",201:"2xx Success - Created",301:"3xx Redirect - Moved Permanently",400:"4xx Client Error - Bad Request",404:"4xx Client Error - Not Found",500:"5xx Server Error - Internal Server Error"}
    var n int;fmt.Scan(&n);for i:=0;i<n;i++{var v int;fmt.Scan(&v);fmt.Printf("%d: %s\n",v,c[v])} }</code></pre>`,
					},
					{
						Title: "URL парсер", Difficulty: "medium",
						Description: `<p>Разбери URL на компоненты:</p>
<p>Ввод: <code>http://api.example.com:8080/users/42?name=alice&role=admin</code></p>
<p>Вывод:</p><pre><code>Scheme: http
Host: api.example.com
Port: 8080
Path: /users/42
Query: name=alice&role=admin</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "net/url.Parse", Definition: "Парсит URL в структуру url.URL с полями Scheme, Host, Path, RawQuery."}},
						TestCases: []TestCase{{Input: "http://api.example.com:8080/users/42?name=alice&role=admin", ExpectedOutput: "Scheme: http\nHost: api.example.com\nPort: 8080\nPath: /users/42\nQuery: name=alice&role=admin"}},
						StarterCode: `package main
import ("fmt";"net/url")
func main() {
    var raw string; fmt.Scan(&raw)
    u, _ := url.Parse(raw)
    fmt.Println("Scheme:", u.Scheme)
    fmt.Println("Host:", u.Hostname())
    fmt.Println("Port:", u.Port())
    fmt.Println("Path:", u.Path)
    fmt.Println("Query:", u.RawQuery)
}`,
						Hints: `<p><code>url.Parse(s)</code> → u.Scheme, u.Hostname(), u.Port(), u.Path, u.RawQuery.</p>`,
						Solution: `<pre><code>package main
import ("fmt";"net/url")
func main() { var r string;fmt.Scan(&r);u,_:=url.Parse(r)
    fmt.Println("Scheme:",u.Scheme);fmt.Println("Host:",u.Hostname());fmt.Println("Port:",u.Port());fmt.Println("Path:",u.Path);fmt.Println("Query:",u.RawQuery) }</code></pre>`,
					},
					{
						Title: "Query параметры", Difficulty: "hard",
						Description: `<p>Разбери query string в пары key=value и выведи в алфавитном порядке:</p>
<p>Ввод: <code>page=2&sort=name&limit=10&order=asc</code></p>
<p>Вывод:</p><pre><code>limit=10
order=asc
page=2
sort=name</code></pre>`,
						Glossary: []GlossaryItem{{Term: "url.ParseQuery", Definition: "Парсит query string в map[string][]string. Поддерживает повторяющиеся ключи."}},
						TestCases: []TestCase{
							{Input: "page=2&sort=name&limit=10&order=asc", ExpectedOutput: "limit=10\norder=asc\npage=2\nsort=name"},
						},
						StarterCode: `package main
import ("fmt";"net/url";"sort")
func main() {
    var qs string; fmt.Scan(&qs)
    vals, _ := url.ParseQuery(qs)
    keys := make([]string, 0, len(vals))
    for k := range vals { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys { fmt.Printf("%s=%s\n", k, vals.Get(k)) }
}`,
						Hints: `<p><code>url.ParseQuery(qs)</code> → map. <code>vals.Get(key)</code> — первое значение.</p>`,
						Solution: `<pre><code>package main
import ("fmt";"net/url";"sort")
func main() { var q string;fmt.Scan(&q);v,_:=url.ParseQuery(q);var k []string
    for x:=range v{k=append(k,x)};sort.Strings(k);for _,x:=range k{fmt.Printf("%s=%s\n",x,v.Get(x))} }</code></pre>`,
					},
				},
			},
			{
				Slug: "chi-middleware", Title: "Chi Router, Middleware и Route Groups", Order: 2,
				Content: `<h1>Chi Router — продвинутое использование</h1>

<h2>Middleware — как это работает внутри</h2>
<p>Middleware — это <strong>матрёшка из Handler'ов</strong>:</p>
<pre><code>// Запрос проходит через все middleware по порядку:
// Logger → Recoverer → Auth → твой handler → Auth → Recoverer → Logger

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        slog.Info("→ request", "method", r.Method, "path", r.URL.Path)

        next.ServeHTTP(w, r)  // вызов следующего handler

        slog.Info("← response", "duration", time.Since(start))
    })
}</code></pre>

<h2>Написание своего middleware</h2>
<pre><code>// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Security headers
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        next.ServeHTTP(w, r)
    })
}</code></pre>

<h2>Route Groups — разделение публичного и защищённого</h2>
<pre><code>r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(CORSMiddleware)

r.Route("/api", func(r chi.Router) {
    // Публичные
    r.Get("/health", healthHandler)
    r.Post("/auth/login", loginHandler)
    r.Post("/auth/register", registerHandler)

    // Требуют авторизации
    r.Group(func(r chi.Router) {
        r.Use(AuthMiddleware)
        r.Get("/me", getMeHandler)
        r.Get("/videos", listVideosHandler)
        r.Post("/videos", createVideoHandler)
        r.Route("/rooms", func(r chi.Router) {
            r.Get("/", listRoomsHandler)
            r.Post("/", createRoomHandler)
            r.Get("/{roomID}", getRoomHandler)
        })
    })

    // Только админ
    r.Group(func(r chi.Router) {
        r.Use(AuthMiddleware)
        r.Use(AdminMiddleware)
        r.Delete("/users/{userID}", deleteUserHandler)
    })
})</code></pre>

<h2>Передача данных через context</h2>
<pre><code>// В middleware: положить данные
ctx := context.WithValue(r.Context(), "user_id", userID)
next.ServeHTTP(w, r.WithContext(ctx))

// В handler: достать данные
userID := r.Context().Value("user_id").(int64)

// Лучше использовать свой тип ключа (избежать коллизий)
type contextKey string
const userIDKey contextKey = "user_id"</code></pre>`,

				Quiz: []Q{
					{Question: "В каком порядке выполняются middleware?", Options: []string{"Рандомно", "В порядке r.Use() — снаружи внутрь, потом обратно", "В обратном порядке", "Параллельно"}, Correct: 1, Explanation: "Middleware работает как матрёшка: Logger → Recoverer → Auth → Handler → Auth → Recoverer → Logger. Каждый видит запрос и ответ."},
					{Question: "Зачем нужен middleware Recoverer?", Options: []string{"Для скорости", "Ловит panic в handler и возвращает 500 вместо обрыва соединения", "Для логирования", "Для кеширования"}, Correct: 1, Explanation: "Без Recoverer panic в handler убивает горутину — клиент получит обрыв соединения. Recoverer ловит panic и возвращает корректный 500."},
					{Question: "Зачем собственный тип для ключей context.WithValue?", Options: []string{"Для красоты", "Избежать коллизий — строковые ключи могут совпасть в разных пакетах", "Обязательно по спецификации", "Для типобезопасности"}, Correct: 1, Explanation: "Если два пакета используют строку \"user_id\" как ключ — коллизия. Собственный тип type contextKey string гарантирует уникальность."},
					{Question: "Что такое Route Group в chi?", Options: []string{"Группа серверов", "Набор роутов с общим префиксом и middleware — изолированная подсекция", "Шаблон URL", "Набор handler-ов"}, Correct: 1, Explanation: "r.Route(\"/api\", func(r chi.Router) { r.Use(Auth); r.Get(\"/users\", ...) }) — все роуты внутри получат /api префикс и Auth middleware."},
					{Question: "Как chi передаёт URL-параметры в handler?", Options: []string{"Через query string", "chi.URLParam(r, \"id\") — извлекает параметры из /users/{id}", "Через заголовки", "Через cookies"}, Correct: 1, Explanation: "chi.URLParam(r, \"id\") извлекает значение из пути /users/{id}. Параметр определяется в фигурных скобках при регистрации роута."},
				},
				Tasks: []T{
					{
						Title:      "Цепочка middleware",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "func(next func()) func()", Definition: "Middleware-паттерн: принимает следующий обработчик, возвращает обёрнутый."},
						},
						Description: `<p>Реализуй middleware-цепочку (без HTTP — чистые функции).</p>
<p>Middleware добавляет текст вокруг вызова next:</p>
<ul>
<li><code>Logger</code>: печатает <code>[LOG] before</code> и <code>[LOG] after</code></li>
<li><code>Auth</code>: печатает <code>[AUTH] check</code> перед next</li>
<li><code>Timer</code>: печатает <code>[TIME] start</code> и <code>[TIME] end</code></li>
</ul>
<p>Handler: печатает <code>[HANDLER] process</code></p>
<p>Порядок: Timer → Logger → Auth → Handler</p>
<p>Вывод:</p>
<pre><code>[TIME] start
[LOG] before
[AUTH] check
[HANDLER] process
[LOG] after
[TIME] end</code></pre>`,
						Hints: `<p>Каждый middleware = func(next func()) func(). Собирай цепочку снаружи внутрь: timer(logger(auth(handler))).</p>`,
						Solution: `<pre><code>package main

import "fmt"

func Logger(next func()) func() {
	return func() {
		fmt.Println("[LOG] before")
		next()
		fmt.Println("[LOG] after")
	}
}

func Auth(next func()) func() {
	return func() {
		fmt.Println("[AUTH] check")
		next()
	}
}

func Timer(next func()) func() {
	return func() {
		fmt.Println("[TIME] start")
		next()
		fmt.Println("[TIME] end")
	}
}

func main() {
	handler := func() { fmt.Println("[HANDLER] process") }
	chain := Timer(Logger(Auth(handler)))
	chain()
}</code></pre>`,
						StarterCode: `package main

import "fmt"

func Logger(next func()) func() {
	return func() {
		fmt.Println("[LOG] before")
		next()
		fmt.Println("[LOG] after")
	}
}

func Auth(next func()) func() {
	// TODO: выведи "[AUTH] check", потом вызови next()
	return func() { next() }
}

func Timer(next func()) func() {
	// TODO: "[TIME] start" → next() → "[TIME] end"
	return func() { next() }
}

func main() {
	handler := func() { fmt.Println("[HANDLER] process") }
	// TODO: собери цепочку Timer → Logger → Auth → handler
	chain := handler
	chain()
}`,
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "[TIME] start\n[LOG] before\n[AUTH] check\n[HANDLER] process\n[LOG] after\n[TIME] end"},
						},
					},
					{
						Title: "URL Pattern matcher", Difficulty: "easy",
						Description: `<p>Сопоставь URL с паттернами. <code>/users/{id}</code> совпадает с <code>/users/42</code>:</p>
<p>Ввод:</p><pre><code>2
/users/{id}
/posts/{slug}
3
/users/42
/posts/hello
/other</code></pre>
<p>Вывод:</p><pre><code>/users/42 -> /users/{id} (id=42)
/posts/hello -> /posts/{slug} (slug=hello)
/other -> 404</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "{param}", Definition: "URL-параметр в chi. chi.URLParam(r,\"id\") извлекает значение."}},
						TestCases: []TestCase{{Input: "2\n/users/{id}\n/posts/{slug}\n3\n/users/42\n/posts/hello\n/other", ExpectedOutput: "/users/42 -> /users/{id} (id=42)\n/posts/hello -> /posts/{slug} (slug=hello)\n/other -> 404"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin); sc.Scan(); var np int; fmt.Sscan(sc.Text(),&np)
    ps:=make([]string,np); for i:=range ps{sc.Scan();ps[i]=sc.Text()}
    sc.Scan(); var nu int; fmt.Sscan(sc.Text(),&nu)
    for i:=0;i<nu;i++{sc.Scan();u:=sc.Text();ok:=false
        for _,p:=range ps{pp:=strings.Split(p,"/");up:=strings.Split(u,"/");if len(pp)!=len(up){continue}
            m:=true;pn,pv:="","";for j:=range pp{if strings.HasPrefix(pp[j],"{"){pn=pp[j][1:len(pp[j])-1];pv=up[j]}else if pp[j]!=up[j]{m=false;break}}
            if m{fmt.Printf("%s -> %s (%s=%s)\n",u,p,pn,pv);ok=true;break}};if !ok{fmt.Printf("%s -> 404\n",u)}} }`,
						Hints: `<p>Сегмент в {} — параметр, остальные должны совпасть буквально.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin);sc.Scan();var np int;fmt.Sscan(sc.Text(),&np);ps:=make([]string,np);for i:=range ps{sc.Scan();ps[i]=sc.Text()}
    sc.Scan();var nu int;fmt.Sscan(sc.Text(),&nu);for i:=0;i<nu;i++{sc.Scan();u:=sc.Text();ok:=false
        for _,p:=range ps{pp:=strings.Split(p,"/");up:=strings.Split(u,"/");if len(pp)!=len(up){continue};m:=true;pn,pv:="","";for j:=range pp{if strings.HasPrefix(pp[j],"{"){pn=pp[j][1:len(pp[j])-1];pv=up[j]}else if pp[j]!=up[j]{m=false;break}};if m{fmt.Printf("%s -> %s (%s=%s)\n",u,p,pn,pv);ok=true;break}};if !ok{fmt.Printf("%s -> 404\n",u)}} }</code></pre>`,
					},
					{
						Title: "Парсер HTTP request line", Difficulty: "easy",
						Description: `<p>Разбери строку HTTP-запроса:</p>
<p>Ввод:</p><pre><code>2
GET /api/users HTTP/1.1
POST /api/login HTTP/1.1</code></pre>
<p>Вывод:</p><pre><code>Method: GET, Path: /api/users
Method: POST, Path: /api/login</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "Request line", Definition: "METHOD PATH VERSION — первая строка HTTP-запроса."}},
						TestCases: []TestCase{{Input: "2\nGET /api/users HTTP/1.1\nPOST /api/login HTTP/1.1", ExpectedOutput: "Method: GET, Path: /api/users\nMethod: POST, Path: /api/login"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); p := strings.Fields(sc.Text()); fmt.Printf("Method: %s, Path: %s\n", p[0], p[1]) } }`,
						Hints: `<p>strings.Fields разбивает по пробелам.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());fmt.Printf("Method: %s, Path: %s\n",p[0],p[1])}}</code></pre>`,
					},
					{
						Title: "Rate limiter", Difficulty: "medium",
						Description: `<p>Token bucket: N запросов разрешено. Каждый забирает токен:</p>
<p>Ввод:</p><pre><code>3
5
r1 r2 r3 r4 r5</code></pre>
<p>Вывод:</p><pre><code>r1: ALLOWED (2 left)
r2: ALLOWED (1 left)
r3: ALLOWED (0 left)
r4: DENIED
r5: DENIED</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "Token Bucket", Definition: "Ведро с N токенами. Каждый запрос = -1. Пусто = отклонить."}},
						TestCases: []TestCase{{Input: "3\n5\nr1 r2 r3 r4 r5", ExpectedOutput: "r1: ALLOWED (2 left)\nr2: ALLOWED (1 left)\nr3: ALLOWED (0 left)\nr4: DENIED\nr5: DENIED"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { var limit int; fmt.Scan(&limit); sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); var count int; fmt.Sscan(sc.Text(), &count); sc.Scan(); t := limit
    for _, r := range strings.Fields(sc.Text()) {
        if t > 0 { t--; fmt.Printf("%s: ALLOWED (%d left)\n", r, t) } else { fmt.Printf("%s: DENIED\n", r) } } }`,
						Hints: `<p>Счётчик tokens--. Когда 0 → DENIED.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){var l int;fmt.Scan(&l);sc:=bufio.NewScanner(os.Stdin);sc.Scan();var c int;fmt.Sscan(sc.Text(),&c);sc.Scan();t:=l
    for _,r:=range strings.Fields(sc.Text()){if t>0{t--;fmt.Printf("%s: ALLOWED (%d left)\n",r,t)}else{fmt.Printf("%s: DENIED\n",r)}}}</code></pre>`,
					},
					{
						Title: "REST router simulator", Difficulty: "hard",
						Description: `<p>Зарегистрируй маршруты и обработай запросы. Если метод не совпал → 405:</p>
<p>Ввод:</p><pre><code>2
GET /users list
POST /users create
3
GET /users
POST /users
PUT /users</code></pre>
<p>Вывод:</p><pre><code>200 list
200 create
405 Method Not Allowed</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "405 Method Not Allowed", Definition: "Путь есть, но метод не зарегистрирован для этого пути."}},
						TestCases: []TestCase{{Input: "2\nGET /users list\nPOST /users create\n3\nGET /users\nPOST /users\nPUT /users", ExpectedOutput: "200 list\n200 create\n405 Method Not Allowed"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin); sc.Scan(); var nr int; fmt.Sscan(sc.Text(),&nr)
    type R struct{m,p,h string}; rs:=make([]R,nr)
    for i:=range rs{sc.Scan();f:=strings.Fields(sc.Text());rs[i]=R{f[0],f[1],f[2]}}
    sc.Scan(); var nq int; fmt.Sscan(sc.Text(),&nq)
    for i:=0;i<nq;i++{sc.Scan();f:=strings.Fields(sc.Text());pm,found:=false,false
        for _,r:=range rs{if r.p==f[1]{pm=true;if r.m==f[0]{fmt.Printf("200 %s\n",r.h);found=true;break}}}
        if !found{if pm{fmt.Println("405 Method Not Allowed")}else{fmt.Println("404 Not Found")}}} }`,
						Hints: `<p>Ищи совпадение пути → потом проверяй метод. Путь есть но метод нет → 405.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();var nr int;fmt.Sscan(sc.Text(),&nr);type R struct{m,p,h string};rs:=make([]R,nr)
    for i:=range rs{sc.Scan();f:=strings.Fields(sc.Text());rs[i]=R{f[0],f[1],f[2]}};sc.Scan();var nq int;fmt.Sscan(sc.Text(),&nq)
    for i:=0;i<nq;i++{sc.Scan();f:=strings.Fields(sc.Text());pm,fd:=false,false;for _,r:=range rs{if r.p==f[1]{pm=true;if r.m==f[0]{fmt.Printf("200 %s\n",r.h);fd=true;break}}};if !fd{if pm{fmt.Println("405 Method Not Allowed")}else{fmt.Println("404 Not Found")}}}}</code></pre>`,
					},
				},
			},
		},
	}
}
