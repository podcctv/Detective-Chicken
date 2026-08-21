package agent

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// ScanProxyFromEnvironment returns the proxy dedicated to scan traffic. The
// explicit Detective Chicken variable wins; standard proxy variables are kept
// as fallbacks so existing LXC/container proxy setups continue to work.
func ScanProxyFromEnvironment() string {
	for _, key := range []string{
		"DETECTIVE_CHICKEN_SCAN_PROXY",
		"ALL_PROXY", "all_proxy",
		"HTTPS_PROXY", "https_proxy",
		"HTTP_PROXY", "http_proxy",
	} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func ValidateScanProxy(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid scan proxy URL")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "socks5", "socks5h":
		return nil
	default:
		return fmt.Errorf("unsupported scan proxy scheme %q; use http, https, socks5, or socks5h", parsed.Scheme)
	}
}
