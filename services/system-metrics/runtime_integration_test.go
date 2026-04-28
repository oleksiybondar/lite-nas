package main

import (
	"context"
	"reflect"
	"testing"

	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-svc/FR-001, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestServicePipelinePublishesSnapshotEvent(t *testing.T) {
	t.Parallel()

	result := runServiceCycleFixture(t)

	if len(result.client.publishCalls) != 1 {
		t.Fatalf("publishCalls = %d, want 1", len(result.client.publishCalls))
	}

	if result.client.publishCalls[0].subject != systemmetricscontract.SnapshotEventSubject {
		t.Fatalf("publish subject = %q, want %q", result.client.publishCalls[0].subject, systemmetricscontract.SnapshotEventSubject)
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/IR-001
func TestServicePipelineStatsRPCReturnsLatestSnapshot(t *testing.T) {
	t.Parallel()

	result := runServiceCycleFixture(t)
	publishedSnapshot := extractPublishedSnapshot(t, result.client)

	response, err := result.server.rpcHandlers[systemmetricscontract.SnapshotRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("stats handler error = %v", err)
	}

	snapshotResponse, ok := response.(systemmetricscontract.GetSnapshotResponse)
	if !ok {
		t.Fatalf("stats response type = %T, want systemmetrics.GetSnapshotResponse", response)
	}

	if !snapshotResponse.Available {
		t.Fatal("stats response Available = false, want true")
	}

	if !reflect.DeepEqual(snapshotResponse.Snapshot, publishedSnapshot) {
		t.Fatalf("stats response = %#v, want %#v", snapshotResponse.Snapshot, publishedSnapshot)
	}
}

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-004, system-metrics-svc/IR-001
func TestServicePipelineHistoryRPCReturnsCollectedSnapshot(t *testing.T) {
	t.Parallel()

	result := runServiceCycleFixture(t)
	publishedSnapshot := extractPublishedSnapshot(t, result.client)

	response, err := result.server.rpcHandlers[systemmetricscontract.HistoryRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("history handler error = %v", err)
	}

	historyResponse, ok := response.(systemmetricscontract.GetHistoryResponse)
	if !ok {
		t.Fatalf("history response type = %T, want systemmetrics.GetHistoryResponse", response)
	}

	wantHistory := []metrics.SystemSnapshot{publishedSnapshot}
	if !reflect.DeepEqual(historyResponse.Items, wantHistory) {
		t.Fatalf("history = %#v, want %#v", historyResponse.Items, wantHistory)
	}
}
