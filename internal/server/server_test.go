package server

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/podcctv/detective-chicken/internal/store"
)

func testAPI() (*API, *store.Memory) {
	st := store.NewMemory(false)
	return New(st, slog.New(slog.NewTextHandler(io.Discard, nil))), st
}

func registerUser(t *testing.T, api *API, username, password string) (*http.Cookie, map[string]any) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "display_name": username, "password": password})
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body)))
	if res.Code != http.StatusCreated {
		t.Fatalf("register user: %d %s", res.Code, res.Body.String())
	}
	var out map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &out)
	cookies := res.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("registration did not set session cookie")
	}
	return cookies[0], out
}

func withCookie(req *http.Request, cookie *http.Cookie) *http.Request {
	req.AddCookie(cookie)
	return req
}

func TestEnrollmentRegistrationAndSignedHeartbeat(t *testing.T) {
	api, st := testAPI()
	cookie, _ := registerUser(t, api, "owner", "correct-horse-battery")
	enrollBody := []byte(`{"name":"HK-Test","provider":"Demo","region":"HK"}`)
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPost, "/api/v1/enrollment-tokens", bytes.NewReader(enrollBody)), cookie))
	if res.Code != 201 {
		t.Fatalf("enrollment: %d %s", res.Code, res.Body.String())
	}
	var enroll map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &enroll)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	regBody, _ := json.Marshal(map[string]string{"public_key": base64.StdEncoding.EncodeToString(pub)})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/register", bytes.NewReader(regBody))
	req.Header.Set("Authorization", "Bearer "+enroll["token"].(string))
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)
	if res.Code != 201 {
		t.Fatalf("register: %d %s", res.Code, res.Body.String())
	}
	var reg map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &reg)
	hb, _ := json.Marshal(map[string]any{"agent_id": reg["agent_id"], "node_id": reg["node_id"], "observed_at": time.Now().UTC(), "agent_version": "test"})
	req = signedRequest(t, http.MethodPost, "/api/v1/heartbeats", hb, reg["agent_id"].(string), priv, "nonce-1")
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)
	if res.Code != 202 {
		t.Fatalf("heartbeat: %d %s", res.Code, res.Body.String())
	}
	n, err := st.Node(reg["node_id"].(string))
	if err != nil || n.Status != "online" {
		t.Fatalf("node was not updated: %#v %v", n, err)
	}
	report, _ := json.Marshal(map[string]any{
		"schema_version": "1.0", "report_id": "rpt_test_000001", "agent_id": reg["agent_id"], "node_id": reg["node_id"], "collected_at": time.Now().UTC(),
		"collector": map[string]any{"name": "ipquality", "adapter_version": "test"},
		"network":   map[string]any{"family": 4, "reported_ip": "203.0.113.9"},
		"quality":   map[string]any{"asn": 64500, "organization": "Example", "country_code": "US", "scores": map[string]any{"ipqs": 61}, "factors": map[string]any{}, "media": map[string]any{}, "mail": map[string]any{}},
	})
	req = signedRequest(t, http.MethodPost, "/api/v1/reports", report, reg["agent_id"].(string), priv, "nonce-2")
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)
	if res.Code != 202 {
		t.Fatalf("report: %d %s", res.Code, res.Body.String())
	}
	n, _ = st.Node(reg["node_id"].(string))
	if n.Risk != 61 || n.MaskedIP != "203.0.*.*" {
		t.Fatalf("report did not update normalized node state: %#v", n)
	}
	req = signedRequest(t, http.MethodPost, "/api/v1/reports", report, reg["agent_id"].(string), priv, "nonce-3")
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)
	if res.Code != 409 {
		t.Fatalf("expected duplicate report rejection, got %d", res.Code)
	}
	res = httptest.NewRecorder()
	req = signedRequest(t, http.MethodPost, "/api/v1/heartbeats", hb, reg["agent_id"].(string), priv, "nonce-1")
	api.Handler().ServeHTTP(res, req)
	if res.Code != 409 {
		t.Fatalf("expected replay detection, got %d", res.Code)
	}
}

func TestDashboardSeed(t *testing.T) {
	api, _ := testAPI()
	api.store = store.NewMemory(true)
	cookie, _ := registerUser(t, api, "admin", "correct-horse-battery")
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil), cookie))
	if res.Code != 200 {
		t.Fatal(res.Code)
	}
	var body struct {
		Nodes []any `json:"nodes"`
	}
	_ = json.Unmarshal(res.Body.Bytes(), &body)
	if len(body.Nodes) < 5 {
		t.Fatalf("expected demo nodes, got %d", len(body.Nodes))
	}
}

func TestAccountLifecycleAndAuthorization(t *testing.T) {
	api, _ := testAPI()
	adminCookie, adminOut := registerUser(t, api, "first.admin", "first-password-123")
	admin := adminOut["user"].(map[string]any)
	if admin["role"] != "admin" {
		t.Fatalf("first user should be admin: %#v", admin)
	}

	second := []byte(`{"username":"member","password":"member-password-123","display_name":"Member"}`)
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(second)))
	if res.Code != http.StatusForbidden {
		t.Fatalf("registration should be closed after bootstrap: %d %s", res.Code, res.Body.String())
	}

	settings := []byte(`{"registration_enabled":true}`)
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPatch, "/api/v1/admin/settings", bytes.NewReader(settings)), adminCookie))
	if res.Code != http.StatusOK {
		t.Fatalf("enable registration: %d %s", res.Code, res.Body.String())
	}

	memberCookie, memberOut := registerUser(t, api, "member", "member-password-123")
	member := memberOut["user"].(map[string]any)
	if member["role"] != "user" {
		t.Fatalf("later user should not be admin: %#v", member)
	}
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil), memberCookie))
	if res.Code != http.StatusForbidden {
		t.Fatalf("member reached admin API: %d", res.Code)
	}

	memberID := member["id"].(string)
	role := []byte(`{"role":"admin"}`)
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+memberID, bytes.NewReader(role)), adminCookie))
	if res.Code != http.StatusOK {
		t.Fatalf("promote member: %d %s", res.Code, res.Body.String())
	}

	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+memberID+"/password-reset", nil), adminCookie))
	if res.Code != http.StatusCreated {
		t.Fatalf("create reset: %d %s", res.Code, res.Body.String())
	}
	var reset map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &reset)
	resetBody, _ := json.Marshal(map[string]string{"token": reset["token"].(string), "new_password": "member-password-456"})
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/auth/password-reset/complete", bytes.NewReader(resetBody)))
	if res.Code != http.StatusOK {
		t.Fatalf("complete reset: %d %s", res.Code, res.Body.String())
	}
	login := []byte(`{"username":"member","password":"member-password-456"}`)
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(login)))
	if res.Code != http.StatusOK {
		t.Fatalf("login after reset: %d %s", res.Code, res.Body.String())
	}
}

func TestNodeOwnershipAndIPVisibility(t *testing.T) {
	api, _ := testAPI()
	ownerCookie, _ := registerUser(t, api, "owner", "owner-password-123")
	settings := []byte(`{"registration_enabled":true}`)
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPatch, "/api/v1/admin/settings", bytes.NewReader(settings)), ownerCookie))
	memberCookie, _ := registerUser(t, api, "member", "member-password-123")

	enrollBody := []byte(`{"name":"Owned Node","provider":"Demo","region":"US","os_family":"alpine","platform":"lxc","arch":"arm64"}`)
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodPost, "/api/v1/enrollment-tokens", bytes.NewReader(enrollBody)), ownerCookie))
	var enroll map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &enroll)
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	regBody, _ := json.Marshal(map[string]string{"public_key": base64.StdEncoding.EncodeToString(pub)})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agents/register", bytes.NewReader(regBody))
	req.Header.Set("Authorization", "Bearer "+enroll["token"].(string))
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)
	var reg map[string]any
	_ = json.Unmarshal(res.Body.Bytes(), &reg)

	report, _ := json.Marshal(map[string]any{
		"schema_version": "1.0", "report_id": "rpt_visibility", "agent_id": reg["agent_id"], "node_id": reg["node_id"], "collected_at": time.Now().UTC(),
		"collector": map[string]any{"name": "ipquality", "adapter_version": "test"},
		"network":   map[string]any{"family": 4, "reported_ip": "203.0.113.99"},
		"quality":   map[string]any{"asn": 64500, "organization": "Example", "country_code": "US", "scores": map[string]any{"ipqs": 12}, "factors": map[string]any{}, "media": map[string]any{}, "mail": map[string]any{}},
	})
	req = signedRequest(t, http.MethodPost, "/api/v1/reports", report, reg["agent_id"].(string), priv, "visibility-nonce")
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)

	nodeID := reg["node_id"].(string)
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodGet, "/api/v1/nodes/"+nodeID, nil), ownerCookie))
	if strings.Contains(res.Body.String(), "203.0.113.99") || !strings.Contains(res.Body.String(), "203.0.*.*") {
		t.Fatalf("default node response leaked or failed to mask IP: %s", res.Body.String())
	}
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodGet, "/api/v1/nodes/"+nodeID+"?full_ip=true", nil), ownerCookie))
	if !strings.Contains(res.Body.String(), `"ip_address":"203.0.113.99"`) {
		t.Fatalf("owner could not reveal full IP: %s", res.Body.String())
	}
	res = httptest.NewRecorder()
	api.Handler().ServeHTTP(res, withCookie(httptest.NewRequest(http.MethodGet, "/api/v1/nodes/"+nodeID+"?full_ip=true", nil), memberCookie))
	if res.Code != http.StatusNotFound {
		t.Fatalf("unrelated member accessed node: %d %s", res.Code, res.Body.String())
	}
}

func signedRequest(t *testing.T, method, path string, body []byte, keyID string, priv ed25519.PrivateKey, nonce string) *http.Request {
	t.Helper()
	created := time.Now().Unix()
	meta := signatureMeta{Created: created, KeyID: keyID, Nonce: nonce}
	digest := contentDigest(body)
	sig := ed25519.Sign(priv, signingBase(method, path, digest, meta))
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Digest", digest)
	req.Header.Set("Signature-Input", `sig1=("@method" "@path" "content-digest");created=`+strconv.FormatInt(created, 10)+`;keyid="`+keyID+`";nonce="`+nonce+`"`)
	req.Header.Set("Signature", "sig1=:"+base64.StdEncoding.EncodeToString(sig)+":")
	return req
}
