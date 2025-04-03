# goversion

[![pkg.go.dev][gopkg-badge]][gopkg]

`goversion` provides utilities for versioning in Go.

## CLI

```
$ go install github.com/tenntenn/goversion/cmd/golatest@latest
$ golatest
go1.24.2
```

## Pacakge

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tenntenn/goversion"
)

func main() {
	// fetch latest go version via "https://go.dev/VERSION?m=text"
	ctx := context.Context()
	latest, err := goversion.FetchLatest(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println(latest.Version, latest.Time)
}
```

<!-- links -->
[gopkg]: https://pkg.go.dev/github.com/tenntenn/goversion
[gopkg-badge]: https://pkg.go.dev/badge/github.com/tenntenn/goversion?status.svg
