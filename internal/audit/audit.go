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

// Provider abstracts the audit behavior for a given ecosystem.
// Each ecosystem implements this to build and parse its audit commands.
type Provider interface {
	BuildAuditCommand(dir string, opts Options) ([]string, error)
	ParseAuditOutput(dir string, data []byte) (*AuditResult, error)
}

// Run executes the audit for the given provider, parses the output,
// filters by severity, and renders the result. Returns an exit code.
func Run(provider Provider, dir string, opts Options) (int, error) {
	args, err := provider.BuildAuditCommand(dir, opts)
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
		return ExitError, fmt.Errorf("audit produced no output")
	}

	// Parse based on provider.
	result, err := provider.ParseAuditOutput(dir, data)
	if err != nil {
		return ExitError, fmt.Errorf("failed to parse audit output: %w", err)
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
