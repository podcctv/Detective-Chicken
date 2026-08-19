package model

import (
	"encoding/json"
	"time"
)

type Node struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	AgentID      string    `json:"agent_id,omitempty"`
	Name         string    `json:"name"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	Family       int       `json:"family"`
	ReportedIP   string    `json:"reported_ip"`
	MaskedIP     string    `json:"masked_ip"`
	ASN          int64     `json:"asn"`
	Organization string    `json:"organization"`
	CountryCode  string    `json:"country_code"`
	Risk         int       `json:"risk"`
	Status       string    `json:"status"`
	Netflix      string    `json:"netflix"`
	ChatGPT      string    `json:"chatgpt"`
	DNSBL        int       `json:"dnsbl"`
	IPChanged    bool      `json:"ip_changed"`
	LastSeen     time.Time `json:"last_seen"`
	LastScan     time.Time `json:"last_scan"`
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
	Alerts      []Alert        `json:"alerts"`
	Regions     map[string]int `json:"regions"`
}
