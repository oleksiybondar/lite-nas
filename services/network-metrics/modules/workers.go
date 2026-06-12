package modules

import (
	serviceconfig "lite-nas/services/network-metrics/config"
	"lite-nas/services/network-metrics/workers"
	sharedworkers "lite-nas/shared/workers"
)

// SourcePaths groups filesystem paths used by the polling worker.
type SourcePaths struct {
	ProcNetDev      string
	SysClassNet     string
	ProcNetSNMP     string
	ProcNetNetstat  string
	ProcNetTCP      string
	ProcNetTCP6     string
	ProcNetUDP      string
	ProcNetUDP6     string
	ProcNetSockstat string
	ProcSoftIRQs    string
}

// Workers groups worker instances used by the service runtime.
type Workers struct {
	Timer   sharedworkers.TimerWorker
	Polling workers.PollingWorker
}

// NewWorkersModule assembles workers required by the network-metrics runtime.
func NewWorkersModule(
	cfg serviceconfig.MetricsConfig,
	channels Channels,
	paths SourcePaths,
) (Workers, error) {
	timerWorker, pollTicks, err := sharedworkers.NewPollingTimerWorker(cfg.PollInterval, 1)
	if err != nil {
		return Workers{}, err
	}

	return Workers{
		Timer: sharedworkers.TimerWorker(timerWorker),
		Polling: workers.NewPollingWorker(
			paths.ProcNetDev,
			paths.SysClassNet,
			paths.ProcNetSNMP,
			paths.ProcNetNetstat,
			paths.ProcNetTCP,
			paths.ProcNetTCP6,
			paths.ProcNetUDP,
			paths.ProcNetUDP6,
			paths.ProcNetSockstat,
			paths.ProcSoftIRQs,
			pollTicks,
			channels.NetworkSnapshots,
			channels.PollErrors,
		),
	}, nil
}
