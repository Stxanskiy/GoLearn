package main

// Lab 1 of pg-start: installing PostgreSQL (ch-pgs-lab1).
//
// The one lab of the course without pgSetup — there is no server at the start.
// The setup removes PostgreSQL from the sandbox-pg image; its packages stay in
// the image's offline repository (/var/cache/glpkg), so `apt update` and
// `apt install postgresql` work with no network, as on a real host. The
// postgres OS user goes too, and so does the image's own pg-start helper: it
// would hand out a role the lesson does not mention.
//
// policy-rc.d is set to allow service starts. The container image ships one
// that forbids them, which left the new cluster down and made the very first
// lesson about pg_ctlcluster and clusters; with starts allowed, apt brings the
// server up the way it does on an ordinary Ubuntu server.
//
// root has no role in this lab (the lesson connects with sudo -u postgres), so
// the checks ask the server as the postgres OS user (pgsLab1AsPostgres).
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.

func init() {
	pgStartLabs["ch-pgs-lab1"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgsLab1Setup,
		Checks: map[int]string{
			1: check(pgsLab1Installed,
				"PostgreSQL установлен",
				"Сейчас: пакет postgresql-16 — $(dpkg-query -W -f='${Status}' postgresql-16 2>/dev/null | grep -v not-installed | grep . || echo 'не установлен')"),
			// The file alone could be typed by hand; the server has to be up too.
			2: check(`[ "$(pg_lsclusters -h 16 main 2>/dev/null | awk '{print $4}')" = online ] && grep -qE '^[[:space:]]*16[[:space:]]+main[[:space:]]+5432[[:space:]]+online([[:space:]]|$)' /root/status.txt 2>/dev/null`,
				"сервер работает, а в /root/status.txt сохранён вывод pg_lsclusters со статусом online",
				"Сейчас: сервер — $(pg_lsclusters -h 16 main 2>/dev/null | awk '{print $4}' | grep . || echo 'не найден'), в /root/status.txt — $(grep -qE '[[:space:]]online([[:space:]]|$)' /root/status.txt 2>/dev/null && echo 'есть строка со статусом online' || { [ -s /root/status.txt ] && echo 'нет строки со статусом online' || echo 'файла нет или он пустой'; })"),
			// version() prints "PostgreSQL 16.15 (Ubuntu …)". psql --version
			// prints "psql (PostgreSQL) 16.15", so it does not pass: the task is
			// to ask the server, not the client.
			3: check(pgsLab1Ver+` && grep -qF "PostgreSQL $v " /root/version.txt 2>/dev/null`,
				"в /root/version.txt сохранён ответ сервера на SELECT version();",
				"Сейчас в /root/version.txt: $( (grep -m1 '[^[:space:]]' /root/version.txt 2>/dev/null || echo 'файла нет или он пустой') | cut -c1-80)"),
			4: check(`[ "`+pgsLab1AsPostgres("SELECT count(*) FROM pg_database WHERE datname = 'pereplet'")+`" = 1 ]`,
				"база pereplet создана",
				"Сейчас баз на сервере: "+pgsLab1AsPostgres("SELECT string_agg(datname, ', ' ORDER BY datname) FROM pg_database")),
		},
	}
}

// pgsLab1Setup takes the server away and leaves the packages: purge, then the
// directories a purge keeps (cluster data and logs survive it on purpose).
var pgsLab1Setup = "set -e\n" + pgInstallGlpg + "glpg fixhosts\n" + `apt-get purge -y -qq postgresql postgresql-16 postgresql-client-16 postgresql-client postgresql-common postgresql-client-common >/dev/null
rm -rf /etc/postgresql /var/lib/postgresql /var/log/postgresql /var/run/postgresql
userdel postgres 2>/dev/null || true
rm -f /usr/local/bin/pg-start
printf '#!/bin/sh\nexit 0\n' > /usr/sbin/policy-rc.d
chmod +x /usr/sbin/policy-rc.d
`

// pgsLab1Installed is a condition: the server package is installed.
const pgsLab1Installed = `[ "$(dpkg-query -W -f='${Status}' postgresql-16 2>/dev/null)" = 'install ok installed' ]`

// pgsLab1Ver sets v to the installed server's major.minor (16.15), read from the
// package rather than written down, so the check survives a point release.
const pgsLab1Ver = `f=$(dpkg-query -W -f='${Version}' postgresql-16 2>/dev/null) && f=${f#*:} && v=${f%%-*} && [ -n "$v" ]`

// pgsLab1AsPostgres is a substitution printing a query's value as the postgres
// OS user, the only one the server knows in this lab.
func pgsLab1AsPostgres(sql string) string {
	return `$(runuser -u postgres -- /usr/local/bin/glpg q postgres ` + b64(sql) + ` 2>/dev/null || echo 'сервер не отвечает')`
}
