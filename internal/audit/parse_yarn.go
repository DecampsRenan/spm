package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
)

// --- Yarn Classic (v1) ---
// Yarn v1 outputs NDJSON lines. We care about lines with type "auditAdvisory".

type yarnClassicLine struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type yarnClassicAdvisory struct {
	Advisory struct {
		ModuleName string `json:"module_name"`
		Severity   string `json:"severity"`
		Title      string `json:"title"`
		URL        string `json:"url"`
		Range      string `json:"vulnerable_versions"`
		Patched    string `json:"patched_versions"`
	} `json:"advisory"`
}

func parseYarnClassic(data []byte) (*AuditResult, error) {
	result := &AuditResult{
		Summary: make(map[Severity]int),
		PM:      "yarn",
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var line yarnClassicLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue // skip non-JSON lines
		}
		if line.Type != "auditAdvisory" {
			continue
		}

		var adv yarnClassicAdvisory
		if err := json.Unmarshal(line.Data, &adv); err != nil {
			continue
		}

		sev := Severity(adv.Advisory.Severity)
		result.Vulnerabilities = append(result.Vulnerabilities, Vulnerability{
			Name:     adv.Advisory.ModuleName,
			Severity: sev,
			Title:    adv.Advisory.Title,
			URL:      adv.Advisory.URL,
			Range:    adv.Advisory.Range,
			Fixed:    adv.Advisory.Patched,
		})
		result.Summary[sev]++
	}

	return result, scanner.Err()
}

// --- Yarn Berry (v2+) ---
// Yarn Berry `yarn npm audit --all --json` outputs a JSON object with advisories.

type yarnBerryOutput struct {
	Advisories map[string]yarnBerryAdvisory `json:"advisories"`
}

type yarnBerryAdvisory struct {
	ModuleName string `json:"module_name"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	Range      string `json:"vulnerable_versions"`
	Patched    string `json:"patched_versions"`
}

func parseYarnBerry(data []byte) (*AuditResult, error) {
	var out yarnBerryOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	result := &AuditResult{
		Summary: make(map[Severity]int),
		PM:      "yarn",
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
