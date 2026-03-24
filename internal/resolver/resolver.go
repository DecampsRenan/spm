package resolver

import "github.com/decampsrenan/spm/internal/detector"

// Resolve translates an spm command + args into the actual package manager command.
// command is the spm verb (install, add, or a script name).
// args are extra arguments (package names, flags, etc.).
func Resolve(pm detector.PackageManager, command string, args []string) []string {
	bin := string(pm)

	switch command {
	case "init":
		switch pm {
		case detector.Pnpm:
			// pnpm init is already non-interactive
			return append([]string{bin, "init"}, args...)
		default:
			// npm, yarn, bun need -y for non-interactive init
			return append([]string{bin, "init", "-y"}, args...)
		}

	case "install", "i":
		return append([]string{bin, "install"}, args...)

	case "add":
		switch pm {
		case detector.NPM:
			return append([]string{bin, "install"}, args...)
		default:
			return append([]string{bin, "add"}, args...)
		}

	case "remove":
		switch pm {
		case detector.NPM:
			return append([]string{bin, "uninstall"}, args...)
		default:
			return append([]string{bin, "remove"}, args...)
		}

	default:
		// Fallback: treat as a script run
		switch pm {
		case detector.NPM:
			return append([]string{bin, "run", command}, args...)
		default:
			// yarn, pnpm, and bun don't need explicit "run"
			return append([]string{bin, command}, args...)
		}
	}
}
