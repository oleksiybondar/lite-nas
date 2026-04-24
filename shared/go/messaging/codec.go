package messaging

import "encoding/json"

// ContentTypeJSON is the MIME type used by JSONCodec.
const ContentTypeJSON = "application/json"

// Codec defines the contract for message serialization and deserialization
// used by the messaging transport layer.
//
// Implementations are responsible for converting between Go values and raw
// byte payloads sent over the wire (e.g. via NATS).
type Codec interface {
	// Marshal serializes a Go value into a byte slice suitable for transport.
	Marshal(value any) ([]byte, error)

	// Unmarshal deserializes a byte slice into the provided target.
	// The target must be a pointer to the expected type.
	Unmarshal(data []byte, target any) error

	// ContentType returns the MIME type of the encoded payload.
	ContentType() string
}

// JSONCodec is a Codec implementation based on encoding/json.
//
// It provides stateless JSON serialization and deserialization and is intended
// to be used as the default codec for messaging payloads.
type JSONCodec struct{}

// NewJSONCodec creates a new JSONCodec instance.
//
// The codec is stateless and safe for concurrent use.
func NewJSONCodec() JSONCodec {
	return JSONCodec{}
}

// Marshal serializes the given value into JSON.
func (JSONCodec) Marshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

// Unmarshal deserializes JSON data into the provided target.
//
// The target must be a pointer to the destination type.
func (JSONCodec) Unmarshal(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

// ContentType returns the MIME type for JSON payloads.
func (JSONCodec) ContentType() string {
	return ContentTypeJSON
}
