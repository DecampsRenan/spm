package ecosystem

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/decampsrenan/spm/internal/audit"
)

type yarnEcosystem struct{}

func (y *yarnEcosystem) Name() PackageManager     { return Yarn }
func (y *yarnEcosystem) ManifestFile() string     { return "package.json" }
func (y *yarnEcosystem) LockFiles() []string      { return []string{"yarn.lock"} }
func (y *yarnEcosystem) ArtifactDirs() []string   { return []string{"node_modules"} }
func (y *yarnEcosystem) HasCommand(_ string) bool { return true }

func (y *yarnEcosystem) Resolve(cmd string, args []string) []string {
	switch cmd {
	case "init":
		// yarn classic needs -y; yarn Berry ignores it harmlessly
		return append([]string{"yarn", "init", "-y"}, args...)
	case "install", "i":
		return append([]string{"yarn", "install"}, args...)
	case "add":
		return append([]string{"yarn", "add"}, args...)
	case "remove":
		return append([]string{"yarn", "remove"}, args...)
	default:
		// yarn doesn't need explicit "run"
		return append([]string{"yarn", cmd}, args...)
	}
}

func (y *yarnEcosystem) BuildAuditCommand(dir string, opts audit.Options) ([]string, error) {
	version, err := detectYarnVersion(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot detect yarn version: %w", err)
	}
	if version >= 2 {
		return buildYarnBerry(opts), nil
	}
	return buildYarnClassic(opts), nil
}

func (y *yarnEcosystem) ParseAuditOutput(dir string, data []byte) (*audit.AuditResult, error) {
	version, err := detectYarnVersion(dir)
	if err != nil {
		// Fall back to classic parse if we can't detect version.
		return parseYarnClassic(data)
	}
	if version >= 2 {
		return parseYarnBerry(data)
	}
	return parseYarnClassic(data)
}

// detectYarnVersion runs `yarn --version` in the given directory and returns
// the major version number.
func detectYarnVersion(dir string) (int, error) {
	cmd := exec.Command("yarn", "--version")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	version := strings.TrimSpace(string(out))
	if len(version) == 0 {
		return 0, fmt.Errorf("empty yarn version output")
	}
	major := version[0]
	if major < '0' || major > '9' {
		return 0, fmt.Errorf("unexpected yarn version format: %s", version)
	}
	return int(major - '0'), nil
}

func buildYarnClassic(opts audit.Options) []string {
	args := []string{"yarn", "audit", "--json"}
	if opts.ProdOnly {
		args = append(args, "--groups", "dependencies")
	}
	return args
}

func buildYarnBerry(opts audit.Options) []string {
	args := []string{"yarn", "npm", "audit", "--all", "--json"}
	if opts.ProdOnly {
		// Yarn Berry doesn't have a direct --prod flag for audit;
		// we filter post-parse.
	}
	return args
}

// --- Yarn Classic (v1) ---

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

func parseYarnClassic(data []byte) (*audit.AuditResult, error) {
	result := &audit.AuditResult{
		Summary: make(map[audit.Severity]int),
		PM:      "yarn",
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var line yarnClassicLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}
		if line.Type != "auditAdvisory" {
			continue
		}

		var adv yarnClassicAdvisory
		if err := json.Unmarshal(line.Data, &adv); err != nil {
			continue
		}

		sev := audit.Severity(adv.Advisory.Severity)
		result.Vulnerabilities = append(result.Vulnerabilities, audit.Vulnerability{
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

func parseYarnBerry(data []byte) (*audit.AuditResult, error) {
	var out yarnBerryOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	result := &audit.AuditResult{
		Summary: make(map[audit.Severity]int),
		PM:      "yarn",
	}

	for _, adv := range out.Advisories {
		sev := audit.Severity(adv.Severity)
		result.Vulnerabilities = append(result.Vulnerabilities, audit.Vulnerability{
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
