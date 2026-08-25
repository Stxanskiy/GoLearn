# Sandbox image for the SQL course: the base lab image plus a real PostgreSQL
# server, so the SQL lessons run against an actual database inside the isolated
# (network-less) lab container.
FROM golearn/sandbox:latest

ENV DEBIAN_FRONTEND=noninteractive

# The base image replaced apt's sources with the offline repo; restore the
# online ones for this build only, then drop them again.
RUN cp /etc/apt/sources.list.d.off/* /etc/apt/sources.list.d/ \
 && apt-get update \
 && apt-get install -y --no-install-recommends postgresql postgresql-client \
 && (cd /etc/apt/sources.list.d && for f in $(ls /etc/apt/sources.list.d.off); do rm -f "$f"; done) \
 && apt-get update \
 && rm -rf /var/lib/apt/lists/partial

# Start the cluster and give root a superuser role, so the student can just run
# `psql` as the terminal user the way the lessons assume.
COPY pg-start /usr/local/bin/pg-start
RUN chmod +x /usr/local/bin/pg-start

CMD ["sleep", "infinity"]
