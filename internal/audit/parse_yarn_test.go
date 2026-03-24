package audit

import (
	"os"
	"testing"
)

func TestParseYarnClassic_Vulns(t *testing.T) {
	data, err := os.ReadFile("testdata/yarn_classic_vulns.ndjson")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseYarnClassic(data)
	if err != nil {
		t.Fatal(err)
	}

	if result.PM != "yarn" {
		t.Errorf("PM = %q, want %q", result.PM, "yarn")
	}
	if len(result.Vulnerabilities) != 2 {
		t.Fatalf("got %d vulns, want 2", len(result.Vulnerabilities))
	}
	if result.Summary[SeverityHigh] != 1 {
		t.Errorf("high = %d, want 1", result.Summary[SeverityHigh])
	}
	if result.Summary[SeverityModerate] != 1 {
		t.Errorf("moderate = %d, want 1", result.Summary[SeverityModerate])
	}
}

func TestParseYarnClassic_Clean(t *testing.T) {
	data, err := os.ReadFile("testdata/yarn_classic_clean.ndjson")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseYarnClassic(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Vulnerabilities) != 0 {
		t.Errorf("got %d vulns, want 0", len(result.Vulnerabilities))
	}
}

func TestParseYarnBerry_Vulns(t *testing.T) {
	data, err := os.ReadFile("testdata/yarn_berry_vulns.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseYarnBerry(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Vulnerabilities) != 1 {
		t.Fatalf("got %d vulns, want 1", len(result.Vulnerabilities))
	}

	v := result.Vulnerabilities[0]
	if v.Name != "node-fetch" {
		t.Errorf("name = %q, want %q", v.Name, "node-fetch")
	}
	if v.Severity != SeverityHigh {
		t.Errorf("severity = %q, want %q", v.Severity, SeverityHigh)
	}
}

func TestParseYarnBerry_Clean(t *testing.T) {
	data, err := os.ReadFile("testdata/yarn_berry_clean.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseYarnBerry(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Vulnerabilities) != 0 {
		t.Errorf("got %d vulns, want 0", len(result.Vulnerabilities))
	}
}
