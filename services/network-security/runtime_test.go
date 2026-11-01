package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	networksecuritycontract "lite-nas/shared/contracts/networksecurity"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// workerServer records the two handlers and lets a test deliver typed events.
type workerServer struct {
	handlers map[string]messaging.MessageHandler
	ready    chan struct{}
}

// Subscribe registers one event handler before the worker begins processing.
func (server *workerServer) Subscribe(subject string, handler messaging.MessageHandler) error {
	server.handlers[subject] = handler
	if len(server.handlers) == 2 {
		close(server.ready)
	}
	return nil
}

// RegisterRPC is unused by the subscription worker.
func (*workerServer) RegisterRPC(string, messaging.RPCHandler) error { return nil }

// UseSubscriptionMiddleware is unused by the subscription worker.
func (*workerServer) UseSubscriptionMiddleware(...messaging.SubscriptionMiddleware) {}

// UseRPCMiddleware is unused by the subscription worker.
func (*workerServer) UseRPCMiddleware(...messaging.RPCMiddleware) {}

// Drain completes after synchronous test callback delivery.
func (*workerServer) Drain() error { return nil }

// Close has no external connection to release.
func (*workerServer) Close() {}

// workerClient records each alert published by the outbound worker.
type workerClient struct {
	alerts chan loggingcontract.AlertPayload
	fail   error
	calls  int
}

// Publish forwards one alert for the test to observe.
func (client *workerClient) Publish(_ context.Context, _ string, payload any) error {
	client.calls++
	if client.fail != nil {
		return client.fail
	}
	client.alerts <- payload.(loggingcontract.AlertPayload)
	return nil
}

// TestReportAlertsAccountsForFailedShutdownDelivery checks bounded shutdown.
// Requirements: network-security/FR-004.
func TestReportAlertsAccountsForFailedShutdownDelivery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := &workerClient{fail: errors.New("transport unavailable")}
	effects := make(chan alertIntent, 1)
	effects <- alertIntent{ID: "secnet_12345678", Detection: scanDetection{At: time.Now()}}
	close(effects)
	reportAlerts(ctx, effects, client, sharedlogger.NewNop())
	if client.calls != 1 {
		t.Fatalf("publication attempts = %d", client.calls)
	}
}

// TestDecodeAttemptRejectsNoncanonicalAddress checks transport address rules.
// Requirements: network-security/FR-001.
func TestDecodeAttemptRejectsNoncanonicalAddress(t *testing.T) {
	payload := []byte(`{"timestamp":"2026-11-01T17:18:42Z","source_ip":"::ffff:192.0.2.4","destination_ip":"192.0.2.10","destination_port":80,"protocol":"TCP"}`)
	_, err := decodeAttempt(messaging.Envelope{Payload: payload}, validator.New(validator.WithRequiredStructEnabled()))
	if err == nil {
		t.Fatal("mapped source address was accepted")
	}
}

// TestDecodeJSONRejectsTrailingValues checks strict event framing.
// Requirements: network-security/FR-001.
func TestDecodeJSONRejectsTrailingValues(t *testing.T) {
	var value map[string]any
	err := decodeJSON(messaging.Envelope{Payload: []byte(`{} {}`)}, 32, &value)
	if err == nil {
		t.Fatal("trailing value was accepted")
	}
}

// TestDecodeJSONRejectsEmptyPayload checks the bounded NATS message boundary.
// Requirements: network-security/FR-001.
func TestDecodeJSONRejectsEmptyPayload(t *testing.T) {
	var value map[string]any
	if err := decodeJSON(messaging.Envelope{}, 32, &value); err == nil {
		t.Fatal("empty event was accepted")
	}
}

// TestDecodeSnapshotAcceptsTimestamp checks optional snapshot context input.
// Requirements: network-security/FR-002.
func TestDecodeSnapshotAcceptsTimestamp(t *testing.T) {
	event, err := decodeSnapshot(messaging.Envelope{Payload: []byte(`{"snapshot":{"timestamp":"2026-11-01T17:18:42Z"}}`)})
	if err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if event.Snapshot.Timestamp.IsZero() {
		t.Fatal("snapshot timestamp was lost")
	}
}

// Request is unused by alert publication.
func (*workerClient) Request(context.Context, string, any, any) error { return nil }

// Drain has no external connection to release.
func (*workerClient) Drain() error { return nil }

// Close has no external connection to release.
func (*workerClient) Close() {}

// TestRunWorkersDeliversPortScan verifies subscriptions, FSM, and alert worker.
// Requirements: network-security/FR-001, network-security/FR-002, network-security/FR-004.
func TestRunWorkersDeliversPortScan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := &workerServer{handlers: make(map[string]messaging.MessageHandler), ready: make(chan struct{})}
	client := &workerClient{alerts: make(chan loggingcontract.AlertPayload, 1)}
	policy := patternConfig{Threshold: 3, MaxPairs: 16, Window: time.Minute, Cooldown: time.Minute}
	cfg := securityConfig{Threshold: 3, MaxPairs: 16, Window: time.Minute, Cooldown: time.Minute, Horizontal: policy, Repeated: policy}
	done := make(chan error, 1)
	go func() { done <- runWorkers(ctx, cfg, server, client, sharedlogger.NewNop()) }()
	select {
	case <-server.ready:
	case <-time.After(3 * time.Second):
		t.Fatal("subscriptions not registered")
	}
	handler := server.handlers[networksecuritycontract.FirewallAttemptSubject]
	for port := 80; port < 83; port++ {
		payload := []byte(fmt.Sprintf(`{"timestamp":%q,"source_ip":"192.0.2.4","destination_ip":"192.0.2.10","destination_port":%d,"protocol":"TCP"}`, time.Now().UTC().Format(time.RFC3339), port))
		if err := handler(ctx, messaging.Envelope{Payload: payload}); err != nil {
			t.Fatalf("deliver attempt: %v", err)
		}
	}
	select {
	case alert := <-client.alerts:
		if alert.Category != "port_scan" {
			t.Fatalf("alert category = %q", alert.Category)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("port scan alert not published")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("workers did not drain")
	}
}

// TestDecodeAttemptRejectsOptionShapedPort verifies JSON port shape.
// Requirements: network-security/FR-001.
func TestDecodeAttemptRejectsOptionShapedPort(t *testing.T) {
	envelope := messaging.Envelope{Payload: []byte(`{"timestamp":"2026-11-01T17:18:42Z","source_ip":"192.0.2.4","destination_ip":"192.0.2.10","destination_port":"-1","protocol":"TCP"}`)}
	_, err := decodeAttempt(envelope, validator.New(validator.WithRequiredStructEnabled()))
	if err == nil {
		t.Fatal("option-shaped port was accepted")
	}
}

// TestDecodeAttemptRejectsUnknownField verifies the exact NATS contract.
// Requirements: network-security/FR-001.
func TestDecodeAttemptRejectsUnknownField(t *testing.T) {
	envelope := messaging.Envelope{Payload: []byte(`{"timestamp":"2026-11-01T17:18:42Z","source_ip":"192.0.2.4","destination_ip":"192.0.2.10","destination_port":80,"protocol":"TCP","command":"x"}`)}
	_, err := decodeAttempt(envelope, validator.New(validator.WithRequiredStructEnabled()))
	if err == nil {
		t.Fatal("unknown field was accepted")
	}
}

// TestPublishAlertUsesSecurityManagerContract verifies alert routing and text.
// Requirements: network-security/FR-004.
func TestPublishAlertUsesSecurityManagerContract(t *testing.T) {
	client := &alertRecordingClient{}
	effect := alertIntent{ID: "secnet_12345678", Detection: scanDetection{
		SourceIP: "192.0.2.4", DestinationIP: "192.0.2.10", PortCount: 12, At: time.Now(),
	}, Window: 2 * time.Minute}
	if err := publishAlert(context.Background(), client, effect); err != nil {
		t.Fatalf("publish alert: %v", err)
	}
	if client.subject != "security-alert" || client.payload.Category != "port_scan" || !strings.Contains(client.payload.Message, "12 distinct ports") {
		t.Fatalf("unexpected alert: %s %+v", client.subject, client.payload)
	}
}

// TestAdditionalAlertsUseTheirCategories verifies routing for new findings.
// Requirements: network-security/FR-005, network-security/FR-006.
func TestAdditionalAlertsUseTheirCategories(t *testing.T) {
	for _, kind := range []string{"horizontal_scan", "repeated_attempts"} {
		t.Run(kind, func(t *testing.T) {
			client := &alertRecordingClient{}
			effect := alertIntent{ID: "secnet_12345678", Detection: scanDetection{
				Kind: kind, SourceIP: "192.0.2.4", DestinationIP: "192.0.2.10",
				DestinationPort: 80, PortCount: 12, At: time.Now(),
			}, Window: 2 * time.Minute}
			if err := publishAlert(context.Background(), client, effect); err != nil {
				t.Fatalf("publish alert: %v", err)
			}
			if client.payload.Category != kind {
				t.Fatalf("alert category = %q", client.payload.Category)
			}
		})
	}
}

// alertRecordingClient captures one Security Logging Manager publication.
type alertRecordingClient struct {
	subject string
	payload loggingcontract.AlertPayload
}

// Publish records the typed payload supplied by the reporter.
func (client *alertRecordingClient) Publish(_ context.Context, subject string, payload any) error {
	client.subject = subject
	client.payload = payload.(loggingcontract.AlertPayload)
	return nil
}

// Request is unused by alert publication.
func (*alertRecordingClient) Request(context.Context, string, any, any) error { return nil }

// Drain is unused by this in-memory client.
func (*alertRecordingClient) Drain() error { return nil }

// Close is unused by this in-memory client.
func (*alertRecordingClient) Close() {}

var _ messaging.Client = (*alertRecordingClient)(nil)

// TestDecodeSnapshotRejectsMissingTimestamp protects the context boundary.
// Requirements: network-security/FR-002.
func TestDecodeSnapshotRejectsMissingTimestamp(t *testing.T) {
	_, err := decodeSnapshot(messaging.Envelope{Payload: []byte(`{"snapshot":{}}`)})
	if err == nil {
		t.Fatal("snapshot without timestamp was accepted")
	}
}

// TestCoordinateRoutesBlockedAttemptsToEachDetector verifies the event pipeline.
// Requirements: network-security/FR-002, network-security/FR-005, network-security/FR-006.
func TestCoordinateRoutesBlockedAttemptsToEachDetector(t *testing.T) {
	for _, kind := range []string{"port_scan", "horizontal_scan", "repeated_attempts"} {
		t.Run(kind, func(t *testing.T) {
			now := time.Now()
			plugin := mustPatternPlugin(t, kind, &now)
			attempts := make(chan networksecuritycontract.FirewallAttempt, 3)
			snapshots := make(chan networkmetricscontract.SnapshotUpdatedEvent)
			effects := make(chan alertIntent, 3)
			for i := 0; i < 3; i++ {
				destination := "192.0.2.10"
				port := uint16(80)
				if kind == "horizontal_scan" {
					destination = []string{"192.0.2.10", "192.0.2.11", "192.0.2.12"}[i]
				}
				if kind == "port_scan" {
					port += uint16(i)
				}
				attempts <- networksecuritycontract.FirewallAttempt{SourceIP: "192.0.2.4", DestinationIP: destination, DestinationPort: port, Protocol: "TCP"}
			}
			close(attempts)
			close(snapshots)
			coordinate(context.Background(), []*portScanPlugin{plugin}, attempts, snapshots, nil, effects, sharedlogger.NewNop())
			if len(effects) != 1 || (<-effects).Detection.Kind != kind {
				t.Fatalf("unexpected %s effects", kind)
			}
		})
	}
}

// TestReportAlertsPublishesQueuedEffect checks the outbound worker's handoff.
// Requirements: network-security/FR-004.
func TestReportAlertsPublishesQueuedEffect(t *testing.T) {
	client := &alertRecordingClient{}
	effects := make(chan alertIntent, 1)
	effects <- alertIntent{ID: "secnet_12345678", Detection: scanDetection{
		Kind: "repeated_attempts", SourceIP: "192.0.2.4", DestinationIP: "192.0.2.10", DestinationPort: 80,
		PortCount: 12, At: time.Now(),
	}, Window: time.Minute}
	close(effects)
	reportAlerts(context.Background(), effects, client, sharedlogger.NewNop())
	if client.payload.EventID != "secnet_12345678" {
		t.Fatalf("reported event = %q", client.payload.EventID)
	}
}

// TestDecodeAttemptAcceptsCanonicalObservation checks the valid NATS boundary.
// Requirements: network-security/FR-001.
func TestDecodeAttemptAcceptsCanonicalObservation(t *testing.T) {
	payload := []byte(`{"timestamp":"2026-11-01T17:18:42Z","source_ip":"192.0.2.4","destination_ip":"192.0.2.10","destination_port":80,"protocol":"TCP"}`)
	attempt, err := decodeAttempt(messaging.Envelope{Payload: payload}, validator.New(validator.WithRequiredStructEnabled()))
	if err != nil {
		t.Fatalf("decode observation: %v", err)
	}
	if attempt.DestinationPort != 80 {
		t.Fatalf("destination port = %d", attempt.DestinationPort)
	}
}
