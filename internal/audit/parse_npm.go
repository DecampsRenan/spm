package audit

import "encoding/json"

// npmAuditOutput matches the npm audit --json output format (npm v7+).
type npmAuditOutput struct {
	Vulnerabilities map[string]npmVuln `json:"vulnerabilities"`
}

type npmVuln struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Via      []any  `json:"via"`
	Range    string `json:"range"`
	FixAvail any    `json:"fixAvailable"`
}

// parseNPM parses the JSON output of `npm audit --json`.
func parseNPM(data []byte) (*AuditResult, error) {
	var out npmAuditOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	result := &AuditResult{
		Summary: make(map[Severity]int),
		PM:      "npm",
	}

	for name, v := range out.Vulnerabilities {
		sev := Severity(v.Severity)
		title := extractNPMTitle(v.Via)
		url := extractNPMURL(v.Via)
		result.Vulnerabilities = append(result.Vulnerabilities, Vulnerability{
			Name:     name,
			Severity: sev,
			Title:    title,
			URL:      url,
			Range:    v.Range,
		})
		result.Summary[sev]++
	}

	return result, nil
}

// extractNPMTitle gets the title from the first advisory object in the via array.
func extractNPMTitle(via []any) string {
	for _, v := range via {
		if m, ok := v.(map[string]any); ok {
			if t, ok := m["title"].(string); ok {
				return t
			}
		}
	}
	return ""
}

// extractNPMURL gets the URL from the first advisory object in the via array.
func extractNPMURL(via []any) string {
	for _, v := range via {
		if m, ok := v.(map[string]any); ok {
			if u, ok := m["url"].(string); ok {
				return u
			}
		}
	}
	return ""
}
