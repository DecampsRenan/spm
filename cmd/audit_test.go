package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withAuditSeverity(t *testing.T, val string) {
	t.Helper()
	old := auditSeverity
	auditSeverity = val
	t.Cleanup(func() { auditSeverity = old })
}

func TestAuditCmdRegistered(t *testing.T) {
	var found bool
	for _, c := range rootCmd.Commands() {
		if c.Use == "audit" {
			found = true
			break
		}
	}
	if !found {
		t.Error("auditCmd should be registered under rootCmd")
	}
}

func TestAuditCmdFlags(t *testing.T) {
	for _, name := range []string{"prod-only", "json", "severity"} {
		t.Run(name, func(t *testing.T) {
			if auditCmd.Flags().Lookup(name) == nil {
				t.Errorf("auditCmd missing flag %q", name)
			}
		})
	}
}

func TestAuditCmdShortHelpPresent(t *testing.T) {
	if auditCmd.Short == "" {
		t.Error("auditCmd.Short should be set")
	}
	if auditCmd.Long == "" {
		t.Error("auditCmd.Long should be set")
	}
}

func TestAuditCmdRunNoProject(t *testing.T) {
	// Empty temp dir — detector should fail to find a manifest.
	dir := t.TempDir()
	t.Chdir(dir)
	withAuditSeverity(t, "")

	err := auditCmd.RunE(auditCmd, nil)
	if err == nil {
		t.Fatal("expected error when no manifest found")
	}
}

func TestAuditCmdRunInvalidSeverity(t *testing.T) {
	// Set up a minimal npm project so detector returns a single detection,
	// which lets RunE reach the severity-validation branch.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	withAuditSeverity(t, "bogus")

	err := auditCmd.RunE(auditCmd, nil)
	if err == nil {
		t.Fatal("expected error for invalid severity")
	}
	if !strings.Contains(err.Error(), "invalid severity") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAuditCmdRunValidSeverityParses(t *testing.T) {
	// A valid severity should parse without returning the "invalid severity" error.
	// We use --dry-run to avoid executing the real `npm audit` command.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	withDryRun(t, true)
	withAuditSeverity(t, "HIGH") // case-insensitive

	if err := auditCmd.RunE(auditCmd, nil); err != nil {
		t.Fatalf("unexpected error with valid severity: %v", err)
	}
}

func TestAuditCmdSeverityFlagTypeIsString(t *testing.T) {
	f := auditCmd.Flags().Lookup("severity")
	if f == nil {
		t.Fatal("severity flag missing")
	}
	if f.Value.Type() != "string" {
		t.Errorf("severity flag type = %q, want string", f.Value.Type())
	}
}

func TestAuditCmdBooleanFlagsDefaultFalse(t *testing.T) {
	for _, name := range []string{"prod-only", "json"} {
		t.Run(name, func(t *testing.T) {
			f := auditCmd.Flags().Lookup(name)
			if f == nil {
				t.Fatalf("flag %q missing", name)
			}
			if f.Value.Type() != "bool" {
				t.Errorf("flag %q type = %q, want bool", name, f.Value.Type())
			}
			if f.DefValue != "false" {
				t.Errorf("flag %q default = %q, want false", name, f.DefValue)
			}
		})
	}
}
