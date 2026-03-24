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

func TestListDeno(t *testing.T) {
	dir := t.TempDir()
	cfg := `{"tasks":{"dev":"deno run --watch main.ts","build":"deno compile main.ts","test":"deno test"}}`
	os.WriteFile(filepath.Join(dir, "deno.json"), []byte(cfg), 0644)

	scripts, err := ListDeno(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Script{
		{Name: "build", Command: "deno compile main.ts"},
		{Name: "dev", Command: "deno run --watch main.ts"},
		{Name: "test", Command: "deno test"},
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

func TestListDenoJsonc(t *testing.T) {
	dir := t.TempDir()
	cfg := `{
  // Development tasks
  "tasks": {
    "dev": "deno run --watch main.ts",
    /* build task */
    "build": "deno compile main.ts"
  }
}`
	os.WriteFile(filepath.Join(dir, "deno.jsonc"), []byte(cfg), 0644)

	scripts, err := ListDeno(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Script{
		{Name: "build", Command: "deno compile main.ts"},
		{Name: "dev", Command: "deno run --watch main.ts"},
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

func TestListDenoFallbackToJsonc(t *testing.T) {
	dir := t.TempDir()
	// No deno.json, only deno.jsonc
	cfg := `{"tasks":{"dev":"deno run main.ts"}}`
	os.WriteFile(filepath.Join(dir, "deno.jsonc"), []byte(cfg), 0644)

	scripts, err := ListDeno(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(scripts) != 1 || scripts[0].Name != "dev" {
		t.Errorf("expected 1 script 'dev', got %v", scripts)
	}
}

func TestListDenoNoTasks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "deno.json"), []byte(`{"imports":{}}`), 0644)

	scripts, err := ListDeno(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scripts != nil {
		t.Errorf("expected nil, got %v", scripts)
	}
}

func TestListDenoNoFile(t *testing.T) {
	dir := t.TempDir()

	_, err := ListDeno(dir)
	if err == nil {
		t.Fatal("expected error when deno.json is missing")
	}
}

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single-line comment",
			input: "{\n// comment\n\"a\": 1\n}",
			want:  "{\n\n\"a\": 1\n}",
		},
		{
			name:  "multi-line comment",
			input: "{/* comment */\"a\": 1}",
			want:  "{\"a\": 1}",
		},
		{
			name:  "comment-like in string",
			input: `{"url": "https://example.com"}`,
			want:  `{"url": "https://example.com"}`,
		},
		{
			name:  "escaped quote in string",
			input: `{"a": "he said \"// not a comment\""}`,
			want:  `{"a": "he said \"// not a comment\""}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(stripJSONComments([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
