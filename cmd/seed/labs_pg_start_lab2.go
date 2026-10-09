package main

// Lab 2 of pg-start: working in psql (ch-pgs-lab2).
//
// The server looks the way lab 1 leaves it — root is a superuser, pereplet is
// empty — plus two databases a colleague brought over from the shop's old
// server, both owned by postgres: archive (one table, old_orders, for the
// student to find with \c and \dt) and old_shop (to be dropped). The quiz
// answers come from what psql shows on this state and do not change with the
// tasks: the port in \conninfo, archive's owner in \l, the table in archive,
// the "invalid command \l;" error.
//
// The tasks leave state behind, so they are checked by it: a database sandbox
// owned by root (CREATE DATABASE from the student's own psql), old_shop gone
// while pereplet and archive — which the task says to leave alone — stay, and
// archive_2025 owned by postgres (the OWNER clause, found in \h CREATE
// DATABASE). The setup's databases carry a description (COMMENT ON DATABASE,
// seen only in \l+), and the checks recognise them by it as well as by name,
// so renaming one is not taken for creating or dropping a database: old_shop
// renamed is still there, and archive renamed to archive_2025 is not new.
//
// This is the lab where the student drops databases, so one of them may drop
// postgres, the database every check connects to. The diagnostic then says so
// (pgsLab2Now) instead of reporting a server that is in fact running.
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.

func init() {
	pgStartLabs["ch-pgs-lab2"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgsLab2Setup,
		Checks: map[int]string{
			1: check(pgTrue("postgres", pgsLab2NewDB("sandbox", "root")),
				"база sandbox создана, её владелец — root",
				"Сейчас: "+pgsLab2Now(pgsLab2DBState("sandbox", "root"))),
			2: check(pgTrue("postgres", pgsLab2Dropped),
				"база old_shop удалена, а pereplet и archive на месте",
				"Сейчас: "+pgsLab2Now(pgsLab2OldShopState)),
			3: check(pgTrue("postgres", pgsLab2NewDB("archive_2025", "postgres")),
				"база archive_2025 создана, её владелец — postgres",
				"Сейчас: "+pgsLab2Now(pgsLab2DBState("archive_2025", "postgres"))),
		},
	}
}

// The descriptions of the databases the setup creates.
const (
	pgsLab2PerepletDesc = "Книжный магазин «Переплёт»"
	pgsLab2ArchiveDesc  = "Архив заказов старого магазина, 2018–2019"
	pgsLab2OldShopDesc  = "Старый интернет-магазин, больше не нужен"
)

// pgsLab2Origin is the SQL for the name a pg_database row had in the setup:
// NULL for a database the setup did not create.
const pgsLab2Origin = "(CASE shobj_description(oid, 'pg_database')" +
	" WHEN '" + pgsLab2PerepletDesc + "' THEN 'pereplet'" +
	" WHEN '" + pgsLab2ArchiveDesc + "' THEN 'archive'" +
	" WHEN '" + pgsLab2OldShopDesc + "' THEN 'old_shop' END)"

// pgsLab2Setup drops every database but the service ones, so a setup run
// again over a used sandbox starts from the same state, then builds the lab's.
var pgsLab2Setup = pgSetup(
	"glpg q postgres " + b64("SELECT datname FROM pg_database WHERE NOT datistemplate AND datname <> 'postgres'") +
		" | while IFS= read -r db; do dropdb --force \"$db\"; done\n" +
		pgFresh("pereplet") +
		pgRun("postgres", `CREATE DATABASE archive OWNER postgres;
CREATE DATABASE old_shop OWNER postgres;
COMMENT ON DATABASE pereplet IS '`+pgsLab2PerepletDesc+`';
COMMENT ON DATABASE archive IS '`+pgsLab2ArchiveDesc+`';
COMMENT ON DATABASE old_shop IS '`+pgsLab2OldShopDesc+`';`) +
		pgRun("archive", `SET ROLE postgres;
CREATE TABLE old_orders (
    id         integer,
    customer   text,
    total      numeric(8,2),
    created_at date
);
INSERT INTO old_orders VALUES
    (1, 'Ирина Соколова', 1250.00, '2018-11-03'),
    (2, 'Павел Орлов',     480.50, '2019-02-17'),
    (3, 'Анна Белова',     990.00, '2019-06-25');`) +
		pgRun("old_shop", `SET ROLE postgres;
CREATE TABLE products (
    id    integer,
    title text,
    price numeric(8,2)
);
INSERT INTO products VALUES
    (1, 'Закладка для книг', 150.00),
    (2, 'Обложка для паспорта', 350.00);`))

// pgsLab2NewDB is the SQL for "database name exists, belongs to owner, and is
// not one of the setup's databases renamed".
func pgsLab2NewDB(name, owner string) string {
	return "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '" + name +
		"' AND pg_get_userbyid(datdba) = '" + owner + "' AND " + pgsLab2Origin + " IS NULL)"
}

// pgsLab2DBState describes database name in words: missing (with the list of
// the databases there are, so a typo shows), one of the setup's renamed, or
// owned by someone other than owner.
func pgsLab2DBState(name, owner string) string {
	return "SELECT coalesce((SELECT CASE " +
		"WHEN " + pgsLab2Origin + " IS NOT NULL THEN 'база " + name + " есть, но это не новая база, а переименованная ' || " + pgsLab2Origin + " " +
		"WHEN pg_get_userbyid(datdba) <> '" + owner + "' THEN 'база " + name + " есть, но её владелец — ' || pg_get_userbyid(datdba) || ', а нужен " + owner + "' " +
		"ELSE 'база " + name + " есть, владелец " + owner + "' END " +
		"FROM pg_database WHERE datname = '" + name + "'), " +
		"'базы " + name + " нет; есть базы: ' || (SELECT string_agg(datname, ', ' ORDER BY datname) FROM pg_database WHERE NOT datistemplate))"
}

// pgsLab2Dropped is the SQL for "old_shop is gone, under its own name or any
// other, and pereplet and archive, which the task says to leave alone, are
// still there".
const pgsLab2Dropped = "SELECT NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'old_shop' OR " + pgsLab2Origin + " = 'old_shop')" +
	" AND (SELECT count(*) FROM pg_database WHERE datname IN ('pereplet', 'archive')) = 2"

// pgsLab2OldShopState says whether old_shop is still there, under its own name
// or another, and whether a database the task says to leave alone went with it.
const pgsLab2OldShopState = "SELECT coalesce((SELECT CASE WHEN datname = 'old_shop' THEN 'база old_shop ещё есть' " +
	"ELSE 'базы old_shop нет, но она не удалена, а переименована в ' || datname END " +
	"FROM pg_database WHERE datname = 'old_shop' OR " + pgsLab2Origin + " = 'old_shop' " +
	"ORDER BY datname <> 'old_shop' LIMIT 1), " +
	"(SELECT CASE count(*) " +
	"WHEN 1 THEN 'база old_shop удалена, но вместе с ней пропала база ' || min(n) || ', а её трогать было нельзя — нажми «Пересоздать»' " +
	"WHEN 2 THEN 'база old_shop удалена, но вместе с ней пропали базы archive и pereplet, а их трогать было нельзя — нажми «Пересоздать»' END " +
	"FROM unnest(ARRAY['archive', 'pereplet']) AS n WHERE n NOT IN (SELECT datname FROM pg_database)), " +
	"'база old_shop удалена')"

// pgsLab2Now is a substitution printing a query's value for a diagnostic. When
// the query cannot run, it asks template1 (which cannot be dropped) why: the
// postgres database is gone, or the server is down.
func pgsLab2Now(sql string) string {
	return `$(glpg q postgres ` + b64(sql) + ` || glpg q template1 ` + b64(pgsLab2NoPostgres) +
		` || echo 'сервер PostgreSQL не отвечает — проверь pg_lsclusters')`
}

// pgsLab2NoPostgres explains a check that could not connect to postgres while
// the server answers.
const pgsLab2NoPostgres = "SELECT CASE WHEN EXISTS (SELECT 1 FROM pg_database WHERE datname = 'postgres') " +
	"THEN 'проверка не смогла подключиться к базе postgres' " +
	"ELSE 'нет служебной базы postgres, а проверка подключается к ней — создай её заново (CREATE DATABASE postgres;) или нажми «Пересоздать»' END"
