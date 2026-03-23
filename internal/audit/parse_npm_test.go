package audit

import (
	"os"
	"testing"
)

func TestParseNPM_Vulns(t *testing.T) {
	data, err := os.ReadFile("testdata/npm_vulns.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseNPM(data)
	if err != nil {
		t.Fatal(err)
	}

	if result.PM != "npm" {
		t.Errorf("PM = %q, want %q", result.PM, "npm")
	}
	if len(result.Vulnerabilities) != 3 {
		t.Fatalf("got %d vulns, want 3", len(result.Vulnerabilities))
	}

	// Check summary counts.
	if result.Summary[SeverityCritical] != 1 {
		t.Errorf("critical = %d, want 1", result.Summary[SeverityCritical])
	}
	if result.Summary[SeverityHigh] != 1 {
		t.Errorf("high = %d, want 1", result.Summary[SeverityHigh])
	}
	if result.Summary[SeverityLow] != 1 {
		t.Errorf("low = %d, want 1", result.Summary[SeverityLow])
	}

	// Check a specific vuln has title and URL from the via array.
	found := false
	for _, v := range result.Vulnerabilities {
		if v.Name == "minimist" {
			found = true
			if v.Title != "Prototype Pollution" {
				t.Errorf("title = %q, want %q", v.Title, "Prototype Pollution")
			}
			if v.URL == "" {
				t.Error("expected URL to be set")
			}
		}
	}
	if !found {
		t.Error("minimist vulnerability not found")
	}
}

func TestParseNPM_Clean(t *testing.T) {
	data, err := os.ReadFile("testdata/npm_clean.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parseNPM(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Vulnerabilities) != 0 {
		t.Errorf("got %d vulns, want 0", len(result.Vulnerabilities))
	}
}
