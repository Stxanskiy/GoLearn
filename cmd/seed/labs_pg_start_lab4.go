package main

// Лабораторная 4 курса pg-start: наполняем магазин (ch-pgs-lab4).
//
// pereplet is where lab 3 left it — the "first version" of §5 of the course
// contract: no NOT NULL and no defaults, not even on in_stock, so a column left
// out of an INSERT really is empty. The catalogue a colleague carried over from Excel
// is in it with Excel's mistakes, and the student adds authors and books
// (INSERT, several rows at once, RETURNING), changes prices and stock (UPDATE
// by a formula, several columns, several rows), removes duplicates (DELETE) and
// moves the supplier's delivery from staging.new_books (INSERT … SELECT),
// logging the load in staging.import_log.
//
// Two tasks are not in the contract's outline. Both come from the "updates"
// set of PostgreSQL Exercises, whose ideas the contract asks for: task 5 fixes
// two columns of one book in one UPDATE (the "update multiple columns"
// exercise), and the ★ task also logs the load with now() and the row count
// (the "insert calculated data" exercise) — import_log, made in lab 3, gets
// its first row.
//
// The checks look at the data, never at the command: a typed constant passes
// where a formula was asked for, separate INSERTs pass where one was asked for.
// What they do insist on is that nothing else moved — every task that changes
// rows also checks that the rows it should not touch are as setup left them,
// because UPDATE and DELETE without WHERE are what this lab is about. Rows are
// recognised by title, except the duplicates, which only their ids tell apart
// (setup fills fresh tables, so ids 1…11 are fixed). Expected values are
// written out here rather than read back from setup's tables, so a student who
// empties staging.new_books after moving it still passes.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab4.sh.

// pgsLab4Pereplet is the database as lab 3 leaves it, plus what this lab works
// on. books keeps lab 3's column order: isbn was added last, after in_stock, so
// "INSERT INTO books SELECT * FROM staging.new_books" cannot work by accident.
//
// The catalogue: «Мастер и Маргарита» three times, the original at id 2, so
// finding the smallest id means looking rather than assuming 1; three books under 100
// pages; «Шинель» with no page count (the NULL a careless condition catches);
// «Идиот», thick and already out of stock (what "set in_stock from the page
// count" would wrongly turn back on); «Ревизор» with transposed digits. Prices
// are round, so a price*1.1 run on every row by mistake is undone exactly by
// price/1.1.
const pgsLab4Pereplet = `
CREATE TABLE authors (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name  text,
    birth_year integer,
    country    text
);

CREATE TABLE books (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title          text,
    price          numeric(8,2),
    page_count     integer,
    published_year integer,
    in_stock       boolean,
    isbn           text
);

CREATE SCHEMA staging;

CREATE TABLE staging.import_log (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    loaded_at timestamptz,
    row_count integer
);

-- Новые поступления от поставщика: задание 8 переносит их в books.
CREATE TABLE staging.new_books (
    title          text,
    price          numeric(8,2),
    page_count     integer,
    published_year integer,
    isbn           text
);

INSERT INTO authors (full_name, birth_year, country) VALUES
    ('Михаил Булгаков',   1891, 'Россия'),
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Антон Чехов',       1860, 'Россия'),
    ('Николай Гоголь',    1809, 'Россия');

INSERT INTO books (title, price, page_count, published_year, in_stock, isbn) VALUES
    ('Преступление и наказание', 590.00,  608, 1866, true,  '978-5-389-04926-0'),
    ('Мастер и Маргарита',       690.00,  480, 1967, true,  '978-5-389-01686-6'),
    ('Вишнёвый сад',             280.00,   96, 1904, true,  '978-5-389-06010-4'),
    ('Идиот',                    540.00,  640, 1869, false, '978-5-389-05178-2'),
    ('Ревизор',                  240.00,  821, 1863, true,  '978-5-389-08562-6'),
    ('Мастер и Маргарита',       690.00,  480, 1967, true,  '978-5-389-01686-6'),
    ('Каштанка',                 190.00,   48, 1887, true,  '978-5-389-10364-1'),
    ('Мёртвые души',             450.00,  352, 1842, true,  '978-5-389-04720-4'),
    ('Шинель',                   260.00, NULL, 1842, true,  '978-5-389-07123-0'),
    ('Белые ночи',               220.00,   80, 1848, true,  '978-5-389-08890-0'),
    ('Мастер и Маргарита',       690.00,  480, 1967, true,  NULL);

INSERT INTO staging.new_books (title, price, page_count, published_year, isbn) VALUES
    ('Анна Каренина',                  790.00, 864, 1878, '978-5-389-21580-1'),
    ('Убийство в Восточном экспрессе', 520.00, 256, 1934, '978-5-04-112380-2'),
    ('451 градус по Фаренгейту',       480.00, 256, 1953, '978-5-04-107463-0');
`

// pgsLab4Known is every book the lab knows the values of, as setup (or, for
// the delivery, staging.new_books) has them: «Евгений Онегин», which the
// student adds, is not here. «Ревизор» is here with its wrong page count and
// year — only its price is compared against this. A CTE to put in front of a
// check's SELECT.
const pgsLab4Known = `WITH known (title, price, page_count, published_year) AS (VALUES
    ('Преступление и наказание', 590.00, 608, 1866),
    ('Мастер и Маргарита', 690.00, 480, 1967),
    ('Вишнёвый сад', 280.00, 96, 1904),
    ('Идиот', 540.00, 640, 1869),
    ('Ревизор', 240.00, 821, 1863),
    ('Каштанка', 190.00, 48, 1887),
    ('Мёртвые души', 450.00, 352, 1842),
    ('Шинель', 260.00, NULL, 1842),
    ('Белые ночи', 220.00, 80, 1848),
    ('Анна Каренина', 790.00, 864, 1878),
    ('Убийство в Восточном экспрессе', 520.00, 256, 1934),
    ('451 градус по Фаренгейту', 480.00, 256, 1953))
`

// pgsLab4Others lists the setup titles other than the duplicated one: the
// books the DELETE task must leave in place.
const pgsLab4Others = `('Преступление и наказание', 'Вишнёвый сад', 'Идиот', 'Ревизор', 'Каштанка', 'Мёртвые души', 'Шинель', 'Белые ночи')`

// pgsLab4Delivery is staging.new_books as setup fills it, for the ★ task.
const pgsLab4Delivery = `WITH want (title, price, page_count, published_year, isbn) AS (VALUES
    ('Анна Каренина', 790.00, 864, 1878, '978-5-389-21580-1'),
    ('Убийство в Восточном экспрессе', 520.00, 256, 1934, '978-5-04-112380-2'),
    ('451 градус по Фаренгейту', 480.00, 256, 1953, '978-5-04-107463-0'))
`

// pgsLab4Moved is the condition "this delivery row is in books as it should
// be": same values, in stock.
const pgsLab4Moved = `EXISTS (SELECT 1 FROM books b WHERE b.title = w.title AND b.price = w.price
    AND b.page_count = w.page_count AND b.published_year = w.published_year AND b.isbn = w.isbn AND b.in_stock)`

// pgsLab4Author shows an author row the way a diagnostic lists it, NULL as ∅
// like the course's ~/.psqlrc shows it.
const pgsLab4Author = `concat_ws(', ', coalesce(full_name, '∅'), coalesce(birth_year::text, '∅'), coalesce(country, '∅'))`

// Task 1. Exactly one Пушкин, with his year and country.
const pgsLab4Pushkin = `SELECT count(*) = 1 AND coalesce(bool_and(birth_year = 1799 AND country = 'Россия'), false)
FROM authors WHERE full_name = 'Александр Пушкин'`

const pgsLab4PushkinNow = `SELECT 'в authors строк: ' || count(*) || coalesce('; похожие на Пушкина: ' ||
    string_agg(` + pgsLab4Author + `, '; ' ORDER BY id) FILTER (WHERE full_name ILIKE '%пушкин%'), ', Пушкина среди них нет')
FROM authors`

// Task 2. The three authors once each with their data, and no row in authors
// beyond the four from setup, Пушкин and these three: a row with a typo is
// left over from a mistake the student is expected to clean up.
const pgsLab4Three = `WITH want (full_name, birth_year, country) AS (VALUES
    ('Лев Толстой', 1828, 'Россия'), ('Агата Кристи', 1890, 'Великобритания'), ('Рэй Брэдбери', 1920, 'США'))
SELECT NOT EXISTS (SELECT 1 FROM want w
        WHERE (SELECT count(*) FROM authors a WHERE a.full_name = w.full_name) <> 1
           OR NOT EXISTS (SELECT 1 FROM authors a WHERE a.full_name = w.full_name AND a.birth_year = w.birth_year AND a.country = w.country))
   AND NOT EXISTS (SELECT 1 FROM authors WHERE full_name IS NULL OR full_name NOT IN
        ('Михаил Булгаков', 'Фёдор Достоевский', 'Антон Чехов', 'Николай Гоголь',
         'Александр Пушкин', 'Лев Толстой', 'Агата Кристи', 'Рэй Брэдбери'))`

const pgsLab4ThreeNow = `SELECT 'в authors, кроме четырёх авторов из Excel: ' ||
    coalesce(string_agg(` + pgsLab4Author + `, '; ' ORDER BY id), 'никого')
FROM authors WHERE id > 4`

// Task 3. One «Евгений Онегин» with the given values — the price may already
// be the raised one of task 4 — and the file holds its id and nothing else.
const pgsLab4Onegin = `SELECT count(*) = 1 AND coalesce(bool_and(price IN (450.00, 495.00) AND page_count = 320
    AND published_year = 1833 AND in_stock), false)
FROM books WHERE title = 'Евгений Онегин'`

const pgsLab4OneginID = `SELECT id FROM books WHERE title = 'Евгений Онегин'`

const pgsLab4OneginNow = `SELECT 'книг «Евгений Онегин» в books: ' || count(*) || coalesce(' (' || string_agg(concat_ws(', ',
        coalesce(price::text, '∅') || ' руб.', coalesce(page_count::text, '∅') || ' стр.',
        coalesce(published_year::text, '∅') || ' г.', 'in_stock ' || coalesce(in_stock::text, '∅')), '; ' ORDER BY id) || ')', '')
FROM books WHERE title = 'Евгений Онегин'`

// pgsLab4OneginOnly is the id of «Евгений Онегин» when there is exactly one,
// and empty otherwise: the file is said to hold another number only when there
// is one book to compare it with.
const pgsLab4OneginOnly = `SELECT CASE WHEN count(*) = 1 THEN min(id)::text ELSE '' END FROM books WHERE title = 'Евгений Онегин'`

// pgsLab4FileNow shows /root/book_id.txt line by line, so the "INSERT 0 1"
// psql adds after RETURNING is visible, and says whether the book's id is in
// it at all — without telling the id. No double quotes: check() escapes
// them in the message, and an escaped quote inside $(…) breaks it; hence the
// x$f comparisons.
var pgsLab4FileNow = `$([ -s /root/book_id.txt ] && head -n 3 /root/book_id.txt | sed 's/[[:space:]]*$//; s/.*/«&»/' | paste -sd ' ' | sed 's|^|в /root/book_id.txt строки: |' || echo файла /root/book_id.txt нет или он пустой)` +
	`$(f=$(tr -d '[:space:]' 2>/dev/null < /root/book_id.txt); i=$(glpg q pereplet ` + b64(pgsLab4OneginOnly) + `); [ x$i = x ] || [ x$f = x ] || [ x$f = x$i ] || { sed 's/[[:space:]]//g' /root/book_id.txt | grep -qx -e $i && echo ' — id книги в файле есть, но кроме него есть и лишние строки' || echo ' — это не id книги «Евгений Онегин»'; })`

// Task 4. «Евгений Онегин» at 495.00, every other known book at its price.
const pgsLab4Price = pgsLab4Known + `SELECT (SELECT count(*) FROM books WHERE title = 'Евгений Онегин') = 1
   AND EXISTS (SELECT 1 FROM books WHERE title = 'Евгений Онегин' AND price = 495.00)
   AND NOT EXISTS (SELECT 1 FROM books b JOIN known k USING (title) WHERE b.price IS DISTINCT FROM k.price)`

const pgsLab4PriceNow = pgsLab4Known + `SELECT 'цена «Евгения Онегина»: ' ||
    coalesce((SELECT string_agg(coalesce(price::text, '∅'), ', ' ORDER BY id) FROM books WHERE title = 'Евгений Онегин'), 'такой книги нет') ||
    coalesce('; цена изменилась ещё у: ' || (SELECT string_agg(DISTINCT b.title, ', ') FROM books b JOIN known k USING (title)
        WHERE b.price IS DISTINCT FROM k.price), '; цены остальных книг прежние')`

// Task 5. «Ревизор» fixed in both columns, the other books' page counts and
// years as they were.
const pgsLab4Revizor = pgsLab4Known + `SELECT (SELECT count(*) FROM books WHERE title = 'Ревизор') = 1
   AND EXISTS (SELECT 1 FROM books WHERE title = 'Ревизор' AND page_count = 128 AND published_year = 1836)
   AND NOT EXISTS (SELECT 1 FROM books b JOIN known k USING (title) WHERE k.title <> 'Ревизор'
        AND (b.page_count, b.published_year) IS DISTINCT FROM (k.page_count, k.published_year))`

const pgsLab4RevizorNow = pgsLab4Known + `SELECT 'у «Ревизора» ' ||
    coalesce((SELECT string_agg(coalesce(page_count::text, '∅') || ' стр. и ' || coalesce(published_year::text, '∅') || ' год', ', ' ORDER BY id)
        FROM books WHERE title = 'Ревизор'), '— такой книги нет') ||
    coalesce('; число страниц или год изменились ещё у: ' || (SELECT string_agg(DISTINCT b.title, ', ') FROM books b JOIN known k USING (title)
        WHERE k.title <> 'Ревизор' AND (b.page_count, b.published_year) IS DISTINCT FROM (k.page_count, k.published_year)),
        '; у остальных книг прежние')`

// Task 6. Every book under 100 pages out of stock — the three from setup still
// there — while «Шинель» (no page count), «Идиот» (already out) and the thick
// books in stock keep what they had.
const pgsLab4Thin = `SELECT NOT EXISTS (SELECT 1 FROM books WHERE page_count < 100 AND in_stock IS DISTINCT FROM false)
   AND (SELECT count(*) FROM books WHERE title IN ('Вишнёвый сад', 'Каштанка', 'Белые ночи') AND in_stock = false) = 3
   AND (SELECT bool_and(in_stock IS TRUE) FROM books WHERE title = 'Шинель') IS TRUE
   AND (SELECT bool_and(in_stock IS FALSE) FROM books WHERE title = 'Идиот') IS TRUE
   AND NOT EXISTS (SELECT 1 FROM books WHERE title IN ('Преступление и наказание', 'Мастер и Маргарита', 'Ревизор', 'Мёртвые души')
        AND in_stock IS DISTINCT FROM true)`

const pgsLab4ThinNow = `SELECT 'книг тоньше 100 страниц: ' || count(*) FILTER (WHERE page_count < 100) ||
    ', из них в наличии: ' || count(*) FILTER (WHERE page_count < 100 AND in_stock) ||
    '; не в наличии (false): ' || coalesce(string_agg(title || ' — ' || coalesce(page_count || ' стр.', 'страниц ∅'), ', ' ORDER BY title)
        FILTER (WHERE in_stock = false), 'ни одной книги') ||
    coalesce('; in_stock пустой (NULL) у: ' || string_agg(title, ', ' ORDER BY title) FILTER (WHERE in_stock IS NULL), '')
FROM books`

// Task 7. One «Мастер и Маргарита», the original at id 2; the duplicates (6
// and 11) gone, not renamed; the other setup books in place.
const pgsLab4Dupes = `SELECT (SELECT count(*) FROM books WHERE title = 'Мастер и Маргарита') = 1
   AND EXISTS (SELECT 1 FROM books WHERE id = 2 AND title = 'Мастер и Маргарита')
   AND NOT EXISTS (SELECT 1 FROM books WHERE id IN (6, 11))
   AND (SELECT count(DISTINCT title) FROM books WHERE title IN ` + pgsLab4Others + `) = 8`

const pgsLab4DupesNow = `SELECT 'строк «Мастер и Маргарита»: ' || count(*) FILTER (WHERE title = 'Мастер и Маргарита') ||
    coalesce(' (id ' || string_agg(id::text, ', ' ORDER BY id) FILTER (WHERE title = 'Мастер и Маргарита') || ')', '') ||
    coalesce('; под номерами дублей теперь: ' || string_agg(coalesce(title, '∅') || ' (id ' || id || ')', ', ' ORDER BY id)
        FILTER (WHERE id IN (6, 11) AND title IS DISTINCT FROM 'Мастер и Маргарита'), '') ||
    '; остальных книг из Excel на месте: ' || count(DISTINCT title) FILTER (WHERE title IN ` + pgsLab4Others + `) || ' из 8'
FROM books`

// Task 8 ★. Each delivered book once in books with staging's values and in
// stock, and one row in the load log: three rows, logged at a real moment —
// not in the future and not a date copied from an example months ago.
const pgsLab4Staging = pgsLab4Delivery + `SELECT NOT EXISTS (SELECT 1 FROM want w
        WHERE (SELECT count(*) FROM books b WHERE b.title = w.title) <> 1 OR NOT ` + pgsLab4Moved + `)
   AND (SELECT count(*) FROM staging.import_log) = 1
   AND EXISTS (SELECT 1 FROM staging.import_log WHERE row_count = 3
        AND loaded_at > now() - interval '30 days' AND loaded_at <= now() + interval '1 minute')`

const pgsLab4StagingNow = pgsLab4Delivery + `SELECT 'в books ' ||
    (SELECT string_agg('«' || w.title || '» — ' || (SELECT count(*) FROM books b WHERE b.title = w.title) || ' шт.', ', ' ORDER BY w.title) FROM want w) ||
    '; совпадают со staging.new_books по цене, страницам, году и ISBN и в наличии: ' ||
    (SELECT count(*) FROM want w WHERE ` + pgsLab4Moved + `) || ' из 3' ||
    '; в staging.import_log записей: ' || (SELECT count(*) FROM staging.import_log) ||
    coalesce(' (' || (SELECT string_agg('row_count ' || coalesce(row_count::text, '∅') || ', loaded_at ' ||
        coalesce(to_char(loaded_at, 'YYYY-MM-DD HH24:MI'), '∅'), '; ' ORDER BY id) FROM staging.import_log) || ')', '')`

func init() {
	pgStartLabs["ch-pgs-lab4"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + "mkdir -p /root/sql\n" + pgRun("pereplet", pgsLab4Pereplet)),
		Checks: map[int]string{
			1: check(pgTrue("pereplet", pgsLab4Pushkin),
				"в authors ровно одна строка «Александр Пушкин», 1799, Россия",
				"Сейчас "+pgVal("pereplet", pgsLab4PushkinNow)),
			2: check(pgTrue("pereplet", pgsLab4Three),
				"в authors по одному разу Лев Толстой (1828, Россия), Агата Кристи (1890, Великобритания), Рэй Брэдбери (1920, США) и нет лишних строк",
				"Сейчас "+pgVal("pereplet", pgsLab4ThreeNow)),
			3: check(pgTrue("pereplet", pgsLab4Onegin)+` && [ "$(tr -d '[:space:]' 2>/dev/null < /root/book_id.txt)" = "`+pgVal("pereplet", pgsLab4OneginID)+`" ]`,
				"в books одна книга «Евгений Онегин» (450.00, 320 стр., 1833, в наличии), а в /root/book_id.txt — только её id",
				"Сейчас "+pgVal("pereplet", pgsLab4OneginNow)+"; "+pgsLab4FileNow),
			4: check(pgTrue("pereplet", pgsLab4Price),
				"у «Евгения Онегина» цена 495.00, у остальных книг цены прежние",
				"Сейчас "+pgVal("pereplet", pgsLab4PriceNow)),
			5: check(pgTrue("pereplet", pgsLab4Revizor),
				"у «Ревизора» 128 страниц и 1836 год, у остальных книг число страниц и год прежние",
				"Сейчас "+pgVal("pereplet", pgsLab4RevizorNow)),
			6: check(pgTrue("pereplet", pgsLab4Thin),
				"книги тоньше 100 страниц не в наличии; книга без числа страниц и остальные книги — как были",
				"Сейчас "+pgVal("pereplet", pgsLab4ThinNow)),
			7: check(pgTrue("pereplet", pgsLab4Dupes),
				"в books одна «Мастер и Маргарита» — с наименьшим id из трёх, остальные книги на месте",
				"Сейчас "+pgVal("pereplet", pgsLab4DupesNow)),
			8: check(pgTrue("pereplet", pgsLab4Staging),
				"три книги из staging.new_books — в books по одному разу, с теми же данными и в наличии; в staging.import_log одна запись: row_count 3 и время загрузки",
				"Сейчас "+pgVal("pereplet", pgsLab4StagingNow)),
		},
	}
}
