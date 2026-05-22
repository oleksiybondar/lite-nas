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

func TestServerHandleMessageRunsSubscriptionMiddlewaresInOrder(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 3)
	srv := &server{subscriptionMiddlewares: []SubscriptionMiddleware{
		traceSubscriptionMiddleware("mw1", &order),
		traceSubscriptionMiddleware("mw2", &order),
	}}

	err := srv.handleMessage(recordMessageHandler(&order), Envelope{Subject: "system.metrics"})
	if err != nil {
		t.Fatalf("handleMessage() error = %v", err)
	}

	assertEqualSlice(t, order, []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	})
}

func TestServerHandleMessageStopsOnMiddlewareError(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 1)
	srv := &server{
		subscriptionMiddlewares: []SubscriptionMiddleware{
			func(
				_ context.Context,
				_ Envelope,
				_ MessageNext,
			) error {
				order = append(order, "mw")
				return errors.New("blocked")
			},
		},
	}

	err := srv.handleMessage(func(context.Context, Envelope) error {
		t.Fatal("handler should not be called")
		return nil
	}, Envelope{Subject: "system.metrics"})
	if !errors.Is(err, ErrHandlerFailed) {
		t.Fatalf("handleMessage() error = %v, want wrapped ErrHandlerFailed", err)
	}

	assertEqualSlice(t, order, []string{"mw"})
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

func TestServerHandleRPCRunsMiddlewaresInOrder(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 3)
	srv := &server{
		codec:      stubCodec{},
		connection: &connection{},
		rpcMiddlewares: []RPCMiddleware{
			traceRPCMiddleware("mw1", &order),
			traceRPCMiddleware("mw2", &order),
		},
	}

	err := srv.handleRPC(
		recordRPCHandler(&order),
		Envelope{Subject: "system.metrics", ReplyTo: "_INBOX.reply"},
	)
	assertErrorIs(t, err, ErrNotConnected)

	assertEqualSlice(t, order, []string{
		"mw1-before",
		"mw2-before",
		"handler",
		"mw2-after",
		"mw1-after",
	})
}

func TestServerHandleRPCStopsOnMiddlewareError(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 1)
	srv := &server{
		codec:      stubCodec{},
		connection: &connection{},
		rpcMiddlewares: []RPCMiddleware{
			func(
				_ context.Context,
				_ Envelope,
				_ RPCNext,
			) (any, error) {
				order = append(order, "mw")
				return nil, errors.New("blocked")
			},
		},
	}

	err := srv.handleRPC(
		func(context.Context, Envelope) (any, error) {
			t.Fatal("handler should not be called")
			return "ok", nil
		},
		Envelope{Subject: "system.metrics", ReplyTo: "_INBOX.reply"},
	)
	if !errors.Is(err, ErrHandlerFailed) {
		t.Fatalf("handleRPC() error = %v, want wrapped ErrHandlerFailed", err)
	}
	assertEqualSlice(t, order, []string{"mw"})
}

func TestServerUseMiddlewareAppendsMiddlewares(t *testing.T) {
	t.Parallel()

	srv := &server{}
	srv.UseSubscriptionMiddleware(
		func(ctx context.Context, envelope Envelope, next MessageNext) error {
			return next(ctx, envelope)
		},
	)
	srv.UseRPCMiddleware(
		func(ctx context.Context, envelope Envelope, next RPCNext) (any, error) {
			return next(ctx, envelope)
		},
	)

	if len(srv.subscriptionMiddlewares) != 1 {
		t.Fatalf("len(subscriptionMiddlewares) = %d, want 1", len(srv.subscriptionMiddlewares))
	}
	if len(srv.rpcMiddlewares) != 1 {
		t.Fatalf("len(rpcMiddlewares) = %d, want 1", len(srv.rpcMiddlewares))
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

func assertEqualSlice(t *testing.T, actual []string, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("len(actual) = %d, want %d (%v)", len(actual), len(expected), actual)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("actual[%d] = %q, want %q (%v)", index, actual[index], expected[index], actual)
		}
	}
}

func traceSubscriptionMiddleware(name string, order *[]string) SubscriptionMiddleware {
	return func(ctx context.Context, envelope Envelope, next MessageNext) error {
		*order = append(*order, name+"-before")
		err := next(ctx, envelope)
		*order = append(*order, name+"-after")
		return err
	}
}

func recordMessageHandler(order *[]string) MessageHandler {
	return func(context.Context, Envelope) error {
		*order = append(*order, "handler")
		return nil
	}
}

func traceRPCMiddleware(name string, order *[]string) RPCMiddleware {
	return func(ctx context.Context, envelope Envelope, next RPCNext) (any, error) {
		*order = append(*order, name+"-before")
		response, err := next(ctx, envelope)
		*order = append(*order, name+"-after")
		return response, err
	}
}

func recordRPCHandler(order *[]string) RPCHandler {
	return func(context.Context, Envelope) (any, error) {
		*order = append(*order, "handler")
		return "ok", nil
	}
}
