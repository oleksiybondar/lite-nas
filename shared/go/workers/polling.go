package workers

import "context"

// PollingWorker runs a collection function for each tick and forwards its result.
// Collection errors are sent without blocking when an error channel is provided.
type PollingWorker[T any] struct {
	BaseWorker
	ticks  <-chan struct{}
	output chan<- T
	errors chan<- error
	poll   func(context.Context) (T, error)
}

// NewPollingWorker binds an existing tick channel and collection function to output channels.
func NewPollingWorker[T any](
	ticks <-chan struct{},
	output chan<- T,
	errors chan<- error,
	poll func(context.Context) (T, error),
) PollingWorker[T] {
	w := PollingWorker[T]{ticks: ticks, output: output, errors: errors, poll: poll}
	w.BaseWorker = NewBaseWorker(w.run)
	return w
}

// run handles ticks until the context is canceled or the tick channel closes.
func (w PollingWorker[T]) run(ctx context.Context) {
	for w.WaitNextPoll(ctx) {
		w.PollAndSend(ctx)
	}
}

// WaitNextPoll reports whether a tick arrived before cancellation or channel closure.
func (w PollingWorker[T]) WaitNextPoll(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-w.ticks:
		return ok
	}
}

// PollAndSend collects one value and forwards it unless the context is canceled.
func (w PollingWorker[T]) PollAndSend(ctx context.Context) {
	snapshot, err := w.poll(ctx)
	if err != nil {
		w.EmitError(ctx, err)
		return
	}
	select {
	case <-ctx.Done():
		return
	case w.output <- snapshot:
	}
}

// EmitError sends one collection error when the receiver is ready.
func (w PollingWorker[T]) EmitError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	select {
	case <-ctx.Done():
		return
	case w.errors <- err:
	default:
	}
}
