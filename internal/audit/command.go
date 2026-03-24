package audit

import (
	"fmt"
	"os/exec"
	"strings"
)

// buildCommand returns the audit command args for the given package manager.
func buildCommand(pm string, dir string, opts Options) ([]string, error) {
	switch pm {
	case "npm":
		return buildNPM(opts), nil
	case "yarn":
		version, err := detectYarnVersion(dir)
		if err != nil {
			return nil, fmt.Errorf("cannot detect yarn version: %w", err)
		}
		if version >= 2 {
			return buildYarnBerry(opts), nil
		}
		return buildYarnClassic(opts), nil
	case "pnpm":
		return buildPnpm(opts), nil
	default:
		return nil, fmt.Errorf("unsupported package manager: %s", pm)
	}
}

func buildNPM(opts Options) []string {
	args := []string{"npm", "audit", "--json"}
	if opts.ProdOnly {
		args = append(args, "--omit=dev")
	}
	return args
}

func buildYarnClassic(opts Options) []string {
	args := []string{"yarn", "audit", "--json"}
	if opts.ProdOnly {
		args = append(args, "--groups", "dependencies")
	}
	return args
}

func buildYarnBerry(opts Options) []string {
	args := []string{"yarn", "npm", "audit", "--all", "--json"}
	if opts.ProdOnly {
		// Yarn Berry doesn't have a direct --prod flag for audit;
		// we filter post-parse. But --severity can be passed.
	}
	return args
}

func buildPnpm(opts Options) []string {
	args := []string{"pnpm", "audit", "--json"}
	if opts.ProdOnly {
		args = append(args, "--prod")
	}
	return args
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
