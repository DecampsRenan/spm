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

	case "remove":
		switch pm {
		case detector.NPM:
			return append([]string{bin, "uninstall"}, args...)
		default:
			return append([]string{bin, "remove"}, args...)
		}

	default:
		// Fallback: treat as a script/task run
		switch pm {
		case detector.NPM:
			return append([]string{bin, "run", command}, args...)
		case detector.Deno:
			return append([]string{bin, "task", command}, args...)
		default:
			// yarn, pnpm, and bun don't need explicit "run"
			return append([]string{bin, command}, args...)
		}
	}
}
