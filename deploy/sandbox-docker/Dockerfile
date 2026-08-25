# Sandbox image for the Docker courses: the base lab image plus a full Docker
# Engine (docker-in-docker). Lab containers using this image are started with
# --privileged, which dockerd requires.
#
# There is no network at runtime, so the images the lessons use are baked in as
# a tarball and loaded into the inner daemon on first start.
FROM golearn/sandbox:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN cp /etc/apt/sources.list.d.off/* /etc/apt/sources.list.d/ \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
      docker.io docker-compose-v2 iptables uidmap fuse-overlayfs \
 && (cd /etc/apt/sources.list.d && for f in $(ls /etc/apt/sources.list.d.off); do rm -f "$f"; done) \
 && apt-get update \
 && rm -rf /var/lib/apt/lists/partial

# Images the Docker lessons use, exported from the host at build time
# (deploy/sandbox-docker/prepare.sh).
COPY images/preload.tar /var/lib/glimages/preload.tar

# Offline stand-in for Docker Hub: a registry:2 storage tree plus a certificate
# issued for registry-1.docker.io. With the CA installed and /etc/hosts pointing
# the Hub names at localhost, `docker pull redis:alpine` works without network.
COPY registry /var/lib/glregistry
COPY certs /etc/glcerts
RUN for host in registry-1.docker.io index.docker.io auth.docker.io docker.io; do         mkdir -p "/etc/docker/certs.d/$host" &&         cp /etc/glcerts/ca.crt "/etc/docker/certs.d/$host/ca.crt";     done  && cp /etc/glcerts/ca.crt /usr/local/share/ca-certificates/golearn-sandbox.crt  && update-ca-certificates >/dev/null 2>&1 || true
COPY docker-start /usr/local/bin/docker-start
RUN chmod +x /usr/local/bin/docker-start

CMD ["sleep", "infinity"]
