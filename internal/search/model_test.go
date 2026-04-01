package search

import (
	"testing"
	"time"
)

func TestFilterVersions(t *testing.T) {
	versions := []VersionInfo{
		{Version: "18.3.1", PublishedAt: time.Now()},
		{Version: "18.3.0", PublishedAt: time.Now()},
		{Version: "18.2.0", PublishedAt: time.Now()},
		{Version: "17.0.2", PublishedAt: time.Now()},
		{Version: "0.14.0-beta.1", PublishedAt: time.Now()},
	}

	tests := []struct {
		name     string
		filter   string
		expected int
	}{
		{"empty filter returns all", "", 5},
		{"filter by major version", "18", 3},
		{"filter by exact version", "18.3.1", 1},
		{"filter by beta", "beta", 1},
		{"no match", "99.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterVersions(versions, tt.filter)
			if len(got) != tt.expected {
				t.Errorf("filterVersions(%q) returned %d results, want %d", tt.filter, len(got), tt.expected)
			}
		})
	}
}

func TestScrollWindow(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		total      int
		windowSize int
		wantStart  int
		wantEnd    int
	}{
		{"total less than window", 0, 5, 10, 0, 5},
		{"cursor at start", 0, 20, 10, 0, 10},
		{"cursor at end", 19, 20, 10, 10, 20},
		{"cursor in middle", 10, 20, 10, 5, 15},
		{"cursor near start", 2, 20, 10, 0, 10},
		{"cursor near end", 18, 20, 10, 10, 20},
		{"single item", 0, 1, 10, 0, 1},
		{"exact fit", 0, 10, 10, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := scrollWindow(tt.cursor, tt.total, tt.windowSize)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("scrollWindow(%d, %d, %d) = (%d, %d), want (%d, %d)",
					tt.cursor, tt.total, tt.windowSize, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"abc", 6, "abc   "},
		{"abc", 3, "abc"},
		{"abc", 2, "abc"},
		{"", 3, "   "},
	}

	for _, tt := range tests {
		got := padRight(tt.input, tt.width)
		if got != tt.expected {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expected)
		}
	}
}
