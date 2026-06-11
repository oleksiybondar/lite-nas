package main

import (
	"context"

	"lite-nas/services/zfs-metrics/modules"
	"lite-nas/services/zfs-metrics/state"
	sharedcontracts "lite-nas/shared/contracts"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	packagedConfigPath = "/etc/lite-nas/zfs-metrics.conf"
	serviceName        = sharedcontracts.ServiceZFSMetrics
)

// run boots infra, starts workers, and serves snapshot publication/RPC.
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	store := state.NewHistoryStore(infra.Config.Metrics.HistorySize)
	channels := modules.NewChannelsModule(1)
	workerModule, err := modules.NewWorkersModule(infra.Config.Metrics, channels)
	if err != nil {
		return err
	}
	if err := registerRPCHandlers(infra.Server, store); err != nil {
		return err
	}
	startWorkers(ctx, workerModule)

	infra.Logger.Info("zfs metrics service started", "config", packagedConfigPath)
	return serveSnapshots(ctx, channels.ZFSSnapshots, channels.PollErrors, store, infra.Client, infra.Logger)
}

// registerRPCHandlers registers snapshot read RPC handlers on messaging server.
func registerRPCHandlers(server messaging.Server, store *state.HistoryStore) error {
	if err := server.RegisterRPC(zfsmetricscontract.SnapshotRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := store.Latest()
		if !ok {
			return zfsmetricscontract.GetSnapshotResponse{Available: false}, nil
		}
		return zfsmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		}, nil
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(zfsmetricscontract.HistoryRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		return zfsmetricscontract.GetHistoryResponse{Items: store.List()}, nil
	}); err != nil {
		return err
	}

	return nil
}

// serveSnapshots processes worker outputs and publishes snapshot events.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.ZFSSnapshot,
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

// handleShutdown logs shutdown and returns context terminal error.
func handleShutdown(ctx context.Context, log sharedlogger.Logger) error {
	log.Info("zfs metrics service stopping")
	return ctx.Err()
}

// handlePollError logs one polling error event when error channel is open.
func handlePollError(log sharedlogger.Logger, err error, ok bool) {
	if !ok {
		return
	}
	log.Error("zfs snapshot poll failed", "error", err)
}

// handleSnapshot updates state and publishes snapshot event.
func handleSnapshot(
	ctx context.Context,
	store *state.HistoryStore,
	client messaging.Client,
	log sharedlogger.Logger,
	snapshot metrics.ZFSSnapshot,
	ok bool,
) bool {
	if !ok {
		return true
	}

	store.Add(snapshot)
	event := zfsmetricscontract.SnapshotUpdatedEvent{Snapshot: snapshot}
	if err := client.Publish(ctx, zfsmetricscontract.SnapshotEventSubject, event); err != nil {
		log.Error("failed to publish zfs snapshot", "subject", zfsmetricscontract.SnapshotEventSubject, "error", err)
	}
	return false
}

// startWorkers starts all runtime workers for polling pipeline.
func startWorkers(ctx context.Context, workerModule modules.Workers) {
	workerModule.Timer.Start(ctx)
	workerModule.Polling.Start(ctx)
}
