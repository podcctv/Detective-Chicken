package agent

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestUninstallScriptSyntaxAndCoverage(t *testing.T) {
	for _, required := range []string{
		"detective-chicken-heartbeat.timer",
		"/etc/init.d/detective-chicken-heartbeat",
		"/etc/cron.d/detective-chicken",
		"/etc/detective-chicken",
		"/usr/local/bin/detective-chicken-agent",
	} {
		if !strings.Contains(uninstallScript, required) {
			t.Fatalf("uninstall script does not clean %s", required)
		}
	}
	if runtime.GOOS == "windows" {
		t.Skip("POSIX shell syntax is verified by Linux CI")
	}
	if output, err := exec.Command("/bin/sh", "-n", "-c", uninstallScript).CombinedOutput(); err != nil {
		t.Fatalf("invalid uninstall shell script: %v: %s", err, output)
	}
}
