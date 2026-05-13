package loggingmanagerservice

import (
	"context"

	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedworkers "lite-nas/shared/workers"
)

// Run starts and blocks a logging-manager service runtime.
func Run(
	ctx context.Context,
	infra Infra,
	subjects sharedloggingmanagernats.Subjects,
	configPath string,
	logName string,
) error {
	if err := sharedloggingmanagernats.RegisterSubscriptions(infra.Server, infra.LoggingManagerCore.Core, subjects); err != nil {
		return err
	}
	if err := sharedloggingmanagernats.RegisterRPCHandlers(infra.Server, infra.LoggingManagerCore.Core, subjects); err != nil {
		return err
	}

	flushTimer, err := sharedworkers.NewTimerWorker(
		sharedworkers.TimerConfig{
			Interval:    infra.Config.LoggingManager.Writer.FlushInterval,
			EmitOnStart: false,
		},
		infra.LoggingManagerCore.WriterFlushCh,
	)
	if err != nil {
		return err
	}

	infra.LoggingManagerCore.CleanupTimer.Start(ctx)
	flushTimer.Start(ctx)
	go runCleanupWorker(ctx, infra.LoggingManagerCore.Core, infra.LoggingManagerCore.CleanupTicksCh, infra.Logger)
	go runWriterWorker(ctx, infra.LoggingManagerCore.Writer, infra.Logger)

	infra.Logger.Info(logName+" service started", "config", configPath)
	<-ctx.Done()
	infra.Logger.Info(logName + " service stopping")
	return ctx.Err()
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
