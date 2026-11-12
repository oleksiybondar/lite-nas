package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	systemcontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/messaging"
)

// alertServer captures transport handlers registered by the service adapter.
type alertServer struct {
	messaging.Server
	handlers map[string]messaging.MessageHandler
	failOn   string
}

// Subscribe records each handler or returns the configured registration error.
func (server *alertServer) Subscribe(subject string, handler messaging.MessageHandler) error {
	if subject == server.failOn {
		return errors.New("subscription unavailable")
	}
	server.handlers[subject] = handler
	return nil
}

// mustAlertHandlers registers both alert producers for an isolated handler test.
func mustAlertHandlers(t *testing.T, output chan<- alertObservation) *alertServer {
	t.Helper()
	server := &alertServer{handlers: make(map[string]messaging.MessageHandler)}
	if err := subscribeAlertCreations(context.Background(), server, output); err != nil {
		t.Fatal(err)
	}
	return server
}

// Requirements: security-automation/IR-002.
// TestAlertSubscriptionsBridgeBothSources checks both configured subjects reach the observation channel.
func TestAlertSubscriptionsBridgeBothSources(t *testing.T) {
	for _, test := range []struct{ subject, kind string }{
		{securitycontract.AlertSubject, "security"}, {systemcontract.AlertSubject, "system"},
	} {
		t.Run(test.kind, func(t *testing.T) {
			output := make(chan alertObservation, 1)
			server := mustAlertHandlers(t, output)
			payload := []byte(`{"access_token":"token","event_id":"sec_1","category":"login","source":"auth"}`)
			if err := server.handlers[test.subject](context.Background(), messaging.Envelope{Payload: payload}); err != nil {
				t.Fatal(err)
			}
			if got := <-output; got.Kind != test.kind || got.Alert.EventID != "sec_1" {
				t.Fatalf("observation = %+v", got)
			}
		})
	}
}

// Requirements: security-automation/IR-002.
// TestAlertSubscriptionRejectsMalformedMessages checks invalid transport payloads are handled as errors.
func TestAlertSubscriptionRejectsMalformedMessages(t *testing.T) {
	cases := []struct{ name, payload string }{
		{"empty", ""},
		{"oversized", strings.Repeat("x", 4097)},
		{"invalid JSON", "{"},
		{"unknown field", `{"access_token":"token","event_id":"sec_1","category":"login","extra":1}`},
		{"trailing JSON", `{"access_token":"token","event_id":"sec_1","category":"login"} {}`},
		{"missing access token", `{"event_id":"sec_1","category":"login"}`},
		{"missing event ID", `{"access_token":"token","category":"login"}`},
		{"invalid event ID", `{"access_token":"token","event_id":"bad-id","category":"login"}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			server := mustAlertHandlers(t, make(chan alertObservation, 1))
			err := server.handlers[securitycontract.AlertSubject](context.Background(), messaging.Envelope{Payload: []byte(test.payload)})
			if err == nil {
				t.Fatal("invalid alert reached automation")
			}
		})
	}
}

// Requirements: security-automation/IR-002.
// TestAlertSubscriptionSkipsOwnAlerts checks automation audit alerts cannot feed their own detector.
func TestAlertSubscriptionSkipsOwnAlerts(t *testing.T) {
	output := make(chan alertObservation, 1)
	server := mustAlertHandlers(t, output)
	payload := []byte(`{"access_token":"token","event_id":"sec_1","category":"audit","source":"security-automation"}`)
	if err := server.handlers[securitycontract.AlertSubject](context.Background(), messaging.Envelope{Payload: payload}); err != nil {
		t.Fatal(err)
	}
	if len(output) != 0 {
		t.Fatal("own alert was queued")
	}
}

// Requirements: security-automation/IR-002.
// TestAlertSubscriptionReportsCanceledSend checks backpressure respects handler cancellation.
func TestAlertSubscriptionReportsCanceledSend(t *testing.T) {
	server := mustAlertHandlers(t, make(chan alertObservation))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	payload := []byte(`{"access_token":"token","event_id":"sec_1","category":"login"}`)
	if err := server.handlers[securitycontract.AlertSubject](ctx, messaging.Envelope{Payload: payload}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled send error = %v", err)
	}
}

// Requirements: security-automation/IR-002.
// TestAlertSubscriptionReportsRegistrationFailure checks startup fails when an alert source is unavailable.
func TestAlertSubscriptionReportsRegistrationFailure(t *testing.T) {
	server := &alertServer{handlers: make(map[string]messaging.MessageHandler), failOn: systemcontract.AlertSubject}
	if err := subscribeAlertCreations(context.Background(), server, make(chan alertObservation, 1)); err == nil {
		t.Fatal("failed subscription accepted")
	}
}
