package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstNonFlagArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"script only", []string{"dev"}, "dev"},
		{"flag then script", []string{"--dry-run", "dev"}, "dev"},
		{"script then flag", []string{"dev", "--dry-run"}, "dev"},
		{"multiple flags then script", []string{"--dry-run", "--verbose", "test"}, "test"},
		{"known command", []string{"install"}, "install"},
		{"flag then known command", []string{"--dry-run", "install"}, "install"},
		{"only flags", []string{"--dry-run", "--help"}, ""},
		{"empty args", []string{}, ""},
		{"short flag then script", []string{"-v", "dev"}, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonFlagArg(tt.args)
			if got != tt.want {
				t.Errorf("firstNonFlagArg(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}

func TestInstallCmdWhitelistsUnknownFlags(t *testing.T) {
	if !installCmd.FParseErrWhitelist.UnknownFlags {
		t.Fatal("installCmd should whitelist unknown flags for pass-through")
	}
}

func TestAddCmdAcceptsNoArgs(t *testing.T) {
	// addCmd no longer requires args — zero args triggers interactive search TUI
	if addCmd.Args != nil {
		err := addCmd.Args(addCmd, []string{})
		if err != nil {
			t.Fatalf("addCmd should accept zero args (interactive mode): %v", err)
		}
	}
}

func TestAddCmdAcceptsArgs(t *testing.T) {
	// addCmd with args should still work (direct add)
	if addCmd.Args != nil {
		err := addCmd.Args(addCmd, []string{"react"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestRemoveCmdRequiresArgs(t *testing.T) {
	err := removeCmd.Args(removeCmd, []string{})
	if err == nil {
		t.Fatal("expected error for empty args")
	}
	if !strings.Contains(err.Error(), "specify at least one package to remove") {
		t.Errorf("expected custom error message, got: %v", err)
	}
}

func TestRemoveCmdAcceptsArgs(t *testing.T) {
	err := removeCmd.Args(removeCmd, []string{"react"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// runClean test helpers

func setupCleanDir(t *testing.T, hasNodeModules bool, lockFile string) string {
	t.Helper()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if lockFile != "" {
		if err := os.WriteFile(filepath.Join(dir, lockFile), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if hasNodeModules {
		nm := filepath.Join(dir, "node_modules", ".package-lock.json")
		if err := os.MkdirAll(filepath.Join(dir, "node_modules"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(nm, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return dir
}

func withDryRun(t *testing.T, val bool) {
	t.Helper()
	old := dryRun
	dryRun = val
	t.Cleanup(func() { dryRun = old })
}

func TestRunCleanDryRun(t *testing.T) {
	dir := setupCleanDir(t, true, "package-lock.json")
	t.Chdir(dir)
	withDryRun(t, true)

	if err := runClean(false, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); os.IsNotExist(err) {
		t.Fatal("node_modules should still exist after dry-run")
	}
}

func TestRunCleanRemovesNodeModules(t *testing.T) {
	dir := setupCleanDir(t, true, "package-lock.json")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(false, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); !os.IsNotExist(err) {
		t.Fatal("node_modules should be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "package-lock.json")); os.IsNotExist(err) {
		t.Fatal("lock file should still exist when --lock not set")
	}
}

func TestRunCleanWithLock(t *testing.T) {
	dir := setupCleanDir(t, true, "package-lock.json")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); !os.IsNotExist(err) {
		t.Fatal("node_modules should be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "package-lock.json")); !os.IsNotExist(err) {
		t.Fatal("lock file should be removed with --lock")
	}
}

func TestRunCleanDryRunWithLock(t *testing.T) {
	dir := setupCleanDir(t, true, "package-lock.json")
	t.Chdir(dir)
	withDryRun(t, true)

	if err := runClean(true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); os.IsNotExist(err) {
		t.Fatal("node_modules should still exist after dry-run")
	}
	if _, err := os.Stat(filepath.Join(dir, "package-lock.json")); os.IsNotExist(err) {
		t.Fatal("lock file should still exist after dry-run")
	}
}

func TestRunCleanNoNodeModules(t *testing.T) {
	dir := setupCleanDir(t, false, "package-lock.json")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(false, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCleanNoLockFileWithoutLockFlag(t *testing.T) {
	// package.json exists but no lock file — should not error when --lock is false
	dir := setupCleanDir(t, true, "")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(false, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); !os.IsNotExist(err) {
		t.Fatal("node_modules should be removed")
	}
}

func TestRunCleanNoPackageJSON(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	withDryRun(t, false)

	err := runClean(false, true)
	if err == nil {
		t.Fatal("expected error when no package.json found")
	}
}

func TestRunCleanNothingToRemove(t *testing.T) {
	// package.json + lock file exist, but no node_modules — should print "Nothing to remove"
	dir := setupCleanDir(t, false, "yarn.lock")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(false, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCleanWithLockYarn(t *testing.T) {
	dir := setupCleanDir(t, true, "yarn.lock")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); !os.IsNotExist(err) {
		t.Fatal("node_modules should be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "yarn.lock")); !os.IsNotExist(err) {
		t.Fatal("yarn.lock should be removed with --lock")
	}
}

func TestRunCleanWithLockPnpm(t *testing.T) {
	dir := setupCleanDir(t, true, "pnpm-lock.yaml")
	t.Chdir(dir)
	withDryRun(t, false)

	if err := runClean(true, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "node_modules")); !os.IsNotExist(err) {
		t.Fatal("node_modules should be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "pnpm-lock.yaml")); !os.IsNotExist(err) {
		t.Fatal("pnpm-lock.yaml should be removed with --lock")
	}
}
