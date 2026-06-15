package main

func mod10_database() M {
	m := M{
		Slug: "database", Title: "PostgreSQL и pgx", Order: 10,
		Description: "SQL глубоко, pgx driver, connection pool, транзакции, миграции, индексы, N+1.",
		Track: "backend", Difficulty: "intermediate", Prerequisites: []string{"http"},
		Lessons: []L{
			{
				Slug: "sql-fundamentals", Title: "SQL: основы и подводные камни", Order: 1,
				Content: `<h1>SQL для Go-разработчика</h1>

<h2>Под капотом: как БД выполняет запрос</h2>
<ol>
<li><strong>Парсинг</strong> — SQL превращается в дерево разбора</li>
<li><strong>Планирование</strong> — оптимизатор выбирает план (индексы, порядок JOIN)</li>
<li><strong>Выполнение</strong> — данные читаются с диска/кеша</li>
</ol>
<pre><code>-- Посмотреть план запроса
EXPLAIN ANALYZE SELECT * FROM videos WHERE title LIKE '%Матрица%';
-- Seq Scan → полное сканирование (медленно на больших таблицах!)
-- Index Scan → использует индекс (быстро)</code></pre>

<h2>Индексы — ускорение запросов</h2>
<pre><code>-- БЕЗ индекса: полное сканирование O(N)
-- С индексом: поиск O(log N)

CREATE INDEX idx_videos_title ON videos(title);
CREATE INDEX idx_rooms_active ON rooms(is_active) WHERE is_active = true; -- partial index

-- Когда индекс НЕ помогает:
-- 1. Таблица маленькая (< 1000 строк)
-- 2. Запрос возвращает > 10% таблицы
-- 3. LIKE '%text%' (начинается с %) — нужен полнотекстовый поиск</code></pre>

<h2>JOIN — объединение таблиц</h2>
<pre><code>-- Комнаты с названием видео
SELECT r.id, r.name, v.title
FROM rooms r
JOIN videos v ON r.video_id = v.id
WHERE r.is_active = true;

-- LEFT JOIN — все комнаты, даже без видео
SELECT r.id, r.name, COALESCE(v.title, 'Нет видео')
FROM rooms r
LEFT JOIN videos v ON r.video_id = v.id;</code></pre>

<h2>Проблема N+1</h2>
<pre><code>// ПЛОХО: N+1 запросов
rooms, _ := repo.GetAllRooms(ctx)          // 1 запрос
for _, room := range rooms {
    video, _ := repo.GetVideo(ctx, room.VideoID) // N запросов!
}

// ХОРОШО: 1 запрос с JOIN
SELECT r.*, v.title as video_title
FROM rooms r JOIN videos v ON r.video_id = v.id</code></pre>

<h2>Транзакции</h2>
<pre><code>tx, err := pool.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx) // безопасно — no-op после Commit

_, err = tx.Exec(ctx, "UPDATE videos SET views = views + 1 WHERE id = $1", videoID)
if err != nil { return err }

_, err = tx.Exec(ctx, "INSERT INTO view_log (video_id) VALUES ($1)", videoID)
if err != nil { return err }

return tx.Commit(ctx) // атомарно: или оба, или ни одного</code></pre>

<h2>pgx: connection pool</h2>
<pre><code>pool, err := pgxpool.New(ctx, databaseURL)
// Pool автоматически:
// - Создаёт соединения по мере необходимости
// - Переиспользует свободные соединения
// - Закрывает протухшие
// - Ограничивает максимум (по умолчанию: CPU * 4)

// Настройка:
cfg, _ := pgxpool.ParseConfig(databaseURL)
cfg.MaxConns = 20
cfg.MinConns = 5
cfg.MaxConnLifetime = time.Hour</code></pre>

<h2>Миграции</h2>
<pre><code># Установка
brew install golang-migrate

# Создать миграцию
migrate create -ext sql -dir migrations -seq add_rooms

# Результат: 000002_add_rooms.up.sql + 000002_add_rooms.down.sql

# up.sql — применить изменения
CREATE TABLE rooms (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    video_id INT REFERENCES videos(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

# down.sql — откатить
DROP TABLE IF EXISTS rooms;

# Применить
migrate -path migrations -database "$DATABASE_URL" up

# Откатить последнюю
migrate -path migrations -database "$DATABASE_URL" down 1</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: SQL инъекция
query := "SELECT * FROM videos WHERE title = '" + userInput + "'"
// userInput: "'; DROP TABLE videos; --"

// ОШИБКА: забыть rows.Close()
rows, _ := pool.Query(ctx, "SELECT ...")
// defer rows.Close() — ОБЯЗАТЕЛЬНО! Иначе утечка соединений

// ОШИБКА: игнорировать rows.Err()
for rows.Next() { ... }
// if err := rows.Err(); err != nil { ... } — может быть ошибка сети!</code></pre>`,

				Quiz: []Q{
					{Question: "Что такое N+1 проблема?", Options: []string{"Ошибка в SQL", "Один запрос на список + N запросов на каждый элемент — вместо одного JOIN", "Проблема с индексами", "Утечка памяти"}, Correct: 1, Explanation: "N+1: получаешь список (1 запрос), потом для каждого элемента делаешь отдельный запрос (N). Решение: JOIN в одном запросе."},
					{Question: "Зачем defer tx.Rollback() если мы делаем Commit?", Options: []string{"Не нужен", "Если ошибка до Commit — Rollback отменит изменения. После Commit — Rollback = no-op", "Для скорости", "Двойная защита"}, Correct: 1, Explanation: "Если функция вернётся с ошибкой до Commit (любой return err), defer Rollback отменит незакоммиченные изменения. После Commit Rollback ничего не делает."},
					{Question: "Когда индекс НЕ помогает?", Options: []string{"Всегда помогает", "Маленькая таблица, запрос LIKE '%text%', возврат >10% строк", "На SELECT", "На INSERT"}, Correct: 1, Explanation: "Индекс — дополнительная структура. На маленьких таблицах seq scan быстрее. LIKE '%text%' не может использовать B-tree индекс."},
					{Question: "Зачем defer rows.Close()?", Options: []string{"Для красоты", "Без этого соединение из pool не вернётся — утечка соединений", "Не нужен — GC закроет", "Для коммита"}, Correct: 1, Explanation: "rows держит соединение из pool. Без Close() оно не вернётся в pool. После нескольких утечек pool исчерпается и приложение зависнет."},
					{Question: "Что делает EXPLAIN ANALYZE?", Options: []string{"Удаляет данные", "Показывает план выполнения запроса с реальным временем — для оптимизации", "Создаёт индекс", "Анализирует таблицу"}, Correct: 1, Explanation: "EXPLAIN ANALYZE выполняет запрос и показывает: какой план выбрал оптимизатор, время каждого шага, использовались ли индексы."},
				},
				Tasks: []T{
					{
						Title: "SQL-билдер с пагинацией",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "LIMIT $1 OFFSET $2", Definition: "Пагинация в PostgreSQL. LIMIT — сколько записей, OFFSET — сколько пропустить."},
							{Term: "$1, $2", Definition: "Параметризованные запросы. Защита от SQL-инъекций. Никогда не конкатенируй!"},
						},
						Description: `<p>Построй SQL-запросы из параметров. На вход — описания запросов:</p>
<ul>
<li><code>SELECT videos page=2 limit=10</code> → <code>SELECT * FROM videos LIMIT 10 OFFSET 10</code></li>
<li><code>SELECT rooms WHERE active=true</code> → <code>SELECT * FROM rooms WHERE active = $1</code></li>
<li><code>INSERT videos title year</code> → <code>INSERT INTO videos (title, year) VALUES ($1, $2) RETURNING id</code></li>
</ul>`,
						Hints: `<p>Парси команду через strings.Fields. OFFSET = (page-1)*limit. Для INSERT считай поля и генерируй $1, $2, ...</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		cmd, table := parts[0], parts[1]

		switch cmd {
		case "SELECT":
			if len(parts) > 2 && parts[2] == "WHERE" {
				fmt.Printf("SELECT * FROM %s WHERE %s = $1\n", table, strings.Split(parts[3], "=")[0])
			} else if len(parts) > 2 {
				var page, limit int
				for _, p := range parts[2:] {
					kv := strings.Split(p, "=")
					if kv[0] == "page" { page, _ = strconv.Atoi(kv[1]) }
					if kv[0] == "limit" { limit, _ = strconv.Atoi(kv[1]) }
				}
				offset := (page - 1) * limit
				fmt.Printf("SELECT * FROM %s LIMIT %d OFFSET %d\n", table, limit, offset)
			}
		case "INSERT":
			fields := parts[2:]
			params := make([]string, len(fields))
			for i := range fields { params[i] = fmt.Sprintf("$%d", i+1) }
			fmt.Printf("INSERT INTO %s (%s) VALUES (%s) RETURNING id\n",
				table, strings.Join(fields, ", "), strings.Join(params, ", "))
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
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		cmd, table := parts[0], parts[1]

		switch cmd {
		case "SELECT":
			// TODO: если есть WHERE → "SELECT * FROM table WHERE field = $1"
			// TODO: если есть page/limit → "SELECT * FROM table LIMIT N OFFSET M"
			// OFFSET = (page-1) * limit
			_ = table
		case "INSERT":
			// TODO: fields → "INSERT INTO table (f1, f2) VALUES ($1, $2) RETURNING id"
			_ = table
		}
	}
}`,
						TestCases: []TestCase{
							{Input: "SELECT videos page=2 limit=10\nINSERT videos title year\nSELECT rooms WHERE active=true", ExpectedOutput: "SELECT * FROM videos LIMIT 10 OFFSET 10\nINSERT INTO videos (title, year) VALUES ($1, $2) RETURNING id\nSELECT * FROM rooms WHERE active = $1"},
						},
					},
					{
						Title: "Детектор N+1 проблемы",
						Difficulty: "hard",
						Glossary: []GlossaryItem{
							{Term: "N+1 problem", Definition: "1 запрос на список + N запросов на каждый элемент. Решение: JOIN или batch-запрос."},
						},
						Description: `<p>На вход — лог SQL-запросов. Найди N+1: когда один SELECT ALL сопровождается множеством SELECT по ID.</p>
<p>Ввод:</p>
<pre><code>SELECT * FROM rooms
SELECT * FROM videos WHERE id = 1
SELECT * FROM videos WHERE id = 2
SELECT * FROM videos WHERE id = 3
SELECT * FROM users WHERE id = 5</code></pre>
<p>Вывод:</p>
<pre><code>N+1 DETECTED: videos (3 queries after rooms)</code></pre>
<p>Логика: SELECT * FROM table → "list". SELECT * FROM table WHERE id = N → "single". Если >1 single подряд для одной таблицы → N+1.</p>`,
						Hints: `<p>Парси таблицу из каждого запроса. Отслеживай: после list-запроса считай single-запросы по таблице. Если count > 1 → N+1.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var listTable string
	counts := map[string]int{}
	afterList := ""

	for scanner.Scan() {
		q := scanner.Text()
		parts := strings.Fields(q)
		// SELECT * FROM table [WHERE id = N]
		if len(parts) >= 4 && parts[0] == "SELECT" {
			table := parts[3]
			if !strings.Contains(q, "WHERE") {
				// List query
				if afterList != "" && counts[afterList] > 1 {
					fmt.Printf("N+1 DETECTED: %s (%d queries after %s)\n", afterList, counts[afterList], listTable)
				}
				listTable = table
				afterList = ""
				counts = map[string]int{}
			} else {
				// Single query
				counts[table]++
				if afterList == "" || afterList == table {
					afterList = table
				}
			}
		}
	}
	if afterList != "" && counts[afterList] > 1 {
		fmt.Printf("N+1 DETECTED: %s (%d queries after %s)\n", afterList, counts[afterList], listTable)
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
	var listTable string
	counts := map[string]int{}

	for scanner.Scan() {
		q := scanner.Text()
		parts := strings.Fields(q)
		if len(parts) < 4 {
			continue
		}
		table := parts[3]

		if !strings.Contains(q, "WHERE") {
			// TODO: это list-запрос. Проверь предыдущие counts.
			// Если для какой-то таблицы > 1 single-запрос → N+1
			listTable = table
			counts = map[string]int{}
		} else {
			// TODO: single-запрос. Увеличь counts[table]
		}
	}

	// TODO: проверь финальные counts
	_ = listTable
	_ = counts
}`,
						TestCases: []TestCase{
							{Input: "SELECT * FROM rooms\nSELECT * FROM videos WHERE id = 1\nSELECT * FROM videos WHERE id = 2\nSELECT * FROM videos WHERE id = 3\nSELECT * FROM users WHERE id = 5", ExpectedOutput: "N+1 DETECTED: videos (3 queries after rooms)"},
						},
					},
					{
						Title: "SQL инъекция детектор", Difficulty: "easy",
						Description: `<p>Проверь SQL-запросы: безопасные используют $1/$2 (параметры), опасные — конкатенацию строк:</p>
<p>Ввод:</p><pre><code>3
SELECT * FROM users WHERE id = $1
SELECT * FROM users WHERE name = 'admin' OR '1'='1'
INSERT INTO videos (title) VALUES ($1)</code></pre>
<p>Вывод:</p><pre><code>SAFE: parameterized query
DANGER: possible SQL injection
SAFE: parameterized query</code></pre>`,
						Glossary: []GlossaryItem{{Term: "SQL injection", Definition: "Атака через конкатенацию: name = ''; DROP TABLE--. Защита: $1 параметры."}},
						TestCases: []TestCase{{Input: "3\nSELECT * FROM users WHERE id = $1\nSELECT * FROM users WHERE name = 'admin' OR '1'='1'\nINSERT INTO videos (title) VALUES ($1)", ExpectedOutput: "SAFE: parameterized query\nDANGER: possible SQL injection\nSAFE: parameterized query"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); q := sc.Text()
        if strings.Contains(q, "$") && !strings.Contains(q, "'") { fmt.Println("SAFE: parameterized query")
        } else { fmt.Println("DANGER: possible SQL injection") } } }`,
						Hints: `<p>$1/$2 → safe. Строковые литералы без параметров → danger.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);for i:=0;i<n;i++{sc.Scan();q:=sc.Text()
    if strings.Contains(q,"$")&&!strings.Contains(q,"'"){fmt.Println("SAFE: parameterized query")}else{fmt.Println("DANGER: possible SQL injection")}}}</code></pre>`,
					},
					{
						Title: "SQL SELECT парсер", Difficulty: "medium",
						Description: `<p>Разбери SELECT-запрос на компоненты:</p>
<p>Ввод: <code>SELECT id, title, year FROM videos WHERE year > 2000 ORDER BY year LIMIT 10</code></p>
<p>Вывод:</p><pre><code>Columns: id, title, year
Table: videos
Where: year > 2000
Order: year
Limit: 10</code></pre>`,
						Glossary: []GlossaryItem{{Term: "SELECT anatomy", Definition: "SELECT columns FROM table WHERE condition ORDER BY col LIMIT n."}},
						TestCases: []TestCase{{Input: "SELECT id, title, year FROM videos WHERE year > 2000 ORDER BY year LIMIT 10", ExpectedOutput: "Columns: id, title, year\nTable: videos\nWhere: year > 2000\nOrder: year\nLimit: 10"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { sc := bufio.NewScanner(os.Stdin); sc.Scan(); q := sc.Text()
    q = strings.TrimPrefix(q, "SELECT ")
    fromIdx := strings.Index(q, " FROM "); cols := q[:fromIdx]; q = q[fromIdx+6:]
    whereIdx := strings.Index(q, " WHERE "); orderIdx := strings.Index(q, " ORDER BY "); limitIdx := strings.Index(q, " LIMIT ")
    table := q; if whereIdx > 0 { table = q[:whereIdx] }
    where := ""; if whereIdx > 0 { end := len(q); if orderIdx > 0 { end = orderIdx }; where = q[whereIdx+7:end] }
    order := ""; if orderIdx > 0 { end := len(q); if limitIdx > 0 { end = limitIdx }; order = q[orderIdx+10:end] }
    limit := ""; if limitIdx > 0 { limit = q[limitIdx+7:] }
    fmt.Println("Columns:", cols); fmt.Println("Table:", table)
    if where != "" { fmt.Println("Where:", where) }
    if order != "" { fmt.Println("Order:", order) }
    if limit != "" { fmt.Println("Limit:", limit) } }`,
						Hints: `<p>Разбей по ключевым словам: FROM, WHERE, ORDER BY, LIMIT.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();q:=strings.TrimPrefix(sc.Text(),"SELECT ")
    fi:=strings.Index(q," FROM ");cols:=q[:fi];q=q[fi+6:]
    wi:=strings.Index(q," WHERE ");oi:=strings.Index(q," ORDER BY ");li:=strings.Index(q," LIMIT ")
    t:=q;if wi>0{t=q[:wi]};w:="";if wi>0{e:=len(q);if oi>0{e=oi};w=q[wi+7:e]}
    o:="";if oi>0{e:=len(q);if li>0{e=li};o=q[oi+10:e]};l:="";if li>0{l=q[li+7:]}
    fmt.Println("Columns:",cols);fmt.Println("Table:",t);if w!=""{fmt.Println("Where:",w)};if o!=""{fmt.Println("Order:",o)};if l!=""{fmt.Println("Limit:",l)}}</code></pre>`,
					},
					{
						Title: "Connection pool симулятор", Difficulty: "hard",
						Description: `<p>Симулируй connection pool: acquire/release. При нехватке → WAIT:</p>
<p>Ввод:</p><pre><code>2
5
acquire c1
acquire c2
acquire c3
release c1
acquire c4</code></pre>
<p>Вывод:</p><pre><code>c1: acquired (1/2 used)
c2: acquired (2/2 used)
c3: WAIT (pool exhausted)
c1: released (1/2 used)
c4: acquired (2/2 used)</code></pre>`,
						Glossary: []GlossaryItem{{Term: "Connection Pool", Definition: "Фиксированное количество соединений. acquire берёт свободное, release возвращает."}},
						TestCases: []TestCase{{Input: "2\n5\nacquire c1\nacquire c2\nacquire c3\nrelease c1\nacquire c4", ExpectedOutput: "c1: acquired (1/2 used)\nc2: acquired (2/2 used)\nc3: WAIT (pool exhausted)\nc1: released (1/2 used)\nc4: acquired (2/2 used)"}},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { var maxConns int; fmt.Scan(&maxConns); var n int; fmt.Scan(&n)
    used := 0; sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); p := strings.Fields(sc.Text()); cmd, name := p[0], p[1]
        switch cmd {
        case "acquire":
            if used < maxConns { used++; fmt.Printf("%s: acquired (%d/%d used)\n", name, used, maxConns)
            } else { fmt.Printf("%s: WAIT (pool exhausted)\n", name) }
        case "release":
            used--; fmt.Printf("%s: released (%d/%d used)\n", name, used, maxConns)
        }
    }
}`,
						Hints: `<p>Счётчик used. acquire → if used < max then used++. release → used--.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){var mx,n int;fmt.Scan(&mx,&n);u:=0;sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text())
        if p[0]=="acquire"{if u<mx{u++;fmt.Printf("%s: acquired (%d/%d used)\n",p[1],u,mx)}else{fmt.Printf("%s: WAIT (pool exhausted)\n",p[1])}}else{u--;fmt.Printf("%s: released (%d/%d used)\n",p[1],u,mx)}}}</code></pre>`,
					},
				},
			},
		},
	}

	m.Lessons = append(m.Lessons, pgxPracticeLesson())
	return m
}

func mod11_architecture() M {
	return M{
		Slug: "architecture", Title: "Чистая архитектура", Order: 11,
		Description: "Handler → Service → Repository, когда это нужно, когда оверинжиниринг, реальные примеры.",
		Track: "backend", Difficulty: "advanced", Prerequisites: []string{"database"},
		Lessons: []L{
			{
				Slug: "layers-deep", Title: "Трёхслойная архитектура на практике", Order: 1,
				Content: `<h1>Handler → Service → Repository</h1>

<h2>Зачем слои?</h2>
<pre><code>// БЕЗ слоёв (всё в handler):
func CreateRoom(w http.ResponseWriter, r *http.Request) {
    var input struct{ Name string; VideoID int64 }
    json.NewDecoder(r.Body).Decode(&input)                    // HTTP
    if input.Name == "" { http.Error(w, "empty", 400); return } // валидация
    var exists bool
    pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM videos WHERE id=$1)", input.VideoID).Scan(&exists) // SQL
    if !exists { http.Error(w, "not found", 404); return }
    pool.Exec(ctx, "INSERT INTO rooms ...", input.Name)       // SQL
    json.NewEncoder(w).Encode(map[string]string{"status":"ok"}) // HTTP
}
// Проблемы: нельзя тестировать без БД, нельзя переиспользовать логику,
// при смене БД нужно переписать handler</code></pre>

<h2>Правильная архитектура</h2>
<pre><code>// HANDLER — только HTTP
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
    var input service.CreateRoomInput
    if err := decodeJSON(r, &input); err != nil {
        respondError(w, 400, "invalid json")
        return
    }
    room, err := h.svc.CreateRoom(r.Context(), input)
    if err != nil { handleError(w, err); return }
    respondJSON(w, 201, room)
}

// SERVICE — бизнес-логика + валидация
func (s *RoomService) CreateRoom(ctx context.Context, input CreateRoomInput) (*Room, error) {
    if input.Name == "" {
        return nil, apperror.BadRequest("имя комнаты обязательно")
    }
    if _, err := s.videoStore.GetByID(ctx, input.VideoID); err != nil {
        return nil, apperror.NotFound("видео не найдено")
    }
    room := &Room{Name: input.Name, VideoID: input.VideoID}
    return room, s.roomStore.Create(ctx, room)
}

// REPOSITORY — только SQL
func (r *RoomRepo) Create(ctx context.Context, room *Room) error {
    return r.pool.QueryRow(ctx,
        "INSERT INTO rooms (name, video_id) VALUES ($1, $2) RETURNING id, created_at",
        room.Name, room.VideoID).Scan(&room.ID, &room.CreatedAt)
}</code></pre>

<h2>Когда это оверинжиниринг?</h2>
<ul>
<li><strong>Маленький проект (< 5 endpoints):</strong> handler + repository достаточно. Service не нужен.</li>
<li><strong>CRUD без логики:</strong> если service просто прокидывает вызов в repo — убери service.</li>
<li><strong>Скрипт/CLI:</strong> не нужна архитектура вообще.</li>
</ul>
<p><strong>Правило:</strong> добавляй слой когда в текущем слое появляется код, который ему не принадлежит (SQL в handler, HTTP в service).</p>

<h2>Dependency Injection — сборка в main.go</h2>
<pre><code>func main() {
    pool := connectDB()

    // Repository (конкретные реализации)
    videoRepo := repository.NewVideoRepo(pool)
    roomRepo := repository.NewRoomRepo(pool)

    // Service (принимают интерфейсы)
    videoSvc := service.NewVideoService(videoRepo)
    roomSvc := service.NewRoomService(roomRepo, videoRepo)

    // Handler (принимают сервисы)
    videoHandler := handler.NewVideoHandler(videoSvc)
    roomHandler := handler.NewRoomHandler(roomSvc)

    // main.go — единственное место, которое знает все конкретные типы
}</code></pre>

<h2>Под капотом: почему интерфейсы в Go — идеальны для DI</h2>
<p>В Java/C# интерфейс нужно явно указать (<code>implements</code>). В Go — <strong>неявная реализация</strong>:</p>
<pre><code>// Service определяет интерфейс (что ему нужно)
type VideoStore interface {
    GetAll(ctx context.Context) ([]Video, error)
    GetByID(ctx context.Context, id int64) (*Video, error)
}

// Repository НЕ ЗНАЕТ про этот интерфейс!
// Но автоматически его реализует (есть нужные методы).
type VideoRepo struct{ pool *pgxpool.Pool }
func (r *VideoRepo) GetAll(ctx context.Context) ([]Video, error) { ... }
func (r *VideoRepo) GetByID(ctx context.Context, id int64) (*Video, error) { ... }

// Для тестов — мок (тоже реализует, без реальной БД):
type MockVideoStore struct {
    videos []Video
    err    error
}
func (m *MockVideoStore) GetAll(ctx context.Context) ([]Video, error) {
    return m.videos, m.err
}</code></pre>

<h2>Правила именования интерфейсов в Go</h2>
<pre><code>// ✅ Интерфейс определяется ТАМ ГДЕ ИСПОЛЬЗУЕТСЯ, не где реализуется
// Service определяет "мне нужен Store с методами X, Y"
// Repository просто реализует методы, не зная о Service

// ✅ Маленькие интерфейсы (1-3 метода)
type Reader interface { Read(p []byte) (n int, err error) }
type VideoStore interface { GetByID(ctx context.Context, id int64) (*Video, error) }

// ❌ Жирные интерфейсы (God interface)
type VideoManager interface {
    GetAll(); GetByID(); Create(); Update(); Delete();
    Search(); Import(); Export(); Validate(); // 10 методов!
}
// Разбей на мелкие: VideoReader, VideoWriter, VideoSearcher</code></pre>

<h2>Частые ошибки архитектуры</h2>
<pre><code>// ❌ Бизнес-логика в handler
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    if price < 0 { ... }         // ← это бизнес-логика, вынеси в service
    if exists { ... }            // ← и это тоже
}

// ❌ HTTP в service
func (s *Service) CreateRoom(...) {
    http.Error(w, "err", 400)   // service НЕ ЗНАЕТ про HTTP!
    // return apperror.BadRequest("...")  ← правильно
}

// ❌ Интерфейс рядом с реализацией
// repository/video_repo.go:
type VideoStore interface { ... }  // ← НЕПРАВИЛЬНО! Определяй где используешь
type VideoRepo struct{ ... }

// ✅ Интерфейс в service (потребитель):
// service/video.go:
type VideoStore interface { ... }  // ← service говорит "мне нужно вот это"</code></pre>`,

				Quiz: []Q{
					{Question: "Где должна быть SQL-логика?", Options: []string{"В handler", "В service", "В repository — только он знает о базе данных", "В main.go"}, Correct: 2, Explanation: "Repository отвечает за работу с хранилищем. Handler не знает SQL, Service не знает SQL. Если нужно сменить PostgreSQL на MongoDB — меняется только Repository."},
					{Question: "Когда service-слой — оверинжиниринг?", Options: []string{"Никогда", "Когда service просто прокидывает вызов в repository без дополнительной логики", "Всегда", "В маленьких проектах"}, Correct: 1, Explanation: "Если service.GetAll() = return repo.GetAll() — он не нужен. Добавь service когда появится валидация, бизнес-правила, вызовы нескольких repo."},
					{Question: "Почему main.go — 'composition root'?", Options: []string{"Так принято", "Это единственное место, знающее все конкретные типы — собирает зависимости", "Компилятор требует", "Для тестирования"}, Correct: 1, Explanation: "main.go создаёт конкретные repo, передаёт их как интерфейсы в service, service в handler. Все остальные пакеты работают через интерфейсы."},
				},
				Tasks: []T{
					{
						Title: "DI симулятор: подмена реализации", Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "interface + конструктор", Definition: "DI: функция принимает интерфейс → подставь любую реализацию."},
						},
						Description: `<p>Интерфейс <code>Store</code> с <code>GetAll() []string</code>. Реализация <code>MemoryStore</code> и <code>PrefixStore</code> (добавляет префикс). Service использует Store.</p>
<p>Ввод: <code>memory</code> → <code>Matrix\nInception</code>. <code>prefix</code> → <code>[v2] Matrix\n[v2] Inception</code></p>`,
						StarterCode: `package main

import "fmt"

type Store interface { GetAll() []string }
type MemoryStore struct{}
func (m *MemoryStore) GetAll() []string { return []string{"Matrix", "Inception"} }

type PrefixStore struct{ prefix string; inner Store }
func (p *PrefixStore) GetAll() []string {
	// TODO: верни inner.GetAll() с prefix перед каждым
	return nil
}

type Service struct{ store Store }
func (s *Service) List() { for _, v := range s.store.GetAll() { fmt.Println(v) } }

func main() {
	var typ string
	fmt.Scan(&typ)
	mem := &MemoryStore{}
	// TODO: if typ == "prefix" → PrefixStore{prefix: "[v2] ", inner: mem}
	svc := &Service{store: mem}
	svc.List()
}`,
						TestCases: []TestCase{
							{Input: "memory", ExpectedOutput: "Matrix\nInception"},
							{Input: "prefix", ExpectedOutput: "[v2] Matrix\n[v2] Inception"},
						},
						Hints: `<p>PrefixStore.GetAll: цикл по inner.GetAll(), каждому добавь prefix.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type Store interface { GetAll() []string }
type MemoryStore struct{}
func (m *MemoryStore) GetAll() []string { return []string{"Matrix", "Inception"} }

type PrefixStore struct{ prefix string; inner Store }
func (p *PrefixStore) GetAll() []string {
	items := p.inner.GetAll()
	result := make([]string, len(items))
	for i, v := range items { result[i] = p.prefix + v }
	return result
}

type Service struct{ store Store }
func (s *Service) List() { for _, v := range s.store.GetAll() { fmt.Println(v) } }

func main() {
	var typ string
	fmt.Scan(&typ)
	mem := &MemoryStore{}
	var store Store = mem
	if typ == "prefix" { store = &PrefixStore{prefix: "[v2] ", inner: mem} }
	svc := &Service{store: store}
	svc.List()
}</code></pre>`,
					},
				},
			},
		},
	}
}

func mod12_testing() M {
	return M{
		Slug: "testing", Title: "Тестирование", Order: 12,
		Description: "Unit тесты, table-driven, моки, testcontainers, benchmarks, coverage, race detector.",
		Track: "backend", Difficulty: "advanced", Prerequisites: []string{"architecture"},
		Lessons: []L{
			{
				Slug: "testing-deep", Title: "Тестирование: от unit до integration", Order: 1,
				Content: `<h1>Тестирование в Go — полный гайд</h1>

<h2>Под капотом: как работает go test</h2>
<ol>
<li>Go находит все файлы <code>*_test.go</code></li>
<li>Компилирует их в отдельный бинарник</li>
<li>Запускает все функции <code>Test*</code>, <code>Benchmark*</code>, <code>Example*</code></li>
<li>Каждый Test* получает свой <code>*testing.T</code></li>
</ol>

<h2>Table-driven тесты — стандарт Go</h2>
<pre><code>func TestFormatSize(t *testing.T) {
    tests := []struct {
        name string
        size int64
        want string
    }{
        {"ноль", 0, "0 Б"},
        {"байты", 999, "999 Б"},
        {"килобайты", 1536, "1.5 КБ"},
        {"мегабайты", 1572864, "1.5 МБ"},
        {"гигабайты", 1610612736, "1.5 ГБ"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := formatSize(tt.size)
            if got != tt.want {
                t.Errorf("formatSize(%d) = %q, want %q", tt.size, got, tt.want)
            }
        })
    }
}</code></pre>

<h2>Моки через интерфейсы</h2>
<pre><code>// Мок — структура с настраиваемым поведением
type mockVideoStore struct {
    videos []model.Video
    err    error
}
func (m *mockVideoStore) GetAll(ctx context.Context) ([]model.Video, error) {
    return m.videos, m.err
}

func TestListVideos_Success(t *testing.T) {
    mock := &mockVideoStore{
        videos: []model.Video{{ID: 1, Title: "Test"}},
    }
    svc := service.NewVideoService(mock)

    videos, err := svc.ListVideos(context.Background())
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(videos) != 1 {
        t.Errorf("got %d videos, want 1", len(videos))
    }
}

func TestListVideos_Error(t *testing.T) {
    mock := &mockVideoStore{err: errors.New("db down")}
    svc := service.NewVideoService(mock)

    _, err := svc.ListVideos(context.Background())
    if err == nil {
        t.Error("expected error, got nil")
    }
}</code></pre>

<h2>HTTP тесты</h2>
<pre><code>func TestHealthEndpoint(t *testing.T) {
    r := chi.NewRouter()
    r.Get("/api/health", healthHandler)

    req := httptest.NewRequest("GET", "/api/health", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != 200 {
        t.Errorf("status = %d, want 200", w.Code)
    }
    var body map[string]string
    json.Unmarshal(w.Body.Bytes(), &body)
    if body["status"] != "ok" {
        t.Errorf("status = %q, want 'ok'", body["status"])
    }
}</code></pre>

<h2>Полезные команды</h2>
<pre><code>go test ./...                 # все тесты
go test -v ./...              # подробно
go test -run TestFormat ./... # конкретный тест
go test -race ./...           # детектор гонок данных
go test -cover ./...          # покрытие
go test -count=1 ./...        # без кеша
go test -bench=. ./...        # бенчмарки
go test -coverprofile=c.out && go tool cover -html=c.out  # визуализация</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: тестировать приватные детали реализации
// Тестируй поведение (вход → выход), не внутреннее устройство

// ОШИБКА: тесты зависят от порядка выполнения
// Каждый тест должен быть независим

// ОШИБКА: не использовать t.Helper()
func assertEqual(t *testing.T, got, want string) {
    t.Helper() // ← ошибка будет указывать на вызывающий код, не на эту функцию
    if got != want { t.Errorf("got %q, want %q", got, want) }
}</code></pre>

<h2>Benchmarks — измерение производительности</h2>
<pre><code>// Файл: video_test.go
func BenchmarkFormatSize(b *testing.B) {
    for i := 0; i < b.N; i++ {
        formatSize(1572864)
    }
}

// Запуск:
// go test -bench=BenchmarkFormatSize -benchmem ./...
// BenchmarkFormatSize-8   12000000   95.2 ns/op   32 B/op   2 allocs/op
//                          ↑ итераций  ↑ время       ↑ память  ↑ аллокации

// Сравнение двух реализаций:
func BenchmarkConcat(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := ""
        for j := 0; j < 100; j++ { s += "x" }  // медленно!
    }
}
func BenchmarkBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var sb strings.Builder
        for j := 0; j < 100; j++ { sb.WriteString("x") }  // быстро!
        _ = sb.String()
    }
}
// Concat:  500000 ns/op, 10000 B/op
// Builder: 1000 ns/op,   256 B/op   ← в 500 раз быстрее!</code></pre>

<h2>TestMain — setup/teardown для всего пакета</h2>
<pre><code>func TestMain(m *testing.M) {
    // Setup: подготовка перед ВСЕМИ тестами пакета
    db := setupTestDB()

    // Запуск всех тестов
    code := m.Run()

    // Teardown: очистка
    db.Close()
    os.Exit(code)
}</code></pre>

<h2>Parallel тесты</h2>
<pre><code>func TestConcurrentAccess(t *testing.T) {
    t.Parallel()  // этот тест может выполняться параллельно с другими

    // t.Parallel() в подтестах:
    for _, tt := range tests {
        tt := tt  // копия! (ловушка замыкания)
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            got := process(tt.input)
            if got != tt.want { t.Errorf(...) }
        })
    }
}</code></pre>

<h2>Золотые правила тестирования</h2>
<ul>
<li><strong>Тестируй поведение, не реализацию</strong> — если рефакторинг ломает тесты, тесты плохие</li>
<li><strong>Один тест = одна проверка</strong> — не мешай логику</li>
<li><strong>Тест читается как документация</strong> — имена: TestCreateRoom_EmptyName_ReturnsError</li>
<li><strong>AAA паттерн:</strong> Arrange (подготовка) → Act (действие) → Assert (проверка)</li>
<li><strong>Покрытие 80% достаточно</strong> — 100% = тесты тестов = бесполезно</li>
</ul>`,

				Quiz: []Q{
					{Question: "Что делает go test -race?", Options: []string{"Ускоряет тесты", "Включает детектор гонок данных — находит конкурентные баги", "Запускает параллельно", "Тестирует производительность"}, Correct: 1, Explanation: "-race включает Go race detector, который находит одновременный доступ к данным без синхронизации. Критично для серверного кода."},
					{Question: "Зачем t.Helper()?", Options: []string{"Ускоряет тесты", "Ошибка указывает на вызывающий код, а не на helper-функцию", "Пропускает тест", "Параллелит тесты"}, Correct: 1, Explanation: "Без t.Helper() ошибка покажет строку внутри assertEqual. С t.Helper() — строку где assertEqual вызвана. Удобнее для отладки."},
					{Question: "Что тестировать моками, а что — реальной БД?", Options: []string{"Всё моками", "Service — моками (unit test), Repository — реальной БД (integration test)", "Всё реальной БД", "Ничего не тестировать"}, Correct: 1, Explanation: "Service тестируется с моками repo (быстро, изолированно). Repository тестируется с реальным PostgreSQL через testcontainers (проверяет SQL)."},
				},
				Tasks: []T{
					{
						Title: "Table-driven: FormatDuration", Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "table-driven", Definition: "Массив тест-кейсов {name, input, want}. Цикл проверяет каждый. Стандарт Go."},
						},
						Description: `<p>Напиши <code>FormatDuration(minutes int) string</code>: 0→"0m", 45→"45m", 60→"1h 0m", 90→"1h 30m".</p>
<p>На вход — числа. На выход — отформатированное время.</p>`,
						StarterCode: `package main

import "fmt"

func FormatDuration(minutes int) string {
	// TODO: < 60 → "Nm", >= 60 → "Xh Ym"
	return ""
}

func main() {
	var n int
	for {
		_, err := fmt.Scan(&n)
		if err != nil { break }
		fmt.Println(FormatDuration(n))
	}
}`,
						TestCases: []TestCase{
							{Input: "0\n45\n60\n90\n150", ExpectedOutput: "0m\n45m\n1h 0m\n1h 30m\n2h 30m"},
						},
						Hints: `<p>if minutes < 60 → Sprintf("%dm") else Sprintf("%dh %dm", m/60, m%60)</p>`,
						Solution: `<pre><code>package main
import "fmt"
func FormatDuration(m int) string {
	if m < 60 { return fmt.Sprintf("%dm", m) }
	return fmt.Sprintf("%dh %dm", m/60, m%60)
}
func main() {
	var n int
	for { _, err := fmt.Scan(&n); if err != nil { break }; fmt.Println(FormatDuration(n)) }
}</code></pre>`,
					},
					{Title: "Тесты для WatchTogether", Difficulty: "hard", Glossary: []GlossaryItem{
					{Term: "func TestXxx(t *testing.T)", Definition: "Функция теста. Имя начинается с Test, принимает *testing.T для проверок."},
					{Term: "t.Run(name, func(t *testing.T))", Definition: "Подтест. Позволяет группировать тест-кейсы в table-driven стиле."},
					{Term: "httptest.NewRecorder()", Definition: "Мок http.ResponseWriter. Записывает ответ сервера для проверки: Code, Body."},
				}, Description: `<p>Напиши тесты: 1) Table-driven для formatSize/formatDuration. 2) Мок VideoStore + тесты VideoService. 3) HTTP тест для /api/health через httptest.</p>`, Hints: `<p>httptest.NewRequest + httptest.NewRecorder + router.ServeHTTP(w, req). Проверяй w.Code и w.Body.</p>`, Solution: `<p>Паттерн из урока. Мок с полями videos/err. t.Run для подтестов. httptest для HTTP.</p>`}},
			},
		},
	}
}

func mod13_auth() M {
	return M{
		Slug: "auth", Title: "Аутентификация и безопасность", Order: 13,
		Description: "JWT под капотом, bcrypt, auth middleware, CORS, OWASP, security headers.",
		Track: "backend", Difficulty: "advanced", Prerequisites: []string{"testing"},
		Lessons: []L{
			{
				Slug: "jwt-deep", Title: "JWT, bcrypt и auth middleware", Order: 1,
				Content: `<h1>Аутентификация — полный разбор</h1>

<h2>Под капотом: как работает JWT</h2>
<p>JWT = три части, разделённые точками:</p>
<pre><code>eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjo0Mn0.signature
     Header              Payload              Signature

Header:  {"alg": "HS256", "typ": "JWT"}  → Base64
Payload: {"user_id": 42, "exp": 1735689600}  → Base64
Signature: HMAC-SHA256(header + "." + payload, secret)</code></pre>

<p><strong>JWT НЕ шифрует данные!</strong> Payload можно прочитать (просто Base64). JWT гарантирует только <strong>целостность</strong> — никто не может подменить payload без знания secret.</p>

<h2>Реализация в Go</h2>
<pre><code>import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID int64 ` + "`" + `json:"user_id"` + "`" + `
    jwt.RegisteredClaims
}

// Генерация токена
func GenerateToken(userID int64, secret string, ttl time.Duration) (string, error) {
    claims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// Валидация токена
func ParseToken(tokenStr, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
        func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
            }
            return []byte(secret), nil
        })
    if err != nil { return nil, err }
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid { return nil, errors.New("невалидный токен") }
    return claims, nil
}</code></pre>

<h2>Auth Middleware</h2>
<pre><code>func AuthMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            if !strings.HasPrefix(header, "Bearer ") {
                respondError(w, 401, "требуется авторизация")
                return
            }
            tokenStr := strings.TrimPrefix(header, "Bearer ")
            claims, err := ParseToken(tokenStr, secret)
            if err != nil {
                respondError(w, 401, "невалидный токен")
                return
            }
            ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}</code></pre>

<h2>Хеширование паролей (bcrypt)</h2>
<pre><code>import "golang.org/x/crypto/bcrypt"

// Хеширование (при регистрации)
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// Cost = 10 по умолчанию → ~100ms на хеш. Это СПЕЦИАЛЬНО — brute force дорогой

// Проверка (при логине)
err := bcrypt.CompareHashAndPassword(hashFromDB, []byte(inputPassword))
if err != nil {
    // НЕВЕРНЫЙ пароль (или хеш повреждён)
}

// НИКОГДА:
// - MD5, SHA-256 для паролей (слишком быстрые → brute force)
// - Хранить пароли в plain text
// - Логировать пароли</code></pre>

<h2>Безопасность: OWASP Top 10</h2>
<ul>
<li><strong>SQL Injection:</strong> используй $1 параметры (pgx делает автоматически)</li>
<li><strong>XSS:</strong> html/template экранирует по умолчанию</li>
<li><strong>CSRF:</strong> для API с JWT — не актуально (нет cookies)</li>
<li><strong>Broken Auth:</strong> JWT с коротким TTL, bcrypt для паролей</li>
<li><strong>Security Headers:</strong> X-Content-Type-Options, X-Frame-Options, HSTS</li>
</ul>

<h2>Cookie vs Header — где хранить токен</h2>
<pre><code>// Вариант 1: Authorization header (SPA, мобильные)
// + Нет CSRF атак
// - Нужен JavaScript (localStorage — XSS риск)
fetch("/api/videos", {
    headers: { "Authorization": "Bearer " + token }
})

// Вариант 2: HttpOnly cookie (SSR, классические сайты)
// + Недоступен из JavaScript → защита от XSS
// - Подвержен CSRF (нужен CSRF-токен)
http.SetCookie(w, &http.Cookie{
    Name:     "token",
    Value:    token,
    HttpOnly: true,   // JavaScript не может прочитать!
    Secure:   true,   // только HTTPS
    SameSite: http.SameSiteStrictMode,  // защита от CSRF
    Path:     "/",
    MaxAge:   3600,
})</code></pre>

<h2>Refresh Tokens — паттерн</h2>
<pre><code>// Access Token: короткий TTL (15 минут)
// Refresh Token: длинный TTL (7 дней), хранится в HttpOnly cookie
//
// Поток:
// 1. Login → access_token (15m) + refresh_token (7d)
// 2. API запросы с access_token
// 3. access_token истёк → POST /api/refresh с refresh_token
// 4. Сервер проверяет refresh_token → выдаёт новый access_token
// 5. refresh_token истёк → повторный логин</code></pre>

<h2>CORS — Cross-Origin Resource Sharing</h2>
<pre><code>// Браузер блокирует запросы с другого домена
// Сервер должен явно разрешить:
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "https://watchtogether.com")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
        if r.Method == "OPTIONS" {
            w.WriteHeader(204)  // preflight — браузер спрашивает "можно?"
            return
        }
        next.ServeHTTP(w, r)
    })
}

// ❌ НИКОГДА: Access-Control-Allow-Origin: *  с credentials
// Атакующий сайт сможет делать запросы от имени пользователя!</code></pre>

<h2>Rate Limiting — защита от abuse</h2>
<pre><code>// Ограничение: не более 100 запросов в минуту с одного IP
// Реализация: Token Bucket или Sliding Window
// В продакшене: nginx rate_limit или golang.org/x/time/rate

import "golang.org/x/time/rate"

limiter := rate.NewLimiter(rate.Every(time.Second/10), 10) // 10 req/s, burst 10

func rateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !limiter.Allow() {
            http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}</code></pre>`,

				Quiz: []Q{
					{Question: "Шифрует ли JWT данные в payload?", Options: []string{"Да, полностью", "Нет — payload просто Base64, читается без ключа. JWT гарантирует целостность, не конфиденциальность", "Зависит от алгоритма", "Частично"}, Correct: 1, Explanation: "JWT payload — Base64. Кто угодно может декодировать и прочитать. Не клади туда пароли или секреты. JWT гарантирует только что payload не подменён."},
					{Question: "Почему bcrypt а не SHA-256 для паролей?", Options: []string{"bcrypt новее", "bcrypt специально медленный (~100ms) — делает brute force непрактичным", "SHA-256 сломан", "Нет разницы"}, Correct: 1, Explanation: "SHA-256 вычисляется за наносекунды → миллиарды попыток в секунду. bcrypt с cost=10 → ~100ms → 10 попыток в секунду. Brute force нереалистичен."},
					{Question: "Зачем проверять метод подписи в ParseToken?", Options: []string{"Для скорости", "Атакующий может подменить alg на 'none' и обойти подпись", "Не нужно", "Для логирования"}, Correct: 1, Explanation: "Атака 'alg none': подменить алгоритм на none в header → подпись не проверяется. Проверка метода предотвращает это."},
					{Question: "Где хранить JWT secret?", Options: []string{"В коде", "В .env файле (не в git) или в системе секретов", "В JWT payload", "В базе данных"}, Correct: 1, Explanation: "Secret НИКОГДА не должен быть в коде или git. Используй переменные окружения (.env, не коммитится) или vault (HashiCorp Vault, AWS Secrets Manager)."},
				},
				Tasks: []T{
					{
						Title: "JWT payload decoder", Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "base64.RawURLEncoding", Definition: "JWT использует base64url (без padding =). RawURLEncoding — без padding."},
							{Term: "JWT = header.payload.signature", Definition: "Три части через точку. Payload — JSON с данными (user_id, exp)."},
						},
						Description: `<p>JWT токен — три base64url-encoded части через точку. Декодируй payload (вторую часть) и выведи user_id и проверь expired.</p>
<p>Ввод: JWT строка и текущее время (unix timestamp).</p>
<p>Вывод: <code>user_id=N status=valid</code> или <code>user_id=N status=expired</code></p>`,
						StarterCode: `package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func main() {
	var token string
	var now int64
	fmt.Scan(&token, &now)

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Println("invalid token")
		return
	}

	// TODO: декодируй parts[1] через base64.RawURLEncoding
	// TODO: unmarshal JSON, извлеки user_id и exp
	// TODO: если exp < now → "expired", иначе "valid"
	_ = parts
}`,
						TestCases: []TestCase{
							{Input: "xxx.eyJ1c2VyX2lkIjo0MiwiZXhwIjoxNzAwMDAwMDAwfQ.zzz 1699999999", ExpectedOutput: "user_id=42 status=valid"},
							{Input: "xxx.eyJ1c2VyX2lkIjo3LCJleHAiOjE2MDAwMDAwMDB9.zzz 1700000000", ExpectedOutput: "user_id=7 status=expired"},
						},
						Hints: `<p>base64.RawURLEncoding.DecodeString(parts[1]). Unmarshal в map[string]any. exp и user_id — float64 в JSON.</p>`,
						Solution: `<pre><code>package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

func main() {
	var token string
	var now int64
	fmt.Scan(&token, &now)

	parts := strings.Split(token, ".")
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]any
	json.Unmarshal(payload, &claims)

	userID := int(claims["user_id"].(float64))
	exp := int64(claims["exp"].(float64))
	status := "valid"
	if exp < now { status = "expired" }
	fmt.Printf("user_id=%d status=%s\n", userID, status)
}</code></pre>`,
					},
					{Title: "Auth система для WatchTogether", Difficulty: "hard", Glossary: []GlossaryItem{
					{Term: "bcrypt.GenerateFromPassword(pw, cost)", Definition: "Хеширует пароль. cost — сложность (10-12 для прод). Результат содержит соль внутри."},
					{Term: "jwt.NewWithClaims(method, claims)", Definition: "Создаёт JWT токен. claims — данные (user_id, exp). method — алгоритм подписи (HS256)."},
					{Term: "r.Header.Get(\"Authorization\")", Definition: "Читает HTTP заголовок. Для JWT: 'Bearer <token>'. Отрежь 'Bearer ' чтобы получить токен."},
				}, Description: `<p>1) Таблица users (id, username, email, password_hash). 2) POST /auth/register (bcrypt). 3) POST /auth/login (JWT). 4) Auth middleware. 5) GET /api/me — данные текущего пользователя.</p>`, Hints: `<p>Register: bcrypt.GenerateFromPassword. Login: bcrypt.CompareHashAndPassword → GenerateToken. Middleware: Bearer token → ParseToken → context.</p>`, Solution: `<p>Полный flow: Register → hash password → save. Login → check password → generate JWT. Middleware → parse JWT → ctx с user_id.</p>`}},
			},
		},
	}
}

// mod14_concurrency moved to mod_concurrency.go as mod14_concurrency_full()
func mod14_concurrency_old_UNUSED() M {
	return M{
		Slug: "concurrency-old", Title: "UNUSED", Order: 99,
		Description: "REPLACED by mod_concurrency.go",
		Track: "backend", Difficulty: "expert", Prerequisites: []string{"auth"},
		Lessons: []L{
			{
				Slug: "goroutines-deep", Title: "Горутины: как это работает внутри", Order: 1,
				Content: `<h1>Конкурентность в Go — под капотом</h1>

<h2>Горутина ≠ поток ОС</h2>
<p>Это ключевое отличие Go от других языков:</p>
<ul>
<li><strong>Поток ОС:</strong> ~1МБ стека, создание ~10мкс, управляется ядром ОС</li>
<li><strong>Горутина Go:</strong> ~2КБ стека (растёт по необходимости), создание ~0.3мкс, управляется Go runtime</li>
</ul>

<p><strong>Модель M:N</strong> — Go runtime мультиплексирует M горутин на N потоков ОС:</p>
<pre><code>// Go runtime содержит свой планировщик:
// G (goroutine) — горутина
// M (machine) — поток ОС
// P (processor) — логический процессор (= GOMAXPROCS, обычно = кол-во CPU)
//
// Каждый P имеет очередь горутин. M берёт G из очереди P и выполняет.
// Если горутина блокируется (I/O, канал) — M переключается на другую G.</code></pre>

<h2>Каналы — подробно</h2>
<pre><code>// Небуферизированный — синхронный рандеву
ch := make(chan int)
// Отправка блокирует, пока кто-то не прочитает
// Чтение блокирует, пока кто-то не отправит

// Буферизированный — асинхронная очередь
ch := make(chan int, 100)
// Отправка блокирует только когда буфер полон
// Чтение блокирует только когда буфер пуст

// Направленные каналы (type safety)
func producer(out chan<- int) { out <- 42 }     // только отправка
func consumer(in <-chan int)  { val := <-in }   // только чтение</code></pre>

<h2>Select — мультиплексирование каналов</h2>
<pre><code>select {
case msg := <-msgCh:
    handleMessage(msg)
case err := <-errCh:
    handleError(err)
case <-ctx.Done():
    return ctx.Err()  // отмена или таймаут
case <-time.After(5 * time.Second):
    return errors.New("таймаут")
default:
    // выполняется если все каналы заблокированы (non-blocking)
}</code></pre>

<h2>sync пакет</h2>
<pre><code>// Mutex — защита общих данных
var mu sync.Mutex
mu.Lock()
sharedMap["key"] = "value"
mu.Unlock()

// RWMutex — много читателей, один писатель
var rw sync.RWMutex
rw.RLock()   // множество горутин могут читать одновременно
data := sharedMap["key"]
rw.RUnlock()

// WaitGroup — ожидание группы горутин
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        process(n)
    }(i)  // ← передаём i как аргумент!
}
wg.Wait()

// Once — выполнить ровно один раз (singleton)
var once sync.Once
var instance *Config
func GetConfig() *Config {
    once.Do(func() {
        instance = loadConfig()
    })
    return instance
}</code></pre>

<h2>Worker Pool — паттерн</h2>
<pre><code>func processVideos(ctx context.Context, paths []string) []Result {
    jobs := make(chan string, len(paths))
    results := make(chan Result, len(paths))

    // Запускаем N воркеров
    numWorkers := runtime.NumCPU()
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for path := range jobs {
                results <- processVideo(ctx, path)
            }
        }()
    }

    // Отправляем задания
    for _, p := range paths {
        jobs <- p
    }
    close(jobs) // сигнал воркерам: заданий больше нет

    // Ждём завершения и закрываем results
    go func() {
        wg.Wait()
        close(results)
    }()

    // Собираем результаты
    var out []Result
    for r := range results {
        out = append(out, r)
    }
    return out
}</code></pre>

<h2>Context — отмена и таймауты</h2>
<pre><code>// Таймаут
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel() // ВСЕГДА defer cancel!

// Проверка отмены в долгой операции
for _, file := range files {
    select {
    case <-ctx.Done():
        return ctx.Err() // "context deadline exceeded" или "context canceled"
    default:
    }
    process(file)
}</code></pre>

<h2>Race Conditions — частые ошибки</h2>
<pre><code>// ОШИБКА: конкурентная запись в map
m := map[string]int{}
go func() { m["a"] = 1 }() // PANIC: concurrent map writes
go func() { m["b"] = 2 }()

// ОШИБКА: замыкание в горутине
for i := 0; i < 5; i++ {
    go func() { fmt.Println(i) }() // все выведут 5!
}
// ПРАВИЛЬНО:
for i := 0; i < 5; i++ {
    go func(n int) { fmt.Println(n) }(i) // 0, 1, 2, 3, 4
}

// Найти гонки: go test -race ./...</code></pre>`,

				Quiz: []Q{
					{Question: "Чем горутина отличается от потока ОС?", Options: []string{"Ничем", "Горутина: ~2КБ стека, управляется Go runtime, тысячи дешёвых. Поток ОС: ~1МБ, управляется ядром", "Горутина медленнее", "Горутина — это поток"}, Correct: 1, Explanation: "Горутины мультиплексируются на потоки ОС (M:N модель). Стек горутины начинается с 2КБ и растёт по необходимости. Можно создать миллионы."},
					{Question: "Когда блокируется отправка в буферизированный канал?", Options: []string{"Всегда", "Когда буфер полон", "Никогда", "Когда нет получателя"}, Correct: 1, Explanation: "Буферизированный канал — очередь. Отправка блокирует только когда очередь полная. Чтение — когда пустая."},
					{Question: "Почему `go func() { fmt.Println(i) }()` в цикле — баг?", Options: []string{"Слишком много горутин", "Замыкание захватывает переменную i, все горутины увидят финальное значение", "Синтаксическая ошибка", "Горутины не могут печатать"}, Correct: 1, Explanation: "Замыкание захватывает переменную (не значение). К моменту выполнения горутин цикл уже завершён и i = финальное значение."},
					{Question: "Что делает sync.Once?", Options: []string{"Запускает горутину один раз", "Гарантирует что функция выполнится ровно один раз даже из множества горутин", "Блокирует выполнение", "Создаёт мьютекс"}, Correct: 1, Explanation: "sync.Once — потокобезопасный singleton. Первый вызов Do() выполняет функцию, все остальные — ждут и возвращают. Идеально для lazy init."},
					{Question: "Зачем defer cancel() после context.WithTimeout?", Options: []string{"Для красоты", "Без cancel() утекает внутренний таймер — goroutine leak", "Отменяет таймаут", "Не нужен"}, Correct: 1, Explanation: "WithTimeout создаёт горутину-таймер. cancel() освобождает её. Без cancel() горутина живёт до истечения таймаута — утечка ресурсов."},
				},
				Tasks: []T{
					{
						Title: "Fan-out/Fan-in: параллельная обработка", Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "ch := make(chan int, n)", Definition: "Буферизированный канал. До n значений без блокировки."},
							{Term: "sync.WaitGroup", Definition: "Счётчик горутин. Add(n), Done(), Wait()."},
						},
						Description: `<p>Обработай каждое число в отдельной горутине (×2), собери через канал, выведи сумму.</p>
<p>Ввод: <code>4\n1 2 3 4</code> → Вывод: <code>20</code></p>`,
						StarterCode: `package main

import (
	"fmt"
	"sync"
)

func main() {
	var n int
	fmt.Scan(&n)
	nums := make([]int, n)
	for i := range nums { fmt.Scan(&nums[i]) }

	results := make(chan int, n)
	var wg sync.WaitGroup

	for _, num := range nums {
		wg.Add(1)
		// TODO: запусти горутину: go func(v int) { defer wg.Done(); results <- v * 2 }(num)
		_ = num
	}

	go func() { wg.Wait(); close(results) }()

	sum := 0
	for v := range results { sum += v }
	fmt.Println(sum)
}`,
						TestCases: []TestCase{
							{Input: "4\n1 2 3 4", ExpectedOutput: "20"},
							{Input: "3\n10 20 30", ExpectedOutput: "120"},
						},
						Hints: `<p>go func(v int) { defer wg.Done(); results <- v * 2 }(num)</p>`,
						Solution: `<pre><code>package main
import ("fmt"; "sync")
func main() {
	var n int; fmt.Scan(&n)
	nums := make([]int, n)
	for i := range nums { fmt.Scan(&nums[i]) }
	results := make(chan int, n)
	var wg sync.WaitGroup
	for _, num := range nums { wg.Add(1); go func(v int) { defer wg.Done(); results <- v * 2 }(num) }
	go func() { wg.Wait(); close(results) }()
	sum := 0; for v := range results { sum += v }
	fmt.Println(sum)
}</code></pre>`,
					},
					{Title: "Worker pool для сканера видео", Difficulty: "hard", Glossary: []GlossaryItem{
						{Term: "go func() { ... }()", Definition: "Запуск горутины — легковесного потока. Дешёвый (~2КБ стека), тысячи одновременно."},
						{Term: "ch := make(chan T, size)", Definition: "Создание канала. size > 0 — буферизированный. Каналы — основной способ коммуникации горутин."},
						{Term: "sync.WaitGroup", Definition: "Счётчик горутин. Add(n) — добавить, Done() — уменьшить, Wait() — ждать пока не 0."},
						{Term: "context.WithCancel(ctx)", Definition: "Создаёт отменяемый контекст. cancel() сигнализирует всем горутинам через ctx.Done()."},
					}, Description: `<p>Обнови videoscanner: 1) Функция ScanDirParallel с worker pool (N = runtime.NumCPU()). 2) Каналы jobs/results. 3) Context для отмены. 4) WaitGroup для ожидания. 5) go test -race.</p>`, Hints: `<p>WalkDir отправляет пути в jobs. Воркеры читают jobs, stat файл, отправляют в results. После wg.Wait() — close(results).</p>`, Solution: `<p>Паттерн worker pool из урока. Ключевое: close(jobs) после отправки, wg.Wait() + close(results) в горутине, range results для сбора.</p>`},
				},
			},
		},
	}
}

func mod15_devops() M {
	return M{
		Slug: "devops", Title: "Docker и Linux", Order: 15,
		Description: "Docker под капотом, multi-stage, docker-compose, Linux, Nginx, systemd, сети.",
		Track: "devops", Difficulty: "intermediate", Prerequisites: []string{"packages"},
		Lessons: []L{
			{
				Slug: "docker-deep", Title: "Docker: от Dockerfile до продакшена", Order: 1,
				Content: `<h1>Docker — как это работает</h1>

<h2>Под капотом</h2>
<p>Docker использует фичи Linux-ядра:</p>
<ul>
<li><strong>Namespaces</strong> — изоляция процессов (PID, сеть, файловая система)</li>
<li><strong>Cgroups</strong> — лимиты ресурсов (CPU, RAM, I/O)</li>
<li><strong>Union FS</strong> — слоёная файловая система (каждая инструкция Dockerfile = слой)</li>
</ul>
<p>Контейнер — это <strong>не виртуальная машина</strong>. Это процесс с изоляцией. Ядро ОС общее.</p>

<h2>Multi-stage build</h2>
<pre><code># Stage 1: сборка (800МБ — компилятор, исходники, зависимости)
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download        # кешируется отдельно от кода
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: runtime (15МБ — только бинарник)
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' appuser
COPY --from=builder /server /server
USER appuser
EXPOSE 8080
CMD ["/server"]</code></pre>

<p><strong>-ldflags="-s -w"</strong> — убирает debug-символы, уменьшает бинарник на 30%.</p>
<p><strong>CGO_ENABLED=0</strong> — статическая линковка, не нужен libc в runtime образе.</p>

<h2>Layer caching — порядок имеет значение</h2>
<pre><code># ХОРОШО: зависимости кешируются отдельно
COPY go.mod go.sum ./      # меняется редко → кеш
RUN go mod download        # кешируется
COPY . .                   # меняется часто → новый слой
RUN go build               # пересобирается

# ПЛОХО: любое изменение кода инвалидирует ВСЁ
COPY . .
RUN go mod download && go build</code></pre>

<h2>Docker Compose для разработки</h2>
<pre><code>services:
  app:
    build: .
    ports: ["8080:8080"]
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/watchtogether?sslmode=disable
      JWT_SECRET: dev-secret-change-in-prod
    volumes:
      - ./videos:/data/videos  # монтируем видео
    depends_on:
      db: { condition: service_healthy }
      redis: { condition: service_healthy }

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: watchtogether
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    ports: ["5433:5432"]
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s

volumes:
  pgdata:</code></pre>

<h2>Частые ошибки</h2>
<pre><code># ОШИБКА: запуск от root
# Если контейнер скомпрометирован — злоумышленник получает root

# ОШИБКА: latest тег
FROM golang:latest  # какая версия? непредсказуемо!
FROM golang:1.22-alpine  # явная версия — воспроизводимость

# ОШИБКА: секреты в Dockerfile
ENV JWT_SECRET=mysecret  # видно в docker history!
# Используй docker secrets или переменные окружения при запуске</code></pre>`,

				Quiz: []Q{
					{Question: "Контейнер — это виртуальная машина?", Options: []string{"Да", "Нет — это изолированный процесс с общим ядром ОС", "Зависит от настроек", "Это микросервис"}, Correct: 1, Explanation: "Контейнер использует ядро хостовой ОС + namespaces для изоляции. VM имеет своё ядро. Контейнер легче, быстрее, но менее изолирован."},
					{Question: "Зачем CGO_ENABLED=0?", Options: []string{"Ускоряет сборку", "Создаёт статический бинарник без зависимости от libc — можно запускать в scratch/alpine", "Уменьшает размер", "Для безопасности"}, Correct: 1, Explanation: "С CGO бинарник зависит от libc. В alpine libc другая (musl). CGO_ENABLED=0 делает полностью статический бинарник."},
					{Question: "Почему go.mod копируется ПЕРЕД исходным кодом в Dockerfile?", Options: []string{"Так быстрее", "Docker кеширует слои — go mod download кешируется отдельно от изменений кода", "Обязательный порядок", "Для безопасности"}, Correct: 1, Explanation: "Docker инвалидирует кеш слоя если его COPY изменился. go.mod меняется редко → go mod download кешируется. Код меняется часто, но это уже следующий слой."},
					{Question: "Что делает depends_on с condition: service_healthy?", Options: []string{"Ничего", "Ждёт пока healthcheck сервиса не пройдёт успешно перед запуском зависимого", "Перезапускает сервис", "Проверяет порт"}, Correct: 1, Explanation: "Без condition app может стартовать раньше, чем БД готова принимать соединения. service_healthy ждёт успешного healthcheck."},
				},
				Tasks: []T{
				{
					Title: "Dockerfile layer optimizer", Difficulty: "medium",
					Glossary: []GlossaryItem{
						{Term: "COPY go.mod → go sum → download → COPY . .", Definition: "Порядок слоёв: зависимости кешируются, код — нет. Быстрый ребилд."},
					},
					Description: `<p>На вход — Dockerfile (по строкам). Проанализируй и выведи проблемы:</p>
<ul>
<li><code>COPY . .</code> перед <code>RUN go mod download</code> → <code>WARN: COPY before deps — cache invalidation</code></li>
<li>Нет <code>USER</code> → <code>WARN: no USER — running as root</code></li>
<li>Нет multi-stage (нет второго FROM) → <code>WARN: no multi-stage — large image</code></li>
<li>Если всё ок → <code>OK</code></li>
</ul>`,
					StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() { lines = append(lines, scanner.Text()) }

	// TODO: проверь порядок COPY/RUN, наличие USER, multi-stage
	hasUser := false
	fromCount := 0
	copyBeforeDeps := false
	copySeen := false

	for _, line := range lines {
		cmd := strings.Fields(line)
		if len(cmd) == 0 { continue }
		switch strings.ToUpper(cmd[0]) {
		case "FROM":
			fromCount++
		case "COPY":
			copySeen = true
		case "RUN":
			// TODO: если copySeen и содержит "go mod download" → copyBeforeDeps
		case "USER":
			hasUser = true
		}
	}

	warned := false
	if copyBeforeDeps { fmt.Println("WARN: COPY before deps — cache invalidation"); warned = true }
	if !hasUser { fmt.Println("WARN: no USER — running as root"); warned = true }
	if fromCount < 2 { fmt.Println("WARN: no multi-stage — large image"); warned = true }
	if !warned { fmt.Println("OK") }
}`,
					TestCases: []TestCase{
						{Input: "FROM golang:1.22\nCOPY . .\nRUN go mod download\nRUN go build -o app\nCMD [\"./app\"]", ExpectedOutput: "WARN: COPY before deps — cache invalidation\nWARN: no USER — running as root\nWARN: no multi-stage — large image"},
						{Input: "FROM golang:1.22 AS builder\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN go build -o app\nFROM alpine\nCOPY --from=builder /app /app\nUSER appuser\nCMD [\"/app\"]", ExpectedOutput: "OK"},
					},
					Hints: `<p>Отслеживай флаг copySeen. Если COPY встречается до RUN с "go mod download" → проблема. Считай FROM для multi-stage.</p>`,
					Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() { lines = append(lines, scanner.Text()) }

	hasUser := false
	fromCount := 0
	copyBeforeDeps := false
	copySeen := false

	for _, line := range lines {
		cmd := strings.Fields(line)
		if len(cmd) == 0 { continue }
		switch strings.ToUpper(cmd[0]) {
		case "FROM": fromCount++
		case "COPY":
			if !strings.Contains(line, "--from") { copySeen = true }
		case "RUN":
			if copySeen && strings.Contains(line, "go mod download") { copyBeforeDeps = true }
		case "USER": hasUser = true
		}
	}

	warned := false
	if copyBeforeDeps { fmt.Println("WARN: COPY before deps — cache invalidation"); warned = true }
	if !hasUser { fmt.Println("WARN: no USER — running as root"); warned = true }
	if fromCount < 2 { fmt.Println("WARN: no multi-stage — large image"); warned = true }
	if !warned { fmt.Println("OK") }
}</code></pre>`,
				},
				{Title: "Dockerize WatchTogether", Difficulty: "hard", Glossary: []GlossaryItem{
				{Term: "FROM golang:1.22 AS builder", Definition: "Multi-stage build: первый этап компилирует, второй — только бинарник. Итоговый образ маленький."},
				{Term: "CGO_ENABLED=0", Definition: "Отключает C-зависимости. Нужно для alpine/scratch — нет glibc."},
				{Term: "COPY --from=builder", Definition: "Копирует файл из предыдущего этапа (builder). Только бинарник, без исходников."},
				{Term: "depends_on + healthcheck", Definition: "docker-compose: depends_on ждёт запуска контейнера, healthcheck — ждёт готовности (порт отвечает)."},
			}, Description: `<p>1) Multi-stage Dockerfile. 2) docker-compose: app + PostgreSQL + Redis. 3) Healthchecks. 4) Volume для видео. 5) .dockerignore.</p>`, Hints: `<p>CGO_ENABLED=0. USER appuser. depends_on с condition. .dockerignore: .git, .idea, tmp/.</p>`, Solution: `<p>Шаблон из урока. Добавь .dockerignore, volumes для видео, явные версии образов.</p>`},
				{Title: "Docker debug — найди и исправь", Difficulty: "medium", Glossary: []GlossaryItem{
					{Term: "docker logs container", Definition: "Показать логи контейнера. Первое место для отладки."},
					{Term: "docker exec -it container sh", Definition: "Войти внутрь контейнера. Полезно для отладки файлов, сети, переменных."},
					{Term: "docker compose ps", Definition: "Показать статус всех контейнеров. Restarting = проблема."},
				}, Description: `<p>Практическая отладка Docker:</p>
<ol>
<li>Контейнер падает с CrashLoopBackOff — как найти причину? Напиши последовательность команд.</li>
<li>Приложение не подключается к БД — как проверить что контейнеры видят друг друга?</li>
<li>Образ слишком большой (1.2GB) — как уменьшить до &lt;50MB?</li>
</ol>
<p>Напиши ответы как пошаговые инструкции с командами.</p>`, Hints: `<p>1) docker logs, docker describe. 2) docker network ls, docker exec ping. 3) multi-stage + alpine/scratch.</p>`, Solution: `<p>1) docker logs app --tail 50. 2) docker exec app ping db + проверь DATABASE_URL. 3) Multi-stage: FROM golang AS builder → FROM alpine, COPY --from=builder.</p>`},
				},
			},
			{
				Slug: "linux-networking", Title: "Linux и сетевые основы", Order: 2,
				Content: `<h1>Linux и сети для DevOps</h1>

<h2>Сетевые основы</h2>
<pre><code># TCP/IP стек (снизу вверх):
# 1. Link Layer (Ethernet, Wi-Fi) — физическая доставка
# 2. Network Layer (IP) — маршрутизация между сетями
# 3. Transport Layer (TCP/UDP) — надёжная доставка / быстрая доставка
# 4. Application Layer (HTTP, DNS) — протоколы приложений

# TCP vs UDP:
# TCP — надёжный (гарантированная доставка, порядок). HTTP, PostgreSQL
# UDP — быстрый (без гарантий). DNS, видеостриминг, игры</code></pre>

<h2>DNS — как имя становится IP</h2>
<pre><code># Браузер: watchtogether.com → DNS resolver → 93.184.216.34
dig watchtogether.com        # DNS lookup
nslookup watchtogether.com   # альтернатива</code></pre>

<h2>Порты</h2>
<pre><code># 0-1023 — системные (нужен root)
# 80 — HTTP, 443 — HTTPS, 22 — SSH, 5432 — PostgreSQL
# 1024-65535 — пользовательские
# 8080 — типичный порт для Go приложений

ss -tlnp                    # какие порты слушают
lsof -i :8080               # что слушает порт 8080</code></pre>

<h2>Nginx как reverse proxy</h2>
<pre><code>server {
    listen 80;
    server_name watchtogether.local;

    # Обычные HTTP запросы → Go app
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # WebSocket → Go app (нужны особые заголовки)
    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}</code></pre>

<h2>Полезные команды Linux</h2>
<pre><code>ps aux | grep myapp        # найти процесс
kill -SIGTERM pid          # graceful stop (Go обрабатывает)
kill -SIGKILL pid          # мгновенное убийство (без cleanup)
df -h                      # место на диске
free -h                    # оперативная память
top / htop                 # нагрузка CPU/RAM
journalctl -u myservice -f # логи systemd сервиса
curl -v http://localhost   # HTTP запрос (подробно)
netstat -tlnp              # открытые порты</code></pre>

<h2>systemd — запуск сервиса</h2>
<pre><code># /etc/systemd/system/watchtogether.service
[Unit]
Description=WatchTogether
After=network.target postgresql.service

[Service]
Type=simple
User=appuser
WorkingDirectory=/opt/watchtogether
ExecStart=/opt/watchtogether/server
Restart=always
RestartSec=5
Environment=PORT=8080

[Install]
WantedBy=multi-user.target

# Команды:
systemctl enable watchtogether   # автозапуск
systemctl start watchtogether    # запустить
systemctl status watchtogether   # статус
systemctl restart watchtogether  # перезапуск</code></pre>`,

				Quiz: []Q{
					{Question: "Чем TCP отличается от UDP?", Options: []string{"TCP быстрее", "TCP гарантирует доставку и порядок, UDP — быстрый но без гарантий", "Ничем", "UDP надёжнее"}, Correct: 1, Explanation: "TCP устанавливает соединение, гарантирует доставку и порядок (HTTP, DB). UDP отправляет без подтверждений — быстрее для стриминга."},
					{Question: "Зачем Nginx перед Go приложением?", Options: []string{"Go не умеет HTTP", "Nginx: TLS терминация, статические файлы, rate limiting, балансировка нагрузки", "Для красивых URL", "Обязательно"}, Correct: 1, Explanation: "Nginx эффективнее для TLS, статики и балансировки. Go app получает чистые HTTP запросы и занимается бизнес-логикой."},
					{Question: "Что делает SIGTERM vs SIGKILL?", Options: []string{"Одно и то же", "SIGTERM просит процесс завершиться (graceful), SIGKILL убивает мгновенно (без cleanup)", "SIGKILL мягче", "SIGTERM для ошибок"}, Correct: 1, Explanation: "SIGTERM (15) — процесс может обработать (закрыть соединения, дописать логи). SIGKILL (9) — мгновенное убийство ядром. Всегда сначала SIGTERM."},
				},
				Tasks: []T{},
			},
		},
	}
}

func mod16_cicd() M {
	return M{
		Slug: "cicd", Title: "CI/CD", Order: 16,
		Description: "GitHub Actions, pipeline stages, secrets, environments, автодеплой.",
		Track: "devops", Difficulty: "advanced", Prerequisites: []string{"devops"},
		Lessons: []L{{
			Slug: "github-actions-deep", Title: "GitHub Actions: полный pipeline", Order: 1,
			Content: `<h1>CI/CD с GitHub Actions</h1>

<h2>Что такое CI/CD?</h2>
<ul>
<li><strong>CI (Continuous Integration):</strong> каждый push автоматически: lint → test → build</li>
<li><strong>CD (Continuous Deployment):</strong> после успешного CI — автоматический деплой</li>
</ul>
<p><strong>Зачем?</strong> Без CI — баги находятся поздно. С CI — каждый push проверяется за минуты.</p>

<h2>Полный pipeline</h2>
<pre><code>name: CI/CD
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }
      - uses: golangci/golangci-lint-action@v4

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports: ["5432:5432"]
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }
      - run: go test -race -coverprofile=coverage.out ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable
      - run: go tool cover -func=coverage.out

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: "1.22" }
      - run: CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server
      - uses: actions/upload-artifact@v4
        with: { name: server-binary, path: server }

  deploy:
    needs: build
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
      - run: echo "Deploy to production..."
      # scp binary to server, restart service, etc.</code></pre>

<h2>Secrets — НИКОГДА в коде</h2>
<pre><code># GitHub → Settings → Secrets → Actions
# Добавить: DATABASE_URL, JWT_SECRET, DEPLOY_KEY

# Использование в workflow:
env:
  DATABASE_URL: $` + "{{ secrets.DATABASE_URL }}" + `
  JWT_SECRET: $` + "{{ secrets.JWT_SECRET }}" + `</code></pre>

<h2>Makefile для локальной разработки</h2>
<pre><code># Makefile
.PHONY: run test lint build

run:
	go run cmd/server/main.go

test:
	go test -race -cover ./...

lint:
	golangci-lint run

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server cmd/server/main.go

docker-up:
	docker compose up -d

docker-down:
	docker compose down

seed:
	go run cmd/seed/main.go

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1</code></pre>

<h2>Кеширование зависимостей — ускорение CI</h2>
<pre><code>  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: "1.22"
        cache: true  # ← кеширует $GOPATH/pkg/mod + build cache
    # Без cache: go mod download каждый раз (~30с)
    # С cache: мгновенно (если go.sum не изменился)</code></pre>

<h2>Matrix strategy — тестирование на нескольких версиях</h2>
<pre><code>  test:
    strategy:
      matrix:
        go-version: ["1.21", "1.22", "1.23"]
        os: [ubuntu-latest, macos-latest]
    runs-on: $` + "{{ matrix.os }}" + `
    steps:
      - uses: actions/setup-go@v5
        with:
          go-version: $` + "{{ matrix.go-version }}" + `
      - run: go test ./...
    # Запустит 6 job: 3 версии Go × 2 ОС</code></pre>

<h2>Environment Protection — деплой с подтверждением</h2>
<pre><code>  deploy:
    environment: production  # ← требует одобрения в GitHub
    runs-on: ubuntu-latest
    steps:
      - run: ssh deploy@server 'cd /app && git pull && make restart'
    # GitHub Settings → Environments → production → Required reviewers
    # Деплой не начнётся пока reviewer не нажмёт Approve</code></pre>

<h2>Частые ошибки CI/CD</h2>
<pre><code># ❌ Не кешировать зависимости → CI занимает 5 минут вместо 1
# ❌ Деплой без тестов → баги в проде
# ❌ Секреты в .env в git → утечка через историю
# ❌ Нет rollback стратегии → невозможно откатить
# ✅ Blue-green deploy: новая версия рядом со старой → переключение → откат если баг</code></pre>`,

			Quiz: []Q{
				{Question: "Что значит needs: [lint, test] для job build?", Options: []string{"Параллельный запуск", "Build запустится только после успешного завершения lint и test", "lint и test не обязательны", "Build запускается первым"}, Correct: 1, Explanation: "needs определяет зависимости. Если lint или test упадёт — build и deploy не запустятся. Это гарантирует что в прод попадает только проверенный код."},
				{Question: "Где хранить секреты для CI?", Options: []string{"В .env в git", "В workflow YAML", "В GitHub Settings → Secrets → Actions", "В Dockerfile"}, Correct: 2, Explanation: "GitHub Secrets зашифрованы, доступны только в workflow runs. Никогда не коммить секреты в git."},
				{Question: "Зачем Makefile если есть go build?", Options: []string{"Обязателен", "Единая точка входа для всех команд: run, test, lint, build, docker, migrate", "Быстрее чем go", "Для совместимости"}, Correct: 1, Explanation: "Makefile — шпаргалка для команд проекта. make test вместо go test -race -cover ./... Новый разработчик сразу видит все доступные действия."},
			},
			Tasks: []T{{Title: "CI pipeline для WatchTogether", Difficulty: "hard", Glossary: []GlossaryItem{
				{Term: "on: push/pull_request", Definition: "Триггеры GitHub Actions. push — при пуше в ветку, pull_request — при создании PR."},
				{Term: "services: postgres", Definition: "Сервис-контейнер в CI. Поднимает PostgreSQL для интеграционных тестов."},
				{Term: "needs: [lint, test]", Definition: "Зависимости между jobs. build запустится только после успешного lint и test."},
			}, Description: `<p>1) .github/workflows/ci.yml: lint → test → build. 2) PostgreSQL service для тестов. 3) Makefile с командами. 4) .golangci.yml с правилами линтера.</p>`, Hints: `<p>services.postgres для тестов. needs для зависимостей jobs. actions/upload-artifact для бинарника.</p>`, Solution: `<p>Шаблон из урока. Ключевое: service_healthy для PostgreSQL, secrets для переменных, needs для порядка.</p>`}},
		}},
	}
}

func mod17_monitoring() M {
	return M{
		Slug: "monitoring", Title: "Мониторинг и логирование", Order: 17,
		Description: "slog structured logging, Prometheus metrics, Grafana, health checks, alerting.",
		Track: "devops", Difficulty: "advanced", Prerequisites: []string{"cicd"},
		Lessons: []L{
			{
				Slug: "slog-deep", Title: "Structured logging с slog", Order: 1,
				Content: `<h1>Логирование — правильный подход</h1>

<h2>Почему structured logging?</h2>
<pre><code>// ПЛОХО — текстовый лог
log.Printf("user %d created room %s for video %d", 42, "movie-night", 7)
// "2024/01/15 10:30:00 user 42 created room movie-night for video 7"
// Попробуй найти все логи пользователя 42... grep? regex? удачи.

// ХОРОШО — структурированный лог
slog.Info("room created",
    "user_id", 42,
    "room_name", "movie-night",
    "video_id", 7,
)
// {"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"room created","user_id":42,"room_name":"movie-night","video_id":7}
// Теперь: jq '.user_id == 42' → все логи пользователя</code></pre>

<h2>slog — стандартный пакет Go (1.21+)</h2>
<pre><code>import "log/slog"

// Для разработки (человекочитаемый)
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Для продакшена (JSON для агрегаторов)
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Уровни:
slog.Debug("детали для отладки")     // DEV only
slog.Info("нормальные события")      // запрос обработан, сервис стартовал
slog.Warn("подозрительно")           // retry удался, кеш промах
slog.Error("ошибка", "error", err)   // БД недоступна, внешний API упал</code></pre>

<h2>Request ID — трассировка запроса</h2>
<pre><code>// Middleware: добавить request_id в каждый запрос
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        w.Header().Set("X-Request-ID", requestID)

        // Все логи этого запроса будут иметь один request_id
        slog.InfoContext(ctx, "request started",
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
        )
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}</code></pre>

<h2>Что НЕ логировать</h2>
<ul>
<li>Пароли, токены, API ключи — НИКОГДА</li>
<li>Полные тела запросов — могут содержать PII (персональные данные)</li>
<li>Каждый SELECT из базы — слишком шумно</li>
<li>Успешные health checks — засоряют логи</li>
</ul>

<h2>slog.With — переиспользование полей</h2>
<pre><code>// Создать логгер с предустановленными полями
logger := slog.Default().With(
    "service", "watchtogether",
    "version", "1.2.0",
)

// Все логи автоматически содержат service и version:
logger.Info("server started", "port", 8080)
// {"service":"watchtogether","version":"1.2.0","msg":"server started","port":8080}

// В handler — добавить request-specific поля:
reqLogger := logger.With("request_id", reqID, "user_id", userID)
reqLogger.Info("processing request")
reqLogger.Error("failed", "error", err)</code></pre>

<h2>Три столпа Observability</h2>
<ol>
<li><strong>Logs (логи)</strong> — что произошло? slog, Loki, ELK</li>
<li><strong>Metrics (метрики)</strong> — как быстро? сколько? Prometheus, Grafana</li>
<li><strong>Traces (трассировки)</strong> — какой путь прошёл запрос? OpenTelemetry, Jaeger</li>
</ol>

<pre><code>// Метрики — числовые показатели за период:
// - request_count (Counter) — сколько запросов
// - request_duration_seconds (Histogram) — распределение latency
// - active_connections (Gauge) — текущее количество
//
// Prometheus собирает метрики через HTTP endpoint /metrics
// Grafana визуализирует графики и алерты</code></pre>

<h2>Health Check endpoint</h2>
<pre><code>// /api/health — для load balancer и мониторинга
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Проверяем все зависимости
    if err := db.Ping(r.Context()); err != nil {
        respondJSON(w, 503, map[string]string{
            "status": "unhealthy",
            "error":  "database unreachable",
        })
        return
    }
    respondJSON(w, 200, map[string]string{
        "status":  "healthy",
        "version": "1.2.0",
        "uptime":  time.Since(startTime).String(),
    })
}
// Kubernetes readinessProbe → /api/health
// Если 503 → трафик не направляется на этот pod</code></pre>

<h2>Что мониторить в Go-сервисе</h2>
<ul>
<li><strong>RED метрики:</strong> Rate (запросы/сек), Errors (% ошибок), Duration (latency)</li>
<li><strong>USE метрики:</strong> Utilization (CPU/RAM), Saturation (очереди), Errors</li>
<li><strong>Go runtime:</strong> goroutines count, heap size, GC pause time</li>
<li><strong>Бизнес-метрики:</strong> комнат создано, видео просмотрено</li>
</ul>`,

				Quiz: []Q{
					{Question: "Зачем JSON формат для логов в продакшене?", Options: []string{"Красивее", "Агрегаторы (Grafana Loki, ELK) могут парсить, фильтровать и визуализировать", "Быстрее", "Обязательно"}, Correct: 1, Explanation: "JSON логи можно загрузить в Loki/ELK/Datadog и искать: 'все ошибки user_id=42 за последний час'. С текстовыми логами это невозможно."},
					{Question: "Зачем Request ID?", Options: []string{"Для подсчёта", "Связать все логи одного HTTP запроса — от входа до ответа — для отладки", "Для авторизации", "Для кеширования"}, Correct: 1, Explanation: "Один запрос может породить десятки логов (middleware, service, repository). Request ID связывает их все, позволяя проследить путь запроса."},
					{Question: "Какой уровень для ошибки подключения к БД?", Options: []string{"Debug", "Info", "Warn", "Error"}, Correct: 3, Explanation: "Ошибка БД напрямую влияет на пользователей — slog.Error. Debug — для деталей отладки, Info — нормальные события, Warn — подозрительно но работает."},
				},
				Tasks: []T{
					{
						Title: "Structured log парсер", Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "JSON log", Definition: "Каждая строка лога — JSON объект с полями level, msg, request_id и др."},
						},
						Description: `<p>На вход — JSON-логи (по строке). Отфильтруй ошибки и выведи <code>request_id: message</code>.</p>
<p>Ввод:</p>
<pre><code>{"level":"INFO","msg":"started","request_id":"abc"}
{"level":"ERROR","msg":"db timeout","request_id":"abc"}
{"level":"INFO","msg":"ok","request_id":"def"}
{"level":"ERROR","msg":"not found","request_id":"def"}</code></pre>
<p>Вывод:</p>
<pre><code>abc: db timeout
def: not found</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var entry map[string]string
		json.Unmarshal([]byte(scanner.Text()), &entry)

		// TODO: если level == "ERROR" → выведи "request_id: msg"
		_ = entry
	}
}`,
						TestCases: []TestCase{
							{Input: "{\"level\":\"INFO\",\"msg\":\"started\",\"request_id\":\"abc\"}\n{\"level\":\"ERROR\",\"msg\":\"db timeout\",\"request_id\":\"abc\"}\n{\"level\":\"INFO\",\"msg\":\"ok\",\"request_id\":\"def\"}\n{\"level\":\"ERROR\",\"msg\":\"not found\",\"request_id\":\"def\"}", ExpectedOutput: "abc: db timeout\ndef: not found"},
						},
						Hints: `<p>json.Unmarshal в map[string]string. Проверяй entry["level"] == "ERROR".</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var entry map[string]string
		json.Unmarshal([]byte(scanner.Text()), &entry)
		if entry["level"] == "ERROR" {
			fmt.Printf("%s: %s\n", entry["request_id"], entry["msg"])
		}
	}
}</code></pre>`,
					},
				},
			},
		},
	}
}

func mod18_advanced() M {
	return M{
		Slug: "advanced", Title: "WebSocket и Redis", Order: 18,
		Description: "Realtime-синхронизация через WebSocket и кеширование на Redis.",
		Track: "backend", Difficulty: "expert", Prerequisites: []string{},
		Lessons: []L{
			{
				Slug: "websocket-deep", Title: "WebSocket для синхронного просмотра", Order: 1,
				Content: `<h1>WebSocket — realtime синхронизация</h1>

<h2>HTTP vs WebSocket</h2>
<pre><code>HTTP:     Client → Request → Server → Response → Connection закрыто
WebSocket: Client ↔ Server (постоянное двустороннее соединение)</code></pre>

<p>Для WatchTogether WebSocket идеален: когда один пользователь нажимает play — все остальные мгновенно получают событие.</p>

<h2>gorilla/websocket</h2>
<pre><code>var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true }, // в проде: проверять Origin!
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil { slog.Error("ws upgrade", "error", err); return }
    defer conn.Close()

    roomID := chi.URLParam(r, "roomID")
    h.hub.Register(roomID, conn)
    defer h.hub.Unregister(roomID, conn)

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil { break }
        h.hub.Broadcast(roomID, msg, conn) // всем кроме отправителя
    }
}</code></pre>

<h2>Hub паттерн — управление комнатами</h2>
<pre><code>type Hub struct {
    rooms map[string]map[*websocket.Conn]bool
    mu    sync.RWMutex
}

func (h *Hub) Register(roomID string, conn *websocket.Conn) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.rooms[roomID] == nil {
        h.rooms[roomID] = make(map[*websocket.Conn]bool)
    }
    h.rooms[roomID][conn] = true
}

func (h *Hub) Broadcast(roomID string, msg []byte, sender *websocket.Conn) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    for conn := range h.rooms[roomID] {
        if conn != sender {
            conn.WriteMessage(websocket.TextMessage, msg)
        }
    }
}

// Sync события
type SyncEvent struct {
    Type     string  ` + "`" + `json:"type"` + "`" + `      // "play", "pause", "seek"
    Position float64 ` + "`" + `json:"position"` + "`" + `  // секунды
    UserID   int64   ` + "`" + `json:"user_id"` + "`" + `
}</code></pre>`,

				Quiz: []Q{
					{Question: "Зачем WebSocket если есть HTTP?", Options: []string{"Моднее", "Постоянное соединение — сервер может push-ить события мгновенно без повторных запросов", "Быстрее", "Безопаснее"}, Correct: 1, Explanation: "HTTP: клиент спрашивает — сервер отвечает. WebSocket: оба могут слать сообщения в любой момент. Для realtime sync — единственный вариант."},
					{Question: "Зачем sync.RWMutex для Hub.rooms?", Options: []string{"Для красоты", "Broadcast (чтение map) — много горутин одновременно, Register (запись) — эксклюзивно", "Обязателен для map", "Для скорости"}, Correct: 1, Explanation: "RLock() позволяет множеству горутин читать map одновременно (broadcast). Lock() — эксклюзивная запись (register/unregister). Обычный Mutex блокировал бы всех."},
				},
				Tasks: []T{{Title: "Sync Room для WatchTogether", Difficulty: "hard", Glossary: []GlossaryItem{
					{Term: "websocket.Upgrader", Definition: "Апгрейдит HTTP-соединение до WebSocket. CheckOrigin — проверка CORS."},
					{Term: "conn.ReadMessage()", Definition: "Блокирующее чтение из WebSocket. Возвращает тип сообщения и данные."},
					{Term: "sync.RWMutex", Definition: "Read-Write мьютекс. RLock — множественное чтение, Lock — эксклюзивная запись."},
				}, Description: `<p>1) Hub с rooms map. 2) WebSocket handler /ws/room/{roomID}. 3) SyncEvent: play/pause/seek. 4) Broadcast всем кроме отправителя. 5) Graceful disconnect.</p>`, Hints: `<p>Hub: map[string]map[*Conn]bool + RWMutex. ReadPump в горутине. Broadcast с RLock.</p>`, Solution: `<p>Паттерн из урока. Register → defer Unregister. ReadMessage loop. Broadcast с исключением sender.</p>`}},
			},
			{
				Slug: "redis-deep", Title: "Redis: кеширование и сессии", Order: 2,
				Content: `<h1>Redis — быстрое хранилище</h1>

<h2>Зачем Redis?</h2>
<p>PostgreSQL запрос: 1-50мс. Redis GET: &lt;1мс. Кешируй то, что читается часто и меняется редко.</p>

<h2>go-redis</h2>
<pre><code>import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// SET с TTL
rdb.Set(ctx, "video:42", videoJSON, 5*time.Minute)

// GET
val, err := rdb.Get(ctx, "video:42").Result()
if err == redis.Nil {
    // cache miss — загрузи из БД
} else if err != nil {
    // реальная ошибка Redis
}

// DELETE (инвалидация кеша)
rdb.Del(ctx, "video:42")</code></pre>

<h2>Cache-Aside паттерн</h2>
<pre><code>func (s *VideoService) GetByID(ctx context.Context, id int64) (*Video, error) {
    key := fmt.Sprintf("video:%d", id)

    // 1. Попробовать кеш
    cached, err := s.cache.Get(ctx, key).Result()
    if err == nil {
        var v Video
        json.Unmarshal([]byte(cached), &v)
        return &v, nil // cache HIT
    }

    // 2. Cache miss → БД
    video, err := s.repo.GetByID(ctx, id)
    if err != nil { return nil, err }

    // 3. Положить в кеш
    data, _ := json.Marshal(video)
    s.cache.Set(ctx, key, data, 5*time.Minute)

    return video, nil
}

// При обновлении видео — инвалидировать кеш!
func (s *VideoService) Update(ctx context.Context, v *Video) error {
    if err := s.repo.Update(ctx, v); err != nil { return err }
    s.cache.Del(ctx, fmt.Sprintf("video:%d", v.ID)) // удалить из кеша
    return nil
}</code></pre>`,

				Quiz: []Q{
					{Question: "Что значит redis.Nil?", Options: []string{"Redis упал", "Ключ не найден — cache miss, нужно загрузить из БД", "Ошибка сети", "Пустое значение"}, Correct: 1, Explanation: "redis.Nil — не ошибка, а штатная ситуация: ключа нет. Загрузи данные из БД и положи в кеш."},
					{Question: "Когда инвалидировать кеш?", Options: []string{"Никогда — TTL достаточно", "При обновлении или удалении данных в БД", "Каждую секунду", "При каждом чтении"}, Correct: 1, Explanation: "При изменении данных — удали соответствующий ключ из кеша. Иначе кеш будет возвращать устаревшие данные."},
				},
				Tasks: []T{},
			},
			{
				Slug: "kubernetes-deep", Title: "Kubernetes: оркестрация", Order: 3,
				Content: `<h1>Kubernetes — управление контейнерами</h1>

<h2>Зачем Kubernetes?</h2>
<p>Docker запускает контейнеры. Kubernetes <strong>оркестрирует</strong> их: масштабирование, балансировка, self-healing, rolling updates, secrets.</p>

<h2>Ключевые концепции</h2>
<ul>
<li><strong>Pod</strong> — минимальная единица. Обычно 1 контейнер.</li>
<li><strong>Deployment</strong> — управляет N репликами Pod'а. Rolling update.</li>
<li><strong>Service</strong> — стабильный IP/DNS для группы Pod'ов.</li>
<li><strong>Ingress</strong> — HTTP маршрутизация извне кластера.</li>
<li><strong>ConfigMap</strong> — конфигурация. <strong>Secret</strong> — секреты.</li>
</ul>

<h2>Deployment манифест</h2>
<pre><code>apiVersion: apps/v1
kind: Deployment
metadata:
  name: watchtogether
spec:
  replicas: 3
  selector:
    matchLabels: { app: watchtogether }
  template:
    metadata:
      labels: { app: watchtogether }
    spec:
      containers:
        - name: app
          image: watchtogether:v1.0
          ports: [{ containerPort: 8080 }]
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef: { name: app-secrets, key: db-url }
          livenessProbe:   # жив ли контейнер?
            httpGet: { path: /api/health, port: 8080 }
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:  # готов ли принимать трафик?
            httpGet: { path: /api/health, port: 8080 }
          resources:
            limits: { memory: "256Mi", cpu: "500m" }
            requests: { memory: "128Mi", cpu: "250m" }</code></pre>

<h2>Основные команды kubectl</h2>
<pre><code>kubectl apply -f deployment.yaml    # применить
kubectl get pods                    # список подов
kubectl logs pod-name               # логи
kubectl describe pod pod-name       # детали
kubectl port-forward svc/app 8080   # проброс порта
kubectl scale deployment app --replicas=5  # масштабирование
kubectl rollout status deployment app      # статус деплоя
kubectl rollout undo deployment app        # откат</code></pre>

<h2>Практика: запуск с minikube</h2>
<p><strong>minikube</strong> — локальный кластер Kubernetes для разработки. Установка и запуск:</p>

<pre><code># Установка (macOS)
brew install minikube

# Запуск кластера
minikube start

# Проверка
kubectl cluster-info
# Kubernetes control plane is running at https://192.168.49.2:8443

# Используем локальный Docker для сборки образов
eval $(minikube docker-env)</code></pre>

<h2>Пошаговый деплой WatchTogether в minikube</h2>
<pre><code># 1. Собрать образ
docker build -t watchtogether:v1 .

# 2. Создать namespace
kubectl create namespace watchtogether

# 3. Применить манифесты
kubectl apply -f k8s/ -n watchtogether

# 4. Проверить поды
kubectl get pods -n watchtogether
# NAME                            READY   STATUS    RESTARTS   AGE
# watchtogether-6d4b8c9f5-abc12   1/1     Running   0          30s
# watchtogether-6d4b8c9f5-def34   1/1     Running   0          30s

# 5. Проброс порта для доступа
kubectl port-forward svc/watchtogether 8080:8080 -n watchtogether
# Открой http://localhost:8080

# 6. Масштабирование
kubectl scale deployment watchtogether --replicas=5 -n watchtogether

# 7. Обновление (rolling update)
docker build -t watchtogether:v2 .
kubectl set image deployment/watchtogether app=watchtogether:v2 -n watchtogether
kubectl rollout status deployment/watchtogether -n watchtogether

# 8. Откат при проблемах
kubectl rollout undo deployment/watchtogether -n watchtogether</code></pre>

<h2>Отладка проблем</h2>
<pre><code># Под не запускается?
kubectl describe pod POD_NAME -n watchtogether
# Смотри Events внизу — там причина

# CrashLoopBackOff?
kubectl logs POD_NAME -n watchtogether --previous
# Покажет логи упавшего контейнера

# ImagePullBackOff?
# Образ не найден. Убедись что eval $(minikube docker-env)

# Проверить ресурсы кластера
kubectl top pods -n watchtogether
kubectl top nodes</code></pre>

<h2>Helm — менеджер пакетов K8s</h2>
<pre><code># Helm — как apt/brew для Kubernetes
brew install helm

# Установить PostgreSQL через Helm
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install db bitnami/postgresql -n watchtogether \
  --set auth.postgresPassword=secret \
  --set auth.database=watchtogether

# Свой Helm chart
helm create watchtogether-chart
# Создаст шаблон с Deployment, Service, Ingress</code></pre>`,

				Quiz: []Q{
					{Question: "Чем livenessProbe отличается от readinessProbe?", Options: []string{"Ничем", "liveness: контейнер жив? (перезапустить если нет). readiness: готов к трафику? (убрать из балансировки)", "liveness быстрее", "readiness для логов"}, Correct: 1, Explanation: "livenessProbe: если fail → K8s перезапускает контейнер. readinessProbe: если fail → убирает из Service (не шлёт трафик), но не перезапускает."},
					{Question: "Что делает kubectl rollout undo?", Options: []string{"Удаляет deployment", "Откатывает к предыдущей версии (предыдущий image)", "Перезапускает поды", "Масштабирует до 0"}, Correct: 1, Explanation: "K8s хранит историю revisions. rollout undo переключает Deployment на предыдущий шаблон Pod (предыдущий image, конфиг). Zero-downtime откат."},
				},
				Tasks: []T{{Title: "K8s манифесты для WatchTogether", Difficulty: "hard", Glossary: []GlossaryItem{
					{Term: "Deployment", Definition: "K8s объект, управляющий ReplicaSet + Pods. Декларативное описание: сколько реплик, какой образ, probes."},
					{Term: "Service (ClusterIP)", Definition: "Виртуальный IP внутри кластера. Балансирует трафик между подами Deployment."},
					{Term: "Secret / ConfigMap", Definition: "Secret — для чувствительных данных (base64). ConfigMap — для конфигов (plain text)."},
					{Term: "livenessProbe / readinessProbe", Definition: "liveness: жив ли контейнер (перезапуск если нет). readiness: готов ли к трафику (убрать из балансировки)."},
				}, Description: `<p>1) Deployment (2 реплики, probes, limits). 2) Service (ClusterIP). 3) Secret для DATABASE_URL. 4) ConfigMap для PORT, LOG_LEVEL.</p>`, Hints: `<p>Используй --- для разделения ресурсов в одном файле. base64 для Secret data. kubectl apply -f.</p>`, Solution: `<p>Deployment + Service + ConfigMap + Secret. Ключевое: probes для /api/health, secretKeyRef для БД, resources limits.</p>`}},
			},
		},
	}
}
