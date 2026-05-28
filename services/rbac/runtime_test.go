package main

import (
	"context"
	"errors"
	"testing"
)

func TestRunReturnsCanceledContextErrorOnGracefulShutdown(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v, want %v", err, context.Canceled)
	}
}
