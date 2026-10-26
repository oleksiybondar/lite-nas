package workers

import (
	"context"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: process-metrics-svc/FR-001
func TestProcessingWorkerFirstSnapshotUsesZeroCPUPct(t *testing.T) {
	t.Parallel()

	current := metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: 200,
		Processes: []metrics.ProcessMetric{
			{
				PID: 100,
				CPU: metrics.ProcessCPU{TotalTicks: 25},
			},
		},
	}

	processed := buildProcessedProcessSnapshot(nil, current)

	if got := processed.Processes[0].CPU.CPUPct; got != 0 {
		t.Fatalf("CPUPct = %v, want 0", got)
	}
}

// processSnapshotFixture builds the same process identity with a chosen CPU sample.
func processSnapshotFixture(systemTicks, processTicks uint64, startTime time.Time) metrics.ProcessMetricsSnapshot {
	return metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: systemTicks,
		Processes: []metrics.ProcessMetric{{
			PID:       100,
			PPID:      1,
			Name:      "alpha",
			StartTime: &startTime,
			CPU:       metrics.ProcessCPU{TotalTicks: processTicks},
		}},
	}
}

// Requirements: process-metrics-svc/FR-001
func TestProcessingWorkerCPUPctUsesProcessIdentity(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name             string
		currentStartTime time.Time
		want             float64
	}{
		{name: "consecutive snapshot", currentStartTime: time.Unix(100, 0).UTC(), want: 25},
		{name: "reused PID", currentStartTime: time.Unix(200, 0).UTC(), want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			previous := processSnapshotFixture(1000, 50, time.Unix(100, 0).UTC())
			current := processSnapshotFixture(1100, 75, test.currentStartTime)
			processed := buildProcessedProcessSnapshot(&previous, current)
			if got := processed.Processes[0].CPU.CPUPct; got != test.want {
				t.Fatalf("CPUPct = %v, want %v", got, test.want)
			}
		})
	}
}

// Requirements: process-metrics-svc/FR-001
func TestProcessingWorkerStartEmitsProcessedSnapshot(t *testing.T) {
	t.Parallel()

	input := make(chan metrics.ProcessMetricsSnapshot, 1)
	output := make(chan metrics.ProcessMetricsSnapshot, 1)
	worker := NewProcessingWorker(input, output)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)

	input <- metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: 200,
		Processes: []metrics.ProcessMetric{
			{
				PID: 100,
				CPU: metrics.ProcessCPU{TotalTicks: 25},
			},
		},
	}

	select {
	case snapshot := <-output:
		if got := snapshot.Processes[0].CPU.CPUPct; got != 0 {
			t.Fatalf("CPUPct = %v, want 0", got)
		}
	case <-time.After(time.Second):
		t.Fatal("processed snapshot was not emitted")
	}
}
