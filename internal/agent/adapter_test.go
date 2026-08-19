package agent

import "testing"

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

func TestParseIPQualityRejectsNonJSON(t *testing.T) {
	if _, err := ParseIPQuality([]byte("network failed"), 4); err == nil {
		t.Fatal("expected error")
	}
}
