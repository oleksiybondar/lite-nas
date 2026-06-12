package services

import (
	"context"
	"reflect"
	"testing"
	"time"

	"lite-nas/shared/messaging"
)

// snapshotGetter describes one service that can fetch a snapshot value.
type snapshotGetter[T any] interface {
	GetSnapshot(context.Context) (T, error)
}

// historyGetter describes one service that can fetch history values.
type historyGetter[T any] interface {
	GetHistory(context.Context) ([]T, error)
}

// metricsClientStub records one RPC subject and request while allowing tests to inject responses.
type metricsClientStub struct {
	subject     string
	request     any
	requestFunc func(context.Context, string, any, any) error
}

// newSnapshotClientStub builds one client stub that populates a snapshot RPC response.
func newSnapshotClientStub[T any, R any](
	t *testing.T,
	want T,
	assign func(*R, T),
) *metricsClientStub {
	t.Helper()

	return &metricsClientStub{
		requestFunc: func(_ context.Context, _ string, _ any, response any) error {
			typed, ok := response.(*R)
			if !ok {
				t.Fatalf("response type = %T, want %T", response, new(R))
			}
			assign(typed, want)
			return nil
		},
	}
}

// newHistoryClientStub builds one client stub that populates a history RPC response.
func newHistoryClientStub[T any, R any](
	t *testing.T,
	want []T,
	assign func(*R, []T),
) *metricsClientStub {
	t.Helper()

	return &metricsClientStub{
		requestFunc: func(_ context.Context, _ string, _ any, response any) error {
			typed, ok := response.(*R)
			if !ok {
				t.Fatalf("response type = %T, want %T", response, new(R))
			}
			assign(typed, want)
			return nil
		},
	}
}

// assertMetricsSubject verifies one service invoked the expected RPC subject.
func assertMetricsSubject(t *testing.T, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("subject = %q, want %q", got, want)
	}
}

// assertMetricsResult verifies one service returned the expected value.
func assertMetricsResult[T any](t *testing.T, methodName string, got T, want T) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v, want %#v", methodName, got, want)
	}
}

// mustGetSnapshot runs one snapshot request and fails the test if the service returns an error.
func mustGetSnapshot[T any, S snapshotGetter[T]](t *testing.T, service S) T {
	t.Helper()

	got, err := service.GetSnapshot(context.Background())
	if err != nil {
		var zero T
		t.Fatalf("GetSnapshot() error = %v", err)
		return zero
	}

	return got
}

// mustGetHistory runs one history request and fails the test if the service returns an error.
func mustGetHistory[T any, S historyGetter[T]](t *testing.T, service S) []T {
	t.Helper()

	got, err := service.GetHistory(context.Background())
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
		return nil
	}

	return got
}

// unixFromSeconds converts one Unix timestamp into a Time value for test fixtures.
func unixFromSeconds(unixSeconds int64) time.Time {
	return time.Unix(unixSeconds, 0)
}

// Publish satisfies the messaging client contract for tests that only exercise request/reply flows.
func (c *metricsClientStub) Publish(context.Context, string, any) error {
	return nil
}

// Request records one RPC invocation and forwards control to the injected response hook.
func (c *metricsClientStub) Request(ctx context.Context, subject string, request any, response any) error {
	c.subject = subject
	c.request = request
	if c.requestFunc == nil {
		return nil
	}
	return c.requestFunc(ctx, subject, request, response)
}

// Drain satisfies the messaging client contract for tests that do not exercise lifecycle behavior.
func (c *metricsClientStub) Drain() error {
	return nil
}

// Close satisfies the messaging client contract for tests that do not exercise lifecycle behavior.
func (c *metricsClientStub) Close() {}

var _ messaging.Client = (*metricsClientStub)(nil)
