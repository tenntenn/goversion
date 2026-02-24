package goversion_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/tenntenn/goversion"
)

func TestFetchStable(t *testing.T) {
	t.Parallel()

	stable := func(ver string, src goversion.Source) *goversion.Stable {
		return &goversion.Stable{
			Version: ver,
			Source:  src,
		}
	}

	cases := map[string]struct {
		releases []dlRelease
		want     *goversion.Stable
		wantErr  bool
	}{
		"latest_stable": {
			releases: []dlRelease{
				{Version: "go1.24.2", Stable: true},
				{Version: "go1.23.8", Stable: true},
			},
			want: stable("go1.24.2", goversion.SourceGoDotDevDL),
		},
		"skip_unstable": {
			releases: []dlRelease{
				{Version: "go1.25rc1", Stable: false},
				{Version: "go1.24.2", Stable: true},
				{Version: "go1.23.8", Stable: true},
			},
			want: stable("go1.24.2", goversion.SourceGoDotDevDL),
		},
		"no_stable": {
			releases: []dlRelease{
				{Version: "go1.25rc1", Stable: false},
			},
			wantErr: true,
		},
		"empty": {
			releases: []dlRelease{},
			wantErr:  true,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(newDLHandler(t, tt.releases))
			defer server.Close()

			f := goversion.NewFetcher()
			f.SetHTTPClient(server.Client())
			f.SetDLURL(server.URL)

			got, err := f.FetchStable(t.Context())
			switch {
			case tt.wantErr && err == nil:
				t.Fatal("expected error does not occur")
			case !tt.wantErr && err != nil:
				t.Fatal("unexpected error:", err)
			case err != nil:
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func newDLHandler(t *testing.T, releases []dlRelease) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Fatal("unexpected error:", err)
		}
	})
}

type dlRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}
