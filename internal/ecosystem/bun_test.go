package ecosystem

import (
	"reflect"
	"testing"

	"github.com/decampsrenan/spm/internal/audit"
)

func TestBunResolve(t *testing.T) {
	eco := &bunEcosystem{}
	tests := []struct {
		name string
		cmd  string
		args []string
		want []string
	}{
		{"init", "init", nil, []string{"bun", "init"}},
		{"install", "install", nil, []string{"bun", "install"}},
		{"install-shorthand", "i", nil, []string{"bun", "install"}},
		{"add", "add", []string{"foo"}, []string{"bun", "add", "foo"}},
		{"remove", "remove", []string{"foo"}, []string{"bun", "remove", "foo"}},
		{"arbitrary", "dev", nil, []string{"bun", "dev"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := eco.Resolve(tt.cmd, tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBunBuildAuditCommand(t *testing.T) {
	eco := &bunEcosystem{}
	_, err := eco.BuildAuditCommand("", audit.Options{})
	if err == nil {
		t.Error("expected error for bun audit, got nil")
	}
}

func TestBunParseAuditOutput(t *testing.T) {
	eco := &bunEcosystem{}
	_, err := eco.ParseAuditOutput("", []byte("{}"))
	if err == nil {
		t.Error("expected error for bun parse, got nil")
	}
}

func TestBunMetadata(t *testing.T) {
	eco := &bunEcosystem{}
	if eco.Name() != Bun {
		t.Errorf("Name() = %q, want %q", eco.Name(), Bun)
	}
	if eco.ManifestFile() != "package.json" {
		t.Errorf("ManifestFile() = %q, want %q", eco.ManifestFile(), "package.json")
	}
	if !reflect.DeepEqual(eco.LockFiles(), []string{"bun.lock", "bun.lockb"}) {
		t.Errorf("LockFiles() = %v", eco.LockFiles())
	}
	if !reflect.DeepEqual(eco.ArtifactDirs(), []string{"node_modules"}) {
		t.Errorf("ArtifactDirs() = %v", eco.ArtifactDirs())
	}
	if !eco.HasCommand("anything") {
		t.Error("HasCommand() should return true")
	}
}
