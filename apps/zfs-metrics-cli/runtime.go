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
)

const (
	defaultConfigPath = "/etc/lite-nas/zfs-metrics-cli.conf"
	serviceName       = sharedcontracts.AppZFSMetricsCLI
)

type requestClient interface {
	Request(ctx context.Context, subject string, request any, response any) error
}

// run parses CLI args, boots runtime infra, and executes selected command.
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
		return executeCurrentCommand(ctx, client, output, writer)
	case workers.ModeHistory:
		return executeHistoryCommand(ctx, client, output, writer)
	default:
		return fmt.Errorf("unsupported invocation mode: %s", invocation.Mode)
	}
}

func executeCurrentCommand(
	ctx context.Context,
	client requestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	var response zfsmetricscontract.GetSnapshotResponse
	if err := client.Request(ctx, zfsmetricscontract.SnapshotRPCSubject, zfsmetricscontract.GetSnapshotRequest{}, &response); err != nil {
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
	client requestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	var response zfsmetricscontract.GetHistoryResponse
	if err := client.Request(ctx, zfsmetricscontract.HistoryRPCSubject, zfsmetricscontract.GetHistoryRequest{}, &response); err != nil {
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
