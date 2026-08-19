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
	installURL := enrollment["install_url"].(string)
	path := strings.TrimPrefix(installURL, "http://example.com")
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
	if res.Code != http.StatusOK {
		t.Fatalf("download installer: %d %s", res.Code, res.Body.String())
	}
	script := res.Body.String()
	for _, expected := range []string{"REQUESTED_OS='alpine'", "REQUESTED_PLATFORM='lxc'", "REQUESTED_ARCH='arm64'", "/api/v1/downloads/agent/$ARCH", "install_systemd", "install_cron", "install_loop"} {
		if !strings.Contains(script, expected) {
			t.Errorf("installer missing %q", expected)
		}
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
