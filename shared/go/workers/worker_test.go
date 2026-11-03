package workers

import (
	"context"
	"testing"
	"time"
)

// TestBaseWorkerStartsBoundRunLoop verifies startup passes the caller's context.
func TestBaseWorkerStartsBoundRunLoop(t *testing.T) {
	ctx := context.WithValue(context.Background(), workerContextKey{}, "caller")
	values := make(chan any, 1)
	worker := NewBaseWorker(func(runCtx context.Context) {
		values <- runCtx.Value(workerContextKey{})
	})
	worker.Start(ctx)

	select {
	case got := <-values:
		if got != "caller" {
			t.Fatalf("context value = %v, want caller", got)
		}
	case <-time.After(time.Second):
		t.Fatal("run loop did not start")
	}
}

// workerContextKey isolates the test's context value from other context users.
type workerContextKey struct{}
