package ecosystem

import "github.com/decampsrenan/spm/internal/audit"

// PackageManager identifies a package manager by name.
type PackageManager string

const (
	NPM  PackageManager = "npm"
	Yarn PackageManager = "yarn"
	Pnpm PackageManager = "pnpm"
	Bun  PackageManager = "bun"
)

// Ecosystem abstracts the behavior of a package manager ecosystem.
// Each implementation knows how to resolve commands, build audit commands,
// parse audit output, and describe its manifest/lock/artifact files.
type Ecosystem interface {
	// Name returns the package manager identifier.
	Name() PackageManager

	// ManifestFile returns the main project manifest (e.g. "package.json").
	ManifestFile() string

	// LockFiles returns all possible lock file names for this ecosystem.
	LockFiles() []string

	// ArtifactDirs returns directories that can be safely cleaned
	// (e.g. "node_modules", "target").
	ArtifactDirs() []string

	// Resolve translates an spm command + args into the native PM command.
	// Returns the full command slice (binary + args).
	Resolve(cmd string, args []string) []string

	// HasCommand reports whether this ecosystem supports the given spm command.
	HasCommand(cmd string) bool

	// BuildAuditCommand returns the command to run a security audit.
	BuildAuditCommand(dir string, opts audit.Options) ([]string, error)

	// ParseAuditOutput parses the raw audit command output into a normalized result.
	ParseAuditOutput(dir string, data []byte) (*audit.AuditResult, error)
}
