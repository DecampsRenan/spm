package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Result holds a single npm registry search result.
type Result struct {
	Name        string
	Description string
	Version     string
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

var httpClient = &http.Client{Timeout: 5 * time.Second}

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
