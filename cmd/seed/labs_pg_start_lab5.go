package main

// Lab 5 of pg-start: fixing the catalog (ch-pgs-lab5).
//
// The catalog runs through labs 4-7 as one story, so this lab starts where lab
// 4 ends: its setup is lab 4's setup with every reference solution of lab 4
// applied — eight authors and fourteen books, the same rows inserted in the
// same order without ids, so the ids and the sequences are the ones a student
// who did lab 4 as asked has. On top of that comes one change, named in the
// lab's introduction: the intern ran the INSERT of «Затерянный мир» from lab 4
// once more, so the book is in twice (ids 14 and 15).
//
// Lesson 10 teaches UPDATE, DELETE, WHERE, IS NULL and BEGIN … COMMIT or
// ROLLBACK; the tasks put them to work on this catalog: an exact new price
// («Евгений Онегин»), a price raised by 10 % with a formula («Идиот»), the
// thin books — under 200 pages — marked out of stock with one UPDATE, the
// fields lab 4 left empty in «Анна Каренина» filled in (the manager promised
// them "later"), and, as the ★ task, the duplicate deleted inside BEGIN …
// COMMIT. The quizzes have the student see UPDATE 0, = NULL finding nothing,
// the *# prompt after BEGIN and ROLLBACK undoing an UPDATE without WHERE.
//
// The checks compare the table with the setup's rows by id, so each one sees
// its own change and also the accident the lesson warns about — a WHERE that
// hit more rows than it should:
//   - a price task wants that book's exact price and every other price as it
//     was. The book of the other price task is left out, so the two can be
//     done in either order, and a deleted row is no price change;
//   - the stock task wants exactly the books under 200 pages marked false and
//     every other in_stock as it was. «Анна Каренина» has no page count, and
//     NULL < 200 is not true, so she is not thin: her in_stock may be NULL (as
//     set up) or true (the next task), never false — which is what marking her
//     by hand, or a WHERE that takes the NULL as well, leaves. The task asks
//     for one UPDATE: the two thin books that were in stock must share an
//     xmin, as in lab 4's "one INSERT". xmin is the writing transaction, so
//     two UPDATEs inside one BEGIN … COMMIT pass as well; the rows are the
//     same either way. Only those two: an UPDATE that skips the thin book
//     already out of stock is just as right.
//     Running the one UPDATE after separate ones rewrites both and passes, so
//     the diagnostic sends the student there rather than to «Пересоздать»;
//   - the «Анна Каренина» task wants 864, true and 1878 in her row, the title
//     once — an INSERT of her data next to the old row is a beginner's way of
//     "filling in", and the diagnostic names it — and no other page count,
//     year or in_stock (bar the thin books') changed;
//   - the duplicate task wants id 15 gone, id 14 kept, the title once, and
//     every other book still there: a DELETE by title removes both and fails.
//
// The tasks pass in any order: each changes its own column or row, and each
// check reads only those, leaving out what another task changes — the price
// tasks each other's book, the «Анна Каренина» task the thin books' in_stock,
// the stock task her in_stock as long as it is not false; the duplicate is no
// thin book. The UPDATE 0 quiz names «Таинственный остров», which is in stock
// already: run with the right letter case, the intern's command still changes
// nothing a check reads.
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
			1: check(pgTrue("pereplet", pgsLab5Price(pgsLab5Onegin, "460.00")),
				"«Евгений Онегин» стоит 460.00, цены других книг не изменились",
				"Сейчас: "+pgsLab5Now(pgsLab5PriceState(pgsLab5Onegin, ""))),
			2: check(pgTrue("pereplet", pgsLab5Price(pgsLab5Idiot, "715.00")),
				"цена «Идиота» выросла ровно на 10 %, цены других книг не изменились",
				"Сейчас: "+pgsLab5Now(pgsLab5PriceState(pgsLab5Idiot, "786.50"))),
			3: check(pgTrue("pereplet", pgsLab5Stock),
				"книги тоньше 200 страниц отмечены in_stock = false одной командой UPDATE, отметки in_stock у других книг не тронуты",
				"Сейчас: "+pgsLab5Now(pgsLab5StockState)),
			4: check(pgTrue("pereplet", pgsLab5Filled),
				"у «Анны Карениной» 864 страницы, она есть на складе, год издания 1878; у других книг число страниц, год и наличие на складе не тронуты",
				"Сейчас: "+pgsLab5Now(pgsLab5FilledState)),
			5: check(pgTrue("pereplet", pgsLab5Dup),
				"дубль «Затерянного мира» удалён: осталась запись с меньшим id, остальные книги на месте",
				"Сейчас: "+pgsLab5Now(pgsLab5DupState)),
		},
	}
}

// pgsLab5Book is a row of books as the lab starts: every value an SQL literal,
// NULL for an empty cell.
type pgsLab5Book struct{ title, price, pages, stock, year string }

// pgsLab5Books is the catalog; a row's id is its place in the list, from 1.
// Rows 1-10 are lab 4's setup, 11-14 what its tasks add (11-12 two books, 13
// «Анна Каренина» with only a price, 14 the intern's «Затерянный мир»), 15 the
// intern's second run of the same INSERT.
var pgsLab5Books = []pgsLab5Book{
	{"Басни", "240.00", "224", "true", "1809"},
	{"Евгений Онегин", "420.00", "320", "true", "1833"},    // check 1: a new price
	{"Капитанская дочка", "210.00", "192", "true", "1836"}, // thin
	{"Повести Белкина", "190.00", "160", "false", "1831"},  // thin, out of stock already
	{"Идиот", "650.00", "640", "true", "1869"},             // check 2: +10 %
	{"Братья Карамазовы", "990.00", "992", "false", "1880"},
	{"Белые ночи", "150.00", "96", "true", "1848"},            // thin
	{"Таинственный остров", "1350.00", "704", "true", "1875"}, // the UPDATE 0 quiz
	{"Вокруг света за восемьдесят дней", "360.00", "320", "false", "1872"},
	{"Дети капитана Гранта", "560.00", "672", "true", "1868"},
	{"Собака Баскервилей", "450.00", "256", "true", "1902"},
	{"Приключения Тома Сойера", "380.00", "288", "false", "1876"},
	{"Анна Каренина", "870.00", "NULL", "NULL", "NULL"}, // check 4: fill in
	{"Затерянный мир", "430.00", "320", "true", "1912"}, // check 5: kept
	{"Затерянный мир", "430.00", "320", "true", "1912"}, // check 5: the duplicate
}

// Ids of the rows the tasks are about.
const (
	pgsLab5Onegin   = 2
	pgsLab5Idiot    = 5
	pgsLab5Karenina = 13
	pgsLab5Keep     = 14
	pgsLab5Gone     = 15
)

// pgsLab5ThinPages: a book with fewer pages than this is thin (the stock task).
const pgsLab5ThinPages = 200

var (
	pgsLab5ThinIDs = pgsLab5IDs(func(id int, b pgsLab5Book) bool { return pgsLab5IsThin(b) })
	// pgsLab5MarkIDs are the thin books in stock: the ones the stock task's
	// UPDATE really changes, and so the ones that share its xmin.
	pgsLab5MarkIDs  = pgsLab5IDs(func(id int, b pgsLab5Book) bool { return pgsLab5IsThin(b) && b.stock == "true" })
	pgsLab5StockOut = append(append([]int{}, pgsLab5ThinIDs...), pgsLab5Karenina)
	pgsLab5RestIDs  = pgsLab5IDs(func(id int, b pgsLab5Book) bool { return id != pgsLab5Keep && id != pgsLab5Gone })
	pgsLab5PriceIDs = []int{pgsLab5Onegin, pgsLab5Idiot}
)

func pgsLab5IsThin(b pgsLab5Book) bool {
	n, err := strconv.Atoi(b.pages)
	return err == nil && n < pgsLab5ThinPages
}

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

// pgsLab5In is ids for SQL's IN (…): "3, 4, 7".
func pgsLab5In(ids ...int) string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = strconv.Itoa(id)
	}
	return strings.Join(s, ", ")
}

func pgsLab5Lit(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// pgsLab5Title is the title of book id as an SQL literal.
func pgsLab5Title(id int) string { return pgsLab5Lit(pgsLab5Books[id-1].title) }

// pgsLab5Schema is the catalog as lab 4 leaves it, plus the duplicate: tables
// as in state S3 (published_year last, as lab 3 adds it), rows in id order.
var pgsLab5Schema = func() string {
	rows := make([]string, len(pgsLab5Books))
	for i, b := range pgsLab5Books {
		rows[i] = "    (" + pgsLab5Lit(b.title) + ", " + b.price + ", " + b.pages + ", " + b.stock + ", " + b.year + ")"
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
    ('Иван Крылов',       1769, 'Россия'),
    ('Александр Пушкин',  1799, 'Россия'),
    ('Фёдор Достоевский', 1821, 'Россия'),
    ('Жюль Верн',         1828, 'Франция'),
    ('Антон Чехов',       1860, 'Россия'),
    ('Лев Толстой',       1828, 'Россия'),
    ('Марк Твен',         1835, 'США'),
    ('Артур Конан Дойл',  1859, 'Великобритания');
INSERT INTO books (title, price, page_count, in_stock, published_year) VALUES
` + strings.Join(rows, ",\n") + ";\n"
}()

// pgsLab5Orig is the setup's rows as a table, o(id, title, price, page_count,
// in_stock, published_year): what the checks compare books with.
var pgsLab5Orig = func() string {
	rows := make([]string, len(pgsLab5Books))
	for i, b := range pgsLab5Books {
		rows[i] = "(" + strconv.Itoa(i+1) + ", " + pgsLab5Lit(b.title) + ", " + b.price + ", " + b.pages + ", " + b.stock + ", " + b.year + ")"
	}
	return "(VALUES " + strings.Join(rows, ", ") + ") AS o(id, title, price, page_count, in_stock, published_year)"
}()

// pgsLab5Changed is SQL for the rows of books, other than except, whose
// column col differs from the setup's.
func pgsLab5Changed(col string, except ...int) string {
	return "SELECT b.id, b.title FROM books b JOIN " + pgsLab5Orig + " USING (id) " +
		"WHERE b.id NOT IN (" + pgsLab5In(except...) + ") AND b." + col + " IS DISTINCT FROM o." + col
}

// pgsLab5Moved is SQL text for a diagnostic, "<what>: <titles>", naming the
// rows pgsLab5Changed finds; NULL when there are none.
func pgsLab5Moved(what, col string, except ...int) string {
	return "(SELECT '" + what + ": ' || string_agg('«' || c.title || '»', ', ' ORDER BY c.id) FROM (" +
		pgsLab5Changed(col, except...) + ") c)"
}

// pgsLab5Lost is SQL text naming the books of ids that are no longer in the
// table, NULL when all are there.
func pgsLab5Lost(ids []int) string {
	return "(SELECT string_agg('«' || o.title || '»', ', ' ORDER BY o.id) FROM " + pgsLab5Orig + " " +
		"WHERE o.id IN (" + pgsLab5In(ids...) + ") AND NOT EXISTS (SELECT 1 FROM books b WHERE b.id = o.id))"
}

// pgsLab5Reset is what to do after a change that cannot be typed back.
const pgsLab5Reset = " — нажми «Пересоздать» и сделай задание заново"

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
// do when books cannot be read at all: it is gone, or a transaction the
// student left open holds a lock on it (BEGIN, then ALTER or DROP TABLE), and
// the check gives up after glpg's lock_timeout.
func pgsLab5Now(sql string) string {
	return `$(glpg q pereplet ` + b64(sql) + ` || echo 'не получилось прочитать таблицу books в базе pereplet — если у тебя открыта транзакция, закончи её (COMMIT; или ROLLBACK;), а если таблицы нет — нажми «Пересоздать»')`
}

// pgsLab5Price is SQL true when book id costs price and no other book's price
// changed, the books of both price tasks left out.
func pgsLab5Price(id int, price string) string {
	return "SELECT coalesce((SELECT price = " + price + " FROM books WHERE id = " + strconv.Itoa(id) + "), false) " +
		"AND NOT EXISTS (" + pgsLab5Changed("price", pgsLab5PriceIDs...) + ")"
}

// pgsLab5PriceState says what book id costs now and which other prices moved.
// twice, when set, is the price the formula gives run two times: the student
// is told so, since the task warns against it.
func pgsLab5PriceState(id int, twice string) string {
	n, t := strconv.Itoa(id), pgsLab5Books[id-1].title
	note := "''"
	if twice != "" {
		note = "CASE WHEN price = " + twice + " THEN ' — это подорожание на 10 % два раза: команда выполнилась дважды' ELSE '' END"
	}
	return "SELECT concat_ws('; ', coalesce(" +
		"(SELECT '«" + t + "» стоит ' || coalesce(price::text, 'NULL — цена пустая') || " + note + " FROM books WHERE id = " + n + "), " +
		"'книги «" + t + "» с id " + n + " в каталоге больше нет — цену меняют командой UPDATE, удалять книгу не нужно; нажми «Пересоздать»'), " +
		"(SELECT m || '" + pgsLab5Reset + "' FROM (SELECT " + pgsLab5Moved("цена изменилась и у других книг", "price", pgsLab5PriceIDs...) + " AS m) x WHERE m IS NOT NULL)" +
		") || " + pgsLab5Tx
}

// pgsLab5OneUpdate is SQL true when the thin books the stock task changes
// share one xmin: one transaction — for a student without BEGIN, one
// statement — wrote them.
var pgsLab5OneUpdate = "(SELECT count(DISTINCT xmin::text) FROM books WHERE id IN (" + pgsLab5In(pgsLab5MarkIDs...) + ")) = 1"

// pgsLab5ThinMarked is SQL true when every thin book is there and false.
var pgsLab5ThinMarked = "(SELECT count(*) FROM books WHERE id IN (" + pgsLab5In(pgsLab5ThinIDs...) + ") AND in_stock IS FALSE) = " +
	strconv.Itoa(len(pgsLab5ThinIDs))

// pgsLab5Stock is SQL true when the thin books are marked false by one
// statement, «Анна Каренина» is not false, and every other in_stock is as set
// up.
var pgsLab5Stock = "SELECT " + pgsLab5ThinMarked + " AND " + pgsLab5OneUpdate + " " +
	"AND NOT EXISTS (" + pgsLab5Changed("in_stock", pgsLab5StockOut...) + ") " +
	"AND NOT EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Karenina) + " AND in_stock IS FALSE)"

// pgsLab5StockState counts the thin books not marked yet — without naming
// them: finding them is the task — says when they were marked one by one, and
// names the books marked by mistake or deleted. «Анна Каренина» marked false
// gets her own note, the NULL point of the task; how to put her back is said
// only when she is the one mistake, since otherwise the note before it already
// sends the student to «Пересоздать».
var pgsLab5StockState = func() string {
	k, thin := strconv.Itoa(pgsLab5Karenina), strconv.Itoa(pgsLab5ThinPages)
	moved := pgsLab5Changed("in_stock", pgsLab5StockOut...)
	return "SELECT concat_ws('; ', " +
		"(SELECT 'тонких книг (меньше " + thin + " страниц) ещё не отмечено: ' || count(*) FROM books " +
		"WHERE id IN (" + pgsLab5In(pgsLab5ThinIDs...) + ") AND in_stock IS NOT FALSE HAVING count(*) > 0), " +
		"CASE WHEN " + pgsLab5ThinMarked + " AND NOT " + pgsLab5OneUpdate + " " +
		"THEN 'тонкие книги отмечены, но разными командами UPDATE, а нужно одной: выполни UPDATE с одним условием — " +
		"«меньше " + thin + " страниц», он перезапишет все тонкие книги разом' END, " +
		"(SELECT m || '" + pgsLab5Reset + "' FROM (SELECT " +
		pgsLab5Moved("отметка in_stock изменилась у книг, которые не тонкие", "in_stock", pgsLab5StockOut...) + " AS m) x WHERE m IS NOT NULL), " +
		"(SELECT '«Анна Каренина» отмечена «нет на складе», а под условие «меньше " + thin + " страниц» она не подходит' || " +
		"CASE WHEN page_count IS NULL THEN ': число страниц у неё не указано (NULL), а про неизвестное число нельзя сказать, что оно меньше " + thin + "' ELSE '' END || " +
		"CASE WHEN NOT EXISTS (" + moved + ") THEN '. Верни ей отметку in_stock = true: менеджер пишет, что книга есть на складе' ELSE '' END " +
		"FROM books WHERE id = " + k + " AND in_stock IS FALSE), " +
		"(SELECT 'из каталога пропали книги: ' || l || ' — удалять их не нужно, только отметить; нажми «Пересоздать»' FROM (SELECT " +
		pgsLab5Lost(pgsLab5ThinIDs) + " AS l) x WHERE l IS NOT NULL)" +
		") || " + pgsLab5Tx
}()

// pgsLab5FilledWant is what the «Анна Каренина» task writes into her row,
// column → SQL literal, in the order the diagnostic lists them.
var pgsLab5FilledWant = [][2]string{{"page_count", "864"}, {"in_stock", "true"}, {"published_year", "1878"}}

// pgsLab5KareninaCount is SQL for how many rows carry her title.
var pgsLab5KareninaCount = "(SELECT count(*) FROM books WHERE title = " + pgsLab5Title(pgsLab5Karenina) + ")"

// pgsLab5Filled is SQL true when her row has the three values, her title is
// in once, and no other book's page count, year or in_stock (the thin books'
// aside) changed.
var pgsLab5Filled = func() string {
	var conds []string
	for _, c := range pgsLab5FilledWant {
		conds = append(conds, c[0]+" IS NOT DISTINCT FROM "+c[1])
	}
	k := strconv.Itoa(pgsLab5Karenina)
	return "SELECT EXISTS (SELECT 1 FROM books WHERE id = " + k + " AND title = " + pgsLab5Title(pgsLab5Karenina) + " AND " +
		strings.Join(conds, " AND ") + ") " +
		"AND " + pgsLab5KareninaCount + " = 1 " +
		"AND NOT EXISTS (" + pgsLab5Changed("page_count", pgsLab5Karenina) + ") " +
		"AND NOT EXISTS (" + pgsLab5Changed("published_year", pgsLab5Karenina) + ") " +
		"AND NOT EXISTS (" + pgsLab5Changed("in_stock", pgsLab5StockOut...) + ")"
}()

// pgsLab5FilledState names each of her three cells that is not what the
// task gives (as it is and as it should be: the values are in the task), a
// second row of her title (an INSERT where an UPDATE was wanted), and the
// other books whose page count, year or stock mark moved.
var pgsLab5FilledState = func() string {
	k := strconv.Itoa(pgsLab5Karenina)
	var diffs []string
	for _, c := range pgsLab5FilledWant {
		diffs = append(diffs, "CASE WHEN "+c[0]+" IS DISTINCT FROM "+c[1]+" THEN '"+c[0]+" = ' || coalesce("+c[0]+"::text, 'NULL') || ' вместо "+c[1]+"' END")
	}
	moved := "concat_ws('; ', " +
		pgsLab5Moved("число страниц изменилось и у других книг", "page_count", pgsLab5Karenina) + ", " +
		pgsLab5Moved("год издания изменился и у других книг", "published_year", pgsLab5Karenina) + ", " +
		pgsLab5Moved("отметка in_stock изменилась и у других книг", "in_stock", pgsLab5StockOut...) + ")"
	return "SELECT concat_ws('; ', coalesce(" +
		"(SELECT 'у «Анны Карениной» ' || coalesce(nullif(concat_ws(', ', " + strings.Join(diffs, ", ") + "), ''), 'всё как в задании') " +
		"FROM books WHERE id = " + k + " AND title = " + pgsLab5Title(pgsLab5Karenina) + "), " +
		"'строки «Анны Карениной» с id " + k + " в каталоге больше нет — менять нужно её строку, удалять и добавлять заново не нужно; нажми «Пересоздать»'), " +
		"(SELECT '«Анна Каренина» записана в каталоге несколько раз (строк: ' || n || '): данные нужно записать в её старую строку командой UPDATE, " +
		"а не добавлять новую командой INSERT. Лишнюю строку — ту, у которой id больше, — удали командой DELETE по её id' " +
		"FROM (SELECT " + pgsLab5KareninaCount + " AS n) x WHERE n > 1), " +
		"nullif(" + moved + ", '') || '" + pgsLab5Reset + "'" +
		") || " + pgsLab5Tx
}()

// pgsLab5Dup is SQL true when the duplicate row is gone, the row it repeated
// is there, the title appears once, and no other book went missing.
var pgsLab5Dup = func() string {
	t := pgsLab5Title(pgsLab5Keep)
	return "SELECT NOT EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Gone) + ") " +
		"AND EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Keep) + " AND title = " + t + ") " +
		"AND (SELECT count(*) FROM books WHERE title = " + t + ") = 1 " +
		"AND (SELECT count(*) FROM books WHERE id IN (" + pgsLab5In(pgsLab5RestIDs...) + ")) = " + strconv.Itoa(len(pgsLab5RestIDs))
}()

// pgsLab5DupState says how many rows of the title are left — not their ids
// while both are there: looking them up is part of the task — which row is
// left, and which other books are gone. What a DELETE took cannot be typed
// back with its id, so those cases end with the reset button, once.
var pgsLab5DupState = func() string {
	t := pgsLab5Title(pgsLab5Keep)
	name := pgsLab5Books[pgsLab5Keep-1].title
	cnt := "(SELECT count(*) FROM books WHERE title = " + t + ")"
	kept := "EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Keep) + " AND title = " + t + ")"
	lost := pgsLab5Lost(pgsLab5RestIDs)
	return "SELECT CASE " +
		"WHEN " + cnt + " > 1 THEN '«" + name + "» записан в каталоге несколько раз: строк с таким названием — ' || " + cnt + " " +
		"WHEN " + cnt + " = 0 THEN 'книги «" + name + "» в каталоге не осталось: удалены обе строки, а одну нужно было оставить' " +
		"WHEN NOT " + kept + " " +
		"THEN 'осталась одна строка «" + name + "», но с id ' || (SELECT min(id) FROM books WHERE title = " + t + ") || " +
		"', а оставить нужно было строку с меньшим id' " +
		"WHEN EXISTS (SELECT 1 FROM books WHERE id = " + strconv.Itoa(pgsLab5Gone) + ") " +
		"THEN 'строка с id " + strconv.Itoa(pgsLab5Gone) + " всё ещё в каталоге — лишнюю строку нужно удалить, а не переименовать' " +
		"ELSE 'дубль удалён' END || " +
		"coalesce('; из каталога пропали и другие книги: ' || " + lost + ", '') || " +
		"CASE WHEN " + cnt + " = 0 OR NOT " + kept + " OR " + lost + " IS NOT NULL THEN ' — нажми «Пересоздать»' ELSE '' END || " +
		pgsLab5Tx
}()
