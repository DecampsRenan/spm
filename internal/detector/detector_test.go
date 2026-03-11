package detector

import (
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
}

func TestDetectNoPackageJSON(t *testing.T) {
	dir := t.TempDir()

	_, err := Detect(dir)
	if err == nil {
		t.Fatal("expected error when no package.json found")
	}
}

func touch(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
		t.Fatal(err)
	}
}
