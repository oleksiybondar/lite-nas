package backgroundtasks

import "encoding/json"

// StartRequest selects a service-owned task implementation. Parameters are
// validated against that implementation's own schema at the service boundary.
type StartRequest struct {
	TaskType   string          `json:"task_type" validate:"required,max=128"`
	Parameters json.RawMessage `json:"parameters" validate:"required"`
}

// StartResponse identifies the newly queued task attempt.
type StartResponse struct {
	TaskID string `json:"task_id" validate:"required,len=32,hexadecimal"`
	State  string `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
}

// RetryRequest identifies a completed attempt whose saved task request will be
// submitted again. Retrying creates a new task ID and preserves the old record.
type RetryRequest struct {
	TaskID string `json:"task_id" validate:"required,len=32,hexadecimal"`
}

// RetryResponse links the previous attempt to the newly queued attempt.
type RetryResponse struct {
	PreviousTaskID string `json:"previous_task_id" validate:"required,len=32,hexadecimal"`
	TaskID         string `json:"task_id" validate:"required,len=32,hexadecimal"`
	State          string `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
}
