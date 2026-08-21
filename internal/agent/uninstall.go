package agent

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const uninstallScript = `
set -u

if command -v systemctl >/dev/null 2>&1; then
  systemctl disable detective-chicken-heartbeat.timer detective-chicken-scan.timer >/dev/null 2>&1 || true
  systemctl stop detective-chicken-heartbeat.timer detective-chicken-scan.timer >/dev/null 2>&1 || true
  rm -f /etc/systemd/system/detective-chicken-heartbeat.service \
    /etc/systemd/system/detective-chicken-heartbeat.timer \
    /etc/systemd/system/detective-chicken-scan.service \
    /etc/systemd/system/detective-chicken-scan.timer
  systemctl daemon-reload >/dev/null 2>&1 || true
fi

if command -v rc-update >/dev/null 2>&1; then
  rc-update del detective-chicken-heartbeat default >/dev/null 2>&1 || true
fi
rm -f /etc/init.d/detective-chicken-heartbeat /etc/cron.d/detective-chicken

for crontab_file in /etc/crontabs/root /var/spool/cron/crontabs/root; do
  if [ -f "$crontab_file" ]; then
    sed -i '/detective-chicken-agent/d' "$crontab_file" 2>/dev/null || true
  fi
done

loop_pid=""
if [ -r /run/detective-chicken-heartbeat.pid ]; then
  loop_pid=$(cat /run/detective-chicken-heartbeat.pid 2>/dev/null || true)
fi

rm -rf /etc/detective-chicken
rm -f /usr/local/bin/detective-chicken-loop \
  /var/log/detective-chicken-agent.log \
  /var/log/detective-chicken-first-scan.log \
  /run/detective-chicken-heartbeat.pid \
  /usr/local/bin/detective-chicken-agent

case "$loop_pid" in
  ''|*[!0-9]*) ;;
  *) kill "$loop_pid" >/dev/null 2>&1 || true ;;
esac
`

func RunLocalUninstall() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("remote uninstall is supported only on linux agents")
	}
	output, err := exec.Command("/bin/sh", "-c", uninstallScript).CombinedOutput()
	if err != nil {
		return fmt.Errorf("remove agent services: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
