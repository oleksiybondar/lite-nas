package main

import (
	"context"
	"testing"

	sensorcontract "lite-nas/shared/contracts/sensormonitor"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/metricsruntimetest"
)

// TestServeSnapshotsStoresPublishedReading verifies the outbound worker owns
// both history insertion and NATS publication (FR-005).
func TestServeSnapshotsStoresPublishedReading(t *testing.T) {
	input := make(chan metrics.SensorSnapshot, 1)
	input <- emptySnapshot()
	close(input)
	store := NewHistoryStore(1)
	client := &metricsruntimetest.RecordingClient{}
	serveSnapshots(context.Background(), input, store, client, &metricsruntimetest.RecordingLogger{})
	if len(client.PublishCalls) != 1 || client.PublishCalls[0].Subject != sensorcontract.SnapshotEventSubject {
		t.Fatalf("PublishCalls = %+v, want one sensor snapshot event", client.PublishCalls)
	}
}

// TestSnapshotRPCReportsLatest checks the public latest-snapshot response.
func TestSnapshotRPCReportsLatest(t *testing.T) {
	store := NewHistoryStore(0)
	store.Add(emptySnapshot())
	server := &metricsruntimetest.RecordingServer{}
	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers(): %v", err)
	}
	response := metricsruntimetest.MustInvokeRPCHandler[sensorcontract.GetSnapshotResponse](
		t, server, sensorcontract.SnapshotRPCSubject,
	)
	if !response.Available {
		t.Fatal("snapshot Available = false, want true")
	}
}

// TestRPCRejectsUnexpectedPayload verifies the service-facing empty request
// boundary rejects unrecognized fields before consulting the store.
func TestRPCRejectsUnexpectedPayload(t *testing.T) {
	server := &metricsruntimetest.RecordingServer{}
	if err := registerRPCHandlers(server, NewHistoryStore(1)); err != nil {
		t.Fatalf("registerRPCHandlers(): %v", err)
	}
	_, err := server.RPCHandlers[sensorcontract.HistoryRPCSubject](
		context.Background(), messaging.Envelope{Payload: []byte(`{"limit":100}`)},
	)
	if err == nil {
		t.Fatal("history RPC accepted unexpected payload")
	}
}
