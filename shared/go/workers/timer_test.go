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

	_, err = NewTimerWorker(TimerConfig{Interval: time.Second}, nil)
	if !errors.Is(err, errNilTimerOutput) {
		t.Fatalf("err = %v, want %v", err, errNilTimerOutput)
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

func TestTimerWorkerEmitsAfterTickerFires(t *testing.T) {
	t.Parallel()

	output := make(chan struct{}, 1)
	worker, err := NewTimerWorker(TimerConfig{Interval: time.Millisecond}, output)
	if err != nil {
		t.Fatalf("NewTimerWorker() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)

	select {
	case <-output:
	case <-time.After(time.Second):
		t.Fatal("expected tick after timer interval")
	}
}

func TestTimerWorkerEmitTickStopsWhenContextCanceled(t *testing.T) {
	t.Parallel()

	worker, err := NewTimerWorker(TimerConfig{Interval: time.Second}, make(chan struct{}))
	if err != nil {
		t.Fatalf("NewTimerWorker() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if worker.emitTick(ctx) {
		t.Fatal("emitTick() = true, want false when context is canceled")
	}
}

func TestTimerWorkerRunLoopStopsWhenContextCanceled(t *testing.T) {
	t.Parallel()

	worker, err := NewTimerWorker(TimerConfig{Interval: time.Second}, make(chan struct{}, 1))
	if err != nil {
		t.Fatalf("NewTimerWorker() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	go func() {
		worker.runLoop(ctx, ticker)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runLoop() did not stop after context cancellation")
	}
}

func TestNewPollingTimerWorkerReturnsConfiguredWorkerAndChannel(t *testing.T) {
	t.Parallel()

	worker, ticks, err := NewPollingTimerWorker(5*time.Second, 2)
	if err != nil {
		t.Fatalf("NewPollingTimerWorker() error = %v", err)
	}
	if ticks == nil {
		t.Fatal("ticks = nil, want allocated channel")
	}
	if got := cap(ticks); got != 2 {
		t.Fatalf("cap(ticks) = %d, want 2", got)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)

	select {
	case <-ticks:
	case <-time.After(time.Second):
		t.Fatal("expected initial polling tick")
	}
}

func TestNewPollingTimerWorkerReturnsValidationError(t *testing.T) {
	t.Parallel()

	_, ticks, err := NewPollingTimerWorker(0, 1)
	if !errors.Is(err, errInvalidTimerInterval) {
		t.Fatalf("err = %v, want %v", err, errInvalidTimerInterval)
	}
	if ticks != nil {
		t.Fatal("ticks != nil, want nil on validation error")
	}
}
