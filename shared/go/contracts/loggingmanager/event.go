package loggingmanager

import (
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
)

// AlertPayload defines event-create contract input including auth context.
type AlertPayload struct {
	AccessToken  string        `json:"access_token" validate:"required,min=1,max=8192"`
	EventID      string        `json:"event_id,omitempty" validate:"omitempty,max=20,loggingmanager_event_id"`
	Category     string        `json:"category" validate:"required,min=1,max=128,printascii"`
	Severity     enum.Severity `json:"severity,omitempty" validate:"omitempty,oneof=info warning error critical"`
	Priority     *int          `json:"priority,omitempty" validate:"omitempty,gte=0,lte=5"`
	CreatedAt    string        `json:"created_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Source       string        `json:"source,omitempty" validate:"omitempty,min=1,max=128,printascii"`
	Message      string        `json:"message,omitempty" validate:"omitempty,max=256"`
	TriggerValue string        `json:"trigger_value,omitempty" validate:"omitempty,max=512"`
}

// ToDTO converts contract payload into core DTO input.
func (payload AlertPayload) ToDTO() loggingmanagerdto.CreateEventInput {
	return loggingmanagerdto.CreateEventInput{
		EventID:   payload.EventID,
		Category:  payload.Category,
		Severity:  payload.Severity,
		Priority:  payload.Priority,
		CreatedAt: payload.CreatedAt,
		Source:    payload.Source,
	}
}

// AlertOccurrencePayload defines occurrence contract input including auth context.
type AlertOccurrencePayload struct {
	AccessToken string         `json:"access_token" validate:"required,min=1,max=8192"`
	RecID       int64          `json:"rec_id,omitempty"`
	EventID     string         `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	EventRecID  int64          `json:"event_rec_id,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	ValueType   enum.ValueType `json:"value_type,omitempty"`
	ValueNum    *float64       `json:"value_num,omitempty"`
	ValueText   *string        `json:"value_text,omitempty"`
	ValueBool   *bool          `json:"value_bool,omitempty"`
	ValueUnit   *string        `json:"value_unit,omitempty"`
}

// ToDTO converts contract occurrence payload into core DTO input.
func (payload AlertOccurrencePayload) ToDTO() loggingmanagerdto.OccurrenceRow {
	return loggingmanagerdto.OccurrenceRow{
		RecID:      payload.RecID,
		EventID:    payload.EventID,
		EventRecID: payload.EventRecID,
		Timestamp:  payload.Timestamp,
		ValueType:  payload.ValueType,
		ValueNum:   payload.ValueNum,
		ValueText:  payload.ValueText,
		ValueBool:  payload.ValueBool,
		ValueUnit:  payload.ValueUnit,
	}
}
