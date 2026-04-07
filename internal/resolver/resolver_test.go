package resolver

import (
	"reflect"
	"testing"

	"github.com/decampsrenan/spm/internal/ecosystem"
)

func TestResolveInstall(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		want []string
	}{
		{ecosystem.NPM, []string{"npm", "install"}},
		{ecosystem.Yarn, []string{"yarn", "install"}},
		{ecosystem.Pnpm, []string{"pnpm", "install"}},
		{ecosystem.Bun, []string{"bun", "install"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "install", nil)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, install) = %v, want %v", tt.pm, got, tt.want)
		}
	}
}

func TestResolveAdd(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		args []string
		want []string
	}{
		{ecosystem.NPM, []string{"react"}, []string{"npm", "install", "react"}},
		{ecosystem.Yarn, []string{"react"}, []string{"yarn", "add", "react"}},
		{ecosystem.Pnpm, []string{"react"}, []string{"pnpm", "add", "react"}},
		{ecosystem.Bun, []string{"react"}, []string{"bun", "add", "react"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "add", tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, add, %v) = %v, want %v", tt.pm, tt.args, got, tt.want)
		}
	}
}

func TestResolveRemove(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		args []string
		want []string
	}{
		{ecosystem.NPM, []string{"react"}, []string{"npm", "uninstall", "react"}},
		{ecosystem.Yarn, []string{"react"}, []string{"yarn", "remove", "react"}},
		{ecosystem.Pnpm, []string{"react"}, []string{"pnpm", "remove", "react"}},
		{ecosystem.Bun, []string{"react"}, []string{"bun", "remove", "react"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "remove", tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, remove, %v) = %v, want %v", tt.pm, tt.args, got, tt.want)
		}
	}
}

func TestResolveRemoveWithExtraFlags(t *testing.T) {
	got := Resolve(ecosystem.NPM, "remove", []string{"react", "--save-dev"})
	want := []string{"npm", "uninstall", "react", "--save-dev"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResolveFallbackScript(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		cmd  string
		want []string
	}{
		{ecosystem.NPM, "dev", []string{"npm", "run", "dev"}},
		{ecosystem.Yarn, "dev", []string{"yarn", "dev"}},
		{ecosystem.Pnpm, "dev", []string{"pnpm", "dev"}},
		{ecosystem.Bun, "dev", []string{"bun", "dev"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, tt.cmd, nil)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, %s) = %v, want %v", tt.pm, tt.cmd, got, tt.want)
		}
	}
}

func TestResolveInit(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		args []string
		want []string
	}{
		{ecosystem.NPM, nil, []string{"npm", "init", "-y"}},
		{ecosystem.Yarn, nil, []string{"yarn", "init", "-y"}},
		{ecosystem.Pnpm, nil, []string{"pnpm", "init"}},
		{ecosystem.Bun, nil, []string{"bun", "init"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "init", tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, init, %v) = %v, want %v", tt.pm, tt.args, got, tt.want)
		}
	}
}

func TestResolveInitWithExtraFlags(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		args []string
		want []string
	}{
		{ecosystem.NPM, []string{"--scope=@myorg"}, []string{"npm", "init", "-y", "--scope=@myorg"}},
		{ecosystem.Bun, []string{"--react"}, []string{"bun", "init", "--react"}},
		{ecosystem.Pnpm, []string{"--react"}, []string{"pnpm", "init", "--react"}},
		{ecosystem.Yarn, []string{"--scope=@myorg"}, []string{"yarn", "init", "-y", "--scope=@myorg"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "init", tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, init, %v) = %v, want %v", tt.pm, tt.args, got, tt.want)
		}
	}
}

func TestResolveInitYarnClassicNeedsNonInteractiveFlag(t *testing.T) {
	got := Resolve(ecosystem.Yarn, "init", nil)
	want := []string{"yarn", "init", "-y"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Yarn init should include -y for Classic compatibility: got %v, want %v", got, want)
	}
}

func TestResolveInitNonInteractivePMsOmitFlag(t *testing.T) {
	tests := []struct {
		pm   ecosystem.PackageManager
		want []string
	}{
		{ecosystem.Pnpm, []string{"pnpm", "init"}},
		{ecosystem.Bun, []string{"bun", "init"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "init", nil)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, init) should not include -y: got %v, want %v", tt.pm, got, tt.want)
		}
		for _, arg := range got {
			if arg == "-y" {
				t.Errorf("Resolve(%s, init) must not include -y, but got %v", tt.pm, got)
			}
		}
	}
}

func TestResolveInitAllPMsPassthroughExtraArgs(t *testing.T) {
	tests := []struct {
		pm       ecosystem.PackageManager
		args     []string
		wantTail []string
	}{
		{ecosystem.NPM, []string{"--scope=@myorg", "--yes"}, []string{"--scope=@myorg", "--yes"}},
		{ecosystem.Yarn, []string{"--scope=@myorg", "--private"}, []string{"--scope=@myorg", "--private"}},
		{ecosystem.Pnpm, []string{"--react", "--typescript"}, []string{"--react", "--typescript"}},
		{ecosystem.Bun, []string{"--react", "--open"}, []string{"--react", "--open"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "init", tt.args)
		tail := got[len(got)-len(tt.wantTail):]
		if !reflect.DeepEqual(tail, tt.wantTail) {
			t.Errorf("Resolve(%s, init, %v): extra args not forwarded correctly, got tail %v, want %v", tt.pm, tt.args, tail, tt.wantTail)
		}
	}
}

func TestResolveWithExtraFlags(t *testing.T) {
	got := Resolve(ecosystem.NPM, "add", []string{"react", "--save-dev"})
	want := []string{"npm", "install", "react", "--save-dev"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
