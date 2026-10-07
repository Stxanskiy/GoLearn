#!/usr/bin/env bash
# Checks one lab without a seeded database.
#
# Usage: scripts/labcheck/lab.sh <module-slug> <lesson-slug>
#
# run.sh reads the labs from a seeded database; this reads them straight from
# the seeder's code (TestDumpLab), so a lab can be verified while it is being
# written. The rules are run.sh's: one container per lesson, the setup once,
# then per task the check must FAIL before the reference solution and PASS
# after it. Checks are flattened to one line as run.sh does. Setup time is
# printed because the runner gives the first command 20 seconds in total.
set -uo pipefail
MODULE="$1" LESSON="$2"
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
for f in "$HERE"/solutions*.sh; do source "$f"; done

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
if ! (cd "$ROOT" && GL_DUMP_LAB="$MODULE/$LESSON" GL_DUMP_OUT="$tmp/lab.json" \
	go test ./cmd/seed -run '^TestDumpLab$' -count=1 >"$tmp/go.log" 2>&1); then
	cat "$tmp/go.log"; exit 1
fi
img=$(jq -r '.image // empty' "$tmp/lab.json"); [ -z "$img" ] && img=golearn/sandbox:latest
c="labcheck-$LESSON-$$"
opts=""
case "$img" in
	*sandbox-docker*) opts="--privileged --memory 2g --cpus 2 --pids-limit 2048" ;;
	*sandbox-k8s*) opts="--privileged --memory 4g --cpus 3 --pids-limit 4096 --tmpfs /run" ;;
	*) opts="--memory 512m --cpus 1 --pids-limit 256" ;;
esac
docker run -d --init --name "$c" --hostname sandbox --network none $opts "$img" sleep infinity >/dev/null || { echo "container start failed"; exit 1; }
trap 'docker rm -f "$c" >/dev/null 2>&1; rm -rf "$tmp"' EXIT

start=$(date +%s.%N)
if ! out=$(jq -r '.setup' "$tmp/lab.json" | docker exec -i "$c" bash 2>&1); then
	echo "SETUP FAILED:"; echo "$out" | tail -20; exit 1
fi
printf 'setup: %.1fs (%s)\n' "$(echo "$(date +%s.%N) - $start" | bc)" "$img"

pass=0 failed=0 missing=0
n=$(jq '.checks | length' "$tmp/lab.json")
for ((i = 0; i < n; i++)); do
	idx=$((i + 1))
	chk=$(jq -r ".checks[$i]" "$tmp/lab.json" | tr '\n' ' ')
	title=$(jq -r ".titles[$i]" "$tmp/lab.json")
	key="sol_$(echo "$LESSON" | tr '-' '_')_$idx"
	sol="${!key:-}"
	if [ -z "$sol" ]; then echo "SKIP  #$idx $title (нет $key)"; missing=$((missing + 1)); continue; fi
	wrapped='cd "$(cat /root/.gl_cwd 2>/dev/null || echo /root)" 2>/dev/null; '"$chk"
	if bout=$(docker exec "$c" bash -c "$wrapped" 2>&1); then before=pass; else before=fail; fi
	sout=$(docker exec "$c" bash -c "$sol" 2>&1); src=$?
	if aout=$(docker exec "$c" bash -c "$wrapped" 2>&1); then after=pass; else after=fail; fi
	if [ "$before" = fail ] && [ "$after" = pass ]; then
		echo "OK    #$idx $title"
		echo "      до:    $(echo "$bout" | tail -1)"
		pass=$((pass + 1))
	elif [ "$after" = pass ]; then
		echo "WEAK  #$idx $title — проверка проходит ещё ДО решения"; failed=$((failed + 1))
	else
		echo "FAIL  #$idx $title — эталонное решение не проходит проверку"
		[ "$src" != 0 ] && echo "      решение упало ($src): $(echo "$sout" | tail -3)"
		echo "      после: $(echo "$aout" | tail -2)"
		failed=$((failed + 1))
	fi
done
echo "──────────────────────────────"
echo "OK: $pass   проблемных: $failed   без решения: $missing"
[ "$failed" = 0 ] && [ "$missing" = 0 ]
