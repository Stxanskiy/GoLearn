package main

// Lab 6 of pg-start: linking the tables (ch-pgs-lab6).
//
// Lessons 13 and 14 teach foreign keys and the rules NOT NULL, DEFAULT, UNIQUE
// and CHECK; here the student takes pereplet from state S3 to S6 of the course
// plan. The setup is the catalog the way lab 5 leaves it — five authors, nine
// books with ids 1-9, the sequence past the duplicate lab 5 deleted — with one
// trap: «Идиот», repriced in lab 5, costs 0.00 ("someone zeroed it while
// repricing"), so ADD CHECK (price > 0) fails until the student finds the row
// and gives it back its price — the order lesson 14 teaches. books has no
// author_id yet; the task gives the book → author table, the ids the student
// looks up with SELECT.
//
// Tasks: author_id with a foreign key, filled in for every book; the catalog's
// rules; customers, orders and order_items as in S6; and the first order — a
// customer, her order and two items, in that order, each row pointing at the
// one before.
//
// The checks look at what the tables do, never at what their rules are called:
//   - a rule is tested by trying to break it in a transaction that is rolled
//     back (pgRejects: 23502 not null, 23503 foreign key, 23505 unique, 23514
//     check), a default by inserting a row without the column and reading the
//     value back (1/(…)::int fails with 22012 when it is not the one wanted) —
//     in books, whose other columns are the student's business, by evaluating
//     the default itself;
//   - test rows carry explicit negative ids, so no check moves a sequence: the
//     student's first customer and first order still get id 1, and the
//     reference solutions rely on that. They go in with OVERRIDING SYSTEM
//     VALUE, which serial ignores and an id GENERATED ALWAYS AS IDENTITY
//     (not taught, but a correct id) needs;
//   - every table's tests start with a row that has to go in, which also pins
//     where its references lead: the test order points at a customer that is
//     only in customers, the test items at two orders that are only in orders
//     and at the lowest and highest book ids, which are not all ids of
//     anything else. author_id is pinned by deleting the authors: only a
//     reference from books to authors stops that;
//   - tables are compared column by column as "name type" (lab 3), id being
//     the primary key and numbered by the database; order_items has no id,
//     and its key is tested as behaviour — a repeated pair is refused, so
//     UNIQUE (order_id, book_id) with NOT NULL passes as well;
//   - each task reads its own table and rows, so the tasks pass in any order
//     the references allow. Task 1 holds after the rules of task 2 and after
//     order items point at books; the tests of tasks 4 and 5 need a test
//     customer (and test orders), as any order does, and the diagnostic says
//     so when customers or orders cannot take one.
//
// The quizzes have one answer once the task they name is done: \d authors
// shows "Referenced by", deleting Gogol (three books point at him) fails with
// a foreign key error, and the order of task 6 has status new.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab6.sh.

import (
	"strconv"
	"strings"
)

func init() {
	pgStartLabs["ch-pgs-lab6"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab6Schema)),
		Checks: map[int]string{
			1: check(pgsLab6AuthorDone,
				"в books есть столбец author_id integer со ссылкой на authors (id), у всех 9 книг проставлен их автор",
				"Сейчас: "+pgsLab6AuthorNow),
			2: check(pgsLab6RulesDone,
				"у книги с обнулённой ценой снова 671.00; в books title и price обязательны, цена больше нуля, in_stock обязателен и по умолчанию true",
				"Сейчас: "+pgsLab6RulesNow),
			3: check(pgsLab6Customers.done(),
				"в базе pereplet есть таблица customers со столбцами и правилами из задания",
				"Сейчас: "+pgsLab6Customers.now()),
			4: check(pgsLab6Orders.done(),
				"в базе pereplet есть таблица orders со столбцами и правилами из задания",
				"Сейчас: "+pgsLab6Orders.now()),
			5: check(pgsLab6Items.done(),
				"в базе pereplet есть таблица order_items: две обязательные ссылки, quantity больше нуля, пара order_id + book_id не повторяется",
				"Сейчас: "+pgsLab6Items.now()),
			6: check(pgTrue("pereplet", pgsLab6Order),
				"записан первый заказ: покупатель Анна Смирнова (anna.smirnova@example.com, Москва), один её заказ со статусом new, в нём «Евгений Онегин» × 2 и «Ревизор» × 1",
				"Сейчас: "+pgsLab6OrderNow),
		},
	}
}

// pgsLab6Authors are lab 5's authors; ids 1-5 in this order.
var pgsLab6Authors = []struct {
	name string
	year int
}{
	{"Александр Пушкин", 1799},
	{"Николай Гоголь", 1809},
	{"Фёдор Достоевский", 1821},
	{"Лев Толстой", 1828},
	{"Антон Чехов", 1860},
}

// pgsLab6Book is a row of the catalog. price is the true price; author is the
// full_name task 1's table gives for the book.
type pgsLab6Book struct {
	title, price string
	pages        int
	inStock      bool
	year         int
	author       string
}

// pgsLab6Books is the catalog lab 5 leaves; a row's id is its place in the list.
var pgsLab6Books = []pgsLab6Book{
	{"Евгений Онегин", "495.00", 320, true, 1833, "Александр Пушкин"},
	{"Мёртвые души", "520.00", 416, true, 1842, "Николай Гоголь"},
	{"Анна Каренина", "780.00", 864, true, 1878, "Лев Толстой"},
	{"Дама с собачкой", "210.00", 48, false, 1899, "Антон Чехов"},
	{"Капитанская дочка", "340.00", 176, true, 1836, "Александр Пушкин"},
	{"Шинель", "190.00", 64, false, 1842, "Николай Гоголь"},
	{"Идиот", "671.00", 640, true, 1869, "Фёдор Достоевский"}, // zeroed by the setup
	{"Ревизор", "280.00", 144, true, 1836, "Николай Гоголь"},
	{"Человек в футляре", "160.00", 32, false, 1898, "Антон Чехов"},
}

// pgsLab6Zeroed is the id of the book whose price the setup sets to 0.00.
const pgsLab6Zeroed = 7

// pgsLab6SetupPrice is a book's price as the setup writes it.
func pgsLab6SetupPrice(id int) string {
	if id == pgsLab6Zeroed {
		return "0.00"
	}
	return pgsLab6Books[id-1].price
}

// pgsLab6Schema is state S3 with lab 5's catalog. Lab 5 deleted the duplicate
// with id 10, so the student's next book would get 11.
var pgsLab6Schema = func() string {
	authors := make([]string, len(pgsLab6Authors))
	for i, a := range pgsLab6Authors {
		authors[i] = "    ('" + a.name + "', " + strconv.Itoa(a.year) + ", 'Россия')"
	}
	books := make([]string, len(pgsLab6Books))
	for i, b := range pgsLab6Books {
		books[i] = "    ('" + b.title + "', " + pgsLab6SetupPrice(i+1) + ", " + strconv.Itoa(b.pages) + ", " +
			strconv.FormatBool(b.inStock) + ", " + strconv.Itoa(b.year) + ")"
	}
	return `CREATE TABLE authors (
    id         serial PRIMARY KEY,
    full_name  text,
    birth_year integer,
    country    text
);
CREATE TABLE books (
    id             serial PRIMARY KEY,
    title          text,
    price          numeric(8,2),
    page_count     integer,
    in_stock       boolean,
    published_year integer
);
INSERT INTO authors (full_name, birth_year, country) VALUES
` + strings.Join(authors, ",\n") + `;
INSERT INTO books (title, price, page_count, in_stock, published_year) VALUES
` + strings.Join(books, ",\n") + `;
ALTER SEQUENCE books_id_seq RESTART WITH ` + strconv.Itoa(len(pgsLab6Books)+2) + `;
`
}()

// ---- shared pieces of the diagnostics ----

// pgsLab6Now is a substitution printing a diagnostic query's text, or what to
// do when the tables cannot be read at all.
func pgsLab6Now(sql string) string {
	return `$(glpg q pereplet ` + b64(sql) + ` || echo 'не получилось прочитать таблицы базы pereplet — если открыта транзакция, закончи её (COMMIT; или ROLLBACK;), если базы нет — нажми «Пересоздать»')`
}

// pgsLab6Tx is a substitution printing a note when the student's psql sits in
// an open transaction, whose changes no other session sees until COMMIT.
var pgsLab6Tx = pgVal("pereplet", "SELECT CASE "+
	"WHEN bool_or(state = 'idle in transaction (aborted)') "+
	"THEN '. В твоей транзакции была ошибка (приглашение pereplet=!#): набери ROLLBACK; и повтори команды' "+
	"WHEN bool_or(state = 'idle in transaction') "+
	"THEN '. У тебя открыта транзакция (приглашение pereplet=*#): проверка не видит её изменений — сохрани их командой COMMIT; или отмени командой ROLLBACK;' "+
	"ELSE '' END FROM pg_stat_activity WHERE datname = 'pereplet' AND pid <> pg_backend_pid()")

// pgsLab6Rule is a rule a task asks for: sql has to end with SQLSTATE state
// (00000: has to run), msg is what the student is told when it does not. msg
// goes into a single-quoted shell word, so it has no apostrophes.
type pgsLab6Rule struct{ sql, state, msg string }

// pgsLab6Broken is shell, for inside a substitution, printing "; msg" for
// each rule broken. prep runs before each rule in its transaction; when prep
// itself fails, prepMsg is printed instead. With gate, a broken first rule —
// the row that has to go in — is the only one reported: the rest would fail
// for the same reason.
func pgsLab6Broken(prep, prepMsg string, gate bool, rules []pgsLab6Rule) string {
	var b strings.Builder
	if prep != "" {
		b.WriteString(`test x$(glpg state pereplet ` + b64(prep) + `) = x00000 || { printf %s '; ` + prepMsg + `'; exit 0; }; `)
	}
	for i, r := range rules {
		b.WriteString(`test x$(glpg state pereplet ` + b64(prep+r.sql) + `) = x` + r.state + ` || `)
		if gate && i == 0 {
			b.WriteString(`{ printf %s '; ` + r.msg + `'; exit 0; }; `)
		} else {
			b.WriteString(`printf %s '; ` + r.msg + `'; `)
		}
	}
	return b.String()
}

// pgsLab6Rejects is a condition: every rule holds (prep run first in each).
func pgsLab6Rejects(prep string, rules []pgsLab6Rule) string {
	conds := make([]string, len(rules))
	for i, r := range rules {
		conds[i] = pgRejects("pereplet", prep+r.sql, r.state)
	}
	return strings.Join(conds, " && ")
}

// pgsLab6IfTable is a substitution running sh only when table exists and no
// other session of pereplet sits in a transaction. Such a session may hold a
// lock (BEGIN; ALTER TABLE … with no COMMIT yet), and every rule would then
// wait out glpg's lock_timeout — longer, all together, than the runner gives a
// check. pgsLab6Tx tells the student to finish the transaction instead.
func pgsLab6IfTable(table, sh string) string {
	return `$(if test x` + pgVal("pereplet", "SELECT to_regclass('"+table+"') IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pg_stat_activity "+
		"WHERE datname = 'pereplet' AND pid <> pg_backend_pid() AND state LIKE 'idle in transaction%')") + ` = xt; then ` + sh + `fi)`
}

// pgsLab6Tables is SQL for the tables in pereplet by name, as \dt lists them.
const pgsLab6Tables = "coalesce((SELECT string_agg(CASE WHEN schemaname = 'public' THEN tablename " +
	"ELSE schemaname || '.' || tablename END, ', ' ORDER BY schemaname <> 'public', schemaname, tablename) " +
	"FROM pg_tables WHERE schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg\\_%'), 'ни одной')"

// pgsLab6Shown is SQL for table's columns the way \d describes them: name,
// type, not null, default — the numbering of a serial id said in words.
func pgsLab6Shown(table string) string {
	def := "pg_get_expr(d.adbin, d.adrelid)"
	return "coalesce((SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod) || " +
		"CASE WHEN a.attnotnull THEN ' not null' ELSE '' END || " +
		"CASE WHEN " + def + " LIKE 'nextval(%' THEN ' (номер ставит база)' ELSE coalesce(' default ' || " + def + ", '') END, " +
		"', ' ORDER BY a.attnum) FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum " +
		"WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped), 'ни одного')"
}

// ---- task 1: author_id ----

// pgsLab6Map is task 1's table as rows m(id, title, author).
var pgsLab6Map = func() string {
	rows := make([]string, len(pgsLab6Books))
	for i, b := range pgsLab6Books {
		rows[i] = "(" + strconv.Itoa(i+1) + ", '" + b.title + "', '" + b.author + "')"
	}
	return "(VALUES " + strings.Join(rows, ", ") + ") AS m(id, title, author)"
}()

// pgsLab6AuthorRefs is SQL for the tables books.author_id references, NULL
// when it references none.
const pgsLab6AuthorRefs = "(SELECT string_agg(c.confrelid::regclass::text, ', ') FROM pg_constraint c " +
	"JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey) " +
	"WHERE c.contype = 'f' AND c.conrelid = to_regclass('books') AND a.attname = 'author_id')"

// pgsLab6AuthorOnDelete is SQL true when that reference has an ON DELETE
// action (CASCADE, SET NULL …): deleting an author then goes through, which
// the check — and the quiz about deleting Gogol — expect to fail.
const pgsLab6AuthorOnDelete = "EXISTS (SELECT 1 FROM pg_constraint c " +
	"JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey) " +
	"WHERE c.contype = 'f' AND c.conrelid = to_regclass('books') AND a.attname = 'author_id' AND c.confdeltype NOT IN ('a', 'r'))"

// pgsLab6AuthorDone: the column, a reference that turns away an author who is
// not there and keeps authors with books from being deleted, and every book of
// the setup by title with the author of task 1's table by full_name.
var pgsLab6AuthorDone = pgHasCol("pereplet", "books", "author_id", "integer") + " && " +
	pgRejects("pereplet", "UPDATE books SET author_id = (SELECT coalesce(max(id), 0) + 1 FROM authors) WHERE id = (SELECT min(id) FROM books)", "23503") + " && " +
	pgRejects("pereplet", "DELETE FROM authors", "23503") + " && " +
	pgTrue("pereplet", "SELECT NOT EXISTS (SELECT 1 FROM "+pgsLab6Map+" WHERE NOT EXISTS (SELECT 1 FROM books b WHERE b.title = m.title)) "+
		"AND NOT EXISTS (SELECT 1 FROM books b JOIN "+pgsLab6Map+" ON m.title = b.title "+
		"LEFT JOIN authors a ON a.id = b.author_id WHERE a.full_name IS DISTINCT FROM m.author)")

// pgsLab6AuthorNow: the column and where it leads (read from the catalog, so
// the query runs whether or not author_id is there), then — only when it is —
// the books without an author or with another one, which the task's table
// names anyway, and books gone from the catalog.
var pgsLab6AuthorNow = pgsLab6Now("SELECT CASE "+
	"WHEN to_regclass('books') IS NULL THEN 'в базе pereplet нет таблицы books — нажми «Пересоздать»' "+
	"WHEN ("+pgColType("books", "author_id")+") IS NULL THEN 'в books нет столбца author_id; столбцы books: ' || "+pgsLab6Shown("books")+" "+
	"ELSE 'author_id ' || ("+pgColType("books", "author_id")+") || "+
	"CASE WHEN ("+pgColType("books", "author_id")+") <> 'integer' THEN ', а нужен тип integer' ELSE '' END || "+
	"CASE WHEN "+pgsLab6AuthorRefs+" IS NULL THEN ', но ни на что не ссылается: нужна ссылка REFERENCES authors (id)' "+
	"WHEN "+pgsLab6AuthorRefs+" <> 'authors' THEN ', ссылается на ' || "+pgsLab6AuthorRefs+" || ', а нужно на authors (id)' "+
	"WHEN "+pgsLab6AuthorOnDelete+" THEN ', ссылается на authors, но с ON DELETE: так база разрешает удалить автора, у которого есть книги — "+
	"убери ссылку (DROP CONSTRAINT, имя есть в \\d books) и добавь заново, без ON DELETE' "+
	"ELSE ', ссылается на authors' END END") +
	pgVal("pereplet", "SELECT coalesce('; ' || nullif(concat_ws('; ', "+
		"(SELECT 'без автора (author_id пустой): ' || count(*) || ' из "+strconv.Itoa(len(pgsLab6Books))+" книг' FROM books b JOIN "+pgsLab6Map+" ON m.title = b.title "+
		"WHERE b.author_id IS NULL HAVING count(*) > 0), "+
		"(SELECT 'не тот автор: ' || string_agg('«' || b.title || '» — сейчас ' || coalesce(a.full_name, 'номер ' || b.author_id || ', а такого автора нет'), ', ' ORDER BY b.id) "+
		"FROM books b JOIN "+pgsLab6Map+" ON m.title = b.title LEFT JOIN authors a ON a.id = b.author_id "+
		"WHERE b.author_id IS NOT NULL AND a.full_name IS DISTINCT FROM m.author), "+
		"(SELECT 'из каталога пропали книги: ' || string_agg(m.title, ', ' ORDER BY m.id) || ' — нажми «Пересоздать»' FROM "+pgsLab6Map+" "+
		"WHERE NOT EXISTS (SELECT 1 FROM books b WHERE b.title = m.title))"+
		"), ''), '')") +
	pgsLab6Tx

// ---- task 2: the catalog's rules ----

// pgsLab6Moved is SQL for the books, other than the zeroed one, whose price
// is not the setup's. A deleted book is no price change.
var pgsLab6Moved = func() string {
	orig := make([]string, len(pgsLab6Books))
	for i := range pgsLab6Books {
		orig[i] = "(" + strconv.Itoa(i+1) + ", " + pgsLab6SetupPrice(i+1) + ")"
	}
	return "SELECT b.id, b.title FROM books b JOIN (VALUES " + strings.Join(orig, ", ") + ") AS o(id, price) USING (id) " +
		"WHERE b.id <> " + strconv.Itoa(pgsLab6Zeroed) + " AND b.price IS DISTINCT FROM o.price"
}()

// pgsLab6Price is SQL true when the zeroed book costs its price again and no
// other book of the setup changed price.
var pgsLab6Price = "SELECT coalesce((SELECT price = " + pgsLab6Books[pgsLab6Zeroed-1].price + " FROM books WHERE id = " +
	strconv.Itoa(pgsLab6Zeroed) + "), false) AND NOT EXISTS (" + pgsLab6Moved + ")"

// pgsLab6BookRules are task 2's rules, each tried on the first book. A
// positive price has to go in too: the rule is "more than zero", no stricter.
// The default of in_stock is evaluated rather than read off an inserted book:
// a book added with title and price only would also need every other column
// of the student's books to be optional, and the task does not forbid making,
// say, author_id NOT NULL. With no default at all, EXECUTE of NULL fails.
var pgsLab6BookRules = []pgsLab6Rule{
	{"UPDATE books SET title = NULL WHERE id = (SELECT min(id) FROM books)", "23502", "title может быть пустым: нужен NOT NULL"},
	{"UPDATE books SET price = NULL WHERE id = (SELECT min(id) FROM books)", "23502", "price может быть пустым: нужен NOT NULL"},
	{"UPDATE books SET price = 0 WHERE id = (SELECT min(id) FROM books)", "23514", "цену 0 база принимает: нужно правило CHECK, что цена больше нуля"},
	{"UPDATE books SET price = 0.01 WHERE id = (SELECT min(id) FROM books)", "00000", "цену 0.01 база не принимает: правило должно запрещать только цены не больше нуля"},
	{"UPDATE books SET in_stock = NULL WHERE id = (SELECT min(id) FROM books)", "23502", "in_stock может быть пустым: нужен NOT NULL"},
	{"DO $$ DECLARE v boolean; BEGIN " +
		"EXECUTE 'SELECT ' || (SELECT pg_get_expr(d.adbin, d.adrelid) FROM pg_attrdef d JOIN pg_attribute a ON a.attrelid = d.adrelid AND a.attnum = d.adnum " +
		"WHERE d.adrelid = 'books'::regclass AND a.attname = 'in_stock') INTO v; " +
		"PERFORM 1 / (v IS TRUE)::int; END $$", "00000", "книга, добавленная без in_stock, получает не true: нужен DEFAULT true"},
}

var pgsLab6RulesDone = pgTrue("pereplet", pgsLab6Price) + " && " + pgsLab6Rejects("", pgsLab6BookRules)

// pgsLab6RulesNow: the zeroed price — without naming the book while it is
// still zero: finding it is part of the task — other prices that moved, then
// the rules the table does not keep yet.
var pgsLab6RulesNow = func() string {
	z := strconv.Itoa(pgsLab6Zeroed)
	p := "(SELECT price FROM books WHERE id = " + z + ")"
	return pgsLab6Now("SELECT concat_ws('; ', CASE "+
		"WHEN NOT EXISTS (SELECT 1 FROM books WHERE id = "+z+") "+
		"THEN 'книги с обнулённой ценой (id "+z+") в каталоге нет — удалять её было не нужно, нажми «Пересоздать»' "+
		"WHEN "+p+" IS NULL THEN 'у книги с обнулённой ценой цена теперь пустая — поставь настоящую цену из задания' "+
		"WHEN "+p+" <= 0 THEN 'книг с ценой 0 или меньше: ' || (SELECT count(*) FROM books WHERE price <= 0) || ' — найди её и поставь настоящую цену из задания' "+
		"WHEN "+p+" <> "+pgsLab6Books[pgsLab6Zeroed-1].price+" THEN 'у книги с обнулённой ценой сейчас ' || "+p+" || ', а настоящая цена — в задании' "+
		"ELSE 'цена исправлена' END, "+
		"(SELECT 'цена изменилась и у других книг: ' || string_agg(m.title, ', ' ORDER BY m.id) || ' — нажми «Пересоздать»' "+
		"FROM ("+pgsLab6Moved+") m))") +
		pgsLab6IfTable("books", pgsLab6Broken("", "", false, pgsLab6BookRules)) +
		pgsLab6Tx
}()

// ---- tasks 3-5: the new tables ----

type pgsLab6Col struct{ name, typ string }

// pgsLab6Table is a table of tasks 3-5: its columns (types as format_type
// prints them), whether id is a serial primary key, the test rows its rules
// need (prep, run first in each rule's transaction) and the rules, the first
// of which is a row that has to go in.
type pgsLab6Table struct {
	name    string
	cols    []pgsLab6Col
	serial  bool
	prep    string
	prepMsg string
	rules   []pgsLab6Rule
}

// pk is SQL true when id on its own is the primary key and the database
// numbers it: serial (DEFAULT nextval) or an identity column.
func (t pgsLab6Table) pk() string {
	return "EXISTS (SELECT 1 FROM pg_constraint c JOIN pg_attribute a ON a.attrelid = c.conrelid AND c.conkey = ARRAY[a.attnum] " +
		"WHERE c.conrelid = to_regclass('" + t.name + "') AND c.contype = 'p' AND a.attname = 'id')"
}

func (t pgsLab6Table) auto() string {
	return "EXISTS (SELECT 1 FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum " +
		"WHERE a.attrelid = to_regclass('" + t.name + "') AND a.attname = 'id' AND NOT a.attisdropped " +
		"AND (a.attidentity IN ('a', 'd') OR pg_get_expr(d.adbin, d.adrelid) LIKE 'nextval(%'))"
}

// shapeOK is SQL true when the table has exactly the columns of the task, with
// their types, and (serial) id is a primary key the database numbers.
func (t pgsLab6Table) shapeOK() string {
	want := make([]string, len(t.cols))
	for i, c := range t.cols {
		want[i] = "'" + c.name + " " + c.typ + "'"
	}
	sql := "SELECT (SELECT array_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod) ORDER BY a.attname) " +
		"FROM pg_attribute a WHERE a.attrelid = to_regclass('" + t.name + "') AND a.attnum > 0 AND NOT a.attisdropped) " +
		"= (SELECT array_agg(w ORDER BY w) FROM unnest(ARRAY[" + strings.Join(want, ", ") + "]) w)"
	if t.serial {
		sql += " AND " + t.pk() + " AND " + t.auto()
	}
	return sql
}

// diff is SQL naming the columns missing, extra or of another type, NULL when
// none are. It never says what the type should be: the task does.
func (t pgsLab6Table) diff() string {
	vals := make([]string, len(t.cols))
	for i, c := range t.cols {
		vals[i] = "('" + c.name + "', '" + c.typ + "', " + strconv.Itoa(i+1) + ")"
	}
	w := "(VALUES " + strings.Join(vals, ", ") + ") w(name, typ, n)"
	h := "(SELECT a.attname::text AS name, format_type(a.atttypid, a.atttypmod) AS typ, a.attnum AS n FROM pg_attribute a " +
		"WHERE a.attrelid = to_regclass('" + t.name + "') AND a.attnum > 0 AND NOT a.attisdropped) h"
	return "nullif(concat_ws('; ', " +
		"(SELECT 'не хватает столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " WHERE w.name NOT IN (SELECT h.name FROM " + h + ")), " +
		"(SELECT 'лишние столбцы: ' || string_agg(h.name, ', ' ORDER BY h.n) FROM " + h + " WHERE h.name NOT IN (SELECT w.name FROM " + w + ")), " +
		"(SELECT 'не тот тип у столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " JOIN " + h + " ON h.name = w.name WHERE h.typ <> w.typ)" +
		"), '')"
}

func (t pgsLab6Table) done() string {
	return pgTrue("pereplet", t.shapeOK()) + " && " + pgsLab6Rejects(t.prep, t.rules)
}

// now: the table as \d shows it and what differs from the task, then the
// rules it does not keep, then a note about an open transaction.
func (t pgsLab6Table) now() string {
	idNote := "''"
	if t.serial {
		idNote = "CASE WHEN (" + pgColType(t.name, "id") + ") IS NULL THEN '' " +
			"WHEN NOT " + t.pk() + " THEN '; id — не первичный ключ' " +
			"WHEN NOT " + t.auto() + " THEN '; номер в id база сама не ставит' ELSE '' END"
	}
	shape := "SELECT CASE WHEN to_regclass('" + t.name + "') IS NULL " +
		"THEN 'в базе pereplet нет таблицы " + t.name + ", таблицы в базе: ' || " + pgsLab6Tables + " " +
		"ELSE 'столбцы " + t.name + ": ' || " + pgsLab6Shown(t.name) + " || coalesce('; ' || " + t.diff() + ", '') || " + idNote + " END"
	return pgsLab6Now(shape) + pgsLab6IfTable(t.name, pgsLab6Broken(t.prep, t.prepMsg, true, t.rules)) + pgsLab6Tx
}

// The test rows. Their ids are negative: no student row has one, and an id of
// their own leaves the sequences alone.
const (
	pgsLab6TestCustomer = "INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (-1, 'check@pereplet.test', 'Проверка');\n"
	pgsLab6TestOrders   = "INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (-1, -1), (-2, -1);\n"
	pgsLab6First        = "(SELECT min(id) FROM books)"
	pgsLab6Last         = "(SELECT max(id) FROM books)"
)

var pgsLab6Customers = pgsLab6Table{
	name:   "customers",
	cols:   []pgsLab6Col{{"id", "integer"}, {"email", "text"}, {"full_name", "text"}, {"city", "text"}, {"registered_at", "date"}},
	serial: true,
	rules: []pgsLab6Rule{
		{"INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (-1, 'check@pereplet.test', 'Проверка')", "00000",
			"покупатель только с email и full_name не добавляется: city должен быть необязательным, а registered_at база должна заполнять сама"},
		{"WITH n AS (INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (-1, 'check@pereplet.test', 'Проверка') RETURNING registered_at) " +
			"SELECT 1 / (((SELECT registered_at FROM n) = current_date) IS TRUE)::int", "00000",
			"покупатель, добавленный без registered_at, получает не сегодняшнюю дату: нужен DEFAULT current_date"},
		{"INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (-1, NULL, 'Проверка')", "23502", "email может быть пустым: нужен NOT NULL"},
		{"INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (-1, 'check@pereplet.test', NULL)", "23502", "full_name может быть пустым: нужен NOT NULL"},
		{"INSERT INTO customers (id, email, full_name, registered_at) OVERRIDING SYSTEM VALUE VALUES (-1, 'check@pereplet.test', 'Проверка', NULL)", "23502",
			"registered_at может быть пустым: нужен NOT NULL"},
		{"INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (-1, 'check@pereplet.test', 'Проверка'), (-2, 'check@pereplet.test', 'Проверка 2')", "23505",
			"один email можно записать двум покупателям: нужен UNIQUE"},
	},
}

var pgsLab6Orders = pgsLab6Table{
	name:    "orders",
	cols:    []pgsLab6Col{{"id", "integer"}, {"customer_id", "integer"}, {"created_at", "date"}, {"status", "text"}},
	serial:  true,
	prep:    pgsLab6TestCustomer,
	prepMsg: "проверка не может добавить в customers покупателя только с email и full_name — сначала сдай задание «Таблица покупателей»",
	rules: []pgsLab6Rule{
		{"INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (-1, -1)", "00000",
			"заказ только с customer_id не добавляется: created_at и status база должна заполнять сама, а customer_id — ссылаться на customers (id)"},
		{"WITH n AS (INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (-1, -1) RETURNING created_at) " +
			"SELECT 1 / (((SELECT created_at FROM n) = current_date) IS TRUE)::int", "00000",
			"заказ, добавленный без created_at, получает не сегодняшнюю дату: нужен DEFAULT current_date"},
		{"WITH n AS (INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (-1, -1) RETURNING status) " +
			"SELECT 1 / (((SELECT status FROM n) = 'new') IS TRUE)::int", "00000",
			"заказ, добавленный без status, получает не new: нужен DEFAULT со словом new"},
		{"INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (-1, NULL)", "23502", "customer_id может быть пустым: нужен NOT NULL"},
		{"INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (-1, -999)", "23503",
			"заказ можно записать покупателю, которого нет: у customer_id нужна ссылка REFERENCES customers (id)"},
		{"INSERT INTO orders (id, customer_id, created_at) OVERRIDING SYSTEM VALUE VALUES (-1, -1, NULL)", "23502", "created_at может быть пустым: нужен NOT NULL"},
		{"INSERT INTO orders (id, customer_id, status) OVERRIDING SYSTEM VALUE VALUES (-1, -1, NULL)", "23502", "status может быть пустым: нужен NOT NULL"},
		{"INSERT INTO orders (id, customer_id, status) OVERRIDING SYSTEM VALUE VALUES (-1, -1, 'returned')", "23514",
			"в status можно записать слово не из списка: нужно правило CHECK со списком new, paid, shipped, cancelled"},
		{"INSERT INTO orders (id, customer_id, status) OVERRIDING SYSTEM VALUE VALUES (-1, -1, 'new'), (-2, -1, 'paid'), (-3, -1, 'shipped'), (-4, -1, 'cancelled')", "00000",
			"база принимает не все статусы из списка: проверь, что в CHECK все четыре слова — new, paid, shipped, cancelled — написаны точно"},
	},
}

var pgsLab6Items = pgsLab6Table{
	name:    "order_items",
	cols:    []pgsLab6Col{{"order_id", "integer"}, {"book_id", "integer"}, {"quantity", "integer"}},
	prep:    pgsLab6TestCustomer + pgsLab6TestOrders,
	prepMsg: "проверка не может добавить в orders заказ только с customer_id — сначала сдай задания «Таблица покупателей» и «Таблица заказов»",
	rules: []pgsLab6Rule{
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-1, " + pgsLab6First + ", 1), (-1, " + pgsLab6Last + ", 2), (-2, " + pgsLab6First + ", 1)", "00000",
			"позиции не добавляются: проверь, что order_id ссылается на orders (id), book_id — на books (id), а не повторяется только пара order_id + book_id"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (NULL, " + pgsLab6First + ", 1)", "23502", "order_id может быть пустым: нужен NOT NULL"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-1, NULL, 1)", "23502", "book_id может быть пустым: нужен NOT NULL"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-1, " + pgsLab6First + ", NULL)", "23502", "quantity может быть пустым: нужен NOT NULL"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-1, " + pgsLab6First + ", 0)", "23514",
			"количество 0 база принимает: нужно правило CHECK, что quantity больше нуля"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-999, " + pgsLab6First + ", 1)", "23503",
			"позицию можно записать в заказ, которого нет: у order_id нужна ссылка REFERENCES orders (id)"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-1, -999, 1)", "23503",
			"в позицию можно записать книгу, которой нет: у book_id нужна ссылка REFERENCES books (id)"},
		{"INSERT INTO order_items (order_id, book_id, quantity) VALUES (-1, " + pgsLab6First + ", 1), (-1, " + pgsLab6First + ", 2)", "23505",
			"одну книгу можно записать в один заказ дважды: нужен первичный ключ из пары PRIMARY KEY (order_id, book_id)"},
	},
}

// ---- task 6: the first order ----

// pgsLab6Anna is the customer of task 6, found by email; the CTEs give her
// rows (c), her orders (o) and their items with the book titles (i).
const pgsLab6Anna = "WITH c AS (SELECT id, full_name, city FROM customers WHERE email = 'anna.smirnova@example.com'), " +
	"o AS (SELECT r.id, r.status FROM orders r JOIN c ON r.customer_id = c.id), " +
	"i AS (SELECT b.title, x.quantity FROM order_items x JOIN o ON x.order_id = o.id JOIN books b ON b.id = x.book_id) "

// pgsLab6Order is SQL true when Anna is there once with her name and city,
// with one order of status new holding exactly the two items of the task.
const pgsLab6Order = pgsLab6Anna + "SELECT (SELECT count(*) FROM c) = 1 " +
	"AND EXISTS (SELECT 1 FROM c WHERE full_name = 'Анна Смирнова' AND city = 'Москва') " +
	"AND (SELECT count(*) FROM o) = 1 AND EXISTS (SELECT 1 FROM o WHERE status = 'new') " +
	"AND (SELECT count(*) FROM i) = 2 " +
	"AND EXISTS (SELECT 1 FROM i WHERE title = 'Евгений Онегин' AND quantity = 2) " +
	"AND EXISTS (SELECT 1 FROM i WHERE title = 'Ревизор' AND quantity = 1)"

// pgsLab6OrderNow walks the order the student builds it in: the customer, her
// order, its items — and stops at the first step that is not there yet.
var pgsLab6OrderNow = `$(glpg q pereplet ` + b64(pgsLab6Anna+"SELECT CASE "+
	"WHEN NOT EXISTS (SELECT 1 FROM c) THEN 'покупателя с email anna.smirnova@example.com в customers нет' || "+
	"coalesce(' (email покупателей сейчас: ' || (SELECT string_agg(email, ', ' ORDER BY id) FROM customers) || ')', '') "+
	"WHEN (SELECT count(*) FROM c) > 1 THEN 'покупатель с email anna.smirnova@example.com записан ' || (SELECT count(*) FROM c) || "+
	"' раза — лишние строки удали (DELETE … WHERE id = …), а у email должно быть правило UNIQUE' "+
	"ELSE concat_ws('; ', "+
	"(SELECT 'у покупателя имя ' || quote_nullable(full_name) || ' и город ' || quote_nullable(city) || ' — в задании другие' FROM c "+
	"WHERE full_name IS DISTINCT FROM 'Анна Смирнова' OR city IS DISTINCT FROM 'Москва'), "+
	"CASE (SELECT count(*) FROM o) WHEN 0 THEN 'покупатель есть, а заказа у него нет' WHEN 1 THEN NULL "+
	"ELSE 'у покупателя заказов: ' || (SELECT count(*) FROM o) || ', а нужен один — лишний удали, сначала его позиции' END, "+
	"(SELECT 'у заказа статус ' || quote_nullable(status) || ', а должен быть new: не указывай status, его поставит база' FROM o "+
	"WHERE (SELECT count(*) FROM o) = 1 AND status IS DISTINCT FROM 'new'), "+
	"CASE WHEN (SELECT count(*) FROM o) = 1 AND NOT (SELECT (SELECT count(*) FROM i) = 2 "+
	"AND EXISTS (SELECT 1 FROM i WHERE title = 'Евгений Онегин' AND quantity = 2) "+
	"AND EXISTS (SELECT 1 FROM i WHERE title = 'Ревизор' AND quantity = 1)) "+
	"THEN coalesce('в заказе: ' || (SELECT string_agg('«' || title || '» × ' || quantity, ', ' ORDER BY title) FROM i) || ' — а в задании другое', "+
	"'в заказе пока нет позиций') END) END") +
	` || echo 'не получилось прочитать таблицы customers, orders и order_items — если их ещё нет, сначала сдай задания про них')` +
	pgsLab6Tx
