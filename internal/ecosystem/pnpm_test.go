package ecosystem

import (
	"os"
	"reflect"
	"testing"

	"github.com/decampsrenan/spm/internal/audit"
)

func TestPnpmBuildAuditCommand(t *testing.T) {
	eco := &pnpmEcosystem{}
	tests := []struct {
		name string
		opts audit.Options
		want []string
	}{
		{"default", audit.Options{}, []string{"pnpm", "audit", "--json"}},
		{"prod-only", audit.Options{ProdOnly: true}, []string{"pnpm", "audit", "--json", "--prod"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := eco.BuildAuditCommand("", tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPnpmParseAuditOutput_Vulns(t *testing.T) {
	eco := &pnpmEcosystem{}
	data, err := os.ReadFile("testdata/pnpm_vulns.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := eco.ParseAuditOutput("", data)
	if err != nil {
		t.Fatal(err)
	}

	if result.PM != "pnpm" {
		t.Errorf("PM = %q, want %q", result.PM, "pnpm")
	}
	if len(result.Vulnerabilities) != 2 {
		t.Fatalf("got %d vulns, want 2", len(result.Vulnerabilities))
	}
	if result.Summary[audit.SeverityHigh] != 1 {
		t.Errorf("high = %d, want 1", result.Summary[audit.SeverityHigh])
	}
	if result.Summary[audit.SeverityModerate] != 1 {
		t.Errorf("moderate = %d, want 1", result.Summary[audit.SeverityModerate])
	}

	for _, v := range result.Vulnerabilities {
		if v.Name == "qs" {
			if v.Fixed != ">=6.5.3" {
				t.Errorf("fixed = %q, want %q", v.Fixed, ">=6.5.3")
			}
		}
	}
}

func TestPnpmParseAuditOutput_Clean(t *testing.T) {
	eco := &pnpmEcosystem{}
	data, err := os.ReadFile("testdata/pnpm_clean.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := eco.ParseAuditOutput("", data)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Vulnerabilities) != 0 {
		t.Errorf("got %d vulns, want 0", len(result.Vulnerabilities))
	}
}
