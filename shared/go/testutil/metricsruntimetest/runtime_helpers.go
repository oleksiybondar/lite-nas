package metricsruntimetest

import (
	"context"
	"testing"
	"time"

	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// PublishCall records one publish request observed by a recording client.
type PublishCall struct {
	Subject string
	Payload any
}

// RecordingClient is a messaging.Client test double for metrics runtime tests.
type RecordingClient struct {
	PublishCalls  []PublishCall
	PublishErr    error
	PublishHook   func()
	PublishSignal chan struct{}
}

// Publish records one publish call and returns the configured publish error.
func (c *RecordingClient) Publish(_ context.Context, subject string, payload any) error {
	c.PublishCalls = append(c.PublishCalls, PublishCall{
		Subject: subject,
		Payload: payload,
	})
	c.signalPublish()

	return c.PublishErr
}

// Request is a no-op because metrics runtime tests only need publish behavior.
func (c *RecordingClient) Request(context.Context, string, any, any) error {
	return nil
}

// Drain is a no-op because these tests do not assert client draining.
func (c *RecordingClient) Drain() error {
	return nil
}

// Close is a no-op because these tests do not assert client closing.
func (c *RecordingClient) Close() {}

func (c *RecordingClient) signalPublish() {
	if c.PublishHook != nil {
		c.PublishHook()
	}
	if c.PublishSignal == nil {
		return
	}

	select {
	case c.PublishSignal <- struct{}{}:
	default:
	}
}

// RecordingServer is a messaging.Server test double for metrics runtime tests.
type RecordingServer struct {
	RPCHandlers       map[string]messaging.RPCHandler
	RegisterErr       error
	RegisterRPCErrors map[string]error
}

// Subscribe is a no-op because runtime tests only exercise RPC registration.
func (s *RecordingServer) Subscribe(string, messaging.MessageHandler) error {
	return nil
}

// RegisterRPC records one RPC handler unless a configured error should be returned.
func (s *RecordingServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if err := s.registerError(subject); err != nil {
		return err
	}

	if s.RPCHandlers == nil {
		s.RPCHandlers = map[string]messaging.RPCHandler{}
	}
	s.RPCHandlers[subject] = handler

	return nil
}

func (s *RecordingServer) registerError(subject string) error {
	if s.RegisterErr != nil {
		return s.RegisterErr
	}

	return s.RegisterRPCErrors[subject]
}

// UseSubscriptionMiddleware is a no-op because runtime tests do not exercise middleware.
func (s *RecordingServer) UseSubscriptionMiddleware(...messaging.SubscriptionMiddleware) {
}

// UseRPCMiddleware is a no-op because runtime tests do not exercise middleware.
func (s *RecordingServer) UseRPCMiddleware(...messaging.RPCMiddleware) {}

// Drain is a no-op because these tests do not assert server draining.
func (s *RecordingServer) Drain() error {
	return nil
}

// Close is a no-op because these tests do not assert server closing.
func (s *RecordingServer) Close() {}

// RecordingLogger is a logger test double for metrics runtime tests.
type RecordingLogger struct {
	Infos  []string
	Warns  []string
	Errors []string
}

// Debug is a no-op because these tests only assert info, warn, and error logs.
func (l *RecordingLogger) Debug(string, ...any) {}

// Info records one info log message.
func (l *RecordingLogger) Info(msg string, _ ...any) {
	l.Infos = append(l.Infos, msg)
}

// Warn records one warning log message.
func (l *RecordingLogger) Warn(msg string, _ ...any) {
	l.Warns = append(l.Warns, msg)
}

// Error records one error log message.
func (l *RecordingLogger) Error(msg string, _ ...any) {
	l.Errors = append(l.Errors, msg)
}

// With returns the same logger so tests can ignore structured logger scoping.
func (l *RecordingLogger) With(...any) sharedlogger.Logger {
	return l
}

// MustInvokeRPCHandler invokes a recorded RPC handler and fails the test on setup errors.
func MustInvokeRPCHandler[Response any](
	t *testing.T,
	server *RecordingServer,
	subject string,
) Response {
	t.Helper()

	handler, ok := server.RPCHandlers[subject]
	if !ok {
		t.Fatalf("RPCHandlers[%q] is not registered", subject)
	}

	response, err := handler(context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("RPCHandlers[%q]() error = %v", subject, err)
	}

	typed, ok := response.(Response)
	if !ok {
		t.Fatalf("RPCHandlers[%q]() response type = %T", subject, response)
	}

	return typed
}

// WaitForPublish blocks until a publish notification is observed or times out.
func WaitForPublish(t *testing.T, client *RecordingClient) {
	t.Helper()

	if client.PublishSignal == nil {
		t.Fatal("recording client has no publish signal channel")
	}

	timeout := time.NewTimer(time.Second)
	defer timeout.Stop()

	select {
	case <-client.PublishSignal:
	case <-timeout.C:
		t.Fatal("publish was not observed")
	}
}
