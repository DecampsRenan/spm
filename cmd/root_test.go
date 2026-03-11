package cmd

import (
	"testing"
)

func TestFirstNonFlagArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"script only", []string{"dev"}, "dev"},
		{"flag then script", []string{"--dry-run", "dev"}, "dev"},
		{"script then flag", []string{"dev", "--dry-run"}, "dev"},
		{"multiple flags then script", []string{"--dry-run", "--verbose", "test"}, "test"},
		{"known command", []string{"install"}, "install"},
		{"flag then known command", []string{"--dry-run", "install"}, "install"},
		{"only flags", []string{"--dry-run", "--help"}, ""},
		{"empty args", []string{}, ""},
		{"short flag then script", []string{"-v", "dev"}, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstNonFlagArg(tt.args)
			if got != tt.want {
				t.Errorf("firstNonFlagArg(%v) = %q, want %q", tt.args, got, tt.want)
			}
		})
	}
}
