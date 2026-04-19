package ecosystem

import "testing"

func TestForPMKnown(t *testing.T) {
	tests := []struct {
		pm   PackageManager
		want PackageManager
	}{
		{NPM, NPM},
		{Yarn, Yarn},
		{Pnpm, Pnpm},
		{Bun, Bun},
	}
	for _, tt := range tests {
		t.Run(string(tt.pm), func(t *testing.T) {
			eco := ForPM(tt.pm)
			if eco == nil {
				t.Fatalf("ForPM(%q) = nil, want non-nil", tt.pm)
			}
			if eco.Name() != tt.want {
				t.Errorf("ForPM(%q).Name() = %q, want %q", tt.pm, eco.Name(), tt.want)
			}
		})
	}
}

func TestForPMUnknown(t *testing.T) {
	if eco := ForPM(PackageManager("cargo")); eco != nil {
		t.Errorf("ForPM(unknown) = %v, want nil", eco)
	}
	if eco := ForPM(PackageManager("")); eco != nil {
		t.Errorf("ForPM(empty) = %v, want nil", eco)
	}
}

func TestAllRegistersEveryEcosystem(t *testing.T) {
	got := All()
	if len(got) != 4 {
		t.Fatalf("All() length = %d, want 4", len(got))
	}

	expected := map[PackageManager]bool{NPM: false, Yarn: false, Pnpm: false, Bun: false}
	for _, eco := range got {
		name := eco.Name()
		seen, known := expected[name]
		if !known {
			t.Errorf("All() contains unexpected ecosystem %q", name)
			continue
		}
		if seen {
			t.Errorf("All() contains duplicate ecosystem %q", name)
		}
		expected[name] = true
	}
	for pm, seen := range expected {
		if !seen {
			t.Errorf("All() missing ecosystem %q", pm)
		}
	}
}

func TestAllEcosystemsExposeManifestAndLocks(t *testing.T) {
	for _, eco := range All() {
		t.Run(string(eco.Name()), func(t *testing.T) {
			if eco.ManifestFile() == "" {
				t.Error("ManifestFile() returned empty string")
			}
			if len(eco.LockFiles()) == 0 {
				t.Error("LockFiles() returned empty slice")
			}
			if len(eco.ArtifactDirs()) == 0 {
				t.Error("ArtifactDirs() returned empty slice")
			}
		})
	}
}
