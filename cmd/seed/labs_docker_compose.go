package main

// Fixtures + auto-checks for "Docker Compose: многосервисные приложения".
//
// Each lesson gets a ready project directory (compose file plus the app files
// it references) in the Docker-enabled sandbox, so `docker compose up` has
// something real to start. All images used here are baked into the image.

// composeApp is a dependency-free HTTP service used as the "app" tier.
const composeApp = `import os
from http.server import BaseHTTPRequestHandler, HTTPServer

NAME = os.environ.get("APP_NAME", "app")
ENVIRONMENT = os.environ.get("APP_ENV", "production")

class H(BaseHTTPRequestHandler):
    def do_GET(self):
        body = f"Hello from {NAME} ({ENVIRONMENT})\n".encode()
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *a):
        pass

HTTPServer(("0.0.0.0", 5000), H).serve_forever()
`

// composeDown stops any leftover stack from a previous attempt.
const composeDown = `for d in app webapp dbapp envapp override healthapp scaleapp; do
  [ -d "/root/$d" ] && (cd "/root/$d" && docker compose down -v --remove-orphans >/dev/null 2>&1)
done
true`

var dockerComposeLabs = map[string]labSpec{
	// ── Lab 1: первый compose-проект ──
	"ch-dkc-lab1": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
docker volume rm -f app_cache_data >/dev/null 2>&1 || true
rm -rf /root/app && mkdir -p /root/app
cat > /root/app/docker-compose.yml <<'YEOF'
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
YEOF
rm -f /root/nginx_conf.txt /root/redis_ping.txt`,
		Checks: map[int]string{
			1: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' app-web-1 2>/dev/null)" = true ]`,
				"сервис web запущен через Compose",
				"cd /root/app && docker compose up -d"),
			2: dcheck(`grep -q 'worker_processes' /root/nginx_conf.txt 2>/dev/null`,
				"конфиг nginx сохранён",
				"cd /root/app && docker compose exec web cat /etc/nginx/nginx.conf > /root/nginx_conf.txt"),
			3: dcheck(`grep -q 'redis:alpine' /root/app/docker-compose.yml && grep -q 'cache_data' /root/app/docker-compose.yml && `+
				`[ "$(docker inspect -f '{{.State.Running}}' app-cache-1 2>/dev/null)" = true ] && `+
				`docker inspect -f '{{range .Mounts}}{{.Name}}:{{.Destination}} {{end}}' app-cache-1 | grep -q 'app_cache_data:/data'`,
				"сервис cache добавлен и запущен с именованным volume",
				"Добавь в docker-compose.yml сервис cache (image: redis:alpine, volumes: - cache_data:/data) и секцию volumes: cache_data:, затем docker compose up -d"),
			4: dcheck(`grep -qi '^PONG$' <(tr -d ' \r' < /root/redis_ping.txt 2>/dev/null)`,
				"redis ответил PONG",
				"cd /root/app && docker compose exec cache redis-cli ping > /root/redis_ping.txt (в файле должно быть только PONG)"),
			5: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' app-cache-1 2>/dev/null)" = false ] && [ "$(docker inspect -f '{{.State.Running}}' app-web-1 2>/dev/null)" = true ]`,
				"cache остановлен, web работает",
				"cd /root/app && docker compose stop cache (именно stop, не down)"),
			6: dcheck(`! docker inspect app-web-1 >/dev/null 2>&1 && ! docker inspect app-cache-1 >/dev/null 2>&1 && `+
				`! docker network inspect app_default >/dev/null 2>&1 && docker volume inspect app_cache_data >/dev/null 2>&1`,
				"стек удалён, volume сохранён",
				"cd /root/app && docker compose down (без флага -v, иначе volume тоже удалится)"),
		},
	},

	// ── Lab 2: сеть, DNS и масштабирование ──
	"ch-dkc-lab2": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
rm -rf /root/webapp && mkdir -p /root/webapp
cat > /root/webapp/app.py <<'PYEOF'
` + composeApp + `PYEOF
cat > /root/webapp/nginx.conf <<'NEOF'
server {
    listen 80;
    location / {
        proxy_pass http://app:5000;
    }
}
NEOF
cat > /root/webapp/docker-compose.yml <<'YEOF'
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
    depends_on:
      - app
    networks:
      - frontend
  app:
    image: python:3.12-alpine
    working_dir: /app
    command: python app.py
    environment:
      APP_NAME: webapp
    volumes:
      - ./app.py:/app/app.py:ro
    depends_on:
      - redis
    networks:
      - frontend
  redis:
    image: redis:alpine
    networks:
      - frontend

networks:
  frontend:
YEOF
rm -f /root/app_dns_ip.txt /root/app_response.txt /root/compose_network.txt`,
		Checks: map[int]string{
			1: dcheck(`[ "$(cd /root/webapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | sort -u | tr '\n' ',')" = "app,redis,web," ]`,
				"все три сервиса запущены",
				"cd /root/webapp && docker compose up -d"),
			2: dcheck(`[ -s /root/app_dns_ip.txt ] && [ "$(tr -d ' \n' < /root/app_dns_ip.txt)" = "$(docker inspect -f '{{(index .NetworkSettings.Networks "webapp_frontend").IPAddress}}' webapp-app-1 2>/dev/null)" ]`,
				"IP сервиса app определён по DNS-имени",
				"cd /root/webapp && docker compose exec web getent hosts app — первый столбец это IP; сохрани только его в /root/app_dns_ip.txt"),
			3: dcheck(`grep -qi 'Hello from webapp' /root/app_response.txt 2>/dev/null`,
				"ответ приложения через nginx сохранён",
				"curl -s http://localhost/ > /root/app_response.txt"),
			4: dcheck(`[ "$(tr -d ' \n' < /root/compose_network.txt 2>/dev/null)" = webapp_frontend ]`,
				"имя compose-сети сохранено",
				"docker network ls --format '{{.Name}}' | grep webapp > /root/compose_network.txt"),
			5: dcheck(`[ "$(cd /root/webapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^app$')" = 2 ]`,
				"запущено 2 экземпляра app",
				"cd /root/webapp && docker compose up -d --scale app=2"),
			6: dcheck(`[ "$(cd /root/webapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^app$')" = 2 ]`,
				"масштабирование подтверждено",
				"cd /root/webapp && docker compose ps — должно быть два контейнера сервиса app"),
			7: dcheck(`[ "$(cd /root/webapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^app$')" = 1 ] && `+
				`[ "$(docker inspect -f '{{.State.Running}}' webapp-web-1 2>/dev/null)" = true ]`,
				"вернулся один экземпляр app, web перезапущен",
				"cd /root/webapp && docker compose up -d --scale app=1 && docker compose restart web"),
			8: dcheck(`[ -z "$(cd /root/webapp && docker compose ps -q 2>/dev/null)" ]`,
				"стек остановлен и удалён",
				"cd /root/webapp && docker compose down"),
		},
	},

	// ── Lab 3: данные и volumes ──
	"ch-dkc-lab3": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
docker volume rm -f dbapp_db_data >/dev/null 2>&1 || true
rm -rf /root/dbapp /root/backups && mkdir -p /root/dbapp
cat > /root/dbapp/docker-compose.yml <<'YEOF'
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: postgres
    volumes:
      - db_data:/var/lib/postgresql/data

volumes:
  db_data:
YEOF
rm -f /root/users_count_before.txt /root/users_count_after.txt /root/db_volume_name.txt /root/volume_mountpoint.txt`,
		Checks: map[int]string{
			1: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' dbapp-db-1 2>/dev/null)" = true ] && docker volume inspect dbapp_db_data >/dev/null 2>&1`,
				"PostgreSQL запущен, volume создан",
				"cd /root/dbapp && docker compose up -d"),
			2: dcheck(`[ "$(cd /root/dbapp && docker compose exec -T db psql -U postgres -tAc "SELECT count(*) FROM users" 2>/dev/null | tr -d ' \r\n')" = 2 ]`,
				"таблица users создана, две строки добавлены",
				"docker compose exec db psql -U postgres -c \"CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100)); INSERT INTO users (name) VALUES ('Alice'), ('Bob');\""),
			3: dcheck(`[ "$(tr -d ' \r\n' < /root/users_count_before.txt 2>/dev/null)" = 2 ]`,
				"количество строк сохранено",
				"cd /root/dbapp && docker compose exec -T db psql -U postgres -tAc 'select count(*) from users' > /root/users_count_before.txt"),
			4: dcheck(`[ "$(tr -d ' \r\n' < /root/users_count_after.txt 2>/dev/null)" = 2 ] && [ "$(docker inspect -f '{{.State.Running}}' dbapp-db-1 2>/dev/null)" = true ]`,
				"данные пережили пересоздание стека",
				"cd /root/dbapp && docker compose down && docker compose up -d — затем снова select count(*) в /root/users_count_after.txt"),
			5: dcheck(`[ "$(tr -d ' \n' < /root/db_volume_name.txt 2>/dev/null)" = dbapp_db_data ]`,
				"имя volume сохранено",
				"docker volume ls --format '{{.Name}}' | grep dbapp > /root/db_volume_name.txt"),
			6: dcheck(`grep -qi 'CREATE TABLE' /root/backups/dump.sql 2>/dev/null && grep -q 'Alice' /root/backups/dump.sql && grep -q 'Bob' /root/backups/dump.sql`,
				"дамп содержит структуру и данные",
				"mkdir -p /root/backups && cd /root/dbapp && docker compose exec -T db pg_dump -U postgres postgres > /root/backups/dump.sql"),
			7: dcheck(`[ -s /root/volume_mountpoint.txt ] && [ "$(tr -d ' \n' < /root/volume_mountpoint.txt)" = "$(docker volume inspect -f '{{.Mountpoint}}' dbapp_db_data 2>/dev/null)" ]`,
				"путь volume на хосте сохранён",
				"docker volume inspect -f '{{.Mountpoint}}' dbapp_db_data > /root/volume_mountpoint.txt"),
			8: dcheck(`! docker volume inspect dbapp_db_data >/dev/null 2>&1 && [ -z "$(cd /root/dbapp && docker compose ps -q 2>/dev/null)" ]`,
				"стек и volume удалены",
				"cd /root/dbapp && docker compose down -v"),
		},
	},

	// ── Lab 4: переменные окружения ──
	"ch-dkc-lab4": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
rm -rf /root/envapp && mkdir -p /root/envapp
cat > /root/envapp/app.py <<'PYEOF'
` + composeApp + `PYEOF
cat > /root/envapp/.env.example <<'EEOF'
APP_PORT=8080
APP_DEBUG=true
REDIS_HOST=cache
REDIS_PORT=6379
APP_NAME=myapp
EEOF
cat > /root/envapp/docker-compose.yml <<'YEOF'
services:
  app:
    image: python:3.12-alpine
    working_dir: /app
    command: python app.py
    ports:
      - "${APP_PORT}:5000"
    environment:
      APP_NAME: ${APP_NAME}
      APP_DEBUG: ${APP_DEBUG}
      REDIS_HOST: ${REDIS_HOST}
      REDIS_PORT: ${REDIS_PORT}
    volumes:
      - ./app.py:/app/app.py:ro
  cache:
    image: redis:alpine
YEOF
rm -f /root/envapp/.env /root/envapp/.env.prod`,
		Checks: map[int]string{
			1: dcheck(`grep -q '^APP_PORT=8080$' /root/envapp/.env 2>/dev/null && grep -q '^APP_DEBUG=true$' /root/envapp/.env && `+
				`grep -q '^REDIS_HOST=cache$' /root/envapp/.env && grep -q '^REDIS_PORT=6379$' /root/envapp/.env && grep -q '^APP_NAME=myapp$' /root/envapp/.env`,
				".env создан со всеми переменными",
				"cp /root/envapp/.env.example /root/envapp/.env (в файле нужны APP_PORT, APP_DEBUG, REDIS_HOST, REDIS_PORT, APP_NAME)"),
			2: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' envapp-app-1 2>/dev/null)" = true ] && `+
				`docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' envapp-app-1 | grep -q '^APP_NAME=myapp$'`,
				"стек запущен, переменные из .env применились",
				"cd /root/envapp && docker compose up -d"),
			3: dcheck(`docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' envapp-app-1 2>/dev/null | grep -q '^APP_DEBUG=false$'`,
				"переменная переопределена из командной строки",
				"cd /root/envapp && APP_DEBUG=false docker compose up -d --force-recreate app"),
			4: dcheck(`grep -q 'monitor' /root/envapp/docker-compose.yml && grep -qE ':-[0-9]+' /root/envapp/docker-compose.yml && `+
				`[ "$(docker inspect -f '{{.State.Running}}' envapp-monitor-1 2>/dev/null)" = true ]`,
				"сервис monitor с дефолтным портом запущен",
				"Добавь в docker-compose.yml сервис monitor с портом вида \"${MONITOR_PORT:-9090}:80\" (image: nginx:alpine) и подними стек"),
			5: dcheck(`grep -q '^APP_DEBUG=false$' /root/envapp/.env.prod 2>/dev/null && grep -q '^APP_PORT=80$' /root/envapp/.env.prod`,
				".env.prod создан",
				"printf 'APP_DEBUG=false\\nAPP_PORT=80\\n' > /root/envapp/.env.prod"),
			6: dcheck(`[ -z "$(cd /root/envapp && docker compose ps -q 2>/dev/null)" ]`,
				"стек остановлен",
				"cd /root/envapp && docker compose down"),
		},
	},

	// ── Lab 5: override-файлы и профили ──
	"ch-dkc-lab5": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
rm -rf /root/override && mkdir -p /root/override
cat > /root/override/app.py <<'PYEOF'
` + composeApp + `PYEOF
cat > /root/override/docker-compose.yml <<'YEOF'
services:
  app:
    image: python:3.12-alpine
    working_dir: /app
    command: python app.py
    environment:
      APP_NAME: overrideapp
    volumes:
      - ./app.py:/app/app.py:ro
YEOF
cat > /root/override/docker-compose.dev.yml <<'YEOF'
services:
  app:
    environment:
      APP_ENV: development
YEOF
rm -f /root/override/docker-compose.override.yml`,
		Checks: map[int]string{
			1: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' override-app-1 2>/dev/null)" = true ]`,
				"базовый стек запущен",
				"cd /root/override && docker compose -f docker-compose.yml up -d"),
			2: dcheck(`[ -z "$(cd /root/override && docker compose -f docker-compose.yml ps -q 2>/dev/null)" ]`,
				"базовый стек остановлен",
				"cd /root/override && docker compose -f docker-compose.yml down"),
			3: dcheck(`docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' override-app-1 2>/dev/null | grep -q '^APP_ENV=development$'`,
				"файлы объединены, APP_ENV=development",
				"cd /root/override && docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d"),
			4: dcheck(`grep -q 'AUTO_OVERRIDE' /root/override/docker-compose.override.yml 2>/dev/null && `+
				`docker inspect -f '{{range .Config.Env}}{{println .}}{{end}}' override-app-1 2>/dev/null | grep -q '^AUTO_OVERRIDE=true$'`,
				"override-файл подхватился автоматически",
				"Создай /root/override/docker-compose.override.yml с environment: AUTO_OVERRIDE: \"true\" для сервиса app, затем cd /root/override && docker compose up -d --force-recreate"),
			5: dcheck(`grep -q 'debug' /root/override/docker-compose.yml && `+
				`[ "$(docker inspect -f '{{.State.Running}}' override-debugger-1 2>/dev/null)" = true ]`,
				"сервис с профилем debug запущен",
				"Добавь сервис debugger с profiles: [debug] (например image: alpine:latest, command: sleep 3600), затем docker compose --profile debug up -d"),
			6: dcheck(`[ -z "$(cd /root/override && docker compose --profile debug ps -q 2>/dev/null)" ]`,
				"весь стек остановлен, включая профильные сервисы",
				"cd /root/override && docker compose --profile debug down"),
		},
	},

	// ── Lab 6: healthcheck и зависимости ──
	"ch-dkc-lab6": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
rm -rf /root/healthapp && mkdir -p /root/healthapp
cat > /root/healthapp/app.py <<'PYEOF'
` + composeApp + `PYEOF
cat > /root/healthapp/docker-compose.yml <<'YEOF'
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: postgres
  app:
    image: python:3.12-alpine
    working_dir: /app
    command: python app.py
    volumes:
      - ./app.py:/app/app.py:ro
    depends_on:
      - db
  web:
    image: nginx:alpine
YEOF`,
		Checks: map[int]string{
			1: dcheck(`grep -q 'healthcheck' /root/healthapp/docker-compose.yml && grep -qE 'pg_isready|pg_isready -U' /root/healthapp/docker-compose.yml && `+
				`grep -q 'service_healthy' /root/healthapp/docker-compose.yml`,
				"healthcheck для db и условие ожидания у app описаны",
				"У сервиса db: healthcheck: test: [\"CMD-SHELL\", \"pg_isready -U postgres\"], interval: 5s; у app: depends_on: db: condition: service_healthy"),
			2: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' healthapp-app-1 2>/dev/null)" = true ]`,
				"стек запущен",
				"cd /root/healthapp && docker compose up -d"),
			3: dcheck(`[ "$(docker inspect -f '{{.State.Health.Status}}' healthapp-db-1 2>/dev/null)" = healthy ]`,
				"база в статусе healthy",
				"Подожди несколько секунд: docker compose ps покажет (healthy) у db"),
			4: dcheck(`grep -q 'wget -qO- http://localhost/' /root/healthapp/docker-compose.yml`,
				"healthcheck для web описан",
				"У сервиса web добавь healthcheck: test: [\"CMD-SHELL\", \"wget -qO- http://localhost/ || exit 1\"], interval: 5s"),
			5: dcheck(`[ "$(docker inspect -f '{{.State.Health.Status}}' healthapp-db-1 2>/dev/null)" = healthy ] && `+
				`[ "$(docker inspect -f '{{.State.Health.Status}}' healthapp-web-1 2>/dev/null)" = healthy ]`,
				"оба сервиса healthy",
				"Пересоздай стек (docker compose up -d) и подожди — оба контейнера должны стать healthy"),
			6: dcheck(`[ -z "$(cd /root/healthapp && docker compose ps -q 2>/dev/null)" ]`,
				"стек остановлен",
				"cd /root/healthapp && docker compose down"),
		},
	},

	// ── Lab 7: масштабирование ──
	"ch-dkc-lab7": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
` + composeDown + `
rm -rf /root/scaleapp && mkdir -p /root/scaleapp
cat > /root/scaleapp/docker-compose.yml <<'YEOF'
services:
  worker:
    image: alpine:latest
    command: sh -c "while true; do sleep 5; done"
YEOF`,
		Checks: map[int]string{
			1: dcheck(`[ "$(cd /root/scaleapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^worker$')" = 1 ]`,
				"один worker запущен",
				"cd /root/scaleapp && docker compose up -d"),
			2: dcheck(`[ "$(cd /root/scaleapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^worker$')" = 3 ]`,
				"три экземпляра worker запущены",
				"cd /root/scaleapp && docker compose up -d --scale worker=3"),
			3: dcheck(`[ "$(cd /root/scaleapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^worker$')" = 3 ]`,
				"масштабирование подтверждено",
				"cd /root/scaleapp && docker compose ps — должно быть ровно 3 контейнера worker"),
			4: dcheck(`grep -q 'replicas' /root/scaleapp/docker-compose.yml && `+
				`[ "$(cd /root/scaleapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^worker$')" = 2 ]`,
				"replicas: 2 описано и применено",
				"Добавь сервису worker секцию deploy: replicas: 2, затем docker compose up -d"),
			5: dcheck(`[ "$(cd /root/scaleapp && docker compose ps --status running --format '{{.Service}}' 2>/dev/null | grep -c '^worker$')" = 1 ]`,
				"остался один worker",
				"cd /root/scaleapp && docker compose up -d --scale worker=1"),
			6: dcheck(`[ -z "$(cd /root/scaleapp && docker compose ps -q 2>/dev/null)" ]`,
				"стек остановлен",
				"cd /root/scaleapp && docker compose down"),
		},
	},
}
