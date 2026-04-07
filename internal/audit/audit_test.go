package audit

import (
	"bytes"
	"os"
	"testing"
)

// stubProvider is a minimal Provider for testing.
type stubProvider struct{}

func (s *stubProvider) BuildAuditCommand(_ string, _ Options) ([]string, error) {
	return []string{"npm", "audit", "--json"}, nil
}

func (s *stubProvider) ParseAuditOutput(_ string, _ []byte) (*AuditResult, error) {
	return &AuditResult{PM: "npm", Summary: make(map[Severity]int)}, nil
}

func TestRunDryRun(t *testing.T) {
	// Capture stdout to verify dry-run message.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode, err := Run(&stubProvider{}, t.TempDir(), Options{DryRun: true})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exitCode != ExitClean {
		t.Errorf("expected exit code %d, got %d", ExitClean, exitCode)
	}
	if out := buf.String(); out == "" {
		t.Error("expected dry-run output, got empty string")
	}
}

func TestFilterBySeverity(t *testing.T) {
	result := &AuditResult{
		PM: "npm",
		Vulnerabilities: []Vulnerability{
			{Name: "a", Severity: SeverityLow},
			{Name: "b", Severity: SeverityModerate},
			{Name: "c", Severity: SeverityHigh},
			{Name: "d", Severity: SeverityCritical},
		},
		Summary: map[Severity]int{
			SeverityLow:      1,
			SeverityModerate: 1,
			SeverityHigh:     1,
			SeverityCritical: 1,
		},
	}

	filtered := filterBySeverity(result, SeverityHigh)
	if len(filtered.Vulnerabilities) != 2 {
		t.Errorf("got %d vulns, want 2", len(filtered.Vulnerabilities))
	}
	if filtered.Summary[SeverityHigh] != 1 || filtered.Summary[SeverityCritical] != 1 {
		t.Errorf("unexpected summary: %v", filtered.Summary)
	}
	if filtered.Summary[SeverityLow] != 0 {
		t.Errorf("low should be filtered out")
	}
}

func TestSeverityRank(t *testing.T) {
	if SeverityRank(SeverityInfo) >= SeverityRank(SeverityLow) {
		t.Error("info should rank below low")
	}
	if SeverityRank(SeverityCritical) <= SeverityRank(SeverityHigh) {
		t.Error("critical should rank above high")
	}
	if SeverityRank("unknown") != -1 {
		t.Error("unknown severity should return -1")
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		want  Severity
	}{
		{"high", true, SeverityHigh},
		{"HIGH", true, SeverityHigh},
		{"Critical", true, SeverityCritical},
		{"invalid", false, ""},
	}
	for _, tt := range tests {
		sev, ok := ParseSeverity(tt.input)
		if ok != tt.valid {
			t.Errorf("ParseSeverity(%q) valid = %v, want %v", tt.input, ok, tt.valid)
		}
		if ok && sev != tt.want {
			t.Errorf("ParseSeverity(%q) = %q, want %q", tt.input, sev, tt.want)
		}
	}
}
