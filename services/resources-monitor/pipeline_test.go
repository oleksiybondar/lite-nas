package main

import (
	"context"
	"errors"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	sensorcontract "lite-nas/shared/contracts/sensormonitor"
	"lite-nas/shared/messaging"
)

// TestSubscriptionQueuesSnapshot verifies the NATS adapter sends one
// decoded event to the worker input channel.
func TestSubscriptionQueuesSnapshot(t *testing.T) {
	server := &stubServer{}
	input := make(chan snapshotInput, 1)
	if err := registerSubscriptions(server, input, t.Context()); err != nil {
		t.Fatalf("registerSubscriptions() error = %v", err)
	}
	envelope := messaging.Envelope{Subject: networkmetricscontract.SnapshotEventSubject, Payload: []byte(`{"snapshot":{"timestamp":"2026-10-30T00:00:00Z"}}`)}
	if err := server.handlers[envelope.Subject](t.Context(), envelope); err != nil {
		t.Fatalf("subscription handler error = %v", err)
	}
	if got := (<-input).subject; got != envelope.Subject {
		t.Fatalf("queued subject = %q, want %q", got, envelope.Subject)
	}
}

// TestSensorSubscriptionRejectsMissingTimestamp verifies the sensor adapter
// validates the public event before queueing rule work.
// Requirements: resources-monitor/SR-001.
func TestSensorSubscriptionRejectsMissingTimestamp(t *testing.T) {
	server := &stubServer{}
	input := make(chan snapshotInput, 1)
	if err := registerSubscriptions(server, input, t.Context()); err != nil {
		t.Fatalf("registerSubscriptions(): %v", err)
	}
	err := server.handlers[sensorcontract.SnapshotEventSubject](t.Context(), messaging.Envelope{
		Subject: sensorcontract.SnapshotEventSubject,
		Payload: []byte(`{"snapshot":{"voltages":[{"key":"EXT5V_V","name":"supply","value":4.8,"unit":"V"}]}}`),
	})
	if err == nil {
		t.Fatal("sensor subscription accepted a snapshot without timestamp")
	}
}

// TestSensorSubscriptionQueuesValidSnapshot verifies a typed sensor event
// reaches the processor worker input.
// Requirements: resources-monitor/FR-006.
func TestSensorSubscriptionQueuesValidSnapshot(t *testing.T) {
	server := &stubServer{}
	input := make(chan snapshotInput, 1)
	if err := registerSubscriptions(server, input, t.Context()); err != nil {
		t.Fatalf("registerSubscriptions(): %v", err)
	}
	envelope := messaging.Envelope{
		Subject: sensorcontract.SnapshotEventSubject,
		Payload: []byte(`{"snapshot":{"timestamp":"2026-11-11T00:00:00Z","voltages":[{"key":"EXT5V_V","name":"supply","value":4.8,"unit":"V"}]}}`),
	}
	if err := server.handlers[envelope.Subject](t.Context(), envelope); err != nil {
		t.Fatalf("sensor subscription: %v", err)
	}
	if got := (<-input).subject; got != sensorcontract.SnapshotEventSubject {
		t.Fatalf("queued subject = %q, want sensor subject", got)
	}
}

// TestSubscriptionAcceptsMissingSnapshot preserves the prior JSON acceptance
// behavior while routing events through the worker queue.
func TestSubscriptionAcceptsMissingSnapshot(t *testing.T) {
	server := &stubServer{}
	input := make(chan snapshotInput, 1)
	if err := registerSubscriptions(server, input, t.Context()); err != nil {
		t.Fatalf("registerSubscriptions() error = %v", err)
	}
	envelope := messaging.Envelope{Subject: networkmetricscontract.SnapshotEventSubject, Payload: []byte(`{}`)}
	if err := server.handlers[envelope.Subject](t.Context(), envelope); err != nil {
		t.Fatalf("subscription handler rejected a JSON object: %v", err)
	}
	if got := (<-input).payload; len(got) != 0 {
		t.Fatalf("queued payload = %v, want empty object", got)
	}
}

// TestSubscriptionRejectsMalformedJSON preserves the prior decoding failure.
func TestSubscriptionRejectsMalformedJSON(t *testing.T) {
	server := &stubServer{}
	input := make(chan snapshotInput, 1)
	if err := registerSubscriptions(server, input, t.Context()); err != nil {
		t.Fatalf("registerSubscriptions() error = %v", err)
	}
	envelope := messaging.Envelope{Subject: networkmetricscontract.SnapshotEventSubject, Payload: []byte(`{"snapshot":`)}
	if err := server.handlers[envelope.Subject](t.Context(), envelope); err == nil {
		t.Fatal("subscription handler accepted malformed JSON")
	}
}

// TestAlertSenderAcknowledgesOutboundDelivery verifies the processor waits for
// the outbound worker before treating an alert as sent.
func TestAlertSenderAcknowledgesOutboundDelivery(t *testing.T) {
	effects := make(chan alertIntent, 1)
	sender := channelAlertSender{effects: effects}
	done := make(chan error, 1)
	go func() {
		done <- sender.Create(t.Context(), loggingmanagercontract.AlertPayload{EventID: "syscpu_00000001"})
	}()
	deliverAlertIntent(t.Context(), pipelineClient{}, <-effects)
	if err := <-done; err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// TestAlertSenderReturnsOutboundFailure verifies delivery errors remain visible
// to the lifecycle processor.
func TestAlertSenderReturnsOutboundFailure(t *testing.T) {
	effects := make(chan alertIntent, 1)
	sender := channelAlertSender{effects: effects}
	done := make(chan error, 1)
	go func() {
		done <- sender.Create(t.Context(), loggingmanagercontract.AlertPayload{EventID: "syscpu_00000001"})
	}()
	deliverAlertIntent(t.Context(), pipelineClient{publishErr: errors.New("publish failed")}, <-effects)
	if err := <-done; err == nil {
		t.Fatal("Create() error = nil, want publication failure")
	}
}

// pipelineClient records only the delivery behavior needed by worker tests.
type pipelineClient struct{ publishErr error }

func (client pipelineClient) Publish(context.Context, string, any) error { return client.publishErr }
func (pipelineClient) Request(context.Context, string, any, any) error   { return nil }
func (pipelineClient) Drain() error                                      { return nil }
func (pipelineClient) Close()                                            {}
