package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCommandIsRootAwareAndQuotesURL(t *testing.T) {
	command := installCommand("https://example.com/install/a'b.sh")
	for _, expected := range []string{
		`if [ "$(id -u)" -eq 0 ]`,
		"runner=sudo",
		"runner=doas",
		`'https://example.com/install/a'"'"'b.sh'`,
	} {
		if !strings.Contains(command, expected) {
			t.Fatalf("install command missing %q: %s", expected, command)
		}
	}
}

func TestGeneratedInstallerAndAgentDownload(t *testing.T) {
	api, _ := testAPI()
	cookie, _ := registerUser(t, api, "installer-admin", "installer-password-123")
	body := []byte(`{"name":"Alpine LXC","os_family":"alpine","platform":"lxc","arch":"arm64"}`)
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPost, "/api/v1/enrollment-tokens", bytes.NewReader(body)), cookie))
	if res.Code != http.StatusCreated {
		t.Fatalf("create enrollment: %d %s", res.Code, res.Body.String())
	}
	var enrollment map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &enrollment)
	installCommand, _ := enrollment["install_command"].(string)
	for _, expected := range []string{"$(id -u)", "runner=sudo", "runner=doas", "${runner:+$runner }env", "DETECTIVE_CHICKEN_SCAN_PROXY"} {
		if !strings.Contains(installCommand, expected) {
			t.Errorf("install command missing %q: %s", expected, installCommand)
		}
	}
	if strings.Contains(installCommand, "| sudo sh") {
		t.Fatalf("install command still requires sudo for root: %s", installCommand)
	}
	installURL := enrollment["install_url"].(string)
	path := strings.TrimPrefix(installURL, "http://example.com")
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
	if res.Code != http.StatusOK {
		t.Fatalf("download installer: %d %s", res.Code, res.Body.String())
	}
	script := res.Body.String()
	for _, expected := range []string{"REQUESTED_OS='alpine'", "REQUESTED_PLATFORM='lxc'", "REQUESTED_ARCH='arm64'", "/api/v1/downloads/agent/$ARCH", "install_systemd", "install_openrc", "install_cron", "install_loop", "Switch to root first", "first IPv4/IPv6 quality report", "usually takes 1-3 minutes"} {
		if !strings.Contains(script, expected) {
			t.Errorf("installer missing %q", expected)
		}
	}
	if strings.Contains(script, "--family 4 scan") {
		t.Fatal("installer still hard-codes IPv4-only scheduled scans")
	}

	dir := t.TempDir()
	t.Setenv("DETECTIVE_CHICKEN_AGENT_DIR", dir)
	binary := []byte("test-agent-binary")
	if err := os.WriteFile(filepath.Join(dir, "detective-chicken-agent-arm64"), binary, 0755); err != nil {
		t.Fatal(err)
	}
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/downloads/agent/arm64", nil))
	if res.Code != http.StatusOK || !bytes.Equal(res.Body.Bytes(), binary) {
		t.Fatalf("agent download failed: %d %q", res.Code, res.Body.Bytes())
	}
}
