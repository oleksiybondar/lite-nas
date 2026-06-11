package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTimerWorkerValidatesDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewTimerWorker(TimerConfig{Interval: 0}, make(chan struct{}, 1))
	if !errors.Is(err, errInvalidTimerInterval) {
		t.Fatalf("err = %v, want %v", err, errInvalidTimerInterval)
	}
}

func TestTimerWorkerEmitsOnStartWhenConfigured(t *testing.T) {
	t.Parallel()

	output := make(chan struct{}, 1)
	worker, err := NewTimerWorker(TimerConfig{Interval: time.Hour, EmitOnStart: true}, output)
	if err != nil {
		t.Fatalf("NewTimerWorker() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)

	select {
	case <-output:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected initial tick")
	}
}

func TestTimerWorkerWaitsWhenEmitOnStartDisabled(t *testing.T) {
	t.Parallel()

	output := make(chan struct{}, 1)
	worker, err := NewTimerWorker(TimerConfig{Interval: time.Second, EmitOnStart: false}, output)
	if err != nil {
		t.Fatalf("NewTimerWorker() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)

	select {
	case <-output:
		t.Fatal("did not expect immediate tick")
	case <-time.After(30 * time.Millisecond):
	}
}
