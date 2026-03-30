package ui

import (
	"strings"
	"testing"
)

func TestSuccessContainsMessage(t *testing.T) {
	result := Success("file removed")
	if !strings.Contains(result, "file removed") {
		t.Errorf("Success() should contain the message, got: %s", result)
	}
}

func TestErrorContainsMessage(t *testing.T) {
	result := Error("something broke")
	if !strings.Contains(result, "something broke") {
		t.Errorf("Error() should contain the message, got: %s", result)
	}
}

func TestWarningContainsMessage(t *testing.T) {
	result := Warning("watch out")
	if !strings.Contains(result, "watch out") {
		t.Errorf("Warning() should contain the message, got: %s", result)
	}
}

func TestInfoContainsMessage(t *testing.T) {
	result := Info("some info")
	if !strings.Contains(result, "some info") {
		t.Errorf("Info() should contain the message, got: %s", result)
	}
}

func TestDimContainsMessage(t *testing.T) {
	result := Dim("secondary text")
	if !strings.Contains(result, "secondary text") {
		t.Errorf("Dim() should contain the message, got: %s", result)
	}
}

func TestCommandContainsArgs(t *testing.T) {
	result := Command([]string{"npm", "install", "react"})
	if !strings.Contains(result, "npm install react") {
		t.Errorf("Command() should contain joined args, got: %s", result)
	}
	if !strings.Contains(result, "Would run") {
		t.Errorf("Command() should contain 'Would run' label, got: %s", result)
	}
}

func TestHeaderContainsMessage(t *testing.T) {
	result := Header("title")
	if !strings.Contains(result, "title") {
		t.Errorf("Header() should contain the message, got: %s", result)
	}
}

func TestPathContainsPath(t *testing.T) {
	result := Path("/foo/bar/node_modules")
	if !strings.Contains(result, "/foo/bar/node_modules") {
		t.Errorf("Path() should contain the path, got: %s", result)
	}
}

func TestDimGradientContainsMessage(t *testing.T) {
	result := DimGradient("log line", 0, 5)
	if !strings.Contains(result, "log line") {
		t.Errorf("DimGradient() should contain the message, got: %s", result)
	}
}

func TestDimGradientSingleLine(t *testing.T) {
	// With total <= 1, should fall back to normal Dim style.
	result := DimGradient("only line", 0, 1)
	if !strings.Contains(result, "only line") {
		t.Errorf("DimGradient() with total=1 should contain the message, got: %s", result)
	}
}

func TestDimGradientAllLevels(t *testing.T) {
	total := 5
	for i := 0; i < total; i++ {
		result := DimGradient("line", i, total)
		if !strings.Contains(result, "line") {
			t.Errorf("DimGradient(level=%d) should contain the message, got: %s", i, result)
		}
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b uint8
		t    float64
		want uint8
	}{
		{0, 100, 0.0, 0},
		{0, 100, 1.0, 100},
		{0, 100, 0.5, 50},
		{100, 0, 0.5, 50},       // reverse direction
		{0x37, 0x6B, 0.0, 0x37}, // dark mode red at start
		{0x37, 0x6B, 1.0, 0x6B}, // dark mode red at end
		{0x9C, 0x4B, 0.0, 0x9C}, // light mode red at start
		{0x9C, 0x4B, 1.0, 0x4B}, // light mode red at end
	}
	for _, tt := range tests {
		got := lerp(tt.a, tt.b, tt.t)
		if got != tt.want {
			t.Errorf("lerp(%d, %d, %.1f) = %d, want %d", tt.a, tt.b, tt.t, got, tt.want)
		}
	}
}
