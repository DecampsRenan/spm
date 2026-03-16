package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Script holds a script name and its command string.
type Script struct {
	Name    string
	Command string
}

// List reads package.json from dir and returns sorted scripts.
func List(dir string) ([]Script, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, fmt.Errorf("cannot read package.json: %w", err)
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("cannot parse package.json: %w", err)
	}

	if len(pkg.Scripts) == 0 {
		return nil, nil
	}

	scripts := make([]Script, 0, len(pkg.Scripts))
	for name, cmd := range pkg.Scripts {
		scripts = append(scripts, Script{Name: name, Command: cmd})
	}
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Name < scripts[j].Name
	})
	return scripts, nil
}
