package audit

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	ExitClean = 0
	ExitVulns = 1
	ExitError = 2
)

// Run executes the audit for the given package manager, parses the output,
// filters by severity, and renders the result. Returns an exit code.
func Run(pm string, dir string, opts Options) (int, error) {
	args, err := buildCommand(pm, dir, opts)
	if err != nil {
		return ExitError, err
	}

	if opts.DryRun {
		fmt.Printf("Would run: %s\n", strings.Join(args, " "))
		return ExitClean, nil
	}

	// Execute the audit command, capturing stdout.
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// We ignore the exit code from the PM — npm/yarn return non-zero
	// when vulnerabilities are found, which is expected.
	_ = cmd.Run()

	data := stdout.Bytes()
	if len(data) == 0 {
		return ExitError, fmt.Errorf("%s audit produced no output", pm)
	}

	// Parse based on PM.
	result, err := parse(pm, dir, data)
	if err != nil {
		return ExitError, fmt.Errorf("failed to parse %s audit output: %w", pm, err)
	}

	// Filter by minimum severity.
	if opts.Severity != "" {
		result = filterBySeverity(result, opts.Severity)
	}

	// Render.
	if opts.JSON {
		if err := renderJSON(result, os.Stdout); err != nil {
			return ExitError, err
		}
	} else {
		renderTable(result, os.Stdout)
	}

	if len(result.Vulnerabilities) > 0 {
		return ExitVulns, nil
	}
	return ExitClean, nil
}

func parse(pm string, dir string, data []byte) (*AuditResult, error) {
	switch pm {
	case "npm":
		return parseNPM(data)
	case "yarn":
		version, err := detectYarnVersion(dir)
		if err != nil {
			// Fall back to classic parse if we can't detect version.
			return parseYarnClassic(data)
		}
		if version >= 2 {
			return parseYarnBerry(data)
		}
		return parseYarnClassic(data)
	case "pnpm":
		return parsePnpm(data)
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

func filterBySeverity(result *AuditResult, minSev Severity) *AuditResult {
	minRank := SeverityRank(minSev)
	filtered := &AuditResult{
		Summary: make(map[Severity]int),
		PM:      result.PM,
	}
	for _, v := range result.Vulnerabilities {
		if SeverityRank(v.Severity) >= minRank {
			filtered.Vulnerabilities = append(filtered.Vulnerabilities, v)
			filtered.Summary[v.Severity]++
		}
	}
	return filtered
}
