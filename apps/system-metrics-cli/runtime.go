package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"lite-nas/apps/system-metrics-cli/modules"
	"lite-nas/apps/system-metrics-cli/workers"
	sharedcontracts "lite-nas/shared/contracts"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	sharedmetricscli "lite-nas/shared/metricscli"
)

const (
	defaultConfigPath = "/etc/lite-nas/system-metrics-cli.conf"
	serviceName       = sharedcontracts.AppSystemMetricsCLI
)

func run(ctx context.Context, args []string) error {
	workerModule := modules.NewWorkersModule(defaultConfigPath)
	return sharedmetricscli.Run(
		ctx,
		args,
		os.Stdout,
		serviceName,
		workerModule.ArgsProcessor.Process,
		func(invocation workers.Invocation) string { return invocation.ConfigPath },
		func(err error) bool { return errors.Is(err, workers.ErrHelpRequested) },
		printUsage,
		func(ctx context.Context, invocation workers.Invocation, client sharedmetricscli.RequestClient, writer io.Writer) error {
			return executeCommand(ctx, invocation, client, workerModule.OutputWriter, writer)
		},
	)
}

func executeCommand(
	ctx context.Context,
	invocation workers.Invocation,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	return sharedmetricscli.ExecuteMode(
		invocation.Mode,
		workers.ModeCurrent,
		workers.ModeHistory,
		func() error { return executeCurrentCommand(ctx, invocation, client, output, writer) },
		func() error { return executeHistoryCommand(ctx, client, output, writer) },
	)
}

func executeCurrentCommand(
	ctx context.Context,
	invocation workers.Invocation,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	response, err := sharedmetricscli.RequestRPC[systemmetricscontract.GetSnapshotResponse](
		ctx,
		client,
		systemmetricscontract.SnapshotRPCSubject,
		systemmetricscontract.GetSnapshotRequest{},
	)
	if err != nil {
		return err
	}

	return output.WriteCurrent(writer, response.Snapshot, invocation.CurrentSelection)
}

func executeHistoryCommand(
	ctx context.Context,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	response, err := sharedmetricscli.RequestRPC[systemmetricscontract.GetHistoryResponse](
		ctx,
		client,
		systemmetricscontract.HistoryRPCSubject,
		systemmetricscontract.GetHistoryRequest{},
	)
	if err != nil {
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
