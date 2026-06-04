package loggingmanager

import (
	"context"
	"errors"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/query"
)

var (
	errNilExecutor           = errors.New("loggingmanager writer transaction executor is required")
	errNilTransactionBuilder = errors.New("loggingmanager writer transaction builder is required")
	errNilQueryInputChannel  = errors.New("loggingmanager writer request input channel is required")
	errInvalidMaxItems       = errors.New("loggingmanager writer max items must be greater than zero")
	errQueryWithEmptySQL     = errors.New("loggingmanager writer query SQL is required")
)

// TransactionBuilder constructs transaction data from a sealed query batch.
//
// Contract:
//   - Implementations must preserve query execution order.
//   - Implementations should treat the provided query slice as read-only input.
//
// Architectural role:
//   - Separates batch-shaping concerns from writer loop and execution concerns.
type TransactionBuilder interface {
	// Build returns transaction framing and statements for the batch.
	Build(queries []query.Query) query.TransactionSQL
}

// DefaultTransactionBuilder frames batches with BEGIN/COMMIT markers and the
// supplied query list.
type DefaultTransactionBuilder struct{}

// Build returns BEGIN/COMMIT framed SQL for the supplied batch.
func (DefaultTransactionBuilder) Build(queries []query.Query) query.TransactionSQL {
	return query.BuildTransactionSQL(queries)
}

// Writer is a single-goroutine async SQL batch writer.
//
// Ownership model:
//   - Writer exclusively owns mutable batch state while Run is active.
//   - Producers and flush triggers are external and communicate only through
//     channels.
//
// Interaction model:
//   - Queries arrive through queryInCh.
//   - Flush signals arrive through flushInCh.
//   - Batch execution is delegated to TransactionExecutor.
//
// Side effects:
//   - Performs persistence I/O only through the injected executor.
type Writer struct {
	executor       TransactionExecutor
	builder        TransactionBuilder
	requestInCh    <-chan WriteRequest
	flushInCh      <-chan struct{}
	maxItems       int
	maxOccurrences int
}

// NewWriter builds a channel-driven writer and validates dependencies.
//
// Preconditions:
//   - executor, builder, and queryInCh must be non-nil.
//   - maxItems must be greater than zero.
//
// Side effects:
//   - None. The writer loop starts only when Run is called.
func NewWriter(
	executor TransactionExecutor,
	builder TransactionBuilder,
	requestInCh <-chan WriteRequest,
	flushInCh <-chan struct{},
	maxItems int,
	maxOccurrences int,
) (*Writer, error) {
	if executor == nil {
		return nil, errNilExecutor
	}
	if builder == nil {
		return nil, errNilTransactionBuilder
	}
	if requestInCh == nil {
		return nil, errNilQueryInputChannel
	}
	if maxItems <= 0 {
		return nil, errInvalidMaxItems
	}

	return &Writer{
		executor:       executor,
		builder:        builder,
		requestInCh:    requestInCh,
		flushInCh:      flushInCh,
		maxItems:       maxItems,
		maxOccurrences: maxOccurrences,
	}, nil
}

// Run executes the writer loop as the single owner of mutable write state.
//
// Lifecycle:
//   - Runs until context cancellation or query input-channel closure.
//   - On cancellation/closure, drains queued input and performs one final flush
//     attempt for accumulated queries.
//
// Error behavior:
//   - Returns dependency/validation or execution errors immediately.
//   - Returns nil on graceful stop after successful finalization.
//
// Concurrency contract:
//   - Run must be called once per Writer instance.
func (w *Writer) Run(ctx context.Context) error {
	batch := make([]query.Query, 0, w.maxItems)
	tailState := newDeferredTailState()

	for {
		stop, err := w.runStep(ctx, &batch, tailState)
		if err != nil {
			return err
		}
		if stop {
			return nil
		}
	}
}

// runStep handles one writer-loop iteration and reports whether Run should stop.
//
// A true stop result indicates graceful termination after finalization work was
// completed or that a terminal error was returned.
func (w *Writer) runStep(
	ctx context.Context,
	batch *[]query.Query,
	tailState *deferredTailState,
) (bool, error) {
	select {
	case <-ctx.Done():
		return true, w.handleContextCancel(batch, tailState)
	case request, ok := <-w.requestInCh:
		err := w.handleWriteRequest(request, ok, batch, tailState)
		if err != nil {
			return true, err
		}
		return !ok, nil
	case <-w.flushInCh:
		return false, w.handleFlushSignal(batch, tailState)
	}
}

// flush persists one sealed batch in a single SQL transaction unit.
//
// Preconditions:
//   - batch is expected to contain at least one query.
//
// Side effects:
//   - Delegates transactional persistence I/O to the configured executor.
func (w *Writer) flush(ctx context.Context, batch []query.Query, tailState *deferredTailState) error {
	transactionSQL := w.builder.Build(w.buildBatchWithDeferredTail(batch, tailState))
	return w.executor.Execute(ctx, transactionSQL)
}

// flushIfNeeded flushes only when the batch is non-empty.
//
// Side effects:
//   - May perform persistence I/O through flush.
func (w *Writer) flushIfNeeded(
	ctx context.Context,
	batch []query.Query,
	tailState *deferredTailState,
) error {
	if len(batch) == 0 {
		return nil
	}

	return w.flush(ctx, batch, tailState)
}

// drainQueryChannel non-blockingly appends queued items before shutdown flush.
//
// Behavior:
//   - Drains only currently available items and stops when the channel would
//     block.
//   - If queryInCh is closed, draining stops immediately.
func (w *Writer) drainWriteChannel(batch *[]query.Query, tailState *deferredTailState) {
	for {
		select {
		case request, ok := <-w.requestInCh:
			if !ok {
				return
			}
			*batch = append(*batch, request.Query)
			tailState.capture(request)
		default:
			return
		}
	}
}

// handleContextCancel drains queued items and performs a final flush attempt.
//
// Side effects:
//   - May execute persistence I/O if the drained batch is non-empty.
func (w *Writer) handleContextCancel(batch *[]query.Query, tailState *deferredTailState) error {
	w.drainWriteChannel(batch, tailState)
	return w.flushIfNeeded(context.Background(), *batch, tailState)
}

// handleWriteRequest appends one query and flushes when maxItems is reached.
//
// When ok is false the input channel is closed and pending data is flushed.
//
// Preconditions:
//   - query.SQL must be non-empty when ok is true.
//
// Side effects:
//   - May execute persistence I/O when flush thresholds are reached.
func (w *Writer) handleWriteRequest(
	request WriteRequest,
	ok bool,
	batch *[]query.Query,
	tailState *deferredTailState,
) error {
	if !ok {
		return w.flushIfNeeded(context.Background(), *batch, tailState)
	}
	if request.Query.SQL == "" {
		return errQueryWithEmptySQL
	}

	*batch = append(*batch, request.Query)
	tailState.capture(request)
	if len(*batch) < w.maxItems {
		return nil
	}

	if err := w.flush(context.Background(), *batch, tailState); err != nil {
		return err
	}
	*batch = (*batch)[:0]
	tailState.reset()
	return nil
}

// handleFlushSignal flushes the current batch when it is non-empty.
//
// Side effects:
//   - May execute persistence I/O through flush.
func (w *Writer) handleFlushSignal(batch *[]query.Query, tailState *deferredTailState) error {
	if len(*batch) == 0 {
		return nil
	}

	if err := w.flush(context.Background(), *batch, tailState); err != nil {
		return err
	}

	*batch = (*batch)[:0]
	tailState.reset()
	return nil
}

type deferredTailState struct {
	runtimeStateByKey  map[string]string
	touchesOccurrences bool
}

// newDeferredTailState constructs empty deferred tail accumulation state.
func newDeferredTailState() *deferredTailState {
	return &deferredTailState{
		runtimeStateByKey: make(map[string]string),
	}
}

// capture accumulates tail intents from one write request.
func (state *deferredTailState) capture(request WriteRequest) {
	for _, update := range request.RuntimeStateUpdates {
		state.runtimeStateByKey[update.Key] = update.Value
	}
	if request.TouchesOccurrences {
		state.touchesOccurrences = true
	}
}

// reset clears accumulated tail intents after a successful flush.
func (state *deferredTailState) reset() {
	for key := range state.runtimeStateByKey {
		delete(state.runtimeStateByKey, key)
	}
	state.touchesOccurrences = false
}

// buildBatchWithDeferredTail appends deferred runtime-state tail queries to the
// current sealed batch.
func (w *Writer) buildBatchWithDeferredTail(
	batch []query.Query,
	tailState *deferredTailState,
) []query.Query {
	queries := make([]query.Query, 0, len(batch)+len(tailState.runtimeStateByKey))
	queries = append(queries, batch...)
	queries = append(queries, runtimeStateTailQueries(tailState.runtimeStateByKey)...)
	return queries
}

// runtimeStateTailQueries converts runtime-state map entries to stable-ordered
// upsert queries.
func runtimeStateTailQueries(runtimeStateByKey map[string]string) []query.Query {
	keysInOrder := []string{
		query.RuntimeStateCurrentEventRecIDKey,
		query.RuntimeStateCurrentEventSeqKey,
		query.RuntimeStateEventIDPrefixKey,
	}
	queries := make([]query.Query, 0, len(runtimeStateByKey))
	for _, key := range keysInOrder {
		value, ok := runtimeStateByKey[key]
		if !ok {
			continue
		}
		queries = append(queries, query.UpsertRuntimeState(dto.RuntimeStateRow{
			Key:   key,
			Value: value,
		}))
	}
	return queries
}
