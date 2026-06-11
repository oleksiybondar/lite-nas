package dto

import "lite-nas/shared/loggingmanager/enum"

// SetStateInput defines a state transition request for one event.
type SetStateInput struct {
	EventID string      `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	Status  enum.Status `json:"status" validate:"required,oneof=high low normal active failure"`
	Message *string     `json:"message,omitempty" validate:"omitempty,max=256"`
}
