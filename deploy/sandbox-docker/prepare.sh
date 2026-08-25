#!/usr/bin/env bash
# Prepares the build context for the Docker sandbox image (run on the host,
# needs network — the lab itself never has any):
#   images/preload.tar  — images loaded into the inner daemon on first start
#   registry/           — a registry:2 storage tree so `docker pull` works offline
#   certs/              — TLS material that makes the local registry answer as Docker Hub
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "$HERE"

PRELOAD=(alpine:3.20 alpine:latest busybox:latest busybox:1.28
         nginx:alpine nginx:latest nginx:1.25-alpine
         redis:alpine redis:7-alpine python:3.12-alpine postgres:16-alpine
         golang:1.21-alpine registry:2)
# What students are asked to `docker pull` (plus the common ones).
PULLABLE=(alpine:3.20 alpine:latest busybox:latest busybox:1.28
          nginx:alpine nginx:latest nginx:1.25-alpine
          redis:alpine redis:7-alpine python:3.12-alpine postgres:16-alpine)

echo "==> pulling images"
for i in "${PRELOAD[@]}"; do docker pull -q "$i" >/dev/null; done

echo "==> exporting preload tarball"
mkdir -p images
docker save "${PRELOAD[@]}" -o images/preload.tar

echo "==> filling the offline registry"
rm -rf registry && mkdir -p registry
docker rm -f gl-prep-registry >/dev/null 2>&1 || true
docker run -d --name gl-prep-registry -p 5999:5000 \
    -v "$HERE/registry:/var/lib/registry" registry:2 >/dev/null
sleep 3
for i in "${PULLABLE[@]}"; do
    docker tag "$i" "127.0.0.1:5999/library/$i"
    docker push -q "127.0.0.1:5999/library/$i" >/dev/null
    docker rmi "127.0.0.1:5999/library/$i" >/dev/null
done
docker rm -f gl-prep-registry >/dev/null

echo "==> generating TLS material for registry-1.docker.io"
rm -rf certs && mkdir -p certs
openssl req -x509 -newkey rsa:2048 -nodes -days 3650 \
    -keyout certs/ca.key -out certs/ca.crt \
    -subj "/CN=GoLearn Sandbox CA" >/dev/null 2>&1
openssl req -newkey rsa:2048 -nodes -keyout certs/registry.key \
    -out certs/registry.csr -subj "/CN=registry-1.docker.io" >/dev/null 2>&1
cat > certs/san.cnf <<SAN
subjectAltName = DNS:registry-1.docker.io, DNS:index.docker.io, DNS:auth.docker.io, DNS:docker.io, IP:127.0.0.1
extendedKeyUsage = serverAuth
SAN
openssl x509 -req -in certs/registry.csr -CA certs/ca.crt -CAkey certs/ca.key \
    -CAcreateserial -out certs/registry.crt -days 3650 \
    -extfile certs/san.cnf >/dev/null 2>&1
rm -f certs/registry.csr certs/ca.key certs/san.cnf certs/ca.srl

echo "==> done: $(du -sh images registry certs | tr '\n' ' ')"
