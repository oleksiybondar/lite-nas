package workers

import "time"

// NewPollingTimerWorker creates a timer worker configured for polling loops.
//
// The returned tick channel receives one empty struct per interval and emits
// one initial tick immediately on start.
func NewPollingTimerWorker(
	interval time.Duration,
	bufferSize int,
) (TimerWorker, <-chan struct{}, error) {
	pollTicks := make(chan struct{}, bufferSize)
	timerWorker, err := NewTimerWorker(
		TimerConfig{
			Interval:    interval,
			EmitOnStart: true,
		},
		pollTicks,
	)
	if err != nil {
		return TimerWorker{}, nil, err
	}

	return timerWorker, pollTicks, nil
}
