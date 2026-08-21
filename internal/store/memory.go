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
	NodeID              string
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
	tasks               map[string][]model.TaskLog
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
		commands: make(map[string][]Command), tasks: make(map[string][]model.TaskLog),
		users: make(map[string]UserAccount), usernames: make(map[string]string),
		sessions: make(map[string]Session), passwordResets: make(map[string]PasswordReset),
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
		statusRaw := mediaString(entry, "Status")
		regionRaw := cleanMediaRegion(mediaString(entry, "Region"))
		if regionRaw == "" {
			regionRaw = defaultRegion
		}
		typeRaw := mediaString(entry, "Type")

		status := "unknown"
		switch strings.ToLower(statusRaw) {
		case "available", "unlocked", "yes", "解锁", "可用", "原生", "允许", "true":
			status = "available"
		case "limited", "partial", "仅自制", "部分解锁", "限制", "待支持", "中国", "cn", "china", "送中":
			status = "limited"
		case "blocked", "no", "不可用", "未解锁", "屏蔽", "质询", "禁会员", "false":
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
		case "bahamut", "gamer":
			serviceID = "bahamut"
			serviceName = "巴哈姆特动画疯"
			category = "streaming"
		case "abema", "abematv":
			serviceID = "abema"
			serviceName = "AbemaTV"
			category = "streaming"
		case "dazn":
			serviceID = "dazn"
			serviceName = "DAZN"
			category = "streaming"
		case "tvb", "tvbanywhere":
			serviceID = "tvb"
			serviceName = "TVB Anywhere"
			category = "streaming"
		case "appletv", "apple_tv", "appletv+":
			serviceID = "appletv"
			serviceName = "Apple TV+"
			category = "streaming"
		}

		qualityDesc := statusRaw
		if customQuality := mediaString(entry, "Quality"); customQuality != "" {
			qualityDesc = customQuality
		} else if typeRaw != "" && typeRaw != statusRaw {
			qualityDesc = typeRaw + " " + statusRaw
		}
		if serviceID == "youtube" && (strings.EqualFold(regionRaw, "CN") || strings.Contains(statusRaw, "中国") || strings.Contains(statusRaw, "送中")) {
			status = "limited"
			regionRaw = "CN"
			qualityDesc = "送中（中国区）"
		}

		latencyMs := mediaInt(entry, "LatencyMs")

		detailDesc := fmt.Sprintf("%s · %s", serviceName, qualityDesc)
		if customDetail := mediaString(entry, "Detail"); customDetail != "" {
			detailDesc = customDetail
		}
		if serviceID == "youtube" && regionRaw == "CN" {
			detailDesc = "YouTube 可达，但内容地区被识别为中国大陆（送中）"
		}

		info := model.UnlockInfo{
			ID:        serviceID,
			Name:      serviceName,
			Category:  category,
			Status:    status,
			Region:    regionRaw,
			Quality:   qualityDesc,
			LatencyMs: latencyMs,
			Detail:    detailDesc,
			CheckedAt: mediaString(entry, "CheckedAt"),
		}

		if category == "ai" {
			if current, exists := ai[serviceID]; !exists || preferUnlockInfo(info, current) {
				ai[serviceID] = info
			}
		} else {
			if current, exists := stream[serviceID]; !exists || preferUnlockInfo(info, current) {
				stream[serviceID] = info
			}
		}
	}

	return model.NodeUnlocks{Streaming: stream, AI: ai}
}

func mediaString(entry map[string]any, wanted string) string {
	for key, value := range entry {
		if strings.EqualFold(key, wanted) {
			return cleanDetectedText(fmt.Sprint(value))
		}
	}
	return ""
}

func mediaInt(entry map[string]any, wanted string) int {
	for key, value := range entry {
		if !strings.EqualFold(key, wanted) {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed)
		case int:
			return typed
		case json.Number:
			parsed, _ := strconv.Atoi(typed.String())
			return parsed
		case string:
			parsed, _ := strconv.Atoi(strings.TrimSpace(typed))
			return parsed
		}
	}
	return 0
}

func cleanMediaRegion(value string) string {
	value = strings.Trim(strings.TrimSpace(value), "[]")
	if len(value) == 2 {
		return strings.ToUpper(value)
	}
	return value
}

func preferUnlockInfo(candidate, current model.UnlockInfo) bool {
	if candidate.ID == "youtube" {
		candidateCN := candidate.Region == "CN" || strings.Contains(candidate.Quality, "送中")
		currentCN := current.Region == "CN" || strings.Contains(current.Quality, "送中")
		if candidateCN != currentCN {
			return candidateCN
		}
	}
	priority := func(status string) int {
		switch status {
		case "available", "limited", "blocked":
			return 2
		default:
			return 1
		}
	}
	return priority(candidate.Status) > priority(current.Status)
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

func fallbackGeoCoordinates(n *model.Node) {

	if n.CountryCode == "" {
		if n.Unlocks.Streaming != nil && n.Unlocks.Streaming["netflix"].Region != "" {
			n.CountryCode = n.Unlocks.Streaming["netflix"].Region
		} else if n.Unlocks.AI != nil && n.Unlocks.AI["chatgpt"].Region != "" {
			n.CountryCode = n.Unlocks.AI["chatgpt"].Region
		} else {
			regUpper := strings.ToUpper(n.Region)
			if strings.Contains(regUpper, "HK") || strings.Contains(n.Region, "香港") {
				n.CountryCode = "HK"
			} else if strings.Contains(regUpper, "TW") || strings.Contains(regUpper, "TAIWAN") || strings.Contains(n.Region, "台湾") {
				n.CountryCode = "TW"
			} else if strings.Contains(regUpper, "JP") || strings.Contains(n.Region, "日本") || strings.Contains(n.Region, "东京") {
				n.CountryCode = "JP"
			} else if strings.Contains(regUpper, "SG") || strings.Contains(n.Region, "新加坡") {
				n.CountryCode = "SG"
			} else if strings.Contains(regUpper, "US") || strings.Contains(n.Region, "美国") || strings.Contains(n.Region, "洛杉矶") {
				n.CountryCode = "US"
			} else if strings.Contains(regUpper, "DE") || strings.Contains(n.Region, "德国") || strings.Contains(n.Region, "法兰克福") {
				n.CountryCode = "DE"
			} else if strings.Contains(regUpper, "NL") || strings.Contains(n.Region, "荷兰") || strings.Contains(n.Region, "阿姆斯特丹") {
				n.CountryCode = "NL"
			}
		}
	}
	if n.Latitude == 0 && n.Longitude == 0 {
		switch n.CountryCode {
		case "HK":
			n.Latitude, n.Longitude = 22.3193, 114.1694
		case "TW":
			n.Latitude, n.Longitude = 25.0330, 121.5654
		case "JP":
			n.Latitude, n.Longitude = 35.6762, 139.6503
		case "SG":
			n.Latitude, n.Longitude = 1.3521, 103.8198
		case "US":
			n.Latitude, n.Longitude = 37.7749, -122.4194
		case "DE":
			n.Latitude, n.Longitude = 50.1109, 8.6821
		case "NL":
			n.Latitude, n.Longitude = 52.3676, 4.9041
		case "GB", "UK":
			n.Latitude, n.Longitude = 51.5074, -0.1278
		}
	}
}

func (m *Memory) seed() {

	now := time.Now().UTC()
	seedNodes := []model.Node{
		{ID: "node_hkg_01", TenantID: "tenant_demo", Name: "HK-CMI-01", Provider: "DMIT", Region: "香港", Family: 4, Families: []int{4, 6}, ReportedIP: "103.145.12.81", MaskedIPv4: "103.145.12.*", MaskedIPv6: "2402:4e00:1000::*", ASN: 906, Organization: "DMIT Cloud Services", CountryCode: "HK", UsageType: "机房", IPType: "原生", Latitude: 22.3193, Longitude: 114.1694, Risk: 18, Status: "online", Netflix: "available", ChatGPT: "available", Unlocks: model.NodeUnlocks{}, DNSBL: 0, LastSeen: now.Add(-42 * time.Second), LastScan: now.Add(-2 * time.Hour)},
		{ID: "node_lax_02", TenantID: "tenant_demo", Name: "US-LAX-02", Provider: "RackNerd", Region: "洛杉矶", Family: 4, Families: []int{4}, ReportedIP: "198.51.100.27", MaskedIPv4: "198.51.100.*", ASN: 62240, Organization: "Clouvider Limited", CountryCode: "US", UsageType: "机房", IPType: "广播", Latitude: 34.0522, Longitude: -118.2437, Risk: 72, Status: "alert", Netflix: "limited", ChatGPT: "available", Unlocks: model.NodeUnlocks{}, DNSBL: 8, IPChanged: true, LastSeen: now.Add(-2 * time.Minute), LastScan: now.Add(-28 * time.Minute)},
		{ID: "node_nrt_03", TenantID: "tenant_demo", Name: "JP-NRT-03", Provider: "Vultr", Region: "东京", Family: 4, Families: []int{4, 6}, ReportedIP: "45.76.201.19", MaskedIPv4: "45.76.201.*", MaskedIPv6: "2001:19f0:7001::*", ASN: 20473, Organization: "The Constant Company", CountryCode: "JP", UsageType: "商宽", IPType: "原生", Latitude: 35.6762, Longitude: 139.6503, Risk: 11, Status: "online", Netflix: "available", ChatGPT: "available", Unlocks: model.NodeUnlocks{}, DNSBL: 0, LastSeen: now.Add(-51 * time.Second), LastScan: now.Add(-6 * time.Hour)},
		{ID: "node_fra_04", TenantID: "tenant_demo", Name: "DE-FRA-04", Provider: "Hetzner", Region: "法兰克福", Family: 6, Families: []int{6}, ReportedIP: "2a01:4f8:c2c:17::1", MaskedIPv6: "2a01:4f8:c2c:17::*", ASN: 24940, Organization: "Hetzner Online", CountryCode: "DE", UsageType: "机房", IPType: "原生", Latitude: 50.1109, Longitude: 8.6821, Risk: 34, Status: "online", Netflix: "blocked", ChatGPT: "available", Unlocks: model.NodeUnlocks{}, DNSBL: 1, LastSeen: now.Add(-74 * time.Second), LastScan: now.Add(-4 * time.Hour)},
		{ID: "node_sin_05", TenantID: "tenant_demo", Name: "SG-SIN-05", Provider: "LightNode", Region: "新加坡", Family: 4, Families: []int{4, 6}, ReportedIP: "172.104.51.233", MaskedIPv4: "172.104.51.*", MaskedIPv6: "2600:3c01::*", Warp6: true, IsWarp: true, ASN: 63949, Organization: "Akamai Connected Cloud", CountryCode: "SG", UsageType: "家宽", IPType: "原生", Latitude: 1.3521, Longitude: 103.8198, Risk: 48, Status: "warning", Netflix: "available", ChatGPT: "blocked", Unlocks: model.NodeUnlocks{}, DNSBL: 3, LastSeen: now.Add(-4 * time.Minute), LastScan: now.Add(-11 * time.Hour)},
		{ID: "node_ams_06", TenantID: "tenant_demo", Name: "NL-AMS-06", Provider: "GreenCloud", Region: "阿姆斯特丹", Family: 4, Families: []int{4}, ReportedIP: "185.22.153.44", MaskedIPv4: "185.22.153.*", ASN: 49544, Organization: "iFog GmbH", CountryCode: "NL", UsageType: "机房", IPType: "原生", Latitude: 52.3676, Longitude: 4.9041, Risk: 22, Status: "offline", Netflix: "available", ChatGPT: "available", Unlocks: model.NodeUnlocks{}, DNSBL: 0, LastSeen: now.Add(-19 * time.Minute), LastScan: now.Add(-13 * time.Hour)},
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
		view := m.nodeView(raw, "", false, false)
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

	for i := range nodes {
		fallbackGeoCoordinates(&nodes[i])
		n := nodes[i]
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
		if n.CountryCode != "" {
			regions[n.CountryCode]++
		}

		// Aggregate unlock service stats
		for _, u := range n.Unlocks.Streaming {
			if u.Status != "available" && u.Status != "limited" && u.Status != "blocked" {
				continue
			}
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
			if u.Status != "available" && u.Status != "limited" && u.Status != "blocked" {
				continue
			}
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
		stats["ai_unlock_rate"] = 100
	}
	if totalStreamChecks > 0 {
		stats["streaming_unlock_rate"] = int(float64(totalStreamUnlocks) / float64(totalStreamChecks) * 100)
	} else {
		stats["streaming_unlock_rate"] = 100
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

func (m *Memory) nodeView(raw model.Node, userID string, admin, fullIP bool) model.Node {
	view := raw
	if view.ScanIntervalMinutes == 0 {
		view.ScanIntervalMinutes = DefaultScanIntervalMinutes
	}
	view.CanViewFullIP = admin || (userID != "" && raw.OwnerUserID == userID)
	view.IPAddress = raw.MaskedIP
	if fullIP && view.CanViewFullIP {
		view.IPAddress = raw.ReportedIP
	}
	if taskList, ok := m.tasks[raw.ID]; ok && len(taskList) > 0 {
		last := taskList[len(taskList)-1]
		view.LastTask = &last
	}
	return view
}

func (m *Memory) nodesForLocked(userID string, admin bool) []model.Node {
	nodes := make([]model.Node, 0, len(m.nodes))
	for _, raw := range m.nodes {
		if !admin && raw.OwnerUserID != userID {
			continue
		}
		nodes = append(nodes, m.nodeView(raw, userID, admin, false))
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Risk > nodes[j].Risk })
	return nodes
}

func (m *Memory) NodesFor(userID string, admin bool) []model.Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nodesForLocked(userID, admin)
}

func (m *Memory) NodeTasks(nodeID, userID string, admin bool) ([]model.TaskLog, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	raw, ok := m.nodes[nodeID]
	if !ok || (!admin && raw.OwnerUserID != userID) {
		return nil, ErrNotFound
	}
	return append([]model.TaskLog(nil), m.tasks[nodeID]...), nil
}

func (m *Memory) DeleteNode(nodeID, userID string, admin, force bool) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, ok := m.nodes[nodeID]
	if !ok || (!admin && raw.OwnerUserID != userID) {
		return 0, ErrNotFound
	}

	agentIDs := make([]string, 0, 1)
	for agentID, agent := range m.agents {
		if agent.NodeID == nodeID {
			agentIDs = append(agentIDs, agentID)
		}
	}

	delete(m.nodes, nodeID)
	delete(m.series, nodeID)
	delete(m.tasks, nodeID)
	for reportID, report := range m.reports {
		if report.NodeID == nodeID {
			delete(m.reports, reportID)
		}
	}
	for token, enrollment := range m.enrollments {
		if enrollment.NodeID == nodeID {
			delete(m.enrollments, token)
		}
	}
	alerts := m.alerts[:0]
	for _, alert := range m.alerts {
		if alert.NodeID != nodeID {
			alerts = append(alerts, alert)
		}
	}
	m.alerts = alerts

	now := time.Now().UTC()
	for _, agentID := range agentIDs {
		if force {
			m.removeAgentLocked(agentID)
			continue
		}
		m.commands[agentID] = []Command{{ID: randomID("cmd"), Type: "uninstall", CreatedAt: now}}
	}
	m.persistLocked()
	return len(agentIDs), nil
}

func (m *Memory) CompleteAgentUninstall(agentID, nodeID, commandID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	agent, ok := m.agents[agentID]
	if !ok || agent.NodeID != nodeID {
		return ErrNotFound
	}
	queued := false
	for _, command := range m.commands[agentID] {
		if command.ID == commandID && command.Type == "uninstall" {
			queued = true
			break
		}
	}
	if !queued {
		return ErrNotFound
	}
	m.removeAgentLocked(agentID)
	m.persistLocked()
	return nil
}

func (m *Memory) removeAgentLocked(agentID string) {
	delete(m.agents, agentID)
	delete(m.commands, agentID)
	prefix := agentID + ":"
	for key := range m.nonces {
		if strings.HasPrefix(key, prefix) {
			delete(m.nonces, key)
		}
	}
}

func (m *Memory) NodeDetailFor(id, userID string, admin, fullIP bool) (model.NodeDetail, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	raw, ok := m.nodes[id]
	if !ok || (!admin && raw.OwnerUserID != userID) {
		return model.NodeDetail{}, ErrNotFound
	}
	detail := model.NodeDetail{
		Node:     m.nodeView(raw, userID, admin, fullIP),
		Series:   append([]model.TrendPoint{}, m.series[id]...),
		Alerts:   make([]model.Alert, 0),
		Tasks:    append([]model.TaskLog(nil), m.tasks[id]...),
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

func (m *Memory) CreateReenrollment(nodeID, userID string, admin bool) (Enrollment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, ok := m.nodes[nodeID]
	if !ok || (!admin && raw.OwnerUserID != userID) {
		return Enrollment{}, ErrNotFound
	}
	interval := raw.ScanIntervalMinutes
	if interval == 0 {
		interval = DefaultScanIntervalMinutes
	}
	e := Enrollment{
		Token:               randomID("et"),
		TenantID:            raw.TenantID,
		OwnerUserID:         raw.OwnerUserID,
		NodeID:              raw.ID,
		NodeName:            raw.Name,
		Provider:            raw.Provider,
		Region:              raw.Region,
		ScanIntervalMinutes: interval,
		ExpiresAt:           time.Now().UTC().Add(30 * time.Minute),
	}
	m.enrollments[e.Token] = e
	m.persistLocked()
	return e, nil
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

	// If re-enrolling an existing node
	if e.NodeID != "" {
		if existing, exists := m.nodes[e.NodeID]; exists {
			agent := AgentKey{AgentID: randomID("agt"), NodeID: existing.ID, TenantID: existing.TenantID, PublicKey: append([]byte(nil), publicKey...)}
			existing.AgentID = agent.AgentID
			existing.Status = "online"
			existing.LastSeen = now
			m.nodes[existing.ID] = existing
			m.agents[agent.AgentID] = agent
			m.persistLocked()
			return existing, agent, nil
		}
	}

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
		Unlocks:             model.NodeUnlocks{},
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

	// Update task log
	if taskList, ok := m.tasks[n.ID]; ok && len(taskList) > 0 {
		last := &taskList[len(taskList)-1]
		if n.QualityStatus == "scanning" && (last.Status == "pending" || last.Status == "running") {
			last.Status = "running"
			last.Message = "探针已接单，正在并发探测 20+ 款 AI 与流媒体服务并核验 IP 纯净度..."
			last.UpdatedAt = time.Now().UTC()
			n.LastTask = last
		} else if n.QualityStatus == "failed" && (last.Status == "pending" || last.Status == "running") {
			last.Status = "failed"
			last.Message = "探测任务异常"
			last.Error = n.LastScanError
			last.UpdatedAt = time.Now().UTC()
			n.LastTask = last
		}
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
	validIP := validReportedIP(r.Network.ReportedIP, r.Network.Family)
	if validIP && !containsFamily(n.Families, r.Network.Family) {
		n.Families = append(n.Families, r.Network.Family)
		sort.Ints(n.Families)
	}
	reportRisk := computeRisk(r.Quality.Scores)
	if validIP && (n.Family == 0 || r.Network.Family == 4 || n.Family != 4) {
		n.Family = r.Network.Family
		n.ReportedIP = r.Network.ReportedIP
		n.MaskedIP = MaskIP(r.Network.ReportedIP)
		applyReportIdentity(&n, r)
	}

	if validIP && r.Network.Family == 4 {
		n.MaskedIPv4 = MaskIP(r.Network.ReportedIP)
		n.Warp4 = reportIsWARP(r)
	} else if validIP && r.Network.Family == 6 {
		n.MaskedIPv6 = MaskIP(r.Network.ReportedIP)
		n.Warp6 = reportIsWARP(r)
	}
	n.IsWarp = n.Warp4 || n.Warp6

	if r.CollectedAt.After(n.LastScan) {
		n.LastScan = r.CollectedAt
	}
	n.LastSeen = time.Now().UTC()
	n.Status = "online"
	n.QualityStatus = "ready"
	n.LastScanError = ""

	// Update task log
	if taskList, ok := m.tasks[n.ID]; ok && len(taskList) > 0 {
		last := &taskList[len(taskList)-1]
		if last.Status == "pending" || last.Status == "running" {
			last.Status = "completed"
			last.Message = fmt.Sprintf("20+ 项服务深度探测完成并成功上报 (Report: %s)", r.ReportID)
			last.UpdatedAt = time.Now().UTC()
			n.LastTask = last
		}
	}

	m.nodes[n.ID] = n
	m.reports[r.ReportID] = r
	m.series[n.ID] = append(m.series[n.ID], model.TrendPoint{At: r.CollectedAt, Risk: reportRisk, IPQS: reportRisk})
	m.persistLocked()
	return nil
}

func applyReportIdentity(node *model.Node, report model.Report) {
	if node == nil {
		return
	}
	quality := report.Quality
	if quality.ASN > 0 {
		node.ASN = quality.ASN
	}
	if organization := cleanDetectedText(quality.Organization); organization != "" {
		node.Organization = organization
	}
	if country := normalizeDetectedCountry(quality.CountryCode); country != "" {
		node.CountryCode = country
	}
	if quality.Latitude != 0 {
		node.Latitude = quality.Latitude
	}
	if quality.Longitude != 0 {
		node.Longitude = quality.Longitude
	}
	if usage := normalizeDetectedUsage(quality.UsageType); usage != "" {
		node.UsageType = usage
	}
	if ipType := normalizeDetectedIPType(quality.IPType); ipType != "" {
		node.IPType = ipType
	}
	node.Risk = computeRisk(quality.Scores)
	node.Netflix = mediaStatus(quality.Media, "netflix")
	node.ChatGPT = mediaStatus(quality.Media, "chatgpt")
	node.Unlocks = parseNodeUnlocks(quality.Media, node.CountryCode)
	node.DNSBL = dnsblCount(quality.Mail)
}

func validReportedIP(value string, family int) bool {
	address, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return (family == 4 && address.Is4()) || (family == 6 && address.Is6())
}

func reportIsWARP(report model.Report) bool {
	if report.Network.IsWARP || strings.EqualFold(strings.TrimSpace(report.Quality.UsageType), "warp") {
		return true
	}
	for key, value := range report.Quality.Factors {
		if strings.EqualFold(key, "warp") && detectedTruthy(value) {
			return true
		}
	}
	return false
}

func detectedTruthy(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		normalized := strings.ToLower(strings.TrimSpace(typed))
		return normalized == "true" || normalized == "yes" || normalized == "on" || normalized == "plus"
	case float64:
		return typed != 0
	case int:
		return typed != 0
	case map[string]any:
		for _, nested := range typed {
			if detectedTruthy(nested) {
				return true
			}
		}
	}
	return false
}

func normalizeDetectedUsage(value string) string {
	lower := strings.ToLower(cleanDetectedText(value))
	switch {
	case lower == "dch", strings.Contains(lower, "datacenter"), strings.Contains(lower, "data center"),
		strings.Contains(lower, "hosting"), strings.Contains(lower, "transit"),
		strings.Contains(lower, "server"), strings.Contains(lower, "cloud"),
		strings.Contains(lower, "cdn"), strings.Contains(lower, "机房"), strings.Contains(lower, "数据中心"):
		return "机房"
	case strings.Contains(lower, "residential"), strings.Contains(lower, "fixed line"),
		strings.Contains(lower, "line isp"), strings.Contains(lower, "broadband"),
		strings.Contains(lower, "mobile"), strings.Contains(lower, "cellular"),
		strings.Contains(lower, "home"), lower == "isp", lower == "mob", strings.Contains(lower, "家宽"),
		strings.Contains(lower, "住宅"), strings.Contains(lower, "移动"):
		return "家宽"
	case lower == "com", strings.Contains(lower, "business"), strings.Contains(lower, "commercial"),
		strings.Contains(lower, "company"), strings.Contains(lower, "education"),
		strings.Contains(lower, "government"), strings.Contains(lower, "organization"),
		strings.Contains(lower, "banking"), strings.Contains(lower, "商宽"),
		strings.Contains(lower, "商业"), strings.Contains(lower, "教育"),
		strings.Contains(lower, "政府"), strings.Contains(lower, "组织"):
		return "商宽"
	default:
		return ""
	}
}

func normalizeDetectedIPType(value string) string {
	lower := strings.ToLower(cleanDetectedText(value))
	switch {
	case strings.Contains(lower, "原生"), strings.Contains(lower, "native"), strings.Contains(lower, "geo-consistent"):
		return "原生"
	case strings.Contains(lower, "广播"), strings.Contains(lower, "broadcast"), strings.Contains(lower, "geo-discrepant"):
		return "广播"
	default:
		return ""
	}
}

func normalizeDetectedCountry(value string) string {
	value = strings.ToUpper(strings.Trim(strings.TrimSpace(cleanDetectedText(value)), "[]"))
	if len(value) != 2 {
		return ""
	}
	return value
}

func cleanDetectedText(value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "", "null", "<nil>", "n/a", "unknown":
		return ""
	default:
		return value
	}
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
	return m.nodeView(n, n.OwnerUserID, true, false), nil
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
		case "limited", "partial", "仅自制", "部分解锁", "待支持", "中国", "cn", "china", "送中":
			return "limited"
		case "blocked", "no", "不可用", "未解锁", "屏蔽", "禁会员":
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
	now := time.Now().UTC()
	c := Command{ID: randomID("cmd"), Type: "scan", CreatedAt: now}
	m.commands[n.AgentID] = append(m.commands[n.AgentID], c)
	n.QualityStatus = "scanning"
	task := model.TaskLog{
		ID:        randomID("task"),
		NodeID:    nodeID,
		Type:      "scan",
		Status:    "pending",
		Message:   "已下发 20+ 项 AI 与流媒体深度探测任务，等待探针接单...",
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.tasks[nodeID] = append(m.tasks[nodeID], task)
	n.LastTask = &task
	m.nodes[nodeID] = n
	m.persistLocked()
	return c, nil
}

func (m *Memory) Commands(agentID string) []Command {
	m.mu.Lock()
	defer m.mu.Unlock()
	c := append([]Command(nil), m.commands[agentID]...)
	remaining := m.commands[agentID][:0]
	for _, command := range m.commands[agentID] {
		if command.Type == "uninstall" {
			remaining = append(remaining, command)
		}
	}
	m.commands[agentID] = remaining
	m.persistLocked()
	return c
}
