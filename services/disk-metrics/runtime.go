package main

import (
	"context"

	"lite-nas/services/disk-metrics/modules"
	"lite-nas/services/disk-metrics/state"
	sharedcontracts "lite-nas/shared/contracts"
	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	packagedConfigPath    = "/etc/lite-nas/disk-metrics.conf"
	serviceName           = sharedcontracts.ServiceDiskMetrics
	sysBlockPath          = "/sys/block"
	procDiskStatsPath     = "/proc/diskstats"
	procSelfMountInfoPath = "/proc/self/mountinfo"
	procMountsPath        = "/proc/mounts"
)

// run boots infrastructure, starts workers, and serves snapshot publication
// and RPC handlers until shutdown.
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	store := state.NewHistoryStore(infra.Config.Metrics.HistorySize)
	channels := modules.NewChannelsModule(1)
	workerModule, err := modules.NewWorkersModule(infra.Config.Metrics, channels, modules.SourcePaths{
		SysBlock:          sysBlockPath,
		ProcDiskStats:     procDiskStatsPath,
		ProcSelfMountInfo: procSelfMountInfoPath,
		ProcMounts:        procMountsPath,
	})
	if err != nil {
		return err
	}

	if err := registerRPCHandlers(infra.Server, store); err != nil {
		return err
	}

	startWorkers(ctx, workerModule)

	infra.Logger.Info("disk metrics service started", "config", packagedConfigPath)
	return serveSnapshots(ctx, channels.DiskSnapshots, channels.PollErrors, store, infra.Client, infra.Logger)
}

// registerRPCHandlers registers snapshot read RPC handlers on the messaging
// server.
func registerRPCHandlers(server messaging.Server, store *state.HistoryStore) error {
	if err := server.RegisterRPC(diskmetricscontract.SnapshotRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := store.Latest()
		if !ok {
			return diskmetricscontract.GetSnapshotResponse{Available: false}, nil
		}

		return diskmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		}, nil
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(diskmetricscontract.HistoryRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		return diskmetricscontract.GetHistoryResponse{Items: store.List()}, nil
	}); err != nil {
		return err
	}

	return nil
}

// serveSnapshots processes worker outputs and publishes snapshot events.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.DiskMetricsSnapshot,
	pollErrors <-chan error,
	store *state.HistoryStore,
	client messaging.Client,
	log sharedlogger.Logger,
) error {
	for {
		select {
		case <-ctx.Done():
			return handleShutdown(ctx, log)
		case err, ok := <-pollErrors:
			handlePollError(log, err, ok)
		case snapshot, ok := <-input:
			shouldStop := handleSnapshot(ctx, store, client, log, snapshot, ok)
			if shouldStop {
				return nil
			}
		}
	}
}

// handleShutdown logs shutdown and returns the context terminal error.
func handleShutdown(ctx context.Context, log sharedlogger.Logger) error {
	log.Info("disk metrics service stopping")
	return ctx.Err()
}

// handlePollError logs one polling error event when the error channel is open.
func handlePollError(log sharedlogger.Logger, err error, ok bool) {
	if !ok {
		return
	}

	log.Error("disk snapshot poll failed", "error", err)
}

// handleSnapshot updates state and publishes a snapshot event.
func handleSnapshot(
	ctx context.Context,
	store *state.HistoryStore,
	client messaging.Client,
	log sharedlogger.Logger,
	snapshot metrics.DiskMetricsSnapshot,
	ok bool,
) bool {
	if !ok {
		return true
	}

	store.Add(snapshot)
	event := diskmetricscontract.SnapshotUpdatedEvent{Snapshot: snapshot}
	if err := client.Publish(ctx, diskmetricscontract.SnapshotEventSubject, event); err != nil {
		log.Error(
			"failed to publish disk snapshot",
			"subject",
			diskmetricscontract.SnapshotEventSubject,
			"error",
			err,
		)
	}

	return false
}

// startWorkers starts all runtime workers for the polling pipeline.
func startWorkers(ctx context.Context, workerModule modules.Workers) {
	workerModule.Timer.Start(ctx)
	workerModule.Polling.Start(ctx)
}
