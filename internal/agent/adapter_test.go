package agent

import (
	"errors"
	"testing"

	"github.com/podcctv/detective-chicken/internal/agent/prober"
)

func TestParseIPQualityWithNoiseAndMissingFields(t *testing.T) {
	raw := []byte("checking...\n{\"Head\":{\"Version\":\"v2026-08-09\"},\"Info\":{\"IP\":\"203.0.113.8\",\"ASN\":64500,\"Organization\":\"Example\",\"CountryCode\":\"US\"},\"Type\":{\"UsageType\":\"hosting\"},\"Score\":{\"IPQS\":18},\"Factor\":{\"Proxy\":false},\"Media\":{},\"Mail\":{}}\ndone")
	report, err := ParseIPQuality(raw, 4)
	if err != nil {
		t.Fatal(err)
	}
	if report.Network.ReportedIP != "203.0.113.8" || report.Quality.ASN != 64500 {
		t.Fatalf("unexpected report: %#v", report)
	}
	if string(report.Quality.Scores["ipqs"]) != "18" {
		t.Fatalf("score was not normalized: %s", report.Quality.Scores["ipqs"])
	}
}

func TestParseIPQualityReadsCurrentTypeConsensusAndIPType(t *testing.T) {
	raw := []byte(`{
		"Head":{"IP":"154.94.1.8","Version":"v2026-08-09"},
		"Info":{"ASN":"AS139646","Organization":"JINX CO., LIMITED","Region":{"Code":"HK"},"Type":"广播IP"},
		"Type":{"Usage":{"IPinfo":"机房","ipregistry":"机房","ipapi":"商业","AbuseIPDB":"数据中心","IP2LOCATION":"DCH"}},
		"Score":{},"Factor":{},
		"Media":{"Youtube":{"Status":"中国","Region":"CN","Type":""}},"Mail":{}
	}`)
	report, err := ParseIPQuality(raw, 4)
	if err != nil {
		t.Fatal(err)
	}
	if report.Quality.ASN != 139646 || report.Quality.Organization != "JINX CO., LIMITED" {
		t.Fatalf("ASN identity not parsed: %#v", report.Quality)
	}
	if report.Quality.UsageType != "机房" || report.Quality.IPType != "广播" || report.Quality.CountryCode != "HK" {
		t.Fatalf("type classification not parsed: %#v", report.Quality)
	}
}

func TestMergeNativeMediaPreservesYouTubeCNAndRemovesCaseDuplicate(t *testing.T) {
	media := map[string]any{
		"Youtube": map[string]any{"Status": "中国", "Region": "CN"},
	}
	mergeNativeMedia(media, "youtube", prober.ProbeResult{ID: "youtube", Name: "YouTube Prem", Status: "available", Region: "US", Quality: "Premium 原生解锁"})
	if len(media) != 1 {
		t.Fatalf("case-duplicate media entries remain: %#v", media)
	}
	entry, ok := media["youtube"].(map[string]any)
	if !ok || entry["Status"] != "中国" || entry["Region"] != "CN" {
		t.Fatalf("YouTube CN result was overwritten: %#v", media)
	}
}

func TestParseIPQualityReadsCurrentHeadIP(t *testing.T) {
	raw := []byte(`{"Head":{"IP":"2001:db8::8","Version":"v2026-08-09"},"Info":{"ASN":64500},"Score":{},"Factor":{},"Media":{},"Mail":{}}`)
	report, err := ParseIPQuality(raw, 6)
	if err != nil {
		t.Fatal(err)
	}
	if report.Network.ReportedIP != "2001:db8::8" {
		t.Fatalf("reported IP = %q", report.Network.ReportedIP)
	}
}

func TestParseIPQualityRejectsNonJSON(t *testing.T) {
	if _, err := ParseIPQuality([]byte("network failed"), 4); err == nil {
		t.Fatal("expected error")
	}
}

func TestCollectorAcceptsValidJSONDespiteUpstreamExitCode(t *testing.T) {
	raw := []byte(`{"Info":{"IP":"203.0.113.8","ASN":64500},"Score":{"IPQS":18},"Factor":{},"Media":{},"Mail":{}}`)
	report, err := finishCollection(raw, "upstream returned status 1", errors.New("exit status 1"), 4)
	if err != nil || report.Network.ReportedIP != "203.0.113.8" {
		t.Fatalf("valid report was discarded: %#v %v", report, err)
	}
}
