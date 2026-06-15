package main

import (
	"context"

	"lite-nas/services/service-metrics/modules"
	servicestate "lite-nas/services/service-metrics/state"
	sharedcontracts "lite-nas/shared/contracts"
	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	packagedConfigPath = "/etc/lite-nas/service-metrics.conf"
	serviceName        = sharedcontracts.ServiceServiceMetrics
	systemSliceDir     = "/sys/fs/cgroup/system.slice"
)

var unitRoots = []string{
	"/etc/systemd/system",
	"/run/systemd/system",
	"/usr/lib/systemd/system",
	"/lib/systemd/system",
}

// run boots infrastructure, starts workers, and serves snapshot publication
// and RPC handlers until shutdown.
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	store := servicestate.NewHistoryStore(infra.Config.Metrics.HistorySize)
	channels := modules.NewChannelsModule(1)
	workerModule, err := modules.NewWorkersModule(infra.Config.Metrics, channels, modules.SourcePaths{
		UnitRoots:      unitRoots,
		SystemSliceDir: systemSliceDir,
	})
	if err != nil {
		return err
	}

	if err := registerRPCHandlers(infra.Server, store); err != nil {
		return err
	}

	startWorkers(ctx, workerModule)

	infra.Logger.Info("service metrics service started", "config", packagedConfigPath)
	return serveSnapshots(ctx, channels.ServiceSnapshots, channels.PollErrors, store, infra.Client, infra.Logger)
}

// registerRPCHandlers registers snapshot read RPC handlers on the messaging server.
func registerRPCHandlers(server messaging.Server, store *servicestate.HistoryStore) error {
	if err := server.RegisterRPC(servicemetricscontract.SnapshotRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := store.Latest()
		if !ok {
			return servicemetricscontract.GetSnapshotResponse{Available: false}, nil
		}

		return servicemetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		}, nil
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(servicemetricscontract.HistoryRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		return servicemetricscontract.GetHistoryResponse{Items: store.List()}, nil
	}); err != nil {
		return err
	}

	return nil
}

// serveSnapshots processes worker outputs and publishes snapshot events.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.ServiceMetricsSnapshot,
	pollErrors <-chan error,
	store *servicestate.HistoryStore,
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
			if handleSnapshot(ctx, store, client, log, snapshot, ok) {
				return nil
			}
		}
	}
}

// handleShutdown logs shutdown and returns the context terminal error.
func handleShutdown(ctx context.Context, log sharedlogger.Logger) error {
	log.Info("service metrics service stopping")
	return ctx.Err()
}

// handlePollError logs one polling error event when the error channel is open.
func handlePollError(log sharedlogger.Logger, err error, ok bool) {
	if !ok {
		return
	}

	log.Error("service snapshot poll failed", "error", err)
}

// handleSnapshot updates state and publishes a snapshot event.
func handleSnapshot(
	ctx context.Context,
	store *servicestate.HistoryStore,
	client messaging.Client,
	log sharedlogger.Logger,
	snapshot metrics.ServiceMetricsSnapshot,
	ok bool,
) bool {
	if !ok {
		return true
	}

	store.Add(snapshot)
	event := servicemetricscontract.SnapshotUpdatedEvent{Snapshot: snapshot}
	if err := client.Publish(ctx, servicemetricscontract.SnapshotEventSubject, event); err != nil {
		log.Error(
			"failed to publish service snapshot",
			"subject",
			servicemetricscontract.SnapshotEventSubject,
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
