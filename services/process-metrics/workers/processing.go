package workers

import (
	"context"

	"lite-nas/shared/metrics"
)

// ProcessingWorker enriches raw process snapshots with computed CPU
// percentages derived from consecutive polling cycles.
type ProcessingWorker struct {
	input    <-chan metrics.ProcessMetricsSnapshot
	output   chan<- metrics.ProcessMetricsSnapshot
	previous *metrics.ProcessMetricsSnapshot
}

// NewProcessingWorker creates a ProcessingWorker with the required input and
// output channels.
func NewProcessingWorker(
	input <-chan metrics.ProcessMetricsSnapshot,
	output chan<- metrics.ProcessMetricsSnapshot,
) ProcessingWorker {
	return ProcessingWorker{
		input:  input,
		output: output,
	}
}

// Start launches the processing worker in a separate goroutine.
func (w *ProcessingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run executes the processing loop until the provided context is canceled or
// the input channel is closed.
func (w *ProcessingWorker) run(ctx context.Context) {
	for {
		snapshot, ok := w.readSnapshot(ctx)
		if !ok {
			return
		}

		w.processAndSend(ctx, snapshot)
	}
}

// readSnapshot waits for the next raw process snapshot or cancellation.
func (w *ProcessingWorker) readSnapshot(ctx context.Context) (metrics.ProcessMetricsSnapshot, bool) {
	select {
	case <-ctx.Done():
		return metrics.ProcessMetricsSnapshot{}, false
	case snapshot, ok := <-w.input:
		if !ok {
			return metrics.ProcessMetricsSnapshot{}, false
		}

		return snapshot, true
	}
}

// processAndSend computes CPU percentages for one process snapshot and sends
// the enriched result to the downstream pipeline.
func (w *ProcessingWorker) processAndSend(
	ctx context.Context,
	snapshot metrics.ProcessMetricsSnapshot,
) {
	processed := buildProcessedProcessSnapshot(w.previous, snapshot)
	w.previous = &snapshot

	select {
	case <-ctx.Done():
		return
	case w.output <- processed:
	}
}

// buildProcessedProcessSnapshot derives per-process CPU percentages from the
// previous and current raw process snapshots.
func buildProcessedProcessSnapshot(
	previous *metrics.ProcessMetricsSnapshot,
	current metrics.ProcessMetricsSnapshot,
) metrics.ProcessMetricsSnapshot {
	processed := current
	processed.Processes = make([]metrics.ProcessMetric, 0, len(current.Processes))

	previousProcessesByPID := indexProcessesByPID(previous)
	systemDeltaTicks := resolveSystemDeltaTicks(previous, current)

	for _, process := range current.Processes {
		process.CPU.CPUPct = calculateProcessCPUPct(
			resolvePreviousProcess(previousProcessesByPID, process),
			process,
			systemDeltaTicks,
		)
		processed.Processes = append(processed.Processes, process)
	}

	return processed
}

// indexProcessesByPID groups the previous snapshot by PID for fast lookup
// during CPU percentage calculation.
func indexProcessesByPID(
	previous *metrics.ProcessMetricsSnapshot,
) map[int]metrics.ProcessMetric {
	if previous == nil {
		return nil
	}

	processesByPID := make(map[int]metrics.ProcessMetric, len(previous.Processes))
	for _, process := range previous.Processes {
		processesByPID[process.PID] = process
	}

	return processesByPID
}

// resolveSystemDeltaTicks returns the host CPU tick delta between two
// consecutive process snapshots.
func resolveSystemDeltaTicks(
	previous *metrics.ProcessMetricsSnapshot,
	current metrics.ProcessMetricsSnapshot,
) uint64 {
	if previous == nil || current.SystemTotalCPUTicks <= previous.SystemTotalCPUTicks {
		return 0
	}

	return current.SystemTotalCPUTicks - previous.SystemTotalCPUTicks
}

// resolvePreviousProcess returns the matching previous process sample when the
// PID and start time identify the same process instance.
func resolvePreviousProcess(
	previousProcessesByPID map[int]metrics.ProcessMetric,
	current metrics.ProcessMetric,
) *metrics.ProcessMetric {
	if previousProcessesByPID == nil {
		return nil
	}

	previous, ok := previousProcessesByPID[current.PID]
	if !ok || !isSameProcessInstance(previous, current) {
		return nil
	}

	return &previous
}

// isSameProcessInstance checks whether two PID samples describe the same live
// process rather than a reused PID.
func isSameProcessInstance(previous metrics.ProcessMetric, current metrics.ProcessMetric) bool {
	if previous.StartTime != nil && current.StartTime != nil {
		return previous.StartTime.Equal(*current.StartTime)
	}

	return previous.PPID == current.PPID && previous.Name == current.Name
}

// calculateProcessCPUPct computes the process share of total host CPU
// capacity across two snapshots.
func calculateProcessCPUPct(
	previous *metrics.ProcessMetric,
	current metrics.ProcessMetric,
	systemDeltaTicks uint64,
) float64 {
	if previous == nil || systemDeltaTicks == 0 || current.CPU.TotalTicks <= previous.CPU.TotalTicks {
		return 0
	}

	processDeltaTicks := current.CPU.TotalTicks - previous.CPU.TotalTicks
	return float64(processDeltaTicks) / float64(systemDeltaTicks) * 100
}
