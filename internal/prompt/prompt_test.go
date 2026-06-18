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

func TestFindDetectionByPM_Found(t *testing.T) {
	detections := []detector.Detection{
		{PM: ecosystem.NPM, Dir: "/a"},
		{PM: ecosystem.Yarn, Dir: "/b"},
		{PM: ecosystem.Pnpm, Dir: "/c"},
	}
	got, err := findDetectionByPM(detections, string(ecosystem.Yarn))
	if err != nil {
		t.Fatal(err)
	}
	if got.PM != ecosystem.Yarn || got.Dir != "/b" {
		t.Errorf("unexpected detection: %+v", got)
	}
}

func TestFindDetectionByPM_NotFound(t *testing.T) {
	detections := []detector.Detection{{PM: ecosystem.NPM, Dir: "/a"}}
	_, err := findDetectionByPM(detections, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown PM")
	}
}

func TestFindDetectionByPM_EmptySlice(t *testing.T) {
	_, err := findDetectionByPM(nil, string(ecosystem.NPM))
	if err == nil {
		t.Fatal("expected error for empty detections")
	}
}

func TestTruncateCmd(t *testing.T) {
	tests := []struct {
		name   string
		cmd    string
		maxLen int
		want   string
	}{
		{"below limit", "vite", 40, "vite"},
		{"at limit", "1234567890", 10, "1234567890"},
		{"truncated", "12345678901234567890", 10, "123456789…"},
		{"empty", "", 10, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateCmd(tt.cmd, tt.maxLen); got != tt.want {
				t.Errorf("truncateCmd(%q, %d) = %q, want %q", tt.cmd, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestScriptOptionLabel_ContainsNameAndCmd(t *testing.T) {
	label := scriptOptionLabel("dev", "vite")
	if label == "" {
		t.Fatal("expected non-empty label")
	}
	// The label is "name — dim(cmd)"; the name and cmd should both appear as substrings.
	if !containsStr(label, "dev") {
		t.Errorf("label missing name: %q", label)
	}
	if !containsStr(label, "vite") {
		t.Errorf("label missing cmd: %q", label)
	}
}

func TestScriptOptionLabel_TruncatesLongCmd(t *testing.T) {
	long := "this-is-a-very-long-command-that-should-be-truncated-by-the-helper"
	label := scriptOptionLabel("build", long)
	if !containsStr(label, "…") {
		t.Errorf("expected ellipsis in truncated label, got %q", label)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
