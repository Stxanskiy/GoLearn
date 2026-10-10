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
// The lesson has the student connect once with sudo -u postgres and then create
// a superuser role for root, after which a plain `psql pereplet` works; the
// later labs' pgSetup leaves root the same role. Until that role exists root
// cannot ask the server anything, so the checks ask as the postgres OS user
// (pgsLab1AsPostgres).
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
			// LOGIN as well as SUPERUSER: CREATE ROLE without LOGIN makes a
			// superuser psql still cannot connect as.
			2: check(`[ "`+pgsLab1AsPostgres("SELECT rolsuper AND rolcanlogin FROM pg_roles WHERE rolname = 'root'")+`" = t ]`,
				"в PostgreSQL есть пользователь root с правами администратора",
				"Сейчас пользователь root — "+pgsLab1AsPostgres(pgsLab1RoleSQL)),
			// The owner shows who created the database: root, connected with a
			// plain psql. One made with sudo -u postgres belongs to postgres.
			3: check(`[ "`+pgsLab1AsPostgres("SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname = 'pereplet'")+`" = root ]`,
				"база pereplet создана пользователем root (через psql, без sudo)",
				"Сейчас: "+pgsLab1AsPostgres("SELECT coalesce((SELECT 'база pereplet есть, но её создал пользователь ' || pg_get_userbyid(datdba) FROM pg_database WHERE datname = 'pereplet'), 'базы pereplet нет')")),
		},
	}
}

// pgsLab1Setup takes the server away and leaves the packages: purge, then the
// directories a purge keeps (cluster data and logs survive it on purpose).
//
// It purges what dpkg says is installed rather than a fixed list of names.
// A named purge is fatal on an image that has no PostgreSQL — apt answers
// "Unable to locate package" and exits 100, `set -e` ends the setup, and the
// boot reports the whole sandbox as failed. Cleanup must never be able to do
// that: there is nothing to remove, which is a reason to carry on, not to
// stop.
//
// This does not make the course work on such an image. Removing a server that
// is not there is fine; installing it again needs the packages, and those come
// from the golden rootfs (deploy/fc-rootfs/Dockerfile.rootfs). It only means a
// stale image costs the student a failing task instead of no sandbox at all.
var pgsLab1Setup = "set -e\n" + pgInstallGlpg + "glpg fixhosts\n" + `installed=$(dpkg-query -W -f='${Package} ${Status}\n' 'postgresql*' 2>/dev/null | awk '$3 == "ok" && $4 == "installed" { print $1 }')
[ -n "$installed" ] && apt-get purge -y -qq $installed >/dev/null || true
rm -rf /etc/postgresql /var/lib/postgresql /var/log/postgresql /var/run/postgresql
userdel postgres 2>/dev/null || true
rm -f /usr/local/bin/pg-start
printf '#!/bin/sh\nexit 0\n' > /usr/sbin/policy-rc.d
chmod +x /usr/sbin/policy-rc.d
`

// pgsLab1Installed is a condition: the server package is installed.
const pgsLab1Installed = `[ "$(dpkg-query -W -f='${Status}' postgresql-16 2>/dev/null)" = 'install ok installed' ]`

// pgsLab1RoleSQL describes root's role in words, or says there is none.
const pgsLab1RoleSQL = "SELECT coalesce((SELECT CASE " +
	"WHEN NOT rolcanlogin THEN 'есть, но не может подключаться (нет LOGIN)' " +
	"WHEN NOT rolsuper THEN 'есть, но не администратор (нет --superuser)' " +
	"ELSE 'есть' END FROM pg_roles WHERE rolname = 'root'), 'не создан')"

// pgsLab1AsPostgres is a substitution printing a query's value as the postgres
// OS user, the only one the server knows in this lab.
func pgsLab1AsPostgres(sql string) string {
	return `$(runuser -u postgres -- /usr/local/bin/glpg q postgres ` + b64(sql) + ` 2>/dev/null || echo 'нет ответа от сервера: PostgreSQL установлен и запущен?')`
}
