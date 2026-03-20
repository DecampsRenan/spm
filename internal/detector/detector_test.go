package detector

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectNPM(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "package-lock.json")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != NPM {
		t.Fatalf("expected npm, got %v", dets)
	}
}

func TestDetectYarn(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "yarn.lock")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Yarn {
		t.Fatalf("expected yarn, got %v", dets)
	}
}

func TestDetectPnpm(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "pnpm-lock.yaml")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Pnpm {
		t.Fatalf("expected pnpm, got %v", dets)
	}
}

func TestDetectBun(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "bun.lock")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Bun {
		t.Fatalf("expected bun, got %v", dets)
	}
}

func TestDetectBunLegacy(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "bun.lockb")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Bun {
		t.Fatalf("expected bun, got %v", dets)
	}
}

func TestDetectBunBothLockFiles(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "bun.lock")
	touch(t, dir, "bun.lockb")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Bun {
		t.Fatalf("expected single bun detection, got %v", dets)
	}
}

func TestDetectWalksUp(t *testing.T) {
	root := t.TempDir()
	touch(t, root, "package.json")
	touch(t, root, "yarn.lock")

	sub := filepath.Join(root, "src", "components")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	dets, err := Detect(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Yarn {
		t.Fatalf("expected yarn from parent, got %v", dets)
	}
	if dets[0].Dir != root {
		t.Fatalf("expected dir %s, got %s", root, dets[0].Dir)
	}
}

func TestDetectMultipleLockFiles(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")
	touch(t, dir, "package-lock.json")
	touch(t, dir, "yarn.lock")

	dets, err := Detect(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 2 {
		t.Fatalf("expected 2 detections, got %d: %v", len(dets), dets)
	}
}

func TestDetectNoLockFile(t *testing.T) {
	dir := t.TempDir()
	touch(t, dir, "package.json")

	_, err := Detect(dir)
	if err == nil {
		t.Fatal("expected error when no lock file found")
	}

	var noLock *ErrNoLockFile
	if !errors.As(err, &noLock) {
		t.Fatalf("expected ErrNoLockFile, got %T: %v", err, err)
	}
	if noLock.Dir != dir {
		t.Fatalf("expected dir %s, got %s", dir, noLock.Dir)
	}
}

func TestDetectWalksUpPastPackageJSONWithoutLockFile(t *testing.T) {
	root := t.TempDir()
	touch(t, root, "package.json")
	touch(t, root, "yarn.lock")

	nested := filepath.Join(root, "packages", "my-lib")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	touch(t, nested, "package.json") // no lock file here

	dets, err := Detect(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 1 || dets[0].PM != Yarn {
		t.Fatalf("expected yarn from root, got %v", dets)
	}
	if dets[0].Dir != root {
		t.Fatalf("expected dir %s, got %s", root, dets[0].Dir)
	}
}

func TestDetectNoPackageJSON(t *testing.T) {
	dir := t.TempDir()

	_, err := Detect(dir)
	if err == nil {
		t.Fatal("expected error when no package.json found")
	}
}

func TestLockFileName(t *testing.T) {
	tests := []struct {
		pm   PackageManager
		want string
	}{
		{NPM, "package-lock.json"},
		{Yarn, "yarn.lock"},
		{Pnpm, "pnpm-lock.yaml"},
		{Bun, "bun.lock"},
		{PackageManager("unknown"), ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.pm), func(t *testing.T) {
			got := LockFileName(tt.pm)
			if got != tt.want {
				t.Errorf("LockFileName(%s) = %q, want %q", tt.pm, got, tt.want)
			}
		})
	}
}

func touch(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
		t.Fatal(err)
	}
}
