#!/usr/bin/env bash
set -euo pipefail

SERVER_URL=""
ENROLL_TOKEN="${ENROLL_TOKEN:-}"
DOWNLOAD_URL="${DETECTIVE_CHICKEN_AGENT_URL:-}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server) SERVER_URL="$2"; shift 2 ;;
    --enroll) ENROLL_TOKEN="$2"; shift 2 ;;
    --agent-url) DOWNLOAD_URL="$2"; shift 2 ;;
    *) echo "Unknown argument: $1" >&2; exit 2 ;;
  esac
done

[[ -n "$SERVER_URL" ]] || { echo "--server is required" >&2; exit 2; }
[[ -n "$ENROLL_TOKEN" ]] || { echo "--enroll or ENROLL_TOKEN is required" >&2; exit 2; }
[[ -n "$DOWNLOAD_URL" ]] || { echo "--agent-url or DETECTIVE_CHICKEN_AGENT_URL is required" >&2; exit 2; }

install -d -m 0700 /etc/detective-chicken
tmp_agent="$(mktemp)"
trap 'rm -f "$tmp_agent"' EXIT
curl -fsSLo "$tmp_agent" --proto '=https' --tlsv1.2 "$DOWNLOAD_URL"
install -m 0755 "$tmp_agent" /usr/local/bin/detective-chicken-agent

cat >/etc/systemd/system/detective-chicken-heartbeat.service <<'UNIT'
[Unit]
Description=Detective Chicken agent heartbeat
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/detective-chicken-agent heartbeat
User=root
UNIT

cat >/etc/systemd/system/detective-chicken-heartbeat.timer <<'UNIT'
[Unit]
Description=Send Detective Chicken heartbeat every two minutes

[Timer]
OnBootSec=2m
OnUnitActiveSec=2m
RandomizedDelaySec=30s
Persistent=true

[Install]
WantedBy=timers.target
UNIT

cat >/etc/systemd/system/detective-chicken-scan.service <<'UNIT'
[Unit]
Description=Detective Chicken VPS IP quality scan
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/detective-chicken-agent --family 4 scan
User=root
Nice=10
IOSchedulingClass=idle
UNIT

cat >/etc/systemd/system/detective-chicken-scan.timer <<'UNIT'
[Unit]
Description=Run Detective Chicken IP quality scan periodically

[Timer]
OnCalendar=*-*-* 00,06,12,18:00:00
RandomizedDelaySec=30m
Persistent=true

[Install]
WantedBy=timers.target
UNIT

/usr/local/bin/detective-chicken-agent --server "$SERVER_URL" --token "$ENROLL_TOKEN" enroll
systemctl daemon-reload
systemctl enable --now detective-chicken-heartbeat.timer detective-chicken-scan.timer
/usr/local/bin/detective-chicken-agent heartbeat
echo "Detective Chicken agent installed and enrolled."
