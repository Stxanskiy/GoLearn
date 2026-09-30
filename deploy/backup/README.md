# Backups

One PostgreSQL database holds everything: accounts, progress, quiz attempts,
submissions, billing — and the uploaded images too, because object storage is
not configured in production and covers fall back to inline data URIs. There is
no second copy of any of it anywhere.

Before this, backups were taken by hand. The newest one on the host was from
26 August, five weeks stale, and nothing rotated or checked them.

## Install on the FC host

```bash
sudo install -d -o berg -m 750 /var/backups/golearn
sudo install -m 755 golearn-backup.sh /usr/local/bin/golearn-backup.sh
sudo install -m 644 golearn-backup.service golearn-backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now golearn-backup.timer
systemctl list-timers golearn-backup.timer
```

## What it does

Nightly at 04:20 it dumps the database out of the `golearn-db` pod, gzips it to
`/var/backups/golearn/golearn-<date>.sql.gz`, and refuses to keep the result
unless it passes three checks: the archive is intact, it is larger than the
schema alone, and it actually contains `CREATE TABLE public.lessons`. A dump
that fails any of them is deleted and the unit fails loudly, because a backup
job that fails quietly buys confidence it has not earned.

Rotation keeps 14 daily, 8 weekly and 12 monthly. At the current database size
(~16 MB, ~3 MB compressed) that is well under 100 MB.

## Checking that it works

A backup nobody has restored is not a backup:

```bash
golearn-backup.sh --verify
```

This restores the newest dump into a **throwaway PostgreSQL container** and
compares row counts for `users`, `lessons`, `tasks` and `progress` against
production. The failure worth catching is a restore that "succeeds" into an
empty database, and comparing counts is what catches it.

It deliberately does not restore into the production instance. The first version
did, and a half-finished restore left a backend stuck in `startup waiting` that
blocked `DROP DATABASE` and queued every later attempt behind it. Verifying a
backup must not be able to disturb the thing it is protecting.

Worth running by hand after any schema migration that changes those tables.

## Restoring

```bash
gunzip -c /var/backups/golearn/golearn-<date>.sql.gz \
  | sudo k3s kubectl -n golearn exec -i deploy/golearn-db -- psql -U golearn -d golearn
```

The dump is taken with `--clean --if-exists`, so it replays over an existing
database. Scale the app to zero first (`kubectl -n golearn scale deploy/golearn
--replicas=0`) — restoring under a live backend leaves it holding rows that no
longer exist.

## What is not backed up

The lab VM images. They are rebuildable from `deploy/fc-rootfs/`, which is the
point of having that recipe in the repository.
