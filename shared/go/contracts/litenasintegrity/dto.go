package litenasintegrity

import (
	"time"

	backgroundtasks "lite-nas/shared/contracts/backgroundtasks"
)

// InfoRequest authorizes one read of active AIDE baseline state.
type InfoRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// InfoResponse reports whether an active baseline file is available plus concise metadata.
type InfoResponse struct {
	BaselineAvailable bool       `json:"baseline_available"`
	BaselinePath      string     `json:"baseline_path"`
	BaselineUpdatedAt *time.Time `json:"baseline_updated_at,omitempty"`
	BaselineSize      int64      `json:"baseline_size,omitempty"`
	Details           string     `json:"details,omitempty"`
}

// RegenerateBaselineRequest requests a privileged AIDE baseline regeneration.
type RegenerateBaselineRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// RegenerateBaselineResponse reports whether baseline regeneration completed.
type RegenerateBaselineResponse struct {
	OK bool `json:"ok"`
}

// TaskRequest identifies a task attempt and the human access token authorizing it.
type TaskRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	TaskID      string `json:"task_id" validate:"required,len=32,hexadecimal"`
}

// StartCheckRequest authorizes one on-demand AIDE check.
type StartCheckRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// StartCheckResponse identifies the queued AIDE check.
type StartCheckResponse struct {
	TaskID string `json:"task_id" validate:"required,len=32,hexadecimal"`
	State  string `json:"state" validate:"required,oneof=queued running succeeded failed cancelled"`
}

// StartBaselineRequest authorizes one on-demand AIDE baseline generation.
type StartBaselineRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// StartBaselineResponse identifies the queued baseline generation task.
type StartBaselineResponse = StartCheckResponse

// TaskStatusResponse is the shared pollable task snapshot.
type TaskStatusResponse = backgroundtasks.StatusCheckResponse

// TaskRetryResponse identifies the replacement AIDE attempt.
type TaskRetryResponse = backgroundtasks.RetryResponse
