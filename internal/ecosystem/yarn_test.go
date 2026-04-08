package ecosystem

import (
	"os"
	"reflect"
	"testing"

	"github.com/decampsrenan/spm/internal/audit"
)

func TestYarnBuildAuditCommandClassic(t *testing.T) {
	tests := []struct {
		name string
		opts audit.Options
		want []string
	}{
		{"default", audit.Options{}, []string{"yarn", "audit", "--json"}},
		{"prod-only", audit.Options{ProdOnly: true}, []string{"yarn", "audit", "--json", "--groups", "dependencies"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildYarnClassic(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYarnBuildAuditCommandBerry(t *testing.T) {
	got := buildYarnBerry(audit.Options{})
	want := []string{"yarn", "npm", "audit", "--all", "--json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestYarnClassicParseAuditOutput_Vulns(t *testing.T) {
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
	if result.Summary[audit.SeverityHigh] != 1 {
		t.Errorf("high = %d, want 1", result.Summary[audit.SeverityHigh])
	}
	if result.Summary[audit.SeverityModerate] != 1 {
		t.Errorf("moderate = %d, want 1", result.Summary[audit.SeverityModerate])
	}
}

func TestYarnClassicParseAuditOutput_Clean(t *testing.T) {
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

func TestYarnBerryParseAuditOutput_Vulns(t *testing.T) {
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
	if v.Severity != audit.SeverityHigh {
		t.Errorf("severity = %q, want %q", v.Severity, audit.SeverityHigh)
	}
}

func TestYarnBerryParseAuditOutput_Clean(t *testing.T) {
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
