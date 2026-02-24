package goversion_test

import (
	"encoding/json"
	"go/version"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tenntenn/goplayground"

	"github.com/tenntenn/goversion"
)

func TestFetchStable(t *testing.T) {
	t.Parallel()

	release := func(ver string, stable bool, src goversion.Source) *goversion.Release {
		return &goversion.Release{
			Version: ver,
			Stable:  stable,
			Source:  src,
		}
	}

	cases := map[string]struct {
		releases []dlRelease
		dlFail   bool
		want     *goversion.Release
		wantErr  bool
	}{
		"go.dev/dl": {
			releases: []dlRelease{
				{Version: "go1.24.2", Stable: true},
				{Version: "go1.23.8", Stable: true},
			},
			want: release("go1.24.2", true, goversion.SourceGoDotDevDL),
		},
		"skip_unstable": {
			releases: []dlRelease{
				{Version: "go1.25rc1", Stable: false},
				{Version: "go1.24.2", Stable: true},
				{Version: "go1.23.8", Stable: true},
			},
			want: release("go1.24.2", true, goversion.SourceGoDotDevDL),
		},
		"playground": {
			dlFail: true,
			want:   release("go1.24.2", false, goversion.SourcePlayground),
		},
		"no_stable": {
			releases: []dlRelease{
				{Version: "go1.25rc1", Stable: false},
			},
			// fallback to playground
			want: release("go1.24.2", false, goversion.SourcePlayground),
		},
		"empty": {
			releases: []dlRelease{},
			// fallback to playground
			want: release("go1.24.2", false, goversion.SourcePlayground),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(newDLHandler(t, tt.releases, tt.dlFail, tt.want))
			defer server.Close()

			f := goversion.NewFetcher()
			f.SetHTTPClient(server.Client())
			f.SetDLURL(server.URL + "?src=dl")
			f.SetPlaygroundURL(server.URL)

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

func newDLHandler(t *testing.T, releases []dlRelease, dlFail bool, want *goversion.Release) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("src") == "dl" {
			if dlFail {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(releases); err != nil {
				t.Fatal("unexpected error:", err)
			}
			return
		}

		// handle play.golang.org fallback
		if want == nil {
			http.Error(w, "want is nil", http.StatusInternalServerError)
			return
		}
		lang := version.Lang(want.Version)
		ver := goplayground.VersionResult{
			Version: want.Version,
			Release: lang,
			Name:    "Go " + strings.TrimPrefix(lang, "go"),
		}
		if err := json.NewEncoder(w).Encode(ver); err != nil {
			t.Fatal("unexpected error:", err)
		}
	})
}

type dlRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}
