package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	backuprestore "lite-nas/services/backup-restore"
)

// main runs the on-demand backup service until a termination signal arrives.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := backuprestore.Serve(ctx, "/etc/lite-nas/backup-restore.conf"); err != nil && !errors.Is(err, context.Canceled) {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
