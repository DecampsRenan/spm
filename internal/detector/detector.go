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
)

var lockFiles = map[string]PackageManager{
	"package-lock.json": NPM,
	"yarn.lock":         Yarn,
	"pnpm-lock.yaml":    Pnpm,
}

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

// Detect walks up from startDir looking for a directory containing package.json
// and at least one known lock file. It stops at $HOME.
// Returns all detected package managers in the first matching directory.
func Detect(startDir string) ([]Detection, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := startDir
	var firstPackageJSONDir string
	for {
		if hasFile(dir, "package.json") {
			if firstPackageJSONDir == "" {
				firstPackageJSONDir = dir
			}
			var detections []Detection
			for lock, pm := range lockFiles {
				if hasFile(dir, lock) {
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

	if firstPackageJSONDir != "" {
		return nil, &ErrNoLockFile{Dir: firstPackageJSONDir}
	}
	return nil, fmt.Errorf("no package.json with a lock file found (searched up to %s)", home)
}

// LockFileName returns the lock file name for the given package manager.
func LockFileName(pm PackageManager) string {
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
