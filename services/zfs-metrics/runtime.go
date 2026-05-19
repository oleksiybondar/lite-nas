package main

import (
	"context"
	"sync"

	"lite-nas/services/zfs-metrics/modules"
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

// snapshotState stores latest available snapshot.
type snapshotState struct {
	mu       sync.RWMutex
	latest   metrics.ZFSSnapshot
	hasValue bool
}

// Set stores the latest snapshot value and marks state as available.
func (s *snapshotState) Set(value metrics.ZFSSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latest = value
	s.hasValue = true
}

// Get returns current snapshot and a flag showing whether value is available.
func (s *snapshotState) Get() (metrics.ZFSSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest, s.hasValue
}

// run boots infra, starts workers, and serves snapshot publication/RPC.
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	state := &snapshotState{}
	channels := modules.NewChannelsModule(1)
	workerModule, err := modules.NewWorkersModule(infra.Config.Metrics, channels)
	if err != nil {
		return err
	}
	if err := registerRPCHandlers(infra.Server, state); err != nil {
		return err
	}
	startWorkers(ctx, workerModule)

	infra.Logger.Info("zfs metrics service started", "config", packagedConfigPath)
	return serveSnapshots(ctx, channels.ZFSSnapshots, channels.PollErrors, state, infra.Client, infra.Logger)
}

// registerRPCHandlers registers snapshot read RPC handlers on messaging server.
func registerRPCHandlers(server messaging.Server, state *snapshotState) error {
	return server.RegisterRPC(zfsmetricscontract.SnapshotRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := state.Get()
		if !ok {
			return zfsmetricscontract.GetSnapshotResponse{Available: false}, nil
		}
		return zfsmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		}, nil
	})
}

// serveSnapshots processes worker outputs and publishes snapshot events.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.ZFSSnapshot,
	pollErrors <-chan error,
	state *snapshotState,
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
			shouldStop := handleSnapshot(ctx, state, client, log, snapshot, ok)
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
	state *snapshotState,
	client messaging.Client,
	log sharedlogger.Logger,
	snapshot metrics.ZFSSnapshot,
	ok bool,
) bool {
	if !ok {
		return true
	}

	state.Set(snapshot)
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
