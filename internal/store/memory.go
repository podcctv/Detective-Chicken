package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/podcctv/jijian/internal/model"
)

var ErrNotFound = errors.New("not found")

type AgentKey struct {
	AgentID   string
	NodeID    string
	TenantID  string
	PublicKey []byte
}

type Enrollment struct {
	Token     string
	TenantID  string
	NodeName  string
	Provider  string
	Region    string
	ExpiresAt time.Time
	Used      bool
}

type Command struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type Memory struct {
	mu          sync.RWMutex
	nodes       map[string]model.Node
	series      map[string][]model.TrendPoint
	alerts      []model.Alert
	agents      map[string]AgentKey
	enrollments map[string]Enrollment
	nonces      map[string]time.Time
	reports     map[string]model.Report
	commands    map[string][]Command
}

func NewMemory(seed bool) *Memory {
	m := &Memory{
		nodes: make(map[string]model.Node), series: make(map[string][]model.TrendPoint),
		agents: make(map[string]AgentKey), enrollments: make(map[string]Enrollment),
		nonces: make(map[string]time.Time), reports: make(map[string]model.Report),
		commands: make(map[string][]Command),
	}
	if seed {
		m.seed()
	}
	return m
}

func randomID(prefix string) string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

func maskIP(ip string) string {
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		if len(parts) > 3 {
			return strings.Join(parts[:3], ":") + ":*"
		}
		return ip
	}
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return strings.Join(parts[:3], ".") + ".*"
	}
	return ip
}

func (m *Memory) seed() {
	now := time.Now().UTC()
	seedNodes := []model.Node{
		{ID: "node_hkg_01", TenantID: "tenant_demo", Name: "HK-CMI-01", Provider: "DMIT", Region: "香港", Family: 4, ReportedIP: "103.145.12.81", ASN: 906, Organization: "DMIT Cloud Services", CountryCode: "HK", Risk: 18, Status: "online", Netflix: "available", ChatGPT: "available", DNSBL: 0, LastSeen: now.Add(-42 * time.Second), LastScan: now.Add(-2 * time.Hour)},
		{ID: "node_lax_02", TenantID: "tenant_demo", Name: "US-LAX-02", Provider: "RackNerd", Region: "洛杉矶", Family: 4, ReportedIP: "198.51.100.27", ASN: 62240, Organization: "Clouvider Limited", CountryCode: "US", Risk: 72, Status: "alert", Netflix: "limited", ChatGPT: "available", DNSBL: 8, IPChanged: true, LastSeen: now.Add(-2 * time.Minute), LastScan: now.Add(-28 * time.Minute)},
		{ID: "node_nrt_03", TenantID: "tenant_demo", Name: "JP-NRT-03", Provider: "Vultr", Region: "东京", Family: 4, ReportedIP: "45.76.201.19", ASN: 20473, Organization: "The Constant Company", CountryCode: "JP", Risk: 11, Status: "online", Netflix: "available", ChatGPT: "available", DNSBL: 0, LastSeen: now.Add(-51 * time.Second), LastScan: now.Add(-6 * time.Hour)},
		{ID: "node_fra_04", TenantID: "tenant_demo", Name: "DE-FRA-04", Provider: "Hetzner", Region: "法兰克福", Family: 6, ReportedIP: "2a01:4f8:c2c:17::1", ASN: 24940, Organization: "Hetzner Online", CountryCode: "DE", Risk: 34, Status: "online", Netflix: "blocked", ChatGPT: "available", DNSBL: 1, LastSeen: now.Add(-74 * time.Second), LastScan: now.Add(-4 * time.Hour)},
		{ID: "node_sin_05", TenantID: "tenant_demo", Name: "SG-SIN-05", Provider: "LightNode", Region: "新加坡", Family: 4, ReportedIP: "172.104.51.233", ASN: 63949, Organization: "Akamai Connected Cloud", CountryCode: "SG", Risk: 48, Status: "warning", Netflix: "available", ChatGPT: "blocked", DNSBL: 3, LastSeen: now.Add(-4 * time.Minute), LastScan: now.Add(-11 * time.Hour)},
		{ID: "node_ams_06", TenantID: "tenant_demo", Name: "NL-AMS-06", Provider: "GreenCloud", Region: "阿姆斯特丹", Family: 4, ReportedIP: "185.22.153.44", ASN: 49544, Organization: "iFog GmbH", CountryCode: "NL", Risk: 22, Status: "offline", Netflix: "available", ChatGPT: "available", DNSBL: 0, LastSeen: now.Add(-19 * time.Minute), LastScan: now.Add(-13 * time.Hour)},
	}
	for i, n := range seedNodes {
		n.MaskedIP = maskIP(n.ReportedIP)
		m.nodes[n.ID] = n
		points := make([]model.TrendPoint, 0, 20)
		for d := 19; d >= 0; d-- {
			base := n.Risk + ((d+i*3)%7 - 3)
			if n.ID == "node_lax_02" && d < 4 {
				base += (4 - d) * 8
			}
			points = append(points, model.TrendPoint{At: now.Add(-time.Duration(d) * 12 * time.Hour), Risk: max(3, base), IPQS: max(2, base-5), Scam: max(1, base-11), DNSBL: n.DNSBL})
		}
		m.series[n.ID] = points
	}
	m.alerts = []model.Alert{
		{ID: "alert_01", NodeID: "node_lax_02", NodeName: "US-LAX-02", Type: "risk_spike", Severity: "critical", Title: "综合风险分显著上升", Detail: "24 小时内从 41 上升到 72，超过变化阈值 20。", CreatedAt: now.Add(-26 * time.Minute)},
		{ID: "alert_02", NodeID: "node_sin_05", NodeName: "SG-SIN-05", Type: "media_degraded", Severity: "warning", Title: "ChatGPT 解锁能力下降", Detail: "状态由 available 变为 blocked。", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "alert_03", NodeID: "node_ams_06", NodeName: "NL-AMS-06", Type: "heartbeat_missing", Severity: "warning", Title: "节点心跳超时", Detail: "已连续 19 分钟未收到心跳。", CreatedAt: now.Add(-9 * time.Minute)},
	}
}

func (m *Memory) Dashboard() model.Dashboard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	nodes := m.nodesLocked()
	stats := map[string]int{"total": len(nodes)}
	regions := map[string]int{}
	var trend []model.TrendPoint
	for _, n := range nodes {
		if n.Status != "offline" {
			stats["online"]++
		}
		if n.Status == "alert" || n.Status == "warning" {
			stats["abnormal"]++
		}
		if n.Risk >= 60 {
			stats["high_risk"]++
		}
		if n.IPChanged {
			stats["ip_changes"]++
		}
		if n.Netflix != "available" || n.ChatGPT != "available" {
			stats["media_degraded"]++
		}
		if n.DNSBL > 0 {
			stats["dnsbl_added"]++
		}
		regions[n.CountryCode]++
	}
	if s := m.series["node_lax_02"]; len(s) > 0 {
		trend = append(trend, s...)
	}
	alerts := append([]model.Alert(nil), m.alerts...)
	return model.Dashboard{GeneratedAt: time.Now().UTC(), Stats: stats, Trend: trend, Nodes: nodes, Alerts: alerts, Regions: regions}
}

func (m *Memory) nodesLocked() []model.Node {
	nodes := make([]model.Node, 0, len(m.nodes))
	for _, n := range m.nodes {
		nodes = append(nodes, n)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Risk > nodes[j].Risk })
	return nodes
}
func (m *Memory) Nodes() []model.Node { m.mu.RLock(); defer m.mu.RUnlock(); return m.nodesLocked() }
func (m *Memory) Node(id string) (model.Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.nodes[id]
	if !ok {
		return n, ErrNotFound
	}
	return n, nil
}
func (m *Memory) Series(id string) ([]model.TrendPoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.series[id]
	if !ok {
		return nil, ErrNotFound
	}
	return append([]model.TrendPoint(nil), s...), nil
}
func (m *Memory) Alerts() []model.Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]model.Alert(nil), m.alerts...)
}

func (m *Memory) CreateEnrollment(tenantID, name, provider, region string) Enrollment {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := Enrollment{Token: randomID("et"), TenantID: tenantID, NodeName: name, Provider: provider, Region: region, ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	m.enrollments[e.Token] = e
	return e
}
func (m *Memory) Register(token string, publicKey []byte) (model.Node, AgentKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.enrollments[token]
	if !ok || e.Used || time.Now().After(e.ExpiresAt) {
		return model.Node{}, AgentKey{}, ErrNotFound
	}
	e.Used = true
	m.enrollments[token] = e
	now := time.Now().UTC()
	node := model.Node{ID: randomID("node"), TenantID: e.TenantID, Name: e.NodeName, Provider: e.Provider, Region: e.Region, Status: "pending", LastSeen: now, LastScan: time.Time{}}
	node.MaskedIP = "等待首次上报"
	agent := AgentKey{AgentID: randomID("agt"), NodeID: node.ID, TenantID: e.TenantID, PublicKey: append([]byte(nil), publicKey...)}
	node.AgentID = agent.AgentID
	m.nodes[node.ID] = node
	m.agents[agent.AgentID] = agent
	return node, agent, nil
}
func (m *Memory) Agent(id string) (AgentKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.agents[id]
	if !ok {
		return a, ErrNotFound
	}
	return a, nil
}
func (m *Memory) UseNonce(agentID, nonce string, ttl time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for k, v := range m.nonces {
		if now.After(v) {
			delete(m.nonces, k)
		}
	}
	key := agentID + ":" + nonce
	if _, ok := m.nonces[key]; ok {
		return false
	}
	m.nonces[key] = now.Add(ttl)
	return true
}
func (m *Memory) SaveHeartbeat(h model.Heartbeat) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[h.NodeID]
	if !ok {
		return ErrNotFound
	}
	n.LastSeen = h.ObservedAt
	n.Status = "online"
	if h.ReportedIP != "" {
		n.ReportedIP = h.ReportedIP
		n.MaskedIP = maskIP(h.ReportedIP)
	}
	m.nodes[n.ID] = n
	return nil
}
func (m *Memory) SaveReport(r model.Report) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.reports[r.ReportID]; ok {
		return errors.New("duplicate report")
	}
	n, ok := m.nodes[r.NodeID]
	if !ok {
		return ErrNotFound
	}
	n.AgentID = r.AgentID
	n.Family = r.Network.Family
	n.ReportedIP = r.Network.ReportedIP
	n.MaskedIP = maskIP(r.Network.ReportedIP)
	n.ASN = r.Quality.ASN
	n.Organization = r.Quality.Organization
	n.CountryCode = r.Quality.CountryCode
	n.LastScan = r.CollectedAt
	n.LastSeen = time.Now().UTC()
	n.Status = "online"
	n.Risk = score(r.Quality.Scores, "ipqs")
	n.Netflix = mediaStatus(r.Quality.Media, "netflix")
	n.ChatGPT = mediaStatus(r.Quality.Media, "chatgpt")
	m.nodes[n.ID] = n
	m.reports[r.ReportID] = r
	m.series[n.ID] = append(m.series[n.ID], model.TrendPoint{At: r.CollectedAt, Risk: n.Risk, IPQS: n.Risk})
	return nil
}
func score(scores map[string]json.RawMessage, key string) int {
	var n int
	if raw, ok := scores[key]; ok {
		_ = json.Unmarshal(raw, &n)
	}
	return n
}
func mediaStatus(media map[string]any, key string) string {
	v, ok := media[key].(map[string]any)
	if !ok {
		return "unknown"
	}
	s, _ := v["status"].(string)
	return s
}
func (m *Memory) CreateScan(nodeID string) (Command, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[nodeID]
	if !ok {
		return Command{}, ErrNotFound
	}
	c := Command{ID: randomID("cmd"), Type: "scan", CreatedAt: time.Now().UTC()}
	m.commands[n.AgentID] = append(m.commands[n.AgentID], c)
	return c, nil
}
func (m *Memory) Commands(agentID string) []Command {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := append([]Command(nil), m.commands[agentID]...)
	m.commands[agentID] = nil
	return c
}
