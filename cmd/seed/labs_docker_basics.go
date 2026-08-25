package main

// Fixtures + auto-checks for "Docker: контейнеры и образы" (docker-basics).
//
// These lessons run in golearn/sandbox-docker — a privileged lab container with
// its own Docker Engine. `docker-start` boots that engine, loads the images the
// lessons use and brings up an offline stand-in for Docker Hub, so even
// `docker pull` works without network.

// dockerBoot is the first line of every Docker lesson setup.
const dockerBoot = `set -e
docker-start`

// dcheck wraps a check so it fails with a clear message when the engine is down
// instead of with a confusing "cannot connect to the Docker daemon".
func dcheck(cond, good, bad string) string {
	// Joined with "; " (not a newline): a check has to stay valid when it is
	// collapsed onto one line.
	return `if ! docker info >/dev/null 2>&1; then ` +
		fail("Docker-движок ещё поднимается. Подожди несколько секунд и нажми «Проверить» снова.") +
		`; fi; ` + check(cond, good, bad)
}

// appPy is a dependency-free HTTP app (no pip install in an offline lab).
const appPy = `import os
from http.server import BaseHTTPRequestHandler, HTTPServer

VERSION = os.environ.get("APP_VERSION", "1.0")

class H(BaseHTTPRequestHandler):
    def do_GET(self):
        body = f"Hello Docker! version={VERSION}\n".encode()
        self.send_response(200)
        self.send_header("Content-Type", "text/plain")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *a):
        pass

HTTPServer(("0.0.0.0", 5000), H).serve_forever()
`

const goMain = `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Go!")
	})
	http.ListenAndServe(":8080", nil)
}
`

var dockerBasicsLabs = map[string]labSpec{
	// ── Lab 1: жизненный цикл контейнера ──
	"ch-dock-lab1": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rm -f mynginx >/dev/null 2>&1 || true
rm -f /root/mynginx_ip.txt`,
		Checks: map[int]string{
			1: dcheck(`docker ps -a --filter ancestor=nginx:alpine --format '{{.ID}}' | grep -q .`,
				"контейнер из nginx:alpine был запущен",
				"Запусти: docker run nginx:alpine (остановить — Ctrl+C)"),
			2: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' mynginx 2>/dev/null)" = true ]`,
				"контейнер mynginx работает в фоне",
				"docker run -d --name mynginx nginx:alpine"),
			3: dcheck(`[ -s /root/mynginx_ip.txt ] && [ "$(tr -d ' \n' < /root/mynginx_ip.txt)" = "$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mynginx 2>/dev/null)" ]`,
				"IP-адрес контейнера сохранён верно",
				"docker inspect mynginx — найди IPAddress. Быстро: docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mynginx > /root/mynginx_ip.txt"),
			4: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' mynginx 2>/dev/null)" = true ] && `+
				`[ "$(date -d "$(docker inspect -f '{{.State.StartedAt}}' mynginx)" +%s)" -gt "$(( $(date -d "$(docker inspect -f '{{.Created}}' mynginx)" +%s) + 2 ))" ]`,
				"контейнер был остановлен и запущен заново",
				"docker stop mynginx затем docker start mynginx (проверка смотрит, что время старта позже времени создания)"),
			5: dcheck(`! docker inspect mynginx >/dev/null 2>&1`,
				"контейнер mynginx удалён",
				"docker stop mynginx && docker rm mynginx"),
		},
	},

	// ── Lab 2: образы и теги ──
	"ch-dock-lab2": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rmi -f myredis:v1 >/dev/null 2>&1 || true
docker rmi -f redis:alpine >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: dcheck(`docker image inspect redis:alpine >/dev/null 2>&1`,
				"образ redis:alpine загружен",
				"docker pull redis:alpine"),
			2: dcheck(`docker image inspect myredis:v1 >/dev/null 2>&1 && [ "$(docker image inspect -f '{{.Id}}' myredis:v1)" = "$(docker image inspect -f '{{.Id}}' redis:alpine)" ]`,
				"тег myredis:v1 указывает на тот же образ",
				"docker tag redis:alpine myredis:v1"),
			3: dcheck(`! docker image inspect myredis:v1 >/dev/null 2>&1 && docker image inspect redis:alpine >/dev/null 2>&1`,
				"тег myredis:v1 удалён, redis:alpine на месте",
				"docker rmi myredis:v1 (удаляется только тег, сам образ остаётся под именем redis:alpine)"),
		},
	},

	// ── Lab 3: сборка своего образа ──
	"ch-dock-lab3": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rm -f myapp myapp2 >/dev/null 2>&1 || true
docker rmi -f myapp:v1 myapp:v2 >/dev/null 2>&1 || true
rm -rf /root/myapp && mkdir -p /root/myapp
cat > /root/myapp/app.py <<'PYEOF'
` + appPy + `PYEOF
cat > /root/myapp/Dockerfile <<'DEOF'
FROM python:3.12-alpine
WORKDIR /app
COPY app.py .
EXPOSE 5000
CMD ["python", "app.py"]
DEOF`,
		Checks: map[int]string{
			1: dcheck(`docker image inspect myapp:v1 >/dev/null 2>&1`,
				"образ myapp:v1 собран",
				"cd /root/myapp && docker build -t myapp:v1 ."),
			2: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' myapp 2>/dev/null)" = true ] && docker port myapp 2>/dev/null | grep -q 5000`,
				"контейнер myapp работает с проброшенным портом",
				"docker run -d --name myapp -p 5000:5000 myapp:v1"),
			3: dcheck(`curl -s --max-time 5 http://localhost:5000/ | grep -q 'Hello Docker'`,
				"приложение отвечает",
				"curl http://localhost:5000 — если ответа нет, проверь docker logs myapp"),
			4: dcheck(`! docker inspect myapp >/dev/null 2>&1 && docker image inspect myapp:v2 >/dev/null 2>&1 && `+
				`[ "$(docker image inspect -f '{{index .Config.Labels "version"}}' myapp:v2 2>/dev/null)" = 2.0 ] && `+
				`docker image inspect -f '{{range .Config.Env}}{{println .}}{{end}}' myapp:v2 | grep -q '^APP_VERSION=2.0$'`,
				"образ myapp:v2 собран с LABEL и ENV",
				"Удали контейнер (docker rm -f myapp), добавь в Dockerfile строки LABEL version=\"2.0\" и ENV APP_VERSION=2.0, затем docker build -t myapp:v2 ."),
			5: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' myapp2 2>/dev/null)" = true ] && curl -s --max-time 5 http://localhost:5001/ | grep -q '2\.0'`,
				"myapp2 отвечает версией 2.0",
				"docker run -d --name myapp2 -p 5001:5000 myapp:v2 — затем curl http://localhost:5001"),
			6: dcheck(`! docker inspect myapp2 >/dev/null 2>&1`,
				"контейнер myapp2 удалён",
				"docker rm -f myapp2"),
		},
	},

	// ── Lab 4: multi-stage ──
	"ch-dock-lab4": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rm -f goapp >/dev/null 2>&1 || true
docker rmi -f goapp:multi goapp:nocache >/dev/null 2>&1 || true
rm -rf /root/goapp && mkdir -p /root/goapp
cat > /root/goapp/main.go <<'GOEOF'
` + goMain + `GOEOF
printf 'module goapp\n\ngo 1.21\n' > /root/goapp/go.mod
cat > /root/goapp/Dockerfile.multistage <<'DEOF'
FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
ENV CGO_ENABLED=0 GOFLAGS=-mod=mod GOPROXY=off
RUN go build -o /out/server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /out/server /app/server
EXPOSE 8080
CMD ["/app/server"]
DEOF`,
		Checks: map[int]string{
			1: dcheck(`docker image inspect goapp:multi >/dev/null 2>&1`,
				"образ goapp:multi собран",
				"cd /root/goapp && docker build -f Dockerfile.multistage -t goapp:multi ."),
			2: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' goapp 2>/dev/null)" = true ] && docker port goapp 2>/dev/null | grep -q 8080`,
				"контейнер goapp работает с портом 8080",
				"docker run -d --name goapp -p 8080:8080 goapp:multi"),
			3: dcheck(`curl -s --max-time 5 http://localhost:8080/ | grep -q 'Hello from Go'`,
				"Go-сервер отвечает",
				"curl http://localhost:8080 — если пусто, смотри docker logs goapp"),
			4: dcheck(`docker image inspect goapp:nocache >/dev/null 2>&1`,
				"образ goapp:nocache собран",
				"cd /root/goapp && docker build --no-cache -f Dockerfile.multistage -t goapp:nocache ."),
			5: dcheck(`! docker inspect goapp >/dev/null 2>&1`,
				"контейнер goapp удалён",
				"docker rm -f goapp"),
			6: dcheck(`! docker inspect goapp >/dev/null 2>&1 && docker image inspect goapp:nocache >/dev/null 2>&1`,
				"контейнера нет, образ остался",
				"Контейнер и образ — разные сущности: docker ps -a покажет контейнеры, docker images — образы"),
		},
	},

	// ── Lab 5: сети ──
	"ch-dock-lab5": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rm -f web1 web2 outsider >/dev/null 2>&1 || true
docker network rm mynet >/dev/null 2>&1 || true
rm -rf /root/docker-lab5
mkdir -p /root/docker-lab5/web1 /root/docker-lab5/web2
echo '<h1>WEB1</h1>' > /root/docker-lab5/web1/index.html
echo '<h1>WEB2</h1>' > /root/docker-lab5/web2/index.html
docker run -d --name outsider nginx:alpine >/dev/null
rm -f /root/web1_ip.txt`,
		Checks: map[int]string{
			1: dcheck(`[ "$(docker network inspect -f '{{.Driver}}' mynet 2>/dev/null)" = bridge ]`,
				"сеть mynet создана",
				"docker network create mynet"),
			2: dcheck(`docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}} {{end}}' web1 2>/dev/null | grep -q mynet && curl -s --max-time 5 http://localhost/ | grep -q WEB1`,
				"web1 в сети mynet и отдаёт свой index.html",
				"docker run -d --name web1 --network mynet -p 80:80 -v /root/docker-lab5/web1:/usr/share/nginx/html nginx:alpine"),
			3: dcheck(`docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}} {{end}}' web2 2>/dev/null | grep -q mynet && `+
				`docker run --rm --network mynet alpine:latest wget -qO- --timeout=5 http://web2/ 2>/dev/null | grep -q WEB2`,
				"web2 доступен по имени внутри mynet",
				"docker run -d --name web2 --network mynet -v /root/docker-lab5/web2:/usr/share/nginx/html nginx:alpine (порт наружу не нужен)"),
			4: dcheck(`[ -s /root/web1_ip.txt ] && [ "$(tr -d ' \n' < /root/web1_ip.txt)" = "$(docker inspect -f '{{(index .NetworkSettings.Networks "mynet").IPAddress}}' web1 2>/dev/null)" ]`,
				"IP web1 в сети mynet сохранён",
				"docker network inspect mynet — найди web1 и его IPv4Address (без маски /16) и запиши в /root/web1_ip.txt"),
			5: dcheck(`! docker inspect web1 >/dev/null 2>&1 && ! docker inspect web2 >/dev/null 2>&1 && ! docker inspect outsider >/dev/null 2>&1 && ! docker network inspect mynet >/dev/null 2>&1`,
				"контейнеры и сеть удалены",
				"docker rm -f web1 web2 outsider && docker network rm mynet"),
		},
	},

	// ── Lab 6: тома и bind mount ──
	"ch-dock-lab6": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rm -f vol_test vol_reader nginx_bind >/dev/null 2>&1 || true
docker volume rm -f mydata >/dev/null 2>&1 || true
rm -rf /root/html && mkdir -p /root/html
echo '<h1>BIND MOUNT OK</h1>' > /root/html/index.html
rm -f /root/mydata_mountpoint.txt`,
		Checks: map[int]string{
			1: dcheck(`docker volume inspect mydata >/dev/null 2>&1`,
				"volume mydata создан",
				"docker volume create mydata"),
			2: dcheck(`docker inspect -f '{{range .Mounts}}{{.Name}}:{{.Destination}} {{end}}' vol_test 2>/dev/null | grep -q 'mydata:/data'`,
				"vol_test запущен с примонтированным mydata",
				"docker run -d --name vol_test -v mydata:/data alpine:latest sleep 3600"),
			3: dcheck(`docker run --rm -v mydata:/data alpine:latest cat /data/test.txt 2>/dev/null | grep -q 'persistent data'`,
				"файл записан в volume",
				"docker exec vol_test sh -c 'echo \"persistent data\" > /data/test.txt'"),
			4: dcheck(`! docker inspect vol_test >/dev/null 2>&1 && docker inspect vol_reader >/dev/null 2>&1 && `+
				`docker run --rm -v mydata:/data alpine:latest cat /data/test.txt 2>/dev/null | grep -q 'persistent data'`,
				"данные пережили удаление контейнера",
				"docker rm -f vol_test, затем docker run --name vol_reader -v mydata:/data alpine:latest cat /data/test.txt"),
			5: dcheck(`[ "$(docker inspect -f '{{.State.Running}}' nginx_bind 2>/dev/null)" = true ] && curl -s --max-time 5 http://localhost/ | grep -q 'BIND MOUNT OK'`,
				"nginx_bind отдаёт файлы из /root/html",
				"docker run -d --name nginx_bind -p 80:80 -v /root/html:/usr/share/nginx/html nginx:alpine"),
			6: dcheck(`[ -s /root/mydata_mountpoint.txt ] && [ "$(tr -d ' \n' < /root/mydata_mountpoint.txt)" = "$(docker volume inspect -f '{{.Mountpoint}}' mydata 2>/dev/null)" ]`,
				"путь тома на хосте сохранён",
				"docker volume inspect mydata — поле Mountpoint. Быстро: docker volume inspect -f '{{.Mountpoint}}' mydata > /root/mydata_mountpoint.txt"),
			7: dcheck(`! docker inspect nginx_bind >/dev/null 2>&1 && ! docker inspect vol_reader >/dev/null 2>&1 && ! docker volume inspect mydata >/dev/null 2>&1`,
				"контейнеры и volume удалены",
				"docker rm -f nginx_bind vol_reader && docker volume rm mydata"),
		},
	},

	// ── Lab 7: registry-теги ──
	"ch-dock-lab7": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rmi -f registry.company.local/devops/mynginx:v1 registry.company.local/devops/myredis:prod >/dev/null 2>&1 || true
docker rmi -f redis:7-alpine >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: dcheck(`docker image inspect registry.company.local/devops/mynginx:v1 >/dev/null 2>&1`,
				"registry-тег для nginx создан",
				"docker tag nginx:alpine registry.company.local/devops/mynginx:v1"),
			2: dcheck(`docker images registry.company.local/devops/mynginx --format '{{.Tag}}' | grep -qx v1`,
				"тег виден в списке образов",
				"docker images registry.company.local/devops/mynginx"),
			3: dcheck(`docker image inspect redis:7-alpine >/dev/null 2>&1`,
				"образ redis:7-alpine загружен",
				"docker pull redis:7-alpine"),
			4: dcheck(`docker image inspect registry.company.local/devops/myredis:prod >/dev/null 2>&1`,
				"registry-тег для redis создан",
				"docker tag redis:7-alpine registry.company.local/devops/myredis:prod"),
			5: dcheck(`! docker image inspect registry.company.local/devops/mynginx:v1 >/dev/null 2>&1 && `+
				`! docker image inspect registry.company.local/devops/myredis:prod >/dev/null 2>&1 && `+
				`docker image inspect nginx:alpine >/dev/null 2>&1`,
				"registry-теги удалены, исходные образы на месте",
				"docker rmi registry.company.local/devops/mynginx:v1 registry.company.local/devops/myredis:prod"),
		},
	},

	// ── Lab 8: итоговый стек ──
	"ch-dock-lab8": {
		Image: sandboxImageDocker,
		Setup: dockerBoot + `
docker rm -f final-api final-worker final-frontend >/dev/null 2>&1 || true
docker network rm final-net >/dev/null 2>&1 || true
docker rmi -f final-api:1.0 final-worker:1.0 final-frontend:1.0 >/dev/null 2>&1 || true
rm -rf /root/final-stack
mkdir -p /root/final-stack/api /root/final-stack/worker /root/final-stack/frontend

cat > /root/final-stack/api/app.py <<'PYEOF'
import json, os, urllib.request
from http.server import BaseHTTPRequestHandler, HTTPServer

VERSION = os.environ.get("APP_VERSION", "0.0")
WORKER = os.environ.get("WORKER_URL", "http://worker:8080")

class H(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith("/api/order"):
            try:
                with urllib.request.urlopen(WORKER + "/health", timeout=3) as r:
                    worker = json.loads(r.read().decode())
            except Exception as e:
                worker = {"error": str(e)}
            payload = {"version": VERSION, "worker": worker}
            payload.update(worker)
            body = json.dumps(payload).encode()
        else:
            body = json.dumps({"service": "api", "version": VERSION}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, *a):
        pass

HTTPServer(("0.0.0.0", 5000), H).serve_forever()
PYEOF

cat > /root/final-stack/worker/main.go <<'GOEOF'
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, ` + "`" + `{"service":"inventory","status":"ok","version":"1.0"}` + "`" + `)
	})
	http.ListenAndServe(":8080", nil)
}
GOEOF
printf 'module worker\n\ngo 1.21\n' > /root/final-stack/worker/go.mod

echo '<h1>Final stack frontend</h1>' > /root/final-stack/frontend/index.html
cat > /root/final-stack/frontend/nginx.conf <<'NEOF'
server {
    listen 80;
    location / {
        root /usr/share/nginx/html;
        index index.html;
    }
    location /api/ {
        proxy_pass http://api:5000;
    }
}
NEOF`,
		Descs: map[int]string{
			// Экспорт курса потерял текст этого задания (в базе осталось "$29").
			7: `<p>Запусти контейнер <code>final-frontend</code> из образа <code>final-frontend:1.0</code> в сети <code>final-net</code> с alias <code>frontend</code>.</p>
<p>Это единственный контейнер стека, опубликованный наружу: пробрось порт <code>80</code> хоста на порт <code>80</code> контейнера.</p>
<p>После запуска главная страница должна открываться по <code>http://localhost/</code>, а запрос <code>http://localhost/api/order</code> — проксироваться в API и возвращать данные worker.</p>`,
		},
		Checks: map[int]string{
			1: dcheck(`docker image inspect final-api:1.0 >/dev/null 2>&1 && `+
				`docker image inspect -f '{{range .Config.Env}}{{println .}}{{end}}' final-api:1.0 | grep -q '^APP_VERSION=1.0$' && `+
				`docker image inspect -f '{{range .Config.Env}}{{println .}}{{end}}' final-api:1.0 | grep -q '^WORKER_URL=' && `+
				`grep -q 'tmp' /root/final-stack/api/.dockerignore 2>/dev/null && grep -q '\*\.log' /root/final-stack/api/.dockerignore`,
				"final-api:1.0 собран, ENV и .dockerignore на месте",
				"В /root/final-stack/api/Dockerfile: FROM python:3.12-alpine, WORKDIR /app, COPY app.py ., ENV APP_VERSION=1.0, ENV WORKER_URL=http://worker:8080, EXPOSE 5000, CMD [\"python\",\"app.py\"]; рядом .dockerignore со строками tmp и *.log"),
			2: dcheck(`docker image inspect final-worker:1.0 >/dev/null 2>&1 && `+
				`grep -qi 'AS builder' /root/final-stack/worker/Dockerfile && grep -qi 'COPY --from=builder' /root/final-stack/worker/Dockerfile && `+
				`! docker run --rm --entrypoint sh final-worker:1.0 -c 'command -v go' >/dev/null 2>&1`,
				"final-worker:1.0 собран multi-stage, компилятора в финальном образе нет",
				"Первый stage: FROM golang:1.21-alpine AS builder + go build; второй: FROM alpine:latest и COPY --from=builder. Для сборки без сети добавь ENV GOPROXY=off GOFLAGS=-mod=mod"),
			3: dcheck(`docker image inspect final-frontend:1.0 >/dev/null 2>&1 && `+
				`docker run --rm --entrypoint sh final-frontend:1.0 -c 'grep -q "proxy_pass http://api:5000" /etc/nginx/conf.d/default.conf && test -f /usr/share/nginx/html/index.html' >/dev/null 2>&1`,
				"final-frontend:1.0 собран с нужными файлами",
				"FROM nginx:alpine; COPY index.html /usr/share/nginx/html/index.html; COPY nginx.conf /etc/nginx/conf.d/default.conf; EXPOSE 80"),
			4: dcheck(`[ "$(docker network inspect -f '{{.Driver}}' final-net 2>/dev/null)" = bridge ]`,
				"сеть final-net создана",
				"docker network create final-net"),
			5: dcheck(`docker run --rm --network final-net alpine:latest wget -qO- --timeout=5 http://worker:8080/health 2>/dev/null | grep -q '"service":"inventory"'`,
				"worker отвечает по алиасу внутри сети",
				"docker run -d --name final-worker --network final-net --network-alias worker final-worker:1.0"),
			6: dcheck(`docker run --rm --network final-net alpine:latest wget -qO- --timeout=5 http://api:5000/api/order 2>/dev/null | grep -q '"service": *"inventory"' && `+
				`docker run --rm --network final-net alpine:latest wget -qO- --timeout=5 http://api:5000/api/order 2>/dev/null | grep -q '"version": *"1.0"'`,
				"API отвечает и ходит в worker",
				"docker run -d --name final-api --network final-net --network-alias api -e APP_VERSION=1.0 -e WORKER_URL=http://worker:8080 final-api:1.0"),
			7: dcheck(`curl -s --max-time 5 http://localhost/ | grep -qi 'frontend' && curl -s --max-time 5 http://localhost/api/order | grep -q 'inventory'`,
				"frontend опубликован и проксирует в API",
				"docker run -d --name final-frontend --network final-net --network-alias frontend -p 80:80 final-frontend:1.0"),
			8: dcheck(`! docker inspect final-frontend >/dev/null 2>&1 && ! docker inspect final-api >/dev/null 2>&1 && `+
				`! docker inspect final-worker >/dev/null 2>&1 && ! docker network inspect final-net >/dev/null 2>&1`,
				"стек остановлен, сеть удалена",
				"docker rm -f final-frontend final-api final-worker && docker network rm final-net"),
		},
	},
}
