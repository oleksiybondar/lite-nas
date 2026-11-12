package workers

import "context"

// Worker starts asynchronous work that runs under a context.
type Worker interface {
	// Start launches the worker's loop under the supplied context.
	Start(context.Context)
}

// BaseWorker supplies goroutine startup for workers with their own run loop.
type BaseWorker struct {
	run func(context.Context)
}

// NewBaseWorker binds a non-nil run loop used when the worker starts.
func NewBaseWorker(run func(context.Context)) BaseWorker {
	return BaseWorker{run: run}
}

// Start launches the bound run loop in a separate goroutine.
func (w BaseWorker) Start(ctx context.Context) {
	go w.run(ctx)
}
