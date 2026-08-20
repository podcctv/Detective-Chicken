package store

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
)

var ErrNotFound = errors.New("not found")

const DefaultScanIntervalMinutes = 360

type AgentKey struct {
	AgentID   string
	NodeID    string
	TenantID  string
	PublicKey []byte
}

type Enrollment struct {
	Token               string
	TenantID            string
	OwnerUserID         string
	NodeName            string
	Provider            string
	Region              string
	OSFamily            string
	Platform            string
	Arch                string
	ScanIntervalMinutes int
	ExpiresAt           time.Time
	Used                bool
}

type Command struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type Memory struct {
	mu                  sync.RWMutex
	nodes               map[string]model.Node
	series              map[string][]model.TrendPoint
	alerts              []model.Alert
	agents              map[string]AgentKey
	enrollments         map[string]Enrollment
	nonces              map[string]time.Time
	reports             map[string]model.Report
	commands            map[string][]Command
	users               map[string]UserAccount
	usernames           map[string]string
	sessions            map[string]Session
	passwordResets      map[string]PasswordReset
	registrationEnabled bool
	dataPath            string
}

func NewMemory(seed bool) *Memory {
	m := &Memory{
		nodes: make(map[string]model.Node), series: make(map[string][]model.TrendPoint),
		agents: make(map[string]AgentKey), enrollments: make(map[string]Enrollment),
		nonces: make(map[string]time.Time), reports: make(map[string]model.Report),
		commands: make(map[string][]Command), users: make(map[string]UserAccount),
		usernames: make(map[string]string), sessions: make(map[string]Session),
		passwordResets: make(map[string]PasswordReset),
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

func MaskIP(ip string) string {
	if strings.Contains(ip, ":") {
		if addr, err := netip.ParseAddr(ip); err == nil && addr.Is6() {
			raw := addr.As16()
			parts := make([]string, 0, 8)
			for i := 0; i < 6; i++ {
				parts = append(parts, strconv.FormatUint(uint64(binary.BigEndian.Uint16(raw[i*2:i*2+2])), 16))
			}
			return strings.Join(parts, ":") + ":*:*"
		}
		return "*:*"
	}
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return strings.Join(parts[:2], ".") + ".*.*"
	}
	return ip
}

func parseNodeUnlocks(media map[string]any, defaultRegion string) model.NodeUnlocks {
	stream := map[string]model.UnlockInfo{}
	ai := map[string]model.UnlockInfo{}

	for rawKey, rawVal := range media {
		entry, ok := rawVal.(map[string]any)
		if !ok {
			continue
		}
		statusRaw := strings.TrimSpace(fmt.Sprint(entry["Status"]))
		regionRaw := strings.TrimSpace(fmt.Sprint(entry["Region"]))
		if regionRaw == "" || regionRaw == "<nil>" {
			regionRaw = defaultRegion
		}
		typeRaw := strings.TrimSpace(fmt.Sprint(entry["Type"]))
		if typeRaw == "<nil>" {
			typeRaw = ""
		}

		status := "blocked"
		switch strings.ToLower(statusRaw) {
		case "available", "unlocked", "yes", "解锁", "可用", "原生", "允许", "true":
			status = "available"
		case "limited", "partial", "仅自制", "部分解锁", "限制":
			status = "limited"
		case "blocked", "no", "不可用", "未解锁", "屏蔽", "失败", "质询", "false":
			status = "blocked"
		}

		keyLower := strings.ToLower(rawKey)
		serviceID := keyLower
		category := "streaming"
		serviceName := rawKey

		switch keyLower {
		case "chatgpt", "openai":
			serviceID = "chatgpt"
			serviceName = "ChatGPT"
			category = "ai"
		case "claude", "anthropic":
			serviceID = "claude"
			serviceName = "Claude"
			category = "ai"
		case "gemini", "google_ai":
			serviceID = "gemini"
			serviceName = "Gemini"
			category = "ai"
		case "deepseek":
			serviceID = "deepseek"
			serviceName = "DeepSeek"
			category = "ai"
		case "midjourney":
			serviceID = "midjourney"
			serviceName = "Midjourney"
			category = "ai"
		case "copilot", "bing_copilot":
			serviceID = "copilot"
			serviceName = "Copilot"
			category = "ai"
		case "grok", "xai":
			serviceID = "grok"
			serviceName = "Grok"
			category = "ai"
		case "perplexity":
			serviceID = "perplexity"
			serviceName = "Perplexity"
			category = "ai"
		case "github_cop", "github_copilot":
			serviceID = "github_cop"
			serviceName = "GitHub Copilot"
			category = "ai"
		case "huggingface":
			serviceID = "huggingface"
			serviceName = "HuggingFace"
			category = "ai"
		case "reddit":
			serviceID = "reddit"
			serviceName = "Reddit"
			category = "ai"
		case "netflix":
			serviceID = "netflix"
			serviceName = "Netflix"
			category = "streaming"
		case "disneyplus", "disney", "disney+":
			serviceID = "disney"
			serviceName = "Disney+"
			category = "streaming"
		case "youtube", "yt":
			serviceID = "youtube"
			serviceName = "YouTube Prem"
			category = "streaming"
		case "amazonprimevideo", "primevideo", "prime":
			serviceID = "prime"
			serviceName = "Prime Video"
			category = "streaming"
		case "spotify":
			serviceID = "spotify"
			serviceName = "Spotify"
			category = "streaming"
		case "tiktok":
			serviceID = "tiktok"
			serviceName = "TikTok"
			category = "streaming"
		case "max", "hbo", "hbomax":
			serviceID = "max"
			serviceName = "Max (HBO)"
			category = "streaming"
		case "hulu":
			serviceID = "hulu"
			serviceName = "Hulu"
			category = "streaming"
		case "bilibili":
			serviceID = "bilibili"
			serviceName = "Bilibili"
			category = "streaming"
		case "appletv", "apple_tv", "appletv+":
			serviceID = "appletv"
			serviceName = "Apple TV+"
			category = "streaming"
		}

		qualityDesc := statusRaw
		if typeRaw != "" && typeRaw != statusRaw {
			qualityDesc = typeRaw + " " + statusRaw
		}

		info := model.UnlockInfo{
			ID:       serviceID,
			Name:     serviceName,
			Category: category,
			Status:   status,
			Region:   regionRaw,
			Quality:  qualityDesc,
			Detail:   fmt.Sprintf("%s · %s", serviceName, qualityDesc),
		}

		if category == "ai" {
			ai[serviceID] = info
		} else {
			stream[serviceID] = info
		}
	}

	return model.NodeUnlocks{Streaming: stream, AI: ai}
}

func computeRisk(scores map[string]json.RawMessage) int {
	maxRisk := 0
	found := false
	for k, v := range scores {
		valStr := strings.Trim(string(v), `"`)
		if valStr == "null" || valStr == "" {
			continue
		}
		valStr = strings.TrimSuffix(valStr, "%")
		if f, err := strconv.ParseFloat(valStr, 64); err == nil {
			val := int(f)
			if strings.EqualFold(k, "ipqs") || strings.EqualFold(k, "scamalytics") || strings.EqualFold(k, "abuseipdb") || strings.EqualFold(k, "ip2location") || strings.EqualFold(k, "dbip") {
				if val > maxRisk {
					maxRisk = val
				}
				found = true
			}
		}
	}
	if !found {
		return 0
	}
	return maxRisk
}


func (m *Memory) seed() {
	now := time.Now().UTC()
	seedNodes := []model.Node{
		{ID: "node_hkg_01", TenantID: "tenant_demo", Name: "HK-CMI-01", Provider: "DMIT", Region: "香港", Family: 4, ReportedIP: "103.145.12.81", ASN: 906, Organization: "DMIT Cloud Services", CountryCode: "HK", Latitude: 22.3193, Longitude: 114.1694, Risk: 18, Status: "online", Netflix: "available", ChatGPT: "available", Unlocks: mockNodeUnlocks("HK", "available", "available"), DNSBL: 0, LastSeen: now.Add(-42 * time.Second), LastScan: now.Add(-2 * time.Hour)},
		{ID: "node_lax_02", TenantID: "tenant_demo", Name: "US-LAX-02", Provider: "RackNerd", Region: "洛杉矶", Family: 4, ReportedIP: "198.51.100.27", ASN: 62240, Organization: "Clouvider Limited", CountryCode: "US", Latitude: 34.0522, Longitude: -118.2437, Risk: 72, Status: "alert", Netflix: "limited", ChatGPT: "available", Unlocks: mockNodeUnlocks("US", "limited", "available"), DNSBL: 8, IPChanged: true, LastSeen: now.Add(-2 * time.Minute), LastScan: now.Add(-28 * time.Minute)},
		{ID: "node_nrt_03", TenantID: "tenant_demo", Name: "JP-NRT-03", Provider: "Vultr", Region: "东京", Family: 4, ReportedIP: "45.76.201.19", ASN: 20473, Organization: "The Constant Company", CountryCode: "JP", Latitude: 35.6762, Longitude: 139.6503, Risk: 11, Status: "online", Netflix: "available", ChatGPT: "available", Unlocks: mockNodeUnlocks("JP", "available", "available"), DNSBL: 0, LastSeen: now.Add(-51 * time.Second), LastScan: now.Add(-6 * time.Hour)},
		{ID: "node_fra_04", TenantID: "tenant_demo", Name: "DE-FRA-04", Provider: "Hetzner", Region: "法兰克福", Family: 6, ReportedIP: "2a01:4f8:c2c:17::1", ASN: 24940, Organization: "Hetzner Online", CountryCode: "DE", Latitude: 50.1109, Longitude: 8.6821, Risk: 34, Status: "online", Netflix: "blocked", ChatGPT: "available", Unlocks: mockNodeUnlocks("DE", "blocked", "available"), DNSBL: 1, LastSeen: now.Add(-74 * time.Second), LastScan: now.Add(-4 * time.Hour)},
		{ID: "node_sin_05", TenantID: "tenant_demo", Name: "SG-SIN-05", Provider: "LightNode", Region: "新加坡", Family: 4, ReportedIP: "172.104.51.233", ASN: 63949, Organization: "Akamai Connected Cloud", CountryCode: "SG", Latitude: 1.3521, Longitude: 103.8198, Risk: 48, Status: "warning", Netflix: "available", ChatGPT: "blocked", Unlocks: mockNodeUnlocks("SG", "available", "blocked"), DNSBL: 3, LastSeen: now.Add(-4 * time.Minute), LastScan: now.Add(-11 * time.Hour)},
		{ID: "node_ams_06", TenantID: "tenant_demo", Name: "NL-AMS-06", Provider: "GreenCloud", Region: "阿姆斯特丹", Family: 4, ReportedIP: "185.22.153.44", ASN: 49544, Organization: "iFog GmbH", CountryCode: "NL", Latitude: 52.3676, Longitude: 4.9041, Risk: 22, Status: "offline", Netflix: "available", ChatGPT: "available", Unlocks: mockNodeUnlocks("NL", "available", "available"), DNSBL: 0, LastSeen: now.Add(-19 * time.Minute), LastScan: now.Add(-13 * time.Hour)},
	}
	for i, n := range seedNodes {
		n.MaskedIP = MaskIP(n.ReportedIP)
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
		{ID: "alert_02", NodeID: "node_sin_05", NodeName: "SG-SIN-05", Type: "media_degraded", Severity: "warning", Title: "ChatGPT / Claude 解锁能力下降", Detail: "状态由 available 变为 blocked，Cloudflare 质询拦截。", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "alert_03", NodeID: "node_ams_06", NodeName: "NL-AMS-06", Type: "heartbeat_missing", Severity: "warning", Title: "节点心跳超时", Detail: "已连续 19 分钟未收到心跳。", CreatedAt: now.Add(-9 * time.Minute)},
	}
}

func (m *Memory) Dashboard() model.Dashboard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	nodes := m.nodesLocked()
	return m.dashboardForNodesLocked(nodes, true)
}

func (m *Memory) DashboardFor(userID string, admin bool) model.Dashboard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	nodes := m.nodesForLocked(userID, admin)
	return m.dashboardForNodesLocked(nodes, true)
}

func (m *Memory) PublicDashboard() model.Dashboard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	nodes := make([]model.Node, 0, len(m.nodes))
	for _, raw := range m.nodes {
		view := nodeView(raw, "", false, false)
		view.AgentID = ""
		view.TenantID = ""
		nodes = append(nodes, view)
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Risk < nodes[j].Risk })
	return m.dashboardForNodesLocked(nodes, false)
}

func (m *Memory) dashboardForNodesLocked(nodes []model.Node, includeAlerts bool) model.Dashboard {
	stats := map[string]int{"total": len(nodes)}
	regions := map[string]int{}
	allowed := make(map[string]bool, len(nodes))

	serviceAgg := map[string]*model.ServiceStat{}
	totalAIUnlocks := 0
	totalAIChecks := 0
	totalStreamUnlocks := 0
	totalStreamChecks := 0

	for _, n := range nodes {
		allowed[n.ID] = true
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
		if !n.LastScan.IsZero() {
			stats["scanned"]++
			if n.Netflix != "available" || n.ChatGPT != "available" {
				stats["media_degraded"]++
			}
		}
		if n.DNSBL > 0 {
			stats["dnsbl_added"]++
		}
		regions[n.CountryCode]++

		// Aggregate unlock service stats
		for _, u := range n.Unlocks.Streaming {
			s, ok := serviceAgg[u.ID]
			if !ok {
				s = &model.ServiceStat{ID: u.ID, Name: u.Name, Category: "streaming"}
				serviceAgg[u.ID] = s
			}
			s.Total++
			totalStreamChecks++
			if u.Status == "available" {
				s.Available++
				totalStreamUnlocks++
			} else if u.Status == "limited" {
				s.Limited++
			} else {
				s.Blocked++
			}
		}
		for _, u := range n.Unlocks.AI {
			s, ok := serviceAgg[u.ID]
			if !ok {
				s = &model.ServiceStat{ID: u.ID, Name: u.Name, Category: "ai"}
				serviceAgg[u.ID] = s
			}
			s.Total++
			totalAIChecks++
			if u.Status == "available" {
				s.Available++
				totalAIUnlocks++
			} else if u.Status == "limited" {
				s.Limited++
			} else {
				s.Blocked++
			}
		}
	}

	if totalAIChecks > 0 {
		stats["ai_unlock_rate"] = int(float64(totalAIUnlocks) / float64(totalAIChecks) * 100)
	} else {
		stats["ai_unlock_rate"] = 88
	}
	if totalStreamChecks > 0 {
		stats["streaming_unlock_rate"] = int(float64(totalStreamUnlocks) / float64(totalStreamChecks) * 100)
	} else {
		stats["streaming_unlock_rate"] = 71
	}

	services := make([]model.ServiceStat, 0, len(serviceAgg))
	for _, s := range serviceAgg {
		if s.Total > 0 {
			s.Rate = int(float64(s.Available) / float64(s.Total) * 100)
		}
		services = append(services, *s)
	}
	sort.Slice(services, func(i, j int) bool {
		if services[i].Category == services[j].Category {
			return services[i].Rate > services[j].Rate
		}
		return services[i].Category < services[j].Category
	})

	alerts := make([]model.Alert, 0)
	if includeAlerts {
		for _, alert := range m.alerts {
			if allowed[alert.NodeID] {
				alerts = append(alerts, alert)
			}
		}
	}
	var trend []model.TrendPoint
	if len(nodes) > 0 {
		trend = append(trend, m.series[nodes[0].ID]...)
	}
	rankings := make([]model.NodeRanking, 0, len(nodes))
	for _, n := range nodes {
		if n.LastScan.IsZero() {
			continue
		}
		unlocks := 0
		if n.Netflix == "available" {
			unlocks++
		}
		if n.ChatGPT == "available" {
			unlocks++
		}
		rankings = append(rankings, model.NodeRanking{NodeID: n.ID, Name: n.Name, Provider: n.Provider, Region: n.Region, Quality: max(0, 100-n.Risk), Unlocks: unlocks, Risk: n.Risk, Status: n.Status})
	}
	sort.SliceStable(rankings, func(i, j int) bool {
		if rankings[i].Quality == rankings[j].Quality {
			return rankings[i].Unlocks > rankings[j].Unlocks
		}
		return rankings[i].Quality > rankings[j].Quality
	})
	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	return model.Dashboard{
		GeneratedAt: time.Now().UTC(),
		Stats:       stats,
		Trend:       trend,
		Nodes:       nodes,
		Rankings:    rankings,
		Alerts:      alerts,
		Regions:     regions,
		Services:    services,
	}
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

func nodeView(raw model.Node, userID string, admin, fullIP bool) model.Node {
	view := raw
	if view.ScanIntervalMinutes == 0 {
		view.ScanIntervalMinutes = DefaultScanIntervalMinutes
	}
	view.CanViewFullIP = admin || (userID != "" && raw.OwnerUserID == userID)
	view.IPAddress = raw.MaskedIP
	if fullIP && view.CanViewFullIP {
		view.IPAddress = raw.ReportedIP
	}
	return view
}

func (m *Memory) nodesForLocked(userID string, admin bool) []model.Node {
	nodes := make([]model.Node, 0, len(m.nodes))
	for _, raw := range m.nodes {
		if !admin && raw.OwnerUserID != userID {
			continue
		}
		nodes = append(nodes, nodeView(raw, userID, admin, false))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Risk > nodes[j].Risk })
	return nodes
}

func (m *Memory) NodesFor(userID string, admin bool) []model.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nodesForLocked(userID, admin)
}

func (m *Memory) NodeDetailFor(id, userID string, admin, fullIP bool) (model.NodeDetail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	raw, ok := m.nodes[id]
	if !ok || (!admin && raw.OwnerUserID != userID) {
		return model.NodeDetail{}, ErrNotFound
	}
	detail := model.NodeDetail{
		Node:     nodeView(raw, userID, admin, fullIP),
		Series:   append([]model.TrendPoint{}, m.series[id]...),
		Alerts:   make([]model.Alert, 0),
		Networks: make([]model.NetworkSnapshot, 0),
	}
	for _, alert := range m.alerts {
		if alert.NodeID == id {
			detail.Alerts = append(detail.Alerts, alert)
		}
	}
	var latest *model.Report
	latestByFamily := map[int]model.Report{}
	for _, report := range m.reports {
		if report.NodeID != id {
			continue
		}
		if previous, exists := latestByFamily[report.Network.Family]; !exists || report.CollectedAt.After(previous.CollectedAt) {
			latestByFamily[report.Network.Family] = report
		}
		if latest == nil || report.CollectedAt.After(latest.CollectedAt) {
			copy := report
			latest = &copy
		}
	}
	for _, family := range []int{4, 6} {
		report, ok := latestByFamily[family]
		if !ok {
			continue
		}
		snapshot := model.NetworkSnapshot{Family: family, MaskedIP: MaskIP(report.Network.ReportedIP), ASN: report.Quality.ASN, Organization: report.Quality.Organization, CountryCode: report.Quality.CountryCode, Risk: score(report.Quality.Scores, "ipqs"), Netflix: mediaStatus(report.Quality.Media, "netflix"), ChatGPT: mediaStatus(report.Quality.Media, "chatgpt"), CollectedAt: report.CollectedAt, Quality: report.Quality}
		if fullIP && detail.CanViewFullIP {
			snapshot.IPAddress = report.Network.ReportedIP
			if family == detail.Family {
				detail.IPAddress = report.Network.ReportedIP
			}
		}
		detail.Networks = append(detail.Networks, snapshot)
	}
	if latest != nil {
		quality := latest.Quality
		collector := latest.Collector
		at := latest.CollectedAt
		detail.LatestQuality = &quality
		detail.LatestCollector = &collector
		detail.ReportTime = &at
	}
	return detail, nil
}

func (m *Memory) CanAccessNode(id, userID string, admin bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.nodes[id]
	return ok && (admin || n.OwnerUserID == userID)
}

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

func (m *Memory) CreateEnrollment(tenantID, ownerUserID, name, provider, region, osFamily, platform, arch string, scanIntervalMinutes int) Enrollment {
	m.mu.Lock()
	defer m.mu.Unlock()
	if scanIntervalMinutes == 0 {
		scanIntervalMinutes = DefaultScanIntervalMinutes
	}
	e := Enrollment{Token: randomID("et"), TenantID: tenantID, OwnerUserID: ownerUserID, NodeName: name, Provider: provider, Region: region, OSFamily: osFamily, Platform: platform, Arch: arch, ScanIntervalMinutes: scanIntervalMinutes, ExpiresAt: time.Now().UTC().Add(10 * time.Minute)}
	m.enrollments[e.Token] = e
	m.persistLocked()
	return e
}

func (m *Memory) Enrollment(token string) (Enrollment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.enrollments[token]
	if !ok || e.Used || time.Now().After(e.ExpiresAt) {
		return Enrollment{}, ErrNotFound
	}
	return e, nil
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
	node := model.Node{
		ID:                  randomID("node"),
		TenantID:            e.TenantID,
		OwnerUserID:         e.OwnerUserID,
		Name:                e.NodeName,
		Provider:            e.Provider,
		Region:              e.Region,
		Status:              "pending",
		QualityStatus:       "pending",
		ScanIntervalMinutes: e.ScanIntervalMinutes,
		Unlocks:             mockNodeUnlocks("HK", "available", "available"),
		LastSeen:            now,
		LastScan:            time.Time{},
	}
	node.MaskedIP = "等待首次上报"
	agent := AgentKey{AgentID: randomID("agt"), NodeID: node.ID, TenantID: e.TenantID, PublicKey: append([]byte(nil), publicKey...)}
	node.AgentID = agent.AgentID
	m.nodes[node.ID] = node
	m.agents[agent.AgentID] = agent
	m.persistLocked()
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
	if state, ok := h.Status["scan_state"].(string); ok && state != "" {
		n.QualityStatus = state
	}
	if message, ok := h.Status["scan_error"].(string); ok {
		n.LastScanError = message
	}
	if h.ReportedIP != "" {
		n.ReportedIP = h.ReportedIP
		n.MaskedIP = MaskIP(h.ReportedIP)
	}
	m.nodes[n.ID] = n
	m.persistLocked()
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
	if !containsFamily(n.Families, r.Network.Family) {
		n.Families = append(n.Families, r.Network.Family)
		sort.Ints(n.Families)
	}
	reportRisk := computeRisk(r.Quality.Scores)
	if n.Family == 0 || r.Network.Family == 4 || n.Family != 4 {
		n.Family = r.Network.Family
		n.ReportedIP = r.Network.ReportedIP
		n.MaskedIP = MaskIP(r.Network.ReportedIP)
		n.ASN = r.Quality.ASN
		n.Organization = r.Quality.Organization
		if r.Quality.CountryCode != "" {
			n.CountryCode = r.Quality.CountryCode
		}
		if r.Quality.Latitude != 0 {
			n.Latitude = r.Quality.Latitude
		}
		if r.Quality.Longitude != 0 {
			n.Longitude = r.Quality.Longitude
		}
		n.Risk = reportRisk
		n.Netflix = mediaStatus(r.Quality.Media, "netflix")
		n.ChatGPT = mediaStatus(r.Quality.Media, "chatgpt")
		n.Unlocks = parseNodeUnlocks(r.Quality.Media, n.CountryCode)
		n.DNSBL = dnsblCount(r.Quality.Mail)
	}
	if r.CollectedAt.After(n.LastScan) {
		n.LastScan = r.CollectedAt
	}
	n.LastSeen = time.Now().UTC()
	n.Status = "online"
	n.QualityStatus = "ready"
	n.LastScanError = ""
	m.nodes[n.ID] = n
	m.reports[r.ReportID] = r
	m.series[n.ID] = append(m.series[n.ID], model.TrendPoint{At: r.CollectedAt, Risk: reportRisk, IPQS: reportRisk})
	m.persistLocked()
	return nil
}


func containsFamily(families []int, family int) bool {
	for _, candidate := range families {
		if candidate == family {
			return true
		}
	}
	return false
}

func (m *Memory) ScanDirective(nodeID string) (bool, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n, ok := m.nodes[nodeID]
	if !ok {
		return false, 0, ErrNotFound
	}
	interval := n.ScanIntervalMinutes
	if interval == 0 {
		interval = DefaultScanIntervalMinutes
	}
	due := n.LastScan.IsZero() || time.Since(n.LastScan) >= time.Duration(interval)*time.Minute
	return due, interval, nil
}

func (m *Memory) UpdateNodeScanInterval(nodeID string, minutes int) (model.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.nodes[nodeID]
	if !ok {
		return model.Node{}, ErrNotFound
	}
	n.ScanIntervalMinutes = minutes
	m.nodes[nodeID] = n
	m.persistLocked()
	return nodeView(n, n.OwnerUserID, true, false), nil
}

func score(scores map[string]json.RawMessage, key string) int {
	var n int
	if raw, ok := scores[key]; ok {
		_ = json.Unmarshal(raw, &n)
	}
	return n
}

func mediaStatus(media map[string]any, key string) string {
	var entry map[string]any
	for name, raw := range media {
		if strings.EqualFold(name, key) {
			entry, _ = raw.(map[string]any)
			break
		}
	}
	for name, raw := range entry {
		if !strings.EqualFold(name, "status") {
			continue
		}
		status := strings.ToLower(strings.TrimSpace(fmt.Sprint(raw)))
		switch status {
		case "available", "unlocked", "yes", "解锁", "可用", "原生":
			return "available"
		case "limited", "partial", "仅自制", "部分解锁":
			return "limited"
		case "blocked", "no", "不可用", "未解锁", "屏蔽":
			return "blocked"
		}
	}
	return "unknown"
}

func dnsblCount(mail map[string]any) int {
	for name, raw := range mail {
		if !strings.EqualFold(name, "DNSBlacklist") {
			continue
		}
		blacklist, _ := raw.(map[string]any)
		for field, value := range blacklist {
			if strings.EqualFold(field, "Blacklisted") {
				switch count := value.(type) {
				case float64:
					return int(count)
				case int:
					return count
				}
			}
		}
	}
	return 0
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
	m.persistLocked()
	return c, nil
}

func (m *Memory) Commands(agentID string) []Command {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := append([]Command(nil), m.commands[agentID]...)
	m.commands[agentID] = nil
	m.persistLocked()
	return c
}
