package ecosystem

import (
	"fmt"

	"github.com/decampsrenan/spm/internal/audit"
)

type bunEcosystem struct{}

func (b *bunEcosystem) Name() PackageManager     { return Bun }
func (b *bunEcosystem) ManifestFile() string     { return "package.json" }
func (b *bunEcosystem) LockFiles() []string      { return []string{"bun.lock", "bun.lockb"} }
func (b *bunEcosystem) ArtifactDirs() []string   { return []string{"node_modules"} }
func (b *bunEcosystem) HasCommand(_ string) bool { return true }

func (b *bunEcosystem) Resolve(cmd string, args []string) []string {
	switch cmd {
	case "init":
		return append([]string{"bun", "init"}, args...)
	case "install", "i":
		return append([]string{"bun", "install"}, args...)
	case "add":
		return append([]string{"bun", "add"}, args...)
	case "remove":
		return append([]string{"bun", "remove"}, args...)
	default:
		return append([]string{"bun", cmd}, args...)
	}
}

func (b *bunEcosystem) BuildAuditCommand(_ string, _ audit.Options) ([]string, error) {
	return nil, fmt.Errorf("bun does not have a built-in audit command")
}

func (b *bunEcosystem) ParseAuditOutput(_ string, _ []byte) (*audit.AuditResult, error) {
	return nil, fmt.Errorf("bun does not have a built-in audit command")
}
