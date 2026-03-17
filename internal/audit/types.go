package audit

import "strings"

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityModerate Severity = "moderate"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

var severityRank = map[Severity]int{
	SeverityInfo:     0,
	SeverityLow:      1,
	SeverityModerate: 2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

// SeverityRank returns the numeric rank of a severity (0=info, 4=critical).
// Unknown severities return -1.
func SeverityRank(s Severity) int {
	if r, ok := severityRank[s]; ok {
		return r
	}
	return -1
}

// ParseSeverity normalizes a string to a Severity, case-insensitive.
// Returns empty string and false if invalid.
func ParseSeverity(s string) (Severity, bool) {
	sev := Severity(strings.ToLower(s))
	if _, ok := severityRank[sev]; ok {
		return sev, true
	}
	return "", false
}

type Vulnerability struct {
	Name     string   `json:"name"`
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Range    string   `json:"range"`
	Fixed    string   `json:"fixedIn"`
}

type AuditResult struct {
	Vulnerabilities []Vulnerability  `json:"vulnerabilities"`
	Summary         map[Severity]int `json:"summary"`
	PM              string           `json:"packageManager"`
}

type Options struct {
	ProdOnly bool
	JSON     bool
	Severity Severity // minimum severity filter
	Notify   bool
	DryRun   bool
}
