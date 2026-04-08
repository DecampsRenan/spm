package prompt

import (
	"testing"

	"github.com/decampsrenan/spm/internal/detector"
	"github.com/decampsrenan/spm/internal/ecosystem"
)

func TestConfirmNonTTY(t *testing.T) {
	// In a test environment, stdin is not a TTY, so Confirm should return an error.
	_, err := Confirm("Delete everything?")
	if err == nil {
		t.Fatal("expected error when stdin is not a TTY")
	}
}

func TestSelectNonTTY(t *testing.T) {
	// In a test environment, stdin is not a TTY, so Select should return an error.
	detections := []detector.Detection{
		{PM: ecosystem.NPM, Dir: "/tmp"},
		{PM: ecosystem.Yarn, Dir: "/tmp"},
	}

	_, err := Select(detections)
	if err == nil {
		t.Fatal("expected error when stdin is not a TTY")
	}
}

func TestSelectScriptNonTTY(t *testing.T) {
	_, err := SelectScript(
		[]string{"dev", "build", "test"},
		[]string{"vite", "tsc", "vitest"},
	)
	if err == nil {
		t.Fatal("expected error when stdin is not a TTY")
	}
}

func TestSelectFromAllNonTTY(t *testing.T) {
	_, err := SelectFromAll("/tmp")
	if err == nil {
		t.Fatal("expected error when stdin is not a TTY")
	}
}
