# GoLearn — Learning Management System

## What This Is
An LMS for learning **DevOps**: Linux, Git, Docker, Kubernetes, Helm, SQL — theory,
quizzes and hands-on labs in a real terminal. Deployed at `learn.prod-factory.ru`
(see `deploy/berg/README.md`).

The Go courses were removed (migration `013`); their content still lives in
`cmd/seed/mod*.go` but is no longer registered in `getAllModules()`.

## Stack
- **Backend:** Go 1.22+, chi router, pgx (PostgreSQL)
- **Frontend:** Server-side HTML templates (Go html/template)
- **Database:** PostgreSQL 16 (via docker-compose, port 5433)
- **No external CSS frameworks** — custom CSS in templates

## How to Run
```bash
# 1. Build the lab sandbox images (once; network is needed only at build time)
docker build -t golearn/sandbox:latest    -f deploy/sandbox/Dockerfile    deploy/sandbox
docker tag   golearn/sandbox:latest golearn/git:latest
docker build -t golearn/sandbox-pg:latest -f deploy/sandbox-pg/Dockerfile deploy/sandbox-pg
bash deploy/sandbox-docker/prepare.sh     # images + offline registry (needs network)
docker build -t golearn/sandbox-docker:latest -f deploy/sandbox-docker/Dockerfile deploy/sandbox-docker
bash deploy/sandbox-k8s/prepare.sh        # k3s/helm binaries + airgap images
docker build -t golearn/sandbox-k8s:latest    -f deploy/sandbox-k8s/Dockerfile    deploy/sandbox-k8s

# 2. Start database
docker compose up -d

# 3. Seed course content (applies pending migrations too)
go run ./cmd/seed

# 4. Start server (also applies migrations, creates the first admin)
go run ./cmd/server
# Open http://localhost:8080 — login: ADMIN_EMAIL / ADMIN_PASSWORD from .env
```

Migrations run automatically on startup (`internal/migrate`, tracked in the
`schema_migrations` table) — no manual psql step.

## Lab sandboxes
Shell labs run one container **per lesson** (`gl-s-u<user>-l<lesson>`), always
with `--network none`. Four images, picked per lesson via `tasks.sandbox_image`:

| Image | Used by | Notes |
|---|---|---|
| `golearn/sandbox` | Linux, Git, тренажёры | CLI tools baked in; offline apt repo; `systemctl` shim (`deploy/sandbox/systemctl`) |
| `golearn/sandbox-pg` | SQL | PostgreSQL server; lesson setup calls `pg-start` |
| `golearn/sandbox-docker` | Docker, Compose | full Docker Engine (**privileged**); `docker-start`; offline Docker Hub stand-in so `docker pull` works |
| `golearn/sandbox-k8s` | Kubernetes, Helm | single-node k3s + kubectl/helm (**privileged**); `k8s-start`; images and traefik baked in |

Consequences to keep in mind:
- every CLI tool and every image a lesson needs must be baked in — there is no network;
- the Docker/Kubernetes images run **privileged** and keep runtime state on a
  per-session volume (`gl-dind-u<user>-l<lesson>`); `internal/runner/shell.go`
  decides this from the image name;
- `mount` (CAP_SYS_ADMIN) still does not work, so those few tasks stay manual.

## Lab fixtures & auto-checks
Course tasks get their environment and validator from `cmd/seed/labs_*.go`
(`labFixtures` in `cmd/seed/labfixtures.go`): one `Setup` per lesson creates the
files the tasks reference, and each task gets a `Check` (exit 0 = solved).
Verify them with the regression harness:
```bash
./scripts/labcheck/run.sh linux-start   # check must FAIL before, PASS after the reference solution
```

## Project Structure
```
cmd/server/     — HTTP server entry point
cmd/seed/       — Database seeder with all course content
internal/
  config/       — Environment config
  handler/      — HTTP handlers (dashboard, lesson, quiz, tasks, progress)
  model/        — Data models
  repository/   — PostgreSQL queries
  templates/    — HTML templates (layouts/, pages/, components/)
  static/       — CSS/JS/images
migrations/     — SQL migrations
```

## Learning Context
- Student builds WatchTogether through lesson tasks
- Each lesson has: theory (HTML content), quiz (multiple choice), practical tasks
- Progress is tracked in the database
- Brain/memory for this project: ~/.claude/projects/-Users-backendraz-GolandProjects-GoLearn/memory/

## Rules
- This is a LEARNING project — explain things thoroughly
- When student asks about a topic, check which lesson covers it
- Update learning_stage.md in brain after each completed lesson
- Language: Russian for explanations, English for code
