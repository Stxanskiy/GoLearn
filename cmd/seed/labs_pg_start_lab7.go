package main

// Лабораторная 7 курса pg-start: проект «база для службы поддержки» (ch-pgs-lab7).
//
// The course project. There is no helpdesk database yet; /root/project holds
// the support team's requirements and three CSV exports that carry their own
// ids. The names and types of every table and column are given, how to express
// the rules is the student's call: the student writes the schema to
// /root/project/schema.sql, loads the files in the order the foreign keys
// allow, moves the identity counters past the loaded ids, and proves that the
// script can be run again without touching what is already there.
//
// Two things are added to the contract's outline. The requirements say that a
// user who has tickets or comments cannot be deleted: chapter 11 teaches three
// delete rules, and the project uses two of them, so the student has to choose
// rather than leave one link unspecified. And task 5 has the student add a new
// customer and his first ticket the way the site will — without ids — so the
// counters are tested by the INSERT the student runs, not only by the check;
// the new rows must get ids above the loaded ones, because retrying an INSERT
// until it lands in a gap between the loaded ids is the wrong fix.
//
// The checks look at the result. Tables are compared column by column, types as
// format_type prints them. Every rule is a statement run in a rolled-back
// transaction that must end with a given SQLSTATE (23502 not null, 23503
// foreign key, 23505 unique, 23514 check, 00000 none), so a constraint passes
// whatever its name, inline or added by ALTER TABLE, CHECK (… IN …) or a list
// of ORs. Probe rows carry ids from 900001 up with OVERRIDING SYSTEM VALUE, so
// they never collide with the student's rows and never move a counter. Where a
// rule is "gets a value by default", the probe ends in SELECT 1/(condition)::int:
// a false condition is a division by zero, 22012. The data is compared with the
// files row by row, by id — the CSV text below is the one source of both the
// files and the comparison. The counters are read without nextval(), as in
// lab 6, so pressing "Проверить" again and again does not move them past the
// loaded ids on the student's behalf.
//
// The ★ check runs the student's schema.sql twice in a scratch database, adds
// rows between the runs, and holds what the script built to the checks of
// tasks 1–3: a script that creates the tables but not their rules, or one that
// drops and recreates them, does not pass.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab7.sh.

import (
	"encoding/csv"
	"strconv"
	"strings"
)

// pgsLab7Requirements is /root/project/requirements.md: the requirements of
// the lab's introduction, without Markdown code marks — it is read with cat
// and less in the terminal.
const pgsLab7Requirements = `# Требования: база службы поддержки «Переплёта»

Служба поддержки отвечает покупателям по почте и в чате на сайте. Сейчас
обращения записывают в таблицу Excel: номера путаются, а переписку по
обращению собирают из писем. Нужна база PostgreSQL, с которой будет работать
сайт.

## Общее

- Имя базы — helpdesk. В ней три таблицы: users, tickets и comments (схема public).
- В каждой таблице ровно те столбцы, что перечислены ниже, с этими именами и типами.
- id в каждой таблице — номер строки: тип bigint, номер ставит сама база
  (identity), первичный ключ.
- «Обязателен» — значение не может быть пустым (NULL).
- «По умолчанию» — значение, которое база ставит сама, если при добавлении
  строки столбец не указан.

## users — пользователи

Все, кто пишет в поддержку или отвечает: покупатели и сотрудники.

| Столбец    | Тип         | Правила                                                    |
|------------|-------------|------------------------------------------------------------|
| id         | bigint      | номер (identity), первичный ключ                           |
| email      | text        | обязателен; двух пользователей с одним email быть не может |
| full_name  | text        | обязателен                                                 |
| created_at | timestamptz | обязателен; по умолчанию — текущее время                   |

## tickets — обращения

| Столбец    | Тип         | Правила                                                                |
|------------|-------------|------------------------------------------------------------------------|
| id         | bigint      | номер обращения (identity), первичный ключ                             |
| author_id  | bigint      | обязателен; ссылка на пользователя из users — кто написал              |
| title      | text        | обязателен; коротко, о чём обращение                                   |
| priority   | text        | обязателен; только low, normal, high или critical; по умолчанию normal |
| status     | text        | обязателен; только open, in_progress или closed; по умолчанию open     |
| created_at | timestamptz | обязателен; по умолчанию — текущее время                               |

## comments — переписка по обращениям

| Столбец    | Тип         | Правила                                                   |
|------------|-------------|-----------------------------------------------------------|
| id         | bigint      | номер комментария (identity), первичный ключ              |
| ticket_id  | bigint      | обязателен; ссылка на обращение из tickets                |
| author_id  | bigint      | обязателен; ссылка на пользователя из users — кто написал |
| body       | text        | обязателен; текст комментария                             |
| created_at | timestamptz | обязателен; по умолчанию — текущее время                  |

## Удаление

- Обращение удаляют вместе со всей перепиской по нему: удалили строку из
  tickets — её комментарии удаляются сами.
- Пользователя, у которого есть обращения или комментарии, удалить нельзя:
  история поддержки должна сохраниться.

## Схема — в файле

- Все команды, которые создают таблицы, — в файле /root/project/schema.sql.
- Файл запускают так: psql -d helpdesk -f /root/project/schema.sql
  Поэтому базу он не создаёт и не выбирает: в нём нет \c и CREATE DATABASE.
- Файл можно запускать повторно: второй запуск проходит без ошибок и не
  трогает таблицы и данные, которые уже есть.

## Данные

Старые обращения выгружены из Excel в CSV — они в этом же каталоге:

| Файл         | Столбцы                            | Таблица  |
|--------------|------------------------------------|----------|
| users.csv    | id,email,full_name                 | users    |
| tickets.csv  | id,author_id,title,priority,status | tickets  |
| comments.csv | id,ticket_id,author_id,body        | comments |

- Номера в файлах настоящие: на них ссылаются другие файлы, а номер обращения
  покупатель видел в письмах («Ваше обращение № 7»). Загрузить их нужно как есть.
- Времени создания в файлах нет — пусть будет время загрузки.
- После загрузки новые строки добавляются без указания id и без ошибок, а
  номера получают после загруженных.
`

// The exports. Staff and customers are made up (the customers are the ones of
// lab 6), their addresses are on reserved example domains; the books in the
// tickets are real. Ids have gaps, as the ids of a live system do: test
// accounts and spam tickets were deleted. A gap is where a retried INSERT lands
// when the counter was not moved. Fields with commas are quoted, and one
// comment quotes the site's error message with "" — what CSV does with a
// double quote inside a quoted field.
const pgsLab7UsersCSV = `id,email,full_name
1,a.grigoreva@pereplet.example,Анна Григорьева
2,m.lebedev@pereplet.example,Максим Лебедев
3,k.romanova@pereplet.example,Ксения Романова
5,anna.smirnova@example.org,Анна Смирнова
6,igor.petrov@example.net,Игорь Петров
8,olga.kuznetsova@example.org,Ольга Кузнецова
9,d.volkov@example.com,Дмитрий Волков
11,maria.sokolova@example.net,Мария Соколова
12,elena.pavlova@example.com,Елена Павлова
14,p.orlov@example.net,Павел Орлов
`

const pgsLab7TicketsCSV = `id,author_id,title,priority,status
1,5,Не пришёл заказ с «Мастером и Маргаритой»,high,closed
2,6,В «Анне Карениной» перепутаны страницы,normal,closed
4,8,"Как вернуть книгу, если она не подошла?",low,closed
5,9,Деньги за заказ списали дважды,critical,in_progress
6,5,Есть ли «Евгений Онегин» в подарочном издании?,low,open
7,11,"Курьер привёз «Идиота», а заказывали «Преступление и наказание»",high,in_progress
9,12,Не получается войти в личный кабинет,normal,open
10,14,Когда снова появится «Вишнёвый сад»?,low,open
11,6,Промокод не работает на «Мёртвые души»,normal,closed
13,3,Каталог не открывается с телефона,high,open
15,8,Помятая обложка у «1984»,normal,in_progress
16,14,Нужен счёт для оплаты от школьной библиотеки,normal,open
`

const pgsLab7CommentsCSV = `id,ticket_id,author_id,body
1,1,1,"Здравствуйте! Проверяем, где посылка, ответим сегодня."
2,1,5,"Спасибо, жду."
3,1,1,"Посылка нашлась на складе службы доставки, завтра курьер её привезёт."
4,2,2,"Пришлите, пожалуйста, фото страниц — заменим книгу."
5,2,6,Отправил фото на почту.
6,2,2,"Заменили, новая книга уже в пути."
8,4,3,"Вернуть книгу можно в течение 14 дней: оформите возврат в личном кабинете."
9,5,1,"Видим два списания, передали в бухгалтерию."
10,5,9,"Прошло три дня, а деньги так и не вернулись."
11,7,2,"Извините за путаницу! Курьер заберёт «Идиота» и привезёт нужную книгу."
12,9,1,Попробуйте сбросить пароль по ссылке «Забыли пароль?» на странице входа.
14,11,2,"Промокод действует только на подборку «Классика за полцены», а «Мёртвые души» в неё не входят."
15,11,6,"Понял, спасибо."
16,13,2,"Повторил на телефоне: вместо каталога страница ""Ошибка 500""."
17,15,3,"Заменим книгу, курьер заберёт помятую."
19,5,1,"Бухгалтерия вернула деньги, проверьте, пожалуйста, выписку."
`

// pgsLab7File is a setup step writing content to path, base64 like the SQL of
// the checks, so no line of the data can end a heredoc early.
func pgsLab7File(path, content string) string {
	return "echo " + b64(content) + " | base64 -d > " + path + "\n"
}

// pgsLab7Data is one export as the checks see it.
type pgsLab7Data struct {
	table string
	cols  []string // the file's header
	want  string   // "want_<table> (cols) AS (VALUES …)", for a WITH
	rows  int
	maxID int
}

// pgsLab7Parse reads a CSV export into a VALUES list: id and *_id columns as
// numbers, everything else as text literals.
func pgsLab7Parse(table, file string) pgsLab7Data {
	recs, err := csv.NewReader(strings.NewReader(file)).ReadAll()
	if err != nil {
		panic("pg-start lab 7: " + table + ".csv: " + err.Error())
	}
	d := pgsLab7Data{table: table, cols: recs[0], rows: len(recs) - 1}
	var rows []string
	for _, rec := range recs[1:] {
		var vals []string
		for i, f := range rec {
			col := d.cols[i]
			switch {
			case col == "id":
				n, err := strconv.Atoi(f)
				if err != nil {
					panic("pg-start lab 7: " + table + ".csv: id " + f)
				}
				d.maxID = max(d.maxID, n)
				vals = append(vals, f)
			case strings.HasSuffix(col, "_id"):
				vals = append(vals, f)
			default:
				vals = append(vals, "'"+strings.ReplaceAll(f, "'", "''")+"'")
			}
		}
		rows = append(rows, "("+strings.Join(vals, ", ")+")")
	}
	d.want = "want_" + table + " (" + strings.Join(d.cols, ", ") + ") AS (VALUES " + strings.Join(rows, ", ") + ")"
	return d
}

var (
	pgsLab7Users    = pgsLab7Parse("users", pgsLab7UsersCSV)
	pgsLab7Tickets  = pgsLab7Parse("tickets", pgsLab7TicketsCSV)
	pgsLab7Comments = pgsLab7Parse("comments", pgsLab7CommentsCSV)
	pgsLab7Files    = []pgsLab7Data{pgsLab7Users, pgsLab7Tickets, pgsLab7Comments}
)

// pgsLab7Table is a table of the requirements: its columns as "name type", the
// type as format_type prints it.
type pgsLab7Table struct {
	name string
	cols []string
}

var pgsLab7Tables = []pgsLab7Table{
	{"users", []string{"id bigint", "email text", "full_name text", "created_at timestamp with time zone"}},
	{"tickets", []string{"id bigint", "author_id bigint", "title text", "priority text", "status text",
		"created_at timestamp with time zone"}},
	{"comments", []string{"id bigint", "ticket_id bigint", "author_id bigint", "body text",
		"created_at timestamp with time zone"}},
}

// problems is the SQL of what is wrong with the table — "таблицы … нет" (and
// where a table of that name is instead), or missing columns, wrong types and
// extra columns — or NULL when it is as the requirements say.
func (t pgsLab7Table) problems() string {
	var want, names []string
	for i, c := range t.cols {
		name, typ, _ := strings.Cut(c, " ")
		want = append(want, "("+strconv.Itoa(i+1)+", '"+name+"', '"+typ+"')")
		names = append(names, "'"+name+"'")
	}
	rel := "to_regclass('" + t.name + "')"
	return `CASE WHEN ` + rel + ` IS NULL THEN 'таблицы ` + t.name + ` нет' || coalesce(' (есть ' ||
        (SELECT string_agg(format('%I.%I', schemaname, tablename), ', ') FROM pg_tables
         WHERE lower(tablename) = '` + t.name + `' AND schemaname NOT IN ('pg_catalog', 'information_schema')) || ' — не то имя или не та схема)', '')
    ELSE 'в ` + t.name + `: ' || nullif(concat_ws(', ',
        (SELECT string_agg(CASE WHEN a.attname IS NULL THEN 'нет столбца ' || w.col
                ELSE 'у ' || w.col || ' тип ' || format_type(a.atttypid, a.atttypmod) || ', а нужен ' || w.typ END, ', ' ORDER BY w.n)
         FROM (VALUES ` + strings.Join(want, ", ") + `) AS w (n, col, typ)
         LEFT JOIN pg_attribute a ON a.attrelid = ` + rel + ` AND a.attname = w.col AND a.attnum > 0 AND NOT a.attisdropped
         WHERE a.attname IS NULL OR format_type(a.atttypid, a.atttypmod) <> w.typ),
        (SELECT CASE count(*) WHEN 1 THEN 'лишний столбец ' ELSE 'лишние столбцы ' END || string_agg(attname, ', ' ORDER BY attnum) FROM pg_attribute
         WHERE attrelid = ` + rel + ` AND attnum > 0 AND NOT attisdropped AND attname NOT IN (` + strings.Join(names, ", ") + `))), '') END`
}

// pgsLab7Shape is "ok", or what is wrong with the three tables, "; " between
// tables.
func pgsLab7Shape() string {
	var parts []string
	for _, t := range pgsLab7Tables {
		parts = append(parts, t.problems())
	}
	return "SELECT coalesce(nullif(concat_ws('; ',\n    " + strings.Join(parts, ",\n    ") + "), ''), 'ok')"
}

// pgsLab7Keys is "ok", or the tables whose id is not an identity column or
// not the primary key, grouped by what is missing.
const pgsLab7Keys = `SELECT coalesce(string_agg(problem || ' в ' || tables, '; ' ORDER BY first), 'ok')
FROM (SELECT concat_ws(' и ', CASE WHEN NOT ident THEN 'номер не ставит база (не identity)' END,
                CASE WHEN NOT pk THEN 'не первичный ключ' END) AS problem,
             string_agg(t, ', ' ORDER BY n) AS tables, min(n) AS first
      FROM (SELECT v.n, v.t,
              EXISTS (SELECT 1 FROM pg_attribute a WHERE a.attrelid = to_regclass(v.t) AND a.attname = 'id' AND a.attidentity IN ('a', 'd')) AS ident,
              EXISTS (SELECT 1 FROM pg_attribute a JOIN pg_constraint c ON c.conrelid = a.attrelid AND c.contype = 'p' AND c.conkey = ARRAY[a.attnum]
                      WHERE a.attrelid = to_regclass(v.t) AND a.attname = 'id') AS pk
            FROM (VALUES (1, 'users'), (2, 'tickets'), (3, 'comments')) AS v (n, t)) x
      WHERE NOT (ident AND pk)
      GROUP BY 1) y`

// pgsLab7Rule is one behaviour the requirements ask for: sql runs in a
// transaction that is rolled back and must end with SQLSTATE want (00000 — no
// error). what names the attempt in a diagnostic, without quotes of either
// kind.
type pgsLab7Rule struct{ what, sql, want string }

// probe is the rule's SQL as it runs: a deferrable constraint would only fire
// at a COMMIT that never comes, so it is made to fire now.
func (r pgsLab7Rule) probe() string { return r.sql + ";\nSET CONSTRAINTS ALL IMMEDIATE" }

// The rows rules lean on: user 900001, his ticket 900002 and his comment on it
// 900003. Every column is given — the defaults may not exist yet — so a
// probe that expects a refusal gets it from the rule under test. The user and
// the ticket have different ids, so a link to the wrong table shows.
const (
	pgsLab7UserRow    = `INSERT INTO users (id, email, full_name, created_at) OVERRIDING SYSTEM VALUE VALUES `
	pgsLab7TicketRow  = `INSERT INTO tickets (id, author_id, title, priority, status, created_at) OVERRIDING SYSTEM VALUE VALUES `
	pgsLab7CommentRow = `INSERT INTO comments (id, ticket_id, author_id, body, created_at) OVERRIDING SYSTEM VALUE VALUES `

	pgsLab7User    = pgsLab7UserRow + `(900001, 'glpg-probe-900001@check.invalid', 'Проверка', now());` + "\n"
	pgsLab7Ticket  = pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', 'normal', 'open', now());` + "\n"
	pgsLab7Comment = pgsLab7Ticket + pgsLab7CommentRow + `(900003, 900002, 900001, 'Проверка', now());` + "\n"

	// pgsLab7Recent is "within a minute of now()": what DEFAULT now() gives,
	// and current_timestamp or clock_timestamp() too.
	pgsLab7Recent = ` BETWEEN now() - interval '1 minute' AND now() + interval '1 minute'`
)

// Task 2: the links, and the delete rule of the links to users.
var pgsLab7KeyRules = []pgsLab7Rule{
	{"обращение от пользователя с id 999999, которого нет",
		pgsLab7TicketRow + `(900002, 999999, 'Проверка', 'normal', 'open', now())`, "23503"},
	{"комментарий к обращению с id 999999, которого нет",
		pgsLab7Ticket + pgsLab7CommentRow + `(900003, 999999, 900001, 'Проверка', now())`, "23503"},
	{"комментарий от пользователя с id 999999, которого нет",
		pgsLab7Ticket + pgsLab7CommentRow + `(900003, 900002, 999999, 'Проверка', now())`, "23503"},
	{"обращение нового пользователя и комментарий к нему",
		pgsLab7Comment + `SELECT 1`, "00000"},
	{"удаление пользователя, у которого есть обращение",
		pgsLab7Ticket + `DELETE FROM users WHERE id = 900001`, "23503"},
	{"удаление пользователя, у которого есть только комментарий",
		pgsLab7Ticket + pgsLab7UserRow + `(900004, 'glpg-probe-900004@check.invalid', 'Проверка', now());
` + pgsLab7CommentRow + `(900003, 900002, 900004, 'Проверка', now());
DELETE FROM users WHERE id = 900004`, "23503"},
}

// Task 3: what is required, unique, one of a list, set by default, and deleted
// along with a ticket.
var pgsLab7Rules = []pgsLab7Rule{
	{"два пользователя с одним email",
		pgsLab7UserRow + `(900001, 'glpg-probe-900001@check.invalid', 'Проверка', now()),
    (900002, 'glpg-probe-900001@check.invalid', 'Проверка', now())`, "23505"},
	{"пользователь без email",
		pgsLab7UserRow + `(900001, NULL, 'Проверка', now())`, "23502"},
	{"пользователь без имени",
		pgsLab7UserRow + `(900001, 'glpg-probe-900001@check.invalid', NULL, now())`, "23502"},
	{"пользователь с created_at = NULL",
		pgsLab7UserRow + `(900001, 'glpg-probe-900001@check.invalid', 'Проверка', NULL)`, "23502"},
	{"новый пользователь без created_at получает текущее время",
		`INSERT INTO users (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (900001, 'glpg-probe-900001@check.invalid', 'Проверка');
SELECT 1/coalesce((SELECT created_at` + pgsLab7Recent + ` FROM users WHERE id = 900001)::int, 0)`, "00000"},
	{"обращение без автора",
		pgsLab7TicketRow + `(900002, NULL, 'Проверка', 'normal', 'open', now())`, "23502"},
	{"обращение без названия",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, NULL, 'normal', 'open', now())`, "23502"},
	{"приоритет urgent",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', 'urgent', 'open', now())`, "23514"},
	{"приоритет NULL",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', NULL, 'open', now())`, "23502"},
	{"статус done",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', 'normal', 'done', now())`, "23514"},
	{"статус NULL",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', 'normal', NULL, now())`, "23502"},
	{"обращение с created_at = NULL",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', 'normal', 'open', NULL)`, "23502"},
	{"обращения с приоритетами low, normal, high, critical и статусами open, in_progress, closed",
		pgsLab7User + pgsLab7TicketRow + `(900002, 900001, 'Проверка', 'low', 'open', now()),
    (900003, 900001, 'Проверка', 'normal', 'in_progress', now()),
    (900004, 900001, 'Проверка', 'high', 'closed', now()),
    (900005, 900001, 'Проверка', 'critical', 'open', now())`, "00000"},
	{"новое обращение без priority получает normal",
		pgsLab7User + `INSERT INTO tickets (id, author_id, title, status, created_at) OVERRIDING SYSTEM VALUE VALUES (900002, 900001, 'Проверка', 'open', now());
SELECT 1/coalesce((SELECT priority = 'normal' FROM tickets WHERE id = 900002)::int, 0)`, "00000"},
	{"новое обращение без status получает open",
		pgsLab7User + `INSERT INTO tickets (id, author_id, title, priority, created_at) OVERRIDING SYSTEM VALUE VALUES (900002, 900001, 'Проверка', 'normal', now());
SELECT 1/coalesce((SELECT status = 'open' FROM tickets WHERE id = 900002)::int, 0)`, "00000"},
	{"новое обращение без created_at получает текущее время",
		pgsLab7User + `INSERT INTO tickets (id, author_id, title, priority, status) OVERRIDING SYSTEM VALUE VALUES (900002, 900001, 'Проверка', 'normal', 'open');
SELECT 1/coalesce((SELECT created_at` + pgsLab7Recent + ` FROM tickets WHERE id = 900002)::int, 0)`, "00000"},
	{"комментарий без обращения",
		pgsLab7Ticket + pgsLab7CommentRow + `(900003, NULL, 900001, 'Проверка', now())`, "23502"},
	{"комментарий без автора",
		pgsLab7Ticket + pgsLab7CommentRow + `(900003, 900002, NULL, 'Проверка', now())`, "23502"},
	{"комментарий без текста",
		pgsLab7Ticket + pgsLab7CommentRow + `(900003, 900002, 900001, NULL, now())`, "23502"},
	{"комментарий с created_at = NULL",
		pgsLab7Ticket + pgsLab7CommentRow + `(900003, 900002, 900001, 'Проверка', NULL)`, "23502"},
	{"новый комментарий без created_at получает текущее время",
		pgsLab7Ticket + `INSERT INTO comments (id, ticket_id, author_id, body) OVERRIDING SYSTEM VALUE VALUES (900003, 900002, 900001, 'Проверка');
SELECT 1/coalesce((SELECT created_at` + pgsLab7Recent + ` FROM comments WHERE id = 900003)::int, 0)`, "00000"},
	{"удаление обращения вместе с его комментариями",
		pgsLab7Comment + `DELETE FROM tickets WHERE id = 900002;
SELECT 1/coalesce((SELECT count(*) = 0 FROM comments WHERE ticket_id = 900002)::int, 0)`, "00000"},
}

// pgsLab7Shown is how many broken rules a diagnostic names; the rest are
// counted. A table with no rules at all breaks most of task 3's, and a list of
// twenty does not fit a line.
const pgsLab7Shown = 4

// pgsLab7Broken is a shell snippet printing "what — what the database did; "
// for each rule that does not behave in db: "база это принимает" where a
// refusal was due, "нет" where a probe's value is wrong (22012), and otherwise
// the error itself, the line a student sees in psql — "null value in column
// "created_at" of relation "tickets" violates not-null constraint" says which
// column is at fault, where a bare SQLSTATE would not. In verbose mode the
// first line of an error is "ERROR:  23502: message".
func pgsLab7Broken(db string, rules []pgsLab7Rule) string {
	var b strings.Builder
	b.WriteString(`n=0; `)
	for _, r := range rules {
		b.WriteString(`e=$({ echo 'BEGIN;'; echo ` + b64(r.probe()) + ` | base64 -d; echo ';'; echo 'ROLLBACK;'; } | ` +
			`psql -X -q -v ON_ERROR_STOP=1 -v VERBOSITY=verbose -d ` + db + ` 2>&1 >/dev/null | grep -m1 '^ERROR:'); ` +
			`s=${e:8:5}; s=${s:-00000}; [ $s = ` + r.want + ` ] || { n=$((n+1)); [ $n -gt ` + strconv.Itoa(pgsLab7Shown) + ` ] || { ` +
			`echo -n '` + r.what + ` — '; case $s in (00000) echo -n 'база это принимает';; (22012) echo -n 'нет';; ` +
			`(*) echo -n "ошибка: ${e#ERROR:  ?????: }";; esac; echo -n '; '; }; }; `)
	}
	b.WriteString(`[ $n -le ` + strconv.Itoa(pgsLab7Shown) + ` ] || echo -n "и ещё правил: $((n-` + strconv.Itoa(pgsLab7Shown) + `))"; `)
	return b.String()
}

// pgsLab7RulesHold is a shell snippet that prints the broken rules after
// prefix and exits 1 when any rule does not behave in db.
func pgsLab7RulesHold(db, prefix string, rules []pgsLab7Rule) string {
	return `r=$(` + pgsLab7Broken(db, rules) + `); [ -z "$r" ] || { r=${r%; }; echo "` + prefix + `$r"; exit 1; }; `
}

// pgsLab7DB is the condition "helpdesk exists".
var pgsLab7DB = pgDBExists("helpdesk")

// pgsLab7InRoot prints where the tables went instead, when a bare psql or a
// psql -f without -d put them into the database root.
var pgsLab7InRoot = `glpg q root ` + b64(`SELECT ' — а в базе root есть ' || string_agg(tablename, ', ' ORDER BY tablename) ||
    ': похоже, команды выполнялись не в той базе'
FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('users', 'tickets', 'comments')`) + `; `

// pgsLab7NeedShape stops a check with a diagnostic unless helpdesk and its
// tables are as task 1 asks: the later checks have nothing to say before that.
var pgsLab7NeedShape = pgsLab7DB + ` || { echo "Сейчас: базы helpdesk нет — её и таблицы создаёт задание «База и три таблицы»"; exit 1; }; ` +
	`p=$(glpg q helpdesk ` + b64(pgsLab7Shape()) + `); [ "$p" = ok ] || ` +
	`{ echo "Сейчас: таблицы ещё не такие, как в требованиях (задание «База и три таблицы»): ${p:-база helpdesk не отвечает}"; exit 1; }; `

// Task 1. helpdesk, the three tables, their columns and types.
var pgsLab7Task1 = pgsLab7DB + ` || { echo -n "Сейчас: базы helpdesk нет"; ` + pgsLab7InRoot + `exit 1; }; ` +
	`p=$(glpg q helpdesk ` + b64(pgsLab7Shape()) + `); [ "$p" = ok ] && exit 0; ` +
	`echo -n "Сейчас: ${p:-база helpdesk не отвечает}"; ` + pgsLab7InRoot + `exit 1`

// Task 2. Identity primary keys, then the links.
var pgsLab7Task2 = pgsLab7NeedShape +
	`k=$(glpg q helpdesk ` + b64(pgsLab7Keys) + `); [ "$k" = ok ] || { echo "Сейчас: ${k:-база helpdesk не отвечает}"; exit 1; }; ` +
	pgsLab7RulesHold("helpdesk", "Сейчас: ", pgsLab7KeyRules)

// Task 3. The rules.
var pgsLab7Task3 = pgsLab7NeedShape + pgsLab7RulesHold("helpdesk", "Сейчас: ", pgsLab7Rules)

// pgsLab7Same is the SQL condition "table row t holds the file's row w".
func pgsLab7Same(d pgsLab7Data) string {
	var eq []string
	for _, c := range d.cols {
		eq = append(eq, "t."+c+" IS NOT DISTINCT FROM w."+c)
	}
	return strings.Join(eq, " AND ")
}

// pgsLab7IDs is a SQL scalar subquery: up to five ids of the rows that from
// (a FROM … WHERE … clause over alias x) gives, then "…"; NULL when none.
func pgsLab7IDs(from string) string {
	return `(SELECT array_to_string((array_agg(x.id ORDER BY x.id))[1:5], ', ') || CASE WHEN count(*) > 5 THEN ', …' ELSE '' END ` +
		from + ` HAVING count(*) > 0)`
}

// pgsLab7DataDiff is "ok" when each table holds every row of its file, with
// the file's id and values, and — when strict — no other row among the file's
// ids. Rows above the file's largest id are always left alone: task 5 adds
// them. Task 5 itself is not strict: a new row that landed in a gap between
// the loaded ids is its problem, and its own diagnostic says so.
func pgsLab7DataDiff(strict bool) string {
	var with, parts, oks []string
	for _, d := range pgsLab7Files {
		w := "want_" + d.table
		same := `(SELECT count(*) FROM ` + w + ` w WHERE EXISTS (SELECT 1 FROM ` + d.table + ` t WHERE ` + pgsLab7Same(d) + `))`
		extra := `FROM ` + d.table + ` x WHERE x.id <= ` + strconv.Itoa(d.maxID) + ` AND x.id NOT IN (SELECT id FROM ` + w + `)`
		with = append(with, d.want)
		ok, extras := same+` = `+strconv.Itoa(d.rows), `''`
		if strict {
			ok += ` AND NOT EXISTS (SELECT 1 ` + extra + `)`
			extras = `coalesce(', строки не из файла: id ' || ` + pgsLab7IDs(extra) + `, '')`
		}
		oks = append(oks, ok)
		parts = append(parts, `CASE WHEN NOT EXISTS (SELECT 1 FROM `+d.table+`) THEN '`+d.table+` — пустая (в `+d.table+`.csv строк: `+strconv.Itoa(d.rows)+`)'
        ELSE '`+d.table+` — совпадает с файлом строк: ' || `+same+` || ' из `+strconv.Itoa(d.rows)+`' ||
        coalesce(', нет строк с id ' || `+pgsLab7IDs(`FROM `+w+` x WHERE NOT EXISTS (SELECT 1 FROM `+d.table+` t WHERE t.id = x.id)`)+`, '') ||
        coalesce(', отличаются от файла строки с id ' || `+pgsLab7IDs(`FROM `+w+` x WHERE EXISTS (SELECT 1 FROM `+d.table+` t WHERE t.id = x.id)
            AND NOT EXISTS (SELECT 1 FROM `+d.table+` t, `+w+` w WHERE w.id = x.id AND `+pgsLab7Same(d)+`)`)+`, '') || `+extras+` END`)
	}
	return `WITH ` + strings.Join(with, ",\n") + `
SELECT CASE WHEN ` + strings.Join(oks, "\n    AND ") + ` THEN 'ok'
    ELSE concat_ws('; ', ` + strings.Join(parts, ",\n    ") + `) END`
}

// Task 4. The data of the files.
var pgsLab7Task4 = pgsLab7NeedShape +
	`p=$(glpg q helpdesk ` + b64(pgsLab7DataDiff(true)) + `); [ "$p" = ok ] || { echo "Сейчас: ${p:-база helpdesk не отвечает}"; exit 1; }`

// pgsLab7Next is the SQL of the value the identity counter of table.id hands
// out next, read without nextval(): the sequence relation is selected through
// query_to_xml, because a plain query cannot take a table name from
// pg_get_serial_sequence. Whatever the student did to the counter — setval,
// setval(…, false), ALTER … RESTART — shows here. NULL when there is no counter.
func pgsLab7Next(table string) string {
	return `(SELECT CASE WHEN (xpath('/row/is_called/text()', x))[1]::text = 'true'
            THEN (xpath('/row/last_value/text()', x))[1]::text::bigint + s.seqincrement
            ELSE (xpath('/row/last_value/text()', x))[1]::text::bigint END
    FROM pg_sequence s, query_to_xml('SELECT last_value, is_called FROM ' || s.seqrelid::regclass, false, true, '') AS x
    WHERE s.seqrelid = to_regclass(pg_get_serial_sequence('` + table + `', 'id')))`
}

// The new customer of task 5 and his ticket.
const (
	pgsLab7NewEmail  = "t.zakharov@example.com"
	pgsLab7NewName   = "Тимур Захаров"
	pgsLab7NewTicket = "Не приходит письмо с подтверждением регистрации"
)

// pgsLab7NewRows is "ok", or what is wrong with the new customer, his ticket
// and the three counters.
func pgsLab7NewRows() string {
	u, t := strconv.Itoa(pgsLab7Users.maxID), strconv.Itoa(pgsLab7Tickets.maxID)
	var counters []string
	for _, d := range pgsLab7Files {
		counters = append(counters, `('`+d.table+`', `+pgsLab7Next(d.table)+`, (SELECT max(id) FROM `+d.table+`))`)
	}
	return `WITH u AS (SELECT * FROM users WHERE email = '` + pgsLab7NewEmail + `'),
t AS (SELECT * FROM tickets WHERE author_id IN (SELECT id FROM u)),
c (tab, next, top) AS (VALUES ` + strings.Join(counters, ",\n    ") + `),
p AS (SELECT concat_ws('; ',
    CASE (SELECT count(*) FROM u)
        WHEN 0 THEN 'пользователя ` + pgsLab7NewEmail + ` в users нет'
        WHEN 1 THEN (SELECT nullif(concat_ws(', ',
            CASE WHEN full_name <> '` + pgsLab7NewName + `' THEN 'у пользователя ` + pgsLab7NewEmail + ` имя «' || full_name || '»' END,
            CASE WHEN id <= ` + u + ` THEN 'пользователь ` + pgsLab7NewEmail + ` получил id ' || id ||
                ' — номер среди загруженных (нужен больше ` + u + `)' END), '') FROM u)
        ELSE 'пользователей ` + pgsLab7NewEmail + ` в users: ' || (SELECT count(*) FROM u) END,
    CASE WHEN (SELECT count(*) FROM u) = 1 THEN CASE (SELECT count(*) FROM t)
        WHEN 0 THEN 'обращений этого пользователя в tickets нет'
        WHEN 1 THEN (SELECT nullif(concat_ws(', ',
            CASE WHEN title <> '` + pgsLab7NewTicket + `' THEN 'у его обращения название «' || title || '»' END,
            CASE WHEN priority IS DISTINCT FROM 'high' THEN 'у его обращения приоритет ' || coalesce(priority, 'NULL') END,
            CASE WHEN status IS DISTINCT FROM 'open' THEN 'у его обращения статус ' || coalesce(status, 'NULL') END,
            CASE WHEN id <= ` + t + ` THEN 'его обращение получило id ' || id ||
                ' — номер среди загруженных (нужен больше ` + t + `)' END), '') FROM t)
        ELSE 'обращений этого пользователя в tickets: ' || (SELECT count(*) FROM t) END END,
    (SELECT string_agg(CASE WHEN next IS NULL THEN 'у ' || tab || '.id нет счётчика'
            ELSE 'счётчик ' || tab || ' выдаст следующим номер ' || next || ', а наибольший id в ' || tab || ' — ' || top END, '; ')
     FROM c WHERE next IS NULL OR next <= top)) AS msg)
SELECT coalesce(nullif(msg, ''), 'ok') FROM p`
}

// Task 5. Needs the data of task 4: with empty tables any counter is ahead.
var pgsLab7Task5 = pgsLab7NeedShape +
	`p=$(glpg q helpdesk ` + b64(pgsLab7DataDiff(false)) + `); [ "$p" = ok ] || ` +
	`{ echo "Сейчас: данные ещё не загружены как надо (задание «Загрузи старые обращения»): ${p:-база helpdesk не отвечает}"; exit 1; }; ` +
	`p=$(glpg q helpdesk ` + b64(pgsLab7NewRows()) + `); [ "$p" = ok ] || { echo "Сейчас: ${p:-база helpdesk не отвечает}"; exit 1; }`

// The ★ task's rows added between the two runs: a second run that drops and
// recreates a table loses them. Their ids are outside the probes' range.
const (
	pgsLab7Keep = `INSERT INTO users (id, email, full_name, created_at) OVERRIDING SYSTEM VALUE
    VALUES (800001, 'glpg-keep-800001@check.invalid', 'Проверка', now());
INSERT INTO tickets (id, author_id, title, priority, status, created_at) OVERRIDING SYSTEM VALUE
    VALUES (800002, 800001, 'Проверка', 'normal', 'open', now());
INSERT INTO comments (id, ticket_id, author_id, body, created_at) OVERRIDING SYSTEM VALUE
    VALUES (800003, 800002, 800001, 'Проверка', now());`
	pgsLab7Kept = `SELECT (SELECT count(*) FROM users WHERE id = 800001) + (SELECT count(*) FROM tickets WHERE id = 800002)
    + (SELECT count(*) FROM comments WHERE id = 800003)`
)

// pgsLab7Schema judges /root/project/schema.sql: run it in a scratch
// database, add a user, a ticket and a comment, run it again, stopping at the
// first error, and hold what it built to the checks of tasks 1–3. The scratch
// database is dropped on every way out.
//
// The file is refused without being run if it connects anywhere (\c) or
// touches a whole database: run as a superuser, a DROP DATABASE helpdesk in it
// would take the student's own work. Comments are stripped first. Errors are
// also looked for in the output, so a file that turns ON_ERROR_STOP off does
// not hide them.
var pgsLab7Schema = `f=/root/project/schema.sql; t=pgs_lab7_check_$$; ` +
	`[ -s $f ] || { echo "Сейчас: файла $f нет или он пустой"; exit 1; }; ` +
	`bad=$(sed 's/--.*$//' $f | grep -oiE '\\(c|connect)([[:space:]]|$)|(create|drop|alter)[[:space:]]+database' | head -n1 | sed 's/[[:space:]]*$//'); ` +
	`[ -z "$bad" ] || { echo "Сейчас: в $f есть «$bad» — скрипт схемы не создаёт, не удаляет и не выбирает базу: её задают при запуске, psql -d"; exit 1; }; ` +
	`trap 'dropdb --if-exists --force $t >/dev/null 2>&1' EXIT; ` +
	`createdb -T template0 $t || { echo "не удалось создать проверочную базу — сообщи авторам курса"; exit 1; }; ` +
	`for i in первом втором; do ` +
	`e=$(psql -X -q -v ON_ERROR_STOP=1 -d $t -f $f 2>&1 >/dev/null); r=$?; ` +
	`if [ $r != 0 ] || printf '%s\n' "$e" | grep -q 'ERROR:'; then ` +
	`echo "Сейчас: при $i запуске в чистой базе ошибка: $(printf '%s\n' "$e" | grep -m1 -E 'ERROR|FATAL' | sed -E 's/^psql:[^:]*:([0-9]+): /строка \1: /')"; exit 1; fi; ` +
	`[ $i = первом ] || continue; ` +
	`p=$(glpg q $t ` + b64(pgsLab7Shape()) + `); [ "$p" = ok ] || { echo "Сейчас: после запуска в чистой базе ${p:-проверить таблицы не удалось}"; exit 1; }; ` +
	`e=$(glpg run $t ` + b64(pgsLab7Keep) + ` 2>&1) || { echo "Сейчас: после первого запуска не удалось добавить пользователя, обращение и комментарий: $(printf '%s\n' "$e" | grep -m1 'ERROR' | sed 's/^.*ERROR: *//')"; exit 1; }; ` +
	`done; ` +
	`[ "$(glpg q $t ` + b64(pgsLab7Kept) + `)" = 3 ] || { echo "Сейчас: оба запуска прошли без ошибок, но второй стёр строки, добавленные после первого, — повторный запуск не должен трогать данные"; exit 1; }; ` +
	`p=$(glpg q $t ` + b64(pgsLab7Shape()) + `); [ "$p" = ok ] || { echo "Сейчас: после второго запуска ${p:-проверить таблицы не удалось}"; exit 1; }; ` +
	`k=$(glpg q $t ` + b64(pgsLab7Keys) + `); [ "$k" = ok ] || { echo "Сейчас: в таблицах, которые создаёт скрипт, ${k:-проверить ключи не удалось}"; exit 1; }; ` +
	pgsLab7RulesHold("$t", "Сейчас: в таблицах, которые создаёт скрипт, не выполняются правила: ", append(append([]pgsLab7Rule{}, pgsLab7KeyRules...), pgsLab7Rules...))

func init() {
	pgStartLabs["ch-pgs-lab7"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup("dropdb --if-exists --force helpdesk\nmkdir -p /root/project\n" +
			pgsLab7File("/root/project/requirements.md", pgsLab7Requirements) +
			pgsLab7File("/root/project/users.csv", pgsLab7UsersCSV) +
			pgsLab7File("/root/project/tickets.csv", pgsLab7TicketsCSV) +
			pgsLab7File("/root/project/comments.csv", pgsLab7CommentsCSV)),
		Checks: map[int]string{
			1: pgCheck(pgsLab7Task1,
				"есть база helpdesk, в ней таблицы users, tickets и comments — ровно со столбцами и типами из требований"),
			2: pgCheck(pgsLab7Task2,
				"id в каждой таблице — identity и первичный ключ; обращение ссылается на пользователя, комментарий — на обращение и на пользователя; пользователя с перепиской удалить нельзя"),
			3: pgCheck(pgsLab7Task3,
				"выполняются все правила из требований: обязательные столбцы, уникальный email, допустимые приоритеты и статусы, значения по умолчанию, удаление обращения вместе с комментариями"),
			4: pgCheck(pgsLab7Task4,
				"в users, tickets и comments — все строки из users.csv, tickets.csv и comments.csv, с теми же id и значениями, и других строк с такими номерами нет"),
			5: pgCheck(pgsLab7Task5,
				"Тимур Захаров и его обращение добавлены с номерами после загруженных; счётчик id каждой из трёх таблиц выдаст следующим номер больше наибольшего id в ней"),
			6: pgCheck(pgsLab7Schema,
				"/root/project/schema.sql — без \\c и CREATE DATABASE; два запуска подряд в чистой базе проходят без ошибок, второй не трогает данные, а таблицы получаются точно по требованиям — с ключами и правилами"),
		},
	}
}
