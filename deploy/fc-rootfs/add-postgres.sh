#!/usr/bin/env bash
# Add PostgreSQL and its offline package repository to an existing golden rootfs.
#
# Why this exists instead of ./build-rootfs.sh:
#
# Dockerfile.rootfs already carries PostgreSQL, but the live goldens on the FC
# host predate it, and rebuilding from the recipe is not a safe way to catch up
# — the live images also carry container images that were loaded into them by
# hand (~2.4 GB under /var/lib/docker, ~573 MB of k3s airgap images), and the
# recipe does not. build-rootfs.sh refuses to install a result that is poorer
# than the live image for exactly that reason. This converges the live image
# towards the recipe for the one thing that is missing, and touches nothing
# else.
#
#   ./add-postgres.sh                      # patch /opt/fc/rootfs-docker.ext4
#   ./add-postgres.sh --check              # only report what the live image has
#   ROOTFS=/opt/fc/rootfs-k8s.ext4 ./add-postgres.sh
#
# The live image is never written to: the work happens on a copy, which lands
# as <rootfs>.new. Installing it is a separate, deliberate step — see README.md.
#
# Needs root (loop mount, chroot) and network on the host. The VM itself has
# none, which is the whole reason the packages are cached into the image.
set -euo pipefail

ROOTFS="${ROOTFS:-/opt/fc/rootfs-docker.ext4}"
NEW="$ROOTFS.new"
CHECK=0
[ "${1:-}" = "--check" ] && CHECK=1

[ -f "$ROOTFS" ] || { echo "нет $ROOTFS" >&2; exit 2; }
[ "$(id -u)" = 0 ] || { echo "нужен root (loop mount и chroot)" >&2; exit 2; }

mnt=""
cleanup() {
	[ -n "$mnt" ] || return 0
	for d in dev/pts dev sys proc; do umount "$mnt/$d" 2>/dev/null || true; done
	mountpoint -q "$mnt" && umount "$mnt"
	rmdir "$mnt" 2>/dev/null || true
}
trap cleanup EXIT

# ── what the live image has now ──
mnt=$(mktemp -d)
mount -o loop,ro "$ROOTFS" "$mnt"
echo "==> $ROOTFS"
if [ -d "$mnt/etc/postgresql" ]; then
	echo "    PostgreSQL: есть ($(ls "$mnt/etc/postgresql" | tr '\n' ' '))"
else
	echo "    PostgreSQL: НЕТ"
fi
if [ -f "$mnt/var/cache/glpkg/Packages" ]; then
	echo "    офлайн-репозиторий: есть ($(ls "$mnt/var/cache/glpkg"/*.deb 2>/dev/null | wc -l) пакетов)"
else
	echo "    офлайн-репозиторий: НЕТ"
fi
echo "    свободно в образе: $(df -h --output=avail "$mnt" | tail -1 | tr -d ' ')"
umount "$mnt"; rmdir "$mnt"; mnt=""

[ "$CHECK" = 1 ] && exit 0

# ── copy, then work on the copy ──
avail=$(df --output=avail -k "$(dirname "$ROOTFS")" | tail -1)
need=$(( $(stat -c %s "$ROOTFS") / 1024 ))
[ "$avail" -gt "$need" ] || { echo "мало места: нужно ~$((need/1024/1024)) ГБ" >&2; exit 1; }

echo "==> копирую в $NEW"
rm -f "$NEW"
cp --sparse=always "$ROOTFS" "$NEW"

mnt=$(mktemp -d)
mount -o loop "$NEW" "$mnt"
for d in proc sys dev dev/pts; do mount --bind "/$d" "$mnt/$d"; done
cp /etc/resolv.conf "$mnt/etc/resolv.conf"

# dpkg runs invoke-rc.d during the install; inside a chroot there is no init to
# talk to. 101 tells it not to try. Removed again below so the golden ends up
# the way the Dockerfile leaves it — with no policy of its own.
cat > "$mnt/usr/sbin/policy-rc.d" <<'EOF'
#!/bin/sh
exit 101
EOF
chmod +x "$mnt/usr/sbin/policy-rc.d"

echo "==> ставлю PostgreSQL и собираю офлайн-репозиторий"
# The same steps as Dockerfile.rootfs, in the same order. Online sources are
# put back for the duration: the image keeps them in sources.list.d.off because
# a VM with no network only gets errors out of them.
chroot "$mnt" /bin/bash -eux <<'EOF'
export DEBIAN_FRONTEND=noninteractive
if [ -d /etc/apt/sources.list.d.off ]; then
	cp -a /etc/apt/sources.list.d.off/. /etc/apt/sources.list.d/ 2>/dev/null || true
	[ -f /etc/apt/sources.list.d/sources.list ] && mv /etc/apt/sources.list.d/sources.list /etc/apt/sources.list
fi
rm -f /etc/apt/sources.list.d/golearn-offline.list
apt-get update
apt-get install -y --no-install-recommends dpkg-dev
mkdir -p /var/cache/glpkg && cd /var/cache/glpkg
# Every .deb `apt install postgresql` would fetch on a clean machine, so the
# student can install it again with no network.
apt-get install -y --print-uris postgresql \
	| grep -oE "https?://[^']+\.deb" | sort -u | xargs -r -n1 -P4 curl -sSOL
dpkg-scanpackages . /dev/null > Packages 2>/dev/null
gzip -kf Packages
apt-get install -y --no-install-recommends postgresql postgresql-client
# Every lab VM boots from this image; only the lessons that need a server start
# one (glpg start).
systemctl disable postgresql 2>/dev/null || true
rm -f /etc/systemd/system/multi-user.target.wants/postgresql.service
mkdir -p /etc/apt/sources.list.d.off
mv /etc/apt/sources.list.d/*.sources /etc/apt/sources.list.d.off/ 2>/dev/null || true
mv /etc/apt/sources.list /etc/apt/sources.list.d.off/ 2>/dev/null || true
echo 'deb [trusted=yes] file:/var/cache/glpkg ./' > /etc/apt/sources.list.d/golearn-offline.list
apt-get update
apt-get clean
rm -rf /var/lib/apt/lists/partial
EOF

rm -f "$mnt/usr/sbin/policy-rc.d" "$mnt/etc/resolv.conf"

# ── prove it before anyone installs it ──
echo "==> проверяю результат"
fail=0
chroot "$mnt" dpkg-query -W -f='${Status}\n' postgresql-16 2>/dev/null | grep -q 'install ok installed' \
	|| { echo "    !! postgresql-16 не установлен"; fail=1; }
[ -d "$mnt/etc/postgresql" ] || { echo "    !! нет /etc/postgresql"; fail=1; }
[ -f "$mnt/var/cache/glpkg/Packages" ] || { echo "    !! нет офлайн-репозитория"; fail=1; }
[ -e "$mnt/etc/systemd/system/multi-user.target.wants/postgresql.service" ] \
	&& { echo "    !! postgresql остался включённым — он будет стартовать в каждой лабе"; fail=1; }
# The things a rebuild loses and this must not: they were loaded by hand.
for d in /var/lib/docker /usr/share/zsh-plugins /root/.local/share/nvim/lazy; do
	a=$(du -sm "$mnt$d" 2>/dev/null | cut -f1); a=${a:-0}
	printf "    %-40s %s МБ\n" "$d" "$a"
done
echo "    свободно в образе: $(df -h --output=avail "$mnt" | tail -1 | tr -d ' ')"

if [ "$fail" = 1 ]; then
	echo
	echo "НЕ СТАВЬ ЭТОТ ОБРАЗ. $NEW оставлен для разбора." >&2
	exit 1
fi

echo
echo "==> готово: $NEW"
echo "    Живой образ не тронут. Установка — отдельный шаг:"
echo
echo "      cd $(dirname "$ROOTFS")"
echo "      cp $(basename "$ROOTFS") $(basename "$ROOTFS").bak     # откат = вернуть это"
echo "      mv $(basename "$NEW") $(basename "$ROOTFS")"
echo "      chown glvm:glvm $(basename "$ROOTFS")"
echo
echo "    Потом слей тёплый пул — он держит VM со старого образа."
