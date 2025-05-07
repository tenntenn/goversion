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

	"github.com/newmo-oss/ctxtime"
	"github.com/tenntenn/goplayground"
)

type Source string

const (
	SourceGoDotDev   Source = "go.dev"
	SourcePlayground Source = "play.golang.org"
)

type Latest struct {
	Version string
	Time    time.Time
	Source  Source
}

type Fetcher struct {
	url           string
	playgroundURL string
	httpClient    *http.Client
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

func (f *Fetcher) SetPlaygroundURL(url string) {
	f.playgroundURL = url
}

func (f *Fetcher) SetHTTPClient(httpClient *http.Client) {
	f.httpClient = httpClient
}

func (f *Fetcher) FetchLatest(ctx context.Context) (*Latest, error) {
	latest, err := f.fromGoDotDev(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version via go.dev/VERSION?m=text: %w", err)
	}

	now := ctxtime.Now(ctx)
	if now.Sub(latest.Time) < 24*time.Hour*30 {
		return latest, nil
	}

	// fallback
	return f.fromPlayground(ctx)
}

func (f *Fetcher) fromGoDotDev(ctx context.Context) (*Latest, error) {
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
		Source:  SourceGoDotDev,
	}, nil
}

func (f *Fetcher) fromPlayground(ctx context.Context) (*Latest, error) {
	cli := &goplayground.Client{
		HTTPClient: f.httpClient,
		BaseURL:    f.playgroundURL,
	}

	ver, err := cli.Version()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version via play.golang.org/version: %w", err)
	}

	return &Latest{
		Version: ver.Version,
		Source:  SourcePlayground,
	}, nil
}

func FetchLatest(ctx context.Context) (*Latest, error) {
	return NewFetcher().FetchLatest(ctx)
}
