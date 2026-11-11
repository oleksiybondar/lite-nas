package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// main runs the operational safety service until a termination signal arrives.
func main() {
	if len(os.Args) > 1 {
		if len(os.Args) == 3 && os.Args[1] == "manual-cpu-cap-helper" {
			if err := runManualCPUCapHelper(os.Args[2]); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
		if len(os.Args) != 4 || os.Args[1] != "cpu-cap-helper" {
			_, _ = fmt.Fprintln(os.Stderr, "invalid helper invocation")
			os.Exit(2)
		}
		if err := runCPUCapHelper(os.Args[2], os.Args[3]); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "operational-safety-automation: %v\n", err)
		os.Exit(1)
	}
}
