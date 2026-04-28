package services

import (
	"context"
	"reflect"
	"testing"
	"time"

	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/metrics"
)

type recordingRequestClient struct {
	subject  string
	response any
}

func (c *recordingRequestClient) Publish(context.Context, string, any) error {
	return nil
}

func (c *recordingRequestClient) Request(_ context.Context, subject string, _ any, response any) error {
	c.subject = subject

	switch out := response.(type) {
	case *systemmetricscontract.GetSnapshotResponse:
		*out = c.response.(systemmetricscontract.GetSnapshotResponse)
	case *systemmetricscontract.GetHistoryResponse:
		*out = c.response.(systemmetricscontract.GetHistoryResponse)
	}

	return nil
}

func (c *recordingRequestClient) Drain() error {
	return nil
}

func (c *recordingRequestClient) Close() {}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestSystemMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := metrics.SystemSnapshot{Timestamp: time.Unix(100, 0)}
	client := &recordingRequestClient{
		response: systemmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  want,
		},
	}
	service := NewSystemMetricsService(client)

	got, err := service.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	if client.subject != systemmetricscontract.SnapshotRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, systemmetricscontract.SnapshotRPCSubject)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetSnapshot() = %#v, want %#v", got, want)
	}
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestSystemMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.SystemSnapshot{
		{Timestamp: time.Unix(100, 0)},
		{Timestamp: time.Unix(101, 0)},
	}
	client := &recordingRequestClient{
		response: systemmetricscontract.GetHistoryResponse{
			Items: want,
		},
	}
	service := NewSystemMetricsService(client)

	got, err := service.GetHistory(context.Background())
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if client.subject != systemmetricscontract.HistoryRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, systemmetricscontract.HistoryRPCSubject)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetHistory() = %#v, want %#v", got, want)
	}
}
