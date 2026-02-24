package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tenntenn/goversion"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gover:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	subcmd := "latest"
	if len(args) > 0 {
		subcmd = args[0]
	}

	switch subcmd {
	case "latest":
		r, err := goversion.FetchLatest(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch latest Go version: %w", err)
		}
		fmt.Println(r.Version)
	case "stable":
		r, err := goversion.FetchStable(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch stable Go version: %w", err)
		}
		fmt.Println(r.Version)
	default:
		return fmt.Errorf("unknown subcommand: %s (available: latest, stable)", subcmd)
	}

	return nil
}
