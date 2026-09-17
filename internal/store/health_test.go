package store

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
)

func TestNodeOfflineTimeout(t *testing.T) {
	st := NewMemory(false)
	st.SetHeartbeatTimeout(100 * time.Millisecond)

	enrollment := st.CreateEnrollment("tenant_1", "owner_1", "Test Node", "Provider", "Region", "auto", "lxc", "amd64", 60)
	node, agentKey, err := st.Register(enrollment.Token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Send initial heartbeat to transition from pending to online
	now := time.Now().UTC()
	err = st.SaveHeartbeat(model.Heartbeat{
		AgentID:    agentKey.AgentID,
		NodeID:     node.ID,
		ObservedAt: now,
	})
	if err != nil {
		t.Fatalf("save heartbeat failed: %v", err)
	}

	// Immediately, node should be online
	detail, err := st.NodeDetailFor(node.ID, "owner_1", false, false)
	if err != nil {
		t.Fatalf("node detail: %v", err)
	}
	if detail.Status != "online" {
		t.Fatalf("expected node status 'online', got %q", detail.Status)
	}

	d := st.DashboardFor("owner_1", false)
	if d.Stats["online"] != 1 || d.Stats["offline"] != 0 {
		t.Fatalf("expected stats online=1, offline=0; got online=%d, offline=%d", d.Stats["online"], d.Stats["offline"])
	}

	// Wait for heartbeat timeout
	time.Sleep(150 * time.Millisecond)

	// Now node should be dynamically computed as offline
	detail, err = st.NodeDetailFor(node.ID, "owner_1", false, false)
	if err != nil {
		t.Fatalf("node detail after timeout: %v", err)
	}
	if detail.Status != "offline" {
		t.Fatalf("expected node status 'offline' after timeout, got %q", detail.Status)
	}

	// Check NodesFor
	nodes := st.NodesFor("owner_1", false)
	if len(nodes) != 1 || nodes[0].Status != "offline" {
		t.Fatalf("expected NodesFor to return offline node, got %#v", nodes)
	}

	// Check DashboardFor stats
	d = st.DashboardFor("owner_1", false)
	if d.Stats["online"] != 0 || d.Stats["offline"] != 1 {
		t.Fatalf("expected stats online=0, offline=1; got online=%d, offline=%d", d.Stats["online"], d.Stats["offline"])
	}

	// Check PublicDashboard
	pub := st.PublicDashboard()
	if len(pub.Nodes) != 1 || pub.Nodes[0].Status != "offline" {
		t.Fatalf("expected PublicDashboard to show offline node, got %#v", pub.Nodes)
	}
	if pub.Stats["online"] != 0 || pub.Stats["offline"] != 1 {
		t.Fatalf("expected public stats online=0, offline=1; got online=%d, offline=%d", pub.Stats["online"], pub.Stats["offline"])
	}

	// Heartbeat arrives again -> node immediately recovers to online
	err = st.SaveHeartbeat(model.Heartbeat{
		AgentID:    agentKey.AgentID,
		NodeID:     node.ID,
		ObservedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("save heartbeat failed: %v", err)
	}

	detail, err = st.NodeDetailFor(node.ID, "owner_1", false, false)
	if err != nil {
		t.Fatalf("node detail: %v", err)
	}
	if detail.Status != "online" {
		t.Fatalf("expected node to recover to 'online', got %q", detail.Status)
	}
}

func TestReconcileNodeHealthAndAlerts(t *testing.T) {
	st := NewMemory(false)
	st.SetHeartbeatTimeout(50 * time.Millisecond)

	enrollment := st.CreateEnrollment("tenant_1", "owner_1", "Alert Node", "Provider", "Region", "auto", "lxc", "amd64", 60)
	node, agentKey, err := st.Register(enrollment.Token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_ = st.SaveHeartbeat(model.Heartbeat{
		AgentID:    agentKey.AgentID,
		NodeID:     node.ID,
		ObservedAt: time.Now().UTC(),
	})

	// Wait for timeout
	time.Sleep(80 * time.Millisecond)

	// Run reconciliation
	st.ReconcileNodeHealth()

	// Should generate a heartbeat_missing alert
	d := st.DashboardFor("owner_1", false)
	foundAlert := false
	for _, a := range d.Alerts {
		if a.NodeID == node.ID && a.Type == "heartbeat_missing" {
			foundAlert = true
			break
		}
	}
	if !foundAlert {
		t.Fatalf("expected heartbeat_missing alert to be generated, alerts: %#v", d.Alerts)
	}

	// Now node reports heartbeat
	_ = st.SaveHeartbeat(model.Heartbeat{
		AgentID:    agentKey.AgentID,
		NodeID:     node.ID,
		ObservedAt: time.Now().UTC(),
	})

	// Alert should be cleared
	d = st.DashboardFor("owner_1", false)
	for _, a := range d.Alerts {
		if a.NodeID == node.ID && a.Type == "heartbeat_missing" {
			t.Fatalf("expected heartbeat_missing alert to be removed upon heartbeat, but found %#v", a)
		}
	}
}

func TestScanningTaskAbortsOnNodeOffline(t *testing.T) {
	st := NewMemory(false)
	st.SetHeartbeatTimeout(50 * time.Millisecond)

	enrollment := st.CreateEnrollment("tenant_1", "owner_1", "Task Node", "Provider", "Region", "auto", "lxc", "amd64", 60)
	node, agentKey, err := st.Register(enrollment.Token, make([]byte, ed25519.PublicKeySize))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	// Trigger a scan command
	cmd, err := st.CreateScan(node.ID)
	if err != nil {
		t.Fatalf("create scan failed: %v", err)
	}
	if cmd.Type != "scan" {
		t.Fatalf("expected scan command, got %#v", cmd)
	}

	// Node acknowledges scan start in heartbeat
	_ = st.SaveHeartbeat(model.Heartbeat{
		AgentID:    agentKey.AgentID,
		NodeID:     node.ID,
		ObservedAt: time.Now().UTC(),
		Status:     map[string]any{"scan_state": "scanning"},
	})

	detail, _ := st.NodeDetailFor(node.ID, "owner_1", false, false)
	if detail.QualityStatus != "scanning" || detail.LastTask == nil || detail.LastTask.Status != "running" {
		t.Fatalf("expected scanning status, got quality=%q task=%#v", detail.QualityStatus, detail.LastTask)
	}

	// Node disconnects and times out
	time.Sleep(80 * time.Millisecond)
	st.ReconcileNodeHealth()

	detail, _ = st.NodeDetailFor(node.ID, "owner_1", false, false)
	if detail.Status != "offline" {
		t.Fatalf("expected offline status, got %q", detail.Status)
	}
	if detail.QualityStatus != "failed" {
		t.Fatalf("expected quality status 'failed', got %q", detail.QualityStatus)
	}
	if detail.LastTask == nil || detail.LastTask.Status != "failed" {
		t.Fatalf("expected task status 'failed', got %#v", detail.LastTask)
	}
}
