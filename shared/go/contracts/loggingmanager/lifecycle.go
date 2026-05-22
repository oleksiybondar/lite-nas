package loggingmanager

import loggingmanagerdto "lite-nas/shared/loggingmanager/dto"

// AcknowledgeAlertInput defines acknowledge contract input including auth context.
type AcknowledgeAlertInput struct {
	AccessToken    string `json:"access_token" validate:"required,min=1,max=8192"`
	EventID        string `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	AcknowledgedBy string `json:"acknowledged_by" validate:"required,min=1,max=128,printascii"`
	AcknowledgedAt string `json:"acknowledged_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// ToDTO converts acknowledge contract input into core DTO input.
func (input AcknowledgeAlertInput) ToDTO() loggingmanagerdto.AcknowledgeEventInput {
	return loggingmanagerdto.AcknowledgeEventInput{
		EventID:        input.EventID,
		AcknowledgedBy: input.AcknowledgedBy,
		AcknowledgedAt: input.AcknowledgedAt,
	}
}

// MuteAlertInput defines mute contract input including auth context.
type MuteAlertInput struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	EventID     string `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	MutedBy     string `json:"muted_by" validate:"required,min=1,max=128,printascii"`
	MutedAt     string `json:"muted_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// ToDTO converts mute contract input into core DTO input.
func (input MuteAlertInput) ToDTO() loggingmanagerdto.MuteEventInput {
	return loggingmanagerdto.MuteEventInput{
		EventID: input.EventID,
		MutedBy: input.MutedBy,
		MutedAt: input.MutedAt,
	}
}
