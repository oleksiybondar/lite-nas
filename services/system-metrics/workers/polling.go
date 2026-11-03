package workers

import (
	"context"
	"time"

	"lite-nas/services/system-metrics/parser"
	"lite-nas/shared/fileio"
	"lite-nas/shared/metrics"
	sharedworkers "lite-nas/shared/workers"
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

// sharedWorker binds service collection to the common polling lifecycle.
func (w PollingWorker) sharedWorker() sharedworkers.PollingWorker[metrics.RawSystemSnapshot] {
	return sharedworkers.NewPollingWorker(w.ticks, w.output, nil, func(_ context.Context) (metrics.RawSystemSnapshot, error) { return w.poll() })
}

// Start launches the polling worker in a separate goroutine.
func (w PollingWorker) Start(ctx context.Context) {
	w.sharedWorker().Start(ctx)
}

// waitNextPoll blocks until the next tick, cancellation, or channel closure.
func (w PollingWorker) waitNextPoll(ctx context.Context) bool {
	return w.sharedWorker().WaitNextPoll(ctx)
}

// pollAndSend runs one service collection cycle and forwards its result.
func (w PollingWorker) pollAndSend(ctx context.Context) {
	w.sharedWorker().PollAndSend(ctx)
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
