package messaging

import (
	"context"
	"fmt"
	"time"

	"lite-nas/shared/config"
	"lite-nas/shared/logger"
)

// Client defines outbound messaging operations built on top of the low-level
// transport connection.
//
// Design choices:
//   - the client is responsible only for outbound flows
//   - payload serialization is delegated to the injected Codec
//   - transport concerns are delegated to the low-level connection wrapper
type Client interface {
	// Publish serializes the given payload and publishes it to the provided
	// subject.
	Publish(ctx context.Context, subject string, payload any) error

	// Request serializes the request payload, sends it to the provided subject,
	// waits for a reply, and deserializes the reply into response.
	Request(ctx context.Context, subject string, request any, response any) error

	// Drain gracefully drains the underlying transport connection.
	Drain() error

	// Close immediately closes the underlying transport connection.
	Close()
}

// client is the default outbound messaging implementation.
//
// Design choices:
//   - it composes a tested low-level connection instead of exposing NATS types
//   - it uses the injected Codec directly instead of wrapping codec calls in
//     trivial helper methods
//   - it keeps timeout policy local because request timeout is a client concern
type client struct {
	connection *connection
	codec      Codec
	timeout    time.Duration
	logger     logger.Logger
}

// NewClient creates a new outbound messaging client.
//
// It establishes the low-level transport connection, stores the configured
// request timeout, and uses the provided Codec for payload serialization.
func NewClient(
	cfg config.MessagingConfig,
	log logger.Logger,
	codec Codec,
) (Client, error) {
	if codec == nil {
		return nil, ErrInvalidConfig
	}

	if cfg.Timeout <= 0 {
		return nil, ErrInvalidConfig
	}

	conn, err := newConnection(cfg, log)
	if err != nil {
		return nil, err
	}

	return &client{
		connection: conn,
		codec:      codec,
		timeout:    cfg.Timeout,
		logger:     log,
	}, nil
}

// Publish serializes the payload and publishes it to the provided subject.
func (c *client) Publish(
	ctx context.Context,
	subject string,
	payload any,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := c.codec.Marshal(payload)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEncodeFailed, err)
	}

	return c.connection.publish(subject, data)
}

// Request serializes the request payload, sends it to the provided subject,
// waits for a reply, and deserializes the reply into response.
func (c *client) Request(
	ctx context.Context,
	subject string,
	request any,
	response any,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	timeout, err := c.resolveTimeout(ctx)
	if err != nil {
		return err
	}

	requestPayload, err := c.codec.Marshal(request)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEncodeFailed, err)
	}

	reply, err := c.connection.request(subject, requestPayload, timeout)
	if err != nil {
		return err
	}

	if err := c.codec.Unmarshal(reply.Data, response); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodeFailed, err)
	}

	return nil
}

// Drain gracefully drains the underlying transport connection.
func (c *client) Drain() error {
	return c.connection.drain()
}

// Close immediately closes the underlying transport connection.
func (c *client) Close() {
	c.connection.close()
}

// resolveTimeout determines the effective request timeout for the current call.
//
// Design choice:
//   - if the context has a deadline, the remaining time is used
//   - otherwise the configured client timeout is used
func (c *client) resolveTimeout(ctx context.Context) (time.Duration, error) {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		return c.timeout, nil
	}

	timeout := time.Until(deadline)
	if timeout <= 0 {
		return 0, context.DeadlineExceeded
	}

	return timeout, nil
}
