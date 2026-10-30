package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	servicemodules "lite-nas/services/resources-monitor/modules"
	"lite-nas/services/resources-monitor/processor"
	servicerules "lite-nas/services/resources-monitor/rules"
	sharedcontracts "lite-nas/shared/contracts"
	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	"lite-nas/shared/eventmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/servicetoken"
	"lite-nas/shared/workers"
)

const (
	packagedConfigPath = "/etc/lite-nas/resources-monitor.conf"
	serviceName        = sharedcontracts.ServiceResourcesMonitor
)

var (
	newInfraModule = servicemodules.NewInfraModule
	loadRules      = servicerules.LoadRules
)

// run constructs runtime dependencies, subscribes to input subjects, and runs
// until shutdown.
func run(ctx context.Context) error {
	return runWithDependencies(ctx, packagedConfigPath, serviceName, newInfraModule, loadRules)
}

// runWithDependencies executes runtime flow with injectable constructors for
// testability.
func runWithDependencies(
	ctx context.Context,
	configPath string,
	appServiceName string,
	infraFactory func(string, string) (servicemodules.Infra, error),
	rulesLoader func([]string) ([]servicerules.Rule, error),
) error {
	infra, err := infraFactory(configPath, appServiceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	loadedRules, err := rulesLoader(infra.Config.Rules.Files)
	if err != nil {
		return err
	}

	initialCounter := initialEventCounter(time.Now())
	stateManager := eventmanager.NewManager(initialCounter)
	// The NATS callback produces incoming snapshots; the processor owns effects.
	// Shutdown drains subscriptions before closing incoming, then waits for both
	// workers to finish before closing the messaging client.
	incoming := make(chan snapshotInput, 32)
	effects := make(chan alertIntent, 32)
	eventProcessor := processor.NewWithSender(loadedRules, stateManager, channelAlertSender{effects: effects}, infra.Logger)

	if err = registerSubscriptions(infra.Server, incoming, ctx); err != nil {
		return err
	}
	processorDone := make(chan struct{})
	reporterDone := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(reporterDone)
		for effect := range effects {
			deliverAlertIntent(workerCtx, infra.Client, effect)
		}
	}).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(processorDone)
		defer close(effects)
		for event := range incoming {
			eventProcessor.ProcessSnapshot(workerCtx, event.subject, event.payload)
		}
	}).Start(ctx)

	infra.AuthRefreshTimer.Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) { runAuthRefreshLoop(workerCtx, infra) }).Start(ctx)

	infra.Logger.Info("resources monitor service started", "config", packagedConfigPath, "rules", len(loadedRules))

	<-ctx.Done()
	if err := infra.Server.Drain(); err != nil {
		infra.Logger.Error("resources monitor subscription drain failed", "error", err)
	}
	close(incoming)
	<-processorDone
	<-reporterDone
	infra.Logger.Info("resources monitor service stopping")
	return ctx.Err()
}

func runAuthRefreshLoop(ctx context.Context, infra servicemodules.Infra) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-infra.AuthRefreshTicks:
			handleAuthRefreshTick(ctx, infra)
		}
	}
}

func handleAuthRefreshTick(ctx context.Context, infra servicemodules.Infra) {
	refreshErr := infra.AuthTokenManager.Refresh(ctx)
	if refreshErr == nil {
		infra.Logger.Debug("resources monitor auth token refreshed")
		return
	}

	if !errors.Is(refreshErr, servicetoken.ErrTokenUnavailable) {
		infra.Logger.Warn("resources monitor auth token refresh failed", "error", refreshErr)
	}

	if err := infra.AuthTokenManager.Login(ctx); err != nil {
		infra.Logger.Warn("resources monitor auth token login failed", "error", err)
		return
	}

	infra.Logger.Info("resources monitor auth token login succeeded")
}

// registerSubscriptions wires subject handlers required by resources-monitor.
func registerSubscriptions(server messaging.Server, input chan<- snapshotInput, serviceCtx context.Context) error {
	handler := func(ctx context.Context, envelope messaging.Envelope) error {
		var payload map[string]any
		if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
			return err
		}
		select {
		case input <- snapshotInput{subject: envelope.Subject, payload: payload}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-serviceCtx.Done():
			return serviceCtx.Err()
		}
	}
	if err := server.Subscribe(networkmetricscontract.SnapshotEventSubject, handler); err != nil {
		return err
	}

	if err := server.Subscribe(systemmetricscontract.SnapshotEventSubject, handler); err != nil {
		return err
	}

	if err := server.Subscribe(diskmetricscontract.SnapshotEventSubject, handler); err != nil {
		return err
	}

	return server.Subscribe(zfsmetricscontract.SnapshotEventSubject, handler)
}

// initialEventCounter derives a non-negative startup seed from current time.
func initialEventCounter(now time.Time) uint64 {
	seconds := now.Unix()
	if seconds <= 0 {
		return 0
	}

	return uint64(seconds) % 99_999_999
}
