package workers

import (
	"context"
	"errors"
	"time"
)

var (
	errInvalidTimerInterval = errors.New("timer interval must be greater than zero")
	errNilTimerOutput       = errors.New("timer output channel is required")
)

// TimerConfig defines runtime behavior for TimerWorker.
type TimerConfig struct {
	Interval    time.Duration
	EmitOnStart bool
}

// TimerWorker emits periodic tick signals into an output channel.
//
// The worker owns ticker lifecycle only; it does not own consumer logic.
type TimerWorker struct {
	config TimerConfig
	output chan<- struct{}
}

// NewTimerWorker creates a TimerWorker with validated configuration.
func NewTimerWorker(config TimerConfig, output chan<- struct{}) (TimerWorker, error) {
	if config.Interval <= 0 {
		return TimerWorker{}, errInvalidTimerInterval
	}
	if output == nil {
		return TimerWorker{}, errNilTimerOutput
	}

	return TimerWorker{
		config: config,
		output: output,
	}, nil
}

// Start launches the timer worker in a separate goroutine.
func (w TimerWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run emits configured ticks until context cancellation.
func (w TimerWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	if w.config.EmitOnStart && !w.emitTick(ctx) {
		return
	}

	w.runLoop(ctx, ticker)
}

// run emits configured ticks until context cancellation.
func (w TimerWorker) runLoop(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !w.emitTick(ctx) {
				return
			}
		}
	}
}

// emitTick sends one tick or stops when context is canceled.
func (w TimerWorker) emitTick(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case w.output <- struct{}{}:
		return true
	}
}
