package messaging

import (
	"context"
	"errors"
	"net/textproto"
	"testing"

	"lite-nas/shared/config"

	"github.com/nats-io/nats.go"
)

func TestNewServerRejectsNilCodec(t *testing.T) {
	t.Parallel()

	_, err := NewServer(config.MessagingConfig{}, &recordingLogger{}, nil)
	assertErrorIs(t, err, ErrInvalidConfig)
}

func TestNewServerReturnsConnectionError(t *testing.T) {
	t.Parallel()

	_, err := NewServer(config.MessagingConfig{
		URL:  "nats://localhost:4222",
		Cert: "/missing/client.crt",
		Key:  "/missing/client.key",
	}, &recordingLogger{}, stubCodec{})
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestServerSubscribeRejectsNilHandler(t *testing.T) {
	t.Parallel()

	err := (&server{connection: &connection{}}).Subscribe("subject", nil)
	assertErrorIs(t, err, ErrHandlerFailed)
}

func TestServerRegisterRPCRejectsNilHandler(t *testing.T) {
	t.Parallel()

	err := (&server{connection: &connection{}}).RegisterRPC("subject", nil)
	assertErrorIs(t, err, ErrHandlerFailed)
}

func TestServerDrainDelegatesToConnection(t *testing.T) {
	t.Parallel()

	if err := (&server{connection: &connection{}}).Drain(); err != nil {
		t.Fatalf("Drain() error = %v", err)
	}
}

func TestServerCloseDelegatesToConnection(t *testing.T) {
	t.Parallel()

	(&server{connection: &connection{}}).Close()
}

func TestServerHandleMessageReturnsHandlerError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("handler failed")
	err := (&server{}).handleMessage(func(context.Context, Envelope) error {
		return expectedErr
	}, Envelope{Subject: "system.metrics"})
	if !errors.Is(err, ErrHandlerFailed) {
		t.Fatalf("handleMessage() error = %v, want wrapped ErrHandlerFailed", err)
	}
}

func TestServerHandleMessageReturnsNilOnSuccess(t *testing.T) {
	t.Parallel()

	err := (&server{}).handleMessage(func(context.Context, Envelope) error {
		return nil
	}, Envelope{Subject: "system.metrics"})
	if err != nil {
		t.Fatalf("handleMessage() error = %v", err)
	}
}

func TestServerHandleRPCRejectsMissingReplySubject(t *testing.T) {
	t.Parallel()

	err := (&server{codec: stubCodec{}, connection: &connection{}}).handleRPC(
		func(context.Context, Envelope) (any, error) { return "ok", nil },
		Envelope{Subject: "system.metrics"},
	)
	if !errors.Is(err, ErrHandlerFailed) {
		t.Fatalf("handleRPC() error = %v, want wrapped ErrHandlerFailed", err)
	}
}

func TestServerHandleRPCReturnsHandlerError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("handler failed")
	err := (&server{codec: stubCodec{}, connection: &connection{}}).handleRPC(
		func(context.Context, Envelope) (any, error) { return nil, expectedErr },
		Envelope{Subject: "system.metrics", ReplyTo: "_INBOX.reply"},
	)
	if !errors.Is(err, ErrHandlerFailed) {
		t.Fatalf("handleRPC() error = %v, want wrapped ErrHandlerFailed", err)
	}
}

func TestServerHandleRPCReturnsEncodeError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("marshal failed")
	err := (&server{
		codec:      codecStub{marshalErr: expectedErr},
		connection: &connection{},
	}).handleRPC(
		func(context.Context, Envelope) (any, error) { return "ok", nil },
		Envelope{Subject: "system.metrics", ReplyTo: "_INBOX.reply"},
	)
	if !errors.Is(err, ErrEncodeFailed) {
		t.Fatalf("handleRPC() error = %v, want wrapped ErrEncodeFailed", err)
	}
}

func TestServerHandleRPCReturnsPublishError(t *testing.T) {
	t.Parallel()

	err := (&server{
		codec:      stubCodec{},
		connection: &connection{},
	}).handleRPC(
		func(context.Context, Envelope) (any, error) { return "ok", nil },
		Envelope{Subject: "system.metrics", ReplyTo: "_INBOX.reply"},
	)
	assertErrorIs(t, err, ErrNotConnected)
}

func TestServerBuildMessageHandlerLogsHandlerErrors(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	srv := &server{logger: log}
	handler := srv.buildMessageHandler(func(context.Context, Envelope) error {
		return errors.New("handler failed")
	})

	handler(&nats.Msg{Subject: "system.metrics", Data: []byte("payload")})

	assertSingleLogEntry(t, log, logEntry{
		level: "error",
		msg:   "message handler failed",
		args:  []any{"subject", "system.metrics", "error", "messaging: handler failed: handler failed"},
	})
}

func TestServerBuildRPCHandlerLogsHandlerErrors(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	srv := &server{
		codec:      stubCodec{},
		connection: &connection{},
		logger:     log,
	}
	handler := srv.buildRPCHandler(func(context.Context, Envelope) (any, error) {
		return nil, errors.New("handler failed")
	})

	handler(&nats.Msg{Subject: "system.metrics", Reply: "_INBOX.reply", Data: []byte("payload")})

	assertSingleLogEntry(t, log, logEntry{
		level: "error",
		msg:   "rpc handler failed",
		args:  []any{"subject", "system.metrics", "reply_to", "_INBOX.reply", "error", "messaging: handler failed: handler failed"},
	})
}

func TestNewEnvelopeFromMessageCopiesTransportFields(t *testing.T) {
	t.Parallel()

	msg := &nats.Msg{
		Subject: "system.metrics",
		Reply:   "_INBOX.reply",
		Data:    []byte("payload"),
		Header: nats.Header{
			textproto.CanonicalMIMEHeaderKey("Trace-ID"): []string{"abc123"},
		},
	}

	envelope := newEnvelopeFromMessage(msg)
	if envelope.Subject != "system.metrics" {
		t.Fatalf("Subject = %q, want system.metrics", envelope.Subject)
	}

	if envelope.ReplyTo != "_INBOX.reply" {
		t.Fatalf("ReplyTo = %q, want _INBOX.reply", envelope.ReplyTo)
	}

	if string(envelope.Payload) != "payload" {
		t.Fatalf("Payload = %q, want payload", string(envelope.Payload))
	}

	if envelope.Headers["Trace-Id"] != "abc123" {
		t.Fatalf("Headers = %#v, want Trace-Id", envelope.Headers)
	}
}

func TestNewHeadersFromMessageReturnsNilWithoutHeaders(t *testing.T) {
	t.Parallel()

	if headers := newHeadersFromMessage(nil); headers != nil {
		t.Fatalf("newHeadersFromMessage(nil) = %#v, want nil", headers)
	}

	if headers := newHeadersFromMessage(&nats.Msg{}); headers != nil {
		t.Fatalf("newHeadersFromMessage(empty) = %#v, want nil", headers)
	}
}

func TestNewHeadersFromMessageSkipsEmptyHeaderValues(t *testing.T) {
	t.Parallel()

	headers := newHeadersFromMessage(&nats.Msg{
		Header: nats.Header{
			"Trace-ID": {},
			"Type":     {"json", "ignored"},
		},
	})

	if len(headers) != 1 {
		t.Fatalf("len(headers) = %d, want 1", len(headers))
	}

	if headers["Type"] != "json" {
		t.Fatalf("headers = %#v, want Type=json", headers)
	}
}
