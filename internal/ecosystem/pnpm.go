package ecosystem

import (
	"encoding/json"

	"github.com/decampsrenan/spm/internal/audit"
)

type pnpmEcosystem struct{}

func (p *pnpmEcosystem) Name() PackageManager     { return Pnpm }
func (p *pnpmEcosystem) ManifestFile() string     { return "package.json" }
func (p *pnpmEcosystem) LockFiles() []string      { return []string{"pnpm-lock.yaml"} }
func (p *pnpmEcosystem) ArtifactDirs() []string   { return []string{"node_modules"} }
func (p *pnpmEcosystem) HasCommand(_ string) bool { return true }

func (p *pnpmEcosystem) Resolve(cmd string, args []string) []string {
	switch cmd {
	case "init":
		return append([]string{"pnpm", "init"}, args...)
	case "install", "i":
		return append([]string{"pnpm", "install"}, args...)
	case "add":
		return append([]string{"pnpm", "add"}, args...)
	case "remove":
		return append([]string{"pnpm", "remove"}, args...)
	default:
		return append([]string{"pnpm", cmd}, args...)
	}
}

func (p *pnpmEcosystem) BuildAuditCommand(_ string, opts audit.Options) ([]string, error) {
	args := []string{"pnpm", "audit", "--json"}
	if opts.ProdOnly {
		args = append(args, "--prod")
	}
	return args, nil
}

func (p *pnpmEcosystem) ParseAuditOutput(_ string, data []byte) (*audit.AuditResult, error) {
	var out pnpmAuditOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	result := &audit.AuditResult{
		Summary: make(map[audit.Severity]int),
		PM:      "pnpm",
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
