package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// routeTransport routes every request to a single test server, preserving the
// original host so the server handler can dispatch based on it.
type routeTransport struct {
	target *url.URL
}

func (r *routeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = r.target.Scheme
	req2.Host = req.URL.Host
	req2.URL.Host = r.target.Host
	return http.DefaultTransport.RoundTrip(req2)
}

// withTestServer swaps both package-level HTTP clients to point at srv for
// the duration of the test.
func withTestServer(t *testing.T, srv *httptest.Server) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	rt := &routeTransport{target: target}

	origHTTP := httpClient
	origDetails := detailsClient
	httpClient = &http.Client{Transport: rt, Timeout: 2 * time.Second}
	detailsClient = &http.Client{Transport: rt, Timeout: 2 * time.Second}
	t.Cleanup(func() {
		httpClient = origHTTP
		detailsClient = origDetails
	})
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1_000, "1.0k"},
		{1_500, "1.5k"},
		{9_999, "10.0k"},
		{10_000, "10k"},
		{47_800, "48k"},
		{999_999, "1000k"},
		{1_000_000, "1.0M"},
		{1_500_000, "1.5M"},
		{10_000_000, "10M"},
		{47_800_000, "48M"},
	}

	for _, tt := range tests {
		got := FormatCount(tt.input)
		if got != tt.expected {
			t.Errorf("FormatCount(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "plain string with git+ prefix",
			raw:      `"git+https://github.com/user/repo.git"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "object with url field",
			raw:      `{"type": "git", "url": "git+https://github.com/user/repo.git"}`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "git protocol",
			raw:      `"git://github.com/user/repo.git"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "ssh protocol",
			raw:      `"ssh://git@github.com/user/repo.git"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "plain https URL",
			raw:      `"https://github.com/user/repo"`,
			expected: "https://github.com/user/repo",
		},
		{
			name:     "empty",
			raw:      ``,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRepoURL(json.RawMessage(tt.raw))
			if got != tt.expected {
				t.Errorf("parseRepoURL(%s) = %q, want %q", tt.raw, got, tt.expected)
			}
		})
	}
}

func TestExtractGitHubRepo(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/facebook/react", "facebook/react"},
		{"https://github.com/user/repo/tree/main", "user/repo"},
		{"https://gitlab.com/user/repo", ""},
		{"", ""},
		{"https://github.com//repo", ""},
		{"https://github.com/user/", ""},
	}

	for _, tt := range tests {
		got := extractGitHubRepo(tt.input)
		if got != tt.expected {
			t.Errorf("extractGitHubRepo(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCleanRepoURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"git+https://github.com/user/repo.git", "https://github.com/user/repo"},
		{"git://github.com/user/repo.git", "https://github.com/user/repo"},
		{"ssh://git@github.com/user/repo.git", "https://github.com/user/repo"},
		{"https://github.com/user/repo", "https://github.com/user/repo"},
	}

	for _, tt := range tests {
		got := cleanRepoURL(tt.input)
		if got != tt.expected {
			t.Errorf("cleanRepoURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseRawString(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		objectKey string
		expected  string
	}{
		{"plain string", `"MIT"`, "type", "MIT"},
		{"object with key", `{"type": "ISC"}`, "type", "ISC"},
		{"object missing key", `{"foo": "bar"}`, "type", ""},
		{"empty", ``, "type", ""},
		{"author string", `"John Doe"`, "name", "John Doe"},
		{"author object", `{"name": "John Doe", "email": "john@example.com"}`, "name", "John Doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRawString(json.RawMessage(tt.raw), tt.objectKey)
			if got != tt.expected {
				t.Errorf("parseRawString(%s, %q) = %q, want %q", tt.raw, tt.objectKey, got, tt.expected)
			}
		})
	}
}

func TestQueryReturnsResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "registry.npmjs.org" {
			t.Errorf("unexpected host: %q", r.Host)
		}
		if !strings.HasPrefix(r.URL.Path, "/-/v1/search") {
			t.Errorf("unexpected path: %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("text"); got != "react" {
			t.Errorf("text=%q, want react", got)
		}
		if got := r.URL.Query().Get("size"); got != "5" {
			t.Errorf("size=%q, want 5", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"objects":[
			{"package":{"name":"react","description":"A JS library","version":"18.2.0"}},
			{"package":{"name":"react-dom","description":"DOM bindings","version":"18.2.0"}}
		]}`))
	}))
	defer srv.Close()
	withTestServer(t, srv)

	results, err := Query(context.Background(), "react", 5)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Name != "react" || results[0].Version != "18.2.0" {
		t.Errorf("unexpected first result: %+v", results[0])
	}
	if results[1].Description != "DOM bindings" {
		t.Errorf("unexpected second description: %q", results[1].Description)
	}
}

func TestQueryHandlesRegistryError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	withTestServer(t, srv)

	_, err := Query(context.Background(), "react", 5)
	if err == nil {
		t.Fatal("expected error on 503 response")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("expected error to mention status code, got: %v", err)
	}
}

func TestQueryHandlesInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()
	withTestServer(t, srv)

	_, err := Query(context.Background(), "react", 5)
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestQueryEscapesText(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"objects":[]}`))
	}))
	defer srv.Close()
	withTestServer(t, srv)

	if _, err := Query(context.Background(), "foo bar&baz", 1); err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if !strings.Contains(gotQuery, "foo+bar%26baz") && !strings.Contains(gotQuery, "foo%20bar%26baz") {
		t.Errorf("query string not URL-escaped: %q", gotQuery)
	}
}

func TestFetchDetailsAggregatesSources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "registry.npmjs.org":
			if r.URL.Path != "/react" {
				t.Errorf("unexpected packument path: %q", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{
				"name": "react",
				"description": "A JS library",
				"homepage": "https://react.dev",
				"license": "MIT",
				"author": {"name": "Meta"},
				"repository": {"type": "git", "url": "git+https://github.com/facebook/react.git"},
				"dist-tags": {"latest": "18.2.0", "next": "19.0.0-rc"},
				"time": {
					"created": "2011-10-26T00:00:00.000Z",
					"modified": "2024-01-01T00:00:00.000Z",
					"18.2.0": "2022-06-14T00:00:00.000Z",
					"17.0.2": "2021-03-22T00:00:00.000Z",
					"18.0.0": "2022-03-29T00:00:00.000Z"
				}
			}`))
		case "api.npmjs.org":
			_, _ = w.Write([]byte(`{"downloads": 42000000, "package": "react"}`))
		case "api.github.com":
			if r.URL.Path != "/repos/facebook/react" {
				t.Errorf("unexpected github path: %q", r.URL.Path)
			}
			_, _ = w.Write([]byte(`{"stargazers_count": 230000}`))
		default:
			t.Errorf("unexpected host: %q", r.Host)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	withTestServer(t, srv)

	details, err := FetchDetails(context.Background(), "react")
	if err != nil {
		t.Fatalf("FetchDetails error: %v", err)
	}
	if details.Name != "react" {
		t.Errorf("Name = %q, want react", details.Name)
	}
	if details.License != "MIT" {
		t.Errorf("License = %q, want MIT", details.License)
	}
	if details.Author != "Meta" {
		t.Errorf("Author = %q, want Meta", details.Author)
	}
	if details.Repository != "https://github.com/facebook/react" {
		t.Errorf("Repository = %q, want cleaned github URL", details.Repository)
	}
	if details.Latest != "18.2.0" {
		t.Errorf("Latest = %q, want 18.2.0", details.Latest)
	}
	if details.DistTags["next"] != "19.0.0-rc" {
		t.Errorf("DistTags[next] = %q, want 19.0.0-rc", details.DistTags["next"])
	}
	if details.WeeklyDownloads != 42000000 {
		t.Errorf("WeeklyDownloads = %d, want 42000000", details.WeeklyDownloads)
	}
	if details.Stars != 230000 {
		t.Errorf("Stars = %d, want 230000", details.Stars)
	}
	if len(details.Versions) != 3 {
		t.Fatalf("expected 3 versions (created/modified filtered), got %d", len(details.Versions))
	}
	if details.Versions[0].Version != "18.2.0" {
		t.Errorf("versions not sorted newest-first: first = %q", details.Versions[0].Version)
	}
	for i := 1; i < len(details.Versions); i++ {
		if details.Versions[i-1].PublishedAt.Before(details.Versions[i].PublishedAt) {
			t.Errorf("versions out of order at index %d", i)
		}
	}
}

func TestFetchDetailsPackumentNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "registry.npmjs.org" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"downloads": 0}`))
	}))
	defer srv.Close()
	withTestServer(t, srv)

	_, err := FetchDetails(context.Background(), "missing-pkg")
	if err == nil {
		t.Fatal("expected error when packument is 404")
	}
}

func TestFetchDetailsSkipsGitHubStarsForNonGitHubRepo(t *testing.T) {
	var githubHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "registry.npmjs.org":
			_, _ = w.Write([]byte(`{
				"name": "thing",
				"repository": "https://gitlab.com/user/thing",
				"dist-tags": {"latest": "1.0.0"},
				"time": {"1.0.0": "2024-01-01T00:00:00.000Z"}
			}`))
		case "api.npmjs.org":
			_, _ = w.Write([]byte(`{"downloads": 5}`))
		case "api.github.com":
			githubHit = true
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	withTestServer(t, srv)

	details, err := FetchDetails(context.Background(), "thing")
	if err != nil {
		t.Fatalf("FetchDetails error: %v", err)
	}
	if githubHit {
		t.Error("github API should not be called for non-github repo")
	}
	if details.Stars != 0 {
		t.Errorf("Stars = %d, want 0", details.Stars)
	}
}

func TestFetchDetailsTreatsDownloadsAsBestEffort(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "registry.npmjs.org":
			_, _ = w.Write([]byte(`{
				"name": "thing",
				"dist-tags": {"latest": "1.0.0"},
				"time": {"1.0.0": "2024-01-01T00:00:00.000Z"}
			}`))
		case "api.npmjs.org":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()
	withTestServer(t, srv)

	details, err := FetchDetails(context.Background(), "thing")
	if err != nil {
		t.Fatalf("FetchDetails should not fail when downloads API errors: %v", err)
	}
	if details.WeeklyDownloads != 0 {
		t.Errorf("WeeklyDownloads = %d, want 0 on API error", details.WeeklyDownloads)
	}
}

func TestFetchDetailsSkipsMalformedTimestamps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "registry.npmjs.org":
			_, _ = w.Write([]byte(`{
				"name": "thing",
				"dist-tags": {"latest": "1.0.0"},
				"time": {
					"1.0.0": "2024-01-01T00:00:00.000Z",
					"2.0.0": "not-a-date"
				}
			}`))
		case "api.npmjs.org":
			_, _ = w.Write([]byte(`{"downloads": 0}`))
		}
	}))
	defer srv.Close()
	withTestServer(t, srv)

	details, err := FetchDetails(context.Background(), "thing")
	if err != nil {
		t.Fatalf("FetchDetails error: %v", err)
	}
	if len(details.Versions) != 1 {
		t.Errorf("expected malformed timestamp to be skipped, got %d versions", len(details.Versions))
	}
}
