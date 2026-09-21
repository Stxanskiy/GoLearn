#!/usr/bin/env bash
# Prepares the build context for the Kubernetes sandbox image (run on the host,
# needs network — the lab itself never has any):
#   bin/k3s, bin/helm, bin/helmfile,
#   bin/registry                        — binaries
#   bin/k3s-airgap-images.tar           — the cluster's own images
#   bin/app-images.tar, registry/       — application images + offline Hub mirror
#   helm-mirror/                        — offline Helm repo (the `bitnami` repo
#                                         the repos lab adds at localhost:8879)
#
# Run deploy/sandbox-docker/prepare.sh first: this script reuses its output.
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
DOCKER_CTX="$HERE/../sandbox-docker"
cd "$HERE"
mkdir -p bin

K3S_VERSION="${K3S_VERSION:-v1.31.5+k3s1}"
HELM_VERSION="${HELM_VERSION:-v3.16.3}"
HELMFILE_VERSION="${HELMFILE_VERSION:-0.169.1}"
BITNAMI_NGINX_VERSION="${BITNAMI_NGINX_VERSION:-18.2.5}"
case "$(uname -m)" in
    arm64|aarch64) ARCH=arm64; K3S_SUFFIX=-arm64 ;;
    x86_64|amd64)  ARCH=amd64; K3S_SUFFIX=       ;;
    *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac
K3S_ENC=$(python3 -c "import urllib.parse,sys;print(urllib.parse.quote(sys.argv[1]))" "$K3S_VERSION")

echo "==> k3s $K3S_VERSION ($ARCH)"
curl -fsSL --retry 3 --retry-delay 2 -o bin/k3s "https://github.com/k3s-io/k3s/releases/download/${K3S_ENC}/k3s${K3S_SUFFIX}"
curl -fsSL --retry 3 --retry-delay 2 -o bin/k3s-airgap-images.tar \
    "https://github.com/k3s-io/k3s/releases/download/${K3S_ENC}/k3s-airgap-images-${ARCH}.tar"

echo "==> helm $HELM_VERSION"
tmp=$(mktemp -d)
curl -fsSL --retry 3 --retry-delay 2 "https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz" | tar -xz -C "$tmp"
mv "$tmp/linux-${ARCH}/helm" bin/helm
rm -rf "$tmp"

echo "==> helmfile $HELMFILE_VERSION"
tmp=$(mktemp -d)
curl -fsSL --retry 3 --retry-delay 2 "https://github.com/helmfile/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION}_linux_${ARCH}.tar.gz" | tar -xz -C "$tmp"
mv "$tmp/helmfile" bin/helmfile
rm -rf "$tmp"

# Offline Helm repository served at localhost:8879 inside the sandbox. The repos
# lab adds it as `bitnami` and pulls the nginx chart from it; without network the
# chart and its index have to be vendored here.
echo "==> offline helm repo (bitnami/nginx $BITNAMI_NGINX_VERSION)"
rm -rf helm-mirror && mkdir -p helm-mirror
curl -fsSL --retry 3 --retry-delay 2 -o "helm-mirror/nginx-${BITNAMI_NGINX_VERSION}.tgz" \
    "https://charts.bitnami.com/bitnami/nginx-${BITNAMI_NGINX_VERSION}.tgz"
# The index is generated here rather than with `helm repo index`, because bin/helm
# is a Linux binary and this script also runs on a macOS build host.
python3 - "helm-mirror" "nginx-${BITNAMI_NGINX_VERSION}.tgz" <<'PY'
# Written by hand rather than with `helm repo index`, because bin/helm is a Linux
# binary and this script also runs on a macOS build host. Only the handful of
# top-level scalars helm needs are read out of Chart.yaml — no PyYAML required.
import hashlib, io, os, sys, tarfile
root, fname = sys.argv[1], sys.argv[2]
blob = open(os.path.join(root, fname), "rb").read()
with tarfile.open(fileobj=io.BytesIO(blob)) as t:
    member = next(m for m in t.getmembers() if m.name.count("/") == 1
                  and m.name.endswith("/Chart.yaml"))
    text = t.extractfile(member).read().decode("utf-8")
meta = {}
for line in text.splitlines():
    m = __import__("re").match(r"^([A-Za-z][A-Za-z0-9_]*):[ \t]*(.+?)[ \t]*$", line)
    if m:
        meta.setdefault(m.group(1), m.group(2).strip().strip("\"'"))

def scalar(v):
    return '"' + str(v).replace("\\", "\\\\").replace('"', '\\"') + '"'

fields = [("apiVersion", meta.get("apiVersion", "v2")), ("name", meta["name"]),
          ("version", meta["version"])]
for k in ("appVersion", "description", "type"):
    if k in meta:
        fields.append((k, meta[k]))
fields.append(("digest", hashlib.sha256(blob).hexdigest()))
fields.append(("created", "2024-01-01T00:00:00Z"))

out = ["apiVersion: v1", "entries:", "  %s:" % meta["name"]]
first = True
for k, v in fields:
    out.append("%s%s: %s" % ("  - " if first else "    ", k, scalar(v)))
    first = False
out += ["    urls:", "    - %s" % scalar("http://localhost:8879/" + fname),
        'generated: "2024-01-01T00:00:00Z"', ""]
open(os.path.join(root, "index.yaml"), "w").write("\n".join(out))
print("   index.yaml:", meta["name"], meta["version"])
PY

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

chmod +x bin/k3s bin/helm bin/helmfile bin/registry
echo "==> done: $(du -sh bin registry helm-mirror | tr '\n' ' ')"
