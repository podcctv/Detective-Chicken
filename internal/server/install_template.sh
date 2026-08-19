#!/bin/sh
set -eu

SERVER_URL='{{.ServerURL}}'
ENROLL_TOKEN='{{.Token}}'
REQUESTED_OS='{{.OSFamily}}'
REQUESTED_PLATFORM='{{.Platform}}'
REQUESTED_ARCH='{{.Arch}}'
AGENT=/usr/local/bin/detective-chicken-agent

if [ "$(id -u)" -ne 0 ]; then
  echo "Please run this installer as root." >&2
  exit 1
fi

detect_os() {
  [ "$REQUESTED_OS" != auto ] && { echo "$REQUESTED_OS"; return; }
  if [ -r /etc/os-release ]; then
    . /etc/os-release
    case "${ID:-}" in
      alpine) echo alpine ;;
      debian|ubuntu|linuxmint|proxmox) echo debian ;;
      rhel|centos|fedora|rocky|almalinux|ol) echo rhel ;;
      arch|manjaro) echo arch ;;
      *) echo "${ID:-unknown}" ;;
    esac
  else
    echo unknown
  fi
}

detect_arch() {
  [ "$REQUESTED_ARCH" != auto ] && { echo "$REQUESTED_ARCH"; return; }
  case "$(uname -m)" in
    x86_64|amd64) echo amd64 ;;
    aarch64|arm64) echo arm64 ;;
    armv7l|armv7) echo armv7 ;;
    *) echo unsupported ;;
  esac
}

detect_platform() {
  [ "$REQUESTED_PLATFORM" != auto ] && { echo "$REQUESTED_PLATFORM"; return; }
  command -v pveversion >/dev/null 2>&1 && { echo pve; return; }
  [ -f /.dockerenv ] && { echo docker; return; }
  grep -qa podman /proc/1/cgroup 2>/dev/null && { echo podman; return; }
  if command -v systemd-detect-virt >/dev/null 2>&1; then
    case "$(systemd-detect-virt 2>/dev/null || true)" in
      lxc) echo lxc; return ;;
      systemd-nspawn) echo incus; return ;;
      none|'') ;;
      *) echo baremetal; return ;;
    esac
  fi
  [ -f /run/.containerenv ] && { echo podman; return; }
  echo baremetal
}

ensure_fetcher() {
  command -v curl >/dev/null 2>&1 && return
  command -v wget >/dev/null 2>&1 && return
  case "$OS_FAMILY" in
    alpine) apk add --no-cache ca-certificates curl ;;
    debian) apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl ;;
    rhel) (dnf install -y ca-certificates curl || yum install -y ca-certificates curl) ;;
    arch) pacman -Sy --noconfirm ca-certificates curl ;;
    *) echo "Install curl or wget, then run this script again." >&2; exit 1 ;;
  esac
}

fetch() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL --proto '=https' --tlsv1.2 "$1" -o "$2"
  else
    wget -qO "$2" "$1"
  fi
}

OS_FAMILY=$(detect_os)
ARCH=$(detect_arch)
PLATFORM=$(detect_platform)
[ "$ARCH" != unsupported ] || { echo "Unsupported architecture: $(uname -m)" >&2; exit 1; }
ensure_fetcher

echo "Installing Detective Chicken: os=$OS_FAMILY platform=$PLATFORM arch=$ARCH"
install -d -m 0700 /etc/detective-chicken
tmp_agent=$(mktemp)
trap 'rm -f "$tmp_agent"' EXIT
fetch "$SERVER_URL/api/v1/downloads/agent/$ARCH" "$tmp_agent"
install -m 0755 "$tmp_agent" "$AGENT"
"$AGENT" --server "$SERVER_URL" --token "$ENROLL_TOKEN" enroll

install_systemd() {
  cat >/etc/systemd/system/detective-chicken-heartbeat.service <<'UNIT'
[Unit]
Description=Detective Chicken heartbeat
After=network-online.target
[Service]
Type=oneshot
ExecStart=/usr/local/bin/detective-chicken-agent heartbeat
UNIT
  cat >/etc/systemd/system/detective-chicken-heartbeat.timer <<'UNIT'
[Unit]
Description=Detective Chicken heartbeat timer
[Timer]
OnBootSec=1m
OnUnitActiveSec=2m
RandomizedDelaySec=30s
Persistent=true
[Install]
WantedBy=timers.target
UNIT
  systemctl disable --now detective-chicken-scan.timer >/dev/null 2>&1 || true
  rm -f /etc/systemd/system/detective-chicken-scan.service /etc/systemd/system/detective-chicken-scan.timer
  systemctl daemon-reload
  systemctl enable --now detective-chicken-heartbeat.timer
}

install_cron() {
  command -v crond >/dev/null 2>&1 || command -v cron >/dev/null 2>&1 || case "$OS_FAMILY" in
    alpine) apk add --no-cache dcron ;;
    debian) DEBIAN_FRONTEND=noninteractive apt-get install -y cron ;;
    rhel) (dnf install -y cronie || yum install -y cronie) ;;
    arch) pacman -S --noconfirm cronie ;;
  esac
  if [ -d /etc/cron.d ]; then
    cat >/etc/cron.d/detective-chicken <<'CRON'
*/2 * * * * root /usr/local/bin/detective-chicken-agent heartbeat >/dev/null 2>&1
CRON
    chmod 0644 /etc/cron.d/detective-chicken
  elif [ -d /etc/crontabs ]; then
    sed -i '/detective-chicken-agent.* scan/d' /etc/crontabs/root 2>/dev/null || true
    grep -q 'detective-chicken-agent heartbeat' /etc/crontabs/root 2>/dev/null || printf '%s\n' '*/2 * * * * /usr/local/bin/detective-chicken-agent heartbeat >/dev/null 2>&1' >>/etc/crontabs/root
  else
    return 1
  fi
  command -v rc-service >/dev/null 2>&1 && rc-service crond start >/dev/null 2>&1 || true
  command -v rc-update >/dev/null 2>&1 && rc-update add crond default >/dev/null 2>&1 || true
}

install_loop() {
  cat >/usr/local/bin/detective-chicken-loop <<'LOOP'
#!/bin/sh
while :; do
  /usr/local/bin/detective-chicken-agent heartbeat >/dev/null 2>&1 || true
  sleep 120
done
LOOP
  chmod 0755 /usr/local/bin/detective-chicken-loop
  nohup /usr/local/bin/detective-chicken-loop >/var/log/detective-chicken-agent.log 2>&1 &
  echo "No init/cron service was available; started a background loop. For containers, bake the installer into the image or persist /etc/detective-chicken."
}

if command -v systemctl >/dev/null 2>&1 && [ "$(cat /proc/1/comm 2>/dev/null || true)" = systemd ]; then
  install_systemd
elif ! install_cron; then
  install_loop
fi

echo "Starting the first IPv4/IPv6 quality report. This usually takes 1-3 minutes and at most 8 minutes."
"$AGENT" heartbeat 2>&1 | tee /var/log/detective-chicken-first-scan.log
echo "Detective Chicken installed and enrolled successfully."
