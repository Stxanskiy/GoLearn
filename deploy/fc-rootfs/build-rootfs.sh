#!/usr/bin/env bash
# Build a golden rootfs image for the Firecracker lab VMs.
#
# This step used to be done by hand on the FC host, which is why the running
# golden could not be reproduced from this repository and why adding anything to
# the lab environment meant editing a 5 GB file in place. Run this instead.
#
#   ./build-rootfs.sh docker      -> rootfs-docker.ext4   (Linux, Git, SQL, Docker)
#   ./build-rootfs.sh k8s         -> rootfs-k8s.ext4      (the above + k3s enabled)
#
# It never touches the live image: the result lands next to it as
# rootfs-<flavour>.ext4.new, and installing it is a separate, deliberate step
# (see README.md). Needs Docker, root for the loop mount, and network.
set -euo pipefail

FLAVOUR="${1:-docker}"
SIZE="${ROOTFS_SIZE:-5G}"
HERE="$(cd "$(dirname "$0")" && pwd)"
OUT="${OUT_DIR:-$HERE}/rootfs-${FLAVOUR}.ext4.new"

case "$FLAVOUR" in
  docker|k8s) ;;
  *) echo "неизвестный вариант: $FLAVOUR (нужен docker или k8s)" >&2; exit 2 ;;
esac

# imgseed/authorized_keys is the VM key's public half. It is not in the
# repository on purpose — the private half lives only on the FC host — so the
# build expects it to be placed here first.
if [ ! -f "$HERE/imgseed/authorized_keys" ]; then
  echo "нет $HERE/imgseed/authorized_keys — положи туда публичную половину ключа VM" >&2
  exit 2
fi

echo "==> собираю образ (вариант: $FLAVOUR)"
docker build -f "$HERE/Dockerfile.rootfs" -t "tot-rootfs:$FLAVOUR" "$HERE"

echo "==> выкладываю файловую систему в $OUT"
cid=$(docker create "tot-rootfs:$FLAVOUR")
cleanup() { docker rm -f "$cid" >/dev/null 2>&1 || true; mountpoint -q "$mnt" && sudo umount "$mnt"; rmdir "$mnt" 2>/dev/null || true; }
mnt=$(mktemp -d)
trap cleanup EXIT

rm -f "$OUT"
truncate -s "$SIZE" "$OUT"
mkfs.ext4 -q -F "$OUT"
sudo mount -o loop "$OUT" "$mnt"
docker export "$cid" | sudo tar -x -C "$mnt"

# k3s only runs in the Kubernetes VMs; leaving it enabled everywhere would cost
# every Linux lesson a cluster it never uses.
if [ "$FLAVOUR" = "k8s" ]; then
  sudo cp "$HERE/rootfs-extra/k3s.service" "$mnt/etc/systemd/system/k3s.service"
  sudo cp "$HERE/rootfs-extra/k8s-start"   "$mnt/usr/local/bin/k8s-start"
  sudo chmod +x "$mnt/usr/local/bin/k8s-start"
  sudo ln -sf /etc/systemd/system/k3s.service "$mnt/etc/systemd/system/multi-user.target.wants/k3s.service"
fi

sudo umount "$mnt"
trap - EXIT
cleanup

echo "==> готово: $OUT"

# Сравнение с живым образом. Этот рецепт был восстановлен с хоста и не знал про
# то, что добавляли руками: предзагруженные образы контейнеров (573 МБ для k3s,
# 2,4 ГБ для Docker). Пересборка молча их потеряла, и в проде перестали
# запускаться Pod'ы и контейнеры — лабораторные выглядели «зависшими».
LIVE="/opt/fc/rootfs-$FLAVOUR.ext4"
if [ -f "$LIVE" ]; then
  mnt_new=$(mktemp -d); mnt_live=$(mktemp -d)
  sudo mount -o loop,ro "$OUT" "$mnt_new"; sudo mount -o loop,ro "$LIVE" "$mnt_live"
  echo "==> сравнение с живым образом:"
  drift=0
  for d in /var/lib/docker /var/lib/rancher/k3s/agent/images /usr/share/zsh-plugins /root/.local/share/nvim/lazy; do
    a=$(sudo du -sm "$mnt_new$d"  2>/dev/null | cut -f1); a=${a:-0}
    b=$(sudo du -sm "$mnt_live$d" 2>/dev/null | cut -f1); b=${b:-0}
    [ "$a" = 0 ] && [ "$b" = 0 ] && continue
    mark=""
    [ "$a" -lt "$((b * 80 / 100))" ] && { mark="  <-- НОВЫЙ ОБРАЗ БЕДНЕЕ ЖИВОГО"; drift=1; }
    printf "    %-44s новый %5s МБ, живой %5s МБ%s\n" "$d" "$a" "$b" "$mark"
  done
  sudo umount "$mnt_new"; sudo umount "$mnt_live"; rmdir "$mnt_new" "$mnt_live"
  if [ "$drift" = 1 ]; then
    echo
    echo "    НЕ СТАВЬ ЭТОТ ОБРАЗ, пока не разберёшься: живой содержит то, чего нет в"
    echo "    рецепте. Перенеси недостающее и добавь шаг в Dockerfile.rootfs."
    exit 1
  fi
fi

echo "    проверь его на одной VM, и только потом ставь на место живого (README.md)"
