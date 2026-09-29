# The lab VM image

Everything a student sees inside a lab — the shell, the prompt, the editor, the
tools — comes from a golden rootfs that Firecracker copies per session. This
directory is the recipe for it.

It used to live only on the FC host, built by hand. That meant the environment
could not be rebuilt from this repository, a change to it was an edit to a live
5 GB file, and nobody could say what was actually in there. The files here were
recovered from `berg:/home/berg/fc-build/` and are now the source of truth.

## What is in the image

Ubuntu 24.04 with systemd, the course CLI tools (docker, kubectl, helm, k3s,
ansible, git, jq, tmux, …), sshd accepting the VM key only, and a static IP read
from the kernel command line (`gl.ip=`).

The shell is **zsh** with a two-line prompt, git branch, completion menus,
`zsh-autosuggestions` and `zsh-syntax-highlighting` — see `rootfs-extra/zshrc`.

The editor is **Neovim** with a small plugin set: tokyonight, treesitter
(yaml, bash, dockerfile, json, lua, markdown), lualine, nvim-tree, gitsigns,
indent-blankline and which-key — see `rootfs-extra/nvim/init.lua`. `vim` and
`vi` are aliased to it.

**The VM has no network.** Anything a plugin or tool needs must be fetched
during the build: lazy.nvim is cloned and `Lazy! sync` plus `TSUpdateSync` run
headless in the Dockerfile, and lazy's update checker is disabled so opening a
file never waits on a request that cannot succeed. The same rule applies to
anything added later.

## Building

Needs Docker, network, and root for the loop mount. Run it **on the FC host**:
the image is x86_64, and one built on an arm64 Mac will not boot there.

```bash
# the public half of the VM key, whose private half lives on the FC host
mkdir -p imgseed && cp /opt/fc/appkey.pub imgseed/authorized_keys

./build-rootfs.sh docker   # -> rootfs-docker.ext4.new
./build-rootfs.sh k8s      # -> rootfs-k8s.ext4.new
```

The script never overwrites a live image; it writes `.new` beside it.

## Installing

A golden is copied at boot, so replacing it affects **new** sessions only —
sessions already running keep the copy they booted from.

```bash
cd /opt/fc
sudo cp rootfs-docker.ext4 rootfs-docker.ext4.bak    # rollback = move this back
sudo mv rootfs-docker.ext4.new rootfs-docker.ext4
sudo chown glvm:glvm rootfs-docker.ext4
```

Then open one lesson and check the shell, `nvim`, and the tools it needs before
letting students onto it. The warm pool pre-boots VMs from the old image, so
drain it (or restart the service) rather than assuming the next student gets
the new one.

Keep one `.bak` around. Two goldens at 5 GB each plus a backup is 15 GB; the
host has ~320 GB free, but `rootfs-docker.ext4.pre-nodocker` from an earlier
hand-edit is still sitting there and can go once this recipe is proven.

## The kernel

`build-kernel.sh` builds the guest kernel (`tot.config`). It is separate because
it changes far less often — and because of a known limitation recorded in
`internal/runner/vmrunner.go`: the guest boots with `acpi=off`, which costs a
full host core per idle VM. Fixing that needs a kernel that boots with ACPI
enabled, not a boot argument.
