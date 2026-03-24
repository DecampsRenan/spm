package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

var severityOrder = []Severity{
	SeverityCritical,
	SeverityHigh,
	SeverityModerate,
	SeverityLow,
	SeverityInfo,
}

var severityLabel = map[Severity]string{
	SeverityCritical: "CRITICAL",
	SeverityHigh:     "HIGH",
	SeverityModerate: "MODERATE",
	SeverityLow:      "LOW",
	SeverityInfo:     "INFO",
}

// renderTable writes a human-readable, grouped list of vulnerabilities.
func renderTable(result *AuditResult, w io.Writer) {
	if len(result.Vulnerabilities) == 0 {
		fmt.Fprintln(w, "No vulnerabilities found.")
		return
	}

	// Sort by severity (critical first), then by name.
	sorted := make([]Vulnerability, len(result.Vulnerabilities))
	copy(sorted, result.Vulnerabilities)
	sort.Slice(sorted, func(i, j int) bool {
		ri, rj := SeverityRank(sorted[i].Severity), SeverityRank(sorted[j].Severity)
		if ri != rj {
			return ri > rj
		}
		return sorted[i].Name < sorted[j].Name
	})

	// Group by severity.
	groups := make(map[Severity][]Vulnerability)
	for _, v := range sorted {
		groups[v.Severity] = append(groups[v.Severity], v)
	}

	first := true
	for _, sev := range severityOrder {
		vulns := groups[sev]
		if len(vulns) == 0 {
			continue
		}

		if !first {
			fmt.Fprintln(w)
		}
		first = false

		label := severityLabel[sev]
		fmt.Fprintf(w, "── %s (%d) ──\n", label, len(vulns))

		for _, v := range vulns {
			title := v.Title
			if title == "" {
				title = "(no title)"
			}
			fmt.Fprintf(w, "  %s: %s\n", v.Name, truncate(title, 70))
			if v.Range != "" {
				fmt.Fprintf(w, "    versions: %s", v.Range)
				if v.Fixed != "" {
					fmt.Fprintf(w, "  fixed: %s", v.Fixed)
				}
				fmt.Fprintln(w)
			}
			if v.URL != "" {
				fmt.Fprintf(w, "    %s\n", v.URL)
			}
		}
	}

	// Summary line.
	fmt.Fprintln(w)
	var parts []string
	for _, sev := range severityOrder {
		if count, ok := result.Summary[sev]; ok && count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, sev))
		}
	}
	total := len(result.Vulnerabilities)
	fmt.Fprintf(w, "%d vulnerabilities found (%s)\n", total, strings.Join(parts, ", "))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// renderJSON writes the audit result as JSON.
func renderJSON(result *AuditResult, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
