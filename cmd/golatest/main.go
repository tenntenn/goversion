package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tentenn/goversion"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "golatest: ", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	latest, err := goversion.FetchLatest(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch latest Go's version: %w", err)
	}
	fmt.Println(latest.Version)
	return nil
}
