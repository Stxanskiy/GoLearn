package main

// Lab 5 of pg-start: fixing the catalog (ch-pgs-lab5).
//
// Lesson 10 teaches UPDATE, DELETE, WHERE, IS NULL and BEGIN … COMMIT or
// ROLLBACK; here the student puts them to work on the catalog of state S3,
// filled with the ten rows of pgsLab5Books: one exact new price, one price
// raised by 10 % with a formula, the thin books marked out of stock, the
// missing page count filled in, and the book that went in twice deleted. The
// quizzes have the student see UPDATE 0, = NULL finding nothing, the *# prompt
// after BEGIN and ROLLBACK undoing an UPDATE without WHERE.
//
// The checks compare the table with the setup's rows by id, so each one sees
// its own change and also the accident the lesson warns about — a WHERE that
// hit more rows than it should:
//   - a price task wants that book's exact price and every other price as it
//     was. The book of the other price task is left out, so the two can be
//     done in either order, and a deleted row is no price change;
//   - the stock task wants exactly the books under 100 pages marked false and
//     every other book still true — the one with no page count included, as
//     NULL < 100 is not true. By ids, so filling that page count first (144,
//     not thin) changes nothing;
//   - the page-count task wants 144 in the row that was NULL and no other page
//     count changed; = NULL updates nothing and stays red;
//   - the duplicate task wants the higher id gone, the lower one kept and every
//     other book still there: a DELETE by title removes both and fails.
//
// A change made in a transaction still open is invisible to the check; each
// diagnostic says so when it sees such a session (pgsLab5Tx), since the lab
// has the student work inside BEGIN.
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.

import (
	"strconv"
	"strings"
)

func init() {
	pgStartLabs["ch-pgs-lab5"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab5Schema)),
		Checks: map[int]string{
			1: check(pgTrue("pereplet", pgsLab5Price(pgsLab5Onegin, "495.00")),
				"«Евгений Онегин» стоит 495.00, цены других книг не изменились",
				"Сейчас: "+pgsLab5Now(pgsLab5PriceState(pgsLab5Onegin))),
			2: check(pgTrue("pereplet", pgsLab5Price(pgsLab5Idiot, "671.00")),
				"цена «Идиота» выросла ровно на 10 %, цены других книг не изменились",
				"Сейчас: "+pgsLab5Now(pgsLab5PriceState(pgsLab5Idiot))),
			3: check(pgTrue("pereplet", pgsLab5Stock),
				"книги тоньше 100 страниц отмечены in_stock = false, остальные книги не тронуты",
				"Сейчас: "+pgsLab5Now(pgsLab5StockState)),
			4: check(pgTrue("pereplet", pgsLab5Pages),
				"у книги, где не было числа страниц, записано 144, у других книг число страниц прежнее",
				"Сейчас: "+pgsLab5Now(pgsLab5PagesState)),
			5: check(pgTrue("pereplet", pgsLab5Dup),
				"дубль удалён: осталась запись с меньшим id, остальные книги на месте",
				"Сейчас: "+pgsLab5Now(pgsLab5DupState)),
		},
	}
}

// pgsLab5Book is a row of the catalog the lab starts from. pages 0 is the
// unknown page count, NULL in the table.
type pgsLab5Book struct {
	title, price string
	pages, year  int
}

// pgsLab5Books is the catalog; a row's id is its place in the list, from 1.
// Three books are under 100 pages, one has no page count and the last row
// repeats the second. Lesson 10 shows the same kinds of trouble on other books,
// so its commands cannot be copied as they are.
var pgsLab5Books = []pgsLab5Book{
	{"Евгений Онегин", "450.00", 320, 1833}, // check 1: a new price
	{"Мёртвые души", "520.00", 416, 1842},   // check 5: kept
	{"Анна Каренина", "780.00", 864, 1878},
	{"Дама с собачкой", "210.00", 48, 1899}, // thin
	{"Капитанская дочка", "340.00", 176, 1836},
	{"Шинель", "190.00", 64, 1842},            // thin; the UPDATE 0 quiz
	{"Идиот", "610.00", 640, 1869},            // check 2: +10 %
	{"Ревизор", "280.00", 0, 1836},            // check 4: no page count
	{"Человек в футляре", "160.00", 32, 1898}, // thin
	{"Мёртвые души", "520.00", 416, 1842},     // check 5: the duplicate
}

// Ids of the rows the tasks are about.
const (
	pgsLab5Onegin  = 1
	pgsLab5Idiot   = 7
	pgsLab5Revizor = 8
	pgsLab5Keep    = 2
	pgsLab5Gone    = 10
)

var (
	pgsLab5ThinIDs    = pgsLab5IDs(func(id int, b pgsLab5Book) bool { return pgsLab5IsThin(b) })
	pgsLab5NotThinIDs = pgsLab5IDs(func(id int, b pgsLab5Book) bool { return !pgsLab5IsThin(b) })
	pgsLab5RestIDs    = pgsLab5IDs(func(id int, b pgsLab5Book) bool { return id != pgsLab5Keep && id != pgsLab5Gone })
	pgsLab5PriceIDs   = []int{pgsLab5Onegin, pgsLab5Idiot}
)

func pgsLab5IsThin(b pgsLab5Book) bool { return b.pages != 0 && b.pages < 100 }

// pgsLab5IDs is the ids of the rows keep says yes to.
func pgsLab5IDs(keep func(id int, b pgsLab5Book) bool) []int {
	var ids []int
	for i, b := range pgsLab5Books {
		if keep(i+1, b) {
			ids = append(ids, i+1)
		}
	}
	return ids
}

// pgsLab5In is ids for SQL's IN (…): "4, 6, 9".
func pgsLab5In(ids ...int) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.Itoa(id)
	}
	return strings.Join(s, ", ")
}

// pgsLab5PagesSQL is a book's page count as SQL: a number or NULL.
func pgsLab5PagesSQL(b pgsLab5Book) string {
	if b.pages == 0 {
		return "NULL"
	}
	return strconv.Itoa(b.pages)
}

// pgsLab5Schema is state S3 the way labs 3 and 4 leave it: authors and books,
// published_year last as lab 3 adds it, and the rows in id order.
var pgsLab5Schema = func() string {
	rows := make([]string, len(pgsLab5Books))
	for i, b := range pgsLab5Books {
		rows[i] = "    ('" + b.title + "', " + b.price + ", " + pgsLab5PagesSQL(b) + ", true, " + strconv.Itoa(b.year) + ")"
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
    ('Александр Пушкин', 1799, 'Россия'),
    ('Николай Гоголь', 1809, 'Россия'),
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Лев Толстой', 1828, 'Россия'),
    ('Антон Чехов', 1860, 'Россия');
INSERT INTO books (title, price, page_count, in_stock, published_year) VALUES
` + strings.Join(rows, ",\n") + ";\n"
}()

// pgsLab5Orig is the setup's rows as a table, o(id, title, price, page_count):
// what the checks compare books with.
var pgsLab5Orig = func() string {
	rows := make([]string, len(pgsLab5Books))
	for i, b := range pgsLab5Books {
		rows[i] = "(" + strconv.Itoa(i+1) + ", '" + b.title + "', " + b.price + ", " + pgsLab5PagesSQL(b) + ")"
	}
	return "(VALUES " + strings.Join(rows, ", ") + ") AS o(id, title, price, page_count)"
}()

// pgsLab5Changed is SQL for the rows of books, other than except, whose
// column col differs from the setup's.
func pgsLab5Changed(col string, except ...int) string {
	return "SELECT b.id, b.title FROM books b JOIN " + pgsLab5Orig + " USING (id) " +
		"WHERE b.id NOT IN (" + pgsLab5In(except...) + ") AND b." + col + " IS DISTINCT FROM o." + col
}

// pgsLab5Moved is SQL text for a diagnostic, "<what>: <titles> — …", naming
// the rows pgsLab5Changed finds; NULL when there are none.
func pgsLab5Moved(what, col string, except ...int) string {
	return "(SELECT '" + what + ": ' || string_agg(c.title, ', ' ORDER BY c.id) || " +
		"' — нажми «Пересоздать» и сделай задание заново' FROM (" + pgsLab5Changed(col, except...) + ") c)"
}

// pgsLab5Tx is SQL text: a note when the student's psql sits in an open
// transaction, whose changes no other session sees until COMMIT, or an empty
// string when there is none.
const pgsLab5Tx = "(SELECT CASE " +
	"WHEN bool_or(state = 'idle in transaction (aborted)') " +
	"THEN '. В твоей транзакции была ошибка (приглашение pereplet=!#): набери ROLLBACK; и повтори команды' " +
	"WHEN bool_or(state = 'idle in transaction') " +
	"THEN '. У тебя открыта транзакция (приглашение pereplet=*#): проверка не видит её изменений — сохрани их командой COMMIT; или отмени командой ROLLBACK;' " +
	"ELSE '' END FROM pg_stat_activity WHERE datname = 'pereplet' AND pid <> pg_backend_pid())"

// pgsLab5Now is a substitution printing a diagnostic query's text, or what to
// do when books cannot be read at all.
func pgsLab5Now(sql string) string {
	return `$(glpg q pereplet ` + b64(sql) + ` || echo 'не получилось прочитать таблицу books в базе pereplet — если её нет, нажми «Пересоздать»')`
}

// pgsLab5Price is SQL true when book id costs price and no other book's price
// changed, the books of both price tasks left out.
func pgsLab5Price(id int, price string) string {
	return "SELECT coalesce((SELECT price = " + price + " FROM books WHERE id = " + strconv.Itoa(id) + "), false) " +
		"AND NOT EXISTS (" + pgsLab5Changed("price", pgsLab5PriceIDs...) + ")"
}

// pgsLab5PriceState says what book id costs now and which other prices moved.
func pgsLab5PriceState(id int) string {
	n, t := strconv.Itoa(id), pgsLab5Books[id-1].title
	return "SELECT concat_ws('; ', coalesce(" +
		"(SELECT '«" + t + "» стоит ' || coalesce(price::text, 'NULL — цена пустая') FROM books WHERE id = " + n + "), " +
		"'книги «" + t + "» с id " + n + " в каталоге больше нет — цену меняют командой UPDATE, удалять книгу не нужно; нажми «Пересоздать»'), " +
		pgsLab5Moved("цена изменилась и у других книг", "price", pgsLab5PriceIDs...) + ") || " + pgsLab5Tx
}

// pgsLab5Stock is SQL true when the thin books, all still there, are marked
// false and every other book of the setup is still true.
var pgsLab5Stock = "SELECT (SELECT count(*) FROM books WHERE id IN (" + pgsLab5In(pgsLab5ThinIDs...) + ") AND in_stock IS FALSE) = " +
	strconv.Itoa(len(pgsLab5ThinIDs)) + " " +
	"AND NOT EXISTS (SELECT 1 FROM books WHERE id IN (" + pgsLab5In(pgsLab5NotThinIDs...) + ") AND in_stock IS NOT TRUE)"

// pgsLab5StockState counts the thin books not marked yet — without naming
// them: finding them is the task — and names the books marked by mistake or
// deleted.
var pgsLab5StockState = "SELECT concat_ws('; ', " +
	"(SELECT 'тонких книг (меньше 100 страниц) ещё не отмечено: ' || count(*) FROM books " +
	"WHERE id IN (" + pgsLab5In(pgsLab5ThinIDs...) + ") AND in_stock IS NOT FALSE HAVING count(*) > 0), " +
	"(SELECT 'лишние книги отмечены «нет на складе»: ' || string_agg(title, ', ' ORDER BY id) || " +
	"' — у них не меньше 100 страниц или число страниц не указано, верни им in_stock = true' FROM books " +
	"WHERE id IN (" + pgsLab5In(pgsLab5NotThinIDs...) + ") AND in_stock IS NOT TRUE), " +
	"(SELECT 'из каталога пропали книги: ' || string_agg(o.title, ', ' ORDER BY o.id) || " +
	"' — удалять их не нужно, только отметить; нажми «Пересоздать»' FROM " + pgsLab5Orig + " " +
	"WHERE o.id IN (" + pgsLab5In(pgsLab5ThinIDs...) + ") AND NOT EXISTS (SELECT 1 FROM books b WHERE b.id = o.id))" +
	") || " + pgsLab5Tx

// pgsLab5Pages is SQL true when the row that had no page count has 144 and no
// other page count changed.
var pgsLab5Pages = "SELECT coalesce((SELECT page_count = 144 FROM books WHERE id = " + strconv.Itoa(pgsLab5Revizor) + "), false) " +
	"AND NOT EXISTS (" + pgsLab5Changed("page_count", pgsLab5Revizor) + ")"

// pgsLab5PagesState says how many books still have no page count, or what the
// row got instead of 144, and which other page counts moved.
var pgsLab5PagesState = func() string {
	n, t := strconv.Itoa(pgsLab5Revizor), pgsLab5Books[pgsLab5Revizor-1].title
	row := "(SELECT page_count FROM books WHERE id = " + n + ")"
	return "SELECT concat_ws('; ', CASE " +
		"WHEN NOT EXISTS (SELECT 1 FROM books WHERE id = " + n + ") " +
		"THEN 'книги без числа страниц («" + t + "», id " + n + ") в каталоге больше нет — нажми «Пересоздать»' " +
		"WHEN " + row + " IS NULL THEN 'книг без числа страниц (page_count IS NULL): ' || (SELECT count(*) FROM books WHERE page_count IS NULL) " +
		"WHEN " + row + " <> 144 THEN 'у книги «" + t + "» page_count = ' || " + row + " || ', а в задании другое число' END, " +
		pgsLab5Moved("число страниц изменилось и у других книг", "page_count", pgsLab5Revizor) + ") || " + pgsLab5Tx
}()

// pgsLab5Dup is SQL true when the duplicate row is gone, the row it repeated
// is there, the title appears once, and no other book went missing.
var pgsLab5Dup = func() string {
	t := pgsLab5Books[pgsLab5Keep-1].title
	return "SELECT NOT EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Gone) + ") " +
		"AND EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Keep) + " AND title = '" + t + "') " +
		"AND (SELECT count(*) FROM books WHERE title = '" + t + "') = 1 " +
		"AND (SELECT count(*) FROM books WHERE id IN (" + pgsLab5In(pgsLab5RestIDs...) + ")) = " + strconv.Itoa(len(pgsLab5RestIDs))
}()

// pgsLab5DupState says whether a title still repeats — without naming the
// book while both rows are there: finding it is part of the task — which row
// is left, and which other books are gone. What a DELETE took cannot be typed
// back with its id, so those cases end with the reset button, once.
var pgsLab5DupState = func() string {
	t := pgsLab5Books[pgsLab5Keep-1].title
	cnt := "(SELECT count(*) FROM books WHERE title = '" + t + "')"
	kept := "EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Keep) + " AND title = '" + t + "')"
	lost := "(SELECT string_agg(o.title, ', ' ORDER BY o.id) FROM " + pgsLab5Orig + " " +
		"WHERE o.id IN (" + pgsLab5In(pgsLab5RestIDs...) + ") AND NOT EXISTS (SELECT 1 FROM books b WHERE b.id = o.id))"
	return "SELECT CASE " +
		"WHEN " + cnt + " > 1 THEN 'одна книга всё ещё записана дважды: строк в books — ' || (SELECT count(*) FROM books) || " +
		"', а разных названий — ' || (SELECT count(DISTINCT title) FROM books) " +
		"WHEN " + cnt + " = 0 THEN 'книги «" + t + "» в каталоге не осталось: удалены обе строки, а одну нужно было оставить' " +
		"WHEN NOT " + kept + " " +
		"THEN 'осталась одна строка «" + t + "», но с id ' || (SELECT min(id) FROM books WHERE title = '" + t + "') || " +
		"', а оставить нужно было строку с меньшим id' " +
		"WHEN EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Gone) + ") " +
		"THEN 'строка с id " + strconv.Itoa(pgsLab5Gone) + " всё ещё в каталоге — лишнюю строку нужно удалить, а не переименовать' " +
		"ELSE 'дубль удалён' END || " +
		"coalesce('; из каталога пропали и другие книги: ' || " + lost + ", '') || " +
		"CASE WHEN " + cnt + " = 0 OR NOT " + kept + " OR " + lost + " IS NOT NULL THEN ' — нажми «Пересоздать»' ELSE '' END || " +
		pgsLab5Tx
}()
