package messaging

// Envelope represents a transport-level message independent of the underlying
// messaging provider.
//
// It contains routing information, optional metadata, and the raw payload
// delivered over the wire. Envelope is intentionally a data-only structure and
// does not perform encoding, decoding, or transport operations.
type Envelope struct {
	// Subject is the message subject used for routing.
	Subject string

	// ReplyTo is the subject where a response should be sent.
	// It is typically set for request/reply style communication.
	ReplyTo string

	// Headers contain optional message metadata such as content type,
	// correlation identifiers, or tracing values.
	Headers map[string]string

	// Payload contains the raw message body.
	// Serialization and deserialization are handled by a Codec.
	Payload []byte
}
