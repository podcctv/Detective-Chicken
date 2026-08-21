package agent

import (
	"path/filepath"
	"testing"
)

func TestConfigureScanProxyPersistsAndClearsEgress(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.json")
	initial := Config{ServerURL: "https://control.example", AgentID: "agent", NodeID: "node", TenantID: "tenant", PrivateKey: "private"}
	if err := SaveConfig(path, initial); err != nil {
		t.Fatal(err)
	}

	configured, err := ConfigureScanProxy(path, " socks5h://backend.example:1080 ")
	if err != nil || configured.ScanProxy != "socks5h://backend.example:1080" {
		t.Fatalf("configure proxy: %#v %v", configured, err)
	}
	reloaded, err := LoadConfig(path)
	if err != nil || reloaded.ScanProxy != configured.ScanProxy {
		t.Fatalf("proxy was not persisted: %#v %v", reloaded, err)
	}

	cleared, err := ConfigureScanProxy(path, "")
	if err != nil || cleared.ScanProxy != "" {
		t.Fatalf("clear proxy: %#v %v", cleared, err)
	}
}
