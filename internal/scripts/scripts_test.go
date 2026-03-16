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

	names, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"build", "dev", "lint", "test"}
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i, name := range names {
		if name != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, name, want[i])
		}
	}
}

func TestListNoScripts(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0644)

	names, err := List(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if names != nil {
		t.Errorf("expected nil, got %v", names)
	}
}

func TestListNoPackageJSON(t *testing.T) {
	dir := t.TempDir()

	_, err := List(dir)
	if err == nil {
		t.Fatal("expected error when package.json is missing")
	}
}
