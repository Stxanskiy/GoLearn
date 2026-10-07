package main

// Лабораторная 3 курса pg-start: создаём базу магазина (ch-pgs-lab3).
//
// The student creates pereplet and its "first version" (§5 of the course
// contract): authors, books with pages that becomes page_count and an isbn
// added by ALTER TABLE, the staging schema with import_log — then writes it
// all down as an idempotent /root/sql/schema.sql.
//
// One task is not in the contract's outline: setup leaves an intern's Excel-like
// draft, "Books", in the database root, and the student removes it. That is
// DROP TABLE and the case trap of quoted names from the theory chapter, and the
// draft sitting in root is also what tells a student who forgot \c pereplet
// that psql put them in the wrong database.
//
// The checks look at the catalog, not at how the student got there: any type
// spelling PostgreSQL normalises to the same name (int8, decimal(8,2),
// timestamptz) passes, and so does an identity BY DEFAULT. The "Сейчас:"
// diagnostics list the table's columns the way \d would, and say where a
// missing table went instead — another schema, other letter case, the
// database root.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab3.sh.

// pgsLab3Draft is what the intern left in the database root: a sheet of the
// shop's Excel file turned into a table as it was — Russian column names in
// quotes, the price as real, page counts as text. Real books, so the student
// recognises them when the catalogue is filled in properly in lab 4.
const pgsLab3Draft = `
DROP TABLE IF EXISTS "Books";
CREATE TABLE "Books" (
    "Название" text,
    "Автор"    text,
    "Цена"     real,
    "Страниц"  text
);
INSERT INTO "Books" VALUES
    ('Мастер и Маргарита',       'Михаил Булгаков',   690,   '480'),
    ('Преступление и наказание', 'Ф. М. Достоевский', 590.5, '608 стр.'),
    ('Вишнёвый сад',             'Чехов А. П.',       280,   '96'),
    ('Мёртвые души',             'Гоголь',            450,   'около 350');
`

// pgsLab3UserTables is the condition part shared by the root clean-up check and
// its diagnostic: the tables outside the system schemas.
const pgsLab3UserTables = "FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema')"

// pgsLab3Now is the "Сейчас" diagnostic for a table the student creates in
// pereplet (table may carry a schema, short is the bare name): the columns as
// \d lists them, or, when there is no such table, what there is instead — the
// same name in another schema or in other letter case, or the table in the
// database root, where psql connects without -d. It has no double quotes:
// check() escapes them in the message, and an escaped quote breaks $(…).
func pgsLab3Now(table, short string) string {
	cols := "SELECT coalesce((SELECT 'в таблице " + table + " столбцы: ' || string_agg(a.attname || ' ' || format_type(a.atttypid, a.atttypmod)" +
		" || CASE WHEN a.attidentity IN ('a', 'd') AND EXISTS (SELECT 1 FROM pg_constraint c WHERE c.conrelid = a.attrelid AND c.contype = 'p' AND c.conkey = ARRAY[a.attnum]) THEN ' (identity, первичный ключ)'" +
		" WHEN a.attidentity IN ('a', 'd') THEN ' (identity, но не первичный ключ)'" +
		" WHEN EXISTS (SELECT 1 FROM pg_constraint c WHERE c.conrelid = a.attrelid AND c.contype = 'p' AND a.attnum = ANY (c.conkey)) THEN ' (первичный ключ, но не identity)' ELSE '' END" +
		", ', ' ORDER BY a.attnum) FROM pg_attribute a WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped)," +
		" 'в базе pereplet таблицы " + table + " нет' || coalesce((SELECT ', но есть ' || string_agg(format('%I.%I', schemaname, tablename), ', ') || ' — не то имя или не та схема' " +
		pgsLab3UserTables + " AND lower(tablename) = '" + short + "'), ''))"
	inRoot := "SELECT '; а в базе root есть ' || string_agg(format('%I.%I', schemaname, tablename), ', ') || ' — похоже, команду выполнили не в той базе' " +
		pgsLab3UserTables + " AND tablename = '" + short + "'"
	return "Сейчас: $(glpg q pereplet " + b64(cols) + " || echo базы pereplet нет)$(glpg q root " + b64(inRoot) + ")"
}

// pgsLab3Want is the schema the ★ task's file must build, column for column:
// the first version of §5, books already with page_count and isbn.
const pgsLab3Want = `VALUES
    ('authors', 'id', 'bigint'), ('authors', 'full_name', 'text'),
    ('authors', 'birth_year', 'integer'), ('authors', 'country', 'text'),
    ('books', 'id', 'bigint'), ('books', 'title', 'text'), ('books', 'price', 'numeric(8,2)'),
    ('books', 'page_count', 'integer'), ('books', 'published_year', 'integer'),
    ('books', 'in_stock', 'boolean'), ('books', 'isbn', 'text'),
    ('staging.import_log', 'id', 'bigint'), ('staging.import_log', 'loaded_at', 'timestamp with time zone'),
    ('staging.import_log', 'row_count', 'integer')`

// pgsLab3Diff compares a database with pgsLab3Want. It prints "ok:" and then
// what differs, if anything; a failed query prints nothing, which the check
// reads as a failure too.
const pgsLab3Diff = `WITH want (t, c, ty) AS (` + pgsLab3Want + `),
tabs AS (SELECT DISTINCT t, to_regclass(t) AS oid FROM want),
got AS (
    SELECT tabs.t, a.attname::text AS c, format_type(a.atttypid, a.atttypmod) AS ty
    FROM tabs JOIN pg_attribute a ON a.attrelid = tabs.oid AND a.attnum > 0 AND NOT a.attisdropped
)
SELECT 'ok:' || concat_ws('; ',
    (SELECT 'нет таблиц: ' || string_agg(t, ', ' ORDER BY t) FROM tabs WHERE oid IS NULL),
    (SELECT 'нет столбцов: ' || string_agg(w.t || '.' || w.c, ', ') FROM want w JOIN tabs USING (t)
        WHERE tabs.oid IS NOT NULL AND NOT EXISTS (SELECT 1 FROM got g WHERE g.t = w.t AND g.c = w.c)),
    (SELECT 'не тот тип: ' || string_agg(g.t || '.' || g.c || ' ' || g.ty, ', ') FROM got g JOIN want w USING (t, c) WHERE g.ty <> w.ty),
    (SELECT 'лишние столбцы: ' || string_agg(g.t || '.' || g.c, ', ') FROM got g
        WHERE NOT EXISTS (SELECT 1 FROM want w WHERE w.t = g.t AND w.c = g.c)),
    (SELECT 'id не identity или не первичный ключ: ' || string_agg(t, ', ') FROM tabs
        WHERE oid IS NOT NULL AND NOT EXISTS (
            SELECT 1 FROM pg_attribute a JOIN pg_constraint k ON k.conrelid = a.attrelid AND k.contype = 'p' AND k.conkey = ARRAY[a.attnum]
            WHERE a.attrelid = tabs.oid AND a.attname = 'id' AND a.attidentity IN ('a', 'd'))))`

// pgsLab3Schema judges /root/sql/schema.sql: run it twice in a scratch
// database, stopping at the first error, and compare what it built with
// pgsLab3Want. The scratch database is dropped on every way out.
//
// The file is refused without being run if it connects anywhere (\c) or
// touches a whole database: run as a superuser, a DROP DATABASE pereplet in it
// would take the student's own work. Comments are stripped first — "-- базу
// создаём через createdb" is fine. Errors are also looked for in the output, so
// a file that turns ON_ERROR_STOP off does not hide them.
var pgsLab3Schema = `f=/root/sql/schema.sql; t=pgs_lab3_check_$$; ` +
	`[ -s $f ] || { echo "файла $f нет или он пустой"; exit 1; }; ` +
	`bad=$(sed 's/--.*$//' $f | grep -oiE '\\(c|connect)([[:space:]]|$)|(create|drop|alter)[[:space:]]+database' | head -n1 | sed 's/[[:space:]]*$//'); ` +
	`[ -z "$bad" ] || { echo "в $f есть «$bad»: скрипт схемы не переключает, не создаёт и не удаляет базы — в какой базе работать, задают при запуске: psql -d"; exit 1; }; ` +
	`trap 'dropdb --if-exists --force $t >/dev/null 2>&1' EXIT; ` +
	`createdb -T template0 $t || { echo "не удалось создать проверочную базу"; exit 1; }; ` +
	`for i in первом втором; do ` +
	`e=$(psql -X -q -v ON_ERROR_STOP=1 -d $t -f $f 2>&1 >/dev/null); r=$?; ` +
	`if [ $r != 0 ] || printf '%s\n' "$e" | grep -q 'ERROR:'; then ` +
	`echo "при $i запуске ошибка: $(printf '%s\n' "$e" | grep -m1 -E 'ERROR|FATAL' | sed -E 's/^psql:[^:]*:([0-9]+): /строка \1: /')"; exit 1; fi; done; ` +
	`p=$(glpg q $t ` + b64(pgsLab3Diff) + `); ` +
	`[ -n "$p" ] || { echo "сверить таблицы не удалось — сообщи авторам курса"; exit 1; }; [ "$p" = ok: ] || { echo "оба запуска прошли без ошибок, но после них ${p#ok:}"; exit 1; }`

func init() {
	pgStartLabs["ch-pgs-lab3"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgSetup("dropdb --if-exists --force pereplet\nmkdir -p /root/sql\n" + pgRun("root", pgsLab3Draft)),
		Checks: map[int]string{
			1: check(pgDBExists("pereplet"),
				"база pereplet создана",
				"Сейчас баз на сервере, не считая шаблонов: "+pgVal("postgres", "SELECT string_agg(datname, ', ' ORDER BY datname) FROM pg_database WHERE NOT datistemplate")),
			2: check(pgIdentityPK("pereplet", "authors")+" && "+pgHasCol("pereplet", "authors", "id", "bigint")+" && "+
				pgHasCol("pereplet", "authors", "full_name", "text")+" && "+pgHasCol("pereplet", "authors", "birth_year", "integer")+" && "+
				pgHasCol("pereplet", "authors", "country", "text"),
				"в базе pereplet есть таблица authors: id bigint (identity, первичный ключ), full_name text, birth_year integer, country text",
				pgsLab3Now("authors", "authors")),
			// pages becomes page_count in task 5; this check must not turn red
			// then, so either name passes.
			3: check(pgIdentityPK("pereplet", "books")+" && "+pgHasCol("pereplet", "books", "id", "bigint")+" && "+
				pgHasCol("pereplet", "books", "title", "text")+" && "+pgHasCol("pereplet", "books", "price", "numeric(8,2)")+" && "+
				"{ "+pgHasCol("pereplet", "books", "pages", "integer")+" || "+pgHasCol("pereplet", "books", "page_count", "integer")+"; } && "+
				pgHasCol("pereplet", "books", "published_year", "integer")+" && "+pgHasCol("pereplet", "books", "in_stock", "boolean"),
				"в базе pereplet есть таблица books: id bigint (identity, первичный ключ), title text, price numeric(8,2), pages (или уже page_count) integer, published_year integer, in_stock boolean",
				pgsLab3Now("books", "books")),
			4: check(pgHasCol("pereplet", "books", "isbn", "text"),
				"в таблице books есть столбец isbn типа text",
				pgsLab3Now("books", "books")),
			5: check(pgHasCol("pereplet", "books", "page_count", "integer")+" && "+pgIs("pereplet", pgColType("books", "pages"), ""),
				"в таблице books столбец page_count integer, а столбца pages нет",
				pgsLab3Now("books", "books")),
			6: check(pgIdentityPK("pereplet", "staging.import_log")+" && "+pgHasCol("pereplet", "staging.import_log", "id", "bigint")+" && "+
				pgHasCol("pereplet", "staging.import_log", "loaded_at", "timestamp with time zone")+" && "+
				pgHasCol("pereplet", "staging.import_log", "row_count", "integer"),
				"в базе pereplet есть схема staging и в ней таблица import_log: id bigint (identity, первичный ключ), loaded_at timestamptz, row_count integer",
				pgsLab3Now("staging.import_log", "import_log")),
			// Every table goes, not just the draft: one made in root by mistake
			// earlier in the lab is the same mess. Renaming the draft or moving
			// it to another schema leaves a table, so it does not pass.
			7: check(pgIs("root", "SELECT count(*) "+pgsLab3UserTables, "0"),
				"в базе root не осталось таблиц: черновик стажёра удалён",
				"Сейчас: $(glpg q root "+b64("SELECT coalesce('в базе root таблицы: ' || string_agg(CASE WHEN schemaname = 'public' THEN tablename ELSE schemaname || '.' || tablename END, ', ' ORDER BY schemaname, tablename), 'в базе root таблиц нет') "+pgsLab3UserTables)+" || echo базы root нет)"),
			8: pgCheck(pgsLab3Schema,
				"/root/sql/schema.sql — без \\c и CREATE DATABASE, два запуска подряд проходят без ошибок, создаются authors, books и staging.import_log ровно с нужными столбцами"),
		},
	}
}
