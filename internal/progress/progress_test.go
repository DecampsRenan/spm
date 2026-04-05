package progress

import (
	"testing"
)

func TestRunEmptyArgs(t *testing.T) {
	err := Run(Config{})
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestRunDryRun(t *testing.T) {
	tests := []struct {
		name   string
		action string
		done   string
	}{
		{"install defaults", "", ""},
		{"custom labels", "Removing", "Removed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Run(Config{
				Args:   []string{"echo", "hello"},
				DryRun: true,
				Action: tt.action,
				Done:   tt.done,
			})
			if err != nil {
				t.Fatalf("dry run returned error: %v", err)
			}
		})
	}
}

func TestRunDryRunEmptyArgs(t *testing.T) {
	err := Run(Config{DryRun: true})
	if err == nil {
		t.Fatal("expected error for empty args in dry-run mode")
	}
}

func TestRunBinaryNotFound(t *testing.T) {
	err := Run(Config{
		Args:   []string{"nonexistent-binary-xyz-12345"},
		Action: "Installing",
		Done:   "Installed",
	})
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}
