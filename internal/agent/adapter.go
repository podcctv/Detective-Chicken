package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/podcctv/detective-chicken/internal/agent/prober"
	"github.com/podcctv/detective-chicken/internal/model"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\([a-zA-Z]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)`)

func stripANSI(b []byte) []byte {
	return ansiRegex.ReplaceAll(b, nil)
}

type Adapter struct {
	ScriptURL string
	Timeout   time.Duration
	ProxyURL  string
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
	if err := ValidateScanProxy(a.ProxyURL); err != nil {
		return model.Report{}, err
	}

	type nativePack struct {
		results  map[string]prober.ProbeResult
		identity prober.NetworkIdentity
	}
	nativeChan := make(chan nativePack, 1)
	go func() {
		p := prober.NewProber(family, 6*time.Second).WithProxy(a.ProxyURL)
		ident := p.ProbeNetworkIdentity(ctx)
		res := p.RunAll(ctx)
		nativeChan <- nativePack{results: res, identity: ident}
	}()

	// 2. Run Upstream Script for ASN, Geo, Risk scores, etc.
	flag := "-4"
	if family == 6 {
		flag = "-6"
	}
	cmd := exec.CommandContext(ctx, "bash", "-c", `curl -fsSL --proto '=https' --tlsv1.2 "$1" | bash -s -- "$2" -j -p -f`, "detective-chicken-ipquality", url, flag)
	if a.ProxyURL != "" {
		cmd.Env = append(os.Environ(), "HTTP_PROXY="+a.ProxyURL, "HTTPS_PROXY="+a.ProxyURL, "ALL_PROXY="+a.ProxyURL)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	pack := <-nativeChan
	nativeResults := pack.results
	identity := pack.identity

	report, collectErr := finishCollection(stdout.Bytes(), stderr.String(), err, family)
	if collectErr != nil && len(nativeResults) > 0 {
		// Fallback report if upstream script totally failed
		report = model.Report{
			SchemaVersion: "1.0",
			ReportID:      newID("rpt"),
			CollectedAt:   time.Now().UTC(),
			Collector: model.Collector{
				Name:           "native-prober",
				AdapterVersion: Version,
			},
			Network: model.Network{
				Family:     family,
				ReportedIP: identity.IP,
				IsWARP:     identity.IsWARP,
			},
			Quality: model.Quality{
				ASN:          identity.ASN,
				Organization: identity.Organization,
				CountryCode:  identity.CountryCode,
				City:         identity.City,
				Latitude:     identity.Latitude,
				Longitude:    identity.Longitude,
				UsageType:    normalizeUsageLabel(identity.UsageType),
			},
		}
	} else if collectErr != nil {
		return model.Report{}, collectErr
	}
	if identity.IP != "" && !ipMatchesFamily(identity.IP, family) {
		return model.Report{}, fmt.Errorf("outbound route returned %s while scanning IPv%d", identity.IP, family)
	}
	if report.Network.ReportedIP != "" && identity.IP != "" && !sameIP(report.Network.ReportedIP, identity.IP) {
		return model.Report{}, fmt.Errorf("collector route mismatch: upstream exit %s differs from native probe exit %s", report.Network.ReportedIP, identity.IP)
	}

	// The native identity probe and every unlock probe share one transport. Use
	// that observed address as the canonical exit instead of any ingress/local IP.
	if identity.IP != "" {
		report.Network.ReportedIP = identity.IP
	}
	report.Network.IsWARP = identity.IsWARP
	if identity.ASN > 0 {
		report.Quality.ASN = identity.ASN
	}
	if identity.Organization != "" {
		report.Quality.Organization = identity.Organization
	}
	if identity.CountryCode != "" {
		report.Quality.CountryCode = identity.CountryCode
	}
	if identity.City != "" {
		report.Quality.City = identity.City
	}
	if identity.Latitude != 0 {
		report.Quality.Latitude = identity.Latitude
	}
	if identity.Longitude != 0 {
		report.Quality.Longitude = identity.Longitude
	}
	if identity.UsageType != "" {
		report.Quality.UsageType = normalizeUsageLabel(identity.UsageType)
	}
	if identity.IsWARP {
		if report.Quality.Factors == nil {
			report.Quality.Factors = make(map[string]any)
		}
		report.Quality.Factors["WARP"] = true
	}
	if a.ProxyURL != "" {
		if report.Quality.Factors == nil {
			report.Quality.Factors = make(map[string]any)
		}
		report.Quality.Factors["ScanProxy"] = true
	}

	// Merge Native Prober results into report.Quality.Media
	if report.Quality.Media == nil {
		report.Quality.Media = make(map[string]any)
	}
	for id, res := range nativeResults {
		mergeNativeMedia(report.Quality.Media, id, res)
	}

	return report, nil
}

func ipMatchesFamily(value string, family int) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return false
	}
	return family == 4 && ip.To4() != nil || family == 6 && ip.To4() == nil && ip.To16() != nil
}

func sameIP(left, right string) bool {
	a := net.ParseIP(strings.TrimSpace(left))
	b := net.ParseIP(strings.TrimSpace(right))
	return a != nil && b != nil && a.Equal(b)
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
	raw = stripANSI(raw)
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
	typeInfo := mapAt(upstream, "Type")
	factors := mapAt(upstream, "Factor")
	scores := rawMap(mapAt(upstream, "Score"))

	countryCode := stringAt(info, "CountryCode", "Country")
	if countryCode == "" {
		if reg, ok := info["Region"].(map[string]any); ok {
			countryCode = stringAt(reg, "Code")
		}
	}
	if countryCode == "" {
		if reg, ok := info["RegisteredRegion"].(map[string]any); ok {
			countryCode = stringAt(reg, "Code")
		}
	}
	countryCode = normalizeCountryCode(countryCode)
	if countryCode == "" {
		if ccMap := mapAt(factors, "CountryCode"); len(ccMap) > 0 {
			countryCode = stringAt(ccMap, "IPinfo", "IP2LOCATION", "SCAMALYTICS", "ipapi", "DBIP")
		}
	}

	city := stringAt(info, "City")
	if city == "" {
		if cityMap, ok := info["City"].(map[string]any); ok {
			city = stringAt(cityMap, "Name", "Subdivisions")
		}
	}

	usageType := classifyUsage(typeInfo, factors)
	companyType := normalizeUsageLabel(stringAt(typeInfo, "CompanyType"))
	if companyType == "" {
		companyType = normalizeUsageLabel(representativeLabel(mapAt(typeInfo, "Company")))
	}
	ipType := normalizeIPTypeLabel(stringAt(info, "Type", "IPType"))

	quality := model.Quality{
		ASN:          int64(numberAt(info, "ASN", "AutonomousSystemNumber")),
		Organization: stringAt(info, "Organization", "Org"),
		CountryCode:  countryCode,
		City:         city,
		Latitude:     numberAt(info, "Latitude"),
		Longitude:    numberAt(info, "Longitude"),
		UsageType:    usageType,
		CompanyType:  companyType,
		IPType:       ipType,
		Scores:       scores,
		Factors:      factors,
		Media:        mapAt(upstream, "Media"),
		Mail:         mapAt(upstream, "Mail"),
	}
	reportedIP := stringAt(head, "IP", "ip")
	if reportedIP == "" {
		reportedIP = stringAt(info, "IP", "ip")
	}
	return model.Report{
		SchemaVersion: "1.0",
		ReportID:      newID("rpt"),
		CollectedAt:   time.Now().UTC(),
		Collector: model.Collector{
			Name:            "ipquality",
			AdapterVersion:  Version,
			UpstreamVersion: stringAt(head, "Version", "version"),
		},
		Network: model.Network{
			Family:     family,
			ReportedIP: reportedIP,
		},
		Quality: quality,
		Raw:     append([]byte(nil), raw...),
	}, nil
}

func extractJSONObject(raw []byte) []byte {
	start := bytes.IndexByte(raw, '{')
	end := bytes.LastIndexByte(raw, '}')
	if start < 0 || end <= start {
		return nil
	}
	return bytes.TrimSpace(raw[start : end+1])
}

func classifyUsage(typeInfo, factors map[string]any) string {
	if direct := normalizeUsageLabel(stringAt(typeInfo, "UsageType")); direct != "" {
		return direct
	}
	if usage := representativeUsage(mapAt(typeInfo, "Usage")); usage != "" {
		return usage
	}
	if company := representativeUsage(mapAt(typeInfo, "Company")); company != "" {
		return company
	}
	if anyTrue(mapAt(factors, "Server")) {
		return "机房"
	}
	return ""
}

func representativeUsage(values map[string]any) string {
	counts := map[string]int{}
	order := make([]string, 0, 3)
	for _, raw := range orderedValues(values) {
		label := normalizeUsageLabel(fmt.Sprint(raw))
		if label == "" {
			continue
		}
		if _, exists := counts[label]; !exists {
			order = append(order, label)
		}
		counts[label]++
	}
	best, bestCount := "", 0
	for _, label := range order {
		if counts[label] > bestCount {
			best, bestCount = label, counts[label]
		}
	}
	return best
}

func representativeLabel(values map[string]any) string {
	for _, raw := range orderedValues(values) {
		if label := cleanString(fmt.Sprint(raw)); label != "" {
			return label
		}
	}
	return ""
}

func orderedValues(values map[string]any) []any {
	if len(values) == 0 {
		return nil
	}
	preferred := []string{"IPinfo", "ipregistry", "ipapi", "AbuseIPDB", "IP2LOCATION"}
	result := make([]any, 0, len(values))
	seen := map[string]bool{}
	for _, wanted := range preferred {
		for key, value := range values {
			if strings.EqualFold(key, wanted) {
				result = append(result, value)
				seen[key] = true
				break
			}
		}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, values[key])
	}
	return result
}

func normalizeUsageLabel(label string) string {
	label = cleanString(label)
	if label == "" {
		return ""
	}
	lower := strings.ToLower(label)
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
	}
	return ""
}

func normalizeIPTypeLabel(label string) string {
	lower := strings.ToLower(cleanString(label))
	switch {
	case strings.Contains(lower, "原生"), strings.Contains(lower, "native"), strings.Contains(lower, "geo-consistent"):
		return "原生"
	case strings.Contains(lower, "广播"), strings.Contains(lower, "broadcast"), strings.Contains(lower, "geo-discrepant"):
		return "广播"
	default:
		return ""
	}
}

func normalizeCountryCode(value string) string {
	value = strings.ToUpper(strings.Trim(strings.TrimSpace(cleanString(value)), "[]"))
	if len(value) != 2 {
		return ""
	}
	return value
}

func cleanString(value string) string {
	value = strings.TrimSpace(value)
	switch strings.ToLower(value) {
	case "", "null", "<nil>", "n/a", "unknown":
		return ""
	default:
		return value
	}
}

func anyTrue(values map[string]any) bool {
	for _, value := range values {
		switch typed := value.(type) {
		case bool:
			if typed {
				return true
			}
		case string:
			if strings.EqualFold(strings.TrimSpace(typed), "true") || strings.EqualFold(strings.TrimSpace(typed), "yes") {
				return true
			}
		case map[string]any:
			if anyTrue(typed) {
				return true
			}
		}
	}
	return false
}

func mergeNativeMedia(media map[string]any, id string, result prober.ProbeResult) {
	entry := map[string]any{
		"ID":        result.ID,
		"Name":      result.Name,
		"Category":  result.Category,
		"Status":    result.Status,
		"Region":    result.Region,
		"Quality":   result.Quality,
		"LatencyMs": result.Latency,
		"Detail":    result.Detail,
		"CheckedAt": time.Now().UTC().Format(time.RFC3339),
	}
	var existing map[string]any
	for key, raw := range media {
		if !strings.EqualFold(key, id) {
			continue
		}
		if candidate, ok := raw.(map[string]any); ok && existing == nil {
			existing = candidate
		}
		delete(media, key)
	}
	if existing == nil || !usableMediaEntry(existing) {
		media[id] = entry
		return
	}
	if id == "youtube" && probeIsYouTubeCN(result) {
		media[id] = entry
		return
	}
	for key, value := range entry {
		if _, ok := valueAt(existing, key); !ok || cleanString(fmt.Sprint(valueAtValue(existing, key))) == "" {
			existing[key] = value
		}
	}
	media[id] = existing
}

func usableMediaEntry(entry map[string]any) bool {
	status, ok := valueAt(entry, "Status")
	if !ok {
		return false
	}
	normalized := strings.ToLower(cleanString(fmt.Sprint(status)))
	switch normalized {
	case "", "failed", "failure", "error", "失败", "检测失败", "结果待确认":
		return false
	default:
		return true
	}
}

func probeIsYouTubeCN(result prober.ProbeResult) bool {
	return strings.EqualFold(strings.TrimSpace(result.Region), "CN") || strings.Contains(result.Quality, "送中")
}

func valueAt(values map[string]any, key string) (any, bool) {
	for candidate, value := range values {
		if strings.EqualFold(candidate, key) {
			return value, true
		}
	}
	return nil, false
}

func valueAtValue(values map[string]any, key string) any {
	value, _ := valueAt(values, key)
	return value
}

func mapAt(m map[string]any, key string) map[string]any {
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
	return map[string]any{}
}
func stringAt(m map[string]any, keys ...string) string {
	for _, key := range keys {
		for k, v := range m {
			if strings.EqualFold(k, key) {
				switch value := v.(type) {
				case string:
					if cleaned := cleanString(value); cleaned != "" {
						return cleaned
					}
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
					f, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(strings.ToUpper(n), "AS")), 64)
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
