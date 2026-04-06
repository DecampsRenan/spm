package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestVHSSnapshots renders a set of `spm` commands through charmbracelet/vhs
// and compares the resulting terminal text output against committed snapshots.
//
// This guarantees that the visual output of important commands (help screens,
// version, …) does not silently change.
//
// The test is skipped when `vhs` is not installed locally. To regenerate the
// snapshots after an intentional change, run:
//
//	UPDATE_VHS_SNAPSHOTS=1 go test ./cmd -run TestVHSSnapshots
func TestVHSSnapshots(t *testing.T) {
	if _, err := exec.LookPath("vhs"); err != nil {
		t.Skip("vhs not installed; skipping snapshot tests (install: https://github.com/charmbracelet/vhs)")
	}

	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}

	// Build a fresh `spm` binary into a temp dir and prepend it to PATH so
	// the tape files can simply invoke `spm`.
	binDir := t.TempDir()
	binName := "spm"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)
	build := exec.Command("go", "build", "-o", binPath, ".")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build spm: %v\n%s", err, out)
	}

	tapesDir := filepath.Join("testdata", "vhs")
	tapes, err := filepath.Glob(filepath.Join(tapesDir, "*.tape"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tapes) == 0 {
		t.Fatal("no .tape files found in testdata/vhs")
	}

	snapshotsDir := filepath.Join(tapesDir, "snapshots")
	if err := os.MkdirAll(snapshotsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	update := os.Getenv("UPDATE_VHS_SNAPSHOTS") != ""

	for _, tape := range tapes {
		tape := tape
		name := strings.TrimSuffix(filepath.Base(tape), ".tape")
		t.Run(name, func(t *testing.T) {
			workDir := t.TempDir()

			// Copy the tape into the work dir so its `Output <name>.txt`
			// directive lands in a known location.
			tapeBytes, err := os.ReadFile(tape)
			if err != nil {
				t.Fatal(err)
			}
			localTape := filepath.Join(workDir, filepath.Base(tape))
			if err := os.WriteFile(localTape, tapeBytes, 0o644); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command("vhs", filepath.Base(tape))
			cmd.Dir = workDir
			cmd.Env = append(os.Environ(),
				"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"NO_COLOR=1",
				"TERM=xterm-256color",
			)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("vhs failed for %s: %v\n%s", tape, err, out)
			}

			gotPath := filepath.Join(workDir, name+".txt")
			got, err := os.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("vhs did not produce expected output %s: %v", gotPath, err)
			}
			normalized := normalizeSnapshot(got)

			snapshotPath := filepath.Join(snapshotsDir, name+".txt")
			if update {
				if err := os.WriteFile(snapshotPath, normalized, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("updated snapshot %s", snapshotPath)
				return
			}

			want, err := os.ReadFile(snapshotPath)
			if err != nil {
				t.Fatalf("missing snapshot %s — run with UPDATE_VHS_SNAPSHOTS=1 to create it: %v", snapshotPath, err)
			}
			if string(normalized) != string(want) {
				t.Errorf("snapshot mismatch for %s\n--- want ---\n%s\n--- got ---\n%s\n(run UPDATE_VHS_SNAPSHOTS=1 to refresh)", name, want, normalized)
			}
		})
	}
}

// normalizeSnapshot strips trailing whitespace on every line and trailing
// blank lines so cosmetic terminal padding does not cause flakiness.
func normalizeSnapshot(b []byte) []byte {
	lines := strings.Split(string(b), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}
