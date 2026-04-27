package main

import (
	"context"

	"lite-nas/services/system-metrics/modules"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	packagedConfigPath = "/etc/lite-nas/system-metrics.conf"
	procStatPath       = "/proc/stat"
	procMemInfoPath    = "/proc/meminfo"

	statsEventSubject = "system.metrics.events.stats"
	statsRPCSubject   = "system.metrics.rpc.stats.get"
	historyRPCSubject = "system.metrics.rpc.history.get"
	serviceName       = "system-metrics"
)

// run assembles the system-metrics runtime, registers RPC handlers, starts the
// workers, and serves processed snapshots until shutdown.
//
// Parameters:
//   - ctx: process-lifetime context cancelled by OS signal handling
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	channels := modules.NewChannelsModule(infra.Config.Metrics.HistorySize)
	ioModule, err := modules.NewIOModule(procStatPath, procMemInfoPath)
	if err != nil {
		return err
	}

	workerModule := modules.NewWorkersModule(
		infra.Config.Metrics,
		channels,
		ioModule,
	)
	stateModule := modules.NewStateModule(infra.Config.Metrics.HistorySize)
	if err := registerRPCHandlers(infra.Server, stateModule.SnapshotStore); err != nil {
		return err
	}

	startWorkers(ctx, workerModule)

	infra.Logger.Info("system metrics service started", "config", packagedConfigPath)

	return serveSnapshots(
		ctx,
		channels.SystemSnapshots,
		stateModule.SnapshotStore,
		infra.Client,
		infra.Logger,
	)
}

// startWorkers starts the polling and processing workers for the service
// runtime.
func startWorkers(ctx context.Context, workerModule modules.Workers) {
	pollingWorker := workerModule.Polling
	processingWorker := workerModule.Processing

	pollingWorker.Start(ctx)
	processingWorker.Start(ctx)
}

// registerRPCHandlers registers the runtime RPC handlers for latest snapshot
// and snapshot history queries.
func registerRPCHandlers(server messaging.Server, store *modules.SnapshotStore) error {
	if err := server.RegisterRPC(statsRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := store.Latest()
		if !ok {
			return map[string]any{}, nil
		}

		return snapshot, nil
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(historyRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		return store.List(), nil
	}); err != nil {
		return err
	}

	return nil
}

// serveSnapshots stores and publishes processed snapshots until the context is
// canceled or the input channel is closed.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.SystemSnapshot,
	store *modules.SnapshotStore,
	client messaging.Client,
	log sharedlogger.Logger,
) error {
	for {
		select {
		case <-ctx.Done():
			return gracefulExit(ctx, log)
		case snapshot, ok := <-input:
			if !ok {
				return nil
			}

			storeAndPublishSnapshot(ctx, snapshot, store, client, log)
		}
	}
}

// gracefulExit logs service shutdown and returns the terminal context error.
func gracefulExit(ctx context.Context, log sharedlogger.Logger) error {
	log.Info("system metrics service stopping")
	return ctx.Err()
}

// storeAndPublishSnapshot stores the latest processed snapshot and publishes it
// as the current stats event.
func storeAndPublishSnapshot(
	ctx context.Context,
	snapshot metrics.SystemSnapshot,
	store *modules.SnapshotStore,
	client messaging.Client,
	log sharedlogger.Logger,
) {
	store.Add(snapshot)
	if err := client.Publish(ctx, statsEventSubject, snapshot); err != nil {
		log.Warn("failed to publish system metrics snapshot", "subject", statsEventSubject, "error", err)
	}
}
