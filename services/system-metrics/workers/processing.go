package workers

import (
	"context"

	serviceerrors "lite-nas/services/system-metrics/errors"
	"lite-nas/shared/metrics"
)

// ProcessingWorker consumes raw system snapshots and produces computed system
// snapshots.
//
// CPU usage requires two raw CPU samples, so the worker keeps the previous raw
// snapshot in memory. The first raw snapshot is stored but does not produce any
// output. Starting from the second raw snapshot, the worker computes CPU usage
// percentages and emits a complete SystemSnapshot.
type ProcessingWorker struct {
	input    <-chan metrics.RawSystemSnapshot
	output   chan<- metrics.SystemSnapshot
	previous *metrics.RawSystemSnapshot
}

// NewProcessingWorker creates a ProcessingWorker with the required input and
// output channels.
func NewProcessingWorker(
	input <-chan metrics.RawSystemSnapshot,
	output chan<- metrics.SystemSnapshot,
) ProcessingWorker {
	return ProcessingWorker{
		input:  input,
		output: output,
	}
}

// Start launches the processing worker in a separate goroutine.
//
// The worker runs until the provided context is canceled or the input channel
// is closed.
func (w *ProcessingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run executes the processing loop until the provided context is canceled or
// the input channel is closed.
//
// The first raw snapshot is stored as a baseline state and does not produce a
// SystemSnapshot. Each subsequent raw snapshot is compared with the previous
// one to compute CPU usage percentages.
func (w *ProcessingWorker) run(ctx context.Context) {
	for {
		rawSnapshot, ok := w.readRawSnapshot(ctx)
		if !ok {
			return
		}

		w.processAndSend(ctx, rawSnapshot)
	}
}

// readRawSnapshot waits for the next raw system snapshot or for worker
// cancellation.
//
// It returns the received snapshot and true when processing should continue.
// It returns false when the worker should stop.
func (w *ProcessingWorker) readRawSnapshot(
	ctx context.Context,
) (metrics.RawSystemSnapshot, bool) {
	select {
	case <-ctx.Done():
		return metrics.RawSystemSnapshot{}, false
	case rawSnapshot, ok := <-w.input:
		if !ok {
			return metrics.RawSystemSnapshot{}, false
		}

		return rawSnapshot, true
	}
}

// processAndSend processes one raw system snapshot and sends the resulting
// SystemSnapshot to the output channel when enough data is available.
//
// If this is the first raw snapshot, it is stored as a baseline state and no
// output is sent.
//
// If CPU usage cannot be computed for the current pair of snapshots, the
// current raw snapshot replaces the previous baseline and the cycle is skipped.
func (w *ProcessingWorker) processAndSend(
	ctx context.Context,
	rawSnapshot metrics.RawSystemSnapshot,
) {
	if w.previous == nil {
		w.previous = &rawSnapshot
		return
	}

	systemSnapshot, err := buildSystemSnapshot(*w.previous, rawSnapshot)
	w.previous = &rawSnapshot

	if err != nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.output <- systemSnapshot:
	}
}

// buildSystemSnapshot builds a computed SystemSnapshot from two consecutive raw
// system snapshots.
//
// CPU values are computed from the difference between the previous and current
// raw CPU counters. Memory values and timestamp are taken from the current raw
// snapshot.
func buildSystemSnapshot(
	previous metrics.RawSystemSnapshot,
	current metrics.RawSystemSnapshot,
) (metrics.SystemSnapshot, error) {
	cpuSample, err := buildCPUSample(previous.CPU, current.CPU)
	if err != nil {
		return metrics.SystemSnapshot{}, err
	}

	return metrics.SystemSnapshot{
		Timestamp: current.Timestamp,
		CPU:       cpuSample,
		Mem:       current.Mem,
	}, nil
}

// buildCPUSample computes total and per-core CPU usage percentages from two raw
// CPU samples.
func buildCPUSample(
	previous metrics.CPURawSample,
	current metrics.CPURawSample,
) (metrics.CPUSample, error) {
	totalUsagePct, err := calculateCPUUsagePct(previous.Total, current.Total)
	if err != nil {
		return metrics.CPUSample{}, err
	}

	perCoreUsage, err := buildPerCoreUsage(previous.Cores, current.Cores)
	if err != nil {
		return metrics.CPUSample{}, err
	}

	return metrics.CPUSample{
		TotalUsagePct: totalUsagePct,
		PerCoreUsage:  perCoreUsage,
	}, nil
}

// buildPerCoreUsage computes usage percentages for CPU cores present in both
// the previous and current raw CPU samples.
func buildPerCoreUsage(
	previous []metrics.CPUCoreRawSample,
	current []metrics.CPUCoreRawSample,
) ([]float64, error) {
	coreCount := min(len(previous), len(current))
	perCoreUsage := make([]float64, 0, coreCount)

	for i := 0; i < coreCount; i++ {
		usagePct, err := calculateCPUUsagePct(previous[i], current[i])
		if err != nil {
			return nil, err
		}

		perCoreUsage = append(perCoreUsage, usagePct)
	}

	return perCoreUsage, nil
}

// calculateCPUUsagePct computes CPU usage percentage from two raw CPU counter
// samples.
//
// Usage is calculated as:
//
//	(totalDelta - idleDelta) / totalDelta * 100
//
// An error is returned if the total counter does not increase between samples.
func calculateCPUUsagePct(
	previous metrics.CPUCoreRawSample,
	current metrics.CPUCoreRawSample,
) (float64, error) {
	totalDelta := current.Total - previous.Total
	idleDelta := current.Idle - previous.Idle

	if totalDelta == 0 {
		return 0, serviceerrors.ErrInvalidCPUDelta
	}

	busyDelta := totalDelta - idleDelta
	return float64(busyDelta) / float64(totalDelta) * 100, nil
}
