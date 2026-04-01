package detector

import (
	"fmt"
	"os"
	"path/filepath"
)

type PackageManager string

const (
	NPM  PackageManager = "npm"
	Yarn PackageManager = "yarn"
	Pnpm PackageManager = "pnpm"
	Bun  PackageManager = "bun"
	Deno PackageManager = "deno"
)

var lockFiles = map[string]PackageManager{
	"package-lock.json": NPM,
	"yarn.lock":         Yarn,
	"pnpm-lock.yaml":    Pnpm,
	"bun.lock":          Bun,
	"bun.lockb":         Bun,
	"deno.lock":         Deno,
}

var projectMarkers = []string{"package.json", "deno.json", "deno.jsonc"}

type Detection struct {
	PM  PackageManager
	Dir string
}

// ErrNoLockFile is returned when a package.json is found but no lock file exists.
type ErrNoLockFile struct {
	Dir string
}

func (e *ErrNoLockFile) Error() string {
	return fmt.Sprintf("no lock file found in %s", e.Dir)
}

// Detect walks up from startDir looking for a directory containing a project
// marker (package.json, deno.json, or deno.jsonc) and at least one known lock
// file. It stops at $HOME.
// Returns all detected package managers in the first matching directory.
func Detect(startDir string) ([]Detection, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := startDir
	var firstProjectDir string
	for {
		if hasProjectMarker(dir) {
			if firstProjectDir == "" {
				firstProjectDir = dir
			}
			var detections []Detection
			seen := make(map[PackageManager]bool)
			for lock, pm := range lockFiles {
				if hasFile(dir, lock) && !seen[pm] {
					seen[pm] = true
					detections = append(detections, Detection{PM: pm, Dir: dir})
				}
			}
			if len(detections) > 0 {
				return detections, nil
			}
		}

		if dir == home || dir == "/" {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if firstProjectDir != "" {
		return nil, &ErrNoLockFile{Dir: firstProjectDir}
	}
	return nil, fmt.Errorf("no project (package.json / deno.json) with a lock file found (searched up to %s)", home)
}

func hasProjectMarker(dir string) bool {
	for _, marker := range projectMarkers {
		if hasFile(dir, marker) {
			return true
		}
	}
	return false
}

// LockFileName returns the lock file name for the given package manager.
func LockFileName(pm PackageManager) string {
	// Bun has two lock files (bun.lock and legacy bun.lockb) in the map.
	// Go map iteration is non-deterministic, so we return the recommended
	// modern format explicitly to avoid returning the legacy binary one.
	if pm == Bun {
		return "bun.lock"
	}
	if pm == Deno {
		return "deno.lock"
	}
	for name, p := range lockFiles {
		if p == pm {
			return name
		}
	}
	return ""
}

func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}
