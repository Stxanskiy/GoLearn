package main

// Lab 1 of pg-start: the student installs PostgreSQL on a clean server.
//
// This is the one lab of the course without pgSetup — there is no server at
// the start. The setup removes PostgreSQL from the sandbox-pg image; its
// packages stay in the image's offline repository (/var/cache/glpkg), so
// `apt update` and `apt install postgresql` work with no network, as on a real
// host. The postgres OS user goes too, and so does the image's own pg-start
// helper: either would give away a step the student is meant to take.
//
// Until task 5 root has no role, so glpg q (which connects as root) cannot see
// anything. The "Сейчас:" diagnostics therefore ask the server as the postgres
// OS user (pgsLab1AsPostgres); the conditions themselves stay with the shared
// helpers.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab1.sh.

func init() {
	pgStartLabs["ch-pgs-lab1"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgsLab1Setup,
		Checks: map[int]string{
			1: check(pgsLab1Installed,
				"пакет postgresql-16 установлен",
				"Сейчас: пакет postgresql-16 — $(dpkg-query -W -f='${Status}' postgresql-16 2>/dev/null | grep -v not-installed | grep . || echo 'не установлен')"),
			// The version alone is not enough: "16.15" typed in from the apt log
			// would pass without running anything. The file has to be the output
			// of a version command — psql/postgres --version and version() print
			// the word PostgreSQL, SHOW server_version the package version.
			2: check(pgsLab1Ver+` && grep -oE '[0-9]+\.[0-9]+' /root/pg_version.txt 2>/dev/null | grep -qxF "$v" && { grep -q PostgreSQL /root/pg_version.txt || grep -qF "$f" /root/pg_version.txt; }`,
				"в /root/pg_version.txt сохранён вывод psql --version (или SELECT version();) с версией установленного PostgreSQL",
				"Сейчас в /root/pg_version.txt: $(grep -m1 '[^[:space:]]' /root/pg_version.txt 2>/dev/null || echo 'файла нет или он пустой')"),
			3: check(`[ "$(pg_lsclusters -h 16 main 2>/dev/null | awk '{print $4}')" = online ] && grep -qE '^[[:space:]]*16[[:space:]]+main[[:space:]]+5432[[:space:]]+online([[:space:]]|$)' /root/cluster.txt 2>/dev/null`,
				"кластер 16 main запущен, а в /root/cluster.txt сохранён вывод pg_lsclusters, где он online",
				"Сейчас: статус кластера 16 main — $(pg_lsclusters -h 16 main 2>/dev/null | awk '{print $4}' | grep . || echo 'кластера нет'), в /root/cluster.txt — $(grep -E '^[[:space:]]*16[[:space:]]+main[[:space:]]' /root/cluster.txt 2>/dev/null | awk '{print $4}' | head -n1 | grep . || echo 'строки кластера 16 main нет')"),
			4: check(pgsLab1Ver+` && grep -qw postgres /root/whoami.txt 2>/dev/null && grep -qF "PostgreSQL $v " /root/whoami.txt`,
				"в /root/whoami.txt сохранён ответ сервера на SELECT current_user, version(); от имени postgres",
				"Сейчас в /root/whoami.txt: имя postgres — $(grep -qw postgres /root/whoami.txt 2>/dev/null && echo есть || echo нет), строка из version() — $(grep -qE 'PostgreSQL [0-9]+[.][0-9]+ ' /root/whoami.txt 2>/dev/null && echo есть || echo нет)"),
			5: check(`[ "$(psql -XtAc 'SELECT current_user' 2>/dev/null)" = root ] && `+
				pgTrue("postgres", "SELECT rolcanlogin AND rolcreatedb AND NOT rolsuper FROM pg_roles WHERE rolname = 'root'")+` && `+
				pgIs("postgres", "SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = 'root'", "root"),
				"psql от root работает без sudo: есть роль root (LOGIN, CREATEDB, не суперпользователь) и база root с владельцем root",
				"Сейчас psql от root отвечает: $(psql -XtAc 'SELECT current_user' 2>&1 | head -n1 | sed -e 's/^bash: line [0-9]*: //' -e 's/^psql: error: //' -e 's/^.*FATAL: *//'); роль root — "+
					pgsLab1AsPostgres(pgsLab1RoleSQL)+"; база root — "+
					pgsLab1AsPostgres("SELECT coalesce((SELECT 'владелец ' || pg_get_userbyid(datdba) FROM pg_database WHERE datname = 'root'), 'нет')")),
			6: check(pgIs("postgres", "SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = 'pereplet'", "root"),
				"база pereplet создана, её владелец — root",
				"Сейчас: "+pgsLab1AsPostgres("SELECT coalesce((SELECT 'база pereplet есть, владелец ' || pg_get_userbyid(datdba) FROM pg_database WHERE datname = 'pereplet'), 'базы pereplet нет')")),
			// The whole line, timestamp and pid included, must be one the server
			// wrote: the phrase alone is in the task text. Whitespace around it is
			// what a copy from the terminal adds, so it is trimmed.
			7: check(`grep -F 'ready to accept connections' /root/ready.txt 2>/dev/null | sed 's/^[[:space:]]*//; s/[[:space:]]*$//' | grep -qxFf - `+pgsLab1Log,
				"в /root/ready.txt — строка журнала сервера о готовности принимать подключения, целиком, как она записана в журнале",
				"Сейчас в /root/ready.txt: $(grep -m1 '[^[:space:]]' /root/ready.txt 2>/dev/null || echo 'файла нет или он пустой')"),
		},
	}
}

// pgsLab1Setup takes the server away and leaves the packages: purge, then the
// directories a purge keeps (cluster data and logs survive it on purpose).
// policy-rc.d is already in the container image and keeps the freshly
// installed cluster down, which the theory explains; the line only makes sure a
// root filesystem without it behaves the same.
var pgsLab1Setup = "set -e\n" + pgInstallGlpg + "glpg fixhosts\n" + `apt-get purge -y -qq postgresql postgresql-16 postgresql-client-16 postgresql-client postgresql-common postgresql-client-common >/dev/null
rm -rf /etc/postgresql /var/lib/postgresql /var/log/postgresql /var/run/postgresql
userdel postgres 2>/dev/null || true
rm -f /usr/local/bin/pg-start
[ -x /usr/sbin/policy-rc.d ] || { printf '#!/bin/sh\nexit 101\n' > /usr/sbin/policy-rc.d; chmod +x /usr/sbin/policy-rc.d; }
`

// pgsLab1Installed is a condition: the server package is installed.
const pgsLab1Installed = `[ "$(dpkg-query -W -f='${Status}' postgresql-16 2>/dev/null)" = 'install ok installed' ]`

// pgsLab1Ver sets f to the installed package version (16.15-0ubuntu0.24.04.1)
// and v to its major.minor (16.15), read from the package rather than written
// down, so the checks survive a point release in the image.
const pgsLab1Ver = `f=$(dpkg-query -W -f='${Version}' postgresql-16 2>/dev/null) && f=${f#*:} && v=${f%%-*} && [ -n "$v" ]`

// pgsLab1Log is the cluster's log file, as pg_lsclusters shows it.
const pgsLab1Log = "/var/log/postgresql/postgresql-16-main.log"

// pgsLab1RoleSQL describes root's role in words, or says there is none.
const pgsLab1RoleSQL = "SELECT coalesce((SELECT concat_ws(', ', " +
	"CASE WHEN rolcanlogin THEN 'вход разрешён' ELSE 'вход запрещён' END, " +
	"CASE WHEN rolcreatedb THEN 'может создавать базы' ELSE 'не может создавать базы' END, " +
	"CASE WHEN rolsuper THEN 'суперпользователь' ELSE 'не суперпользователь' END) " +
	"FROM pg_roles WHERE rolname = 'root'), 'нет')"

// pgsLab1AsPostgres is a "Сейчас:" substitution printing a query's value as
// the postgres OS user, who gets in by peer authentication whether or not root
// has a role yet.
func pgsLab1AsPostgres(sql string) string {
	return `$(runuser -u postgres -- /usr/local/bin/glpg q postgres ` + b64(sql) + ` 2>/dev/null || echo 'сервер не отвечает')`
}
