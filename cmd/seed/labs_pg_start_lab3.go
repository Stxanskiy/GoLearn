package main

// Lab 3 of pg-start: the first tables (ch-pgs-lab3).
//
// The student creates authors and books in pereplet the way lesson 6 teaches
// (state S3 of the course plan), adds books.published_year with ALTER TABLE and
// drops bookz — a table of books an intern started and left with a typo in its
// name and an integer price. The quizzes read \d of the student's own authors,
// the error a repeated CREATE TABLE gives, and \d bookz before it goes.
//
// The checks look at the catalog, never at how the table was typed:
//   - columns are compared as "name type" the way format_type prints them, so
//     int, int4 and integer are all integer. When they differ, the diagnostic
//     lists the columns as they are and names the ones that are missing, extra
//     or of another type — never the type a column should have;
//   - id has to be the primary key and numbered by the database: serial
//     (DEFAULT nextval) and an identity column both pass;
//   - published_year has to be added to a table that already exists. A column
//     written into CREATE TABLE gets its pg_attribute row in the same
//     transaction as the table's system columns, whose rows nothing rewrites
//     later (renaming the table or other columns, VACUUM FULL or FREEZE); ADD
//     COLUMN writes it in a later one, so the xmins differ. Edges, none taught
//     in lesson 6: anything that rewrites the column's own row — ALTER COLUMN
//     … TYPE, renaming published_year away and back — passes (an ALTER TABLE
//     all the same), and CREATE plus ADD COLUMN inside one BEGIN … COMMIT
//     fails. The books check of task 2 leaves published_year out, so it stays
//     green after task 3;
//   - the last task wants bookz gone and no table left but authors and books in
//     any schema of the student's, so neither renaming bookz nor moving it out
//     of public passes for dropping it.
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.

import (
	"sort"
	"strconv"
	"strings"
)

func init() {
	pgStartLabs["ch-pgs-lab3"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup(pgFresh("pereplet") + pgRun("pereplet", pgsLab3Schema)),
		Checks: map[int]string{
			1: check(pgTrue("pereplet", pgsLab3Table("authors", "", pgsLab3Authors)),
				"в базе pereplet есть таблица authors: id serial PRIMARY KEY, full_name text, birth_year integer, country text",
				"Сейчас: "+pgsLab3Now(pgsLab3State("authors", "", pgsLab3Authors))+pgVal("postgres", pgsLab3Elsewhere("authors"))),
			2: check(pgTrue("pereplet", pgsLab3Table("books", "published_year", pgsLab3Books)),
				"в базе pereplet есть таблица books: id, title, price, page_count, in_stock — с нужными типами",
				"Сейчас: "+pgsLab3Now(pgsLab3State("books", "published_year", pgsLab3Books))+pgVal("postgres", pgsLab3Elsewhere("books"))),
			3: check(pgTrue("pereplet", "SELECT EXISTS ("+pgsLab3Year+" AND format_type(p.atttypid, p.atttypmod) = 'integer' AND p.xmin <> s.xmin)"),
				"в таблицу books командой ALTER TABLE добавлен столбец published_year integer",
				"Сейчас: "+pgsLab3Now(pgsLab3YearState)),
			4: check(pgTrue("pereplet", "SELECT NOT EXISTS (SELECT 1 FROM "+pgsLab3UserTables+" AND NOT (schemaname = 'public' AND tablename IN ('authors', 'books')))"),
				"таблица bookz удалена, лишних таблиц в базе pereplet нет",
				"Сейчас: "+pgsLab3Now("SELECT 'таблицы в базе pereplet: ' || "+pgsLab3Tables)),
		},
	}
}

// pgsLab3Schema is the table the intern left: a typo in the name and an integer
// price. The quiz before the last task asks what, besides the name, is wrong.
const pgsLab3Schema = `CREATE TABLE bookz (
    id    serial PRIMARY KEY,
    title text,
    price integer
);
`

// pgsLab3Col is a column a task asks for, its type as format_type prints it.
type pgsLab3Col struct{ name, typ string }

// The columns of tasks 1 and 2, in the order the tasks list them.
var (
	pgsLab3Authors = []pgsLab3Col{{"id", "integer"}, {"full_name", "text"}, {"birth_year", "integer"}, {"country", "text"}}
	pgsLab3Books   = []pgsLab3Col{{"id", "integer"}, {"title", "text"}, {"price", "numeric(8,2)"}, {"page_count", "integer"}, {"in_stock", "boolean"}}
)

// pgsLab3UserTables is the FROM of the student's tables: everything \dt could
// ever list, whatever the schema, without the server's own.
const pgsLab3UserTables = `pg_tables WHERE schemaname <> 'information_schema' AND schemaname NOT LIKE 'pg\_%'`

// pgsLab3Tables is SQL for the student's tables by name, as \dt lists them (a
// table outside public with its schema).
const pgsLab3Tables = "coalesce((SELECT string_agg(CASE WHEN schemaname = 'public' THEN tablename " +
	"ELSE schemaname || '.' || tablename END, ', ' ORDER BY schemaname <> 'public', schemaname, tablename) " +
	"FROM " + pgsLab3UserTables + "), 'ни одной')"

// pgsLab3Now is a substitution printing a query's value in pereplet, or saying
// that pereplet cannot be reached (the student dropped it, most likely) rather
// than leaving "Сейчас:" empty.
func pgsLab3Now(sql string) string {
	return `$(glpg q pereplet ` + b64(sql) + ` || echo 'не получается подключиться к базе pereplet (её удалили?) — нажми «Пересоздать», и она вернётся')`
}

// pgsLab3Table is SQL true when table has exactly the columns want (leaving out
// the column skip), id is its primary key and the database numbers it.
func pgsLab3Table(table, skip string, want []pgsLab3Col) string {
	cols := make([]string, len(want))
	for i, c := range want {
		cols[i] = c.name + " " + c.typ
	}
	sort.Strings(cols) // byte order, as the C.UTF-8 database sorts attname
	return "SELECT " + pgsLab3Sorted(table, skip) + " = '" + strings.Join(cols, ",") + "' AND " +
		pgsLab3PK(table) + " AND " + pgsLab3Auto(table)
}

// pgsLab3Sorted is SQL for table's columns as "name type", sorted by name and
// joined with commas, leaving out the column skip: the shape the checks compare.
func pgsLab3Sorted(table, skip string) string {
	return "(SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod), ',' ORDER BY a.attname) " +
		"FROM pg_attribute a WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 " +
		"AND NOT a.attisdropped AND a.attname <> '" + skip + "')"
}

// pgsLab3Shown is SQL for table's columns as "name type" in the order \d shows
// them: what the student is told the table has now.
func pgsLab3Shown(table string) string {
	return "coalesce((SELECT string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod), ', ' ORDER BY a.attnum) " +
		"FROM pg_attribute a WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped), 'ни одного')"
}

// pgsLab3Diff is SQL naming the columns of table that differ from want — the
// missing ones, the extra ones (skip is not extra) and those of another type —
// or NULL when they match. It never says what the type should be.
func pgsLab3Diff(table, skip string, want []pgsLab3Col) string {
	vals := make([]string, len(want))
	for i, c := range want {
		vals[i] = "('" + c.name + "', '" + c.typ + "', " + strconv.Itoa(i+1) + ")"
	}
	w := "(VALUES " + strings.Join(vals, ", ") + ") w(name, typ, n)"
	h := "(SELECT a.attname::text AS name, format_type(a.atttypid, a.atttypmod) AS typ, a.attnum AS n FROM pg_attribute a " +
		"WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped AND a.attname <> '" + skip + "') h"
	return "nullif(concat_ws('; ', " +
		"(SELECT 'не хватает столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " WHERE w.name NOT IN (SELECT h.name FROM " + h + ")), " +
		"(SELECT 'лишние столбцы: ' || string_agg(h.name, ', ' ORDER BY h.n) FROM " + h + " WHERE h.name NOT IN (SELECT w.name FROM " + w + ")), " +
		"(SELECT 'не тот тип у столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " JOIN " + h + " ON h.name = w.name WHERE h.typ <> w.typ)" +
		"), '')"
}

// pgsLab3PK is SQL true when table.id on its own is the primary key.
func pgsLab3PK(table string) string {
	return "EXISTS (SELECT 1 FROM pg_constraint c JOIN pg_attribute a " +
		"ON a.attrelid = c.conrelid AND c.conkey = ARRAY[a.attnum] " +
		"WHERE c.conrelid = to_regclass('" + table + "') AND c.contype = 'p' AND a.attname = 'id')"
}

// pgsLab3Auto is SQL true when the database numbers table.id itself: serial
// (DEFAULT nextval) or an identity column.
func pgsLab3Auto(table string) string {
	return "EXISTS (SELECT 1 FROM pg_attribute a LEFT JOIN pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum " +
		"WHERE a.attrelid = to_regclass('" + table + "') AND a.attname = 'id' AND NOT a.attisdropped " +
		"AND (a.attidentity IN ('a', 'd') OR pg_get_expr(d.adbin, d.adrelid) LIKE 'nextval(%'))"
}

// pgsLab3State is SQL describing table in pereplet for a "Сейчас:" line: that
// it is missing and which tables there are, or its columns, what differs from
// want and how its id is numbered. A missing id is among the missing columns.
func pgsLab3State(table, skip string, want []pgsLab3Col) string {
	return "SELECT CASE WHEN to_regclass('" + table + "') IS NULL " +
		"THEN 'в базе pereplet нет таблицы " + table + ", таблицы в базе: ' || " + pgsLab3Tables + " " +
		"ELSE 'столбцы " + table + ": ' || " + pgsLab3Shown(table) + " || coalesce('; ' || " + pgsLab3Diff(table, skip, want) + ", '') || CASE " +
		"WHEN (" + pgColType(table, "id") + ") IS NULL THEN '' " +
		"WHEN NOT " + pgsLab3PK(table) + " THEN '; id — не первичный ключ' " +
		"WHEN NOT " + pgsLab3Auto(table) + " THEN '; id — первичный ключ, но номер в нём база сама не ставит' " +
		"ELSE '; id — первичный ключ с автонумерацией' END END"
}

// pgsLab3Elsewhere is SQL, run in the database postgres, that notices the
// table made there — the "не та база" mistake lesson 6 warns about.
func pgsLab3Elsewhere(table string) string {
	return "SELECT CASE WHEN to_regclass('" + table + "') IS NOT NULL " +
		"THEN '; таблица " + table + " нашлась в базе postgres — проверь приглашение: работать нужно в pereplet=#' ELSE '' END"
}

// pgsLab3Year finds books.published_year (p) next to the table's ctid (s): a
// system column, written by CREATE TABLE and never rewritten afterwards.
const pgsLab3Year = "SELECT 1 FROM pg_attribute p JOIN pg_attribute s ON s.attrelid = p.attrelid AND s.attnum = -1 " +
	"WHERE p.attrelid = to_regclass('books') AND p.attname = 'published_year' AND NOT p.attisdropped"

// pgsLab3YearState says what is missing for task 3. The last branch is the
// column written straight into CREATE TABLE.
var pgsLab3YearState = "SELECT CASE " +
	"WHEN to_regclass('books') IS NULL THEN 'в базе pereplet нет таблицы books, таблицы в базе: ' || " + pgsLab3Tables + " " +
	"WHEN (" + pgColType("books", "published_year") + ") IS NULL THEN 'в books нет столбца published_year, столбцы books: ' || " + pgsLab3Shown("books") + " " +
	"WHEN (" + pgColType("books", "published_year") + ") <> 'integer' " +
	"THEN 'у published_year тип ' || (" + pgColType("books", "published_year") + ") || ', а нужен integer' " +
	"ELSE 'published_year создан вместе с таблицей в CREATE TABLE, а в этом задании его нужно добавить в готовую таблицу командой ALTER TABLE' END"
