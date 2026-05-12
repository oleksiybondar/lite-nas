package dto

import "lite-nas/shared/loggingmanager/enum"

// CreateEventInput defines input required to create a new event row and its
// default lifecycle/state companions.
type CreateEventInput struct {
	EventID   string        `json:"event_id,omitempty" validate:"omitempty,max=20,loggingmanager_event_id"`
	Category  string        `json:"category" validate:"required,min=1,max=128,printascii"`
	Severity  enum.Severity `json:"severity,omitempty" validate:"omitempty,oneof=info warning error critical"`
	Priority  *int          `json:"priority,omitempty" validate:"omitempty,gte=0,lte=5"`
	CreatedAt string        `json:"created_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Source    string        `json:"source,omitempty" validate:"omitempty,min=1,max=128,printascii"`
}
