package prompt

import (
	"testing"

	"github.com/decampsrenan/spm/internal/detector"
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
		{PM: detector.NPM, Dir: "/tmp"},
		{PM: detector.Yarn, Dir: "/tmp"},
	}

	_, err := Select(detections)
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
