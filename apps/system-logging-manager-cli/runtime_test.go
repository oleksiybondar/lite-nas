package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"lite-nas/apps/system-logging-manager-cli/workers"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/model"
)

type stubMessagingClient struct {
	publishSubject string
	publishPayload any
	publishErr     error

	requestSubject string
	requestBody    any
	requestErr     error
	requestFill    func(response any)
}

func (c *stubMessagingClient) Publish(_ context.Context, subject string, payload any) error {
	c.publishSubject = subject
	c.publishPayload = payload
	return c.publishErr
}

func (c *stubMessagingClient) Request(_ context.Context, subject string, request any, response any) error {
	c.requestSubject = subject
	c.requestBody = request
	if c.requestErr != nil {
		return c.requestErr
	}
	if c.requestFill != nil {
		c.requestFill(response)
	}
	return nil
}

type stubOutputWriter struct {
	events          []model.Event
	eventsJSONMode  bool
	ok              loggingmanagercontract.OKResponse
	okJSONMode      bool
	writeEventsErr  error
	writeOKErr      error
	writeEventsCall int
	writeOKCall     int
}

func (w *stubOutputWriter) WriteEvents(_ io.Writer, events []model.Event, jsonOutput bool) error {
	w.writeEventsCall++
	w.events = events
	w.eventsJSONMode = jsonOutput
	return w.writeEventsErr
}

func (w *stubOutputWriter) WriteOK(_ io.Writer, response loggingmanagercontract.OKResponse, jsonOutput bool) error {
	w.writeOKCall++
	w.ok = response
	w.okJSONMode = jsonOutput
	return w.writeOKErr
}

func TestExecuteCommandCreateEventPublishesToAlertSubject(t *testing.T) {
	t.Parallel()

	client := &stubMessagingClient{}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command: workers.CommandCreateEvent,
			Data:    `{"category":"disk_health","severity":"warning"}`,
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if client.publishSubject != systemloggingmanagercontract.AlertSubject {
		t.Fatalf("publish subject = %q, want %q", client.publishSubject, systemloggingmanagercontract.AlertSubject)
	}

	payload, ok := client.publishPayload.(loggingmanagercontract.AlertPayload)
	if !ok {
		t.Fatalf("publish payload type = %T, want AlertPayload", client.publishPayload)
	}
	if payload.Category != "disk_health" {
		t.Fatalf("publish payload category = %q, want disk_health", payload.Category)
	}
}

func TestExecuteCommandCreateOccurrenceUsesEventIDFlag(t *testing.T) {
	t.Parallel()

	client := &stubMessagingClient{}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command: workers.CommandCreateOccurrence,
			EventID: "event_1",
			Data:    `{"timestamp":"2026-05-12T20:00:00Z","value_type":"text","value_text":"warning"}`,
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if client.publishSubject != systemloggingmanagercontract.AlertOccurrenceSubject {
		t.Fatalf("publish subject = %q, want %q", client.publishSubject, systemloggingmanagercontract.AlertOccurrenceSubject)
	}

	payload, ok := client.publishPayload.(loggingmanagerdto.OccurrenceRow)
	if !ok {
		t.Fatalf("publish payload type = %T, want OccurrenceRow", client.publishPayload)
	}
	if payload.EventID != "event_1" {
		t.Fatalf("publish payload event id = %q, want event_1", payload.EventID)
	}
}

func TestExecuteCommandListUsesActiveEventsSubject(t *testing.T) {
	t.Parallel()

	client, _, err := executeListActiveEventsFixture(t)
	if err != nil {
		t.Fatalf("executeListActiveEventsFixture() error = %v", err)
	}

	if client.requestSubject != systemloggingmanagercontract.GetActiveAlertsRPCSubject {
		t.Fatalf("request subject = %q, want %q", client.requestSubject, systemloggingmanagercontract.GetActiveAlertsRPCSubject)
	}
}

func TestExecuteCommandListUsesProvidedPagination(t *testing.T) {
	t.Parallel()

	client, _, err := executeListActiveEventsFixture(t)
	if err != nil {
		t.Fatalf("executeListActiveEventsFixture() error = %v", err)
	}

	request, ok := client.requestBody.(loggingmanagercontract.ListAlertsInput)
	if !ok {
		t.Fatalf("request type = %T, want ListAlertsInput", client.requestBody)
	}
	if request.Page != 2 || request.PageSize != 50 {
		t.Fatalf("request pagination = (%d,%d), want (2,50)", request.Page, request.PageSize)
	}
}

func TestExecuteCommandListWritesEventsInJSONMode(t *testing.T) {
	t.Parallel()

	_, output, err := executeListActiveEventsFixture(t)
	if err != nil {
		t.Fatalf("executeListActiveEventsFixture() error = %v", err)
	}

	if output.writeEventsCall != 1 {
		t.Fatalf("WriteEvents() calls = %d, want 1", output.writeEventsCall)
	}
	if !output.eventsJSONMode {
		t.Fatal("WriteEvents() jsonOutput = false, want true")
	}
}

func TestExecuteCommandMutationRequestsAcknowledgeRPC(t *testing.T) {
	t.Parallel()

	client := &stubMessagingClient{
		requestFill: func(response any) {
			out := response.(*loggingmanagercontract.OKResponse)
			out.OK = true
		},
	}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command:    workers.CommandAcknowledgeEvent,
			Data:       `{"event_id":"event_1","acknowledged_by":"operator"}`,
			JSONOutput: true,
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if client.requestSubject != systemloggingmanagercontract.AcknowledgeAlertRPCSubject {
		t.Fatalf("request subject = %q, want %q", client.requestSubject, systemloggingmanagercontract.AcknowledgeAlertRPCSubject)
	}
	if output.writeOKCall != 1 {
		t.Fatalf("WriteOK() calls = %d, want 1", output.writeOKCall)
	}
	if !output.ok.OK {
		t.Fatal("WriteOK() response ok = false, want true")
	}
}

func TestExecuteCommandReturnsUnsupportedCommandError(t *testing.T) {
	t.Parallel()

	err := executeCommand(
		context.Background(),
		workers.Invocation{Command: workers.Command("unknown")},
		&stubMessagingClient{},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCommand() error = nil, want error")
	}
}

func TestExecuteCommandReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")
	client := &stubMessagingClient{requestErr: wantErr}

	err := executeCommand(
		context.Background(),
		workers.Invocation{Command: workers.CommandGetAlerts, Page: 1},
		client,
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCommand() error = %v, want %v", err, wantErr)
	}
}

func TestDecodeOccurrencePayloadReturnsJSONError(t *testing.T) {
	t.Parallel()

	_, err := decodeOccurrencePayload("{")
	if err == nil {
		t.Fatal("decodeOccurrencePayload() error = nil, want parse error")
	}
}

func TestExecuteCommandCreateEventReturnsJSONError(t *testing.T) {
	t.Parallel()

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command: workers.CommandCreateEvent,
			Data:    "{",
		},
		&stubMessagingClient{},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCommand() error = nil, want decode error")
	}
}

func TestExecuteCommandCreateOccurrenceReturnsJSONError(t *testing.T) {
	t.Parallel()

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command: workers.CommandCreateOccurrence,
			EventID: "event_1",
			Data:    "{",
		},
		&stubMessagingClient{},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCommand() error = nil, want decode error")
	}
}

func TestExecuteCommandListReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")
	client := &stubMessagingClient{
		requestFill: func(response any) {
			out := response.(*loggingmanagercontract.ListAlertsResponse)
			out.Items = []model.Event{}
		},
	}
	output := &stubOutputWriter{writeEventsErr: wantErr}

	err := executeCommand(
		context.Background(),
		workers.Invocation{Command: workers.CommandGetAlerts, Page: 1},
		client,
		output,
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCommand() error = %v, want %v", err, wantErr)
	}
}

func TestExecuteCommandMutationReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")
	client := &stubMessagingClient{
		requestFill: func(response any) {
			out := response.(*loggingmanagercontract.OKResponse)
			out.OK = true
		},
	}
	output := &stubOutputWriter{writeOKErr: wantErr}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command: workers.CommandMuteEvent,
			Data:    `{"event_id":"event_1","muted_by":"ops"}`,
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCommand() error = %v, want %v", err, wantErr)
	}
}

func TestPrintUsageIncludesCommands(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	printUsage(&out)
	if !strings.Contains(out.String(), "createEvent") {
		t.Fatalf("printUsage() output missing command list: %q", out.String())
	}
}

func TestRunReturnsCanceledOnHelp(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"--help"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v, want %v", err, context.Canceled)
	}
}

func TestRunReturnsArgProcessingError(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"--cmd", "unsupported"})
	if err == nil {
		t.Fatal("run() error = nil, want argument validation error")
	}
}

func TestExecuteCommandMutationReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")
	client := &stubMessagingClient{requestErr: wantErr}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command: workers.CommandUpdateEventState,
			Data:    `{"event_id":"event_1","status":"active"}`,
		},
		client,
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCommand() error = %v, want %v", err, wantErr)
	}
}

func executeListActiveEventsFixture(t *testing.T) (*stubMessagingClient, *stubOutputWriter, error) {
	t.Helper()

	client := &stubMessagingClient{
		requestFill: func(response any) {
			out := response.(*loggingmanagercontract.ListAlertsResponse)
			out.Items = []model.Event{{Event: loggingmanagerdto.EventRow{EventID: "event_2"}}}
		},
	}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Command:    workers.CommandGetActiveEvents,
			Page:       2,
			PageSize:   50,
			JSONOutput: true,
		},
		client,
		output,
		&bytes.Buffer{},
	)

	return client, output, err
}
