package services

import (
	"context"
	"reflect"
	"testing"

	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	"lite-nas/shared/metrics"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestZFSMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := zfsServiceSnapshotFixture(100)
	client := &zfsMetricsClientStub{
		requestFunc: func(_ context.Context, subject string, request any, response any) error {
			typed, ok := response.(*zfsmetricscontract.GetSnapshotResponse)
			if !ok {
				t.Fatalf("response type = %T, want *GetSnapshotResponse", response)
			}
			typed.Snapshot = want
			return nil
		},
	}
	service := NewZFSMetricsService(client)

	got, err := service.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	if client.subject != zfsmetricscontract.SnapshotRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, zfsmetricscontract.SnapshotRPCSubject)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetSnapshot() = %#v, want %#v", got, want)
	}
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestZFSMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.ZFSSnapshot{
		zfsServiceSnapshotFixture(100),
		zfsServiceSnapshotFixture(101),
	}
	client := &zfsMetricsClientStub{
		requestFunc: func(_ context.Context, subject string, request any, response any) error {
			typed, ok := response.(*zfsmetricscontract.GetHistoryResponse)
			if !ok {
				t.Fatalf("response type = %T, want *GetHistoryResponse", response)
			}
			typed.Items = want
			return nil
		},
	}
	service := NewZFSMetricsService(client)

	got, err := service.GetHistory(context.Background())
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if client.subject != zfsmetricscontract.HistoryRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, zfsmetricscontract.HistoryRPCSubject)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetHistory() = %#v, want %#v", got, want)
	}
}

func zfsServiceSnapshotFixture(unixSeconds int64) metrics.ZFSSnapshot {
	return metrics.ZFSSnapshot{Timestamp: unixFromSeconds(unixSeconds)}
}

type zfsMetricsClientStub struct {
	subject     string
	requestFunc func(context.Context, string, any, any) error
}

func (c *zfsMetricsClientStub) Publish(context.Context, string, any) error {
	return nil
}

func (c *zfsMetricsClientStub) Request(ctx context.Context, subject string, request any, response any) error {
	c.subject = subject
	if c.requestFunc == nil {
		return nil
	}
	return c.requestFunc(ctx, subject, request, response)
}

func (c *zfsMetricsClientStub) Drain() error {
	return nil
}

func (c *zfsMetricsClientStub) Close() {}
