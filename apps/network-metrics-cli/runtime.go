package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"lite-nas/apps/network-metrics-cli/modules"
	"lite-nas/apps/network-metrics-cli/workers"
	sharedcontracts "lite-nas/shared/contracts"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	"lite-nas/shared/metrics"
	sharedmetricscli "lite-nas/shared/metricscli"
)

const (
	defaultConfigPath = "/etc/lite-nas/network-metrics-cli.conf"
	serviceName       = sharedcontracts.AppNetworkMetricsCLI
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
	response, err := sharedmetricscli.RequestRPC[networkmetricscontract.GetSnapshotResponse](
		ctx,
		client,
		networkmetricscontract.SnapshotRPCSubject,
		networkmetricscontract.GetSnapshotRequest{},
	)
	if err != nil {
		return err
	}
	if !response.Available {
		return errors.New("network snapshot is not available yet")
	}

	return output.WriteCurrent(writer, response.Snapshot, invocation.CurrentSelection)
}

func executeHistoryCommand(
	ctx context.Context,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	writer io.Writer,
) error {
	items, err := requestHistory(ctx, client)
	if err != nil {
		return err
	}

	return output.WriteHistory(writer, items)
}

func requestHistory(
	ctx context.Context,
	client sharedmetricscli.RequestClient,
) ([]metrics.NetworkMetricsSnapshot, error) {
	response, err := sharedmetricscli.RequestRPC[networkmetricscontract.GetHistoryResponse](
		ctx,
		client,
		networkmetricscontract.HistoryRPCSubject,
		networkmetricscontract.GetHistoryRequest{},
	)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(
		writer,
		"Usage: network-metrics-cli [--config=/etc/lite-nas/network-metrics-cli.conf] [--interfaces] [--protocols] [--sockets] [--pressure] [--history]",
	)
}
