package main

import (
	"context"
	"errors"
	"testing"
	"time"

	servicestate "lite-nas/services/service-metrics/state"
	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/metricsruntimetest"
)

// Requirements: service-metrics/FR-013.
// TestServeSnapshotsPublishesCollectedEvent verifies a queued snapshot reaches
// the existing outbound subject.
func TestServeSnapshotsPublishesCollectedEvent(t *testing.T) {
	input := make(chan metrics.ServiceMetricsSnapshot, 1)
	input <- metrics.ServiceMetricsSnapshot{Timestamp: time.Unix(100, 0)}
	close(input)
	client := &metricsruntimetest.RecordingClient{}
	if err := serveSnapshots(t.Context(), input, nil, servicestate.NewHistoryStore(1), client,
		&metricsruntimetest.RecordingLogger{}); err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
	if got := client.PublishCalls[0].Subject; got != servicemetricscontract.SnapshotEventSubject {
		t.Fatalf("published subject = %q", got)
	}
}

// Requirements: service-metrics/FR-013.
// TestServeSnapshotsKeepsPublicationFailureVisible verifies the consumer
// retains the original logging policy for a failed publish.
func TestServeSnapshotsKeepsPublicationFailureVisible(t *testing.T) {
	input := make(chan metrics.ServiceMetricsSnapshot, 1)
	input <- metrics.ServiceMetricsSnapshot{}
	close(input)
	log := &metricsruntimetest.RecordingLogger{}
	client := &metricsruntimetest.RecordingClient{PublishErr: errors.New("broker unavailable")}
	if err := serveSnapshots(t.Context(), input, nil, servicestate.NewHistoryStore(1), client, log); err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
	if len(log.Errors) != 1 {
		t.Fatalf("error log count = %d, want 1", len(log.Errors))
	}
}

// Requirements: service-metrics/FR-013.
// TestServeSnapshotsStopsOnCancellation verifies the publisher worker's
// existing cancellation result.
func TestServeSnapshotsStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := serveSnapshots(ctx, make(chan metrics.ServiceMetricsSnapshot), nil,
		servicestate.NewHistoryStore(1), &metricsruntimetest.RecordingClient{},
		&metricsruntimetest.RecordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
}

// Requirements: service-metrics/FR-011.
// TestSnapshotRPCReadsPublisherStore verifies the publisher's store remains
// visible through the service's request handler.
func TestSnapshotRPCReadsPublisherStore(t *testing.T) {
	store := servicestate.NewHistoryStore(1)
	store.Add(metrics.ServiceMetricsSnapshot{Timestamp: time.Unix(100, 0)})
	server := &metricsruntimetest.RecordingServer{}
	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}
	response := metricsruntimetest.MustInvokeRPCHandler[servicemetricscontract.GetSnapshotResponse](
		t, server, servicemetricscontract.SnapshotRPCSubject)
	if !response.Available {
		t.Fatal("snapshot RPC did not find the publisher's stored snapshot")
	}
}
