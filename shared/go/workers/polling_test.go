package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestPollingWorkerEmitsAfterTick verifies collection and output on a received tick.
func TestPollingWorkerEmitsAfterTick(t *testing.T) {
	ticks := make(chan struct{}, 1)
	output := make(chan int, 1)
	worker := NewPollingWorker(ticks, output, nil, func(context.Context) (int, error) {
		return 42, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)
	ticks <- struct{}{}

	select {
	case got := <-output:
		if got != 42 {
			t.Fatalf("snapshot = %d, want 42", got)
		}
	case <-time.After(time.Second):
		t.Fatal("snapshot was not emitted")
	}
}

// TestPollingWorkerForwardsCollectionError verifies nonblocking error delivery.
func TestPollingWorkerForwardsCollectionError(t *testing.T) {
	want := errors.New("collection failed")
	errorOutput := make(chan error, 1)
	worker := NewPollingWorker[int](nil, nil, errorOutput, func(context.Context) (int, error) {
		return 0, want
	})
	worker.PollAndSend(context.Background())
	select {
	case got := <-errorOutput:
		if !errors.Is(got, want) {
			t.Fatalf("error = %v, want %v", got, want)
		}
	default:
		t.Fatal("collection error was not emitted")
	}
}

// TestPollingWorkerStopsOnClosedTicks verifies that a closed source ends the loop.
func TestPollingWorkerStopsOnClosedTicks(t *testing.T) {
	ticks := make(chan struct{})
	close(ticks)
	worker := NewPollingWorker[int](ticks, nil, nil, nil)
	if worker.WaitNextPoll(context.Background()) {
		t.Fatal("WaitNextPoll() = true after tick channel closed")
	}
}
