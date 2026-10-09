package main

// Lab 4 of pg-start: filling the catalog (ch-pgs-lab4).
//
// The setup is state S3 of the course plan — authors and books as lab 3 leaves
// them — with part of the catalog already in: four authors and ten books. The
// student adds rows with INSERT the way lesson 8 teaches: one author, three
// authors with a single command, two books, a book of which only the title and
// the price are known (the rest stays NULL), and, as the ★ task, an intern's
// INSERT with the three mistakes lesson 8 lists (double quotes, a decimal
// comma, an id of its own). SELECT is the subject of the quizzes: the student
// runs a query and picks what it shows.
//
// The checks look at the rows, never at how they were typed:
//   - each expected row is found by its key (full_name, title) and has to be
//     there exactly once with every value as the task gives it, NULL included
//     (IS NOT DISTINCT FROM). The student cannot UPDATE or DELETE yet — that is
//     lesson 10 — so a row added twice or with a wrong value is fixed with
//     «Пересоздать», and the diagnostic says so;
//   - "three authors with one INSERT": rows written by one statement share the
//     xmin of its transaction, three INSERTs in autocommit give three. Three
//     INSERTs between BEGIN and COMMIT (lesson 10) share one too and pass;
//   - a key that differs only in letter case or spaces is named as a near miss:
//     text is compared exactly, which lesson 8 stresses. Like a wrong value it
//     sends the student to «Пересоздать»;
//   - the intern's book has to get its id from the sequence: an id typed by
//     hand past what the sequence has handed out fails, and so does 0 or a
//     negative one, a beginner's guess after "Key (id)=(1) already exists"
//     (the setup's rows hold 1-10, so no free id the sequence gave is below 1).
//     The very next number typed by hand before any other book goes in cannot
//     be told apart once the sequence has moved past it, and passes;
//   - every check reads only its own rows, so the tasks pass in any order.
//
// The quizzes have one right answer on the setup data whatever the student
// adds: every author the tasks name was born after 1769, every book they name
// costs from 380 to 870 (none is the dearest, none is under 300), and none is
// «Дети капитана Гранта».
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.

import "strings"

func init() {
	pgStartLabs["ch-pgs-lab4"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab4Schema)),
		Checks: map[int]string{
			1: check(pgTrue("pereplet", "SELECT "+pgsLab4Done(pgsLab4Chekhov)),
				"в таблице authors есть автор Антон Чехов (1860, Россия), один раз",
				"Сейчас: "+pgsLab4Now("NULL", pgsLab4Broken(pgsLab4Chekhov), pgsLab4Chekhov)),
			2: check(pgTrue("pereplet", "SELECT "+pgsLab4Done(pgsLab4Three...)+" AND "+pgsLab4OneInsert),
				"в таблице authors есть Лев Толстой, Марк Твен и Артур Конан Дойл, добавленные одной командой INSERT",
				"Сейчас: "+pgsLab4Now(pgsLab4ThreeApart, pgsLab4ThreeRedo, pgsLab4Three...)),
			3: check(pgTrue("pereplet", "SELECT "+pgsLab4Done(pgsLab4Books...)),
				"в таблице books есть «Собака Баскервилей» и «Приключения Тома Сойера» с данными из задания, по одному разу",
				"Сейчас: "+pgsLab4Now("NULL", pgsLab4Broken(pgsLab4Books...), pgsLab4Books...)),
			4: check(pgTrue("pereplet", "SELECT "+pgsLab4Done(pgsLab4Karenina)),
				"в таблице books есть «Анна Каренина» с ценой 870.00, а page_count, in_stock и published_year пустые (NULL)",
				"Сейчас: "+pgsLab4Now("NULL", pgsLab4Broken(pgsLab4Karenina), pgsLab4Karenina)),
			5: check(pgTrue("pereplet", "SELECT "+pgsLab4Done(pgsLab4LostWorld)),
				"в таблице books есть «Затерянный мир» (430.00, 320 страниц, на складе, 1912) с номером, который поставила база",
				"Сейчас: "+pgsLab4Now("NULL", pgsLab4Broken(pgsLab4LostWorld), pgsLab4LostWorld)),
		},
	}
}

// pgsLab4Schema is state S3 with the catalog the lab starts from. The rows go
// in without ids, so the sequences stand where the student's rows continue.
const pgsLab4Schema = `CREATE TABLE authors (
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
    ('Жюль Верн',         1828, 'Франция');
INSERT INTO books (title, price, page_count, in_stock, published_year) VALUES
    ('Басни',                            240.00,  224, true,  1809),
    ('Евгений Онегин',                   420.00,  320, true,  1833),
    ('Капитанская дочка',                210.00,  192, true,  1836),
    ('Повести Белкина',                  190.00,  160, false, 1831),
    ('Идиот',                            650.00,  640, true,  1869),
    ('Братья Карамазовы',                990.00,  992, false, 1880),
    ('Белые ночи',                       150.00,   96, true,  1848),
    ('Таинственный остров',             1350.00,  704, true,  1875),
    ('Вокруг света за восемьдесят дней', 360.00,  320, false, 1872),
    ('Дети капитана Гранта',             560.00,  672, true,  1868);
`

// pgsLab4Row is a row a task asks for: found in table by its key column, with
// the values cols lists. A want is an SQL literal as the task writes it, NULL
// for "leave it empty".
type pgsLab4Row struct {
	table, keyCol, key string
	cols               []pgsLab4Col
	// autoID: id has to come from the sequence, not be typed by hand.
	autoID bool
}

type pgsLab4Col struct {
	name, want string
	text       bool // shown quoted in the diagnostic, the way it is typed
}

func pgsLab4Author(name, year, country string) pgsLab4Row {
	return pgsLab4Row{table: "authors", keyCol: "full_name", key: name, cols: []pgsLab4Col{
		{name: "birth_year", want: year},
		{name: "country", want: "'" + country + "'", text: true},
	}}
}

func pgsLab4Book(title, price, pages, stock, year string) pgsLab4Row {
	return pgsLab4Row{table: "books", keyCol: "title", key: title, cols: []pgsLab4Col{
		{name: "price", want: price},
		{name: "page_count", want: pages},
		{name: "in_stock", want: stock},
		{name: "published_year", want: year},
	}}
}

var (
	pgsLab4Chekhov = pgsLab4Author("Антон Чехов", "1860", "Россия")
	pgsLab4Three   = []pgsLab4Row{
		pgsLab4Author("Лев Толстой", "1828", "Россия"),
		pgsLab4Author("Марк Твен", "1835", "США"),
		pgsLab4Author("Артур Конан Дойл", "1859", "Великобритания"),
	}
	pgsLab4Books = []pgsLab4Row{
		pgsLab4Book("Собака Баскервилей", "450.00", "256", "true", "1902"),
		pgsLab4Book("Приключения Тома Сойера", "380.00", "288", "false", "1876"),
	}
	pgsLab4Karenina  = pgsLab4Book("Анна Каренина", "870.00", "NULL", "NULL", "NULL")
	pgsLab4LostWorld = func() pgsLab4Row {
		r := pgsLab4Book("Затерянный мир", "430.00", "320", "true", "1912")
		r.autoID = true
		return r
	}()
)

// pgsLab4OneInsert is SQL true when the three authors of task 2 share one
// xmin: one statement wrote them.
var pgsLab4OneInsert = "(SELECT count(DISTINCT xmin::text) FROM authors WHERE full_name IN (" +
	pgsLab4Keys(pgsLab4Three) + ")) = 1"

// pgsLab4ThreeApart names what task 2 needs beyond its rows: one command for
// all three, which neither three INSERTs nor adding the rest afterwards gives.
var pgsLab4ThreeApart = "CASE WHEN " + pgsLab4Done(pgsLab4Three...) + " AND NOT " + pgsLab4OneInsert +
	" THEN 'все трое в таблице, но добавлены разными командами INSERT, а нужно одной' " +
	"WHEN " + pgsLab4Three[0].count() + " + " + pgsLab4Three[1].count() + " + " + pgsLab4Three[2].count() + " > 0 " +
	"AND " + pgsLab4Three[0].count() + " * " + pgsLab4Three[1].count() + " * " + pgsLab4Three[2].count() + " = 0 " +
	"THEN 'троих нужно добавить одной командой, а часть из них уже в таблице' END"

// pgsLab4ThreeRedo: for task 2 any of the three already in (even misspelt),
// while the task is not done, means starting over — adding the rest would not
// make it one INSERT.
var pgsLab4ThreeRedo = "NOT (" + pgsLab4Done(pgsLab4Three...) + " AND " + pgsLab4OneInsert + ") AND ((" +
	pgsLab4Three[0].count() + " + " + pgsLab4Three[1].count() + " + " + pgsLab4Three[2].count() + ") > 0 OR " +
	pgsLab4NearMiss(pgsLab4Three...) + ")"

func pgsLab4Lit(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func pgsLab4Keys(rows []pgsLab4Row) string {
	var keys []string
	for _, r := range rows {
		keys = append(keys, pgsLab4Lit(r.key))
	}
	return strings.Join(keys, ", ")
}

func (r pgsLab4Row) where() string { return r.keyCol + " = " + pgsLab4Lit(r.key) }

func (r pgsLab4Row) count() string {
	return "(SELECT count(*) FROM " + r.table + " WHERE " + r.where() + ")"
}

// seqTop is SQL for the last id the table's sequence handed out (0 if none).
func (r pgsLab4Row) seqTop() string {
	return "coalesce(pg_sequence_last_value(pg_get_serial_sequence('" + r.table + "', 'id')::regclass), 0)"
}

// match is SQL true for a row that has every value the task gives.
func (r pgsLab4Row) match() string {
	var conds []string
	for _, c := range r.cols {
		conds = append(conds, c.name+" IS NOT DISTINCT FROM "+c.want)
	}
	if r.autoID {
		conds = append(conds, "id BETWEEN 1 AND "+r.seqTop())
	}
	return strings.Join(conds, " AND ")
}

func (r pgsLab4Row) done() string {
	return r.count() + " = 1 AND EXISTS (SELECT 1 FROM " + r.table + " WHERE " + r.where() + " AND " + r.match() + ")"
}

// problem is SQL text saying what is wrong with a row that is in the table,
// NULL when nothing is (or when it is simply missing — see missing): there
// with other letters or spaces in its key, more than once, or with values other
// than the task's, each named as it is and as it should be.
func (r pgsLab4Row) problem() string {
	var diffs []string
	for _, c := range r.cols {
		got := "coalesce(" + c.name + "::text, 'NULL')"
		if c.text {
			got = "quote_nullable(" + c.name + ")"
		}
		diffs = append(diffs, "CASE WHEN "+c.name+" IS DISTINCT FROM "+c.want+" THEN '"+c.name+" = ' || "+got+
			" || ' вместо "+strings.ReplaceAll(c.want, "'", "''")+"' END")
	}
	if r.autoID {
		diffs = append(diffs, "CASE WHEN id NOT BETWEEN 1 AND "+r.seqTop()+" THEN 'номер id = ' || id || ' вписан вручную, а его должна поставить база' END")
	}
	name := "«" + r.key + "»"
	return "CASE " +
		"WHEN " + r.count() + " = 0 THEN " +
		"(SELECT 'есть ' || quote_literal(" + r.keyCol + ") || ', а нужно ' || " + pgsLab4Lit(pgsLab4Lit(r.key)) +
		" || ': текст сравнивается точно, буква в букву' FROM " + r.table + " WHERE " + r.alike() + " ORDER BY id LIMIT 1) " +
		"WHEN " + r.count() + " > 1 THEN " + pgsLab4Lit(name+" в таблице "+r.table+" повторяется (строк: ") + " || " + r.count() + " || ')' " +
		"ELSE (SELECT " + pgsLab4Lit(name+": ") + " || nullif(concat_ws(', ', " + strings.Join(diffs, ", ") + "), '') " +
		"FROM " + r.table + " WHERE " + r.where() + ") END"
}

// alike is SQL true for a key that differs from the task's only in letter case
// or spaces — the near miss problem() points out.
func (r pgsLab4Row) alike() string {
	return "lower(regexp_replace(btrim(" + r.keyCol + "), '\\s+', ' ', 'g')) = lower(" + pgsLab4Lit(r.key) + ")"
}

// missing is SQL text naming the row when it is not in the table at all, not
// even as a near miss; NULL otherwise.
func (r pgsLab4Row) missing() string {
	return "CASE WHEN " + r.count() + " = 0 AND NOT EXISTS (SELECT 1 FROM " + r.table + " WHERE " + r.alike() + ") " +
		"THEN " + pgsLab4Lit("«"+r.key+"»") + " END"
}

// pgsLab4Done is SQL true when each row is there once with its values.
func pgsLab4Done(rows ...pgsLab4Row) string {
	var conds []string
	for _, r := range rows {
		conds = append(conds, r.done())
	}
	return strings.Join(conds, " AND ")
}

// nearMiss is SQL true when the row is in only as a near miss (see alike).
func (r pgsLab4Row) nearMiss() string {
	return "(" + r.count() + " = 0 AND EXISTS (SELECT 1 FROM " + r.table + " WHERE " + r.alike() + "))"
}

// pgsLab4NearMiss is SQL true when any of rows is in only as a near miss.
func pgsLab4NearMiss(rows ...pgsLab4Row) string {
	var conds []string
	for _, r := range rows {
		conds = append(conds, r.nearMiss())
	}
	return strings.Join(conds, " OR ")
}

// pgsLab4Broken is SQL true when a row is in but wrong — twice, with other
// values, or with its key misspelt: only starting over fixes that before
// lesson 10. (Adding the right row next to a misspelt one would pass too, but
// leaves a stray row in the catalog; the diagnostic names the one clean way.)
func pgsLab4Broken(rows ...pgsLab4Row) string {
	var conds []string
	for _, r := range rows {
		conds = append(conds, "("+r.count()+" > 0 AND NOT ("+r.done()+"))", r.nearMiss())
	}
	return strings.Join(conds, " OR ")
}

// pgsLab4Now is a task's "Сейчас:" diagnostic: what is wrong with each of its
// rows, then extra (SQL text, NULL when there is nothing to add), and — when
// redo holds — how to start over, since fixing and deleting rows comes only in
// the next lesson. If the server cannot answer it says that instead of nothing.
func pgsLab4Now(extra, redo string, rows ...pgsLab4Row) string {
	var miss, probs []string
	for _, r := range rows {
		miss = append(miss, r.missing())
		probs = append(probs, r.problem())
	}
	absent := pgsLab4Lit("в таблице "+rows[0].table+" нет: ") + " || nullif(concat_ws(', ', " + strings.Join(miss, ", ") + "), '')"
	sql := "SELECT coalesce(nullif(concat_ws('; ', " + absent + ", " + strings.Join(probs, ", ") + ", " + extra + "), ''), " +
		"'расхождений не видно, нажми «Проверить» ещё раз') || CASE WHEN " + redo +
		" THEN '. Исправлять и удалять строки ты научишься в следующем уроке, а сейчас нажми «Пересоздать» и выполни задание заново' ELSE '' END"
	return `$(glpg q pereplet ` + b64(sql) + ` || echo 'не удалось прочитать таблицы authors и books: их нет или сервер не отвечает. Нажми «Пересоздать»')`
}
