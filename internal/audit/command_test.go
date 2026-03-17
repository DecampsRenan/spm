package audit

import (
	"reflect"
	"testing"
)

func TestBuildNPM(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{"default", Options{}, []string{"npm", "audit", "--json"}},
		{"prod-only", Options{ProdOnly: true}, []string{"npm", "audit", "--json", "--omit=dev"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNPM(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildYarnClassic(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{"default", Options{}, []string{"yarn", "audit", "--json"}},
		{"prod-only", Options{ProdOnly: true}, []string{"yarn", "audit", "--json", "--groups", "dependencies"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildYarnClassic(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildYarnBerry(t *testing.T) {
	got := buildYarnBerry(Options{})
	want := []string{"yarn", "npm", "audit", "--all", "--json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBuildPnpm(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want []string
	}{
		{"default", Options{}, []string{"pnpm", "audit", "--json"}},
		{"prod-only", Options{ProdOnly: true}, []string{"pnpm", "audit", "--json", "--prod"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPnpm(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
