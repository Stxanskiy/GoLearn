#!/bin/bash
# glpg — the PostgreSQL course's lab helper.
#
# Every lab setup of the course writes this file to /usr/local/bin/glpg (see
# pgSetup in cmd/seed/labs_pg_common.go), and the checks call it. Living in the
# setup rather than in the sandbox image means a fix here reaches students with
# the next seed, without rebuilding the production root filesystem by hand.
#
# SQL always arrives base64-encoded: checks are stored as one line (the labcheck
# harness flattens newlines) and quoting SQL inside a quoted shell word is where
# such checks break.
#
#   glpg start                         start the cluster; root gets a superuser role
#   glpg fixhosts                      make the hostname resolve (quiet sudo)
#   glpg run   DB B64SQL               run SQL, stop at the first error (setups)
#   glpg q     DB B64SQL               print the result: tuples only, unaligned, '|'
#   glpg state DB B64SQL               run SQL in a rolled-back transaction and
#                                      print the SQLSTATE of its error, 00000 if none
#   glpg answer FILE DBS B64REF MODE   run the student's query file read-only on
#                                      each database of DBS (comma-separated) and
#                                      compare with the reference: MODE "set" or
#                                      "ordered"; prints what differs, exit 1
#   glpg outfile FILE DB B64REF        compare a file the student saved with the
#                                      reference query's -At output
set -u

decode() { printf '%s' "$1" | base64 -d; }

psql_() { psql -X -q -v ON_ERROR_STOP=1 "$@"; }

# rows prints a query's rows the way the comparison sees them: no header, no
# footer, '|' between columns, NULL shown as ∅ so it differs from ''.
rows() { # rows DB [psql source args...]
	local db=$1; shift
	PGOPTIONS='-c default_transaction_read_only=on -c statement_timeout=10s' \
		psql_ -A -t -F'|' -P null='∅' -d "$db" "$@"
}

normalize() { sed -e 's/[[:space:]]*$//' | sed -e :a -e '/^\n*$/{$d;N;ba' -e '}'; }

cols() { head -n1 | awk -F'|' 'NF { print NF; exit } END { if (!NR) print 0 }'; }

count() { if [ -z "$1" ]; then echo 0; else printf '%s\n' "$1" | wc -l; fi; }

cmd=${1:-}
shift || true

# fixhosts makes the sandbox's own hostname resolve. With --network none Docker
# leaves it out of /etc/hosts, and every `sudo -u postgres …` the course teaches
# then starts with "unable to resolve host" — harmless, but a beginner reads it
# as the command having failed.
fixhosts() {
	local h
	h=$(hostname 2>/dev/null) || return 0
	grep -qw -- "$h" /etc/hosts 2>/dev/null || echo "127.0.1.1 $h" >> /etc/hosts 2>/dev/null || true
}

case "$cmd" in
fixhosts)
	fixhosts
	;;

start)
	fixhosts
	VER=$(ls /etc/postgresql 2>/dev/null | head -1)
	[ -n "$VER" ] || { echo "PostgreSQL не установлен" >&2; exit 1; }
	pg_ctlcluster "$VER" main start >/dev/null 2>&1 || true
	for _ in $(seq 1 50); do
		su postgres -c 'psql -X -tAc "SELECT 1"' >/dev/null 2>&1 && break
		sleep 0.2
	done
	su postgres -c 'psql -X -tAc "SELECT 1"' >/dev/null 2>&1 || { echo "сервер PostgreSQL не запустился" >&2; exit 1; }
	su postgres -c "psql -X -tAc \"SELECT 1 FROM pg_roles WHERE rolname='root'\"" | grep -q 1 \
		|| su postgres -c "psql -X -qc \"CREATE ROLE root SUPERUSER LOGIN CREATEDB CREATEROLE\"" >/dev/null
	su postgres -c "psql -X -tAc \"SELECT 1 FROM pg_database WHERE datname='root'\"" | grep -q 1 \
		|| su postgres -c "createdb -O root root" >/dev/null
	;;

run)
	decode "$2" | psql_ -d "$1" -f -
	;;

q)
	psql_ -A -t -F'|' -d "$1" -c "$(decode "$2")" 2>/dev/null
	;;

state)
	# ON_ERROR_STOP ends the session at the first error, so a later statement in
	# the aborted transaction cannot overwrite its code with 25P02; closing the
	# connection rolls the transaction back either way.
	err=$({ echo 'BEGIN;'; decode "$2"; echo ';'; echo 'ROLLBACK;'; } \
		| psql -X -q -v ON_ERROR_STOP=1 -v VERBOSITY=sqlstate -d "$1" -f - 2>&1 >/dev/null)
	code=$(printf '%s\n' "$err" | grep -oE 'ERROR:  [0-9A-Z]{5}' | head -1 | awk '{print $2}')
	echo "${code:-00000}"
	;;

answer)
	file=$1 dbs=$2 ref=$(decode "$3") mode=${4:-set}
	if [ ! -s "$file" ]; then
		echo "файл $file не найден или пустой"
		exit 1
	fi
	if ! grep -qiE '^[[:space:]]*(select|with|table|values)|[^[:alnum:]_](select|with)[^[:alnum:]_]' "$file"; then
		echo "в $file нет запроса SELECT"
		exit 1
	fi
	for db in ${dbs//,/ }; do
		if ! got=$(rows "$db" -f "$file" 2>&1); then
			if printf '%s\n' "$got" | grep -q 'read-only transaction'; then
				echo "запрос пытается изменить данные, а здесь нужен только SELECT"
			else
				echo "запрос завершился ошибкой: $(printf '%s\n' "$got" | grep -m1 -E 'ERROR' | sed 's/^.*ERROR: *//')"
			fi
			exit 1
		fi
		want=$(rows "$db" -c "$ref" 2>&1) || { echo "эталон не выполнился в базе $db — сообщи авторам курса"; exit 1; }
		got=$(printf '%s\n' "$got" | normalize)
		want=$(printf '%s\n' "$want" | normalize)
		if [ "$mode" != ordered ]; then
			got=$(printf '%s\n' "$got" | LC_ALL=C sort)
			want=$(printf '%s\n' "$want" | LC_ALL=C sort)
		fi
		[ "$got" = "$want" ] && continue
		gc=$(printf '%s\n' "$got" | cols) wc=$(printf '%s\n' "$want" | cols)
		gn=$(count "$got") wn=$(count "$want")
		if [ "$gc" != "$wc" ] && [ "$gn" != 0 ] && [ "$wn" != 0 ]; then
			echo "столбцов в результате $gc, а нужно $wc"
		elif [ "$gn" != "$wn" ]; then
			echo "запрос вернул строк: $gn, а ожидается $wn"
		elif [ "$mode" = ordered ] && [ "$(printf '%s\n' "$got" | LC_ALL=C sort)" = "$(printf '%s\n' "$want" | LC_ALL=C sort)" ]; then
			echo "строки те, но порядок не тот — проверь ORDER BY"
		else
			echo "строк столько, сколько нужно ($gn), но значения отличаются"
		fi
		if [ "$db" != "${dbs%%,*}" ]; then
			echo "(на основной базе ответ совпал, а на проверочной копии с другими данными — нет: запрос не должен зависеть от конкретных строк)"
		fi
		exit 1
	done
	;;

outfile)
	file=$1 db=$2 ref=$(decode "$3")
	if [ ! -s "$file" ]; then
		echo "файл $file не найден или пустой"
		exit 1
	fi
	want=$(psql_ -A -t -d "$db" -c "$ref" 2>&1 | normalize) || { echo "эталон не выполнился — сообщи авторам курса"; exit 1; }
	got=$(normalize < "$file")
	[ "$got" = "$want" ] && exit 0
	gn=$(count "$got") wn=$(count "$want")
	if [ "$gn" != "$wn" ]; then
		echo "в файле строк: $gn, а ожидается $wn"
	elif printf '%s\n' "$got" | grep -qE '^[-+]+$|^\(.* (строк|строки|строка|rows?)\)$' \
		|| { printf '%s\n' "$got" | grep -q '|' && ! printf '%s\n' "$want" | grep -q '|'; }; then
		echo "в файле остались заголовок, рамка или разделители — нужны только значения"
	else
		echo "строк столько, сколько нужно ($gn), но содержимое отличается"
	fi
	exit 1
	;;

*)
	echo "glpg: неизвестная команда '$cmd'" >&2
	exit 2
	;;
esac
