package workers

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/services/system-metrics/config"
	"lite-nas/shared/metrics"
)

type fakeReader struct {
	data []byte
	err  error
}

func (r fakeReader) Read() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.data, nil
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerPollFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(metrics.RawSystemSnapshot) any
		want any
	}{
		{name: "cpu total", got: func(snapshot metrics.RawSystemSnapshot) any { return snapshot.CPU.Total.Total }, want: uint64(126)},
		{name: "memory used bytes", got: func(snapshot metrics.RawSystemSnapshot) any { return snapshot.Mem.UsedBytes }, want: uint64(1024 * 750)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			snapshot := loadPollingSnapshotFixture(t)
			if got := testCase.got(snapshot); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerPollReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read failed")
	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{err: expectedErr},
		fakeReader{},
		make(chan metrics.RawSystemSnapshot, 1),
	)

	if _, err := worker.poll(); !errors.Is(err, expectedErr) {
		t.Fatalf("poll() error = %v, want %v", err, expectedErr)
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerWaitNextPollStopsOnContextCancellation(t *testing.T) {
	t.Parallel()

	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{},
		fakeReader{},
		make(chan metrics.RawSystemSnapshot, 1),
	)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if worker.waitNextPoll(ctx, ticker) {
		t.Fatal("expected waitNextPoll() to stop on canceled context")
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerWaitNextPollContinuesOnTick(t *testing.T) {
	t.Parallel()

	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{},
		fakeReader{},
		make(chan metrics.RawSystemSnapshot, 1),
	)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	if !worker.waitNextPoll(context.Background(), ticker) {
		t.Fatal("expected waitNextPoll() to continue on tick")
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerPollAndSendDoesNotEmitOnCanceledContext(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.RawSystemSnapshot)
	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{data: []byte("cpu  10 20 30 40 5 6 7 8 0 0\ncpu0 1 2 3 4 1 0 0 0 0 0\n")},
		fakeReader{data: []byte("MemTotal: 1000 kB\nMemAvailable: 250 kB\n")},
		output,
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker.pollAndSend(ctx)

	select {
	case <-output:
		t.Fatal("did not expect output on canceled context")
	default:
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerPollReturnsMemReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("mem read failed")
	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{data: []byte("cpu  10 20 30 40 5 6 7 8 0 0\ncpu0 1 2 3 4 1 0 0 0 0 0\n")},
		fakeReader{err: expectedErr},
		make(chan metrics.RawSystemSnapshot, 1),
	)

	if _, err := worker.poll(); !errors.Is(err, expectedErr) {
		t.Fatalf("poll() error = %v, want %v", err, expectedErr)
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerPollReturnsCPUParseError(t *testing.T) {
	t.Parallel()

	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{data: []byte("invalid cpu data")},
		fakeReader{data: []byte("MemTotal: 1000 kB\nMemAvailable: 250 kB\n")},
		make(chan metrics.RawSystemSnapshot, 1),
	)

	if _, err := worker.poll(); err == nil {
		t.Fatal("expected cpu parse error")
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerPollReturnsMemParseError(t *testing.T) {
	t.Parallel()

	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{data: []byte("cpu  10 20 30 40 5 6 7 8 0 0\ncpu0 1 2 3 4 1 0 0 0 0 0\n")},
		fakeReader{data: []byte("invalid mem data")},
		make(chan metrics.RawSystemSnapshot, 1),
	)

	if _, err := worker.poll(); err == nil {
		t.Fatal("expected mem parse error")
	}
}

// Requirements: system-metrics-svc/FR-001
func TestPollingWorkerStartEmitsSnapshot(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.RawSystemSnapshot, 1)
	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Hour},
		fakeReader{data: []byte("cpu  10 20 30 40 5 6 7 8 0 0\ncpu0 1 2 3 4 1 0 0 0 0 0\n")},
		fakeReader{data: []byte("MemTotal: 1000 kB\nMemAvailable: 250 kB\n")},
		output,
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)

	select {
	case <-output:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected emitted snapshot")
	}
}

func loadPollingSnapshotFixture(t *testing.T) metrics.RawSystemSnapshot {
	t.Helper()

	worker := NewPollingWorker(
		config.MetricsConfig{PollInterval: time.Second},
		fakeReader{data: []byte("cpu  10 20 30 40 5 6 7 8 0 0\ncpu0 1 2 3 4 1 0 0 0 0 0\n")},
		fakeReader{data: []byte("MemTotal: 1000 kB\nMemAvailable: 250 kB\n")},
		make(chan metrics.RawSystemSnapshot, 1),
	)

	snapshot, err := worker.poll()
	if err != nil {
		t.Fatalf("poll() error = %v", err)
	}

	return snapshot
}
