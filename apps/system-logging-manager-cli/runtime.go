package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"lite-nas/apps/system-logging-manager-cli/modules"
	"lite-nas/apps/system-logging-manager-cli/workers"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
)

const (
	defaultConfigPath = "/etc/lite-nas/system-logging-manager-cli.conf"
	appName           = "system-logging-manager-cli"
)

type messagingClient interface {
	Publish(ctx context.Context, subject string, payload any) error
	Request(ctx context.Context, subject string, request any, response any) error
}

type commandExecutor func(
	ctx context.Context,
	invocation workers.Invocation,
	client messagingClient,
	output workers.OutputWriter,
	stdout io.Writer,
) error

// run executes CLI command flow.
func run(ctx context.Context, args []string) error {
	workersModule := modules.NewWorkersModule(defaultConfigPath)
	invocation, err := workersModule.ArgsProcessor.Process(args)
	if err != nil {
		if errors.Is(err, workers.ErrHelpRequested) {
			printUsage(os.Stdout)
			return context.Canceled
		}
		return err
	}

	infra, err := modules.NewInfraModule(invocation.ConfigPath, appName)
	if err != nil {
		return err
	}
	defer infra.Close()

	return executeCommand(ctx, invocation, infra.Client, workersModule.OutputWriter, os.Stdout)
}

func executeCommand(
	ctx context.Context,
	invocation workers.Invocation,
	client messagingClient,
	output workers.OutputWriter,
	stdout io.Writer,
) error {
	executor, ok := commandExecutors[invocation.Command]
	if !ok {
		return fmt.Errorf("unsupported command: %s", invocation.Command)
	}
	return executor(ctx, invocation, client, output, stdout)
}

var commandExecutors = map[workers.Command]commandExecutor{
	workers.CommandCreateEvent:                   executeCreateEventExecutor,
	workers.CommandCreateOccurrence:              executeCreateOccurrenceExecutor,
	workers.CommandGetAlerts:                     executeListCommandExecutor(systemloggingmanagercontract.GetAlertsRPCSubject),
	workers.CommandGetActiveEvents:               executeListCommandExecutor(systemloggingmanagercontract.GetActiveAlertsRPCSubject),
	workers.CommandGetActiveUnacknowledgedEvents: executeListCommandExecutor(systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject),
	workers.CommandUpdateEventState:              executeRPCMutationCommandExecutor[loggingmanagercontract.UpdateAlertStateInput](systemloggingmanagercontract.UpdateAlertStateRPCSubject),
	workers.CommandAcknowledgeEvent:              executeRPCMutationCommandExecutor[loggingmanagercontract.AcknowledgeAlertInput](systemloggingmanagercontract.AcknowledgeAlertRPCSubject),
	workers.CommandMuteEvent:                     executeRPCMutationCommandExecutor[loggingmanagercontract.MuteAlertInput](systemloggingmanagercontract.MuteAlertRPCSubject),
}

func executeCreateEventExecutor(
	ctx context.Context,
	invocation workers.Invocation,
	client messagingClient,
	_ workers.OutputWriter,
	_ io.Writer,
) error {
	return executeCreateEventCommand(ctx, invocation, client)
}

func executeCreateOccurrenceExecutor(
	ctx context.Context,
	invocation workers.Invocation,
	client messagingClient,
	_ workers.OutputWriter,
	_ io.Writer,
) error {
	return executeCreateOccurrenceCommand(ctx, invocation, client)
}

func executeListCommandExecutor(subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation workers.Invocation,
		client messagingClient,
		output workers.OutputWriter,
		stdout io.Writer,
	) error {
		return executeListCommand(ctx, invocation, client, output, stdout, subject)
	}
}

func executeRPCMutationCommandExecutor[T any](subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation workers.Invocation,
		client messagingClient,
		output workers.OutputWriter,
		stdout io.Writer,
	) error {
		return executeRPCMutationCommand[T](ctx, invocation, client, output, stdout, subject)
	}
}

func executeCreateEventCommand(ctx context.Context, invocation workers.Invocation, client messagingClient) error {
	var payload loggingmanagercontract.AlertPayload
	if err := decodeJSON(invocation.Data, &payload); err != nil {
		return err
	}

	return client.Publish(ctx, systemloggingmanagercontract.AlertSubject, payload)
}

func executeCreateOccurrenceCommand(ctx context.Context, invocation workers.Invocation, client messagingClient) error {
	payload, err := decodeOccurrencePayload(invocation.Data)
	if err != nil {
		return err
	}
	payload.EventID = invocation.EventID

	return client.Publish(ctx, systemloggingmanagercontract.AlertOccurrenceSubject, payload)
}

func executeListCommand(
	ctx context.Context,
	invocation workers.Invocation,
	client messagingClient,
	output workers.OutputWriter,
	stdout io.Writer,
	subject string,
) error {
	request := loggingmanagercontract.ListAlertsInput{
		Page:     invocation.Page,
		PageSize: invocation.PageSize,
	}
	response := loggingmanagercontract.ListAlertsResponse{}
	if err := client.Request(ctx, subject, request, &response); err != nil {
		return err
	}
	return output.WriteEvents(stdout, response.Items, invocation.JSONOutput)
}

func executeRPCMutationCommand[T any](
	ctx context.Context,
	invocation workers.Invocation,
	client messagingClient,
	output workers.OutputWriter,
	stdout io.Writer,
	subject string,
) error {
	var request T
	if err := decodeJSON(invocation.Data, &request); err != nil {
		return err
	}

	response := loggingmanagercontract.OKResponse{}
	if err := client.Request(ctx, subject, request, &response); err != nil {
		return err
	}
	return output.WriteOK(stdout, response, invocation.JSONOutput)
}

func printUsage(writer io.Writer) {
	_, _ = fmt.Fprintln(
		writer,
		"Usage: system-logging-manager-cli --cmd <command> [--config=/etc/lite-nas/system-logging-manager-cli.conf] [--data '<json>'] [--eventID '<eventID>'] [--page <page>] [--pageSize <size>] [--json]",
	)
	_, _ = fmt.Fprintln(writer, "Commands: createEvent, createOccurrence, getAlerts, getActiveEvents, getActiveUnacknowledgedEvents, updateEventState, acknowledgeEvent, muteEvent")
}

func decodeJSON(data string, target any) error {
	if err := json.Unmarshal([]byte(data), target); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}
	return nil
}

type occurrencePayloadInput struct {
	RecID      *int64          `json:"rec_id,omitempty"`
	EventID    string          `json:"event_id,omitempty"`
	EventRecID *int64          `json:"event_rec_id,omitempty"`
	Timestamp  string          `json:"timestamp,omitempty"`
	ValueType  enum.ValueType  `json:"value_type,omitempty"`
	ValueNum   *float64        `json:"value_num,omitempty"`
	ValueText  *string         `json:"value_text,omitempty"`
	ValueBool  *bool           `json:"value_bool,omitempty"`
	ValueUnit  *string         `json:"value_unit,omitempty"`
	LegacyBody json.RawMessage `json:"-"`
}

func decodeOccurrencePayload(data string) (loggingmanagerdto.OccurrenceRow, error) {
	var payload occurrencePayloadInput
	if err := decodeJSON(data, &payload); err != nil {
		return loggingmanagerdto.OccurrenceRow{}, err
	}

	row := loggingmanagerdto.OccurrenceRow{
		EventID:   payload.EventID,
		Timestamp: payload.Timestamp,
		ValueType: payload.ValueType,
		ValueNum:  payload.ValueNum,
		ValueText: payload.ValueText,
		ValueBool: payload.ValueBool,
		ValueUnit: payload.ValueUnit,
	}

	if payload.RecID != nil {
		row.RecID = *payload.RecID
	}
	if payload.EventRecID != nil {
		row.EventRecID = *payload.EventRecID
	}
	return row, nil
}
