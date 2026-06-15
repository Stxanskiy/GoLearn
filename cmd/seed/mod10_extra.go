package main

func pgxPracticeLesson() L {
	return L{
		Slug: "pgx-practice", Title: "pgx: pool, транзакции, миграции", Order: 2,
		Content: `<h1>pgx на практике</h1>
<h2>Connection Pool</h2>
<p>Pool переиспользует TCP-соединения к БД. Без pool каждый запрос = новое соединение = медленно.</p>
<pre><code>pool, _ := pgxpool.New(ctx, databaseURL)
// config.MaxConns = 20, config.MinConns = 5, config.MaxConnLifetime = time.Hour</code></pre>

<h2>CRUD операции</h2>
<pre><code>// INSERT с RETURNING id
var id int
pool.QueryRow(ctx,
    "INSERT INTO videos (title, year) VALUES ($1, $2) RETURNING id",
    "Matrix", 1999).Scan(&id)

// SELECT список (с пагинацией)
rows, err := pool.Query(ctx,
    "SELECT id, title FROM videos ORDER BY id LIMIT $1 OFFSET $2",
    limit, offset)
defer rows.Close()  // ОБЯЗАТЕЛЬНО!
for rows.Next() {
    var v Video
    rows.Scan(&v.ID, &v.Title)
    videos = append(videos, v)
}
if err := rows.Err(); err != nil { return err }  // проверяй после цикла!

// UPDATE
tag, _ := pool.Exec(ctx, "UPDATE videos SET title = $1 WHERE id = $2", title, id)
fmt.Println(tag.RowsAffected())

// DELETE
pool.Exec(ctx, "DELETE FROM videos WHERE id = $1", id)</code></pre>

<h2>Транзакции — всё или ничего</h2>
<pre><code>tx, err := pool.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx)  // страховка: no-op после Commit

// Перевод денег: атомарная операция
_, err = tx.Exec(ctx,
    "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
if err != nil { return err }

_, err = tx.Exec(ctx,
    "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
if err != nil { return err }

return tx.Commit(ctx)  // атомарно: или оба UPDATE, или ни одного</code></pre>

<h2>Миграции с golang-migrate</h2>
<pre><code># Создать файлы миграции
migrate create -ext sql -dir migrations -seq add_rooms

# Применить все
migrate -database "postgres://user:pass@localhost/db" -path migrations up

# Откатить последнюю
migrate -database "postgres://..." -path migrations down 1

# Текущая версия
migrate -database "postgres://..." -path migrations version</code></pre>

<h2>Защита от SQL-инъекций</h2>
<pre><code>// НИКОГДА: конкатенация
query := "SELECT * FROM users WHERE name = '" + name + "'"
// name = "'; DROP TABLE users; --"  → УДАЛИТ ТАБЛИЦУ!

// ВСЕГДА: параметры ($1, $2)
pool.Query(ctx, "SELECT * FROM users WHERE name = $1", name)
// Экранируется автоматически — инъекция невозможна</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Зачем defer rows.Close()?",
				Options:     []string{"Не нужен", "Без Close() соединение из pool не вернётся — утечка", "Ускоряет запрос", "Для красоты"},
				Correct:     1,
				Explanation: "rows удерживает соединение из pool. Без Close() оно навсегда занято. Pool исчерпается → приложение зависнет.",
			},
			{
				Question:    "Что произойдёт если транзакция не закоммичена?",
				Options:     []string{"Авто-коммит", "Соединение заблокировано, данные в неопределённом состоянии", "Ничего", "Ошибка"},
				Correct:     1,
				Explanation: "Незавершённая транзакция блокирует соединение и может блокировать строки в БД. defer tx.Rollback — страховка.",
			},
			{
				Question:    "Почему используют $1, $2 вместо ?, ? (как в MySQL)?",
				Options:     []string{"Нет разницы", "PostgreSQL использует нумерованные плейсхолдеры $N — можно ссылаться на один параметр несколько раз", "MySQL быстрее", "Go не поддерживает ?"},
				Correct:     1,
				Explanation: "$1, $2 — синтаксис PostgreSQL. Преимущество: можно использовать $1 дважды в одном запросе. MySQL ? — позиционные, нельзя переиспользовать.",
			},
			{
				Question:    "Зачем pool.MaxConns устанавливать разумное значение?",
				Options:     []string{"Для красоты", "Каждое соединение = TCP + память на стороне PostgreSQL. Слишком много → БД перегружена, слишком мало → очереди", "Go ограничение", "Для безопасности"},
				Correct:     1,
				Explanation: "PostgreSQL рекомендует max_connections = CPU * 2 + disk_count. Для Go-приложения обычно 20-50 коннектов в pool достаточно. Больше — контекстные переключения в БД.",
			},
			{
				Question:    "Что делает SELECT ... FOR UPDATE?",
				Options:     []string{"Обновляет данные", "Блокирует выбранные строки от изменения другими транзакциями до COMMIT/ROLLBACK", "Создаёт индекс", "Кеширует результат"},
				Correct:     1,
				Explanation: "FOR UPDATE — pessimistic lock. Строка заблокирована для записи до конца транзакции. Используется для атомарных операций: проверить баланс → списать.",
			},
		},
		Tasks: []T{
			{
				Title:       "Pool monitor", Difficulty: "easy",
				Description: `<p>Рассчитай утилизацию connection pool. На вход: maxConns, текущие acquired, idle:</p>
<p>Ввод: <code>20 15 5</code></p>
<p>Вывод:</p><pre><code>Max: 20
Acquired: 15
Idle: 5
Utilization: 75%</code></pre>`,
				Glossary: []GlossaryItem{{Term: "Pool utilization", Definition: "acquired / maxConns * 100. Высокая (>80%) → нужно увеличить pool."}},
				TestCases: []TestCase{{Input: "20 15 5", ExpectedOutput: "Max: 20\nAcquired: 15\nIdle: 5\nUtilization: 75%"}},
				StarterCode: `package main
import "fmt"
func main() { var max, acq, idle int; fmt.Scan(&max, &acq, &idle)
    fmt.Printf("Max: %d\nAcquired: %d\nIdle: %d\nUtilization: %d%%\n", max, acq, idle, acq*100/max) }`,
				Hints: `<p>Utilization = acquired * 100 / maxConns.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func main(){var m,a,i int;fmt.Scan(&m,&a,&i);fmt.Printf("Max: %d\nAcquired: %d\nIdle: %d\nUtilization: %d%%\n",m,a,i,a*100/m)}</code></pre>`,
			},
			{
				Title:       "Перевод денег (транзакция)",
				Difficulty:  "medium",
				Description: `<p>Симулируй транзакцию перевода денег. Операции: balance, transfer, rollback:</p>
<p>Ввод:</p><pre><code>5
balance Alice 100
balance Bob 50
transfer Alice Bob 30
transfer Bob Alice 200
balance Alice</code></pre>
<p>Вывод:</p><pre><code>Alice: 100
Bob: 50
Transfer: Alice -> Bob: 30 OK
Transfer: Bob -> Alice: 200 FAILED (insufficient funds)
Alice: 70</code></pre>`,
				Glossary: []GlossaryItem{{Term: "Transaction", Definition: "Атомарная операция: или все шаги выполнены, или ни один. В Go: tx.Begin → операции → tx.Commit."}},
				TestCases: []TestCase{{Input: "5\nbalance Alice 100\nbalance Bob 50\ntransfer Alice Bob 30\ntransfer Bob Alice 200\nbalance Alice", ExpectedOutput: "Alice: 100\nBob: 50\nTransfer: Alice -> Bob: 30 OK\nTransfer: Bob -> Alice: 200 FAILED (insufficient funds)\nAlice: 70"}},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { sc := bufio.NewScanner(os.Stdin); var n int; fmt.Scan(&n)
    accs := map[string]int{}
    for i := 0; i < n; i++ { sc.Scan(); p := strings.Fields(sc.Text())
        switch p[0] {
        case "balance":
            if len(p) == 3 { var amt int; fmt.Sscan(p[2], &amt); accs[p[1]] = amt; fmt.Printf("%s: %d\n", p[1], amt)
            } else { fmt.Printf("%s: %d\n", p[1], accs[p[1]]) }
        case "transfer":
            var amt int; fmt.Sscan(p[3], &amt)
            if accs[p[1]] >= amt { accs[p[1]] -= amt; accs[p[2]] += amt; fmt.Printf("Transfer: %s -> %s: %d OK\n", p[1], p[2], amt)
            } else { fmt.Printf("Transfer: %s -> %s: %d FAILED (insufficient funds)\n", p[1], p[2], amt) }
        }
    }
}`,
				Hints: `<p>Map для балансов. Transfer: проверь баланс → if ok: from -= amt, to += amt.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){sc:=bufio.NewScanner(os.Stdin);var n int;fmt.Scan(&n);a:=map[string]int{}
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text())
        switch p[0]{case "balance":if len(p)==3{var v int;fmt.Sscan(p[2],&v);a[p[1]]=v;fmt.Printf("%s: %d\n",p[1],v)}else{fmt.Printf("%s: %d\n",p[1],a[p[1]])}
        case "transfer":var v int;fmt.Sscan(p[3],&v);if a[p[1]]>=v{a[p[1]]-=v;a[p[2]]+=v;fmt.Printf("Transfer: %s -> %s: %d OK\n",p[1],p[2],v)}else{fmt.Printf("Transfer: %s -> %s: %d FAILED (insufficient funds)\n",p[1],p[2],v)}}}}</code></pre>`,
			},
			{
				Title: "DSN парсер", Difficulty: "easy",
				Description: `<p>Разбери PostgreSQL DSN строку на компоненты:</p>
<p>Ввод: <code>postgres://admin:secret@db.example.com:5432/myapp?sslmode=verify-ca</code></p>
<p>Вывод:</p><pre><code>User: admin
Host: db.example.com
Port: 5432
Database: myapp
SSLMode: verify-ca</code></pre>`,
				Glossary: []GlossaryItem{{Term: "DSN", Definition: "Data Source Name: postgres://user:pass@host:port/dbname?params."}},
				TestCases: []TestCase{{Input: "postgres://admin:secret@db.example.com:5432/myapp?sslmode=verify-ca", ExpectedOutput: "User: admin\nHost: db.example.com\nPort: 5432\nDatabase: myapp\nSSLMode: verify-ca"}},
				StarterCode: `package main
import ("fmt";"net/url")
func main() {
    var dsn string; fmt.Scan(&dsn)
    u, _ := url.Parse(dsn)
    fmt.Println("User:", u.User.Username())
    fmt.Println("Host:", u.Hostname())
    fmt.Println("Port:", u.Port())
    fmt.Println("Database:", u.Path[1:])
    fmt.Println("SSLMode:", u.Query().Get("sslmode"))
}`,
				Hints: `<p><code>url.Parse(dsn)</code> — работает для postgres:// как для http://.</p>`,
				Solution: `<pre><code>package main
import ("fmt";"net/url")
func main(){var d string;fmt.Scan(&d);u,_:=url.Parse(d)
    fmt.Println("User:",u.User.Username());fmt.Println("Host:",u.Hostname());fmt.Println("Port:",u.Port());fmt.Println("Database:",u.Path[1:]);fmt.Println("SSLMode:",u.Query().Get("sslmode"))}</code></pre>`,
			},
			{
				Title: "SQL миграция парсер", Difficulty: "easy",
				Description: `<p>Прочитай SQL миграцию и определи тип операции:</p>
<p>Ввод:</p><pre><code>4
CREATE TABLE users (id SERIAL PRIMARY KEY)
ALTER TABLE users ADD COLUMN email TEXT
DROP TABLE sessions
INSERT INTO roles (name) VALUES ('admin')</code></pre>
<p>Вывод:</p><pre><code>CREATE TABLE: users
ALTER TABLE: users
DROP TABLE: sessions
INSERT INTO: roles</code></pre>`,
				Glossary: []GlossaryItem{{Term: "Migration", Definition: "SQL-файл с изменениями схемы. UP — применить, DOWN — откатить."}},
				TestCases: []TestCase{{Input: "4\nCREATE TABLE users (id SERIAL PRIMARY KEY)\nALTER TABLE users ADD COLUMN email TEXT\nDROP TABLE sessions\nINSERT INTO roles (name) VALUES ('admin')", ExpectedOutput: "CREATE TABLE: users\nALTER TABLE: users\nDROP TABLE: sessions\nINSERT INTO: roles"}},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); p := strings.Fields(sc.Text())
        op := p[0] + " " + p[1]; table := p[2]
        fmt.Printf("%s: %s\n", op, table) } }`,
				Hints: `<p>Первые два слова = операция. Третье = таблица.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());fmt.Printf("%s %s: %s\n",p[0],p[1],p[2])}}</code></pre>`,
			},
			{
				Title: "CRUD симулятор", Difficulty: "hard",
				Description: `<p>Симулируй in-memory базу данных с CRUD операциями:</p>
<p>Ввод:</p><pre><code>6
INSERT Alice 25
INSERT Bob 30
SELECT 1
UPDATE 1 26
SELECT 1
DELETE 2</code></pre>
<p>Вывод:</p><pre><code>Inserted: id=1 name=Alice age=25
Inserted: id=2 name=Bob age=30
Found: id=1 name=Alice age=25
Updated: id=1 age=26
Found: id=1 name=Alice age=26
Deleted: id=2</code></pre>`,
				Glossary: []GlossaryItem{{Term: "CRUD", Definition: "Create, Read, Update, Delete — четыре базовые операции с данными."}},
				TestCases: []TestCase{{Input: "6\nINSERT Alice 25\nINSERT Bob 30\nSELECT 1\nUPDATE 1 26\nSELECT 1\nDELETE 2", ExpectedOutput: "Inserted: id=1 name=Alice age=25\nInserted: id=2 name=Bob age=30\nFound: id=1 name=Alice age=25\nUpdated: id=1 age=26\nFound: id=1 name=Alice age=26\nDeleted: id=2"}},
				StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
type Row struct{ ID, Age int; Name string }
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    db := map[int]Row{}; nextID := 1
    for i := 0; i < n; i++ { sc.Scan(); p := strings.Fields(sc.Text())
        switch p[0] {
        case "INSERT":
            var age int; fmt.Sscan(p[2], &age); db[nextID] = Row{nextID, age, p[1]}
            fmt.Printf("Inserted: id=%d name=%s age=%d\n", nextID, p[1], age); nextID++
        case "SELECT":
            var id int; fmt.Sscan(p[1], &id); r := db[id]
            fmt.Printf("Found: id=%d name=%s age=%d\n", r.ID, r.Name, r.Age)
        case "UPDATE":
            var id, age int; fmt.Sscan(p[1], &id); fmt.Sscan(p[2], &age)
            r := db[id]; r.Age = age; db[id] = r
            fmt.Printf("Updated: id=%d age=%d\n", id, age)
        case "DELETE":
            var id int; fmt.Sscan(p[1], &id); delete(db, id)
            fmt.Printf("Deleted: id=%d\n", id)
        }
    }
}`,
				Hints: `<p>Map[int]Row как таблица. nextID++ как SERIAL.</p>`,
				Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
type R struct{I,A int;N string}
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);db:=map[int]R{};nid:=1
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());switch p[0]{
        case "INSERT":var a int;fmt.Sscan(p[2],&a);db[nid]=R{nid,a,p[1]};fmt.Printf("Inserted: id=%d name=%s age=%d\n",nid,p[1],a);nid++
        case "SELECT":var id int;fmt.Sscan(p[1],&id);r:=db[id];fmt.Printf("Found: id=%d name=%s age=%d\n",r.I,r.N,r.A)
        case "UPDATE":var id,a int;fmt.Sscan(p[1],&id);fmt.Sscan(p[2],&a);r:=db[id];r.A=a;db[id]=r;fmt.Printf("Updated: id=%d age=%d\n",id,a)
        case "DELETE":var id int;fmt.Sscan(p[1],&id);delete(db,id);fmt.Printf("Deleted: id=%d\n",id)}}}</code></pre>`,
			},
		},
	}
}
