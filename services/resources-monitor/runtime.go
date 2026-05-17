package main

import (
	"context"
	"time"

	servicemodules "lite-nas/services/resources-monitor/modules"
	"lite-nas/services/resources-monitor/processor"
	servicerules "lite-nas/services/resources-monitor/rules"
	sharedcontracts "lite-nas/shared/contracts"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/eventmanager"
	"lite-nas/shared/messaging"
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

	infra.Logger.Info("resources monitor service started", "config", packagedConfigPath, "rules", len(loadedRules))

	<-ctx.Done()
	infra.Logger.Info("resources monitor service stopping")
	return ctx.Err()
}

// registerSubscriptions wires subject handlers required by resources-monitor.
func registerSubscriptions(server messaging.Server, eventProcessor *processor.Processor) error {
	return server.Subscribe(systemmetricscontract.SnapshotEventSubject, eventProcessor.HandleEnvelope)
}

// initialEventCounter derives a non-negative startup seed from current time.
func initialEventCounter(now time.Time) uint64 {
	seconds := now.Unix()
	if seconds <= 0 {
		return 0
	}

	return uint64(seconds) % 99_999_999
}
