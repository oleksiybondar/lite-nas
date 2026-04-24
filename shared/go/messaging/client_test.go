package messaging

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/shared/config"

	"github.com/nats-io/nats.go"
)

func TestNewClientRejectsNilCodec(t *testing.T) {
	t.Parallel()

	_, err := NewClient(config.MessagingConfig{Timeout: time.Second}, &recordingLogger{}, nil)
	assertErrorIs(t, err, ErrInvalidConfig)
}

func TestNewClientRejectsNonPositiveTimeout(t *testing.T) {
	t.Parallel()

	_, err := NewClient(config.MessagingConfig{Timeout: 0}, &recordingLogger{}, stubCodec{})
	assertErrorIs(t, err, ErrInvalidConfig)
}

func TestNewClientReturnsConnectionError(t *testing.T) {
	t.Parallel()

	_, err := NewClient(config.MessagingConfig{
		URL:     "nats://localhost:4222",
		Timeout: time.Second,
		Cert:    "/missing/client.crt",
		Key:     "/missing/client.key",
	}, &recordingLogger{}, stubCodec{})
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestClientPublishReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := (&client{codec: stubCodec{}, connection: &connection{}}).Publish(ctx, "subject", "payload")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Publish() error = %v, want %v", err, context.Canceled)
	}
}

func TestClientPublishReturnsEncodeError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("marshal failed")
	err := (&client{
		codec:      codecStub{marshalErr: expectedErr},
		connection: &connection{},
	}).Publish(context.Background(), "subject", "payload")

	if !errors.Is(err, ErrEncodeFailed) {
		t.Fatalf("Publish() error = %v, want wrapped ErrEncodeFailed", err)
	}
}

func TestClientPublishReturnsConnectionError(t *testing.T) {
	t.Parallel()

	err := (&client{
		codec:      stubCodec{},
		connection: &connection{},
	}).Publish(context.Background(), "subject", "payload")
	assertErrorIs(t, err, ErrNotConnected)
}

func TestClientRequestReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := (&client{
		codec:      stubCodec{},
		connection: &connection{},
		timeout:    time.Second,
	}).Request(ctx, "subject", "request", &struct{}{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Request() error = %v, want %v", err, context.Canceled)
	}
}

func TestClientRequestReturnsDeadlineExceededForExpiredContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	err := (&client{
		codec:      stubCodec{},
		connection: &connection{},
		timeout:    time.Second,
	}).Request(ctx, "subject", "request", &struct{}{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Request() error = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestClientRequestReturnsEncodeError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("marshal failed")
	err := (&client{
		codec:      codecStub{marshalErr: expectedErr},
		connection: &connection{},
		timeout:    time.Second,
	}).Request(context.Background(), "subject", "request", &struct{}{})

	if !errors.Is(err, ErrEncodeFailed) {
		t.Fatalf("Request() error = %v, want wrapped ErrEncodeFailed", err)
	}
}

func TestClientRequestReturnsConnectionError(t *testing.T) {
	t.Parallel()

	err := (&client{
		codec:      stubCodec{},
		connection: &connection{},
		timeout:    time.Second,
	}).Request(context.Background(), "subject", "request", &struct{}{})
	assertErrorIs(t, err, ErrNotConnected)
}

func TestClientDrainDelegatesToConnection(t *testing.T) {
	t.Parallel()

	if err := (&client{connection: &connection{}}).Drain(); err != nil {
		t.Fatalf("Drain() error = %v", err)
	}
}

func TestClientCloseDelegatesToConnection(t *testing.T) {
	t.Parallel()

	(&client{connection: &connection{}}).Close()
}

func TestClientResolveTimeoutReturnsConfiguredTimeoutWithoutDeadline(t *testing.T) {
	t.Parallel()

	timeout, err := (&client{timeout: 3 * time.Second}).resolveTimeout(context.Background())
	if err != nil {
		t.Fatalf("resolveTimeout() error = %v", err)
	}

	if timeout != 3*time.Second {
		t.Fatalf("resolveTimeout() = %v, want 3s", timeout)
	}
}

func TestClientResolveTimeoutReturnsRemainingDeadline(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(200*time.Millisecond))
	defer cancel()

	timeout, err := (&client{timeout: time.Second}).resolveTimeout(ctx)
	if err != nil {
		t.Fatalf("resolveTimeout() error = %v", err)
	}

	if timeout <= 0 || timeout > 200*time.Millisecond {
		t.Fatalf("resolveTimeout() = %v, want remaining positive deadline", timeout)
	}
}

func TestClientResolveTimeoutReturnsDeadlineExceededForExpiredDeadline(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Millisecond))
	defer cancel()

	_, err := (&client{timeout: time.Second}).resolveTimeout(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("resolveTimeout() error = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestClientRequestReturnsDecodeError(t *testing.T) {
	t.Parallel()

	reply := &nats.Msg{Data: []byte("reply")}
	err := requestAndDecodeWithCodec(codecStub{unmarshalErr: errors.New("decode failed")}, reply, &struct{}{})
	if !errors.Is(err, ErrDecodeFailed) {
		t.Fatalf("requestAndDecodeWithCodec() error = %v, want wrapped ErrDecodeFailed", err)
	}
}

func requestAndDecodeWithCodec(codec Codec, reply *nats.Msg, response any) error {
	if err := codec.Unmarshal(reply.Data, response); err != nil {
		return wrapDecodeError(err)
	}

	return nil
}
