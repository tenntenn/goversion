package goversion

import (
	"context"
	"fmt"
	"go/version"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

type Latest struct {
	Version string
	Time    time.Time
}

type Fetcher struct {
	url        string
	httpClient *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		url:        "https://go.dev/VERSION?m=text",
		httpClient: http.DefaultClient,
	}
}

func (f *Fetcher) SetURL(url string) {
	f.url = url
}

func (f *Fetcher) SetHTTPClient(httpClient *http.Client) {
	f.httpClient = httpClient
}

func (f *Fetcher) FetchLatest(ctx context.Context) (*Latest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to %q: %w", f.url, err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest Go version via go.dev (status=%d): %w", resp.StatusCode, err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	lines := slices.Collect(strings.Lines(string(data)))
	if len(lines) < 2 {
		return nil, fmt.Errorf("failed to parse body")
	}

	goversion := strings.TrimSpace(lines[0])
	if !version.IsValid(goversion) {
		return nil, fmt.Errorf("invalid go version (%q): %w", goversion, err)
	}

	tmstr := strings.TrimSpace(lines[1])
	tm, err := time.Parse("time 2006-01-02T15:04:05Z", tmstr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time (%q): %w", tmstr, err)
	}

	return &Latest{
		Version: goversion,
		Time:    tm,
	}, nil
}

func FetchLatest(ctx context.Context) (*Latest, error) {
	return NewFetcher().FetchLatest(ctx)
}
