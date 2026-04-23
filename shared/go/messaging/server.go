package messaging

import (
	"context"
	"fmt"

	"lite-nas/shared/config"
	"lite-nas/shared/logger"

	"github.com/nats-io/nats.go"
)

// MessageHandler handles an inbound message represented as a transport-level
// envelope.
//
// Design choices:
//   - handlers receive Envelope instead of raw NATS types
//   - payload decoding is left to the application layer
type MessageHandler func(ctx context.Context, envelope Envelope) error

// RPCHandler handles an inbound RPC request represented as a transport-level
// envelope and returns a response payload to be serialized by the server.
//
// Design choices:
//   - request metadata remains available through Envelope
//   - response encoding is handled by the shared Codec
type RPCHandler func(ctx context.Context, envelope Envelope) (any, error)

// Server defines inbound messaging operations built on top of the low-level
// transport connection.
//
// Design choices:
//   - the server is responsible only for inbound flows
//   - explicit registration keeps subject ownership visible at app/service level
//   - the shared server does not introduce routing magic or orchestration
type Server interface {
	// Subscribe registers a handler for messages received on the provided subject.
	Subscribe(subject string, handler MessageHandler) error

	// RegisterRPC registers a request/reply handler for the provided subject.
	RegisterRPC(subject string, handler RPCHandler) error

	// Drain gracefully drains the underlying transport connection.
	Drain() error

	// Close immediately closes the underlying transport connection.
	Close()
}

// server is the default inbound messaging implementation.
//
// Design choices:
//   - it composes the tested low-level connection instead of exposing NATS types
//   - it converts raw NATS messages into Envelope before invoking handlers
//   - it uses the injected Codec only where serialization is required
type server struct {
	connection *connection
	codec      Codec
	logger     logger.Logger
}

// NewServer creates a new inbound messaging server.
//
// It establishes the low-level transport connection and uses the provided Codec
// for RPC response serialization.
func NewServer(
	cfg config.MessagingConfig,
	log logger.Logger,
	codec Codec,
) (Server, error) {
	if codec == nil {
		return nil, ErrInvalidConfig
	}

	conn, err := newConnection(cfg, log)
	if err != nil {
		return nil, err
	}

	return &server{
		connection: conn,
		codec:      codec,
		logger:     log,
	}, nil
}

// Subscribe registers a message handler for the provided subject.
func (s *server) Subscribe(subject string, handler MessageHandler) error {
	if handler == nil {
		return ErrHandlerFailed
	}

	return s.connection.subscribe(subject, s.buildMessageHandler(handler))
}

// RegisterRPC registers a request/reply handler for the provided subject.
func (s *server) RegisterRPC(subject string, handler RPCHandler) error {
	if handler == nil {
		return ErrHandlerFailed
	}

	return s.connection.subscribe(subject, s.buildRPCHandler(handler))
}

// Drain gracefully drains the underlying transport connection.
func (s *server) Drain() error {
	return s.connection.drain()
}

// Close immediately closes the underlying transport connection.
func (s *server) Close() {
	s.connection.close()
}

// buildMessageHandler adapts a MessageHandler into the callback shape required
// by the low-level NATS subscription API.
//
// Design choice:
//   - the adapter stays thin
//   - actual processing and logging are delegated to named methods
func (s *server) buildMessageHandler(handler MessageHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		envelope := newEnvelopeFromMessage(msg)

		if err := s.handleMessage(handler, envelope); err != nil {
			s.logMessageHandlerError(envelope, err)
		}
	}
}

// buildRPCHandler adapts an RPCHandler into the callback shape required by the
// low-level NATS subscription API.
//
// Design choice:
//   - the adapter stays thin
//   - actual RPC processing and logging are delegated to named methods
func (s *server) buildRPCHandler(handler RPCHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		envelope := newEnvelopeFromMessage(msg)

		if err := s.handleRPC(handler, envelope); err != nil {
			s.logRPCHandlerError(envelope, err)
		}
	}
}

// handleMessage executes a message handler for a single inbound envelope.
func (s *server) handleMessage(
	handler MessageHandler,
	envelope Envelope,
) error {
	if err := handler(context.Background(), envelope); err != nil {
		return fmt.Errorf("%w: %w", ErrHandlerFailed, err)
	}

	return nil
}

// handleRPC executes an RPC handler for a single inbound envelope and publishes
// the encoded response to the reply subject.
//
// Design choices:
//   - request metadata remains available through Envelope
//   - the server owns response encoding and reply publishing
//   - errors are returned to the caller so logging stays centralized
func (s *server) handleRPC(
	handler RPCHandler,
	envelope Envelope,
) error {
	if envelope.ReplyTo == "" {
		return fmt.Errorf("%w: missing reply subject", ErrHandlerFailed)
	}

	response, err := handler(context.Background(), envelope)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrHandlerFailed, err)
	}

	payload, err := s.codec.Marshal(response)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEncodeFailed, err)
	}

	if err := s.connection.publish(envelope.ReplyTo, payload); err != nil {
		return err
	}

	return nil
}

// logMessageHandlerError logs a message handler failure with transport context.
func (s *server) logMessageHandlerError(
	envelope Envelope,
	err error,
) {
	s.logger.Error(
		"message handler failed",
		"subject", envelope.Subject,
		"error", err.Error(),
	)
}

// logRPCHandlerError logs an RPC handler failure with transport context.
func (s *server) logRPCHandlerError(
	envelope Envelope,
	err error,
) {
	s.logger.Error(
		"rpc handler failed",
		"subject", envelope.Subject,
		"reply_to", envelope.ReplyTo,
		"error", err.Error(),
	)
}

// newEnvelopeFromMessage converts a raw NATS message into a transport-level
// envelope.
//
// Design choice:
//   - the shared messaging layer maps transport-specific messages into a
//     transport-neutral structure before handing them to application logic
func newEnvelopeFromMessage(msg *nats.Msg) Envelope {
	return Envelope{
		Subject: msg.Subject,
		ReplyTo: msg.Reply,
		Headers: newHeadersFromMessage(msg),
		Payload: msg.Data,
	}
}

// newHeadersFromMessage converts NATS headers into a plain string map.
func newHeadersFromMessage(msg *nats.Msg) map[string]string {
	if msg == nil || len(msg.Header) == 0 {
		return nil
	}

	headers := make(map[string]string, len(msg.Header))

	for key, values := range msg.Header {
		if len(values) == 0 {
			continue
		}

		headers[key] = values[0]
	}

	return headers
}
