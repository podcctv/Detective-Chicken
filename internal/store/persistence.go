package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/podcctv/detective-chicken/internal/model"
)

type diskState struct {
	Version             int                           `json:"version"`
	Nodes               map[string]model.Node         `json:"nodes"`
	NodeOwners          map[string]string             `json:"node_owners,omitempty"`
	Series              map[string][]model.TrendPoint `json:"series"`
	Alerts              []model.Alert                 `json:"alerts"`
	Agents              map[string]AgentKey           `json:"agents"`
	Enrollments         map[string]Enrollment         `json:"enrollments"`
	Reports             map[string]model.Report       `json:"reports"`
	Commands            map[string][]Command          `json:"commands"`
	Users               map[string]UserAccount        `json:"users"`
	Sessions            map[string]Session            `json:"sessions"`
	PasswordResets      map[string]PasswordReset      `json:"password_resets"`
	RegistrationEnabled bool                          `json:"registration_enabled"`
}

func NewPersistent(path string, seed bool) (*Memory, error) {
	if path == "" {
		return NewMemory(seed), nil
	}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		m := NewMemory(seed)
		m.dataPath = path
		if err := m.persist(); err != nil {
			return nil, err
		}
		return m, nil
	}
	if err != nil {
		return nil, err
	}
	var state diskState
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("decode data file: %w", err)
	}
	m := NewMemory(false)
	m.dataPath = path
	m.nodes = nonNil(state.Nodes)
	m.series = nonNil(state.Series)
	m.alerts = state.Alerts
	m.agents = nonNil(state.Agents)
	m.enrollments = nonNil(state.Enrollments)
	m.reports = nonNil(state.Reports)
	m.restoreReportedIPs()
	m.commands = nonNil(state.Commands)
	m.users = nonNil(state.Users)
	m.restoreNodeOwners(state.NodeOwners)
	m.sessions = nonNil(state.Sessions)
	m.passwordResets = nonNil(state.PasswordResets)
	m.registrationEnabled = state.RegistrationEnabled
	m.usernames = make(map[string]string, len(m.users))
	for id, account := range m.users {
		m.usernames[normalizeUsername(account.User.Username)] = id
	}
	return m, nil
}

func (m *Memory) restoreNodeOwners(owners map[string]string) {
	for nodeID, ownerID := range owners {
		node, ok := m.nodes[nodeID]
		if ok && ownerID != "" {
			node.OwnerUserID = ownerID
			m.nodes[nodeID] = node
		}
	}
	if len(owners) != 0 || len(m.users) != 1 {
		return
	}
	var onlyUserID string
	for userID := range m.users {
		onlyUserID = userID
	}
	for nodeID, node := range m.nodes {
		if node.OwnerUserID == "" {
			node.OwnerUserID = onlyUserID
			m.nodes[nodeID] = node
		}
	}
}

func (m *Memory) restoreReportedIPs() {
	latest := make(map[string]map[int]model.Report)
	for _, report := range m.reports {
		if _, ok := m.nodes[report.NodeID]; !ok {
			continue
		}
		if latest[report.NodeID] == nil {
			latest[report.NodeID] = make(map[int]model.Report)
		}
		current, ok := latest[report.NodeID][report.Network.Family]
		if !ok || report.CollectedAt.After(current.CollectedAt) {
			latest[report.NodeID][report.Network.Family] = report
		}
	}
	for nodeID, reports := range latest {
		node := m.nodes[nodeID]
		node.Families = nil
		node.MaskedIPv4 = ""
		node.MaskedIPv6 = ""
		node.Warp4 = false
		node.Warp6 = false
		for _, family := range []int{4, 6} {
			report, ok := reports[family]
			if !ok || !validReportedIP(report.Network.ReportedIP, family) {
				continue
			}
			node.Families = append(node.Families, family)
			if family == 4 {
				node.MaskedIPv4 = MaskIP(report.Network.ReportedIP)
				node.Warp4 = reportIsWARP(report)
			} else {
				node.MaskedIPv6 = MaskIP(report.Network.ReportedIP)
				node.Warp6 = reportIsWARP(report)
			}
		}
		node.IsWarp = node.Warp4 || node.Warp6

		primary, ok := reports[4]
		if !ok || !validReportedIP(primary.Network.ReportedIP, 4) {
			primary, ok = reports[6]
		}
		if ok && validReportedIP(primary.Network.ReportedIP, primary.Network.Family) {
			node.Family = primary.Network.Family
			node.ReportedIP = primary.Network.ReportedIP
			node.MaskedIP = MaskIP(primary.Network.ReportedIP)
			applyReportIdentity(&node, primary)
		}
		m.nodes[nodeID] = node
	}
}

func nonNil[K comparable, V any](value map[K]V) map[K]V {
	if value == nil {
		return make(map[K]V)
	}
	return value
}

func (m *Memory) persist() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.writeStateLocked()
}

func (m *Memory) persistLocked() {
	if err := m.writeStateLocked(); err != nil {
		fmt.Fprintf(os.Stderr, "detective-chicken: persist state: %v\n", err)
	}
}

func (m *Memory) writeStateLocked() error {
	if m.dataPath == "" {
		return nil
	}
	nodeOwners := make(map[string]string, len(m.nodes))
	for nodeID, node := range m.nodes {
		if node.OwnerUserID != "" {
			nodeOwners[nodeID] = node.OwnerUserID
		}
	}
	state := diskState{Version: 2, Nodes: m.nodes, NodeOwners: nodeOwners, Series: m.series, Alerts: m.alerts, Agents: m.agents, Enrollments: m.enrollments, Reports: m.reports, Commands: m.commands, Users: m.users, Sessions: m.sessions, PasswordResets: m.passwordResets, RegistrationEnabled: m.registrationEnabled}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(m.dataPath), 0700); err != nil {
		return err
	}
	tmp := m.dataPath + ".tmp"
	if err := os.WriteFile(tmp, raw, 0600); err != nil {
		return err
	}
	return replaceStateFile(tmp, m.dataPath)
}

func replaceStateFile(tmp, target string) error {
	if runtime.GOOS != "windows" {
		return os.Rename(tmp, target)
	}
	backup := target + ".bak"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.Rename(tmp, target)
		}
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Rename(backup, target)
		return err
	}
	_ = os.Remove(backup)
	return nil
}
