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
	err := Run([]string{"nonexistent-binary-xyz-12345"}, false, false, false)
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestRunVibesMode(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	err := Run([]string{"echo", "hello"}, false, true, false)
	if err != nil {
		t.Fatalf("vibes mode returned error: %v", err)
	}
}

func TestRunVibesDryRun(t *testing.T) {
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

func TestRunNotifyMode(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	err := Run([]string{"echo", "hello"}, false, false, true)
	if err != nil {
		t.Fatalf("notify mode returned error: %v", err)
	}
}

func TestRunNotifyDryRun(t *testing.T) {
	err := Run([]string{"echo", "hello"}, true, false, true)
	if err != nil {
		t.Fatalf("dry run with notify returned error: %v", err)
	}
}

func TestRunVibesAndNotify(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	err := Run([]string{"echo", "hello"}, false, true, true)
	if err != nil {
		t.Fatalf("vibes+notify mode returned error: %v", err)
	}
}

func TestRunNotifyEmptyArgs(t *testing.T) {
	err := Run([]string{}, false, false, true)
	if err == nil {
		t.Fatal("expected error for empty args in notify mode")
	}
}

func TestRunNotifyBinaryNotFound(t *testing.T) {
	t.Setenv("SPM_DISABLE_AUDIO", "1")
	err := Run([]string{"nonexistent-binary-xyz-12345"}, false, false, true)
	if err == nil {
		t.Fatal("expected error for missing binary in notify mode")
	}
}
