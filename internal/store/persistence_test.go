package store

import (
	"crypto/ed25519"
	"path/filepath"
	"testing"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
)

func TestPersistentAccountsAndSessions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	first, err := NewPersistent(path, false)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := first.RegisterUser("admin", "Administrator", "admin-password-123")
	if err != nil || admin.Role != "admin" {
		t.Fatalf("create first admin: %#v %v", admin, err)
	}
	first.SetRegistrationEnabled(true)
	member, err := first.RegisterUser("member", "Member", "member-password-123")
	if err != nil || member.Role != "user" {
		t.Fatalf("create member: %#v %v", member, err)
	}
	_, token, _, err := first.CreateSession("member", "member-password-123")
	if err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewPersistent(path, false)
	if err != nil {
		t.Fatal(err)
	}
	user, tenantID, err := reloaded.UserBySession(token)
	if err != nil || user.ID != member.ID || tenantID != "tenant_default" {
		t.Fatalf("session did not survive reload: %#v %q %v", user, tenantID, err)
	}
	if !reloaded.PublicSettings().RegistrationEnabled || len(reloaded.Users()) != 2 {
		t.Fatalf("settings or users did not survive reload: %#v", reloaded.PublicSettings())
	}
}

func TestPublicDashboardRemovesControlPlaneIdentifiers(t *testing.T) {
	st := NewMemory(false)
	enrollment := st.CreateEnrollment("tenant_secret", "owner_secret", "Public Node", "Provider", "Region", "auto", "lxc", "amd64", 60)
	_, _, err := st.Register(enrollment.Token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatal(err)
	}
	dashboard := st.PublicDashboard()
	if len(dashboard.Nodes) != 1 || dashboard.Nodes[0].AgentID != "" || dashboard.Nodes[0].TenantID != "" || dashboard.Nodes[0].ScanIntervalMinutes != 60 {
		t.Fatalf("public node leaked identifiers or lost safe schedule metadata: %#v", dashboard.Nodes)
	}
	if dashboard.Nodes[0].CanViewFullIP || dashboard.Nodes[0].IPAddress != dashboard.Nodes[0].MaskedIP {
		t.Fatalf("public node was granted full-IP access: %#v", dashboard.Nodes[0])
	}
}

func TestPersistentStoreRestoresPrivateIPFromLatestReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	st, err := NewPersistent(path, false)
	if err != nil {
		t.Fatal(err)
	}
	enrollment := st.CreateEnrollment("tenant", "owner", "Node", "Provider", "Region", "auto", "lxc", "amd64", 360)
	node, agentKey, err := st.Register(enrollment.Token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatal(err)
	}
	report := model.Report{ReportID: "report-v4", NodeID: node.ID, AgentID: agentKey.AgentID, CollectedAt: time.Now().UTC(), Network: model.Network{Family: 4, ReportedIP: "203.0.113.99"}}
	if err := st.SaveReport(report); err != nil {
		t.Fatal(err)
	}

	reloaded, err := NewPersistent(path, false)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := reloaded.NodeDetailFor(node.ID, "owner", false, true)
	if err != nil || detail.IPAddress != "203.0.113.99" || detail.MaskedIP != "203.0.*.*" {
		t.Fatalf("private IP was not restored after restart: %#v %v", detail.Node, err)
	}
}

func TestMaskIPLastTwoSegments(t *testing.T) {
	cases := map[string]string{
		"203.0.113.99":       "203.0.*.*",
		"2a01:4f8:c2c:17::1": "2a01:4f8:c2c:17:0:0:*:*",
	}
	for input, want := range cases {
		if got := MaskIP(input); got != want {
			t.Errorf("MaskIP(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsWarpDetectsCloudflareVPN(t *testing.T) {
	cloudflare := model.Quality{
		ASN:          13335,
		Organization: "Cloudflare",
		Factors: map[string]any{
			"VPN":   map[string]any{"IP2LOCATION": true},
			"Proxy": map[string]any{"IP2LOCATION": false},
		},
	}
	if !isWarp(cloudflare) {
		t.Fatal("expected WARP detection for Cloudflare ASN with VPN flag")
	}
	generic := model.Quality{
		ASN:          64500,
		Organization: "Example",
		Factors: map[string]any{
			"VPN":   map[string]any{"IP2LOCATION": true},
			"Proxy": map[string]any{"IP2LOCATION": false},
		},
	}
	if isWarp(generic) {
		t.Fatal("expected no WARP for non-Cloudflare VPN")
	}
	noProxy := model.Quality{ASN: 13335, Organization: "Cloudflare", Factors: map[string]any{}}
	if isWarp(noProxy) {
		t.Fatal("expected no WARP when VPN/Proxy factors are absent")
	}
}

func TestSaveReportCapturesIPTypeYouTubeAndWarp(t *testing.T) {
	st := NewMemory(false)
	enrollment := st.CreateEnrollment("tenant", "owner", "Node", "Provider", "Region", "auto", "lxc", "amd64", 360)
	node, agentKey, err := st.Register(enrollment.Token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatal(err)
	}
	report := model.Report{
		ReportID:   "report-v4",
		NodeID:     node.ID,
		AgentID:    agentKey.AgentID,
		CollectedAt: time.Now().UTC(),
		Network:    model.Network{Family: 4, ReportedIP: "162.158.0.1"},
		Quality: model.Quality{
			ASN:         13335,
			Organization: "Cloudflare",
			UsageType:   "机房",
			Factors: map[string]any{
				"VPN":   map[string]any{"IP2LOCATION": true},
				"Proxy": map[string]any{"IP2LOCATION": false},
			},
			Media: map[string]any{
				"Netflix": map[string]any{"Status": "Yes"},
				"ChatGPT": map[string]any{"Status": "Yes"},
				"Youtube": map[string]any{"Status": "Yes", "Region": "CN"},
			},
		},
	}
	if err := st.SaveReport(report); err != nil {
		t.Fatal(err)
	}
	stored, err := st.Node(node.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.UsageType != "机房" {
		t.Fatalf("usage type not stored: %q", stored.UsageType)
	}
	if stored.YouTube != "cn" {
		t.Fatalf("youtube 送中 not detected: %q", stored.YouTube)
	}
	if !stored.WarpV4 {
		t.Fatal("WARP not detected for Cloudflare VPN IPv4")
	}
}
