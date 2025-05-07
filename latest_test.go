package goversion_test

import (
	"encoding/json"
	"fmt"
	"go/version"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/newmo-oss/ctxtime/ctxtimetest"
	"github.com/newmo-oss/testid"

	"github.com/tenntenn/goplayground"

	"github.com/tenntenn/goversion"
)

func TestFetchLatest(t *testing.T) {
	t.Parallel()

	latest := func(ver string, tm time.Time, src goversion.Source) *goversion.Latest {
		return &goversion.Latest{
			Version: ver,
			Time:    tm,
			Source:  src,
		}
	}

	baseTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	var zeroTime time.Time

	cases := map[string]struct {
		now     time.Time
		want    *goversion.Latest
		wantErr bool
	}{
		"go.dev":     {baseTime, latest("go1.24.2", baseTime, goversion.SourceGoDotDev), false},
		"playground": {baseTime.Add(24 * time.Hour * 31), latest("go1.24.2", zeroTime, goversion.SourcePlayground), false},
		"ng":         {zeroTime, nil, true},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := testid.WithValue(t.Context(), testid.New(t))
			ctxtimetest.SetFixedNow(t, ctx, tt.now)

			server := httptest.NewServer(newHandler(t, tt.want))
			f := goversion.NewFetcher()
			f.SetHTTPClient(server.Client())
			f.SetURL(server.URL + "?src=go.dev")
			f.SetPlaygroundURL(server.URL)

			got, err := f.FetchLatest(ctx)
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

func newHandler(t *testing.T, want *goversion.Latest) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want == nil {
			http.Error(w, "want is nil", http.StatusInternalServerError)
			return
		}

		// handle go.dev
		if r.FormValue("src") == "go.dev" {
			fmt.Fprintln(w, want.Version)
			fmt.Fprintln(w, want.Time.Format("time 2006-01-02T15:04:05Z"))
		} else { // handle play.golang.org
			lang := version.Lang(want.Version)
			ver := goplayground.VersionResult{
				Version: want.Version,
				Release: lang,
				Name:    "Go " + strings.TrimPrefix(lang, "go"),
			}
			if err := json.NewEncoder(w).Encode(ver); err != nil {
				t.Fatal("unexpected error:", err)
			}
		}
	})
}
