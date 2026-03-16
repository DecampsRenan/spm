package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// List reads package.json from dir and returns sorted script names.
func List(dir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, fmt.Errorf("cannot read package.json: %w", err)
	}

	var pkg struct {
		Scripts map[string]json.RawMessage `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("cannot parse package.json: %w", err)
	}

	if len(pkg.Scripts) == 0 {
		return nil, nil
	}

	names := make([]string, 0, len(pkg.Scripts))
	for name := range pkg.Scripts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
