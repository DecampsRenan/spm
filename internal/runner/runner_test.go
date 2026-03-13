package runner

import (
	"testing"
)

func TestRunDryRun(t *testing.T) {
	err := Run([]string{"echo", "hello"}, true, false, false)
	if err != nil {
		t.Fatalf("dry run returned error: %v", err)
	}
}

func TestRunEmptyArgs(t *testing.T) {
	err := Run([]string{}, false, false, false)
	if err == nil {
		t.Fatal("expected error for empty args")
	}
}

func TestRunDryRunEmptyArgs(t *testing.T) {
	err := Run([]string{}, true, false, false)
	if err == nil {
		t.Fatal("expected error for empty args in dry-run mode")
	}
}

func TestRunBinaryNotFound(t *testing.T) {
	// Use a binary name that definitely doesn't exist
	err := Run([]string{"nonexistent-binary-xyz-12345"}, false, false, false)
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestRunVibesMode(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	// vibes mode runs the command as a child process instead of syscall.Exec
	err := Run([]string{"echo", "hello"}, false, true, false)
	if err != nil {
		t.Fatalf("vibes mode returned error: %v", err)
	}
}

func TestRunVibesDryRun(t *testing.T) {
	// dry-run should take precedence over vibes
	err := Run([]string{"echo", "hello"}, true, true, false)
	if err != nil {
		t.Fatalf("dry run with vibes returned error: %v", err)
	}
}

func TestRunVibesBinaryNotFound(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	err := Run([]string{"nonexistent-binary-xyz-12345"}, false, true, false)
	if err == nil {
		t.Fatal("expected error for missing binary in vibes mode")
	}
}

func TestRunVibesEmptyArgs(t *testing.T) {
	err := Run([]string{}, false, true, false)
	if err == nil {
		t.Fatal("expected error for empty args in vibes mode")
	}
}
