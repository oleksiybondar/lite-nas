package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"lite-nas/apps/zfs-metrics-cli/modules"
	"lite-nas/apps/zfs-metrics-cli/workers"
	sharedcontracts "lite-nas/shared/contracts"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	sharedmetricscli "lite-nas/shared/metricscli"
)

const (
	defaultConfigPath = "/etc/lite-nas/zfs-metrics-cli.conf"
	serviceName       = sharedcontracts.AppZFSMetricsCLI
)

// run parses CLI args, boots runtime infra, and executes selected command.
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
		func() error { return executeCurrentCommand(ctx, client, output, writer) },
		func() error { return executeHistoryCommand(ctx, client, output, writer) },
	)
}

func executeCurrentCommand(
	ctx context.Context,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	response, err := sharedmetricscli.RequestRPC[zfsmetricscontract.GetSnapshotResponse](
		ctx,
		client,
		zfsmetricscontract.SnapshotRPCSubject,
		zfsmetricscontract.GetSnapshotRequest{},
	)
	if err != nil {
		return err
	}

	if !response.Available {
		_, err := fmt.Fprintln(writer, "No ZFS snapshot available yet.")
		return err
	}

	return output.WriteCurrent(writer, response.Snapshot)
}

func executeHistoryCommand(
	ctx context.Context,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	response, err := sharedmetricscli.RequestRPC[zfsmetricscontract.GetHistoryResponse](
		ctx,
		client,
		zfsmetricscontract.HistoryRPCSubject,
		zfsmetricscontract.GetHistoryRequest{},
	)
	if err != nil {
		return err
	}

	return output.WriteHistory(writer, response.Items)
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(
		writer,
		"Usage: zfs-metrics-cli [--config=/etc/lite-nas/zfs-metrics-cli.conf] [--history]",
	)
}
