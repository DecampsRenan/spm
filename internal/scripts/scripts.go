package scripts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

// ListDeno reads deno.json (or deno.jsonc) from dir and returns sorted tasks.
func ListDeno(dir string) ([]Script, error) {
	data, err := os.ReadFile(filepath.Join(dir, "deno.json"))
	if err != nil {
		// Fall back to deno.jsonc.
		data, err = os.ReadFile(filepath.Join(dir, "deno.jsonc"))
		if err != nil {
			return nil, fmt.Errorf("cannot read deno.json or deno.jsonc: %w", err)
		}
		data = stripJSONComments(data)
	}

	var cfg struct {
		Tasks map[string]string `json:"tasks"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse deno.json: %w", err)
	}

	if len(cfg.Tasks) == 0 {
		return nil, nil
	}

	scripts := make([]Script, 0, len(cfg.Tasks))
	for name, cmd := range cfg.Tasks {
		scripts = append(scripts, Script{Name: name, Command: cmd})
	}
	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].Name < scripts[j].Name
	})
	return scripts, nil
}

// stripJSONComments removes single-line (//) and multi-line (/* */) comments
// from JSONC content while preserving strings that contain comment-like sequences.
func stripJSONComments(data []byte) []byte {
	s := string(data)
	var b strings.Builder
	b.Grow(len(s))

	i := 0
	for i < len(s) {
		// String literal — copy verbatim including any comment-like content.
		if s[i] == '"' {
			b.WriteByte(s[i])
			i++
			for i < len(s) {
				b.WriteByte(s[i])
				if s[i] == '\\' {
					i++
					if i < len(s) {
						b.WriteByte(s[i])
					}
				} else if s[i] == '"' {
					i++
					break
				}
				i++
			}
			continue
		}
		// Single-line comment.
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '/' {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		// Multi-line comment.
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '*' {
			i += 2
			for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
				i++
			}
			if i+1 < len(s) {
				i += 2
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return []byte(b.String())
}
