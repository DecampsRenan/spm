package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/semver"
)

// Options configures the upgrade behavior.
type Options struct {
	CurrentVersion string // e.g. "0.4.0" or "dev"
	Alpha          bool   // include pre-releases
	Force          bool   // reinstall even if up to date
}

// Result holds the outcome of an upgrade check.
type Result struct {
	CurrentVersion string
	LatestVersion  string // tag like "v0.5.0"
	DownloadURL    string
	TargetPath     string // path to current executable
	AlreadyLatest  bool
}

// Release represents a GitHub release.
type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseFetcher abstracts GitHub API access for testing.
type ReleaseFetcher interface {
	FetchReleases() ([]Release, error)
}

// Downloader abstracts HTTP downloads of release archives for testing.
type Downloader interface {
	Download(url string) (io.ReadCloser, error)
}

// HTTPDownloader downloads archives via http.Get.
type HTTPDownloader struct {
	Client *http.Client
}

func (d *HTTPDownloader) Download(url string) (io.ReadCloser, error) {
	client := d.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading release: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

// GitHubFetcher fetches releases from the GitHub API.
type GitHubFetcher struct {
	Client *http.Client
	Repo   string // e.g. "decampsrenan/spm"
}

func (f *GitHubFetcher) FetchReleases() ([]Release, error) {
	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", f.Repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decoding releases: %w", err)
	}
	return releases, nil
}

// Plan checks for available updates and returns what would happen.
func Plan(fetcher ReleaseFetcher, opts Options) (*Result, error) {
	if opts.CurrentVersion == "dev" {
		return nil, fmt.Errorf("cannot upgrade a dev build — install a release version from https://github.com/decampsrenan/spm#getting-started")
	}

	currentTag := "v" + opts.CurrentVersion
	if !semver.IsValid(currentTag) {
		return nil, fmt.Errorf("invalid current version %q", opts.CurrentVersion)
	}

	releases, err := fetcher.FetchReleases()
	if err != nil {
		return nil, err
	}

	var latest *Release
	for i := range releases {
		r := &releases[i]
		if !opts.Alpha && r.Prerelease {
			continue
		}
		if !semver.IsValid(r.TagName) {
			continue
		}
		if latest == nil || semver.Compare(r.TagName, latest.TagName) > 0 {
			latest = r
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("no releases found")
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locating executable: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("resolving executable path: %w", err)
	}

	result := &Result{
		CurrentVersion: currentTag,
		LatestVersion:  latest.TagName,
		TargetPath:     execPath,
	}

	if semver.Compare(currentTag, latest.TagName) >= 0 && !opts.Force {
		result.AlreadyLatest = true
		return result, nil
	}

	archiveName := expectedArchiveName(latest.TagName)
	for _, a := range latest.Assets {
		if a.Name == archiveName {
			result.DownloadURL = a.BrowserDownloadURL
			return result, nil
		}
	}

	return nil, fmt.Errorf("no matching archive %q found in release %s", archiveName, latest.TagName)
}

// Execute downloads and replaces the current binary.
func Execute(downloader Downloader, result *Result) error {
	if isHomebrew(result.TargetPath) {
		return fmt.Errorf("spm appears to be installed via Homebrew — use `brew upgrade spm` instead")
	}

	body, err := downloader.Download(result.DownloadURL)
	if err != nil {
		return err
	}
	defer body.Close()

	binary, err := extractBinary(body)
	if err != nil {
		return err
	}

	// Get permissions from original binary.
	info, err := os.Stat(result.TargetPath)
	if err != nil {
		return fmt.Errorf("reading target permissions: %w", err)
	}

	// Write to temp file in the same directory for atomic rename.
	dir := filepath.Dir(result.TargetPath)
	tmp, err := os.CreateTemp(dir, "spm-upgrade-*")
	if err != nil {
		return fmt.Errorf("creating temp file (permission denied? try sudo): %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Chmod(info.Mode()); err != nil {
		tmp.Close()
		return fmt.Errorf("setting permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, result.TargetPath); err != nil {
		return fmt.Errorf("replacing binary (permission denied? try sudo): %w", err)
	}

	return nil
}

// expectedArchiveName returns the archive name for the current OS/arch.
func expectedArchiveName(tag string) string {
	version := strings.TrimPrefix(tag, "v")
	return fmt.Sprintf("spm_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
}

// extractBinary reads a tar.gz stream and returns the spm binary content.
func extractBinary(r io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("decompressing archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading archive: %w", err)
		}
		if hdr.Name == "spm" || filepath.Base(hdr.Name) == "spm" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("extracting binary: %w", err)
			}
			return data, nil
		}
	}
	return nil, fmt.Errorf("spm binary not found in archive")
}

// isHomebrew checks if the binary path is inside a Homebrew prefix.
func isHomebrew(path string) bool {
	return strings.Contains(path, "/Cellar/") || strings.Contains(path, "/Caskroom/")
}

// ExpectedArchiveName is exported for testing.
func ExpectedArchiveName(tag string) string {
	return expectedArchiveName(tag)
}
