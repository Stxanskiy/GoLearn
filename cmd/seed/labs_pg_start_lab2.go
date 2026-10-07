package main

// Лабораторная 2 курса pg-start: осваиваем psql (ch-pgs-lab2).
//
// The student looks around the server (\l, \c, \dt, \d), runs a ready-made
// script, saves query output to files the way a shell script would read it
// (psql -At, or \t \a \o inside psql), finds an option of CREATE DATABASE in the
// built-in help and sets up ~/.psqlrc.
//
// Every check below runs psql with -X (glpg does), so the student's own
// ~/.psqlrc — the last task — cannot change what an earlier check sees. Only
// the ~/.psqlrc check itself runs psql without it.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab2.sh.

// pgsLab2Pereplet is the canonical authors and books (§5 of the course
// contract) with a small catalogue, plus the empty visits log hello.sql writes
// to. One book has no page count and no ISBN, so a NULL shows up in the data
// the student looks at once ~/.psqlrc makes NULL visible.
const pgsLab2Pereplet = `
CREATE TABLE authors (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name  text NOT NULL,
    birth_year integer,
    country    text
);

CREATE TABLE books (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title          text NOT NULL,
    author_id      bigint REFERENCES authors (id),
    price          numeric(8,2) NOT NULL CONSTRAINT books_price_positive CHECK (price > 0),
    page_count     integer,
    published_year integer,
    isbn           text UNIQUE,
    in_stock       boolean NOT NULL DEFAULT true
);

CREATE TABLE visits (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    note       text NOT NULL,
    visited_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO authors (full_name, birth_year, country) VALUES
    ('Михаил Булгаков', 1891, 'Россия'),
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Антон Чехов', 1860, 'Россия'),
    ('Николай Гоголь', 1809, 'Россия'),
    ('Джордж Оруэлл', 1903, 'Великобритания'),
    ('Эрих Мария Ремарк', 1898, 'Германия');

INSERT INTO books (title, author_id, price, page_count, published_year, isbn, in_stock) VALUES
    ('Мастер и Маргарита',       1, 690.00, 480, 1967, '978-5-389-01686-6', true),
    ('Собачье сердце',           1, 350.00, 160, 1987, '978-5-389-07435-4', true),
    ('Преступление и наказание', 2, 590.00, 608, 1866, '978-5-389-04926-0', true),
    ('Идиот',                    2, 540.00, 640, 1869, '978-5-389-05178-2', false),
    ('Вишнёвый сад',             3, 280.00,  96, 1904, '978-5-389-06010-4', true),
    ('Мёртвые души',             4, 450.00, 352, 1842, '978-5-389-04720-4', true),
    ('Ревизор',                  4, 240.00, 128, 1836, '978-5-389-08562-6', true),
    ('1984',                     5, 520.00, 320, 1949, '978-5-17-080115-2', true),
    ('Скотный двор',             5, 300.00, NULL, 1945, NULL, true),
    ('Три товарища',             6, 610.00, 480, 1936, '978-5-17-087883-3', true);
`

// pgsLab2Archive is the second database: the shop's old orders, moved out of
// Excel as they were. It is there to be seen in \l and to be the wrong place
// to look for pereplet's tables.
const pgsLab2Archive = `
CREATE TABLE old_orders (
    order_no   integer PRIMARY KEY,
    ordered_on date NOT NULL,
    customer   text NOT NULL,
    book_title text NOT NULL,
    quantity   integer NOT NULL
);

INSERT INTO old_orders (order_no, ordered_on, customer, book_title, quantity) VALUES
    (1, '2023-11-03', 'Ирина Соколова', 'Мастер и Маргарита', 1),
    (2, '2023-12-18', 'Павел Орлов', '1984', 2);
`

// pgsLab2Hello is /root/sql/hello.sql. The \echo is deliberate: a script run
// with psql -f or \i executes meta-commands too.
const pgsLab2Hello = `-- hello.sql — первый скрипт «Переплёта».
-- Записывает в журнал visits, что psql работает,
-- и показывает, сколько записей в журнале теперь.
\echo 'Пишу в журнал visits...'
INSERT INTO visits (note) VALUES ('psql работает');
SELECT count(*) AS visits_total FROM visits;
`

// pgsLab2Lines prints a file the way the file checks read it: every line
// trimmed, blank lines dropped. `psql -t` pads values with spaces and ends with
// an empty line; that is still "one name per line".
func pgsLab2Lines(file string) string {
	return `sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' ` + file + ` 2>/dev/null | grep -v '^$'`
}

// pgsLab2FileNow tells how a file looks now, for a "Сейчас" diagnostic: missing,
// or how many lines it has. It has no double quotes on purpose — check()
// escapes them in the message, and an escaped quote inside $(…) breaks it.
func pgsLab2FileNow(file string) string {
	return `$([ -s ` + file + ` ] || echo файла ` + file + ` нет или он пустой)` +
		`$([ -s ` + file + ` ] && echo в ` + file + ` непустых строк: $(` + pgsLab2Lines(file) + ` | grep -c .))`
}

// pgsLab2NotDB filters a file's lines down to those that are not the name of
// any database on the server: a header, a frame, "(6 rows)", a stray column.
var pgsLab2NotDB = pgsLab2Lines("/root/dbs.txt") + ` | grep -vxF -f <(glpg q postgres ` + b64("SELECT datname FROM pg_database") + `)`

// pgsLab2TitlesRef is what /root/titles.txt must hold, line for line.
const pgsLab2TitlesRef = "SELECT title FROM books ORDER BY title"

// pgsLab2Titles compares /root/titles.txt with the reference the way pgOutfile
// does, and prints what differs.
var pgsLab2Titles = "glpg outfile /root/titles.txt pereplet " + b64(pgsLab2TitlesRef)

// pgsLab2Psql is what the ~/.psqlrc task is judged by: plain psql, reading the
// student's startup file, as the task text tells them to run it.
const pgsLab2Psql = `HOME=/root psql -d pereplet -Atc 'SELECT NULL' 2>&1`

func init() {
	pgStartLabs["ch-pgs-lab2"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab2Pereplet) +
			pgFresh("archive") + pgRun("archive", pgsLab2Archive) +
			"mkdir -p /root/sql\n" +
			"echo " + b64(pgsLab2Hello) + " | base64 -d > /root/sql/hello.sql\n"),
		Checks: map[int]string{
			// The databases setup made must all be there (the student may list
			// sandbox, or the templates, too), and nothing that is not a name.
			1: check(`[ "$(`+pgsLab2Lines("/root/dbs.txt")+` | sort -u | grep -cxE 'postgres|root|pereplet|archive')" = 4 ] && [ -z "$(`+pgsLab2NotDB+`)" ]`,
				"в /root/dbs.txt имена всех баз сервера, по одному в строке, и ничего кроме имён",
				"Сейчас "+pgsLab2FileNow("/root/dbs.txt")+"$("+pgsLab2NotDB+" | grep -c . | sed -e '/^0$/d' -e 's/.*/, из них не имена баз: &/')$("+pgsLab2NotDB+" | head -n1 | sed 's/.*/, например «&»/'); баз на сервере, не считая шаблонов: "+pgVal("postgres", "SELECT count(*) FROM pg_database WHERE NOT datistemplate")),
			2: check(pgTrue("pereplet", "SELECT EXISTS (SELECT 1 FROM visits WHERE note = 'psql работает')"),
				"скрипт /root/sql/hello.sql выполнен в базе pereplet: в visits есть запись «psql работает»",
				"Сейчас записей в таблице visits базы pereplet: "+pgVal("pereplet", "SELECT count(*) FROM visits")),
			3: check(`[ -s /root/authors_count.txt ] && [ "$(`+pgsLab2Lines("/root/authors_count.txt")+`)" = "`+pgVal("pereplet", "SELECT count(*) FROM authors")+`" ]`,
				"в /root/authors_count.txt одно число — сколько авторов в таблице authors",
				"Сейчас "+pgsLab2FileNow("/root/authors_count.txt")+"$("+pgsLab2Lines("/root/authors_count.txt")+" | head -n1 | sed 's/.*/, первая — «&»/')"),
			// pgOutfile's comparison, with two more hints than it gives: the two
			// near misses of this task are a forgotten ORDER BY and psql -t
			// without -A, and to glpg both are just "содержимое отличается".
			4: check(pgsLab2Titles+" >/dev/null 2>&1",
				"в /root/titles.txt названия всех книг по алфавиту, по одному в строке, без заголовка и рамок",
				"Сейчас: $("+pgsLab2Titles+")"+
					"$(grep -q '^[[:space:]]' /root/titles.txt 2>/dev/null && echo '; в начале строк пробелы — так psql выравнивает столбец')"+
					"$(cmp -s <("+pgsLab2Lines("/root/titles.txt")+" | LC_ALL=C sort) <(glpg q pereplet "+b64("SELECT title FROM books")+" | LC_ALL=C sort)"+
					" && ! cmp -s <("+pgsLab2Lines("/root/titles.txt")+") <(glpg q pereplet "+b64(pgsLab2TitlesRef)+")"+
					" && echo '; названия те, но порядок не алфавитный')"),
			5: check(pgIs("postgres", "SELECT datconnlimit FROM pg_database WHERE datname = 'sandbox'", "5"),
				"база sandbox создана, к ней разрешено не больше 5 подключений",
				"Сейчас: "+pgVal("postgres", "SELECT coalesce((SELECT 'база sandbox есть, лимит подключений — ' || CASE WHEN datconnlimit = -1 THEN 'не задан' ELSE datconnlimit::text END FROM pg_database WHERE datname = 'sandbox'), "+
					"(SELECT 'базы sandbox нет, но есть ' || datname || ' — имя должно быть строчными буквами' FROM pg_database WHERE lower(datname) = 'sandbox' LIMIT 1), 'базы sandbox нет')")),
			// Behaviour, not the file: ~/.psqlrc-16 or the system psqlrc work too.
			// "?column?|∅" is the same NULL with \x on in the student's file.
			6: check(`o=$(`+pgsLab2Psql+`); printf '%s\n' "$o" | grep -qE '(^|\|)∅[[:space:]]*$' && printf '%s\n' "$o" | grep -qE '^Time: [0-9]'`,
				"psql сам включает замер времени и показывает NULL как ∅",
				"Сейчас psql -d pereplet -Atc 'SELECT NULL' выводит: $("+pgsLab2Psql+" | head -n 6 | sed -e 's/^$/(пустая строка)/' -e 's/.*/«&»/' | tr '\\n' ' ')"),
		},
	}
}
