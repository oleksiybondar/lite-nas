package loggingmanager

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeExecutor struct {
	executeCalls int
	lastTxSQL    TransactionSQL
	executeErr   error
}

func (executor *fakeExecutor) Execute(_ context.Context, txSQL TransactionSQL) error {
	executor.executeCalls++
	executor.lastTxSQL = txSQL
	if executor.executeErr != nil {
		return executor.executeErr
	}
	return nil
}

type fakeBuilder struct {
	buildCalls int
}

func (b *fakeBuilder) Build(queries []Query) TransactionSQL {
	b.buildCalls++
	return BuildTransactionSQL(queries)
}

func TestNewWriterValidatesDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewWriter(nil, &fakeBuilder{}, make(chan Query), nil, 1)
	if !errors.Is(err, errNilExecutor) {
		t.Fatalf("err = %v, want %v", err, errNilExecutor)
	}
}

func TestRunFlushesOnMaxItems(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{}
	builder := &fakeBuilder{}
	queryInCh := make(chan Query, 4)
	writer := mustNewWriter(t, executor, builder, queryInCh, nil, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(writer, ctx)

	queryInCh <- Query{SQL: "INSERT INTO events VALUES(?)", Args: []any{"a"}}
	queryInCh <- Query{SQL: "INSERT INTO events VALUES(?)", Args: []any{"b"}}

	waitForCondition(t, time.Second, func() bool { return executor.executeCalls >= 1 })
	cancel()
	waitDone(t, done)
}

func TestRunFlushesOnFlushSignal(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{}
	builder := &fakeBuilder{}
	queryInCh := make(chan Query, 4)
	flushInCh := make(chan struct{}, 1)
	writer := mustNewWriter(t, executor, builder, queryInCh, flushInCh, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(writer, ctx)
	queryInCh <- Query{SQL: "INSERT INTO occurrences VALUES(?)", Args: []any{"x"}}
	flushInCh <- struct{}{}
	time.Sleep(10 * time.Millisecond)
	flushInCh <- struct{}{}

	waitForCondition(t, time.Second, func() bool { return executor.executeCalls >= 1 })
	cancel()
	waitDone(t, done)
}

func TestRunFlushesPendingBatchOnContextCancel(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{}
	builder := &fakeBuilder{}
	queryInCh := make(chan Query, 4)
	writer := mustNewWriter(t, executor, builder, queryInCh, nil, 10)

	ctx, cancel := context.WithCancel(context.Background())
	done := runWriterAsync(writer, ctx)

	queryInCh <- Query{SQL: "UPDATE lifecycle SET muted = 1 WHERE rec_id = ?", Args: []any{1}}
	cancel()
	waitDone(t, done)

	if executor.executeCalls < 1 {
		t.Fatal("expected flush on cancel")
	}
}

func TestRunReturnsErrorForEmptySQL(t *testing.T) {
	t.Parallel()

	executor := &fakeExecutor{}
	builder := &fakeBuilder{}
	queryInCh := make(chan Query, 1)
	writer := mustNewWriter(t, executor, builder, queryInCh, nil, 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := runWriterAsync(writer, ctx)
	queryInCh <- Query{}

	err := <-done
	if !errors.Is(err, errQueryWithEmptySQL) {
		t.Fatalf("err = %v, want %v", err, errQueryWithEmptySQL)
	}
}

func mustNewWriter(
	t *testing.T,
	executor TransactionExecutor,
	builder TransactionBuilder,
	queryInCh <-chan Query,
	flushInCh <-chan struct{},
	maxItems int,
) *Writer {
	t.Helper()

	writer, err := NewWriter(executor, builder, queryInCh, flushInCh, maxItems)
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
