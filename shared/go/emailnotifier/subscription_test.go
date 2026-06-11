package emailnotifier

import (
	"context"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
	sharedmessaging "lite-nas/shared/messaging"
)

func TestAlertSubscriptionHandlerForwardsValidatedPayload(t *testing.T) {
	t.Parallel()

	validate := mustNewLoggingManagerValidator(t)
	output := make(chan loggingmanagercontract.AlertPayload, 1)
	handler := NewAlertSubscriptionHandler(validate, output)
	payload := buildSubscriptionAlertPayload()
	codec := sharedmessaging.NewJSONCodec()
	data := mustMarshalAlertPayload(t, codec, payload)

	if err := handler(context.Background(), sharedmessaging.Envelope{Payload: data}); err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	select {
	case got := <-output:
		if got.EventID != payload.EventID {
			t.Fatalf("got.EventID = %q, want %q", got.EventID, payload.EventID)
		}
		if got.Message != payload.Message {
			t.Fatalf("got.Message = %q, want %q", got.Message, payload.Message)
		}
	default:
		t.Fatal("expected forwarded alert payload")
	}
}

func TestAlertSubscriptionHandlerRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	validate := mustNewLoggingManagerValidator(t)
	output := make(chan loggingmanagercontract.AlertPayload, 1)
	handler := NewAlertSubscriptionHandler(validate, output)

	if err := handler(context.Background(), sharedmessaging.Envelope{Payload: []byte("{")}); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestAlertSubscriptionHandlerRejectsSchemaViolation(t *testing.T) {
	t.Parallel()

	validate := mustNewLoggingManagerValidator(t)
	output := make(chan loggingmanagercontract.AlertPayload, 1)
	handler := NewAlertSubscriptionHandler(validate, output)
	codec := sharedmessaging.NewJSONCodec()
	data := mustMarshalAlertPayload(t, codec, loggingmanagercontract.AlertPayload{
		AccessToken: "token",
	})

	if err := handler(context.Background(), sharedmessaging.Envelope{Payload: data}); err == nil {
		t.Fatal("expected validation error")
	}
}

func mustNewLoggingManagerValidator(t *testing.T) sharedloggingmanager.InputValidator {
	t.Helper()

	validate, err := sharedloggingmanager.NewInputValidator()
	if err != nil {
		t.Fatalf("NewInputValidator() error = %v", err)
	}

	return validate
}

func mustMarshalAlertPayload(
	t *testing.T,
	codec sharedmessaging.JSONCodec,
	payload loggingmanagercontract.AlertPayload,
) []byte {
	t.Helper()

	data, err := codec.Marshal(payload)
	if err != nil {
		t.Fatalf("codec.Marshal() error = %v", err)
	}

	return data
}

func buildSubscriptionAlertPayload() loggingmanagercontract.AlertPayload {
	priority := 2

	return loggingmanagercontract.AlertPayload{
		AccessToken:  "token-1",
		EventID:      "sysram_00000001",
		Category:     "system.metrics.mem.used",
		Severity:     sharedloggingenum.SeverityWarning,
		Priority:     &priority,
		CreatedAt:    "2026-05-31T10:00:00Z",
		Source:       "system-metrics",
		Message:      "RAM usage is above threshold",
		TriggerValue: "93",
	}
}
