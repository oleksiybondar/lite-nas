package main

import (
	"context"
	"errors"
	"time"

	servicemodules "lite-nas/services/resources-monitor/modules"
	"lite-nas/services/resources-monitor/processor"
	servicerules "lite-nas/services/resources-monitor/rules"
	sharedcontracts "lite-nas/shared/contracts"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	"lite-nas/shared/eventmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/servicetoken"
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
	eventProcessor := processor.New(loadedRules, stateManager, infra.Client, infra.Logger)

	if err = registerSubscriptions(infra.Server, eventProcessor); err != nil {
		return err
	}

	infra.AuthRefreshTimer.Start(ctx)
	go runAuthRefreshLoop(ctx, infra)

	infra.Logger.Info("resources monitor service started", "config", packagedConfigPath, "rules", len(loadedRules))

	<-ctx.Done()
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
func registerSubscriptions(server messaging.Server, eventProcessor *processor.Processor) error {
	if err := server.Subscribe(systemmetricscontract.SnapshotEventSubject, eventProcessor.HandleEnvelope); err != nil {
		return err
	}

	return server.Subscribe(zfsmetricscontract.SnapshotEventSubject, eventProcessor.HandleEnvelope)
}

// initialEventCounter derives a non-negative startup seed from current time.
func initialEventCounter(now time.Time) uint64 {
	seconds := now.Unix()
	if seconds <= 0 {
		return 0
	}

	return uint64(seconds) % 99_999_999
}
