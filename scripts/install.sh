#!/usr/bin/env bash
set -euo pipefail

SERVER_URL=""
ENROLL_TOKEN="${ENROLL_TOKEN:-}"
DOWNLOAD_URL="${JIJIAN_AGENT_URL:-}"

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
[[ -n "$DOWNLOAD_URL" ]] || { echo "--agent-url or JIJIAN_AGENT_URL is required" >&2; exit 2; }

install -d -m 0700 /etc/jijian
tmp_agent="$(mktemp)"
trap 'rm -f "$tmp_agent"' EXIT
curl -fsSLo "$tmp_agent" --proto '=https' --tlsv1.2 "$DOWNLOAD_URL"
install -m 0755 "$tmp_agent" /usr/local/bin/jijian-agent

cat >/etc/systemd/system/jijian-heartbeat.service <<'UNIT'
[Unit]
Description=JiJian agent heartbeat
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/jijian-agent heartbeat
User=root
UNIT

cat >/etc/systemd/system/jijian-heartbeat.timer <<'UNIT'
[Unit]
Description=Send JiJian heartbeat every two minutes

[Timer]
OnBootSec=2m
OnUnitActiveSec=2m
RandomizedDelaySec=30s
Persistent=true

[Install]
WantedBy=timers.target
UNIT

cat >/etc/systemd/system/jijian-scan.service <<'UNIT'
[Unit]
Description=JiJian VPS IP quality scan
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/jijian-agent --family 4 scan
User=root
Nice=10
IOSchedulingClass=idle
UNIT

cat >/etc/systemd/system/jijian-scan.timer <<'UNIT'
[Unit]
Description=Run JiJian IP quality scan periodically

[Timer]
OnCalendar=*-*-* 00,06,12,18:00:00
RandomizedDelaySec=30m
Persistent=true

[Install]
WantedBy=timers.target
UNIT

/usr/local/bin/jijian-agent --server "$SERVER_URL" --token "$ENROLL_TOKEN" enroll
systemctl daemon-reload
systemctl enable --now jijian-heartbeat.timer jijian-scan.timer
/usr/local/bin/jijian-agent heartbeat
echo "JiJian agent installed and enrolled."
