package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// main runs the authenticated journal command worker until shutdown.
func main() {
	if len(os.Args) > 1 {
		if err := runVacuum(os.Args[1:]); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "journal-manager vacuum: %v\n", err)
			os.Exit(1)
		}
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintf(os.Stderr, "journal-manager: %v\n", err)
		os.Exit(1)
	}
}

// runVacuum is the narrow privileged entrypoint used by the sudoers rule.
func runVacuum(args []string) error {
	if len(args) != 2 || args[0] != "vacuum" {
		return fmt.Errorf("expected vacuum DAYS")
	}
	days, err := strconv.Atoi(args[1])
	if err != nil || days < 1 || days > 365 || strconv.Itoa(days) != args[1] {
		return fmt.Errorf("invalid retention days")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = (execRunner{}).Output(ctx, journalctlBinary, "--rotate", "--vacuum-time="+args[1]+"days", "--no-pager")
	return err
}
