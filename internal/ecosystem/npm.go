package ecosystem

import (
	"encoding/json"

	"github.com/decampsrenan/spm/internal/audit"
)

type npmEcosystem struct{}

func (n *npmEcosystem) Name() PackageManager     { return NPM }
func (n *npmEcosystem) ManifestFile() string     { return "package.json" }
func (n *npmEcosystem) LockFiles() []string      { return []string{"package-lock.json"} }
func (n *npmEcosystem) ArtifactDirs() []string   { return []string{"node_modules"} }
func (n *npmEcosystem) HasCommand(_ string) bool { return true }

func (n *npmEcosystem) Resolve(cmd string, args []string) []string {
	switch cmd {
	case "init":
		return append([]string{"npm", "init", "-y"}, args...)
	case "install", "i":
		return append([]string{"npm", "install"}, args...)
	case "add":
		return append([]string{"npm", "install"}, args...)
	case "remove":
		return append([]string{"npm", "uninstall"}, args...)
	default:
		return append([]string{"npm", "run", cmd}, args...)
	}
}

func (n *npmEcosystem) BuildAuditCommand(_ string, opts audit.Options) ([]string, error) {
	args := []string{"npm", "audit", "--json"}
	if opts.ProdOnly {
		args = append(args, "--omit=dev")
	}
	return args, nil
}

func (n *npmEcosystem) ParseAuditOutput(_ string, data []byte) (*audit.AuditResult, error) {
	var out npmAuditOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	result := &audit.AuditResult{
		Summary: make(map[audit.Severity]int),
		PM:      "npm",
	}

	for name, v := range out.Vulnerabilities {
		sev := audit.Severity(v.Severity)
		title := extractNPMTitle(v.Via)
		url := extractNPMURL(v.Via)
		result.Vulnerabilities = append(result.Vulnerabilities, audit.Vulnerability{
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

// npm audit JSON types

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
