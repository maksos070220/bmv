package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/maksos070220/bmv/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.Execute(ctx); err != nil {
		if ctx.Err() != nil {
			_, err = fmt.Fprintln(os.Stderr, "Operation cancelled")
			os.Exit(130)
		}

		os.Exit(1)
	}
}
