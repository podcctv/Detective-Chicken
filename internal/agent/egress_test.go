package agent

import "testing"

func TestValidateScanProxy(t *testing.T) {
	for _, value := range []string{"", "http://127.0.0.1:8080", "https://proxy.example:8443", "socks5://127.0.0.1:1080", "socks5h://user:pass@proxy.example:1080"} {
		if err := ValidateScanProxy(value); err != nil {
			t.Errorf("ValidateScanProxy(%q): %v", value, err)
		}
	}
	for _, value := range []string{"127.0.0.1:1080", "ftp://proxy.example"} {
		if err := ValidateScanProxy(value); err == nil {
			t.Errorf("ValidateScanProxy(%q) unexpectedly succeeded", value)
		}
	}
}

func TestScanProxyFromEnvironmentUsesDedicatedThenSystemProxy(t *testing.T) {
	for _, key := range []string{"DETECTIVE_CHICKEN_SCAN_PROXY", "ALL_PROXY", "all_proxy", "HTTPS_PROXY", "https_proxy", "HTTP_PROXY", "http_proxy"} {
		t.Setenv(key, "")
	}
	t.Setenv("HTTPS_PROXY", "http://system-proxy:8080")
	if got := ScanProxyFromEnvironment(); got != "http://system-proxy:8080" {
		t.Fatalf("system proxy = %q", got)
	}
	t.Setenv("DETECTIVE_CHICKEN_SCAN_PROXY", "socks5h://scan-proxy:1080")
	if got := ScanProxyFromEnvironment(); got != "socks5h://scan-proxy:1080" {
		t.Fatalf("dedicated proxy = %q", got)
	}
}
