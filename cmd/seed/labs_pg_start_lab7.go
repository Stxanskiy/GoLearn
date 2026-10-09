package main

// Lab 7 of pg-start: importing the catalog (ch-pgs-lab7).
//
// Lesson 16 teaches \copy: loading a CSV file with a header, a file from the
// Russian Excel (DELIMITER ';'), ids that come with the file (integer PRIMARY
// KEY, no serial), reading CONTEXT when a line breaks the load, and exporting a
// query's result. Here the student puts it to work on state S6 and ends at S7:
// genres and book_genres are created and loaded from two files the setup puts
// in /root, and a report for the manager is exported to /root/in_stock.csv.
//
// The catalog runs through labs 4-7 as one story, so the setup is lab 6's
// setup with every reference solution of lab 6 applied: eight authors and
// fourteen books, inserted without ids in lab 6's order, each book with its
// author and «Идиот» at its price again; books_id_seq goes on past lab 5's
// deleted duplicate (id 15); Анна Смирнова's first order, two copies of
// «Евгений Онегин» and one of «Анна Каренина». Nothing else changed while the
// student was away: what is new are the manager's two files in /root, which
// the introduction names.
//
//   - genres.csv comes from the Russian Excel: ';' between the values, ids 1–6.
//     Loaded the lesson's way, without DELIMITER, it stops on line 2 with
//     invalid input syntax for type integer: "1;роман";
//   - book_genres.csv has ',' and gives every book of the catalog its genres,
//     by the books' real ids, with one pair written twice (line 12 repeats
//     line 7, «Идиот» — роман), so the load stops on the primary key with
//     CONTEXT … line 12 and nothing is loaded. The student removes the repeat
//     in nano and loads again. The check looks at the table, never at the
//     file the student edits.
//
// The quizzes have the student read line N in CONTEXT, see that a failed \copy
// leaves the table empty, and say why genres.id has no serial.
//
// The table checks test what a table does, not how it is written: each rule
// inserts rows in a transaction that is rolled back (pgRejects/pgAccepts),
// with ids from 900001 up and names «проверка …» that no file has, so the
// rules hold the same before and after the data is loaded and the tasks can be
// done in any order. When a rule fails, the student is told what the database
// does now — "a genre without an id is accepted" — not which keyword to type.
// Tables that behave alike pass alike: id integer NOT NULL UNIQUE as well as
// PRIMARY KEY, UNIQUE (book_id, genre_id) on two NOT NULL columns as well as
// the pair as primary key.
//
// The report is compared with what psql itself exports from the database at
// check time (COPY … TO STDOUT with the options the task asks for), so any way
// that writes the same file passes: \copy with the new or the old option
// syntax, \o with \pset format csv. Eight books of the catalog are in stock,
// at eight different prices, so ORDER BY price has one order.
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.

import (
	"sort"
	"strconv"
	"strings"
)

func init() {
	pgStartLabs["ch-pgs-lab7"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab7Schema) + pgsLab7Files),
		Checks: map[int]string{
			1: pgsLab7Table("genres", []string{"id integer", "name text"}, pgsLab7GenresRules,
				"в базе pereplet есть таблица genres: id — целое число, первичный ключ, номер из файла (без serial); name — текст, обязательный, без повторов"),
			2: check(pgTrue("pereplet", pgsLab7GenresLoaded),
				"в genres загружены жанры из /root/genres.csv — все "+strconv.Itoa(len(pgsLab7Genres))+", с номерами из файла",
				"Сейчас: "+pgsLab7Rows("genres", "в базе pereplet нет таблицы genres — сначала создай её (первое задание)",
					pgsLab7GenresState, "не получилось сравнить таблицу genres с файлом — проверь её столбцы: id integer и name text (первое задание)")),
			3: pgsLab7Table("book_genres", []string{"book_id integer", "genre_id integer"}, pgsLab7LinksRules,
				"в базе pereplet есть таблица book_genres: book_id и genre_id — обязательные ссылки на books и genres, первичный ключ — пара столбцов"),
			4: check(pgTrue("pereplet", pgsLab7LinksLoaded),
				"в book_genres загружены пары из /root/book_genres.csv — все "+strconv.Itoa(len(pgsLab7Pairs))+", каждая по одному разу",
				"Сейчас: "+pgsLab7Rows("book_genres", "в базе pereplet нет таблицы book_genres — сначала создай её (задание «Таблица связей»)",
					pgsLab7LinksState, "не получилось сравнить таблицу book_genres с файлом — проверь её столбцы: book_id integer и genre_id integer (задание «Таблица связей»)")),
			5: check(pgsLab7Run(pgsLab7Report)+" >/dev/null",
				"в /root/in_stock.csv — книги в наличии: заголовок title;price, от дешёвых к дорогим, через точку с запятой",
				"Сейчас: $("+pgsLab7Run(pgsLab7Report)+")"),
		},
	}
}

// pgsLab7Schema is state S6 as lab 6 leaves it — its setup with every
// reference solution applied: the catalog with each book's author, the rules,
// and the first order. Rows go in without ids in lab 6's order, so the ids are
// the ones the student has; setval runs inside a DO block, which prints
// nothing.
const pgsLab7Schema = `CREATE TABLE authors (
    id         serial PRIMARY KEY,
    full_name  text NOT NULL,
    birth_year integer,
    country    text
);
CREATE TABLE books (
    id             serial PRIMARY KEY,
    title          text NOT NULL,
    price          numeric(8,2) NOT NULL CHECK (price > 0),
    page_count     integer,
    in_stock       boolean NOT NULL DEFAULT true,
    published_year integer,
    author_id      integer REFERENCES authors (id)
);
CREATE TABLE customers (
    id            serial PRIMARY KEY,
    email         text NOT NULL UNIQUE,
    full_name     text NOT NULL,
    city          text,
    registered_at date NOT NULL DEFAULT current_date
);
CREATE TABLE orders (
    id          serial PRIMARY KEY,
    customer_id integer NOT NULL REFERENCES customers (id),
    created_at  date NOT NULL DEFAULT current_date,
    status      text NOT NULL DEFAULT 'new'
                CHECK (status IN ('new', 'paid', 'shipped', 'cancelled'))
);
CREATE TABLE order_items (
    order_id integer NOT NULL REFERENCES orders (id),
    book_id  integer NOT NULL REFERENCES books (id),
    quantity integer NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, book_id)
);
INSERT INTO authors (full_name, birth_year, country) VALUES
    ('Иван Крылов', 1769, 'Россия'),
    ('Александр Пушкин', 1799, 'Россия'),
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Жюль Верн', 1828, 'Франция'),
    ('Антон Чехов', 1860, 'Россия'),
    ('Лев Толстой', 1828, 'Россия'),
    ('Марк Твен', 1835, 'США'),
    ('Артур Конан Дойл', 1859, 'Великобритания');
INSERT INTO books (title, price, page_count, in_stock, published_year, author_id) VALUES
    ('Басни', 240.00, 224, true, 1809, 1),
    ('Евгений Онегин', 460.00, 320, true, 1833, 2),
    ('Капитанская дочка', 210.00, 192, false, 1836, 2),
    ('Повести Белкина', 190.00, 160, false, 1831, 2),
    ('Идиот', 715.00, 640, true, 1869, 3),
    ('Братья Карамазовы', 990.00, 992, false, 1880, 3),
    ('Белые ночи', 150.00, 96, false, 1848, 3),
    ('Таинственный остров', 1350.00, 704, true, 1875, 4),
    ('Вокруг света за восемьдесят дней', 360.00, 320, false, 1872, 4),
    ('Дети капитана Гранта', 560.00, 672, true, 1868, 4),
    ('Собака Баскервилей', 450.00, 256, true, 1902, 8),
    ('Приключения Тома Сойера', 380.00, 288, false, 1876, 7),
    ('Анна Каренина', 870.00, 864, true, 1878, 6),
    ('Затерянный мир', 430.00, 320, true, 1912, 8);
DO $$ BEGIN PERFORM setval('books_id_seq', 15); END $$;
INSERT INTO customers (email, full_name, city) VALUES
    ('anna.smirnova@example.com', 'Анна Смирнова', 'Москва');
INSERT INTO orders (customer_id) VALUES (1);
INSERT INTO order_items (order_id, book_id, quantity) VALUES
    (1, 2, 2),
    (1, 13, 1);
`

// pgsLab7Genres is genres.csv: a genre's id is its place in the list, from 1.
var pgsLab7Genres = []string{"роман", "повесть", "поэзия", "приключения", "детектив", "фантастика"}

// pgsLab7Pairs is book_genres.csv without its repeat: (book_id, genre_id), in
// file order, by the ids books has after lab 6. Every book of the catalog has
// a genre and every genre a book.
var pgsLab7Pairs = [][2]int{
	{1, 3},         // Басни: поэзия
	{2, 1}, {2, 3}, // Евгений Онегин: роман, поэзия
	{3, 2},         // Капитанская дочка: повесть
	{4, 2},         // Повести Белкина: повесть
	{5, 1},         // Идиот: роман — the pair written twice
	{6, 1},         // Братья Карамазовы: роман
	{7, 2},         // Белые ночи: повесть
	{8, 1}, {8, 4}, // Таинственный остров: роман, приключения
	{9, 1}, {9, 4}, // Вокруг света за восемьдесят дней: роман, приключения
	{10, 1}, {10, 4}, // Дети капитана Гранта: роман, приключения
	{11, 2}, {11, 5}, // Собака Баскервилей: повесть, детектив
	{12, 1}, {12, 4}, // Приключения Тома Сойера: роман, приключения
	{13, 1},                   // Анна Каренина: роман
	{14, 1}, {14, 4}, {14, 6}, // Затерянный мир: роман, приключения, фантастика
}

// The repeat: pgsLab7Pairs[pgsLab7Twice] is written again after the first
// pgsLab7RepeatAfter pairs, between two books' lines. With the header as line
// 1 that is line 12 of the file repeating line 7, as the quiz on CONTEXT says.
const (
	pgsLab7Twice       = 5
	pgsLab7RepeatAfter = 10
)

// pgsLab7GenresCSV is the file from the Russian Excel: ';' between values.
var pgsLab7GenresCSV = func() string {
	var b strings.Builder
	b.WriteString("id;name\n")
	for i, g := range pgsLab7Genres {
		b.WriteString(strconv.Itoa(i+1) + ";" + g + "\n")
	}
	return b.String()
}()

// pgsLab7LinksCSV is book_genres.csv: ',' between values and one pair twice.
var pgsLab7LinksCSV = func() string {
	line := func(p [2]int) string { return strconv.Itoa(p[0]) + "," + strconv.Itoa(p[1]) + "\n" }
	var b strings.Builder
	b.WriteString("book_id,genre_id\n")
	for i, p := range pgsLab7Pairs {
		if i == pgsLab7RepeatAfter {
			b.WriteString(line(pgsLab7Pairs[pgsLab7Twice]))
		}
		b.WriteString(line(p))
	}
	return b.String()
}()

// pgsLab7Files puts the manager's files in /root, owned by root and mode 644
// as a saved file would be (docker exec runs the setup with umask 0), and
// removes a report left from an earlier attempt, so the last check starts red.
var pgsLab7Files = "echo " + b64(pgsLab7GenresCSV) + " | base64 -d > /root/genres.csv\n" +
	"echo " + b64(pgsLab7LinksCSV) + " | base64 -d > /root/book_genres.csv\n" +
	"chmod 644 /root/genres.csv /root/book_genres.csv\n" +
	"rm -f /root/in_stock.csv\n"

// pgsLab7GenresFile is genres.csv as a table, f(id, name).
var pgsLab7GenresFile = func() string {
	rows := make([]string, len(pgsLab7Genres))
	for i, g := range pgsLab7Genres {
		rows[i] = "(" + strconv.Itoa(i+1) + ", '" + g + "')"
	}
	return "(VALUES " + strings.Join(rows, ", ") + ") AS f(id, name)"
}()

// pgsLab7LinksFile is the distinct pairs of book_genres.csv as a table,
// f(book_id, genre_id).
var pgsLab7LinksFile = func() string {
	rows := make([]string, len(pgsLab7Pairs))
	for i, p := range pgsLab7Pairs {
		rows[i] = "(" + strconv.Itoa(p[0]) + ", " + strconv.Itoa(p[1]) + ")"
	}
	return "(VALUES " + strings.Join(rows, ", ") + ") AS f(book_id, genre_id)"
}()

// pgsLab7Tx is SQL text: a note when the student's psql sits in an open
// transaction — lesson 10 taught the BEGIN habit, and a table created or
// loaded inside one is invisible to the check until COMMIT — or an empty
// string.
const pgsLab7Tx = "(SELECT CASE " +
	"WHEN bool_or(state = 'idle in transaction (aborted)') " +
	"THEN '. В твоей транзакции была ошибка (приглашение pereplet=!#): набери ROLLBACK; и повтори команды' " +
	"WHEN bool_or(state = 'idle in transaction') " +
	"THEN '. У тебя открыта транзакция (приглашение pereplet=*#): проверка не видит её изменений — сохрани их командой COMMIT;' " +
	"ELSE '' END FROM pg_stat_activity WHERE datname = 'pereplet' AND pid <> pg_backend_pid())"

// pgsLab7Rows is a substitution describing the rows of table for a data
// check: missing when there is no such table, else what sql says about it,
// else fallback (the table has other columns, or the student's open
// transaction holds it) — and the open-transaction note. Whether the table
// exists is asked on its own: a query naming a table that is not there fails
// as a whole, whatever its CASE would have said.
func pgsLab7Rows(table, missing, sql, fallback string) string {
	return "$(" + pgsLab7Run(
		"if [ \"$(glpg q pereplet "+b64("SELECT to_regclass('"+table+"') IS NULL")+")\" = t ]; then\n"+
			"\tprintf '%s' "+pgsLab7Sh(missing)+"\n"+
			"elif s=$(glpg q pereplet "+b64(sql)+"); then\n"+
			"\tprintf '%s' \"$s\"\n"+
			"else\n"+
			"\tprintf '%s' "+pgsLab7Sh(fallback)+"\n"+
			"fi\n"+
			"glpg q pereplet "+b64("SELECT "+pgsLab7Tx)+"\n") + ")"
}

// pgsLab7Run is a command running a generated bash script in the sandbox. The
// script travels base64-encoded, so it may quote as it likes and still sit in
// a one-line check, inside the double quotes of a diagnostic too.
func pgsLab7Run(script string) string {
	return "echo " + b64(script) + " | base64 -d | bash 2>/dev/null"
}

// pgsLab7Sh is s quoted for bash.
func pgsLab7Sh(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// pgsLab7Rule is one thing a table has to do: running sql (rolled back) ends
// with SQLSTATE want. say is what the student is told when it does not — what
// the database does now; by overrides it for particular SQLSTATEs.
type pgsLab7Rule struct {
	sql, want, say string
	by             map[string]string
}

func (r pgsLab7Rule) cond() string { return pgRejects("pereplet", r.sql, r.want) }

// pgsLab7TwoGenres is the first statement of the book_genres rules: two genres
// no file has, for the pairs the rules try.
const pgsLab7TwoGenres = "INSERT INTO genres (id, name) VALUES (900001, 'проверка 1'), (900002, 'проверка 2'); "

// pgsLab7GenresRules: genres takes a genre with its number from the file, and
// refuses one without a number, a repeated number, a genre without a name and
// a repeated name. Lesson 16 makes id integer PRIMARY KEY with no serial,
// lesson 14 gives NOT NULL and UNIQUE.
var pgsLab7GenresRules = []pgsLab7Rule{
	{sql: "INSERT INTO genres (id, name) VALUES (900001, 'проверка 1'), (900002, 'проверка 2')", want: "00000",
		say: "жанр с номером и названием, как в файле, в genres не добавляется — посмотри на таблицу: \\d genres"},
	{sql: "INSERT INTO genres (name) VALUES ('проверка 1')", want: "23502",
		say: "жанр без номера id база принимает — а номер жанра должен браться только из файла: id без serial, и пустым он быть не может"},
	{sql: "INSERT INTO genres (id, name) VALUES (900001, 'проверка 1'), (900001, 'проверка 2')", want: "23505",
		say: "два жанра с одним номером id база принимает — id должен быть первичным ключом"},
	{sql: "INSERT INTO genres (id, name) VALUES (900001, NULL)", want: "23502",
		say: "жанр без названия (пустой name) база принимает — название обязательно"},
	{sql: "INSERT INTO genres (id, name) VALUES (900001, 'проверка'), (900002, 'проверка')", want: "23505",
		say: "два жанра с одинаковым названием база принимает — названия не должны повторяться"},
}

// pgsLab7LinksRules: book_genres takes several genres for a book and several
// books for a genre, and refuses a pair twice, a book or genre that does not
// exist and an empty book or genre. The book is the last of books (max(id)),
// which no author has an id of, so a book_id that points to authors fails too.
var pgsLab7LinksRules = []pgsLab7Rule{
	{sql: pgsLab7TwoGenres +
		"INSERT INTO book_genres (book_id, genre_id) SELECT max(id), 900001 FROM books; " +
		"INSERT INTO book_genres (book_id, genre_id) SELECT max(id), 900002 FROM books; " +
		"INSERT INTO book_genres (book_id, genre_id) SELECT min(id), 900001 FROM books", want: "00000",
		say: "правильная пара «книга — жанр» в book_genres не добавляется — посмотри на таблицу: \\d book_genres",
		by: map[string]string{
			"23505": "одной книге не получается записать два жанра или одному жанру — две книги: первичный ключ должен быть из двух столбцов сразу, book_id и genre_id",
			"23503": "пару с книгой из books и жанром из genres база не принимает — проверь, куда ведут ссылки: book_id на books (id), genre_id на genres (id)",
		}},
	{sql: pgsLab7TwoGenres +
		"INSERT INTO book_genres (book_id, genre_id) SELECT max(id), 900001 FROM books; " +
		"INSERT INTO book_genres (book_id, genre_id) SELECT max(id), 900001 FROM books", want: "23505",
		say: "одну и ту же пару «книга — жанр» база принимает дважды — нужен первичный ключ из пары столбцов"},
	{sql: pgsLab7TwoGenres + "INSERT INTO book_genres (book_id, genre_id) VALUES (900009, 900001)", want: "23503",
		say: "пару с книгой, которой нет в books, база принимает — book_id должен ссылаться на books (id)"},
	{sql: "INSERT INTO book_genres (book_id, genre_id) SELECT max(id), 900009 FROM books", want: "23503",
		say: "пару с жанром, которого нет в genres, база принимает — genre_id должен ссылаться на genres (id)"},
	{sql: pgsLab7TwoGenres + "INSERT INTO book_genres (book_id, genre_id) VALUES (NULL, 900001)", want: "23502",
		say: "пару без книги (пустой book_id) база принимает — book_id обязателен"},
	{sql: "INSERT INTO book_genres (book_id, genre_id) SELECT max(id), NULL::integer FROM books", want: "23502",
		say: "пару без жанра (пустой genre_id) база принимает — genre_id обязателен"},
}

// pgsLab7Table is the check of a table the student creates: its columns are
// exactly cols ("name type", in any order) and it does what rules say. The
// diagnostic shows the columns and, when they are right, every rule that
// fails, in the student's words.
func pgsLab7Table(table string, cols []string, rules []pgsLab7Rule, good string) string {
	conds := []string{pgTrue("pereplet", pgsLab7Shape(table, cols))}
	for _, r := range rules {
		conds = append(conds, r.cond())
	}
	return check(strings.Join(conds, " && "), good, "Сейчас: $("+pgsLab7Run(pgsLab7TableDiag(table, cols, rules))+")")
}

// pgsLab7Shape is SQL true when table's columns are exactly cols.
func pgsLab7Shape(table string, cols []string) string {
	want := append([]string(nil), cols...)
	sort.Strings(want) // byte order, as attname sorts
	return "SELECT coalesce((SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod), ',' ORDER BY a.attname) " +
		"FROM pg_attribute a WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped), '') = '" +
		strings.Join(want, ",") + "'"
}

// pgsLab7Tables is SQL for the student's tables in pereplet, as \dt lists them.
const pgsLab7Tables = "coalesce((SELECT string_agg(tablename, ', ' ORDER BY tablename) FROM pg_tables " +
	"WHERE schemaname = 'public'), 'ни одной')"

// pgsLab7Columns is SQL describing table for a "Сейчас:" line: that it is
// missing, or its columns as \d shows them and which differ from cols —
// missing, extra or of another type, never what the type should be.
func pgsLab7Columns(table string, cols []string) string {
	vals := make([]string, len(cols))
	for i, c := range cols {
		name, typ, _ := strings.Cut(c, " ")
		vals[i] = "('" + name + "', '" + typ + "', " + strconv.Itoa(i+1) + ")"
	}
	w := "(VALUES " + strings.Join(vals, ", ") + ") w(name, typ, n)"
	h := "(SELECT a.attname::text AS name, format_type(a.atttypid, a.atttypmod) AS typ, a.attnum AS n FROM pg_attribute a " +
		"WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped) h"
	return "SELECT CASE WHEN to_regclass('" + table + "') IS NULL " +
		"THEN 'в базе pereplet нет таблицы " + table + ", таблицы в базе: ' || " + pgsLab7Tables + " " +
		"ELSE concat_ws('; ', 'столбцы " + table + ": ' || " +
		"(SELECT string_agg(h.name || ' ' || h.typ, ', ' ORDER BY h.n) FROM " + h + "), " +
		"(SELECT 'не хватает столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " WHERE w.name NOT IN (SELECT h.name FROM " + h + ")), " +
		"(SELECT 'лишние столбцы: ' || string_agg(h.name, ', ' ORDER BY h.n) FROM " + h + " WHERE h.name NOT IN (SELECT w.name FROM " + w + ")), " +
		"(SELECT 'не тот тип у столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " JOIN " + h + " ON h.name = w.name WHERE h.typ <> w.typ)" +
		") END"
}

// pgsLab7TableDiag is the bash script of a table check's diagnostic. A rule
// that cannot get at the table (55P03: an open transaction of the student's
// holds it) stops the rest, which would each wait out the lock timeout too.
// So does the first rule, the rows that have to go in: when they do not, every
// later rule fails on the same rows, whatever it tests — a book_id that points
// at authors would also read as "a pair is taken twice" — and only the first
// one says what is wrong.
func pgsLab7TableDiag(table string, cols []string, rules []pgsLab7Rule) string {
	var b strings.Builder
	b.WriteString("s=$(glpg q pereplet " + b64(pgsLab7Columns(table, cols)) + ") || " +
		"{ echo 'не получается подключиться к базе pereplet (её удалили?) — нажми «Пересоздать»'; exit; }\n")
	b.WriteString("printf '%s' \"$s\"\n")
	b.WriteString("stop=\n")
	b.WriteString("if [ \"$(glpg q pereplet " + b64(pgsLab7Shape(table, cols)) + ")\" = t ]; then\n")
	for i, r := range rules {
		gate := ""
		if i == 0 {
			gate = "stop=1; "
		}
		b.WriteString("\t[ -z \"$stop\" ] && st=$(glpg state pereplet " + b64(r.sql) + ") && [ \"$st\" != " + r.want + " ] &&\n\t\t{ " + gate + "case $st in\n")
		b.WriteString("\t\t55P03) stop=1; printf '; %s' " + pgsLab7Sh("проверка не дождалась таблицы "+table+": её держит незакрытая транзакция") + " ;;\n")
		states := make([]string, 0, len(r.by))
		for st := range r.by {
			states = append(states, st)
		}
		sort.Strings(states)
		for _, st := range states {
			b.WriteString("\t\t" + st + ") printf '; %s' " + pgsLab7Sh(r.by[st]) + " ;;\n")
		}
		b.WriteString("\t\t*) printf '; %s' " + pgsLab7Sh(r.say) + " ;;\n\t\tesac; }\n")
	}
	b.WriteString("fi\n")
	b.WriteString("if [ \"$(glpg q postgres " + b64("SELECT to_regclass('"+table+"') IS NOT NULL") + ")\" = t ]; then\n" +
		"\tprintf '; %s' " + pgsLab7Sh("таблица "+table+" нашлась в базе postgres — проверь приглашение: работать нужно в pereplet=#") + "\n" +
		"fi\n")
	b.WriteString("glpg q pereplet " + b64("SELECT "+pgsLab7Tx) + "\n")
	return b.String()
}

// pgsLab7GenresLoaded is SQL true when genres holds exactly the rows of
// genres.csv.
var pgsLab7GenresLoaded = "SELECT (SELECT count(*) FROM genres) = " + strconv.Itoa(len(pgsLab7Genres)) + " " +
	"AND NOT EXISTS (SELECT id, name FROM " + pgsLab7GenresFile + " EXCEPT SELECT id, name FROM genres)"

// pgsLab7GenresState says what genres holds compared with the file: a genre
// held twice (a table without its key, loaded twice), genres missing or not
// from it. The file is the student's to read, so its rows may be named.
var pgsLab7GenresState = "SELECT CASE " +
	"WHEN NOT EXISTS (SELECT 1 FROM genres) THEN 'таблица genres пустая. Если \\copy ответил ERROR, в таблицу не попало ничего: " +
	"прочитай ошибку, исправь команду и выполни её снова' " +
	"ELSE concat_ws('; ', 'в genres строк: ' || (SELECT count(*) FROM genres) || ', а жанров в файле " + strconv.Itoa(len(pgsLab7Genres)) + "', " +
	"(SELECT 'жанры, записанные дважды: ' || string_agg(d.id || ';' || d.name, ', ' ORDER BY d.id) || " +
	"' — у таблицы нет первичного ключа на id (первое задание)' " +
	"FROM (SELECT id, name FROM genres GROUP BY id, name HAVING count(*) > 1) d), " +
	"(SELECT 'нет жанров из файла: ' || string_agg(f.id || ';' || f.name, ', ' ORDER BY f.id) FROM " + pgsLab7GenresFile + " " +
	"WHERE NOT EXISTS (SELECT 1 FROM genres g WHERE g.id = f.id AND g.name = f.name)), " +
	"(SELECT 'строки не из файла: ' || string_agg(coalesce(g.id::text, 'NULL') || ';' || coalesce(g.name, 'NULL'), ', ' ORDER BY g.id) FROM genres g " +
	"WHERE NOT EXISTS (SELECT 1 FROM " + pgsLab7GenresFile + " WHERE f.id = g.id AND f.name = g.name))) END"

// pgsLab7LinksLoaded is SQL true when book_genres holds each pair of
// book_genres.csv once and nothing else.
var pgsLab7LinksLoaded = "SELECT (SELECT count(*) FROM book_genres) = " + strconv.Itoa(len(pgsLab7Pairs)) + " " +
	"AND NOT EXISTS (SELECT book_id, genre_id FROM " + pgsLab7LinksFile + " EXCEPT SELECT book_id, genre_id FROM book_genres)"

// pgsLab7Pair is SQL text for a pair the way the file writes it: 5,1.
func pgsLab7Pair(t string) string {
	return "coalesce(" + t + ".book_id::text, 'NULL') || ',' || coalesce(" + t + ".genre_id::text, 'NULL')"
}

// pgsLab7LinksState says what book_genres holds compared with the file: a
// pair held twice (a table without its key), pairs missing or not from it.
var pgsLab7LinksState = "SELECT CASE " +
	"WHEN NOT EXISTS (SELECT 1 FROM book_genres) THEN 'таблица book_genres пустая. Если \\copy ответил ERROR, в таблицу не попало ничего: " +
	"найди в файле строку из CONTEXT, исправь файл и загрузи его снова' " +
	"ELSE concat_ws('; ', 'в book_genres строк: ' || (SELECT count(*) FROM book_genres) || ', а разных пар в файле " + strconv.Itoa(len(pgsLab7Pairs)) + "', " +
	"(SELECT 'пары, записанные дважды: ' || string_agg(" + pgsLab7Pair("d") + ", ', ' ORDER BY d.book_id, d.genre_id) || " +
	"' — у таблицы нет первичного ключа из пары столбцов (задание «Таблица связей»)' " +
	"FROM (SELECT book_id, genre_id FROM book_genres GROUP BY book_id, genre_id HAVING count(*) > 1) d), " +
	"(SELECT 'нет пар из файла: ' || string_agg(" + pgsLab7Pair("f") + ", ', ' ORDER BY f.book_id, f.genre_id) FROM " + pgsLab7LinksFile + " " +
	"WHERE NOT EXISTS (SELECT 1 FROM book_genres b WHERE b.book_id = f.book_id AND b.genre_id = f.genre_id)), " +
	"(SELECT 'пары не из файла: ' || string_agg(" + pgsLab7Pair("b") + ", ', ' ORDER BY b.book_id, b.genre_id) FROM book_genres b " +
	"WHERE NOT EXISTS (SELECT 1 FROM " + pgsLab7LinksFile + " WHERE f.book_id = b.book_id AND f.genre_id = b.genre_id))) END"

// pgsLab7ReportSQL is the report as psql itself writes it from the database:
// the reference the student's file is compared with.
const pgsLab7ReportSQL = "COPY (SELECT title, price FROM books WHERE in_stock = true ORDER BY price) " +
	"TO STDOUT WITH (FORMAT csv, HEADER, DELIMITER ';')"

// pgsLab7Report is the bash script of the report check: silent with exit 0
// when /root/in_stock.csv is what psql exports, otherwise it says what is
// wrong with the file — the delimiter, the header, how many books, their
// order — and exits 1. It never prints the rows the file should hold. An
// empty file is what a \copy … TO that failed leaves: psql opens the file
// before the server reads the query, so a stray ';' inside the parentheses
// empties a report that was right a minute ago.
var pgsLab7Report = `f=/root/in_stock.csv
if [ ! -e "$f" ]; then
	echo "файла $f нет: выгрузи отчёт командой \\copy … TO '$f', а если уже выгружал — проверь путь после TO"
	exit 1
fi
if [ ! -s "$f" ]; then
	echo "файл $f пустой: так бывает, когда \\copy … TO ответил ERROR — файл он уже открыл, а записать ничего не успел. Исправь команду по ошибке и выполни её снова"
	exit 1
fi
want=$(glpg q pereplet ` + b64(pgsLab7ReportSQL) + `)
wn=$(( $(printf '%s\n' "$want" | wc -l) - 1 ))
if [ "$wn" -lt 1 ]; then
	echo "не получилось выгрузить из базы pereplet книги в наличии — таблица books изменилась? Нажми «Пересоздать»"
	exit 1
fi
got=$(cat "$f")
[ "$got" = "$want" ] && exit 0
if grep -q $'\r' "$f"; then
	echo "в конце строк файла стоят невидимые символы \\r, как в файлах из Windows — выгрузи файл заново командой \\copy"
	exit 1
fi
h=$(head -n1 "$f" | head -c 120)
wh=$(printf '%s\n' "$want" | head -n1)
gn=$(( $(printf '%s\n' "$got" | wc -l) - 1 ))
if [ "$h" = "${wh//;/,}" ]; then
	echo "значения в файле разделены запятыми, а русский Excel ждёт точку с запятой — укажи разделитель: DELIMITER ';'"
elif [ "$h" = "$(printf '%s\n' "$want" | sed -n 2p)" ]; then
	echo "в первой строке файла сразу книга, а должен быть заголовок $wh — добавь HEADER"
elif printf '%s' "$h" | grep -q '|'; then
	echo "в файле таблица с рамкой из | — так psql показывает результат на экране; файл CSV выгружают командой \\copy … TO … WITH (FORMAT csv, …)"
elif [ "$h" != "$wh" ]; then
	echo "первая строка файла — «$h», а должен быть заголовок $wh: столбцы title и price, в таком порядке, через точку с запятой"
elif [ "$gn" != "$wn" ]; then
	echo "книг в файле: $gn, а в наличии на складе: $wn — проверь условие WHERE"
elif [ "$(printf '%s\n' "$got" | LC_ALL=C sort)" = "$(printf '%s\n' "$want" | LC_ALL=C sort)" ]; then
	echo "книги те, но порядок не тот: нужно от дешёвых к дорогим — проверь ORDER BY"
else
	echo "книг в файле столько, сколько нужно ($gn), но строки другие — проверь WHERE и столбцы: название и цена"
fi
exit 1
`
