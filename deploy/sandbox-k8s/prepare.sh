#!/usr/bin/env bash
# Prepares the build context for the Kubernetes sandbox image (run on the host,
# needs network — the lab itself never has any):
#   bin/k3s, bin/helm, bin/registry     — binaries
#   bin/k3s-airgap-images.tar           — the cluster's own images
#   bin/app-images.tar, registry/       — application images + offline Hub mirror
#
# Run deploy/sandbox-docker/prepare.sh first: this script reuses its output.
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
DOCKER_CTX="$HERE/../sandbox-docker"
cd "$HERE"
mkdir -p bin

K3S_VERSION="${K3S_VERSION:-v1.31.5+k3s1}"
HELM_VERSION="${HELM_VERSION:-v3.16.3}"
case "$(uname -m)" in
    arm64|aarch64) ARCH=arm64; K3S_SUFFIX=-arm64 ;;
    x86_64|amd64)  ARCH=amd64; K3S_SUFFIX=       ;;
    *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac
K3S_ENC=$(python3 -c "import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1]))" "$K3S_VERSION")

echo "==> k3s $K3S_VERSION ($ARCH)"
curl -sL -o bin/k3s "https://github.com/k3s-io/k3s/releases/download/${K3S_ENC}/k3s${K3S_SUFFIX}"
curl -sL -o bin/k3s-airgap-images.tar \
    "https://github.com/k3s-io/k3s/releases/download/${K3S_ENC}/k3s-airgap-images-${ARCH}.tar"

echo "==> helm $HELM_VERSION"
tmp=$(mktemp -d)
curl -sL "https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz" | tar -xz -C "$tmp"
mv "$tmp/linux-${ARCH}/helm" bin/helm
rm -rf "$tmp"

echo "==> registry binary (from the registry:2 image)"
cid=$(docker create registry:2)
docker cp "$cid:/bin/registry" bin/registry >/dev/null
docker rm "$cid" >/dev/null

echo "==> application images + offline registry (from the Docker sandbox context)"
if [ ! -f "$DOCKER_CTX/images/preload.tar" ]; then
    echo "run deploy/sandbox-docker/prepare.sh first" >&2
    exit 1
fi
cp "$DOCKER_CTX/images/preload.tar" bin/app-images.tar
rm -rf registry && cp -R "$DOCKER_CTX/registry" registry

chmod +x bin/k3s bin/helm bin/registry
echo "==> done: $(du -sh bin registry | tr '\n' ' ')"
