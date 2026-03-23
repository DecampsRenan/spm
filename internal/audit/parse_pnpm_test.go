package audit

import (
	"os"
	"testing"
)

func TestParsePnpm_Vulns(t *testing.T) {
	data, err := os.ReadFile("testdata/pnpm_vulns.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parsePnpm(data)
	if err != nil {
		t.Fatal(err)
	}

	if result.PM != "pnpm" {
		t.Errorf("PM = %q, want %q", result.PM, "pnpm")
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

	// Verify fixed version is parsed.
	for _, v := range result.Vulnerabilities {
		if v.Name == "qs" {
			if v.Fixed != ">=6.5.3" {
				t.Errorf("fixed = %q, want %q", v.Fixed, ">=6.5.3")
			}
		}
	}
}

func TestParsePnpm_Clean(t *testing.T) {
	data, err := os.ReadFile("testdata/pnpm_clean.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := parsePnpm(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Vulnerabilities) != 0 {
		t.Errorf("got %d vulns, want 0", len(result.Vulnerabilities))
	}
}
