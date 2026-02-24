package goversion

import (
	"context"
	"encoding/json"
	"fmt"
	"go/version"
	"net/http"

	"github.com/newmo-oss/ctxtime"
	"time"
)

const SourceGoDotDevDL Source = "go.dev/dl"

func (f *Fetcher) SetDLURL(url string) {
	f.dlURL = url
}

func (f *Fetcher) FetchStable(ctx context.Context) (*Latest, error) {
	stable, err := f.fromGoDotDevDL(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stable version via go.dev/dl/?mode=json: %w", err)
	}

	now := ctxtime.Now(ctx)
	if now.Sub(stable.Time) < 24*time.Hour*30 {
		return stable, nil
	}

	// fallback
	return f.fromPlayground(ctx)
}

type dlRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func (f *Fetcher) fromGoDotDevDL(ctx context.Context) (*Latest, error) {
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
			return &Latest{
				Version: r.Version,
				Source:  SourceGoDotDevDL,
			}, nil
		}
	}

	return nil, fmt.Errorf("no stable version found")
}

func FetchStable(ctx context.Context) (*Latest, error) {
	return NewFetcher().FetchStable(ctx)
}
