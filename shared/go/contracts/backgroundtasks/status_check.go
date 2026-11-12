package backgroundtasks

import "time"

// StatusCheckRequest identifies one task attempt to poll.
type StatusCheckRequest struct {
	TaskID string `json:"task_id" validate:"required,len=32,hexadecimal"`
}

// StatusCheckProgress carries optional command-specific progress information.
// Percent is omitted when the task implementation cannot report a reliable value.
type StatusCheckProgress struct {
	Percent *int   `json:"percent,omitempty" validate:"omitempty,min=0,max=100"`
	Message string `json:"message,omitempty" validate:"omitempty,max=512"`
}

// StatusCheckResponse is a pollable snapshot of one in-process task attempt.
// Output is a bounded tail and may be truncated; task state is not durable
// across an owning service process restart.
type StatusCheckResponse struct {
	TaskID          string               `json:"task_id" validate:"required,len=32,hexadecimal"`
	State           string               `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
	StartedAt       time.Time            `json:"started_at" validate:"required"`
	UpdatedAt       time.Time            `json:"updated_at" validate:"required"`
	FinishedAt      *time.Time           `json:"finished_at,omitempty"`
	ExitCode        *int                 `json:"exit_code,omitempty"`
	Error           string               `json:"error,omitempty" validate:"omitempty,max=2048"`
	Output          string               `json:"output" validate:"max=1048576"`
	OutputTruncated bool                 `json:"output_truncated"`
	Progress        *StatusCheckProgress `json:"progress,omitempty"`
}
