package nats

import (
	"context"
	"database/sql"
	"testing"
	"time"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"

	_ "modernc.org/sqlite"
)

func TestRegisterSubscriptionsRegistersBothSubjects(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	core := mustNewCore(t)
	subjects := Subjects{AlertSubject: "alert", AlertOccurrenceSubject: "occurrence"}

	if err := RegisterSubscriptions(server, core, subjects); err != nil {
		t.Fatalf("RegisterSubscriptions() error = %v", err)
	}
	if len(server.subscriptions) != 2 {
		t.Fatalf("subscriptions count = %d, want 2", len(server.subscriptions))
	}
}

func TestRegisterRPCHandlersRegistersAllSubjects(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	core := mustNewCore(t)
	subjects := Subjects{
		GetAlertsRPCSubject:                     "getAlerts",
		GetAlertRPCSubject:                      "getAlert",
		GetActiveAlertsRPCSubject:               "getActive",
		GetUnacknowledgedActiveAlertsRPCSubject: "getUnacknowledged",
		UpdateAlertStateRPCSubject:              "update",
		AcknowledgeAlertRPCSubject:              "ack",
		MuteAlertRPCSubject:                     "mute",
	}

	if err := RegisterRPCHandlers(server, core, subjects); err != nil {
		t.Fatalf("RegisterRPCHandlers() error = %v", err)
	}
	if len(server.rpcHandlers) != 7 {
		t.Fatalf("rpc handlers count = %d, want 7", len(server.rpcHandlers))
	}
}

func TestHandleAlertAcceptsValidPayload(t *testing.T) {
	t.Parallel()

	core := mustNewCore(t)
	handler := handleAlert(core)

	err := handler(context.Background(), sharedmessaging.Envelope{
		Payload: mustMarshal(t, loggingmanagercontract.AlertPayload{
			AccessToken: "token",
			Category:    "disk",
		}),
	})
	if err != nil {
		t.Fatalf("handleAlert() error = %v", err)
	}
}

func TestHandleAlertOccurrenceAcceptsValidPayload(t *testing.T) {
	t.Parallel()

	core := mustNewCore(t)
	handler := handleAlertOccurrence(core)

	err := handler(context.Background(), sharedmessaging.Envelope{
		Payload: mustMarshal(t, loggingmanagercontract.AlertOccurrencePayload{
			AccessToken: "token",
			EventID:     "event_1",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			ValueType:   "text",
		}),
	})
	if err != nil {
		t.Fatalf("handleAlertOccurrence() error = %v", err)
	}
}

func TestReadRPCHandlersReturnSuccessResponses(t *testing.T) {
	t.Parallel()

	core := mustNewCore(t)
	listEnv := sharedmessaging.Envelope{Payload: mustMarshal(t, loggingmanagercontract.ListAlertsInput{
		AccessToken: "token",
		Page:        1,
		PageSize:    10,
	})}
	getEnv := sharedmessaging.Envelope{Payload: mustMarshal(t, loggingmanagercontract.GetAlertInput{
		AccessToken: "token",
		EventID:     "event_1",
	})}

	if _, err := handleGetAlertsRPC(core)(context.Background(), listEnv); err != nil {
		t.Fatalf("handleGetAlertsRPC() error = %v", err)
	}
	if _, err := handleGetAlertRPC(core)(context.Background(), getEnv); err != nil {
		t.Fatalf("handleGetAlertRPC() error = %v", err)
	}
	if _, err := handleGetActiveAlertsRPC(core)(context.Background(), listEnv); err != nil {
		t.Fatalf("handleGetActiveAlertsRPC() error = %v", err)
	}
	if _, err := handleGetUnacknowledgedActiveAlertsRPC(core)(context.Background(), listEnv); err != nil {
		t.Fatalf("handleGetUnacknowledgedActiveAlertsRPC() error = %v", err)
	}
}

func TestHandleUpdateAlertStateRPCReturnsOKFalseWhenEventMissing(t *testing.T) {
	t.Parallel()

	core := mustNewCore(t)
	updateEnv := sharedmessaging.Envelope{Payload: mustMarshal(t, loggingmanagercontract.UpdateAlertStateInput{
		AccessToken: "token",
		EventID:     "event_1",
		Status:      "failure",
	})}

	updateResult, err := handleUpdateAlertStateRPC(core)(context.Background(), updateEnv)
	if err != nil {
		t.Fatalf("handleUpdateAlertStateRPC() error = %v", err)
	}
	if !isOKFalse(updateResult) {
		t.Fatalf("update result = %#v, want OK=false", updateResult)
	}
}

func TestHandleAcknowledgeAlertRPCReturnsOKFalseWhenEventMissing(t *testing.T) {
	t.Parallel()

	core := mustNewCore(t)
	ackEnv := sharedmessaging.Envelope{Payload: mustMarshal(t, loggingmanagercontract.AcknowledgeAlertInput{
		AccessToken:    "token",
		EventID:        "event_1",
		AcknowledgedBy: "operator",
	})}
	ackResult, err := handleAcknowledgeAlertRPC(core)(context.Background(), ackEnv)
	if err != nil {
		t.Fatalf("handleAcknowledgeAlertRPC() error = %v", err)
	}
	if !isOKFalse(ackResult) {
		t.Fatalf("ack result = %#v, want OK=false", ackResult)
	}
}

func TestHandleMuteAlertRPCReturnsOKFalseWhenEventMissing(t *testing.T) {
	t.Parallel()

	core := mustNewCore(t)
	muteEnv := sharedmessaging.Envelope{Payload: mustMarshal(t, loggingmanagercontract.MuteAlertInput{
		AccessToken: "token",
		EventID:     "event_1",
		MutedBy:     "operator",
	})}
	muteResult, err := handleMuteAlertRPC(core)(context.Background(), muteEnv)
	if err != nil {
		t.Fatalf("handleMuteAlertRPC() error = %v", err)
	}
	if !isOKFalse(muteResult) {
		t.Fatalf("mute result = %#v, want OK=false", muteResult)
	}
}

type recordingServer struct {
	subscriptions []string
	rpcHandlers   []string
}

func (server *recordingServer) Subscribe(subject string, _ sharedmessaging.MessageHandler) error {
	server.subscriptions = append(server.subscriptions, subject)
	return nil
}

func (server *recordingServer) RegisterRPC(subject string, _ sharedmessaging.RPCHandler) error {
	server.rpcHandlers = append(server.rpcHandlers, subject)
	return nil
}

func (server *recordingServer) UseSubscriptionMiddleware(...sharedmessaging.SubscriptionMiddleware) {}

func (server *recordingServer) UseRPCMiddleware(...sharedmessaging.RPCMiddleware) {}

func (server *recordingServer) Drain() error { return nil }

func (server *recordingServer) Close() {}

func mustNewCore(t *testing.T) *sharedloggingmanager.Core {
	t.Helper()

	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	core, err := sharedloggingmanager.NewCore(context.Background(), sharedloggingmanager.CoreDeps{
		DB:             db,
		WriterInputCh:  make(chan sharedloggingmanager.WriteRequest, 64),
		Clock:          time.Now,
		MaxEvents:      100,
		MaxOccurrences: 1000,
		EventIDPrefix:  "event",
		Validator:      mustNewInputValidator(t),
	})
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}
	return core
}

func mustNewInputValidator(t *testing.T) sharedloggingmanager.InputValidator {
	t.Helper()
	validate, err := sharedloggingmanager.NewInputValidator()
	if err != nil {
		t.Fatalf("NewInputValidator() error = %v", err)
	}
	return validate
}

func mustMarshal(t *testing.T, payload any) []byte {
	t.Helper()
	data, err := sharedmessaging.NewJSONCodec().Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	return data
}

func isOKFalse(value any) bool {
	result, ok := value.(loggingmanagercontract.OKResponse)
	return ok && !result.OK
}
