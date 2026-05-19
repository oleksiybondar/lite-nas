package main

import (
	"context"
	"fmt"
)

// run is a temporary stub until zfs-metrics-cli request/format workflow is implemented.
func run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Println("zfs-metrics-cli stub: not implemented yet")
		return nil
	}
}
