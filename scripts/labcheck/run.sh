#!/usr/bin/env bash
# Regression harness for the labs' auto-checks.
#
# Usage: scripts/labcheck/run.sh <module-slug>      (default: linux-start)
#
# Verifies every authored check: for each lesson, spin one sandbox container,
# apply the lesson setup, then per task assert check FAILS before the reference
# solution and PASSES after it.
set -uo pipefail
MODULE="${1:-linux-start}"
HERE="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"
source "$HERE/solutions.sh"
source "$HERE/solutions-docker.sh"
source "$HERE/solutions-compose.sh"
source "$HERE/solutions-k8s.sh"
PSQL=(docker compose -f "$ROOT/docker-compose.yml" exec -T db psql -U golearn -d golearn -tAF'|')

pass=0; failed=0; missing=0
lessons=$("${PSQL[@]}" -c "SELECT DISTINCT l.slug, l.order_num FROM lessons l JOIN modules m ON m.id=l.module_id JOIN tasks t ON t.lesson_id=l.id WHERE m.slug='$MODULE' AND t.check_script<>'' ORDER BY l.order_num;" | cut -d'|' -f1)

for lesson in $lessons; do
  c="verify-$lesson"
  docker rm -f "$c" >/dev/null 2>&1
  img=$("${PSQL[@]}" -c "SELECT t.sandbox_image FROM tasks t JOIN lessons l ON l.id=t.lesson_id JOIN modules m ON m.id=l.module_id WHERE m.slug='$MODULE' AND l.slug='$lesson' AND t.sandbox_image<>'' LIMIT 1;")
  [ -z "$img" ] && img=golearn/sandbox:latest
  opts=""
  case "$img" in
    *sandbox-docker*)
      docker volume rm -f "$c-dind" >/dev/null 2>&1
      opts="--privileged --memory 2g --cpus 2 --pids-limit 2048 -v $c-dind:/var/lib/docker" ;;
    *sandbox-k8s*)
      docker volume rm -f "$c-dind" >/dev/null 2>&1
      opts="--privileged --memory 4g --cpus 3 --pids-limit 4096 --tmpfs /run -v $c-dind:/var/lib/rancher" ;;
  esac
  docker run -d --init --name "$c" --network none $opts "$img" sleep infinity >/dev/null || { echo "container start failed"; exit 1; }

  # Same rule as repository.LessonSandbox: every distinct task setup of the
  # lesson runs once, in task order, in the single shared container.
  setup=$("${PSQL[@]}" -c "SELECT string_agg(s, E'\n' ORDER BY o) FROM (SELECT t.setup_script AS s, min(t.order_num) AS o FROM tasks t JOIN lessons l ON l.id=t.lesson_id JOIN modules m ON m.id=l.module_id WHERE m.slug='$MODULE' AND l.slug='$lesson' AND t.setup_script<>'' GROUP BY t.setup_script) x;")
  [ -n "$setup" ] && docker exec "$c" bash -c "$setup" >/dev/null 2>&1

  n=0
  while IFS='|' read -r idx chk; do
    [ -z "$idx" ] && continue
    n=$((n+1))
    key="sol_$(echo "$lesson" | tr '-' '_')_$idx"
    sol="${!key:-}"
    if [ -z "$sol" ]; then echo "SKIP  $lesson#$idx (нет эталонного решения)"; missing=$((missing+1)); continue; fi

    wrapped='cd "$(cat /root/.gl_cwd 2>/dev/null || echo /root)" 2>/dev/null; '"$chk"
    docker exec "$c" bash -c "$wrapped" >/dev/null 2>&1 && before=pass || before=fail
    docker exec "$c" bash -c "$sol" >/dev/null 2>&1
    if out=$(docker exec "$c" bash -c "$wrapped" 2>&1); then after=pass; else after=fail; fi

    if [ "$after" = pass ] && [ "$before" = fail ]; then
      pass=$((pass+1))
    elif [ "$after" = pass ]; then
      echo "WEAK  $lesson#$idx — проверка проходит ещё ДО решения"; failed=$((failed+1))
    else
      echo "FAIL  $lesson#$idx — эталонное решение не проходит проверку"
      echo "      $(echo "$out" | tail -1)"
      failed=$((failed+1))
    fi
  done < <("${PSQL[@]}" -c "SELECT row_number() OVER (ORDER BY t.order_num, t.id), replace(t.check_script, E'\n', ' ') FROM tasks t JOIN lessons l ON l.id=t.lesson_id JOIN modules m ON m.id=l.module_id WHERE m.slug='$MODULE' AND l.slug='$lesson' ORDER BY t.order_num, t.id;")

  docker rm -f "$c" >/dev/null 2>&1
  docker volume rm -f "$c-dind" >/dev/null 2>&1
done
echo "──────────────────────────────"
echo "OK: $pass   проблемных: $failed   без решения: $missing"
