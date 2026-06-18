package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// mockDownloader implements Downloader for testing.
type mockDownloader struct {
	body []byte
	err  error
}

func (m *mockDownloader) Download(url string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.body)), nil
}

// mockFetcher implements ReleaseFetcher for testing.
type mockFetcher struct {
	releases []Release
	err      error
}

func (m *mockFetcher) FetchReleases() ([]Release, error) {
	return m.releases, m.err
}

func stableRelease(tag string) Release {
	version := tag[1:] // strip "v"
	return Release{
		TagName:    tag,
		Prerelease: false,
		Assets: []Asset{
			{
				Name:               fmt.Sprintf("spm_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH),
				BrowserDownloadURL: fmt.Sprintf("https://example.com/spm_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH),
			},
		},
	}
}

func alphaRelease(tag string) Release {
	version := tag[1:]
	r := stableRelease(tag)
	r.Prerelease = true
	r.Assets = []Asset{
		{
			Name:               fmt.Sprintf("spm_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH),
			BrowserDownloadURL: fmt.Sprintf("https://example.com/spm_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH),
		},
	}
	return r
}

func TestPlan_DevBuildRefused(t *testing.T) {
	_, err := Plan(&mockFetcher{}, Options{CurrentVersion: "dev"})
	if err == nil {
		t.Fatal("expected error for dev build")
	}
	if got := err.Error(); got == "" || !contains(got, "cannot upgrade a dev build") {
		t.Errorf("unexpected error message: %s", got)
	}
}

func TestPlan_InvalidVersion(t *testing.T) {
	_, err := Plan(&mockFetcher{}, Options{CurrentVersion: "not-semver"})
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestPlan_AlreadyLatest(t *testing.T) {
	fetcher := &mockFetcher{releases: []Release{stableRelease("v0.4.0")}}
	result, err := Plan(fetcher, Options{CurrentVersion: "0.4.0"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.AlreadyLatest {
		t.Error("expected AlreadyLatest to be true")
	}
}

func TestPlan_AlreadyLatestWithForce(t *testing.T) {
	fetcher := &mockFetcher{releases: []Release{stableRelease("v0.4.0")}}
	result, err := Plan(fetcher, Options{CurrentVersion: "0.4.0", Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.AlreadyLatest {
		t.Error("expected AlreadyLatest to be false with Force")
	}
	if result.DownloadURL == "" {
		t.Error("expected DownloadURL to be set")
	}
}

func TestPlan_NewVersionAvailable(t *testing.T) {
	fetcher := &mockFetcher{releases: []Release{stableRelease("v0.5.0")}}
	result, err := Plan(fetcher, Options{CurrentVersion: "0.4.0"})
	if err != nil {
		t.Fatal(err)
	}
	if result.AlreadyLatest {
		t.Error("expected AlreadyLatest to be false")
	}
	if result.LatestVersion != "v0.5.0" {
		t.Errorf("expected LatestVersion v0.5.0, got %s", result.LatestVersion)
	}
	if result.DownloadURL == "" {
		t.Error("expected DownloadURL to be set")
	}
}

func TestPlan_AlphaFiltering(t *testing.T) {
	fetcher := &mockFetcher{releases: []Release{
		alphaRelease("v0.6.0-alpha.1"),
		stableRelease("v0.5.0"),
	}}

	// Without alpha: should pick stable
	result, err := Plan(fetcher, Options{CurrentVersion: "0.4.0"})
	if err != nil {
		t.Fatal(err)
	}
	if result.LatestVersion != "v0.5.0" {
		t.Errorf("expected v0.5.0 without alpha, got %s", result.LatestVersion)
	}

	// With alpha: should pick alpha
	result, err = Plan(fetcher, Options{CurrentVersion: "0.4.0", Alpha: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.LatestVersion != "v0.6.0-alpha.1" {
		t.Errorf("expected v0.6.0-alpha.1 with alpha, got %s", result.LatestVersion)
	}
}

func TestPlan_NoReleases(t *testing.T) {
	fetcher := &mockFetcher{releases: []Release{}}
	_, err := Plan(fetcher, Options{CurrentVersion: "0.4.0"})
	if err == nil {
		t.Fatal("expected error for no releases")
	}
}

func TestPlan_FetchError(t *testing.T) {
	fetcher := &mockFetcher{err: fmt.Errorf("network error")}
	_, err := Plan(fetcher, Options{CurrentVersion: "0.4.0"})
	if err == nil {
		t.Fatal("expected error for fetch failure")
	}
}

func TestPlan_PicksHighestVersion(t *testing.T) {
	fetcher := &mockFetcher{releases: []Release{
		stableRelease("v0.3.0"),
		stableRelease("v0.5.0"),
		stableRelease("v0.4.0"),
	}}
	result, err := Plan(fetcher, Options{CurrentVersion: "0.2.0"})
	if err != nil {
		t.Fatal(err)
	}
	if result.LatestVersion != "v0.5.0" {
		t.Errorf("expected v0.5.0, got %s", result.LatestVersion)
	}
}

func TestExpectedArchiveName(t *testing.T) {
	name := ExpectedArchiveName("v0.5.0")
	expected := fmt.Sprintf("spm_0.5.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	if name != expected {
		t.Errorf("expected %s, got %s", expected, name)
	}
}

func TestIsHomebrew(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/usr/local/Cellar/spm/0.4.0/bin/spm", true},
		{"/opt/homebrew/Caskroom/spm/0.4.0/spm", true},
		{"/usr/local/bin/spm", false},
		{"/home/user/.local/bin/spm", false},
	}
	for _, tt := range tests {
		if got := isHomebrew(tt.path); got != tt.want {
			t.Errorf("isHomebrew(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestExtractBinary(t *testing.T) {
	content := []byte("fake-spm-binary")
	archive := createTarGz(t, "spm", content)

	got, err := extractBinary(bytes.NewReader(archive))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("extracted content mismatch")
	}
}

func TestExtractBinary_NotFound(t *testing.T) {
	archive := createTarGz(t, "other-file", []byte("data"))
	_, err := extractBinary(bytes.NewReader(archive))
	if err == nil {
		t.Fatal("expected error when binary not found")
	}
}

func TestExecute_HomebrewRefused(t *testing.T) {
	result := &Result{
		TargetPath:  "/opt/homebrew/Cellar/spm/0.4.0/bin/spm",
		DownloadURL: "https://example.com/spm.tar.gz",
	}
	err := Execute(&mockDownloader{}, result)
	if err == nil {
		t.Fatal("expected error for Homebrew install")
	}
	if !contains(err.Error(), "Homebrew") {
		t.Errorf("expected Homebrew error, got: %v", err)
	}
}

func TestExecute_DownloadError(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "spm")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	result := &Result{TargetPath: target, DownloadURL: "https://example.com/spm.tar.gz"}

	err := Execute(&mockDownloader{err: fmt.Errorf("connection refused")}, result)
	if err == nil {
		t.Fatal("expected error for download failure")
	}
	if !contains(err.Error(), "connection refused") {
		t.Errorf("expected download error to propagate, got: %v", err)
	}
}

func TestExecute_InvalidArchive(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "spm")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	result := &Result{TargetPath: target, DownloadURL: "https://example.com/spm.tar.gz"}

	err := Execute(&mockDownloader{body: []byte("not a tar.gz")}, result)
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestExecute_BinaryNotInArchive(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "spm")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	archive := createTarGz(t, "other-file", []byte("data"))
	result := &Result{TargetPath: target, DownloadURL: "https://example.com/spm.tar.gz"}

	err := Execute(&mockDownloader{body: archive}, result)
	if err == nil {
		t.Fatal("expected error when binary not in archive")
	}
}

func TestExecute_SuccessfullyReplacesBinary(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "spm")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	newContent := []byte("new-spm-binary")
	archive := createTarGz(t, "spm", newContent)
	result := &Result{TargetPath: target, DownloadURL: "https://example.com/spm.tar.gz"}

	if err := Execute(&mockDownloader{body: archive}, result); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newContent) {
		t.Errorf("binary not replaced: got %q, want %q", got, newContent)
	}
}

func TestExecute_MissingTarget(t *testing.T) {
	archive := createTarGz(t, "spm", []byte("new"))
	result := &Result{
		TargetPath:  "/nonexistent/path/spm",
		DownloadURL: "https://example.com/spm.tar.gz",
	}
	err := Execute(&mockDownloader{body: archive}, result)
	if err == nil {
		t.Fatal("expected error when target doesn't exist")
	}
}

func TestHTTPDownloader_Success(t *testing.T) {
	want := []byte("archive-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(want)
	}))
	defer srv.Close()

	d := &HTTPDownloader{}
	body, err := d.Download(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer body.Close()
	got, _ := io.ReadAll(body)
	if !bytes.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHTTPDownloader_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	d := &HTTPDownloader{}
	_, err := d.Download(srv.URL)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !contains(err.Error(), "404") {
		t.Errorf("expected status in error, got: %v", err)
	}
}

func TestHTTPDownloader_NetworkError(t *testing.T) {
	d := &HTTPDownloader{}
	// Use an invalid URL to trigger a transport error.
	_, err := d.Download("http://127.0.0.1:1/definitely-not-listening")
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestExecute_ReplacesFile(t *testing.T) {
	dir := t.TempDir()

	// Create a fake "current" binary.
	target := filepath.Join(dir, "spm")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Serve a fake archive via a temp file + file:// would be complex,
	// so we test extractBinary and the file replacement logic separately.
	// The integration of HTTP download is covered by the Plan tests with mock fetcher.

	newContent := []byte("new-spm-binary")

	// Simulate what Execute does after download:
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}

	tmp, err := os.CreateTemp(dir, "spm-upgrade-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.Write(newContent); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Chmod(info.Mode()); err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	if err := os.Rename(tmp.Name(), target); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newContent) {
		t.Errorf("binary was not replaced correctly")
	}

	gotInfo, _ := os.Stat(target)
	if gotInfo.Mode() != info.Mode() {
		t.Errorf("permissions changed: %v → %v", info.Mode(), gotInfo.Mode())
	}
}

// createTarGz creates a tar.gz archive with a single file.
func createTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
