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

## When the live image is older than this recipe

A lesson's setup script runs against whatever the golden contains. Add
something here, do not rebuild, and every lab whose setup depends on it fails —
as one opaque line, "песочница не запустилась", because the boot treats a
failed setup as a failed bring-up.

This happened with PostgreSQL: the server and its offline package repository
were added to `Dockerfile.rootfs` above, the live `rootfs-docker.ext4` was not
rebuilt, and every lab of the PostgreSQL course died in setup — `glpg start`
answers "PostgreSQL не установлен" and `set -e` ends the script.

How to tell this apart from a VM that will not boot, without reading the pod's
log:

```bash
curl -s https://learn.prod-factory.ru/metrics | grep vm_start_failures
# reason="setup"        the VM booted; a lesson's setup script failed  <- this case
# reason="boot"         Firecracker started and sshd never came up
# reason="boot-timeout" the VM did not answer in time
# reason="host"         the SSH to the FC host itself failed
```

Opening the lab as an **admin** prints the host's own output under the error,
which names the failing command. A student still sees one line.

Nothing in the backend can substitute for the image: a lab that needs a package
the golden does not carry cannot be made to work from the application side.

But a full rebuild is the wrong way to catch up. The live goldens also hold
container images that were loaded into them by hand — ~2.4 GB under
`/var/lib/docker`, ~573 MB of k3s airgap images — and this recipe does not
produce them, which is why `build-rootfs.sh` refuses to install a result poorer
than the live image. Adding the one missing thing is safer than rebuilding
everything around it:

```bash
sudo ./add-postgres.sh --check     # what the live image has now
sudo ./add-postgres.sh             # -> rootfs-docker.ext4.new, live file untouched
```

It works on a copy, mirrors the PostgreSQL block of `Dockerfile.rootfs` step
for step, and refuses to hand over a result where the package did not install,
the offline repository is missing, or the unit was left enabled — a server
starting in every lab VM is not what the course wants. Installing the result is
the same deliberate step as above.

## The kernel

`build-kernel.sh` builds the guest kernel (`tot.config`). It is separate because
it changes far less often — and because of a known limitation recorded in
`internal/runner/vmrunner.go`: the guest boots with `acpi=off`, which costs a
full host core per idle VM. Fixing that needs a kernel that boots with ACPI
enabled, not a boot argument.
