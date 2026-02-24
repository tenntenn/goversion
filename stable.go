package goversion

import (
	"context"
	"encoding/json"
	"fmt"
	"go/version"
	"net/http"
)

const SourceGoDotDevDL Source = "go.dev/dl"

func (f *Fetcher) SetDLURL(url string) {
	f.dlURL = url
}

func (f *Fetcher) FetchStable(ctx context.Context) (*Release, error) {
	stable, err := f.fromGoDotDevDL(ctx)
	if err == nil {
		return stable, nil
	}

	// fallback
	return f.fromPlayground(ctx)
}

type dlRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func (f *Fetcher) fromGoDotDevDL(ctx context.Context) (*Release, error) {
	dlURL := f.dlURL
	if dlURL == "" {
		dlURL = "https://go.dev/dl/?mode=json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dlURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to %q: %w", dlURL, err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Go versions via go.dev/dl (status=%d)", resp.StatusCode)
	}

	var releases []dlRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, r := range releases {
		if r.Stable && version.IsValid(r.Version) {
			return &Release{
				Version: r.Version,
				Stable:  true,
				Source:  SourceGoDotDevDL,
			}, nil
		}
	}

	return nil, fmt.Errorf("no stable version found")
}

func FetchStable(ctx context.Context) (*Release, error) {
	return NewFetcher().FetchStable(ctx)
}
