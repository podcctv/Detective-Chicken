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
	"testing"
	"time"

	"github.com/podcctv/detective-chicken/internal/store"
)

func testAPI() (*API, *store.Memory) {
	st := store.NewMemory(false)
	return New(st, slog.New(slog.NewTextHandler(io.Discard, nil))), st
}

func TestEnrollmentRegistrationAndSignedHeartbeat(t *testing.T) {
	api, st := testAPI()
	enrollBody := []byte(`{"name":"HK-Test","provider":"Demo","region":"HK"}`)
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/v1/enrollment-tokens", bytes.NewReader(enrollBody)))
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
	if n.Risk != 61 || n.MaskedIP != "203.0.113.*" {
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
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil))
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
