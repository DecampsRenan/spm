package detector

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/decampsrenan/spm/internal/ecosystem"
)

// lockFileMap is built from the registered ecosystems at init time.
var lockFileMap map[string]ecosystem.PackageManager

// manifestFiles is the set of manifest files across all ecosystems.
var manifestFiles map[string]bool

func init() {
	lockFileMap = make(map[string]ecosystem.PackageManager)
	manifestFiles = make(map[string]bool)
	for _, eco := range ecosystem.All() {
		for _, lf := range eco.LockFiles() {
			lockFileMap[lf] = eco.Name()
		}
		manifestFiles[eco.ManifestFile()] = true
	}
}

type Detection struct {
	PM  ecosystem.PackageManager
	Dir string
}

// ErrNoLockFile is returned when a manifest is found but no lock file exists.
type ErrNoLockFile struct {
	Dir string
}

func (e *ErrNoLockFile) Error() string {
	return fmt.Sprintf("no lock file found in %s", e.Dir)
}

// Detect walks up from startDir looking for a directory containing a manifest
// file and at least one known lock file. It stops at $HOME.
// Returns all detected package managers in the first matching directory.
func Detect(startDir string) ([]Detection, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := startDir
	var firstManifestDir string
	for {
		if hasManifest(dir) {
			if firstManifestDir == "" {
				firstManifestDir = dir
			}
			var detections []Detection
			seen := make(map[ecosystem.PackageManager]bool)
			for lock, pm := range lockFileMap {
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

	if firstManifestDir != "" {
		return nil, &ErrNoLockFile{Dir: firstManifestDir}
	}
	return nil, fmt.Errorf("no project manifest with a lock file found (searched up to %s)", home)
}

// LockFileName returns the preferred lock file name for the given package manager.
func LockFileName(pm ecosystem.PackageManager) string {
	eco := ecosystem.ForPM(pm)
	if eco == nil {
		return ""
	}
	locks := eco.LockFiles()
	if len(locks) == 0 {
		return ""
	}
	// Return the first (preferred) lock file.
	return locks[0]
}

// hasManifest checks if any known manifest file exists in the directory.
func hasManifest(dir string) bool {
	for mf := range manifestFiles {
		if hasFile(dir, mf) {
			return true
		}
	}
	return false
}

func hasFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}
