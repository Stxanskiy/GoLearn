package main

// Лабораторная 6 курса pg-start: импорт каталога (ch-pgs-lab6).
//
// pereplet is the canonical schema of §5 of the course contract — what lab 5
// leaves — with every table empty, and /root/import holds what the shop
// exported from Excel, its CRM and its old website. The student loads the files
// with \copy, finds and fixes the one row that does not load, exports a report
// for the warehouse, keeps the website's genre numbers and fixes the identity
// counter after them, and generates test customers with generate_series.
//
// One task is not in the contract's outline: genres.csv brings its own ids
// (task 5). Chapter 15 teaches what that does to an identity column — COPY
// writes the ids as they are and the counter stays at 1, so the next INSERT
// without an id fails — and the course project (lab 7) depends on the fix, so
// the student meets it here first. The ids have gaps, as ids of a live site
// do: retrying the INSERT until it goes through lands on a free gap (4), the
// counter is still behind max(id), and the check says so.
//
// The checks compare the data with what the files hold, written out here: a
// student may edit the files (task 3 asks for it), so the files themselves
// are no reference. They look at the result, never at the command — \copy,
// COPY … FROM STDIN fed by the shell, a server-side COPY from a copy in /tmp
// or plain INSERTs all pass; customers.csv may be fixed in the file or loaded
// with a DMY datestyle. The genre check reads the identity counter without
// calling nextval(), so pressing "Проверить" again and again does not move it
// past the taken numbers on the student's behalf.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab6.sh.

// pgsLab6Schema is §5 of the course contract, the database after lab 5.
const pgsLab6Schema = `
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

CREATE TABLE genres (
    id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE book_genres (
    book_id  bigint REFERENCES books (id) ON DELETE CASCADE,
    genre_id bigint REFERENCES genres (id),
    PRIMARY KEY (book_id, genre_id)
);

CREATE TABLE customers (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         text NOT NULL UNIQUE,
    full_name     text NOT NULL,
    city          text,
    registered_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id bigint NOT NULL REFERENCES customers (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    status      text NOT NULL DEFAULT 'new'
                CHECK (status IN ('new', 'paid', 'shipped', 'cancelled'))
);

CREATE TABLE order_items (
    order_id bigint REFERENCES orders (id) ON DELETE CASCADE,
    book_id  bigint REFERENCES books (id),
    quantity integer NOT NULL CHECK (quantity > 0),
    price    numeric(8,2) NOT NULL,
    PRIMARY KEY (order_id, book_id)
);
`

// The files in /root/import. authors.csv is the plain case: commas, a header,
// the table's own column order without id.
const pgsLab6AuthorsCSV = `full_name,birth_year,country
Александр Пушкин,1799,Россия
Николай Гоголь,1809,Россия
Фёдор Достоевский,1821,Россия
Лев Толстой,1828,Россия
Антон Чехов,1860,Россия
Михаил Булгаков,1891,Россия
Агата Кристи,1890,Великобритания
Эрнест Хемингуэй,1899,США
Джордж Оруэлл,1903,Великобритания
Рэй Брэдбери,1920,США
`

// books.csv comes from another program: semicolons, isbn first, so neither
// the default delimiter nor the table's column order loads it. «Шинель» has no
// page count (an empty field, which CSV format reads as NULL). Five books are
// out of stock, among them «Прощай, оружие!», whose comma the export must quote,
// and «451 градус по Фаренгейту», which sorts before the Cyrillic titles.
const pgsLab6BooksCSV = `isbn;title;price;page_count;published_year;in_stock
978-5-389-07435-4;Евгений Онегин;450.00;320;1833;true
978-5-389-04720-4;Мёртвые души;450.00;352;1842;true
978-5-389-07123-0;Шинель;260.00;;1842;true
978-5-389-04926-0;Преступление и наказание;590.00;608;1866;true
978-5-389-05178-2;Идиот;540.00;640;1869;false
978-5-389-21580-1;Анна Каренина;790.00;864;1878;true
978-5-389-10364-1;Каштанка;190.00;48;1887;false
978-5-389-06010-4;Вишнёвый сад;280.00;96;1904;true
978-5-389-01686-6;Мастер и Маргарита;690.00;480;1967;true
978-5-04-112380-2;Убийство в Восточном экспрессе;520.00;256;1934;true
978-5-17-090631-1;Прощай, оружие!;560.00;320;1929;false
978-5-17-094367-5;Скотный двор;350.00;128;1945;false
978-5-17-080115-9;1984;520.00;320;1949;true
978-5-04-107463-0;451 градус по Фаренгейту;480.00;256;1953;false
`

// customers.csv: line 9 has the date the way a Russian spreadsheet writes it,
// which the sandbox's DateStyle (ISO, MDY) rejects; one customer has no city.
// The customers are made up, and so are their addresses (example.* domains).
const pgsLab6CustomersCSV = `email,full_name,city,registered_at
anna.smirnova@example.org,Анна Смирнова,Москва,2023-02-14
igor.petrov@example.net,Игорь Петров,Казань,2023-03-02
olga.kuznetsova@example.org,Ольга Кузнецова,Санкт-Петербург,2023-04-19
d.volkov@example.com,Дмитрий Волков,Екатеринбург,2023-06-07
maria.sokolova@example.net,Мария Соколова,Новосибирск,2023-08-23
sergey.morozov@example.org,Сергей Морозов,,2023-09-11
elena.pavlova@example.com,Елена Павлова,Казань,2023-11-05
n.fedorov@example.org,Николай Федоров,Москва,31.12.2023
tatiana.belova@example.net,Татьяна Белова,Самара,2024-01-16
a.nikitin@example.com,Алексей Никитин,Нижний Новгород,2024-02-28
irina.zaitseva@example.org,Ирина Зайцева,Санкт-Петербург,2024-03-09
p.orlov@example.net,Павел Орлов,Краснодар,2024-04-21
`

// genres.csv: the old website's genres with their ids, gaps included.
const pgsLab6GenresCSV = `id,name
1,Роман
2,Повесть
3,Рассказ
5,Пьеса
6,Поэзия
8,Детектив
9,Фантастика
12,Антиутопия
`

// pgsLab6File is a setup step writing content to path. Base64, like the SQL
// of the checks, so no line of the data can end a heredoc early.
func pgsLab6File(path, content string) string {
	return "echo " + b64(content) + " | base64 -d > " + path + "\n"
}

// Task 1. Every author of the file exactly once, and nothing else.
const pgsLab6AuthorsWant = `WITH want (full_name, birth_year, country) AS (VALUES
    ('Александр Пушкин', 1799, 'Россия'),
    ('Николай Гоголь', 1809, 'Россия'),
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Лев Толстой', 1828, 'Россия'),
    ('Антон Чехов', 1860, 'Россия'),
    ('Михаил Булгаков', 1891, 'Россия'),
    ('Агата Кристи', 1890, 'Великобритания'),
    ('Эрнест Хемингуэй', 1899, 'США'),
    ('Джордж Оруэлл', 1903, 'Великобритания'),
    ('Рэй Брэдбери', 1920, 'США'))
`

// pgsLab6AuthorRows counts the rows of authors equal to the file's row w.
const pgsLab6AuthorRows = `(SELECT count(*) FROM authors a WHERE a.full_name = w.full_name AND a.birth_year = w.birth_year AND a.country = w.country)`

const pgsLab6Authors = pgsLab6AuthorsWant + `SELECT (SELECT count(*) FROM authors) = 10
   AND NOT EXISTS (SELECT 1 FROM want w WHERE ` + pgsLab6AuthorRows + ` <> 1)`

const pgsLab6AuthorsNow = pgsLab6AuthorsWant + `SELECT 'в authors строк: ' || (SELECT count(*) FROM authors) ||
    ' (в authors.csv авторов: 10); авторов из файла с теми же годом и страной: ' ||
    (SELECT count(*) FROM want w WHERE ` + pgsLab6AuthorRows + ` > 0) || ' из 10' ||
    coalesce('; из них встречаются больше одного раза: ' || nullif((SELECT count(*) FROM want w WHERE ` + pgsLab6AuthorRows + ` > 1), 0), '') ||
    coalesce('; строк, которых нет в файле: ' || nullif((SELECT count(*) FROM authors a WHERE NOT EXISTS (SELECT 1 FROM want w
        WHERE a.full_name = w.full_name AND a.birth_year = w.birth_year AND a.country = w.country)), 0), '')`

// Task 2. The 14 books of the file with all their columns, «Шинель» with a
// NULL page count. isbn is unique, so 14 rows that each match a book of the
// file are exactly the file.
const pgsLab6BooksWant = `WITH want (isbn, title, price, page_count, published_year, in_stock) AS (VALUES
    ('978-5-389-07435-4', 'Евгений Онегин', 450.00, 320, 1833, true),
    ('978-5-389-04720-4', 'Мёртвые души', 450.00, 352, 1842, true),
    ('978-5-389-07123-0', 'Шинель', 260.00, NULL, 1842, true),
    ('978-5-389-04926-0', 'Преступление и наказание', 590.00, 608, 1866, true),
    ('978-5-389-05178-2', 'Идиот', 540.00, 640, 1869, false),
    ('978-5-389-21580-1', 'Анна Каренина', 790.00, 864, 1878, true),
    ('978-5-389-10364-1', 'Каштанка', 190.00, 48, 1887, false),
    ('978-5-389-06010-4', 'Вишнёвый сад', 280.00, 96, 1904, true),
    ('978-5-389-01686-6', 'Мастер и Маргарита', 690.00, 480, 1967, true),
    ('978-5-04-112380-2', 'Убийство в Восточном экспрессе', 520.00, 256, 1934, true),
    ('978-5-17-090631-1', 'Прощай, оружие!', 560.00, 320, 1929, false),
    ('978-5-17-094367-5', 'Скотный двор', 350.00, 128, 1945, false),
    ('978-5-17-080115-9', '1984', 520.00, 320, 1949, true),
    ('978-5-04-107463-0', '451 градус по Фаренгейту', 480.00, 256, 1953, false))
`

const pgsLab6Books = pgsLab6BooksWant + `SELECT (SELECT count(*) FROM books) = 14
   AND NOT EXISTS (SELECT 1 FROM want w WHERE NOT EXISTS (SELECT 1 FROM books b
        WHERE b.isbn = w.isbn AND b.title = w.title AND b.price = w.price AND b.page_count IS NOT DISTINCT FROM w.page_count
          AND b.published_year = w.published_year AND b.in_stock = w.in_stock))`

// pgsLab6BooksNow names the columns that differ from the file, matching the
// books by ISBN — a column list in the wrong order loads without an error when
// the swapped columns share a type (page_count and published_year), and this
// is where the student sees it.
const pgsLab6BooksNow = pgsLab6BooksWant + `, j AS (SELECT w.*, b.title AS b_title, b.price AS b_price, b.page_count AS b_pages,
        b.published_year AS b_year, b.in_stock AS b_stock
    FROM want w JOIN books b ON b.isbn = w.isbn)
SELECT 'в books строк: ' || (SELECT count(*) FROM books) || ' (в books.csv книг: 14); книг с ISBN из файла: ' || count(*) ||
    coalesce('; не совпадают с файлом: ' || nullif(concat_ws(', ',
        'title (строк: ' || nullif(count(*) FILTER (WHERE b_title IS DISTINCT FROM title), 0) || ')',
        'price (строк: ' || nullif(count(*) FILTER (WHERE b_price IS DISTINCT FROM price), 0) || ')',
        'page_count (строк: ' || nullif(count(*) FILTER (WHERE b_pages IS DISTINCT FROM page_count), 0) || ')',
        'published_year (строк: ' || nullif(count(*) FILTER (WHERE b_year IS DISTINCT FROM published_year), 0) || ')',
        'in_stock (строк: ' || nullif(count(*) FILTER (WHERE b_stock IS DISTINCT FROM in_stock), 0) || ')'), ''), '')
FROM j`

// Task 3. The 12 customers of the file with their data, the fixed one on
// 2023-12-31, and no rows besides them and the ★ task's test customers. email
// is unique, so each of them is there once.
const pgsLab6CustomersWant = `WITH want (email, full_name, city, registered_at) AS (VALUES
    ('anna.smirnova@example.org', 'Анна Смирнова', 'Москва', date '2023-02-14'),
    ('igor.petrov@example.net', 'Игорь Петров', 'Казань', date '2023-03-02'),
    ('olga.kuznetsova@example.org', 'Ольга Кузнецова', 'Санкт-Петербург', date '2023-04-19'),
    ('d.volkov@example.com', 'Дмитрий Волков', 'Екатеринбург', date '2023-06-07'),
    ('maria.sokolova@example.net', 'Мария Соколова', 'Новосибирск', date '2023-08-23'),
    ('sergey.morozov@example.org', 'Сергей Морозов', NULL, date '2023-09-11'),
    ('elena.pavlova@example.com', 'Елена Павлова', 'Казань', date '2023-11-05'),
    ('n.fedorov@example.org', 'Николай Федоров', 'Москва', date '2023-12-31'),
    ('tatiana.belova@example.net', 'Татьяна Белова', 'Самара', date '2024-01-16'),
    ('a.nikitin@example.com', 'Алексей Никитин', 'Нижний Новгород', date '2024-02-28'),
    ('irina.zaitseva@example.org', 'Ирина Зайцева', 'Санкт-Петербург', date '2024-03-09'),
    ('p.orlov@example.net', 'Павел Орлов', 'Краснодар', date '2024-04-21'))
`

// pgsLab6CustomerSame: customer c holds the file's row w.
const pgsLab6CustomerSame = `(c.email = w.email AND c.full_name = w.full_name AND c.city IS NOT DISTINCT FROM w.city AND c.registered_at::date = w.registered_at)`

// pgsLab6Stray: a row that is neither a customer of the file nor a test one.
const pgsLab6Stray = `(SELECT count(*) FROM customers c WHERE c.email NOT IN (SELECT email FROM want) AND c.email NOT ILIKE 'test%')`

const pgsLab6Customers = pgsLab6CustomersWant + `SELECT NOT EXISTS (SELECT 1 FROM want w WHERE NOT EXISTS (SELECT 1 FROM customers c WHERE ` + pgsLab6CustomerSame + `))
   AND ` + pgsLab6Stray + ` = 0`

const pgsLab6CustomersNow = pgsLab6CustomersWant + `SELECT 'в customers строк: ' || (SELECT count(*) FROM customers) ||
    '; покупателей из customers.csv с данными из файла: ' ||
    (SELECT count(*) FROM want w WHERE EXISTS (SELECT 1 FROM customers c WHERE ` + pgsLab6CustomerSame + `)) || ' из 12' ||
    coalesce('; имя, город или дата отличаются у: ' || (SELECT string_agg(w.email, ', ' ORDER BY w.email)
        FROM want w JOIN customers c ON c.email = w.email WHERE ` + pgsLab6CustomerSame + ` IS NOT TRUE), '') ||
    coalesce('; строк не из файла: ' || nullif(` + pgsLab6Stray + `, 0), '')`

// Task 4 needs the catalogue of task 2: exporting an empty or wrong books
// table would give a file that matches the database and says nothing.
const pgsLab6OutOfStock = `COPY (SELECT title, price FROM books WHERE NOT in_stock ORDER BY title) TO STDOUT WITH (FORMAT csv, HEADER)`

const pgsLab6OutFile = "/root/export/out_of_stock.csv"

// pgsLab6Header is the first line of the export, trailing blanks dropped as
// glpg outfile drops them. A file without the header, with semicolons or with
// psql's table frame shows it in its first line, and saying so is clearer than
// the comparison's "the content differs". No double quotes: the line ends up
// inside the echo of the failure message.
const pgsLab6Header = `$(head -n 1 ` + pgsLab6OutFile + ` 2>/dev/null | sed 's/[[:space:]]*$//')`

// pgsLab6Unordered is the condition "the export has the right lines in another
// order" — the ORDER BY left out, which the comparison would only report as
// different content.
var pgsLab6Unordered = `[ -s ` + pgsLab6OutFile + ` ] && [ "$(sed 's/[[:space:]]*$//' ` + pgsLab6OutFile + ` | LC_ALL=C sort)" = "$(glpg q pereplet ` + b64(pgsLab6OutOfStock) + ` | LC_ALL=C sort)" ]` +
	` && [ "$(sed 's/[[:space:]]*$//' ` + pgsLab6OutFile + `)" != "$(glpg q pereplet ` + b64(pgsLab6OutOfStock) + `)" ]`

// Task 5. The genres of the file under their own ids, «Приключения» once,
// nothing else, and an identity counter whose next value is past every id.
//
// seq.next is the value the counter of genres.id hands out next, read without
// nextval(): the sequence relation is selected through query_to_xml, because a
// plain query cannot take a table name from pg_get_serial_sequence. Whatever
// the student did to the counter — setval, setval(…, false), ALTER … RESTART —
// shows here. No row (and so NULL) when genres.id has no counter at all.
const pgsLab6GenresWant = `WITH want (id, name) AS (VALUES
    (1, 'Роман'), (2, 'Повесть'), (3, 'Рассказ'), (5, 'Пьеса'),
    (6, 'Поэзия'), (8, 'Детектив'), (9, 'Фантастика'), (12, 'Антиутопия')),
seq AS (
    SELECT CASE WHEN (xpath('/row/is_called/text()', x))[1]::text = 'true'
                THEN (xpath('/row/last_value/text()', x))[1]::text::bigint + s.seqincrement
                ELSE (xpath('/row/last_value/text()', x))[1]::text::bigint END AS next
    FROM pg_sequence s,
         query_to_xml('SELECT last_value, is_called FROM ' || s.seqrelid::regclass, false, true, '') AS x
    WHERE s.seqrelid = to_regclass(pg_get_serial_sequence('genres', 'id')))
`

const pgsLab6GenreRows = `(SELECT count(*) FROM want w WHERE EXISTS (SELECT 1 FROM genres g WHERE g.id = w.id AND g.name = w.name))`

const pgsLab6Genres = pgsLab6GenresWant + `SELECT ` + pgsLab6GenreRows + ` = 8
   AND (SELECT count(*) FROM genres WHERE name = 'Приключения') = 1
   AND (SELECT count(*) FROM genres) = 9
   AND coalesce((SELECT next FROM seq) > (SELECT max(id) FROM genres), false)`

const pgsLab6GenresNow = pgsLab6GenresWant + `SELECT 'в genres строк: ' || (SELECT count(*) FROM genres) ||
    '; жанров из genres.csv с их номерами: ' || ` + pgsLab6GenreRows + ` || ' из 8' ||
    '; «Приключения»: ' || coalesce((SELECT string_agg('id ' || id, ', ' ORDER BY id) FROM genres WHERE name = 'Приключения'), 'нет') ||
    coalesce('; счётчик id выдаст следующим номер ' || (SELECT next FROM seq), '; у genres.id нет счётчика (identity)') ||
    coalesce(', а наибольший id в genres — ' || (SELECT max(id) FROM genres), '')`

// Task 6 ★. test1…test1000@example.com, each with its number in the name, and
// no other test addresses.
const pgsLab6TestRows = `(SELECT count(*) FROM generate_series(1, 1000) AS g
        JOIN customers c ON c.email = 'test' || g || '@example.com' AND c.full_name = 'Тестовый покупатель ' || g)`

const pgsLab6Test = `SELECT (SELECT count(*) FROM customers WHERE email LIKE 'test%@example.com') = 1000
   AND ` + pgsLab6TestRows + ` = 1000`

const pgsLab6TestNow = `SELECT 'покупателей с email test…@example.com: ' || (SELECT count(*) FROM customers WHERE email LIKE 'test%@example.com') ||
    '; из них testN@example.com с именем «Тестовый покупатель N» для N от 1 до 1000: ' || ` + pgsLab6TestRows + ` ||
    coalesce('; например, не так: ' || (SELECT c.email || ' — «' || c.full_name || '»' FROM customers c
        WHERE c.email ILIKE 'test%' AND NOT EXISTS (SELECT 1 FROM generate_series(1, 1000) AS g
            WHERE c.email = 'test' || g || '@example.com' AND c.full_name = 'Тестовый покупатель ' || g)
        ORDER BY c.id LIMIT 1), '')`

const (
	pgsLab6BooksGood  = "в books 14 книг из books.csv — с ISBN, названием, ценой, числом страниц, годом и наличием из файла; где число страниц не указано — NULL"
	pgsLab6ExportGood = "в /root/export/out_of_stock.csv — книги не в наличии в формате CSV: заголовок title,price и строки по алфавиту названий"
)

func init() {
	pgStartLabs["ch-pgs-lab6"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab6Schema) +
			"mkdir -p /root/import /root/export\nrm -f /root/export/*.csv\n" +
			pgsLab6File("/root/import/authors.csv", pgsLab6AuthorsCSV) +
			pgsLab6File("/root/import/books.csv", pgsLab6BooksCSV) +
			pgsLab6File("/root/import/customers.csv", pgsLab6CustomersCSV) +
			pgsLab6File("/root/import/genres.csv", pgsLab6GenresCSV)),
		Checks: map[int]string{
			1: check(pgTrue("pereplet", pgsLab6Authors),
				"в authors 10 авторов из authors.csv — каждый по одному разу, с годом рождения и страной из файла",
				"Сейчас "+pgVal("pereplet", pgsLab6AuthorsNow)),
			2: check(pgTrue("pereplet", pgsLab6Books),
				pgsLab6BooksGood,
				"Сейчас "+pgVal("pereplet", pgsLab6BooksNow)),
			3: check(pgTrue("pereplet", pgsLab6Customers),
				"в customers 12 покупателей из customers.csv — с email, именем, городом и датой регистрации из файла; исправленная дата — 31 декабря 2023 года",
				"Сейчас "+pgVal("pereplet", pgsLab6CustomersNow)),
			4: "if ! " + pgTrue("pereplet", pgsLab6Books) + "; then " +
				fail(studentFail(pgsLab6ExportGood, "Сейчас "+pgVal("pereplet", pgsLab6BooksNow)+" — выгрузку проверим, когда каталог будет загружен (задание 2)")) +
				"; fi; if [ -s " + pgsLab6OutFile + ` ] && [ "` + pgsLab6Header + `" != title,price ]; then ` +
				fail(studentFail(pgsLab6ExportGood, "Сейчас первая строка файла: «"+pgsLab6Header+"», а должна быть заголовком title,price")) +
				"; fi; if " + pgsLab6Unordered + "; then " +
				fail(studentFail(pgsLab6ExportGood, "Сейчас в файле нужные строки, но не по алфавиту названий: после заголовка идёт «$(sed -n 2p "+pgsLab6OutFile+")»")) +
				"; fi; " + pgOutfile(pgsLab6OutFile, "pereplet", pgsLab6OutOfStock, pgsLab6ExportGood),
			5: check(pgTrue("pereplet", pgsLab6Genres),
				"в genres 9 строк: 8 жанров из genres.csv с их номерами и «Приключения»; следующий номер, который выдаст счётчик id, больше всех номеров в таблице",
				"Сейчас "+pgVal("pereplet", pgsLab6GenresNow)),
			6: check(pgTrue("pereplet", pgsLab6Test),
				"в customers 1000 тестовых покупателей: testN@example.com и «Тестовый покупатель N» для каждого N от 1 до 1000",
				"Сейчас "+pgVal("pereplet", pgsLab6TestNow)),
		},
	}
}
