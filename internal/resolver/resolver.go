package resolver

import "github.com/decampsrenan/spm/internal/ecosystem"

// Resolve translates an spm command + args into the actual package manager command.
// command is the spm verb (install, add, or a script name).
// args are extra arguments (package names, flags, etc.).
func Resolve(pm ecosystem.PackageManager, command string, args []string) []string {
	eco := ecosystem.ForPM(pm)
	if eco == nil {
		// Unknown PM — best effort: use PM name as binary.
		return append([]string{string(pm), command}, args...)
	}
	return eco.Resolve(command, args)
}
