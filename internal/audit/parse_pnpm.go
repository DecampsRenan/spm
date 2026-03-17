package audit

import "encoding/json"

// pnpmAuditOutput matches the pnpm audit --json output format.
type pnpmAuditOutput struct {
	Advisories map[string]pnpmAdvisory `json:"advisories"`
}

type pnpmAdvisory struct {
	ModuleName string `json:"module_name"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Range      string `json:"vulnerable_versions"`
	Patched    string `json:"patched_versions"`
}

func parsePnpm(data []byte) (*AuditResult, error) {
	var out pnpmAuditOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	result := &AuditResult{
		Summary: make(map[Severity]int),
		PM:      "pnpm",
	}

	for _, adv := range out.Advisories {
		sev := Severity(adv.Severity)
		result.Vulnerabilities = append(result.Vulnerabilities, Vulnerability{
			Name:     adv.ModuleName,
			Severity: sev,
			Title:    adv.Title,
			URL:      adv.URL,
			Range:    adv.Range,
			Fixed:    adv.Patched,
		})
		result.Summary[sev]++
	}

	return result, nil
}
