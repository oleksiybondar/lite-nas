package main

import (
	"context"

	"lite-nas/services/process-metrics/modules"
	processstate "lite-nas/services/process-metrics/state"
	sharedcontracts "lite-nas/shared/contracts"
	processmetricscontract "lite-nas/shared/contracts/processmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	packagedConfigPath = "/etc/lite-nas/process-metrics.conf"
	serviceName        = sharedcontracts.ServiceProcessMetrics
	procRootPath       = "/proc"
)

// run boots infrastructure, starts workers, and serves snapshot publication
// and RPC handlers until shutdown.
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	store := processstate.NewSnapshotStore()
	if err := registerRPCHandlers(infra.Server, store); err != nil {
		return err
	}

	if !infra.Config.ProcessMetrics.Enabled {
		infra.Logger.Info("process metrics service disabled", "config", packagedConfigPath)
		<-ctx.Done()
		infra.Logger.Info("process metrics service stopping")
		return ctx.Err()
	}

	channels := modules.NewChannelsModule(1)
	workerModule, err := modules.NewWorkersModule(
		infra.Config.ProcessMetrics,
		channels,
		modules.SourcePaths{ProcRoot: procRootPath},
	)
	if err != nil {
		return err
	}

	startWorkers(ctx, workerModule)

	infra.Logger.Info("process metrics service started", "config", packagedConfigPath)
	return serveSnapshots(
		ctx,
		channels.ProcessedProcessSnapshots,
		channels.PollErrors,
		store,
		infra.Client,
		infra.Logger,
	)
}

// registerRPCHandlers registers snapshot read RPC handlers on the messaging server.
func registerRPCHandlers(server messaging.Server, store *processstate.SnapshotStore) error {
	return server.RegisterRPC(processmetricscontract.SnapshotRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := store.Latest()
		if !ok {
			return processmetricscontract.GetSnapshotResponse{Available: false}, nil
		}

		return processmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		}, nil
	})
}

// serveSnapshots processes worker outputs and publishes snapshot events.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.ProcessMetricsSnapshot,
	pollErrors <-chan error,
	store *processstate.SnapshotStore,
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
	log.Info("process metrics service stopping")
	return ctx.Err()
}

// handlePollError logs one polling error event when the error channel is open.
func handlePollError(log sharedlogger.Logger, err error, ok bool) {
	if !ok {
		return
	}

	log.Error("process snapshot poll failed", "error", err)
}

// handleSnapshot updates state and publishes a snapshot event.
func handleSnapshot(
	ctx context.Context,
	store *processstate.SnapshotStore,
	client messaging.Client,
	log sharedlogger.Logger,
	snapshot metrics.ProcessMetricsSnapshot,
	ok bool,
) bool {
	if !ok {
		return true
	}

	store.Add(snapshot)
	event := processmetricscontract.SnapshotUpdatedEvent{Snapshot: snapshot}
	if err := client.Publish(ctx, processmetricscontract.SnapshotEventSubject, event); err != nil {
		log.Error(
			"failed to publish process snapshot",
			"subject",
			processmetricscontract.SnapshotEventSubject,
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
	workerModule.Processing.Start(ctx)
}
