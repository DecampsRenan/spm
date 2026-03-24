package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/decampsrenan/spm/internal/detector"
)

func TestRunInitPackageJsonExists(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit([]string{"npm"})
	if err == nil {
		t.Fatal("expected error when package.json exists")
	}
	if !strings.Contains(err.Error(), "package.json already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInitInvalidPM(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	err := runInit([]string{"pip"})
	if err == nil {
		t.Fatal("expected error for invalid PM")
	}
	if !strings.Contains(err.Error(), `unknown package manager "pip"`) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunInitDryRunNPM(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withDryRun(t, true)

	err := runInit([]string{"npm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// In dry-run, no package.json should be created
	if _, err := os.Stat(filepath.Join(dir, "package.json")); !os.IsNotExist(err) {
		t.Fatal("package.json should not exist after dry-run")
	}
}

func TestRunInitDryRunYarn(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withDryRun(t, true)

	err := runInit([]string{"yarn"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitDryRunPnpm(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withDryRun(t, true)

	err := runInit([]string{"pnpm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitDryRunBun(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withDryRun(t, true)

	err := runInit([]string{"bun"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitDryRunWithExtraFlags(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withDryRun(t, true)

	// Extra flags like --react should be passed through
	err := runInit([]string{"bun", "--react"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunInitNoArgNonTTY(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Without args and without TTY, should error (can't prompt)
	err := runInit(nil)
	if err == nil {
		t.Fatal("expected error when no PM and no TTY")
	}
}

func TestValidPMs(t *testing.T) {
	expected := map[string]detector.PackageManager{
		"npm":  detector.NPM,
		"yarn": detector.Yarn,
		"pnpm": detector.Pnpm,
		"bun":  detector.Bun,
	}
	for name, want := range expected {
		got, ok := validPMs[name]
		if !ok {
			t.Errorf("validPMs missing %q", name)
			continue
		}
		if got != want {
			t.Errorf("validPMs[%q] = %v, want %v", name, got, want)
		}
	}
}

func TestInitCmdDisablesFlagParsing(t *testing.T) {
	if !initCmd.DisableFlagParsing {
		t.Fatal("initCmd should have DisableFlagParsing=true to pass flags through to PM")
	}
}
