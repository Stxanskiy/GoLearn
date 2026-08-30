# GoLearn sandbox image.
#
# Lab containers run with --network none, so nothing can be installed at
# runtime: every tool a lesson asks for must be baked in here. Keep this list
# aligned with what the course tasks actually use (Linux fundamentals, text
# processing, permissions, processes, networking inspection, git).
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive \
    LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    TERM=xterm-256color

RUN apt-get update && apt-get install -y --no-install-recommends \
      bash bash-completion coreutils findutils diffutils patch \
      grep sed gawk util-linux \
      vim nano less tree file bc \
      jq curl wget ca-certificates \
      git \
      tar gzip bzip2 xz-utils zip unzip \
      procps psmisc lsof strace time \
      iproute2 iputils-ping net-tools dnsutils netcat-openbsd socat \
      rsync cron at sudo openssh-client tmux \
      man-db manpages \
      python3 python3-venv zsh \
      ansible-core \
      nginx-light \
    && yes | unminimize >/dev/null 2>&1 || true \
    && rm -rf /var/lib/apt/lists/*

# Offline apt repository.
#
# Lab containers have no network, but the course teaches `apt install`. A few
# packages are downloaded at build time into a local file repo, and the online
# sources are removed, so `apt-get update && apt-get install -y htop` behaves
# exactly like the lesson describes — offline. htop is deliberately NOT
# preinstalled: installing it is the exercise. The original online sources are
# kept (disabled) in /etc/apt/sources.list.d.off so derived images can restore
# them at build time.
RUN apt-get update \
 && apt-get install -y --no-install-recommends dpkg-dev \
 && mkdir -p /var/cache/glpkg \
 && cd /var/cache/glpkg \
 && apt-get install -y --print-uris htop ncdu sl cowsay figlet \
      | grep -oE "https?://[^']+\.deb" | sort -u | xargs -r -n1 -P4 curl -sSOL \
 && dpkg-scanpackages . /dev/null > Packages 2>/dev/null \
 && gzip -kf Packages \
 && mkdir -p /etc/apt/sources.list.d.off \
 && (mv /etc/apt/sources.list.d/*.sources /etc/apt/sources.list.d.off/ 2>/dev/null || true) \
 && (mv /etc/apt/sources.list.d/*.list /etc/apt/sources.list.d.off/ 2>/dev/null || true) \
 && (mv /etc/apt/sources.list /etc/apt/sources.list.d.off/ 2>/dev/null || true) \
 && echo 'deb [trusted=yes] file:/var/cache/glpkg ./' > /etc/apt/sources.list.d/golearn-offline.list \
 && apt-get update \
 && rm -rf /var/lib/apt/lists/partial

# Containers have no systemd as PID 1; this shim implements the subset of
# systemctl the systemd lessons use, so those tasks are actually doable.
COPY systemctl /usr/local/bin/systemctl
RUN chmod +x /usr/local/bin/systemctl

# Git identity so `git commit` works immediately in git labs.
RUN git config --system user.name  "Student" \
 && git config --system user.email "student@golearn.local" \
 && git config --system init.defaultBranch main \
 && git config --system advice.detachedHead false \
 && git config --system --add safe.directory '*'

# A readable prompt plus the cwd file the exec-based checks read.
RUN printf '%s\n' \
    'export TERM=xterm-256color' \
    'export PS1="\[\e[1;32m\]student@sandbox\[\e[0m\]:\[\e[1;34m\]\w\[\e[0m\]\$ "' \
    'alias ll="ls -alF"' \
    'alias la="ls -A"' \
    'cd "$(cat /root/.gl_cwd 2>/dev/null || echo /root)" 2>/dev/null || cd /root' \
    'PROMPT_COMMAND=${PROMPT_COMMAND:+$PROMPT_COMMAND; }'"'"'pwd > /root/.gl_cwd 2>/dev/null'"'"'' \
    >> /root/.bashrc

WORKDIR /root
CMD ["sleep", "infinity"]
