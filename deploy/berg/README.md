# Deploying GoLearn on Berg

CD is automatic: a push to `main` triggers `.github/workflows/deploy.yml`, which
runs on the self-hosted runner **on the server**. It syncs the checkout into
`/opt/golearn/src`, builds the image, applies migrations, seeds the content and
restarts the app. No inbound SSH and no stored key are involved.

The app itself needs nothing else. The **labs**, however, run inside a separate
sandbox VM reached over SSH (`SANDBOX_SSH_*` in `docker-compose.yml`), and that
VM needs a one-time preparation.

## One-time: build the lab image on the sandbox VM

Every shell lab now runs in `golearn/sandbox:latest` instead of a bare
`ubuntu:24.04`: the courses need the CLI tools, the offline apt repo and the
`systemctl` shim that image provides, and the lesson fixtures assume them.
Until the image exists on the sandbox VM, labs fail with a message naming it.

On the sandbox VM (needs network **only while building**):

```bash
git clone https://github.com/Stxanskiy/GoLearn.git /opt/golearn-src   # or rsync deploy/sandbox
cd /opt/golearn-src
docker build -t golearn/sandbox:latest -f deploy/sandbox/Dockerfile deploy/sandbox
docker tag  golearn/sandbox:latest golearn/git:latest                 # git trainer
```

Roughly 650 MB and a few minutes. Rebuild it whenever `deploy/sandbox/`
changes — the lessons' auto-checks depend on what is installed there.

### SQL course (optional)

The SQL lessons run in an image with a PostgreSQL server:

```bash
docker build -t golearn/sandbox-pg:latest -f deploy/sandbox-pg/Dockerfile deploy/sandbox-pg
```

Without it only the SQL course is affected; everything else keeps working.

## Docker and Kubernetes courses are disabled here

Their labs need a container runtime inside the sandbox, which means
`--privileged` — effectively root on the sandbox VM for anyone who can open a
lab. On this server they are off (`SANDBOX_PRIVILEGED: "0"`), and opening such a
lab shows an explanatory message instead of starting a container.

To enable them (only on a host where that risk is acceptable, ≥ 8 GB RAM and
≥ 40 GB free disk):

1. On the sandbox VM, build the images — see `deploy/sandbox-docker/prepare.sh`
   and `deploy/sandbox-k8s/prepare.sh`, then `docker build` each context.
   Note they are built for the VM's own architecture; images built on an Apple
   Silicon machine (arm64) will not run on an x86_64 server.
2. Set `SANDBOX_PRIVILEGED: "1"` in `docker-compose.yml` and redeploy.

## After the first deploy

The database gets its schema and content automatically (migrations run on
startup, the seeder loads the courses). If the `users` table is empty, the
server creates the first admin from `ADMIN_EMAIL` / `ADMIN_PASSWORD`
(defaults `admin@golearn.local` / `golearn123`) — **change that password after
the first login**, or set both variables in `docker-compose.yml` before
deploying.
