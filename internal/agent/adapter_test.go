package agent

import (
	"errors"
	"testing"
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

func TestParseIPQualityReadsUsageTypeMap(t *testing.T) {
	// The real ipquality JSON nests Usage/Company as provider→label maps.
	raw := []byte(`{
		"Head":{"IP":"203.0.113.8","Version":"v2026-08-09"},
		"Info":{"IP":"203.0.113.8","ASN":"3462","Organization":"Example","CountryCode":"TW"},
		"Type":{"Usage":{"IPinfo":"ISP","ipregistry":"ISP","ipapi":"ISP","AbuseIPDB":"Line ISP"},"Company":{"IPinfo":"ISP","ipregistry":"ISP","ipapi":"ISP"}},
		"Score":{"IPQS":"18"},
		"Factor":{"Proxy":{"IP2LOCATION":false},"VPN":{"IP2LOCATION":false}},
		"Media":{"Netflix":{"Status":"Yes"},"ChatGPT":{"Status":"Yes"},"Youtube":{"Status":"Yes","Region":"CN"}},
		"Mail":{}
	}`)
	report, err := ParseIPQuality(raw, 4)
	if err != nil {
		t.Fatal(err)
	}
	if report.Quality.ASN != 3462 {
		t.Fatalf("ASN not parsed: %#v", report.Quality.ASN)
	}
	if report.Quality.UsageType != "家宽" {
		t.Fatalf("usage type not normalized, got %q", report.Quality.UsageType)
	}
}

func TestCollectorAcceptsValidJSONDespiteUpstreamExitCode(t *testing.T) {
	raw := []byte(`{"Info":{"IP":"203.0.113.8","ASN":64500},"Score":{"IPQS":18},"Factor":{},"Media":{},"Mail":{}}`)
	report, err := finishCollection(raw, "upstream returned status 1", errors.New("exit status 1"), 4)
	if err != nil || report.Network.ReportedIP != "203.0.113.8" {
		t.Fatalf("valid report was discarded: %#v %v", report, err)
	}
}
