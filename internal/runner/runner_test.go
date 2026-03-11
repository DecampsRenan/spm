package runner

import (
	"testing"
)

func TestRunDryRun(t *testing.T) {
	err := Run([]string{"echo", "hello"}, true)
	if err != nil {
		t.Fatalf("dry run returned error: %v", err)
	}
}

func TestRunEmptyArgs(t *testing.T) {
	err := Run([]string{}, false)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestRunDryRunEmptyArgs(t *testing.T) {
	err := Run([]string{}, true)
	if err == nil {
		t.Fatal("expected error for empty args in dry-run mode")
	}
}

func TestRunBinaryNotFound(t *testing.T) {
	// Use a binary name that definitely doesn't exist
	err := Run([]string{"nonexistent-binary-xyz-12345"}, false)
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}
