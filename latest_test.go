package goversion_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/tenntenn/goversion"
)

func TestFetchLatest(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want    string
		wantErr bool
	}{
		"ok": {"go1.24.2\ntime 2025-03-26T19:09:39Z\n", false},
		"ng": {"", true},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(newHandler(t, tt.want))
			f := goversion.NewFetcher()
			f.SetHTTPClient(server.Client())
			f.SetURL(server.URL)

			latest, err := f.FetchLatest(t.Context())
			switch {
			case tt.wantErr && err == nil:
				t.Fatal("expected error does not occur")
			case !tt.wantErr && err != nil:
				t.Fatal("unexpected error:", err)
			case err != nil:
				return
			}

			t.Log(latest)

			var got bytes.Buffer
			fmt.Fprintln(&got, latest.Version)
			fmt.Fprintln(&got, latest.Time.Format("time 2006-01-02T15:04:05Z"))
			if diff := cmp.Diff(tt.want, got.String()); diff != "" {
				t.Error(diff)
			}
		})
	}

}

func newHandler(t *testing.T, want string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, want)
	})
}
