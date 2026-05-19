package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// main starts zfs-metrics-cli process and handles termination signals.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "zfs-metrics-cli: %v\n", err)
		os.Exit(1)
	}
}
