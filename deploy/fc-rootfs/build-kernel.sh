#!/bin/bash
set -e
cd /out
[ -f linux-6.1.128.tar.xz ] || curl -sL https://cdn.kernel.org/pub/linux/kernel/v6.x/linux-6.1.128.tar.xz -o linux-6.1.128.tar.xz
rm -rf /build/linux-6.1.128 && mkdir -p /build && tar -C /build -xf linux-6.1.128.tar.xz
cd /build/linux-6.1.128
cp /out/vmlinux-6.1.128.config .config
echo "[$(date +%T)] enabling options..."
# critical boot drivers (Firecracker virtio-mmio) + fs + netfilter for docker/k8s
for o in VIRTIO VIRTIO_MMIO VIRTIO_BLK VIRTIO_NET VIRTIO_PCI VIRTIO_RING VIRTIO_MENU \
         BLK_DEV VIRTIO_MMIO_CMDLINE_DEVICES BLOCK EXT4_FS DEVTMPFS DEVTMPFS_MOUNT BINFMT_ELF BINFMT_SCRIPT UNIX INET \
         IP_NF_RAW IP6_NF_RAW NETFILTER_XT_TABLE_RAW NETFILTER_XT_TABLE_FILTER NETFILTER_XT_TABLE_NAT NETFILTER_XT_TABLE_MANGLE \
         NF_TABLES NF_TABLES_INET NF_TABLES_IPV4 NF_TABLES_IPV6 NFT_CT NFT_NAT NFT_MASQ NFT_REJECT NFT_COMPAT NFT_LIMIT NFT_LOG \
         VXLAN IPVLAN MACVLAN DUMMY NETFILTER_XT_MATCH_COMMENT NETFILTER_XT_MATCH_STATISTIC NETFILTER_XT_MATCH_MARK NETFILTER_XT_TARGET_MARK NETFILTER_XT_MATCH_MULTIPORT NETFILTER_XT_MATCH_RECENT NETFILTER_XT_NAT NETFILTER_XT_TARGET_MASQUERADE NETFILTER_XT_TARGET_REDIRECT NETFILTER_XT_MATCH_CONNTRACK NETFILTER_XT_MATCH_ADDRTYPE NETFILTER_XT_TARGET_DNAT IP_SET NETFILTER_XT_SET NF_CONNTRACK_MARK BRIDGE_NETFILTER; do
  scripts/config --file .config -e $o
done
make olddefconfig >/dev/null 2>&1
cp .config /out/tot.config
echo "[$(date +%T)] verify critical:"; for o in VIRTIO_BLK EXT4_FS IP_NF_RAW NF_TABLES; do grep -E "^CONFIG_$o=" .config || echo "  !! $o MISSING"; done
echo "[$(date +%T)] compiling -j$(nproc)..."
make vmlinux -j"$(nproc)" >/tmp/mk.log 2>&1 && tail -2 /tmp/mk.log || { echo BUILD_FAIL; tail -20 /tmp/mk.log; exit 1; }
cp vmlinux /out/vmlinux-6.1.128-tot
echo "[$(date +%T)] DONE size=$(ls -lh /out/vmlinux-6.1.128-tot|awk "{print \$5}")"
