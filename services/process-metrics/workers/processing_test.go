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

// Requirements: process-metrics-svc/FR-001
func TestProcessingWorkerCalculatesCPUPctFromConsecutiveSnapshots(t *testing.T) {
	t.Parallel()

	previousStartTime := time.Unix(100, 0).UTC()
	currentStartTime := previousStartTime
	previous := &metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: 1000,
		Processes: []metrics.ProcessMetric{
			{
				PID:       100,
				PPID:      1,
				Name:      "alpha",
				StartTime: &previousStartTime,
				CPU: metrics.ProcessCPU{
					TotalTicks: 50,
				},
			},
		},
	}
	current := metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: 1100,
		Processes: []metrics.ProcessMetric{
			{
				PID:       100,
				PPID:      1,
				Name:      "alpha",
				StartTime: &currentStartTime,
				CPU: metrics.ProcessCPU{
					TotalTicks: 75,
				},
			},
		},
	}

	processed := buildProcessedProcessSnapshot(previous, current)

	if got := processed.Processes[0].CPU.CPUPct; got != 25 {
		t.Fatalf("CPUPct = %v, want 25", got)
	}
}

// Requirements: process-metrics-svc/FR-001
func TestProcessingWorkerResetsCPUPctWhenPIDWasReused(t *testing.T) {
	t.Parallel()

	previousStartTime := time.Unix(100, 0).UTC()
	currentStartTime := time.Unix(200, 0).UTC()
	previous := &metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: 1000,
		Processes: []metrics.ProcessMetric{
			{
				PID:       100,
				PPID:      1,
				Name:      "alpha",
				StartTime: &previousStartTime,
				CPU: metrics.ProcessCPU{
					TotalTicks: 50,
				},
			},
		},
	}
	current := metrics.ProcessMetricsSnapshot{
		SystemTotalCPUTicks: 1100,
		Processes: []metrics.ProcessMetric{
			{
				PID:       100,
				PPID:      1,
				Name:      "alpha",
				StartTime: &currentStartTime,
				CPU: metrics.ProcessCPU{
					TotalTicks: 75,
				},
			},
		},
	}

	processed := buildProcessedProcessSnapshot(previous, current)

	if got := processed.Processes[0].CPU.CPUPct; got != 0 {
		t.Fatalf("CPUPct = %v, want 0", got)
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
