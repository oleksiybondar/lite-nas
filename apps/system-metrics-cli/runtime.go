package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"lite-nas/apps/system-metrics-cli/modules"
	"lite-nas/apps/system-metrics-cli/workers"
	"lite-nas/shared/metrics"
)

const (
	defaultConfigPath = "/etc/liteNAS/system-metrics-cli.conf"
	statsRPCSubject   = "system.metrics.rpc.stats.get"
	historyRPCSubject = "system.metrics.rpc.history.get"
	serviceName       = "system-metrics-cli"
)

type requestClient interface {
	Request(ctx context.Context, subject string, request any, response any) error
}

func run(ctx context.Context, args []string) error {
	workerModule := modules.NewWorkersModule(defaultConfigPath)

	invocation, err := workerModule.ArgsProcessor().Process(args)
	if err != nil {
		if errors.Is(err, workers.ErrHelpRequested) {
			printUsage(os.Stdout)
			return context.Canceled
		}

		return err
	}

	infra, err := modules.NewInfraModule(invocation.ConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	return executeCommand(ctx, invocation, infra.Client(), workerModule.OutputWriter(), os.Stdout)
}

func executeCommand(
	ctx context.Context,
	invocation workers.Invocation,
	client requestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	switch invocation.Mode {
	case workers.ModeCurrent:
		return executeCurrentCommand(ctx, invocation, client, output, writer)
	case workers.ModeHistory:
		return executeHistoryCommand(ctx, client, output, writer)
	default:
		return fmt.Errorf("unsupported invocation mode: %s", invocation.Mode)
	}
}

func executeCurrentCommand(
	ctx context.Context,
	invocation workers.Invocation,
	client requestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	var snapshot metrics.SystemSnapshot
	if err := client.Request(ctx, statsRPCSubject, map[string]any{}, &snapshot); err != nil {
		return err
	}

	return output.WriteCurrent(writer, snapshot, invocation.CurrentSelection)
}

func executeHistoryCommand(
	ctx context.Context,
	client requestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	var history []metrics.SystemSnapshot
	if err := client.Request(ctx, historyRPCSubject, map[string]any{}, &history); err != nil {
		return err
	}

	return output.WriteHistory(writer, history)
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(
		writer,
		"Usage: system-metrics-cli [--config=/etc/liteNAS/system-metrics-cli.conf] [--cpu] [--ram] [--history]",
	)
}
