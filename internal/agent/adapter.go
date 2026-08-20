package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/podcctv/detective-chicken/internal/model"
)

type Adapter struct {
	ScriptURL string
	Timeout   time.Duration
}

func (a Adapter) Collect(ctx context.Context, family int) (model.Report, error) {
	if family != 4 && family != 6 {
		return model.Report{}, errors.New("family must be 4 or 6")
	}
	url := a.ScriptURL
	if url == "" {
		url = "https://IP.Check.Place"
	}
	timeout := a.Timeout
	if timeout == 0 {
		timeout = 8 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	flag := "-4"
	if family == 6 {
		flag = "-6"
	}
	cmd := exec.CommandContext(ctx, "bash", "-c", `curl -fsSL --proto '=https' --tlsv1.2 "$1" | bash -s -- "$2" -j -p -f`, "detective-chicken-ipquality", url, flag)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return finishCollection(stdout.Bytes(), stderr.String(), err, family)
}

func finishCollection(stdout []byte, stderr string, runErr error, family int) (model.Report, error) {
	report, parseErr := ParseIPQuality(stdout, family)
	if parseErr == nil {
		return report, nil
	}
	if runErr != nil {
		return model.Report{}, fmt.Errorf("ipquality failed: %w: %s", runErr, strings.TrimSpace(stderr))
	}
	return model.Report{}, parseErr
}

func ParseIPQuality(raw []byte, family int) (model.Report, error) {
	raw = extractJSONObject(raw)
	if len(raw) == 0 {
		return model.Report{}, errors.New("no JSON object found in collector output")
	}
	var upstream map[string]any
	if err := json.Unmarshal(raw, &upstream); err != nil {
		return model.Report{}, fmt.Errorf("decode upstream JSON: %w", err)
	}
	head := mapAt(upstream, "Head")
	info := mapAt(upstream, "Info")
	typeMap := mapAt(upstream, "Type")
	usageMap := mapAt(typeMap, "Usage", "UsageType")
	companyMap := mapAt(typeMap, "Company", "CompanyType")
	usageLabel := representativeLabel(usageMap)
	if usageLabel == "" {
		usageLabel = representativeLabel(companyMap)
	}
	scores := rawMap(mapAt(upstream, "Score"))
	quality := model.Quality{
		ASN:          int64(numberAt(info, "ASN", "Asn")),
		Organization: stringAt(info, "Organization", "Org"),
		CountryCode:  stringAt(info, "CountryCode", "Country"),
		UsageType:    normalizeIPType(usageLabel),
		CompanyType:  normalizeIPType(representativeLabel(companyMap)),
		Scores:       scores,
		Factors:      mapAt(upstream, "Factor"),
		Media:        mapAt(upstream, "Media"),
		Mail:         mapAt(upstream, "Mail"),
	}
	reportedIP := stringAt(head, "IP", "ip")
	if reportedIP == "" {
		reportedIP = stringAt(info, "IP", "ip")
	}
	return model.Report{SchemaVersion: "1.0", ReportID: newID("rpt"), CollectedAt: time.Now().UTC(), Collector: model.Collector{Name: "ipquality", AdapterVersion: Version, UpstreamVersion: stringAt(head, "Version", "version")}, Network: model.Network{Family: family, ReportedIP: reportedIP}, Quality: quality, Raw: append([]byte(nil), raw...)}, nil
}

func extractJSONObject(raw []byte) []byte {
	start := bytes.IndexByte(raw, '{')
	end := bytes.LastIndexByte(raw, '}')
	if start < 0 || end <= start {
		return nil
	}
	return bytes.TrimSpace(raw[start : end+1])
}
// representativeLabel returns the most common provider→value label from a map
// such as Type.Usage = {"IPinfo":"ISP","ipapi":"ISP","AbuseIPDB":"Line ISP"}.
func representativeLabel(m map[string]any) string {
	if len(m) == 0 {
		return ""
	}
	counts := map[string]int{}
	var order []string
	for _, v := range m {
		var label string
		switch val := v.(type) {
		case string:
			label = strings.TrimSpace(val)
		case bool:
			label = strconv.FormatBool(val)
		default:
			continue
		}
		if label == "" || strings.EqualFold(label, "null") {
			continue
		}
		if _, ok := counts[label]; !ok {
			order = append(order, label)
		}
		counts[label]++
	}
	if len(counts) == 0 {
		return ""
	}
	best := order[0]
	bestCount := 0
	for _, k := range order {
		if counts[k] > bestCount {
			best = k
			bestCount = counts[k]
		}
	}
	return best
}

// normalizeIPType maps a raw usage/company label (e.g. "Data Center",
// "ISP", "Line ISP", "Business") to a coarse Chinese category used by the UI:
// 机房 (datacenter) / 家宽 (residential) / 商宽 (business) / 移动 / 教育 / 政府.
func normalizeIPType(label string) string {
	if label == "" {
		return ""
	}
	l := strings.ToLower(label)
	switch {
	case strings.Contains(l, "data center"), strings.Contains(l, "datacenter"),
		strings.Contains(l, "dch"), strings.Contains(l, "hosting"),
		strings.Contains(l, "cloud"), strings.Contains(l, "server"),
		strings.Contains(l, "idc"), strings.Contains(l, "机房"):
		return "机房"
	case strings.Contains(l, "residential"), strings.Contains(l, "isp"),
		strings.Contains(l, "line isp"), strings.Contains(l, "fixed"),
		strings.Contains(l, "home"), strings.Contains(l, "broadband"),
		strings.Contains(l, "家宽"), strings.Contains(l, "住宅"):
		return "家宽"
	case strings.Contains(l, "business"), strings.Contains(l, "commercial"),
		strings.Contains(l, "company"), strings.Contains(l, "商宽"),
		strings.Contains(l, "商业"):
		return "商宽"
	case strings.Contains(l, "mobile"), strings.Contains(l, "cellular"),
		strings.Contains(l, "手机"), strings.Contains(l, "移动"):
		return "移动"
	case strings.Contains(l, "education"), strings.Contains(l, "edu"),
		strings.Contains(l, "学校"), strings.Contains(l, "教育"):
		return "教育"
	case strings.Contains(l, "government"), strings.Contains(l, "gov"),
		strings.Contains(l, "政府"):
		return "政府"
	}
	return label
}

func mapAt(m map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if v, ok := m[key].(map[string]any); ok {
			return v
		}
		for k, v := range m {
			if strings.EqualFold(k, key) {
				if mm, ok := v.(map[string]any); ok {
					return mm
				}
			}
		}
	}
	return map[string]any{}
}
func stringAt(m map[string]any, keys ...string) string {
	for _, key := range keys {
		for k, v := range m {
			if strings.EqualFold(k, key) {
				switch value := v.(type) {
				case string:
					return value
				case json.Number:
					return value.String()
				case float64:
					return strconv.FormatFloat(value, 'f', -1, 64)
				}
			}
		}
	}
	return ""
}
func numberAt(m map[string]any, keys ...string) float64 {
	for _, key := range keys {
		for k, v := range m {
			if strings.EqualFold(k, key) {
				switch n := v.(type) {
				case float64:
					return n
				case json.Number:
					f, _ := n.Float64()
					return f
				case string:
					f, _ := strconv.ParseFloat(strings.TrimPrefix(strings.ToUpper(n), "AS"), 64)
					return f
				}
			}
		}
	}
	return 0
}
func rawMap(m map[string]any) map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	for k, v := range m {
		raw, _ := json.Marshal(v)
		out[strings.ToLower(k)] = raw
	}
	return out
}
func newID(prefix string) string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
