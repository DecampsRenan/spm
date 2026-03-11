package resolver

import "github.com/decampsrenan/spm/internal/detector"

// Resolve translates an spm command + args into the actual package manager command.
// command is the spm verb (install, add, or a script name).
// args are extra arguments (package names, flags, etc.).
func Resolve(pm detector.PackageManager, command string, args []string) []string {
	bin := string(pm)

	switch command {
	case "install", "i":
		return append([]string{bin, "install"}, args...)

	case "add":
		switch pm {
		case detector.NPM:
			return append([]string{bin, "install"}, args...)
		default:
			return append([]string{bin, "add"}, args...)
		}

	default:
		// Fallback: treat as a script run
		switch pm {
		case detector.NPM:
			return append([]string{bin, "run", command}, args...)
		default:
			// yarn and pnpm don't need explicit "run"
			return append([]string{bin, command}, args...)
		}
	}
}
