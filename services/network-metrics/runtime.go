package main

import (
	"context"

	"lite-nas/services/network-metrics/modules"
	"lite-nas/services/network-metrics/state"
	sharedcontracts "lite-nas/shared/contracts"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	packagedConfigPath = "/etc/lite-nas/network-metrics.conf"
	serviceName        = sharedcontracts.ServiceNetworkMetrics

	procNetDevPath      = "/proc/net/dev"
	sysClassNetPath     = "/sys/class/net"
	procNetSNMPPath     = "/proc/net/snmp"
	procNetNetstatPath  = "/proc/net/netstat"
	procNetTCPPath      = "/proc/net/tcp"
	procNetTCP6Path     = "/proc/net/tcp6"
	procNetUDPPath      = "/proc/net/udp"
	procNetUDP6Path     = "/proc/net/udp6"
	procNetSockstatPath = "/proc/net/sockstat"
	procSoftIRQsPath    = "/proc/softirqs"
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
		ProcNetDev:      procNetDevPath,
		SysClassNet:     sysClassNetPath,
		ProcNetSNMP:     procNetSNMPPath,
		ProcNetNetstat:  procNetNetstatPath,
		ProcNetTCP:      procNetTCPPath,
		ProcNetTCP6:     procNetTCP6Path,
		ProcNetUDP:      procNetUDPPath,
		ProcNetUDP6:     procNetUDP6Path,
		ProcNetSockstat: procNetSockstatPath,
		ProcSoftIRQs:    procSoftIRQsPath,
	})
	if err != nil {
		return err
	}

	if err := registerRPCHandlers(infra.Server, store); err != nil {
		return err
	}

	startWorkers(ctx, workerModule)

	infra.Logger.Info("network metrics service started", "config", packagedConfigPath)
	return serveSnapshots(ctx, channels.NetworkSnapshots, channels.PollErrors, store, infra.Client, infra.Logger)
}

// registerRPCHandlers registers snapshot read RPC handlers on the messaging
// server.
func registerRPCHandlers(server messaging.Server, store *state.HistoryStore) error {
	if err := server.RegisterRPC(networkmetricscontract.SnapshotRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		snapshot, ok := store.Latest()
		if !ok {
			return networkmetricscontract.GetSnapshotResponse{Available: false}, nil
		}

		return networkmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		}, nil
	}); err != nil {
		return err
	}

	if err := server.RegisterRPC(networkmetricscontract.HistoryRPCSubject, func(_ context.Context, _ messaging.Envelope) (any, error) {
		return networkmetricscontract.GetHistoryResponse{Items: store.List()}, nil
	}); err != nil {
		return err
	}

	return nil
}

// serveSnapshots processes worker outputs and publishes snapshot events.
func serveSnapshots(
	ctx context.Context,
	input <-chan metrics.NetworkMetricsSnapshot,
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
	log.Info("network metrics service stopping")
	return ctx.Err()
}

// handlePollError logs one polling error event when the error channel is open.
func handlePollError(log sharedlogger.Logger, err error, ok bool) {
	if !ok {
		return
	}

	log.Error("network snapshot poll failed", "error", err)
}

// handleSnapshot updates state and publishes a snapshot event.
func handleSnapshot(
	ctx context.Context,
	store *state.HistoryStore,
	client messaging.Client,
	log sharedlogger.Logger,
	snapshot metrics.NetworkMetricsSnapshot,
	ok bool,
) bool {
	if !ok {
		return true
	}

	store.Add(snapshot)
	event := networkmetricscontract.SnapshotUpdatedEvent{Snapshot: snapshot}
	if err := client.Publish(ctx, networkmetricscontract.SnapshotEventSubject, event); err != nil {
		log.Error(
			"failed to publish network snapshot",
			"subject",
			networkmetricscontract.SnapshotEventSubject,
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
