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
	cmd := exec.CommandContext(ctx, "bash", "-c", `curl -fsSL --proto '=https' --tlsv1.2 "$1" | bash -s -- "$2" -j -p`, "detective-chicken-ipquality", url, flag)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return model.Report{}, fmt.Errorf("ipquality failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return ParseIPQuality(stdout.Bytes(), family)
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
	info := mapAt(upstream, "Info")
	scores := rawMap(mapAt(upstream, "Score"))
	quality := model.Quality{ASN: int64(numberAt(info, "ASN")), Organization: stringAt(info, "Organization", "Org"), CountryCode: stringAt(info, "CountryCode", "Country"), UsageType: stringAt(mapAt(upstream, "Type"), "UsageType", "Usage"), CompanyType: stringAt(mapAt(upstream, "Type"), "CompanyType", "Company"), Scores: scores, Factors: mapAt(upstream, "Factor"), Media: mapAt(upstream, "Media"), Mail: mapAt(upstream, "Mail")}
	reportedIP := stringAt(info, "IP", "ip")
	return model.Report{SchemaVersion: "1.0", ReportID: newID("rpt"), CollectedAt: time.Now().UTC(), Collector: model.Collector{Name: "ipquality", AdapterVersion: Version, UpstreamVersion: stringAt(mapAt(upstream, "Head"), "Version", "version")}, Network: model.Network{Family: family, ReportedIP: reportedIP}, Quality: quality, Raw: append([]byte(nil), raw...)}, nil
}

func extractJSONObject(raw []byte) []byte {
	start := bytes.IndexByte(raw, '{')
	end := bytes.LastIndexByte(raw, '}')
	if start < 0 || end <= start {
		return nil
	}
	return bytes.TrimSpace(raw[start : end+1])
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
func numberAt(m map[string]any, key string) float64 {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			switch n := v.(type) {
			case float64:
				return n
			case string:
				f, _ := strconv.ParseFloat(strings.TrimPrefix(strings.ToUpper(n), "AS"), 64)
				return f
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
