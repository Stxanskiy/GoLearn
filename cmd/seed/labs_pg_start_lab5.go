package main

// Лабораторная 5 курса pg-start: связываем таблицы (ch-pgs-lab5).
//
// pereplet holds authors and books in their canonical shape (§5 of the course
// contract) except for two things the student adds: books has no author_id and
// no books_price_positive. The catalogue came over from Excel, and «Война и
// мир» came with a price of 0 — the row that makes "ADD CONSTRAINT … CHECK
// (price > 0)" fail until the data is fixed. customers, orders, order_items,
// genres and book_genres do not exist: the student creates them, places the
// shop's first order through all three order tables, and, in the ★ task, links
// books to genres many-to-many. After the lab the database is the canonical
// schema that lab 6 starts from.
//
// The contract's outline is kept with two changes. Fixing the price comes
// third, right after the author link: both change a table that already holds
// data, before the student moves on to creating new ones. And a task is added
// before the ★ — the first order, entered into customers, orders and
// order_items — because it is the one place where the links are used rather
// than only declared, and the quiz after it needs a customer with an order.
// Where a link keeps the default delete rule, the tasks say "без ON DELETE"
// rather than what that rule does: that quiz is where the student sees it.
//
// The checks test what the tables do, not how they were written: each rule is
// a statement run in a rolled-back transaction (glpg state) that must end with
// a given SQLSTATE — a duplicate email with 23505, an order for a missing
// customer with 23503 — whatever the constraints are called and wherever they
// were declared. Probe rows carry explicit ids from 900001 up (OVERRIDING
// SYSTEM VALUE), so they never collide with the student's rows and never move
// an identity counter. Where a rule is "gets a value by default", the probe
// ends in SELECT 1/(condition)::int: a false condition is a division by zero,
// 22012. Column types are compared as format_type prints them. The one name a
// check insists on is books_price_positive: the task asks for it, and the
// theory explains why a constraint gets a name.
//
// Each failed check says what is wrong now: a missing table (and whether it
// was created in the root database, where a bare psql connects), columns that
// differ from the task, or the rules the database does not enforce yet — never
// the statement that fixes it.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab5.sh.

import (
	"sort"
	"strconv"
	"strings"
)

// pgsLab5Pereplet is the database the lab starts from. Ids are fixed — a fresh
// database, authors 1…6 and books 1…10 in this order — but no check relies on
// them: rows are recognised by name and title. Two books have no ISBN, which
// the quiz on UNIQUE and NULL looks at.
const pgsLab5Pereplet = `
CREATE TABLE authors (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name  text NOT NULL,
    birth_year integer,
    country    text
);

CREATE TABLE books (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title          text NOT NULL,
    price          numeric(8,2) NOT NULL,
    page_count     integer,
    published_year integer,
    isbn           text UNIQUE,
    in_stock       boolean NOT NULL DEFAULT true
);

INSERT INTO authors (full_name, birth_year, country) VALUES
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Михаил Булгаков',   1891, 'Россия'),
    ('Антон Чехов',       1860, 'Россия'),
    ('Александр Пушкин',  1799, 'Россия'),
    ('Лев Толстой',       1828, 'Россия'),
    ('Агата Кристи',      1890, 'Великобритания');

INSERT INTO books (title, price, page_count, published_year, isbn, in_stock) VALUES
    ('Преступление и наказание',       590.00,  608, 1866, '978-5-389-04926-0', true),
    ('Мастер и Маргарита',             690.00,  480, 1967, '978-5-389-01686-6', true),
    ('Вишнёвый сад',                   280.00,   96, 1904, '978-5-389-06010-4', false),
    ('Идиот',                          540.00,  640, 1869, '978-5-389-05178-2', false),
    ('Каштанка',                       190.00,   48, 1887, '978-5-389-10364-1', false),
    ('Евгений Онегин',                 495.00,  320, 1833, NULL,                true),
    ('Анна Каренина',                  790.00,  864, 1878, '978-5-389-21580-1', true),
    ('Убийство в Восточном экспрессе', 520.00,  256, 1934, '978-5-04-112380-2', true),
    ('Война и мир',                      0.00, 1300, 1869, '978-5-389-06256-6', true),
    ('Капитанская дочка',              260.00,  224, 1836, NULL,                true);
`

// pgsLab5Rule is one behaviour a task asks for: sql runs in a transaction that
// is rolled back and must end with SQLSTATE want (00000 — no error). what
// names the attempt in a diagnostic, without quotes of either kind.
type pgsLab5Rule struct{ what, sql, want string }

// probe is the rule's SQL as it runs: a deferrable constraint would only
// fire at a COMMIT that never comes, so it is made to fire now.
func (r pgsLab5Rule) probe() string { return r.sql + ";\nSET CONSTRAINTS ALL IMMEDIATE" }

// pgsLab5Hold is the condition "every rule behaves".
func pgsLab5Hold(rules []pgsLab5Rule) string {
	var conds []string
	for _, r := range rules {
		conds = append(conds, pgRejects("pereplet", r.probe(), r.want))
	}
	return strings.Join(conds, " && ")
}

// pgsLab5Broken prints "what — what the database did; " for each rule that
// does not behave: "база это принимает" where a refusal was due, "нет" where a
// probe's value is wrong (22012), and otherwise the error itself, the line a
// student sees in psql — "null value in column "created_at" of relation
// "orders" violates not-null constraint" says which column is at fault, where
// a bare SQLSTATE would not. The probe runs as glpg state runs it, with the
// message kept: in verbose mode the first line is "ERROR:  23502: message".
// The case patterns open with "(": check() trims unbalanced closing
// parentheses off the end of a diagnostic.
func pgsLab5Broken(rules []pgsLab5Rule) string {
	var b strings.Builder
	for _, r := range rules {
		b.WriteString(`e=$({ echo 'BEGIN;'; echo ` + b64(r.probe()) + ` | base64 -d; echo ';'; echo 'ROLLBACK;'; } | ` +
			`psql -X -q -v ON_ERROR_STOP=1 -v VERBOSITY=verbose -d pereplet 2>&1 >/dev/null | grep -m1 '^ERROR:'); ` +
			`s=${e:8:5}; s=${s:-00000}; [ $s = ` + r.want + ` ] || { echo -n '` + r.what + ` — '; ` +
			`case $s in (00000) echo -n 'база это принимает';; (22012) echo -n 'нет';; (*) echo -n 'ошибка:' ${e#ERROR:  ?????: };; esac; ` +
			`echo -n '; '; }; `)
	}
	return b.String()
}

// pgsLab5Table is a table a task creates: its columns as "name type" (the type
// as format_type prints it) and its primary key — "id" for an identity id,
// otherwise the columns of a composite key.
type pgsLab5Table struct {
	name string
	pk   []string
	cols []string
}

// shape is the SQL of what is wrong with the table, problems joined by "; ",
// or "ok".
func (t pgsLab5Table) shape() string {
	var want []string
	for i, c := range t.cols {
		name, typ, _ := strings.Cut(c, " ")
		want = append(want, "("+strconv.Itoa(i+1)+", '"+name+"', '"+typ+"')")
	}
	var pkOK, pkBad string
	if len(t.pk) == 1 && t.pk[0] == "id" {
		pkOK = `EXISTS (SELECT 1 FROM pg_attribute a JOIN pg_constraint c
            ON c.conrelid = a.attrelid AND c.contype = 'p' AND c.conkey = ARRAY[a.attnum]
            WHERE a.attrelid = to_regclass('` + t.name + `') AND a.attname = 'id' AND a.attidentity IN ('a', 'd'))`
		pkBad = "id — не identity или не первичный ключ"
	} else {
		pk := append([]string(nil), t.pk...)
		sort.Strings(pk)
		pkOK = `EXISTS (SELECT 1 FROM pg_constraint c WHERE c.conrelid = to_regclass('` + t.name + `') AND c.contype = 'p'
            AND (SELECT array_agg(a.attname::text ORDER BY a.attname::text) FROM pg_attribute a
                 WHERE a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)) = ARRAY['` + strings.Join(pk, "', '") + `'])`
		pkBad = "первичный ключ — не пара " + strings.Join(t.pk, " + ")
	}
	return `WITH want (n, col, typ) AS (VALUES ` + strings.Join(want, ", ") + `),
got AS (SELECT w.n, w.col, w.typ, format_type(a.atttypid, a.atttypmod) AS t FROM want w
    LEFT JOIN pg_attribute a ON a.attrelid = to_regclass('` + t.name + `') AND a.attname = w.col
        AND a.attnum > 0 AND NOT a.attisdropped)
SELECT CASE WHEN to_regclass('` + t.name + `') IS NULL THEN 'таблицы ` + t.name + ` в базе pereplet нет'
    ELSE coalesce(nullif(concat_ws('; ',
        (SELECT string_agg(CASE WHEN t IS NULL THEN 'нет столбца ' || col
                ELSE 'у ' || col || ' тип ' || t || ', а нужен ' || typ END, '; ' ORDER BY n)
            FROM got WHERE t IS DISTINCT FROM typ),
        CASE WHEN ` + pkOK + ` THEN NULL ELSE '` + pkBad + `' END), ''), 'ok') END`
}

// ok is the condition "the table is as the task asks".
func (t pgsLab5Table) ok() string { return pgIs("pereplet", t.shape(), "ok") }

// now prints the table's problems followed by "; " and sets bad when there are
// any. A table missing from pereplet but present in root is the classic slip of
// this course: a bare psql connects to the database root.
func (t pgsLab5Table) now() string {
	exists := b64("SELECT to_regclass('" + t.name + "') IS NOT NULL")
	return `d=$(glpg q pereplet ` + b64(t.shape()) + `); case $d in (ok) ;; ` +
		`('') bad=1; echo -n 'база pereplet не отвечает; ';; ` +
		`(таблицы*) bad=1; echo -n $d; [ x$(glpg q root ` + exists + `) = xt ] && echo -n ' — зато она есть в базе root: проверь, к какой базе подключён psql'; echo -n '; ';; ` +
		`(*) bad=1; echo -n $d'; ';; esac; `
}

// pgsLab5TablesNow is the diagnostic of a task that creates tables: what is
// wrong with their shape, or, once they are right, the rules that do not hold;
// then whatever extra prints (a shell snippet ending in "; ").
func pgsLab5TablesNow(tables []pgsLab5Table, rules []pgsLab5Rule, extra string) string {
	var b strings.Builder
	b.WriteString(`$({ bad=; `)
	for _, t := range tables {
		b.WriteString(t.now())
	}
	b.WriteString(`case $bad in ('') ` + pgsLab5Broken(rules) + `;; esac; ` + extra + `} | sed 's/; *$//')`)
	return b.String()
}

// The rows rules lean on: a customer, an order of that customer, a genre.
// Their ids are far above any the student's rows will get. A probe that
// expects a refusal fills every other column with a valid value, so the
// refusal it gets comes from the rule under test and not from a missing
// default elsewhere.
const (
	pgsLab5ProbeCustomer = `INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE
    VALUES (900001, 'glpg-probe-900001@check.invalid', 'Проверка'); `
	pgsLab5ProbeOrder = pgsLab5ProbeCustomer + `INSERT INTO orders (id, customer_id, created_at, status) OVERRIDING SYSTEM VALUE
    VALUES (900001, 900001, now(), 'new'); `
	pgsLab5ProbeGenre = `INSERT INTO genres (id, name) OVERRIDING SYSTEM VALUE VALUES (900001, 'glpg: жанр 900001'); `
	// pgsLab5ProbeBook is book 900001, a copy of the first book under another
	// title and without an ISBN: whatever the student required of a book
	// (author_id NOT NULL, say), the copy has it.
	pgsLab5ProbeBook = `INSERT INTO books OVERRIDING SYSTEM VALUE
SELECT (jsonb_populate_record(NULL::books, to_jsonb(b) || jsonb_build_object('id', 900001, 'title', 'Проверка', 'isbn', NULL))).*
FROM books b WHERE b.id = ` + pgsLab5FirstBook + `;`
	pgsLab5FirstBook = `(SELECT min(id) FROM books)`
	pgsLab5LastBook  = `(SELECT max(id) FROM books)`
	// pgsLab5Recent is "within a minute of now()": what a DEFAULT now() gives,
	// and current_timestamp or clock_timestamp() too, while CURRENT_DATE does not.
	pgsLab5Recent = ` BETWEEN now() - interval '1 minute' AND now() + interval '1 minute'`
)

// Task 1. books.author_id, bigint, a foreign key to authors(id) with the
// default delete rule: a book cannot point at a missing author, can point at
// one that exists (so the key is on authors, not on some other table), and an
// author with books cannot be deleted (no CASCADE, no SET NULL).
var (
	pgsLab5AuthorRules = []pgsLab5Rule{
		{"книге указан автор с id 999999, которого нет",
			`UPDATE books SET author_id = 999999 WHERE id = ` + pgsLab5FirstBook, "23503"},
		{"книге указан новый автор из authors",
			`INSERT INTO authors (id, full_name) OVERRIDING SYSTEM VALUE VALUES (900001, 'Проверка');
UPDATE books SET author_id = 900001 WHERE id = ` + pgsLab5FirstBook, "00000"},
		{"удаление автора, у которого есть книга",
			`INSERT INTO authors (id, full_name) OVERRIDING SYSTEM VALUE VALUES (900001, 'Проверка');
UPDATE books SET author_id = 900001 WHERE id = ` + pgsLab5FirstBook + `;
DELETE FROM authors WHERE id = 900001`, "23503"},
	}
)

// pgsLab5AuthorShape checks only the new column: books' own primary key is
// setup's.
func pgsLab5AuthorShape() string {
	return `SELECT CASE WHEN f IS NULL THEN 'в books нет столбца author_id'
    WHEN f <> 'bigint' THEN 'у author_id тип ' || f || ', а нужен bigint' ELSE 'ok' END
FROM (SELECT (` + pgColType("books", "author_id") + `) AS f) x`
}

// pgsLab5AuthorNow: the column, then the rules.
var pgsLab5AuthorNow = `$({ d=$(glpg q pereplet ` + b64(pgsLab5AuthorShape()) + `); case $d in (ok) ` +
	pgsLab5Broken(pgsLab5AuthorRules) + `;; (*) echo -n $d;; esac; } | sed 's/; *$//')`

// Task 2. Every book of the catalogue once, with the author the task's table
// gives it, no book in books without an author, and every author in authors
// once: a second «Лев Толстой» added to point a book at is the duplicate this
// lab is about getting rid of. The diagnostic reads author_id through
// to_jsonb, so it works before task 1 too.
const pgsLab5Wrote = `WITH want (title, author) AS (VALUES
    ('Преступление и наказание', 'Фёдор Достоевский'),
    ('Мастер и Маргарита', 'Михаил Булгаков'),
    ('Вишнёвый сад', 'Антон Чехов'),
    ('Идиот', 'Фёдор Достоевский'),
    ('Каштанка', 'Антон Чехов'),
    ('Евгений Онегин', 'Александр Пушкин'),
    ('Анна Каренина', 'Лев Толстой'),
    ('Убийство в Восточном экспрессе', 'Агата Кристи'),
    ('Война и мир', 'Лев Толстой'),
    ('Капитанская дочка', 'Александр Пушкин')),
b AS (SELECT id, title, (to_jsonb(books) ->> 'author_id')::bigint AS author_id FROM books)
`

const pgsLab5Authors = pgsLab5Wrote + `SELECT (SELECT count(*) FROM pg_attribute WHERE attrelid = 'books'::regclass AND attname = 'author_id' AND NOT attisdropped) = 1
   AND NOT EXISTS (SELECT 1 FROM b WHERE author_id IS NULL)
   AND NOT EXISTS (SELECT 1 FROM want w WHERE (SELECT count(*) FROM b WHERE b.title = w.title) <> 1
        OR NOT EXISTS (SELECT 1 FROM b JOIN authors a ON a.id = b.author_id WHERE b.title = w.title AND a.full_name = w.author))
   AND NOT EXISTS (SELECT 1 FROM authors GROUP BY full_name HAVING count(*) > 1)`

const pgsLab5AuthorsNow = pgsLab5Wrote + `SELECT CASE
    WHEN NOT EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = 'books'::regclass AND attname = 'author_id' AND NOT attisdropped)
        THEN 'в books нет столбца author_id — его добавляет задание 1'
    ELSE 'книг без автора: ' || (SELECT count(*) FROM b WHERE author_id IS NULL) ||
        coalesce('; автор не тот, что в таблице задания, у книг: ' || (SELECT string_agg(DISTINCT '«' || b.title || '»', ', ')
            FROM b JOIN want w USING (title) LEFT JOIN authors a ON a.id = b.author_id
            WHERE b.author_id IS NOT NULL AND a.full_name IS DISTINCT FROM w.author), '') ||
        coalesce('; в каталоге не по одному разу: ' || (SELECT string_agg('«' || w.title || '»', ', ')
            FROM want w WHERE (SELECT count(*) FROM b WHERE b.title = w.title) <> 1), '') ||
        coalesce('; в authors не по одному разу: ' || (SELECT string_agg(full_name, ', ' ORDER BY full_name)
            FROM (SELECT full_name FROM authors GROUP BY full_name HAVING count(*) > 1) d), '')
    END`

// Task 3. «Война и мир» at 890.00, the other books at setup's prices (an
// UPDATE without WHERE is the mistake to catch), and a validated CHECK named
// books_price_positive that refuses 0 and a negative price but lets 0.01 in.
const pgsLab5Prices = `WITH known (title, price) AS (VALUES
    ('Преступление и наказание', 590.00), ('Мастер и Маргарита', 690.00), ('Вишнёвый сад', 280.00),
    ('Идиот', 540.00), ('Каштанка', 190.00), ('Евгений Онегин', 495.00), ('Анна Каренина', 790.00),
    ('Убийство в Восточном экспрессе', 520.00), ('Капитанская дочка', 260.00))
`

const pgsLab5Price = pgsLab5Prices + `SELECT (SELECT count(*) FROM books WHERE title = 'Война и мир') = 1
   AND EXISTS (SELECT 1 FROM books WHERE title = 'Война и мир' AND price = 890.00)
   AND NOT EXISTS (SELECT 1 FROM books b JOIN known k USING (title) WHERE b.price <> k.price)
   AND EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'books'::regclass AND contype = 'c'
        AND conname = 'books_price_positive' AND convalidated)`

const pgsLab5PriceNow = pgsLab5Prices + `SELECT 'цена «Войны и мира»: ' ||
    coalesce((SELECT string_agg(price::text, ', ') FROM books WHERE title = 'Война и мир'), 'такой книги нет') ||
    coalesce('; цена изменилась ещё у: ' || (SELECT string_agg(DISTINCT '«' || b.title || '»', ', ')
        FROM books b JOIN known k USING (title) WHERE b.price <> k.price), '') || '; ' ||
    coalesce('ограничения CHECK у books: ' || (SELECT string_agg(conname || CASE WHEN convalidated THEN '' ELSE ' NOT VALID' END, ', ' ORDER BY conname)
        FROM pg_constraint WHERE conrelid = 'books'::regclass AND contype = 'c'), 'ограничений CHECK у books нет')`

var pgsLab5PriceRules = []pgsLab5Rule{
	{"цена 0", `UPDATE books SET price = 0 WHERE id = ` + pgsLab5FirstBook, "23514"},
	{"цена -10", `UPDATE books SET price = -10 WHERE id = ` + pgsLab5FirstBook, "23514"},
	{"цена 0.01", `UPDATE books SET price = 0.01 WHERE id = ` + pgsLab5FirstBook, "00000"},
	{"цена 25000.00", `UPDATE books SET price = 25000.00 WHERE id = ` + pgsLab5FirstBook, "00000"},
}

const pgsLab5HasPriceCheck = `SELECT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid = 'books'::regclass AND conname = 'books_price_positive')`

// pgsLab5PriceFull is the diagnostic: prices and constraints, and the rules
// once the named constraint is there.
var pgsLab5PriceFull = `$({ glpg q pereplet ` + b64(pgsLab5PriceNow) + ` | tr -d '\n'; echo -n '; '; ` +
	`[ x$(glpg q pereplet ` + b64(pgsLab5HasPriceCheck) + `) = xt ] && { ` + pgsLab5Broken(pgsLab5PriceRules) + `}; } | sed 's/; *$//')`

// Task 4. customers as in §5: email required and unique, a name required, the
// city optional, registered_at required with now() by default.
const pgsLab5CustomerRow = `INSERT INTO customers (id, email, full_name, city, registered_at) OVERRIDING SYSTEM VALUE VALUES `

var (
	pgsLab5Customers = pgsLab5Table{name: "customers", pk: []string{"id"}, cols: []string{
		"id bigint", "email text", "full_name text", "city text", "registered_at timestamp with time zone"}}

	pgsLab5CustomerRules = []pgsLab5Rule{
		{"два покупателя с одним email",
			pgsLab5CustomerRow + `(900001, 'glpg-probe-900001@check.invalid', 'Проверка', 'Казань', now()),
    (900002, 'glpg-probe-900001@check.invalid', 'Проверка', 'Казань', now())`, "23505"},
		{"покупатель без email",
			pgsLab5CustomerRow + `(900001, NULL, 'Проверка', 'Казань', now())`, "23502"},
		{"покупатель без имени",
			pgsLab5CustomerRow + `(900001, 'glpg-probe-900001@check.invalid', NULL, 'Казань', now())`, "23502"},
		{"registered_at = NULL",
			pgsLab5CustomerRow + `(900001, 'glpg-probe-900001@check.invalid', 'Проверка', 'Казань', NULL)`, "23502"},
		{"покупатель без города и registered_at получает текущее время",
			`INSERT INTO customers (id, email, full_name) OVERRIDING SYSTEM VALUE VALUES (900001, 'glpg-probe-900001@check.invalid', 'Проверка');
SELECT 1/coalesce((SELECT registered_at` + pgsLab5Recent + ` FROM customers WHERE id = 900001)::int, 0)`, "00000"},
	}
)

// Task 5. orders: a required link to customers with the default delete rule,
// created_at required with now(), status required, one of four, 'new' by
// default.
const pgsLab5OrderRow = `INSERT INTO orders (id, customer_id, created_at, status) OVERRIDING SYSTEM VALUE VALUES `

var (
	pgsLab5Orders = pgsLab5Table{name: "orders", pk: []string{"id"}, cols: []string{
		"id bigint", "customer_id bigint", "created_at timestamp with time zone", "status text"}}

	pgsLab5OrderRules = []pgsLab5Rule{
		{"заказ покупателю с id 999999, которого нет",
			pgsLab5OrderRow + `(900001, 999999, now(), 'new')`, "23503"},
		{"заказ без покупателя",
			pgsLab5OrderRow + `(900001, NULL, now(), 'new')`, "23502"},
		{"заказ со статусом lost",
			pgsLab5ProbeCustomer + pgsLab5OrderRow + `(900001, 900001, now(), 'lost')`, "23514"},
		{"заказ со статусом NULL",
			pgsLab5ProbeCustomer + pgsLab5OrderRow + `(900001, 900001, now(), NULL)`, "23502"},
		{"заказы со статусами new, paid, shipped и cancelled",
			pgsLab5ProbeCustomer + pgsLab5OrderRow + `(900001, 900001, now(), 'new'), (900002, 900001, now(), 'paid'),
    (900003, 900001, now(), 'shipped'), (900004, 900001, now(), 'cancelled')`, "00000"},
		{"заказ без статуса и created_at получает new и текущее время",
			pgsLab5ProbeCustomer + `INSERT INTO orders (id, customer_id) OVERRIDING SYSTEM VALUE VALUES (900001, 900001);
SELECT 1/coalesce((SELECT status = 'new' AND created_at` + pgsLab5Recent + ` FROM orders WHERE id = 900001)::int, 0)`, "00000"},
		{"created_at = NULL",
			pgsLab5ProbeCustomer + pgsLab5OrderRow + `(900001, 900001, NULL, 'new')`, "23502"},
		{"удаление покупателя, у которого есть заказ",
			pgsLab5ProbeOrder + `DELETE FROM customers WHERE id = 900001`, "23503"},
	}
)

// Task 6. order_items: a pair key, quantity required and above zero, price
// required; deleting an order takes its items along, deleting a book that was
// ordered is refused.
var (
	pgsLab5Items = pgsLab5Table{name: "order_items", pk: []string{"order_id", "book_id"}, cols: []string{
		"order_id bigint", "book_id bigint", "quantity integer", "price numeric(8,2)"}}

	pgsLab5ItemRules = []pgsLab5Rule{
		{"позиция с количеством 0",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, ` + pgsLab5FirstBook + `, 0, 100)`, "23514"},
		{"позиция с количеством -1",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, ` + pgsLab5FirstBook + `, -1, 100)`, "23514"},
		{"позиция без количества",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, ` + pgsLab5FirstBook + `, NULL, 100)`, "23502"},
		{"позиция без цены",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, ` + pgsLab5FirstBook + `, 1, NULL)`, "23502"},
		{"одна книга дважды в одном заказе",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price)
    VALUES (900001, ` + pgsLab5FirstBook + `, 1, 100), (900001, ` + pgsLab5FirstBook + `, 2, 100)`, "23505"},
		{"две разные книги в одном заказе",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price)
    VALUES (900001, ` + pgsLab5FirstBook + `, 1, 100), (900001, ` + pgsLab5LastBook + `, 2, 100)`, "00000"},
		{"позиция заказа с id 999999, которого нет",
			`INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (999999, ` + pgsLab5FirstBook + `, 1, 100)`, "23503"},
		{"позиция с книгой id 999999, которой нет",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, 999999, 1, 100)`, "23503"},
		{"удаление заказа вместе с его позициями",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, ` + pgsLab5FirstBook + `, 1, 100);
DELETE FROM orders WHERE id = 900001;
SELECT 1/coalesce((SELECT count(*) = 0 FROM order_items WHERE order_id = 900001)::int, 0)`, "00000"},
		{"удаление книги, которая есть в заказе",
			pgsLab5ProbeOrder + `INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (900001, ` + pgsLab5FirstBook + `, 1, 100);
DELETE FROM books WHERE id = ` + pgsLab5FirstBook, "23503"},
	}
)

// Task 7. The first order: Ирина Соколова once, with one order in status new,
// and in it exactly two items with the books' prices copied in.
const pgsLab5OrderTables = `SELECT to_regclass('customers') IS NOT NULL AND to_regclass('orders') IS NOT NULL AND to_regclass('order_items') IS NOT NULL`

const pgsLab5Irina = `WITH c AS (SELECT * FROM customers WHERE email = 'irina.sokolova@example.com'),
o AS (SELECT * FROM orders WHERE customer_id IN (SELECT id FROM c)),
i AS (SELECT b.title, oi.quantity, oi.price FROM order_items oi JOIN books b ON b.id = oi.book_id
      WHERE oi.order_id IN (SELECT id FROM o))
`

const pgsLab5FirstOrder = pgsLab5Irina + `SELECT (SELECT count(*) FROM c) = 1
   AND EXISTS (SELECT 1 FROM c WHERE full_name = 'Ирина Соколова' AND city = 'Казань')
   AND (SELECT count(*) FROM o) = 1
   AND EXISTS (SELECT 1 FROM o WHERE status = 'new')
   AND (SELECT count(*) FROM i) = 2
   AND EXISTS (SELECT 1 FROM i WHERE title = 'Мастер и Маргарита' AND quantity = 1 AND price = 690.00)
   AND EXISTS (SELECT 1 FROM i WHERE title = 'Капитанская дочка' AND quantity = 2 AND price = 260.00)`

const pgsLab5FirstOrderNow = pgsLab5Irina + `SELECT 'покупателей с email irina.sokolova@example.com: ' || (SELECT count(*) FROM c) ||
    coalesce(' — ' || (SELECT string_agg(full_name || ', ' || coalesce(city, 'город ∅'), '; ') FROM c), '') ||
    '; заказов у неё: ' || (SELECT count(*) FROM o) ||
    coalesce(', статус ' || (SELECT string_agg(status, ', ' ORDER BY id) FROM o), '') ||
    '; позиций в заказе: ' || (SELECT count(*) FROM i) ||
    coalesce(' — ' || (SELECT string_agg('«' || title || '» × ' || quantity || ' по ' || price, ', ' ORDER BY title) FROM i), '')`

var pgsLab5FirstOrderFull = `$(if [ x$(glpg q pereplet ` + b64(pgsLab5OrderTables) + `) = xt ]; then glpg q pereplet ` +
	b64(pgsLab5FirstOrderNow) + `; else echo нет таблиц customers, orders и order_items — их создают задания 4–6; fi)`

// Task 8 ★. genres and book_genres as in §5, and three links: «Мастер и
// Маргарита» in two genres, «Роман» on two books — both sides of
// many-to-many. Deleting a book takes its links along; a genre in use cannot
// be deleted. The book the delete probe removes is a fresh copy rather than
// one of the catalogue: a book in an order could not be deleted at all.
var (
	pgsLab5Genres     = pgsLab5Table{name: "genres", pk: []string{"id"}, cols: []string{"id bigint", "name text"}}
	pgsLab5BookGenres = pgsLab5Table{name: "book_genres", pk: []string{"book_id", "genre_id"}, cols: []string{
		"book_id bigint", "genre_id bigint"}}

	pgsLab5GenreRules = []pgsLab5Rule{
		{"два жанра с одним названием",
			`INSERT INTO genres (id, name) OVERRIDING SYSTEM VALUE VALUES (900001, 'glpg: жанр'), (900002, 'glpg: жанр')`, "23505"},
		{"жанр без названия",
			`INSERT INTO genres (id, name) OVERRIDING SYSTEM VALUE VALUES (900001, NULL)`, "23502"},
		{"одна пара книга + жанр дважды",
			pgsLab5ProbeGenre + `INSERT INTO book_genres (book_id, genre_id) VALUES (` + pgsLab5FirstBook + `, 900001), (` + pgsLab5FirstBook + `, 900001)`, "23505"},
		{"книга в двух жанрах и жанр у двух книг",
			pgsLab5ProbeGenre + `INSERT INTO genres (id, name) OVERRIDING SYSTEM VALUE VALUES (900002, 'glpg: жанр 900002');
INSERT INTO book_genres (book_id, genre_id)
    VALUES (` + pgsLab5FirstBook + `, 900001), (` + pgsLab5FirstBook + `, 900002), (` + pgsLab5LastBook + `, 900001)`, "00000"},
		{"связь с жанром id 999999, которого нет",
			`INSERT INTO book_genres (book_id, genre_id) VALUES (` + pgsLab5FirstBook + `, 999999)`, "23503"},
		{"связь с книгой id 999999, которой нет",
			pgsLab5ProbeGenre + `INSERT INTO book_genres (book_id, genre_id) VALUES (999999, 900001)`, "23503"},
		{"удаление книги вместе с её жанрами",
			pgsLab5ProbeGenre + pgsLab5ProbeBook + `
INSERT INTO book_genres (book_id, genre_id) VALUES (900001, 900001);
DELETE FROM books WHERE id = 900001;
SELECT 1/coalesce((SELECT count(*) = 0 FROM book_genres WHERE book_id = 900001)::int, 0)`, "00000"},
		{"удаление жанра, к которому привязана книга",
			pgsLab5ProbeGenre + `INSERT INTO book_genres (book_id, genre_id) VALUES (` + pgsLab5FirstBook + `, 900001);
DELETE FROM genres WHERE id = 900001`, "23503"},
	}
)

const pgsLab5GenreOf = `(SELECT array_agg(g.name ORDER BY g.name) FROM book_genres bg
    JOIN genres g ON g.id = bg.genre_id JOIN books b ON b.id = bg.book_id WHERE b.title = `

const pgsLab5Linked = `SELECT ` + pgsLab5GenreOf + `'Мастер и Маргарита') = ARRAY['Роман', 'Фантастика']
   AND ` + pgsLab5GenreOf + `'Анна Каренина') = ARRAY['Роман']`

const pgsLab5LinkedNow = `SELECT 'жанры «Мастера и Маргариты»: ' || coalesce(array_to_string(` + pgsLab5GenreOf + `'Мастер и Маргарита'), ', '), 'нет') ||
    '; жанры «Анны Карениной»: ' || coalesce(array_to_string(` + pgsLab5GenreOf + `'Анна Каренина'), ', '), 'нет')`

const pgsLab5GenreTables = `SELECT to_regclass('genres') IS NOT NULL AND to_regclass('book_genres') IS NOT NULL`

// pgsLab5GenresFull: the tables and their rules, then the links once both
// tables exist.
var pgsLab5GenresFull = pgsLab5TablesNow([]pgsLab5Table{pgsLab5Genres, pgsLab5BookGenres}, pgsLab5GenreRules,
	`[ x$(glpg q pereplet `+b64(pgsLab5GenreTables)+`) = xt ] && { glpg q pereplet `+b64(pgsLab5LinkedNow)+` | tr -d '\n'; echo -n '; '; }; `)

func init() {
	pgStartLabs["ch-pgs-lab5"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + "mkdir -p /root/sql\n" + pgRun("pereplet", pgsLab5Pereplet)),
		Checks: map[int]string{
			1: check(pgIs("pereplet", pgsLab5AuthorShape(), "ok")+" && "+pgsLab5Hold(pgsLab5AuthorRules),
				"в books есть столбец author_id (bigint) — внешний ключ на authors (id), без ON DELETE",
				"Сейчас: "+pgsLab5AuthorNow),
			2: check(pgTrue("pereplet", pgsLab5Authors),
				"у каждой книги в author_id — её автор из таблицы задания, пустых author_id нет, каждый автор в authors записан один раз",
				"Сейчас: "+pgVal("pereplet", pgsLab5AuthorsNow)),
			3: check(pgTrue("pereplet", pgsLab5Price)+" && "+pgsLab5Hold(pgsLab5PriceRules),
				"цена «Войны и мира» — 890.00, у остальных книг прежняя; ограничение books_price_positive не пускает цену 0 и ниже",
				"Сейчас: "+pgsLab5PriceFull),
			4: check(pgsLab5Customers.ok()+" && "+pgsLab5Hold(pgsLab5CustomerRules),
				"таблица customers как в задании: столбцы, типы, email обязателен и не повторяется, имя обязательно, registered_at обязателен и по умолчанию — текущее время",
				"Сейчас: "+pgsLab5TablesNow([]pgsLab5Table{pgsLab5Customers}, pgsLab5CustomerRules, "")),
			5: check(pgsLab5Orders.ok()+" && "+pgsLab5Hold(pgsLab5OrderRules),
				"таблица orders как в задании: столбцы, типы, обязательная ссылка на покупателя без ON DELETE, статус обязателен и один из четырёх, по умолчанию new и текущее время",
				"Сейчас: "+pgsLab5TablesNow([]pgsLab5Table{pgsLab5Orders}, pgsLab5OrderRules, "")),
			6: check(pgsLab5Items.ok()+" && "+pgsLab5Hold(pgsLab5ItemRules),
				"таблица order_items как в задании: столбцы, типы, ключ из пары заказ + книга, количество больше нуля, позиции удаляются вместе с заказом, ссылка на книгу без ON DELETE",
				"Сейчас: "+pgsLab5TablesNow([]pgsLab5Table{pgsLab5Items}, pgsLab5ItemRules, "")),
			7: check(pgTrue("pereplet", pgsLab5FirstOrder),
				"у Ирины Соколовой (irina.sokolova@example.com, Казань) один заказ, статус new: «Мастер и Маргарита» × 1 по 690.00 и «Капитанская дочка» × 2 по 260.00",
				"Сейчас: "+pgsLab5FirstOrderFull),
			8: check(pgsLab5Genres.ok()+" && "+pgsLab5BookGenres.ok()+" && "+pgsLab5Hold(pgsLab5GenreRules)+" && "+pgTrue("pereplet", pgsLab5Linked),
				"таблицы genres и book_genres как в задании; у «Мастера и Маргариты» жанры Роман и Фантастика, у «Анны Карениной» — Роман",
				"Сейчас: "+pgsLab5GenresFull),
		},
	}
}
