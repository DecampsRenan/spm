package ecosystem

import (
	"os"
	"reflect"
	"testing"

	"github.com/decampsrenan/spm/internal/audit"
)

func TestNPMBuildAuditCommand(t *testing.T) {
	eco := &npmEcosystem{}
	tests := []struct {
		name string
		opts audit.Options
		want []string
	}{
		{"default", audit.Options{}, []string{"npm", "audit", "--json"}},
		{"prod-only", audit.Options{ProdOnly: true}, []string{"npm", "audit", "--json", "--omit=dev"}},
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

func TestNPMParseAuditOutput_Vulns(t *testing.T) {
	eco := &npmEcosystem{}
	data, err := os.ReadFile("testdata/npm_vulns.json")
	if err != nil {
		t.Fatal(err)
	}

	result, err := eco.ParseAuditOutput("", data)
	if err != nil {
		t.Fatal(err)
	}

	if result.PM != "npm" {
		t.Errorf("PM = %q, want %q", result.PM, "npm")
	}
	if len(result.Vulnerabilities) != 3 {
		t.Fatalf("got %d vulns, want 3", len(result.Vulnerabilities))
	}
	if result.Summary[audit.SeverityCritical] != 1 {
		t.Errorf("critical = %d, want 1", result.Summary[audit.SeverityCritical])
	}
	if result.Summary[audit.SeverityHigh] != 1 {
		t.Errorf("high = %d, want 1", result.Summary[audit.SeverityHigh])
	}
	if result.Summary[audit.SeverityLow] != 1 {
		t.Errorf("low = %d, want 1", result.Summary[audit.SeverityLow])
	}

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

func TestNPMParseAuditOutput_Clean(t *testing.T) {
	eco := &npmEcosystem{}
	data, err := os.ReadFile("testdata/npm_clean.json")
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
