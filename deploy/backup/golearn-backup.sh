#!/usr/bin/env bash
# Back up the GoLearn database from the k3s cluster, verify the dump, and rotate.
#
# Everything a student has done lives in this one database: accounts, progress,
# quiz attempts, submissions, billing — and, because object storage is not
# configured in production, the uploaded images too. There is nothing else to
# back up, and no second copy of any of it.
#
# Run on the FC host (berg) as a user with sudo, via golearn-backup.timer.
#
#   golearn-backup.sh            take a backup, verify it, rotate
#   golearn-backup.sh --verify   restore the newest backup into a scratch
#                                database and compare row counts (does not
#                                touch the live one)
#
# It exits non-zero and says why on any failure. A backup job that fails quietly
# is worse than no backup job: it buys confidence it has not earned.
set -euo pipefail

DIR="${BACKUP_DIR:-/var/backups/golearn}"
NS="${K8S_NAMESPACE:-golearn}"
DEPLOY="${DB_DEPLOY:-golearn-db}"
DB="${DB_NAME:-golearn}"
USER="${DB_USER:-golearn}"
KEEP_DAILY="${KEEP_DAILY:-14}"
KEEP_WEEKLY="${KEEP_WEEKLY:-8}"
KEEP_MONTHLY="${KEEP_MONTHLY:-12}"
# Below this the dump is certainly truncated: the schema alone is larger.
MIN_BYTES="${MIN_BYTES:-100000}"

kube() { sudo k3s kubectl -n "$NS" "$@"; }
die()  { echo "backup: $*" >&2; exit 1; }

# ── verify mode ─────────────────────────────────────────────────────────────
# Restores the newest dump into a throwaway PostgreSQL container and compares
# row counts against production.
#
# It deliberately does NOT restore into the production instance. The first
# version did, and a half-finished restore left a backend stuck in "startup
# waiting" that blocked DROP DATABASE and queued every later attempt behind it.
# Verifying a backup must not be able to disturb the thing it is protecting.
if [ "${1:-}" = "--verify" ]; then
  command -v docker >/dev/null || die "нужен docker: дамп разворачивается в одноразовом контейнере"
  newest=$(ls -1t "$DIR"/golearn-*.sql.gz 2>/dev/null | head -1) || true
  [ -n "${newest:-}" ] || die "в $DIR нет ни одного дампа"
  echo "проверяю $newest"
  gzip -t "$newest" || die "архив повреждён"

  cname="golearn-verify-$$"
  cleanup() { docker rm -f "$cname" >/dev/null 2>&1 || true; }
  trap cleanup EXIT

  docker run -d --rm --name "$cname" -e POSTGRES_PASSWORD=verify \
    -e POSTGRES_USER="$USER" -e POSTGRES_DB="$DB" postgres:16-alpine >/dev/null \
    || die "не удалось поднять контейнер для проверки"

  # pg_isready is not enough: the image starts a temporary server to run its
  # init scripts and then restarts it, so a readiness check can pass against a
  # socket that is about to disappear. Wait for a real query over TCP instead.
  pg() { docker exec -e PGPASSWORD=verify "$cname" psql -h 127.0.0.1 -U "$USER" -d "$DB" "$@"; }
  for i in $(seq 1 90); do
    pg -tAc "SELECT 1" >/dev/null 2>&1 && break
    [ "$i" = 90 ] && die "контейнер с postgres не принял подключение"
    sleep 1
  done

  gunzip -c "$newest" \
    | docker exec -i -e PGPASSWORD=verify "$cname" psql -h 127.0.0.1 -U "$USER" -d "$DB" -q >/dev/null 2>&1 \
    || die "дамп не восстанавливается"

  # A restore that "succeeds" into an empty database is the failure mode worth
  # catching, so compare what landed against what production actually holds.
  bad=0
  for t in users lessons tasks progress; do
    # timeout cannot run a shell function, so this spells the command out.
    live=$(timeout 60 sudo k3s kubectl -n "$NS" exec "deploy/$DEPLOY" -- psql -U "$USER" -d "$DB" -tAc "SELECT count(*) FROM $t" 2>/dev/null | tr -d '\r')
    back=$(pg -tAc "SELECT count(*) FROM $t" 2>/dev/null | tr -d '\r')
    printf '  %-10s прод: %-8s из бекапа: %s\n' "$t" "${live:-?}" "${back:-?}"
    [ -n "$live" ] && [ "$live" = "$back" ] || bad=1
  done
  [ "$bad" = 0 ] || die "строки не сходятся — дамп неполный"
  echo "дамп восстанавливается, строки сходятся"
  exit 0
fi

# ── take a backup ───────────────────────────────────────────────────────────
# The directory is created at install time and owned by the service user; doing
# it here would need root, and a backup job with root is a bigger thing than it
# needs to be.
[ -d "$DIR" ] || die "нет каталога $DIR — создай его: sudo install -d -o \"$(id -un)\" -m 750 $DIR"
[ -w "$DIR" ] || die "каталог $DIR недоступен на запись пользователю $(id -un)"
avail=$(df -Pk "$DIR" | awk 'NR==2{print $4}')
[ "$avail" -gt 1048576 ] || die "на разделе с $DIR меньше 1 ГБ свободно"

stamp=$(date +%F-%H%M)
out="$DIR/golearn-$stamp.sql.gz"
tmp="$out.part"

# --clean --if-exists so the dump can be replayed over an existing database;
# without it a restore onto a live pod fails halfway and leaves it half-written.
timeout 600 sudo k3s kubectl -n "$NS" exec "deploy/$DEPLOY" -- pg_dump -U "$USER" -d "$DB" --clean --if-exists \
  | gzip -9 > "$tmp" || { rm -f "$tmp"; die "pg_dump не отработал"; }

size=$(stat -c%s "$tmp" 2>/dev/null || stat -f%z "$tmp")
[ "$size" -ge "$MIN_BYTES" ] || { rm -f "$tmp"; die "дамп подозрительно мал ($size байт)"; }
gzip -t "$tmp" || { rm -f "$tmp"; die "получился повреждённый архив"; }
# grep -c, not grep -q: -q exits at the first match, gunzip takes SIGPIPE, and
# pipefail turns a found table into a failed pipeline.
[ "$(gunzip -c "$tmp" | grep -c "^CREATE TABLE public\.lessons ")" -ge 1 ] \
  || { rm -f "$tmp"; die "в дампе нет таблицы lessons — это не наша база"; }

mv "$tmp" "$out"
echo "готово: $out ($(numfmt --to=iec "$size" 2>/dev/null || echo "$size байт"))"

# ── rotation ────────────────────────────────────────────────────────────────
# Daily for two weeks, then one per week, then one per month. Keeping a year of
# monthlies costs ~50 MB at the current database size, so the limit here is
# taste, not disk.
keep=$(mktemp); trap 'rm -f "$keep"' EXIT
ls -1 "$DIR"/golearn-*.sql.gz 2>/dev/null | sort -r | head -n "$KEEP_DAILY" >> "$keep"

pick() { # pick <format> <count>: newest dump for each distinct period
  local fmt="$1" count="$2" seen=""
  for f in $(ls -1 "$DIR"/golearn-*.sql.gz 2>/dev/null | sort -r); do
    local day period
    day=$(basename "$f" | sed -E 's/^golearn-([0-9]{4}-[0-9]{2}-[0-9]{2}).*/\1/')
    period=$(date -d "$day" +"$fmt" 2>/dev/null) || continue
    case " $seen " in *" $period "*) continue ;; esac
    seen="$seen $period"
    echo "$f" >> "$keep"
    [ "$(echo "$seen" | wc -w)" -ge "$count" ] && break
  done
}
pick "%G-%V" "$KEEP_WEEKLY"    # ISO week
pick "%Y-%m" "$KEEP_MONTHLY"

sort -u "$keep" -o "$keep"
removed=0
for f in $(ls -1 "$DIR"/golearn-*.sql.gz 2>/dev/null); do
  grep -qxF "$f" "$keep" || { rm -f "$f"; removed=$((removed + 1)); }
done
echo "хранится: $(wc -l < "$keep") дампов, удалено при ротации: $removed"
