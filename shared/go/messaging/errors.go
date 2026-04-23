package messaging

import "errors"

var (
	// ErrNotConnected indicates that an operation was attempted on a client
	// or server that does not have an active connection to the messaging system.
	ErrNotConnected = errors.New("messaging: not connected")

	// ErrInvalidConfig indicates that the provided messaging configuration
	// is invalid or incomplete.
	ErrInvalidConfig = errors.New("messaging: invalid config")

	// ErrInvalidSubject indicates that a provided subject is empty or does
	// not meet expected formatting rules.
	ErrInvalidSubject = errors.New("messaging: invalid subject")

	// ErrEncodeFailed indicates a failure while serializing a payload using
	// the configured codec.
	ErrEncodeFailed = errors.New("messaging: encode failed")

	// ErrDecodeFailed indicates a failure while deserializing a payload using
	// the configured codec.
	ErrDecodeFailed = errors.New("messaging: decode failed")

	// ErrRequestTimeout indicates that a request/reply operation exceeded the
	// allowed time without receiving a response.
	ErrRequestTimeout = errors.New("messaging: request timeout")

	// ErrPublishFailed indicates that a message could not be published to the
	// messaging system.
	ErrPublishFailed = errors.New("messaging: publish failed")

	// ErrSubscribeFailed indicates that a subscription could not be created
	// or registered.
	ErrSubscribeFailed = errors.New("messaging: subscribe failed")

	// ErrHandlerFailed indicates that a message or RPC handler returned an
	// error during processing.
	ErrHandlerFailed = errors.New("messaging: handler failed")
)
