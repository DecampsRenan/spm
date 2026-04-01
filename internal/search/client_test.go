package search

import (
	"encoding/json"
	"testing"
)

func TestFormatCount(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1_000, "1.0k"},
		{1_500, "1.5k"},
		{9_999, "10.0k"},
		{10_000, "10k"},
		{47_800, "48k"},
		{999_999, "1000k"},
		{1_000_000, "1.0M"},
		{1_500_000, "1.5M"},
		{10_000_000, "10M"},
		{47_800_000, "48M"},
	}

	for _, tt := range tests {
		got := FormatCount(tt.input)
		if got != tt.expected {
			t.Errorf("FormatCount(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "plain string with git+ prefix",
			raw:      `"git+https://github.com/user/repo.git"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "object with url field",
			raw:      `{"type": "git", "url": "git+https://github.com/user/repo.git"}`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "git protocol",
			raw:      `"git://github.com/user/repo.git"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "ssh protocol",
			raw:      `"ssh://git@github.com/user/repo.git"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "plain https URL",
			raw:      `"https://github.com/user/repo"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "empty",
			raw:      ``,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRepoURL(json.RawMessage(tt.raw))
			if got != tt.expected {
				t.Errorf("parseRepoURL(%s) = %q, want %q", tt.raw, got, tt.expected)
			}
		})
	}
}

func TestExtractGitHubRepo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/facebook/react", "facebook/react"},
		{"https://github.com/user/repo/tree/main", "user/repo"},
		{"https://gitlab.com/user/repo", ""},
		{"", ""},
		{"https://github.com//repo", ""},
		{"https://github.com/user/", ""},
	}

	for _, tt := range tests {
		got := extractGitHubRepo(tt.input)
		if got != tt.expected {
			t.Errorf("extractGitHubRepo(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCleanRepoURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git+https://github.com/user/repo.git", "https://github.com/user/repo"},
		{"git://github.com/user/repo.git", "https://github.com/user/repo"},
		{"ssh://git@github.com/user/repo.git", "https://github.com/user/repo"},
		{"https://github.com/user/repo", "https://github.com/user/repo"},
	}

	for _, tt := range tests {
		got := cleanRepoURL(tt.input)
		if got != tt.expected {
			t.Errorf("cleanRepoURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseRawString(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		objectKey string
		expected  string
	}{
		{"plain string", `"MIT"`, "type", "MIT"},
		{"object with key", `{"type": "ISC"}`, "type", "ISC"},
		{"object missing key", `{"foo": "bar"}`, "type", ""},
		{"empty", ``, "type", ""},
		{"author string", `"John Doe"`, "name", "John Doe"},
		{"author object", `{"name": "John Doe", "email": "john@example.com"}`, "name", "John Doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRawString(json.RawMessage(tt.raw), tt.objectKey)
			if got != tt.expected {
				t.Errorf("parseRawString(%s, %q) = %q, want %q", tt.raw, tt.objectKey, got, tt.expected)
			}
		})
	}
}
