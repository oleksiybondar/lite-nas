package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"lite-nas/apps/system-metrics-cli/modules"
	"lite-nas/apps/system-metrics-cli/workers"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
)

const (
	defaultConfigPath = "/etc/lite-nas/system-metrics-cli.conf"
	serviceName       = "system-metrics-cli"
)

type requestClient interface {
	Request(ctx context.Context, subject string, request any, response any) error
}

func run(ctx context.Context, args []string) error {
	workerModule := modules.NewWorkersModule(defaultConfigPath)

	invocation, err := workerModule.ArgsProcessor.Process(args)
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

	return executeCommand(ctx, invocation, infra.Client, workerModule.OutputWriter, os.Stdout)
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
	var response systemmetricscontract.GetSnapshotResponse
	if err := client.Request(ctx, systemmetricscontract.SnapshotRPCSubject, systemmetricscontract.GetSnapshotRequest{}, &response); err != nil {
		return err
	}

	return output.WriteCurrent(writer, response.Snapshot, invocation.CurrentSelection)
}

func executeHistoryCommand(
	ctx context.Context,
	client requestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	var response systemmetricscontract.GetHistoryResponse
	if err := client.Request(ctx, systemmetricscontract.HistoryRPCSubject, systemmetricscontract.GetHistoryRequest{}, &response); err != nil {
		return err
	}

	return output.WriteHistory(writer, response.Items)
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(
		writer,
		"Usage: system-metrics-cli [--config=/etc/lite-nas/system-metrics-cli.conf] [--cpu] [--ram] [--history]",
	)
}
