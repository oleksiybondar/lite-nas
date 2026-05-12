package main

import (
	"context"

	"lite-nas/services/system-logging-manager/modules"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedworkers "lite-nas/shared/workers"
)

const (
	packagedConfigPath = "/etc/lite-nas/system-logging-manager.conf"
	serviceName        = "system-logging-manager"
)

// run assembles and starts the system logging-manager runtime.
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(ctx, packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	subjects := buildNATSSubjects()
	if err := sharedloggingmanagernats.RegisterSubscriptions(infra.Server, infra.LoggingManagerCore.Core, subjects); err != nil {
		return err
	}
	if err := sharedloggingmanagernats.RegisterRPCHandlers(infra.Server, infra.LoggingManagerCore.Core, subjects); err != nil {
		return err
	}

	flushTimer, err := newFlushTimerWorker(infra)
	if err != nil {
		return err
	}

	startRuntimeWorkers(ctx, infra, flushTimer)

	infra.Logger.Info("system logging manager service started", "config", packagedConfigPath)
	<-ctx.Done()
	infra.Logger.Info("system logging manager service stopping")
	return ctx.Err()
}

func buildNATSSubjects() sharedloggingmanagernats.Subjects {
	return sharedloggingmanagernats.Subjects{
		AlertSubject:                            systemloggingmanagercontract.AlertSubject,
		AlertOccurrenceSubject:                  systemloggingmanagercontract.AlertOccurrenceSubject,
		GetAlertsRPCSubject:                     systemloggingmanagercontract.GetAlertsRPCSubject,
		GetActiveAlertsRPCSubject:               systemloggingmanagercontract.GetActiveAlertsRPCSubject,
		GetUnacknowledgedActiveAlertsRPCSubject: systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		UpdateAlertStateRPCSubject:              systemloggingmanagercontract.UpdateAlertStateRPCSubject,
		AcknowledgeAlertRPCSubject:              systemloggingmanagercontract.AcknowledgeAlertRPCSubject,
		MuteAlertRPCSubject:                     systemloggingmanagercontract.MuteAlertRPCSubject,
	}
}

func newFlushTimerWorker(infra modules.Infra) (sharedworkers.TimerWorker, error) {
	return sharedworkers.NewTimerWorker(
		sharedworkers.TimerConfig{
			Interval:    infra.Config.LoggingManager.Writer.FlushInterval,
			EmitOnStart: false,
		},
		infra.LoggingManagerCore.WriterFlushCh,
	)
}

func startRuntimeWorkers(ctx context.Context, infra modules.Infra, flushTimer sharedworkers.TimerWorker) {
	infra.LoggingManagerCore.CleanupTimer.Start(ctx)
	flushTimer.Start(ctx)
	go runCleanupWorker(ctx, infra.LoggingManagerCore.Core, infra.LoggingManagerCore.CleanupTicksCh, infra.Logger)
	go runWriterWorker(ctx, infra.LoggingManagerCore.Writer, infra.Logger)
}

func runCleanupWorker(
	ctx context.Context,
	core *sharedloggingmanager.Core,
	ticks <-chan struct{},
	log interface {
		Warn(msg string, args ...any)
	},
) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			if err := core.Cleanup(ctx); err != nil {
				log.Warn("logging manager cleanup failed", "error", err)
			}
		}
	}
}

func runWriterWorker(
	ctx context.Context,
	writer *sharedloggingmanager.Writer,
	log interface {
		Error(msg string, args ...any)
	},
) {
	if err := writer.Run(ctx); err != nil {
		log.Error("logging manager writer stopped with error", "error", err.Error())
	}
}
