package loggingmanager

import (
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
)

// UpdateAlertStateInput defines state-update contract input including auth context.
type UpdateAlertStateInput struct {
	AccessToken string      `json:"access_token" validate:"required,min=1,max=8192"`
	EventID     string      `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	Status      enum.Status `json:"status" validate:"required,oneof=high low normal active failure"`
	Message     *string     `json:"message,omitempty" validate:"omitempty,max=256"`
}

// ToDTO converts state-update contract input into core DTO input.
func (input UpdateAlertStateInput) ToDTO() loggingmanagerdto.SetStateInput {
	return loggingmanagerdto.SetStateInput{
		EventID: input.EventID,
		Status:  input.Status,
		Message: input.Message,
	}
}
