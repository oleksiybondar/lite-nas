package loggingmanagercli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	"lite-nas/shared/loggingmanager/enum"
)

// Subjects defines CLI publish/request subjects for a logging-manager domain.
type Subjects struct {
	AlertSubject                            string
	AlertOccurrenceSubject                  string
	GetAlertsRPCSubject                     string
	GetAlertRPCSubject                      string
	GetActiveAlertsRPCSubject               string
	GetUnacknowledgedActiveAlertsRPCSubject string
	UpdateAlertStateRPCSubject              string
	AcknowledgeAlertRPCSubject              string
	MuteAlertRPCSubject                     string
}

// Invocation contains validated CLI execution settings.
type Invocation struct {
	ConfigPath string
	Command    Command
	Data       string
	EventID    string
	Page       int
	PageSize   int
	JSONOutput bool
}

// MessagingClient represents required NATS client behavior for command execution.
type MessagingClient interface {
	Publish(ctx context.Context, subject string, payload any) error
	Request(ctx context.Context, subject string, request any, response any) error
}

// OutputWriter renders command output in table or JSON modes.
type OutputWriter interface {
	WriteEvents(writer io.Writer, events []loggingmanagercontract.ListAlertItem, jsonOutput bool) error
	WriteOK(writer io.Writer, response loggingmanagercontract.OKResponse, jsonOutput bool) error
}

type commandExecutor func(
	ctx context.Context,
	invocation Invocation,
	client MessagingClient,
	output OutputWriter,
	stdout io.Writer,
) error

// ExecuteCommand runs a parsed command invocation.
func ExecuteCommand(
	ctx context.Context,
	invocation Invocation,
	client MessagingClient,
	output OutputWriter,
	stdout io.Writer,
	subjects Subjects,
) error {
	executor, ok := commandExecutors(subjects)[invocation.Command]
	if !ok {
		return fmt.Errorf("unsupported command: %s", invocation.Command)
	}
	return executor(ctx, invocation, client, output, stdout)
}

// PrintUsage writes CLI usage for a specific app/config path pair.
func PrintUsage(writer io.Writer, appName string, defaultConfigPath string) {
	_, _ = fmt.Fprintf(
		writer,
		"Usage: %s --cmd <command> [--config=%s] [--data '<json>'] [--eventID '<eventID>'] [--page <page>] [--pageSize <size>] [--json]\n",
		appName,
		defaultConfigPath,
	)
	_, _ = fmt.Fprintln(writer, "Commands: createEvent, createOccurrence, getEvent, getAlerts|getEvents, getActiveEvents, getActiveUnacknowledgedEvents, updateEventState, acknowledgeEvent, muteEvent")
}

func commandExecutors(subjects Subjects) map[Command]commandExecutor {
	return map[Command]commandExecutor{
		CommandCreateEvent:                   executeCreateEventExecutor(subjects.AlertSubject),
		CommandCreateOccurrence:              executeCreateOccurrenceExecutor(subjects.AlertOccurrenceSubject),
		CommandGetEvent:                      executeGetEventCommandExecutor(subjects.GetAlertRPCSubject),
		CommandGetAlerts:                     executeListCommandExecutor(subjects.GetAlertsRPCSubject),
		CommandGetActiveEvents:               executeListCommandExecutor(subjects.GetActiveAlertsRPCSubject),
		CommandGetActiveUnacknowledgedEvents: executeListCommandExecutor(subjects.GetUnacknowledgedActiveAlertsRPCSubject),
		CommandUpdateEventState:              executeRPCMutationCommandExecutor[loggingmanagercontract.UpdateAlertStateInput](subjects.UpdateAlertStateRPCSubject),
		CommandAcknowledgeEvent:              executeRPCMutationCommandExecutor[loggingmanagercontract.AcknowledgeAlertInput](subjects.AcknowledgeAlertRPCSubject),
		CommandMuteEvent:                     executeRPCMutationCommandExecutor[loggingmanagercontract.MuteAlertInput](subjects.MuteAlertRPCSubject),
	}
}

func executeCreateEventExecutor(subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation Invocation,
		client MessagingClient,
		_ OutputWriter,
		_ io.Writer,
	) error {
		var payload loggingmanagercontract.AlertPayload
		if err := decodeJSON(invocation.Data, &payload); err != nil {
			return err
		}
		return client.Publish(ctx, subject, payload)
	}
}

func executeCreateOccurrenceExecutor(subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation Invocation,
		client MessagingClient,
		_ OutputWriter,
		_ io.Writer,
	) error {
		payload, err := decodeOccurrencePayload(invocation.Data)
		if err != nil {
			return err
		}
		payload.EventID = invocation.EventID
		return client.Publish(ctx, subject, payload)
	}
}

func executeListCommandExecutor(subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation Invocation,
		client MessagingClient,
		output OutputWriter,
		stdout io.Writer,
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
}

func executeGetEventCommandExecutor(subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation Invocation,
		client MessagingClient,
		output OutputWriter,
		stdout io.Writer,
	) error {
		request := loggingmanagercontract.GetAlertInput{EventID: invocation.EventID}
		response := loggingmanagercontract.GetAlertResponse{}
		if err := client.Request(ctx, subject, request, &response); err != nil {
			return err
		}

		items := []loggingmanagercontract.ListAlertItem{}
		if response.Item != nil {
			items = append(items, *response.Item)
		}
		return output.WriteEvents(stdout, items, invocation.JSONOutput)
	}
}

func executeRPCMutationCommandExecutor[T any](subject string) commandExecutor {
	return func(
		ctx context.Context,
		invocation Invocation,
		client MessagingClient,
		output OutputWriter,
		stdout io.Writer,
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

func decodeOccurrencePayload(data string) (loggingmanagercontract.AlertOccurrencePayload, error) {
	var payload occurrencePayloadInput
	if err := decodeJSON(data, &payload); err != nil {
		return loggingmanagercontract.AlertOccurrencePayload{}, err
	}

	row := loggingmanagercontract.AlertOccurrencePayload{
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

func decodeJSON(data string, target any) error {
	if err := json.Unmarshal([]byte(data), target); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}
	return nil
}

// Run parses CLI args, loads infra using provided callback, and executes command.
func Run(
	ctx context.Context,
	args []string,
	defaultConfigPath string,
	appName string,
	subjects Subjects,
	loadInfra func(configPath string, serviceName string) (closeFn func(), client MessagingClient, err error),
	stdout io.Writer,
) error {
	processor := NewArgsProcessor(defaultConfigPath)
	output := NewOutputWriter()
	invocation, err := processor.Process(args)
	if err != nil {
		if errors.Is(err, ErrHelpRequested) {
			PrintUsage(stdout, appName, defaultConfigPath)
			return context.Canceled
		}
		return err
	}

	closeFn, client, err := loadInfra(invocation.ConfigPath, appName)
	if err != nil {
		return err
	}
	defer closeFn()

	return ExecuteCommand(ctx, invocation, client, output, stdout, subjects)
}

// StderrErrorLine writes a standardized app error line to stderr.
func StderrErrorLine(appName string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", appName, err)
}
