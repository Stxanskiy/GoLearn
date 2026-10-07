# Sandbox image for the PostgreSQL courses: the base lab image plus a real
# PostgreSQL server, so the lessons run against an actual database inside the
# isolated (network-less) lab container.
FROM golearn/sandbox:latest

ENV DEBIAN_FRONTEND=noninteractive

# The base image replaced apt's sources with the offline repo; restore the
# online ones for this build only, then drop them again.
#
# The first lab of the course is installing PostgreSQL itself, and the lab
# container has no network. So before the server goes in, every package that
# `apt install postgresql` would fetch on a clean machine — recommends included,
# because the student's apt installs them — is downloaded into the base image's
# offline repo. The lab's setup removes the server and the student puts it back
# from there, exactly as on a real host.
RUN cp /etc/apt/sources.list.d.off/* /etc/apt/sources.list.d/ \
 && apt-get update \
 && cd /var/cache/glpkg \
 && apt-get install -y --print-uris postgresql \
      | grep -oE "https?://[^']+\.deb" | sort -u | xargs -r -n1 -P4 curl -sSOL \
 && dpkg-scanpackages . /dev/null > Packages 2>/dev/null \
 && gzip -kf Packages \
 && apt-get install -y --no-install-recommends postgresql postgresql-client \
 && (cd /etc/apt/sources.list.d && for f in $(ls /etc/apt/sources.list.d.off); do rm -f "$f"; done) \
 && apt-get update \
 && rm -rf /var/lib/apt/lists/partial

# Start the cluster and give root a superuser role, so the student can just run
# `psql` as the terminal user the way the lessons assume. The course's lab
# setups carry the same logic themselves (glpg start, cmd/seed/pgcourse), so
# they do not depend on this file being in the production root filesystem.
COPY pg-start /usr/local/bin/pg-start
RUN chmod +x /usr/local/bin/pg-start

CMD ["sleep", "infinity"]
