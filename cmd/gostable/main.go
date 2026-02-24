package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tenntenn/goversion"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "gostable: ", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	stable, err := goversion.FetchStable(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch stable Go's version: %w", err)
	}
	fmt.Println(stable.Version)
	return nil
}
