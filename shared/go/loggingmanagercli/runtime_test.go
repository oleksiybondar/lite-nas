package loggingmanagercli

import (
	"context"
	"errors"
	"io"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

type runtimeClientStub struct {
	publishSubject string
	publishPayload any
	publishErr     error

	requestSubject string
	requestPayload any
	requestErr     error
	responseBody   any
}

func (stub *runtimeClientStub) Publish(_ context.Context, subject string, payload any) error {
	stub.publishSubject = subject
	stub.publishPayload = payload
	return stub.publishErr
}

func (stub *runtimeClientStub) Request(_ context.Context, subject string, request any, response any) error {
	stub.requestSubject = subject
	stub.requestPayload = request
	if stub.requestErr != nil {
		return stub.requestErr
	}
	switch out := response.(type) {
	case *loggingmanagercontract.ListAlertsResponse:
		*out = stub.responseBody.(loggingmanagercontract.ListAlertsResponse)
	case *loggingmanagercontract.GetAlertResponse:
		*out = stub.responseBody.(loggingmanagercontract.GetAlertResponse)
	case *loggingmanagercontract.OKResponse:
		*out = stub.responseBody.(loggingmanagercontract.OKResponse)
	}
	return nil
}

type runtimeOutputWriterStub struct {
	eventsCalls int
	okCalls     int
	lastEvents  []loggingmanagercontract.ListAlertItem
	lastOK      loggingmanagercontract.OKResponse
}

func (stub *runtimeOutputWriterStub) WriteEvents(_ io.Writer, events []loggingmanagercontract.ListAlertItem, _ bool) error {
	stub.eventsCalls++
	stub.lastEvents = events
	return nil
}

func (stub *runtimeOutputWriterStub) WriteOK(_ io.Writer, response loggingmanagercontract.OKResponse, _ bool) error {
	stub.okCalls++
	stub.lastOK = response
	return nil
}

func TestExecuteCreateEventExecutorPublishesPayload(t *testing.T) {
	t.Parallel()

	client := &runtimeClientStub{}
	exec := executeCreateEventExecutor("system-alert")
	err := exec(context.Background(), Invocation{
		Data: `{"event_id":"event_1","category":"disk","severity":"warning"}`,
	}, client, &runtimeOutputWriterStub{}, io.Discard)
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}
	if client.publishSubject != "system-alert" {
		t.Fatalf("publish subject = %q, want system-alert", client.publishSubject)
	}
	if _, ok := client.publishPayload.(loggingmanagercontract.AlertPayload); !ok {
		t.Fatalf("publish payload type = %T, want AlertPayload", client.publishPayload)
	}
}

func TestExecuteCreateOccurrenceExecutorPublishesPayload(t *testing.T) {
	t.Parallel()

	client := &runtimeClientStub{}
	exec := executeCreateOccurrenceExecutor("system-occurrence")
	err := exec(context.Background(), Invocation{
		EventID: "event_1",
		Data:    `{"timestamp":"2026-05-13T10:00:00Z","value_type":"text","value_text":"x"}`,
	}, client, &runtimeOutputWriterStub{}, io.Discard)
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}
	payload, ok := client.publishPayload.(loggingmanagercontract.AlertOccurrencePayload)
	if !ok {
		t.Fatalf("publish payload type = %T, want AlertOccurrencePayload", client.publishPayload)
	}
	if payload.EventID != "event_1" {
		t.Fatalf("EventID = %q, want event_1", payload.EventID)
	}
}

func TestExecuteListCommandExecutorRequestsAndWrites(t *testing.T) {
	t.Parallel()

	client := &runtimeClientStub{
		responseBody: loggingmanagercontract.ListAlertsResponse{
			Items: []loggingmanagercontract.ListAlertItem{{EventID: "event_1"}},
		},
	}
	writer := &runtimeOutputWriterStub{}
	exec := executeListCommandExecutor("system-logging-manager.getAlerts")
	err := exec(context.Background(), Invocation{Page: 1, PageSize: 5}, client, writer, io.Discard)
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}
	if client.requestSubject != "system-logging-manager.getAlerts" {
		t.Fatalf("request subject = %q, want system-logging-manager.getAlerts", client.requestSubject)
	}
	if writer.eventsCalls != 1 {
		t.Fatalf("events calls = %d, want 1", writer.eventsCalls)
	}
}

func TestExecuteGetEventCommandExecutorWritesSingleItemWhenFound(t *testing.T) {
	t.Parallel()

	item := loggingmanagercontract.ListAlertItem{EventID: "event_1"}
	client := &runtimeClientStub{
		responseBody: loggingmanagercontract.GetAlertResponse{Item: &item},
	}
	writer := &runtimeOutputWriterStub{}
	exec := executeGetEventCommandExecutor("system-logging-manager.getAlert")
	err := exec(context.Background(), Invocation{EventID: "event_1"}, client, writer, io.Discard)
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}
	if len(writer.lastEvents) != 1 {
		t.Fatalf("len(lastEvents) = %d, want 1", len(writer.lastEvents))
	}
}

func TestExecuteRPCMutationCommandExecutorDecodesAndWritesOK(t *testing.T) {
	t.Parallel()

	client := &runtimeClientStub{
		responseBody: loggingmanagercontract.OKResponse{OK: true},
	}
	writer := &runtimeOutputWriterStub{}
	exec := executeRPCMutationCommandExecutor[loggingmanagercontract.AcknowledgeAlertInput]("system-logging-manager.acknowledgeAlert")
	err := exec(context.Background(), Invocation{
		Data: `{"event_id":"event_1","acknowledged_by":"operator","acknowledged_at":"2026-05-13T10:01:00Z"}`,
	}, client, writer, io.Discard)
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}
	if writer.okCalls != 1 || !writer.lastOK.OK {
		t.Fatalf("WriteOK result = %#v, want OK=true", writer.lastOK)
	}
}

func TestExecuteCommandReturnsUnsupportedCommandError(t *testing.T) {
	t.Parallel()

	err := ExecuteCommand(
		context.Background(),
		Invocation{Command: "unknown"},
		&runtimeClientStub{},
		&runtimeOutputWriterStub{},
		io.Discard,
		Subjects{},
	)
	if err == nil {
		t.Fatal("expected unsupported command error")
	}
}

func TestDecodeOccurrencePayloadRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := decodeOccurrencePayload("{")
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestExecuteListCommandExecutorPropagatesRequestError(t *testing.T) {
	t.Parallel()

	requestErr := errors.New("request failed")
	client := &runtimeClientStub{requestErr: requestErr}
	exec := executeListCommandExecutor("system-logging-manager.getAlerts")
	err := exec(context.Background(), Invocation{Page: 1}, client, &runtimeOutputWriterStub{}, io.Discard)
	if !errors.Is(err, requestErr) {
		t.Fatalf("executor() error = %v, want %v", err, requestErr)
	}
}
