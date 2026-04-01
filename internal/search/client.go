package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Result holds a single npm registry search result.
type Result struct {
	Name        string
	Description string
	Version     string
}

// VersionInfo holds a single npm package version with its publication date.
type VersionInfo struct {
	Version     string
	PublishedAt time.Time
}

// PackageDetails holds detailed npm package information for display.
type PackageDetails struct {
	Name            string
	Description     string
	Homepage        string
	Repository      string // cleaned https URL
	License         string
	Author          string
	Latest          string
	DistTags        map[string]string // tag → version (e.g. "latest", "next", "beta")
	Versions        []VersionInfo     // sorted by published date, newest first
	WeeklyDownloads int64
	Stars           int
}

type searchResponse struct {
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Version     string `json:"version"`
		} `json:"package"`
	} `json:"objects"`
}

type packageDetailsResponse struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Homepage    string            `json:"homepage"`
	License     json.RawMessage   `json:"license"`
	Author      json.RawMessage   `json:"author"`
	Repository  json.RawMessage   `json:"repository"`
	DistTags    map[string]string `json:"dist-tags"`
	Time        map[string]string `json:"time"`
}

type downloadsResponse struct {
	Downloads int64  `json:"downloads"`
	Package   string `json:"package"`
}

type githubRepoResponse struct {
	StargazersCount int `json:"stargazers_count"`
}

var httpClient = &http.Client{Timeout: 5 * time.Second}
var detailsClient = &http.Client{Timeout: 15 * time.Second}

// Query searches the npm registry for packages matching the given text.
func Query(ctx context.Context, text string, size int) ([]Result, error) {
	u := fmt.Sprintf(
		"https://registry.npmjs.org/-/v1/search?text=%s&size=%d",
		url.QueryEscape(text), size,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry returned %d", resp.StatusCode)
	}

	var sr searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}

	results := make([]Result, len(sr.Objects))
	for i, obj := range sr.Objects {
		results[i] = Result{
			Name:        obj.Package.Name,
			Description: obj.Package.Description,
			Version:     obj.Package.Version,
		}
	}

	return results, nil
}

// FetchDetails retrieves detailed package information from the npm registry,
// fetching the packument and weekly downloads concurrently, then stars.
func FetchDetails(ctx context.Context, name string) (*PackageDetails, error) {
	type packResult struct {
		data *packageDetailsResponse
		err  error
	}

	packCh := make(chan packResult, 1)
	dlCh := make(chan int64, 1)

	go func() {
		data, err := fetchPackument(ctx, name)
		packCh <- packResult{data, err}
	}()

	go func() {
		count, _ := fetchWeeklyDownloads(ctx, name)
		dlCh <- count
	}()

	pack := <-packCh
	downloads := <-dlCh

	if pack.err != nil {
		return nil, pack.err
	}

	raw := pack.data
	repoURL := parseRepoURL(raw.Repository)

	// Fetch GitHub stars with a short timeout (best-effort).
	var stars int
	if ghRepo := extractGitHubRepo(repoURL); ghRepo != "" {
		ghCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		stars, _ = fetchGitHubStars(ghCtx, ghRepo)
	}

	var versions []VersionInfo
	for version, timeStr := range raw.Time {
		if version == "created" || version == "modified" {
			continue
		}
		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			continue
		}
		versions = append(versions, VersionInfo{Version: version, PublishedAt: t})
	}
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].PublishedAt.After(versions[j].PublishedAt)
	})

	return &PackageDetails{
		Name:            raw.Name,
		Description:     raw.Description,
		Homepage:        raw.Homepage,
		Repository:      repoURL,
		License:         parseRawString(raw.License, "type"),
		Author:          parseRawString(raw.Author, "name"),
		Latest:          raw.DistTags["latest"],
		DistTags:        raw.DistTags,
		Versions:        versions,
		WeeklyDownloads: downloads,
		Stars:           stars,
	}, nil
}

func fetchPackument(ctx context.Context, name string) (*packageDetailsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://registry.npmjs.org/"+name, nil)
	if err != nil {
		return nil, err
	}
	resp, err := detailsClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm registry returned %d", resp.StatusCode)
	}
	var raw packageDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return &raw, nil
}

func fetchWeeklyDownloads(ctx context.Context, name string) (int64, error) {
	u := "https://api.npmjs.org/downloads/point/last-week/" + url.PathEscape(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("downloads API returned %d", resp.StatusCode)
	}
	var dl downloadsResponse
	if err := json.NewDecoder(resp.Body).Decode(&dl); err != nil {
		return 0, err
	}
	return dl.Downloads, nil
}

func fetchGitHubStars(ctx context.Context, repo string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repo, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github API returned %d", resp.StatusCode)
	}
	var gh githubRepoResponse
	if err := json.NewDecoder(resp.Body).Decode(&gh); err != nil {
		return 0, err
	}
	return gh.StargazersCount, nil
}

// parseRepoURL extracts a clean https URL from npm's repository field.
func parseRepoURL(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return cleanRepoURL(s)
	}
	var obj struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return cleanRepoURL(obj.URL)
	}
	return ""
}

func cleanRepoURL(u string) string {
	u = strings.TrimPrefix(u, "git+")
	u = strings.TrimSuffix(u, ".git")
	u = strings.NewReplacer(
		"git://github.com/", "https://github.com/",
		"ssh://git@github.com/", "https://github.com/",
	).Replace(u)
	return u
}

// extractGitHubRepo returns "owner/repo" from a GitHub URL, or "".
func extractGitHubRepo(repoURL string) string {
	const prefix = "github.com/"
	idx := strings.Index(repoURL, prefix)
	if idx < 0 {
		return ""
	}
	path := repoURL[idx+len(prefix):]
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "/" + parts[1]
}

// FormatCount formats a large integer compactly (e.g. 47_800_000 → "47.8M").
func FormatCount(n int64) string {
	switch {
	case n >= 1_000_000:
		f := float64(n) / 1_000_000
		if f >= 10 {
			return fmt.Sprintf("%.0fM", f)
		}
		return fmt.Sprintf("%.1fM", f)
	case n >= 1_000:
		f := float64(n) / 1_000
		if f >= 10 {
			return fmt.Sprintf("%.0fk", f)
		}
		return fmt.Sprintf("%.1fk", f)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// parseRawString tries to parse a JSON raw message as a plain string first,
// then as an object extracting the given key.
func parseRawString(raw json.RawMessage, objectKey string) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var obj map[string]string
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj[objectKey]
	}
	return ""
}
