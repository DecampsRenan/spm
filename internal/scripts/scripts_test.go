package scripts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	dir := t.TempDir()
	pkg := `{"name":"test","scripts":{"dev":"vite","build":"tsc","test":"vitest","lint":"eslint ."}}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkg), 0644)

	scripts, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Script{
		{Name: "build", Command: "tsc"},
		{Name: "dev", Command: "vite"},
		{Name: "lint", Command: "eslint ."},
		{Name: "test", Command: "vitest"},
	}
	if len(scripts) != len(want) {
		t.Fatalf("got %d scripts, want %d", len(scripts), len(want))
	}
	for i, s := range scripts {
		if s.Name != want[i].Name || s.Command != want[i].Command {
			t.Errorf("scripts[%d] = %+v, want %+v", i, s, want[i])
		}
	}
}

func TestListNoScripts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644)

	scripts, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scripts != nil {
		t.Errorf("expected nil, got %v", scripts)
	}
}

func TestListNoPackageJSON(t *testing.T) {
	dir := t.TempDir()

	_, err := List(dir)
	if err == nil {
		t.Fatal("expected error when package.json is missing")
	}
}
