package dto

import "time"

// ResponseMeta defines the common browser-facing response envelope metadata.
//
// Only Success and Timestamp are required in the initial gateway skeleton.
// The remaining fields are part of the target envelope shape and may be
// populated incrementally as tracing and error-code wiring is added.
type ResponseMeta struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
	Code      string    `json:"code,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
}
