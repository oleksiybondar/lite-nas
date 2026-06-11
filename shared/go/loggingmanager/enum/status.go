package enum

// Status defines supported event-state status values.
type Status string

const (
	StatusHigh    Status = "high"
	StatusLow     Status = "low"
	StatusNormal  Status = "normal"
	StatusActive  Status = "active"
	StatusFailure Status = "failure"
)
