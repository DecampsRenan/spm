package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderTable_WithVulns(t *testing.T) {
	result := &AuditResult{
		PM: "npm",
		Vulnerabilities: []Vulnerability{
			{Name: "lodash", Severity: SeverityHigh, Title: "Prototype Pollution", URL: "https://example.com/1", Range: "<4.17.21"},
			{Name: "minimist", Severity: SeverityCritical, Title: "Prototype Pollution", URL: "https://example.com/2", Range: "<1.2.6", Fixed: ">=1.2.6"},
		},
		Summary: map[Severity]int{
			SeverityHigh:     1,
			SeverityCritical: 1,
		},
	}

	var buf bytes.Buffer
	renderTable(result, &buf)
	out := buf.String()

	// Severity group headers present.
	if !strings.Contains(out, "CRITICAL") {
		t.Error("missing CRITICAL group header")
	}
	if !strings.Contains(out, "HIGH") {
		t.Error("missing HIGH group header")
	}
	// Critical should appear before high (sorted by severity desc).
	critIdx := strings.Index(out, "CRITICAL")
	highIdx := strings.Index(out, "HIGH")
	if critIdx > highIdx {
		t.Error("critical should appear before high")
	}
	// Package names and titles.
	if !strings.Contains(out, "minimist: Prototype Pollution") {
		t.Error("missing minimist entry")
	}
	if !strings.Contains(out, "lodash: Prototype Pollution") {
		t.Error("missing lodash entry")
	}
	// Version range and fix info.
	if !strings.Contains(out, "versions: <1.2.6") {
		t.Error("missing version range")
	}
	if !strings.Contains(out, "fixed: >=1.2.6") {
		t.Error("missing fixed version")
	}
	// URLs on their own line.
	if !strings.Contains(out, "https://example.com/2") {
		t.Error("missing URL")
	}
	// Summary line.
	if !strings.Contains(out, "2 vulnerabilities found") {
		t.Error("missing summary line")
	}
}

func TestRenderTable_Clean(t *testing.T) {
	result := &AuditResult{
		PM:      "npm",
		Summary: make(map[Severity]int),
	}

	var buf bytes.Buffer
	renderTable(result, &buf)

	if !strings.Contains(buf.String(), "No vulnerabilities found") {
		t.Error("expected clean message")
	}
}

func TestRenderJSON(t *testing.T) {
	result := &AuditResult{
		PM: "npm",
		Vulnerabilities: []Vulnerability{
			{Name: "lodash", Severity: SeverityHigh, Title: "Prototype Pollution"},
		},
		Summary: map[Severity]int{SeverityHigh: 1},
	}

	var buf bytes.Buffer
	if err := renderJSON(result, &buf); err != nil {
		t.Fatal(err)
	}

	var decoded AuditResult
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(decoded.Vulnerabilities) != 1 {
		t.Errorf("got %d vulns, want 1", len(decoded.Vulnerabilities))
	}
}
