package resolver

import (
	"reflect"
	"testing"

	"github.com/decampsrenan/spm/internal/detector"
)

func TestResolveInstall(t *testing.T) {
	tests := []struct {
		pm   detector.PackageManager
		want []string
	}{
		{detector.NPM, []string{"npm", "install"}},
		{detector.Yarn, []string{"yarn", "install"}},
		{detector.Pnpm, []string{"pnpm", "install"}},
		{detector.Bun, []string{"bun", "install"}},
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
		pm   detector.PackageManager
		args []string
		want []string
	}{
		{detector.NPM, []string{"react"}, []string{"npm", "install", "react"}},
		{detector.Yarn, []string{"react"}, []string{"yarn", "add", "react"}},
		{detector.Pnpm, []string{"react"}, []string{"pnpm", "add", "react"}},
		{detector.Bun, []string{"react"}, []string{"bun", "add", "react"}},
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
		pm   detector.PackageManager
		args []string
		want []string
	}{
		{detector.NPM, []string{"react"}, []string{"npm", "uninstall", "react"}},
		{detector.Yarn, []string{"react"}, []string{"yarn", "remove", "react"}},
		{detector.Pnpm, []string{"react"}, []string{"pnpm", "remove", "react"}},
		{detector.Bun, []string{"react"}, []string{"bun", "remove", "react"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "remove", tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, remove, %v) = %v, want %v", tt.pm, tt.args, got, tt.want)
		}
	}
}

func TestResolveRemoveWithExtraFlags(t *testing.T) {
	got := Resolve(detector.NPM, "remove", []string{"react", "--save-dev"})
	want := []string{"npm", "uninstall", "react", "--save-dev"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestResolveFallbackScript(t *testing.T) {
	tests := []struct {
		pm   detector.PackageManager
		cmd  string
		want []string
	}{
		{detector.NPM, "dev", []string{"npm", "run", "dev"}},
		{detector.Yarn, "dev", []string{"yarn", "dev"}},
		{detector.Pnpm, "dev", []string{"pnpm", "dev"}},
		{detector.Bun, "dev", []string{"bun", "dev"}},
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
		pm   detector.PackageManager
		args []string
		want []string
	}{
		{detector.NPM, nil, []string{"npm", "init", "-y"}},
		{detector.Yarn, nil, []string{"yarn", "init", "-y"}},
		{detector.Pnpm, nil, []string{"pnpm", "init"}},
		{detector.Bun, nil, []string{"bun", "init", "-y"}},
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
		pm   detector.PackageManager
		args []string
		want []string
	}{
		{detector.NPM, []string{"--scope=@myorg"}, []string{"npm", "init", "-y", "--scope=@myorg"}},
		{detector.Bun, []string{"--react"}, []string{"bun", "init", "-y", "--react"}},
		{detector.Pnpm, []string{"--react"}, []string{"pnpm", "init", "--react"}},
		{detector.Yarn, []string{"--scope=@myorg"}, []string{"yarn", "init", "-y", "--scope=@myorg"}},
	}
	for _, tt := range tests {
		got := Resolve(tt.pm, "init", tt.args)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Resolve(%s, init, %v) = %v, want %v", tt.pm, tt.args, got, tt.want)
		}
	}
}

func TestResolveWithExtraFlags(t *testing.T) {
	got := Resolve(detector.NPM, "add", []string{"react", "--save-dev"})
	want := []string{"npm", "install", "react", "--save-dev"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
