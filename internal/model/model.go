package model

import (
	"encoding/json"
	"time"
)

type UnlockInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"` // "streaming" | "ai" | "social"
	Status    string `json:"status"`   // "available" | "limited" | "blocked" | "unknown"
	Region    string `json:"region,omitempty"`
	Quality   string `json:"quality,omitempty"`
	LatencyMs int    `json:"latency_ms,omitempty"`
	Detail    string `json:"detail,omitempty"`
	CheckedAt string `json:"checked_at,omitempty"`
}

type NodeUnlocks struct {
	Streaming map[string]UnlockInfo `json:"streaming"`
	AI        map[string]UnlockInfo `json:"ai"`
}

type ServiceStat struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Total     int    `json:"total"`
	Available int    `json:"available"`
	Limited   int    `json:"limited"`
	Blocked   int    `json:"blocked"`
	Rate      int    `json:"rate"`
}

type Node struct {
	ID                  string      `json:"id"`
	TenantID            string      `json:"tenant_id"`
	OwnerUserID         string      `json:"-"`
	AgentID             string      `json:"agent_id,omitempty"`
	Name                string      `json:"name"`
	Provider            string      `json:"provider"`
	Region              string      `json:"region"`
	Family              int         `json:"family"`
	Families            []int       `json:"families,omitempty"`
	ReportedIP          string      `json:"-"`
	MaskedIP            string      `json:"masked_ip"`
	IPAddress           string      `json:"ip_address,omitempty"`
	CanViewFullIP       bool        `json:"can_view_full_ip"`
	ASN                 int64       `json:"asn"`
	Organization        string      `json:"organization"`
	CountryCode         string      `json:"country_code"`
	Latitude            float64     `json:"latitude,omitempty"`
	Longitude           float64     `json:"longitude,omitempty"`
	Risk                int         `json:"risk"`
	Status              string      `json:"status"`
	Netflix             string      `json:"netflix"`
	ChatGPT             string      `json:"chatgpt"`
	Unlocks             NodeUnlocks `json:"unlocks"`
	DNSBL               int         `json:"dnsbl"`
	IPChanged           bool        `json:"ip_changed"`
	LastSeen            time.Time   `json:"last_seen"`
	LastScan            time.Time   `json:"last_scan"`
	ScanIntervalMinutes int         `json:"scan_interval_minutes"`
	QualityStatus       string      `json:"quality_status"`
	LastScanError       string      `json:"last_scan_error,omitempty"`
	LastTask            *TaskLog    `json:"last_task,omitempty"`
}

type TaskLog struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	Type      string    `json:"type"`       // "scan" | "reinstall"
	Status    string    `json:"status"`     // "pending" | "running" | "completed" | "failed"
	Message   string    `json:"message"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublicSettings struct {
	RegistrationEnabled bool `json:"registration_enabled"`
	Bootstrapped        bool `json:"bootstrapped"`
}

type TrendPoint struct {
	At    time.Time `json:"at"`
	Risk  int       `json:"risk"`
	IPQS  int       `json:"ipqs"`
	Scam  int       `json:"scamalytics"`
	DNSBL int       `json:"dnsbl"`
}

type Alert struct {
	ID           string     `json:"id"`
	NodeID       string     `json:"node_id"`
	NodeName     string     `json:"node_name"`
	Type         string     `json:"type"`
	Severity     string     `json:"severity"`
	Title        string     `json:"title"`
	Detail       string     `json:"detail"`
	CreatedAt    time.Time  `json:"created_at"`
	Acknowledged bool       `json:"acknowledged"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type Collector struct {
	Name            string `json:"name"`
	AdapterVersion  string `json:"adapter_version"`
	UpstreamVersion string `json:"upstream_version,omitempty"`
}

type Network struct {
	Family     int    `json:"family"`
	ReportedIP string `json:"reported_ip"`
}

type Quality struct {
	ASN          int64                      `json:"asn"`
	Organization string                     `json:"organization"`
	CountryCode  string                     `json:"country_code"`
	City         string                     `json:"city,omitempty"`
	Latitude     float64                    `json:"latitude,omitempty"`
	Longitude    float64                    `json:"longitude,omitempty"`
	UsageType    string                     `json:"usage_type"`
	CompanyType  string                     `json:"company_type"`
	Scores       map[string]json.RawMessage `json:"scores"`
	Factors      map[string]any             `json:"factors"`
	Media        map[string]any             `json:"media"`
	Mail         map[string]any             `json:"mail"`
}


type Report struct {
	SchemaVersion string          `json:"schema_version"`
	ReportID      string          `json:"report_id"`
	AgentID       string          `json:"agent_id"`
	NodeID        string          `json:"node_id"`
	CollectedAt   time.Time       `json:"collected_at"`
	Collector     Collector       `json:"collector"`
	Network       Network         `json:"network"`
	Quality       Quality         `json:"quality"`
	Raw           json.RawMessage `json:"raw_upstream,omitempty"`
}

type Heartbeat struct {
	AgentID      string         `json:"agent_id"`
	NodeID       string         `json:"node_id"`
	ObservedAt   time.Time      `json:"observed_at"`
	AgentVersion string         `json:"agent_version"`
	ReportedIP   string         `json:"reported_ip,omitempty"`
	Status       map[string]any `json:"status,omitempty"`
}

type Dashboard struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Stats       map[string]int `json:"stats"`
	Trend       []TrendPoint   `json:"trend"`
	Nodes       []Node         `json:"nodes"`
	Rankings    []NodeRanking  `json:"rankings"`
	Alerts      []Alert        `json:"alerts"`
	Regions     map[string]int `json:"regions"`
	Services    []ServiceStat  `json:"services,omitempty"`
}

type NodeRanking struct {
	Rank     int    `json:"rank"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
	Quality  int    `json:"quality"`
	Unlocks  int    `json:"unlocks"`
	Risk     int    `json:"risk"`
	Status   string `json:"status"`
}

type NodeDetail struct {
	Node
	Series          []TrendPoint      `json:"series"`
	Alerts          []Alert           `json:"alerts"`
	Tasks           []TaskLog         `json:"tasks"`
	Networks        []NetworkSnapshot `json:"networks"`
	LatestQuality   *Quality          `json:"latest_quality,omitempty"`
	LatestCollector *Collector        `json:"latest_collector,omitempty"`
	ReportTime      *time.Time        `json:"report_time,omitempty"`
}


type NetworkSnapshot struct {
	Family       int         `json:"family"`
	MaskedIP     string      `json:"masked_ip"`
	IPAddress    string      `json:"ip_address,omitempty"`
	ASN          int64       `json:"asn"`
	Organization string      `json:"organization"`
	CountryCode  string      `json:"country_code"`
	Risk         int         `json:"risk"`
	Netflix      string      `json:"netflix"`
	ChatGPT      string      `json:"chatgpt"`
	Unlocks      NodeUnlocks `json:"unlocks,omitempty"`
	CollectedAt  time.Time   `json:"collected_at"`
	Quality      Quality     `json:"quality"`
}
