package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// main runs the security automation NATS worker process until a termination signal.
func main() {
	if len(os.Args) > 1 {
		if len(os.Args) != 7 || os.Args[1] != "ufw-helper" {
			_, _ = fmt.Fprintln(os.Stderr, "security-automation: invalid helper invocation")
			os.Exit(2)
		}
		if err := runUFWHelper(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6]); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "security-automation: %v\n", err)
			os.Exit(1)
		}
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "security-automation: %v\n", err)
		os.Exit(1)
	}
}
