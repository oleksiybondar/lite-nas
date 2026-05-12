package workers

import (
	"context"
	"time"

	"lite-nas/services/system-metrics/parser"
	"lite-nas/shared/fileio"
	"lite-nas/shared/metrics"
)

// PollingWorker periodically reads system metric sources and emits raw system
// snapshots into an output channel.
//
// The worker is intentionally limited to polling and parsing. It does not
// calculate CPU percentages, maintain history, or publish messages. Those
// responsibilities belong to downstream pipeline stages.
type PollingWorker struct {
	cpuReader fileio.Reader
	memReader fileio.Reader
	cpuParser parser.CPUStatParser
	memParser parser.MemStatParser
	ticks     <-chan struct{}
	output    chan<- metrics.RawSystemSnapshot
}

// NewPollingWorker creates a PollingWorker with the dependencies required for
// periodic metrics polling.
//
// The worker depends on readers for CPU/memory data, a tick input channel, and
// an output channel used to forward raw polling results.
func NewPollingWorker(
	cpuReader fileio.Reader,
	memReader fileio.Reader,
	ticks <-chan struct{},
	output chan<- metrics.RawSystemSnapshot,
) PollingWorker {
	return PollingWorker{
		cpuReader: cpuReader,
		memReader: memReader,
		cpuParser: parser.CPUStatParser{},
		memParser: parser.MemStatParser{},
		ticks:     ticks,
		output:    output,
	}
}

// Start launches the polling worker in a separate goroutine.
//
// The worker runs until the provided context is canceled.
func (w PollingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run executes the polling loop until the provided context is canceled.
//
// A polling cycle is executed for each incoming tick signal.
func (w PollingWorker) run(ctx context.Context) {
	for {
		if !w.waitNextPoll(ctx) {
			return
		}

		w.pollAndSend(ctx)
	}
}

// waitNextPoll blocks until the next poll tick arrives or the context is
// canceled.
//
// It returns true when the next polling cycle should proceed. It returns false
// when the worker should stop.
func (w PollingWorker) waitNextPoll(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-w.ticks:
		if !ok {
			return false
		}
		return true
	}
}

// pollAndSend performs one polling cycle and sends the resulting raw system
// snapshot to the output channel.
//
// If polling or parsing fails, the cycle is skipped. If the context is
// canceled before the send completes, the snapshot is dropped.
func (w PollingWorker) pollAndSend(ctx context.Context) {
	snapshot, err := w.poll()
	if err != nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.output <- snapshot:
	}
}

// poll reads and parses CPU and memory data for one polling cycle.
//
// The returned RawSystemSnapshot contains:
//   - the collection timestamp
//   - raw CPU counters
//   - computed memory values
func (w PollingWorker) poll() (metrics.RawSystemSnapshot, error) {
	cpuData, err := w.cpuReader.Read()
	if err != nil {
		return metrics.RawSystemSnapshot{}, err
	}

	memData, err := w.memReader.Read()
	if err != nil {
		return metrics.RawSystemSnapshot{}, err
	}

	cpuSample, err := w.cpuParser.Parse(string(cpuData))
	if err != nil {
		return metrics.RawSystemSnapshot{}, err
	}

	memSample, err := w.memParser.Parse(string(memData))
	if err != nil {
		return metrics.RawSystemSnapshot{}, err
	}

	return metrics.RawSystemSnapshot{
		Timestamp: time.Now(),
		CPU:       cpuSample,
		Mem:       memSample,
	}, nil
}
