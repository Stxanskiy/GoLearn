# GoLearn — Learning Management System

## What This Is
An LMS for learning **DevOps**: Linux, Git, Docker, Kubernetes, Helm — theory,
quizzes and hands-on labs in a real terminal. Deployed at `tot.prod-factory.ru`
(see `deploy/berg/README.md`).

Course content is JSON in `cmd/seed/content/`, compiled into the seeder with
`go:embed` and imported by `cmd/seed/import.go`. No course is written in Go any
more — the Go courses and the last `mod*.go` files are gone.

There is no SQL course. "Экспресс-курс по SQL" was withdrawn in migration `030`:
the first task of every one of its labs ran `psql -U student`, and the sandbox
only ever created a `root` role, so each lab failed on its first command. It was
never published. Its content, fixtures and reference solutions are deleted; the
`database` track keeps its order_num band so a replacement lands in the right
place rather than at the end of the DevOps path.

The database is the source of truth for anything edited in the studio: a lesson
or module with `edited_at` set (migration `026`) is never overwritten or pruned
by the seeder, so a deploy does not undo an author's changes.

## Stack
- **Backend:** Go 1.22+, chi router, pgx (PostgreSQL)
- **Frontend:** separate repo (`space-symbol/frontend-tot`, Next.js). This service
  renders no pages — it is a JSON API, and `api/openapi.yaml` is the contract
  between the two, served at `/docs`.
- **Database:** PostgreSQL 16 (via docker-compose, port 5433)

## How to Run
```bash
# 1. Build the lab sandbox images (once; network is needed only at build time)
docker build -t golearn/sandbox:latest    -f deploy/sandbox/Dockerfile    deploy/sandbox
docker tag   golearn/sandbox:latest golearn/git:latest
# sandbox-pg is only needed if the SQL course comes back (see below) — skip it
docker build -t golearn/sandbox-pg:latest -f deploy/sandbox-pg/Dockerfile deploy/sandbox-pg
bash deploy/sandbox-docker/prepare.sh     # images + offline registry (needs network)
docker build -t golearn/sandbox-docker:latest -f deploy/sandbox-docker/Dockerfile deploy/sandbox-docker
bash deploy/sandbox-k8s/prepare.sh        # k3s/helm binaries + airgap images
docker build -t golearn/sandbox-k8s:latest    -f deploy/sandbox-k8s/Dockerfile    deploy/sandbox-k8s

# 2. Local settings (database, first admin, object storage)
cp .env.example .env

# 3. Start database and object storage (RustFS on :9010, console :9011)
docker compose up -d

# 4. Seed course content (applies pending migrations too)
go run ./cmd/seed

# 5. Start server (also applies migrations, creates the first admin)
go run ./cmd/server
# JSON API on http://localhost:8080 (PORT): /api/v1, Swagger UI at /docs,
# /healthz, /readyz, Prometheus metrics at /metrics. There are no pages here —
# the UI is frontend-tot; log in there with ADMIN_EMAIL / ADMIN_PASSWORD from .env
```

Uploaded images (course and specialization icons, covers, lesson pictures) go to the
S3 bucket; the database keeps their URLs. The settings come from `.env` (`S3_ENDPOINT`,
`S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_PUBLIC_URL`); without `S3_ENDPOINT` the
server runs anyway — icon uploads answer `503 storage_disabled` and covers fall back to
inline data URIs. The bucket is prepared on the first upload, so storage that starts after
the server needs no restart.

Migrations run automatically on startup (`internal/migrate`, tracked in the
`schema_migrations` table) — no manual psql step. A migration is identified by
its full file name, so never rename one that has shipped; `025` exists twice
(`025_billing`, `025_trainer_flag`) for that reason. Take the next number from
`ls migrations/ | tail -1` rather than from this file — the latest is `030`, so
the next one is `031`.

## Tests
```bash
go test ./...
```
The repository tests run against a real database: set `TEST_DATABASE_URL` (or
`DATABASE_URL`) and they apply `migrations/` themselves; without one they skip.
CI (`.github/workflows/ci.yml`) runs them on every branch and PR against a
`postgres:16` service and fails if any of them skipped. `deploy.yml` runs only
on `main` and runs no tests.

## Lab sandboxes
Two interchangeable backends behind `runner.Engine`, picked by `runner.Dispatcher`:

| Backend | When | Where |
|---|---|---|
| `VMRunner` | `FC_ENABLED=1` + an FC host (`FC_SSH_*`) — production | one Firecracker micro-VM per lesson, reached over SSH (`internal/runner/vmrunner.go`) |
| `ShellRunner` | otherwise; `SANDBOX_LOCAL=1` runs it on this machine's Docker | one container per lesson (`internal/runner/shell.go`) |

The FC host (golden rootfs, guest kernel, `gl-tap`, the `glvm` account) is set up by
hand and is **not** reproducible from this repo, so local development uses
`SANDBOX_LOCAL=1` and the images below. `SANDBOX_PRIVILEGED=1` additionally unlocks
the Docker and Kubernetes courses.

Shell labs run one container **per lesson** (`gl-s-u<user>-l<lesson>`), always
with `--network none`. Four images, picked per lesson via `tasks.sandbox_image`:

| Image | Used by | Notes |
|---|---|---|
| `golearn/sandbox` | Linux, Git, тренажёры | CLI tools baked in; offline apt repo; `systemctl` shim (`deploy/sandbox/systemctl`) |
| `golearn/sandbox-pg` | — | PostgreSQL server; unused since the SQL course was withdrawn (migration `030`), kept because rebuilding it is the expensive part of bringing that course back |
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
cmd/seed/       — Seeder: embedded JSON course content (content/), importer, lab fixtures
api/            — OpenAPI contract for the JSON API (openapi.yaml)
internal/
  api/          — JSON API /api/v1 for the Next.js frontend; after editing the spec run `go generate ./internal/api/apigen`
  auth/         — Session cookie, rate limits, ADMIN_EMAILS rules
  config/       — Environment config
  model/        — Data models
  repository/   — PostgreSQL queries
  apidocs/      — serves api/openapi.yaml and a Swagger UI for it at /docs
  catalog/      — course taxonomy, progress status rules, cover rendering
  content/      — renders stored lesson/task content (HTML or Markdown) into safe HTML
  courseio/     — native course interchange format (import/export)
  lab/          — lab rules: session keys, path jail, preview, git helpers
  migrate/      — applies migrations/ once each, tracked in schema_migrations
  obs/          — Prometheus metrics (/metrics), request ids, structured request log
  runner/       — lab sandboxes: VMRunner (Firecracker) and ShellRunner (containers)
  simulators/   — built-in simulator scenarios and their one-time seeding
  storage/      — uploaded images in S3-compatible storage
migrations/     — SQL migrations
```

## Learning Context
- Each lesson has: theory (HTML or Markdown), quiz (multiple choice), practical tasks
  (auto-checked labs in a sandbox)
- Progress is tracked in the database
- Brain/memory for this project: ~/.claude/projects/-Users-backendraz-GolandProjects-GoLearn/memory/

## Rules
- This is a LEARNING project — explain things thoroughly
- When student asks about a topic, check which lesson covers it
- Update learning_stage.md in brain after each completed lesson
- Language: Russian for explanations, English for code
