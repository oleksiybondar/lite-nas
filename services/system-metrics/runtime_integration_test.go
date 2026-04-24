package main

import (
	"context"
	"reflect"
	"testing"

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

	if result.client.publishCalls[0].subject != statsEventSubject {
		t.Fatalf("publish subject = %q, want %q", result.client.publishCalls[0].subject, statsEventSubject)
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/IR-001
func TestServicePipelineStatsRPCReturnsLatestSnapshot(t *testing.T) {
	t.Parallel()

	result := runServiceCycleFixture(t)
	publishedSnapshot := extractPublishedSnapshot(t, result.client)

	response, err := result.server.rpcHandlers[statsRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("stats handler error = %v", err)
	}

	statsSnapshot, ok := response.(metrics.SystemSnapshot)
	if !ok {
		t.Fatalf("stats response type = %T, want metrics.SystemSnapshot", response)
	}

	if !reflect.DeepEqual(statsSnapshot, publishedSnapshot) {
		t.Fatalf("stats response = %#v, want %#v", statsSnapshot, publishedSnapshot)
	}
}

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-004, system-metrics-svc/IR-001
func TestServicePipelineHistoryRPCReturnsCollectedSnapshot(t *testing.T) {
	t.Parallel()

	result := runServiceCycleFixture(t)
	publishedSnapshot := extractPublishedSnapshot(t, result.client)

	response, err := result.server.rpcHandlers[historyRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("history handler error = %v", err)
	}

	history, ok := response.([]metrics.SystemSnapshot)
	if !ok {
		t.Fatalf("history response type = %T, want []metrics.SystemSnapshot", response)
	}

	wantHistory := []metrics.SystemSnapshot{publishedSnapshot}
	if !reflect.DeepEqual(history, wantHistory) {
		t.Fatalf("history = %#v, want %#v", history, wantHistory)
	}
}
