package workers

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	serviceerrors "lite-nas/services/system-metrics/errors"
	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-svc/FR-001
func TestBuildSystemSnapshotFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(metrics.SystemSnapshot) any
		want any
	}{
		{name: "total cpu usage", got: func(snapshot metrics.SystemSnapshot) any { return snapshot.CPU.TotalUsagePct }, want: 75.0},
		{name: "per core usage count", got: func(snapshot metrics.SystemSnapshot) any { return len(snapshot.CPU.PerCoreUsage) }, want: 1},
		{name: "timestamp", got: func(snapshot metrics.SystemSnapshot) any { return snapshot.Timestamp }, want: time.Unix(200, 0)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			snapshot := loadSystemSnapshotFixture(t)
			if got := testCase.got(snapshot); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

// Requirements: system-metrics-svc/FR-001
func TestBuildSystemSnapshotPerCoreUsageValue(t *testing.T) {
	t.Parallel()

	snapshot := loadSystemSnapshotFixture(t)
	if got := snapshot.CPU.PerCoreUsage[0]; got != 83.33333333333334 {
		t.Fatalf("PerCoreUsage[0] = %v, want 83.33333333333334", got)
	}
}

// Requirements: system-metrics-svc/FR-001
func TestBuildSystemSnapshotMemPayload(t *testing.T) {
	t.Parallel()

	snapshot := loadSystemSnapshotFixture(t)
	if got := snapshot.Mem.UsedBytes; got != uint64(5) {
		t.Fatalf("Mem.UsedBytes = %d, want 5", got)
	}
}

func TestCalculateCPUUsagePctRejectsZeroDelta(t *testing.T) {
	t.Parallel()

	_, err := calculateCPUUsagePct(
		metrics.CPUCoreRawSample{Total: 10, Idle: 2},
		metrics.CPUCoreRawSample{Total: 10, Idle: 2},
	)
	if !errors.Is(err, serviceerrors.ErrInvalidCPUDelta) {
		t.Fatalf("calculateCPUUsagePct() error = %v, want %v", err, serviceerrors.ErrInvalidCPUDelta)
	}
}

// Requirements: system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestProcessingWorkerProcessAndSendDoesNotEmitBaseline(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.SystemSnapshot, 1)
	worker := NewProcessingWorker(make(chan metrics.RawSystemSnapshot), output)

	first, _ := loadProcessingSnapshotsFixture()

	worker.processAndSend(context.Background(), first)

	select {
	case <-output:
		t.Fatal("did not expect output from baseline snapshot")
	default:
	}
}

// Requirements: system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestProcessingWorkerProcessAndSendEmitsAfterBaseline(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.SystemSnapshot, 1)
	worker := NewProcessingWorker(make(chan metrics.RawSystemSnapshot), output)

	first, second := loadProcessingSnapshotsFixture()

	worker.processAndSend(context.Background(), first)
	worker.processAndSend(context.Background(), second)

	select {
	case <-output:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected processed snapshot")
	}
}

func TestProcessingWorkerReadRawSnapshotStopsOnClosedInput(t *testing.T) {
	t.Parallel()

	input := make(chan metrics.RawSystemSnapshot)
	close(input)

	worker := NewProcessingWorker(input, make(chan metrics.SystemSnapshot, 1))
	_, ok := worker.readRawSnapshot(context.Background())
	if ok {
		t.Fatal("expected readRawSnapshot() to stop on closed input")
	}
}

func TestProcessingWorkerReadRawSnapshotStopsOnContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker := NewProcessingWorker(make(chan metrics.RawSystemSnapshot), make(chan metrics.SystemSnapshot, 1))
	_, ok := worker.readRawSnapshot(ctx)
	if ok {
		t.Fatal("expected readRawSnapshot() to stop on canceled context")
	}
}

// Requirements: system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestProcessingWorkerProcessAndSendDoesNotEmitOnCanceledContext(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.SystemSnapshot)
	worker := NewProcessingWorker(make(chan metrics.RawSystemSnapshot), output)
	first, second := loadProcessingSnapshotsFixture()
	worker.previous = &first

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker.processAndSend(ctx, second)

	select {
	case <-output:
		t.Fatal("did not expect output on canceled context")
	default:
	}
}

// Requirements: system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestProcessingWorkerStartEmitsProcessedSnapshot(t *testing.T) {
	t.Parallel()

	input := make(chan metrics.RawSystemSnapshot, 2)
	output := make(chan metrics.SystemSnapshot, 1)
	worker := NewProcessingWorker(input, output)
	first, second := loadProcessingSnapshotsFixture()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)
	input <- first
	input <- second

	select {
	case <-output:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected processed snapshot")
	}
}

func TestBuildSystemSnapshotReturnsCPUError(t *testing.T) {
	t.Parallel()

	_, err := buildSystemSnapshot(
		metrics.RawSystemSnapshot{
			CPU: metrics.CPURawSample{Total: metrics.CPUCoreRawSample{Total: 10, Idle: 5}},
		},
		metrics.RawSystemSnapshot{
			CPU: metrics.CPURawSample{Total: metrics.CPUCoreRawSample{Total: 10, Idle: 5}},
		},
	)
	if !errors.Is(err, serviceerrors.ErrInvalidCPUDelta) {
		t.Fatalf("buildSystemSnapshot() error = %v, want %v", err, serviceerrors.ErrInvalidCPUDelta)
	}
}

func TestBuildCPUSampleReturnsPerCoreError(t *testing.T) {
	t.Parallel()

	_, err := buildCPUSample(
		metrics.CPURawSample{
			Total: metrics.CPUCoreRawSample{Total: 10, Idle: 5},
			Cores: []metrics.CPUCoreRawSample{{Total: 10, Idle: 5}},
		},
		metrics.CPURawSample{
			Total: metrics.CPUCoreRawSample{Total: 20, Idle: 10},
			Cores: []metrics.CPUCoreRawSample{{Total: 10, Idle: 5}},
		},
	)
	if !errors.Is(err, serviceerrors.ErrInvalidCPUDelta) {
		t.Fatalf("buildCPUSample() error = %v, want %v", err, serviceerrors.ErrInvalidCPUDelta)
	}
}

func TestBuildPerCoreUsageReturnsMinimumCoreCount(t *testing.T) {
	t.Parallel()

	usage, err := buildPerCoreUsage(
		[]metrics.CPUCoreRawSample{
			{Total: 10, Idle: 2},
			{Total: 10, Idle: 3},
		},
		[]metrics.CPUCoreRawSample{
			{Total: 20, Idle: 4},
		},
	)
	if err != nil {
		t.Fatalf("buildPerCoreUsage() error = %v", err)
	}

	want := []float64{80}
	if !reflect.DeepEqual(usage, want) {
		t.Fatalf("buildPerCoreUsage() = %#v, want %#v", usage, want)
	}
}

func loadSystemSnapshotFixture(t *testing.T) metrics.SystemSnapshot {
	t.Helper()

	previous := metrics.RawSystemSnapshot{
		CPU: metrics.CPURawSample{
			Total: metrics.CPUCoreRawSample{Total: 100, Idle: 40},
			Cores: []metrics.CPUCoreRawSample{
				{Total: 40, Idle: 20},
			},
		},
	}
	current := metrics.RawSystemSnapshot{
		Timestamp: time.Unix(200, 0),
		CPU: metrics.CPURawSample{
			Total: metrics.CPUCoreRawSample{Total: 160, Idle: 55},
			Cores: []metrics.CPUCoreRawSample{
				{Total: 70, Idle: 25},
			},
		},
		Mem: metrics.MemSample{UsedBytes: 5},
	}

	snapshot, err := buildSystemSnapshot(previous, current)
	if err != nil {
		t.Fatalf("buildSystemSnapshot() error = %v", err)
	}

	return snapshot
}

func loadProcessingSnapshotsFixture() (metrics.RawSystemSnapshot, metrics.RawSystemSnapshot) {
	return metrics.RawSystemSnapshot{
			CPU: metrics.CPURawSample{Total: metrics.CPUCoreRawSample{Total: 10, Idle: 5}},
		},
		metrics.RawSystemSnapshot{
			CPU: metrics.CPURawSample{Total: metrics.CPUCoreRawSample{Total: 20, Idle: 10}},
		}
}
