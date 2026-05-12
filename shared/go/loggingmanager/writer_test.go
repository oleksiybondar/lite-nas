package loggingmanager

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/shared/loggingmanager/query"
)

type fakeExecutor struct {
	executeCalls int
	lastTxSQL    query.TransactionSQL
	txHistory    []query.TransactionSQL
	executeErr   error
}

func (executor *fakeExecutor) Execute(_ context.Context, txSQL query.TransactionSQL) error {
	executor.executeCalls++
	executor.lastTxSQL = txSQL
	executor.txHistory = append(executor.txHistory, txSQL)
	if executor.executeErr != nil {
		return executor.executeErr
	}
	return nil
}

type fakeBuilder struct {
	buildCalls int
}

func (b *fakeBuilder) Build(queries []query.Query) query.TransactionSQL {
	b.buildCalls++
	return query.BuildTransactionSQL(queries)
}

func TestNewWriterValidatesDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewWriter(nil, &fakeBuilder{}, make(chan WriteRequest), nil, 1, 100)
	if !errors.Is(err, errNilExecutor) {
		t.Fatalf("err = %v, want %v", err, errNilExecutor)
	}
}

func TestRunFlushesOnMaxItems(t *testing.T) {
	t.Parallel()

	rig := newWriterTestRig(t, 4, 0, 2, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(rig.writer, ctx)

	rig.queryInCh <- WriteRequest{Query: query.Query{SQL: "INSERT INTO events VALUES(?)", Args: []any{"a"}}}
	rig.queryInCh <- WriteRequest{Query: query.Query{SQL: "INSERT INTO events VALUES(?)", Args: []any{"b"}}}

	waitForCondition(t, time.Second, func() bool { return rig.executor.executeCalls >= 1 })
	cancel()
	waitDone(t, done)
}

func TestRunFlushesOnFlushSignal(t *testing.T) {
	t.Parallel()

	rig := newWriterTestRig(t, 4, 1, 100, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(rig.writer, ctx)
	rig.queryInCh <- WriteRequest{
		Query:              query.Query{SQL: "INSERT INTO occurrences VALUES(?)", Args: []any{"x"}},
		TouchesOccurrences: true,
	}
	rig.flushInCh <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	rig.flushInCh <- struct{}{}

	waitForCondition(t, time.Second, func() bool { return rig.executor.executeCalls >= 1 })
	cancel()
	waitDone(t, done)
}

func TestRunFlushesPendingBatchOnContextCancel(t *testing.T) {
	t.Parallel()

	rig := newWriterTestRig(t, 4, 0, 10, 100)

	ctx, cancel := context.WithCancel(context.Background())
	done := runWriterAsync(rig.writer, ctx)

	rig.queryInCh <- WriteRequest{Query: query.Query{SQL: "UPDATE lifecycle SET muted = 1 WHERE rec_id = ?", Args: []any{1}}}
	cancel()
	waitDone(t, done)

	if rig.executor.executeCalls < 1 {
		t.Fatal("expected flush on cancel")
	}
}

func TestRunReturnsErrorForEmptySQL(t *testing.T) {
	t.Parallel()

	rig := newWriterTestRig(t, 1, 0, 10, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(rig.writer, ctx)
	rig.queryInCh <- WriteRequest{Query: query.Query{}}

	err := <-done
	if !errors.Is(err, errQueryWithEmptySQL) {
		t.Fatalf("err = %v, want %v", err, errQueryWithEmptySQL)
	}
}

func TestRunFlushesMixedByCountAndTick(t *testing.T) {
	t.Parallel()

	rig := newWriterTestRig(t, 16, 1, 5, 1000)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(rig.writer, ctx)
	for idx := range 8 {
		rig.queryInCh <- WriteRequest{
			Query: query.Query{
				SQL:  "INSERT INTO events VALUES(?)",
				Args: []any{idx},
			},
		}
	}

	waitForCondition(t, time.Second, func() bool { return rig.executor.executeCalls >= 1 })
	rig.flushInCh <- struct{}{}
	waitForCondition(t, time.Second, func() bool { return rig.executor.executeCalls >= 2 })

	cancel()
	waitDone(t, done)

	if len(rig.executor.txHistory) < 2 {
		t.Fatalf("txHistory len = %d, want at least 2", len(rig.executor.txHistory))
	}
	if got := len(rig.executor.txHistory[0].Queries); got != 5 {
		t.Fatalf("first tx queries = %d, want 5", got)
	}
	if got := len(rig.executor.txHistory[1].Queries); got != 3 {
		t.Fatalf("second tx queries = %d, want 3", got)
	}
}

type writerTestRig struct {
	executor  *fakeExecutor
	queryInCh chan WriteRequest
	flushInCh chan struct{}
	writer    *Writer
}

func newWriterTestRig(
	t *testing.T,
	queryBuffer int,
	flushBuffer int,
	maxItems int,
	maxOccurrences int,
) writerTestRig {
	t.Helper()

	executor := &fakeExecutor{}
	builder := &fakeBuilder{}
	queryInCh := make(chan WriteRequest, queryBuffer)
	var flushInCh chan struct{}
	if flushBuffer > 0 {
		flushInCh = make(chan struct{}, flushBuffer)
	}
	writer := mustNewWriter(t, executor, builder, queryInCh, flushInCh, maxItems, maxOccurrences)
	return writerTestRig{
		executor:  executor,
		queryInCh: queryInCh,
		flushInCh: flushInCh,
		writer:    writer,
	}
}

func mustNewWriter(
	t *testing.T,
	executor TransactionExecutor,
	builder TransactionBuilder,
	queryInCh <-chan WriteRequest,
	flushInCh <-chan struct{},
	maxItems int,
	maxOccurrences int,
) *Writer {
	t.Helper()

	writer, err := NewWriter(executor, builder, queryInCh, flushInCh, maxItems, maxOccurrences)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}
	return writer
}

func runWriterAsync(writer *Writer, ctx context.Context) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- writer.Run(ctx)
	}()
	return done
}

func waitDone(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("writer.Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for writer to stop")
	}
}

func waitForCondition(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}
