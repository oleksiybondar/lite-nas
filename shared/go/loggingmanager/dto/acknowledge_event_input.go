package dto

// AcknowledgeEventInput defines lifecycle acknowledgement transition input.
type AcknowledgeEventInput struct {
	EventID        string `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	AcknowledgedBy string `json:"acknowledged_by" validate:"required,min=1,max=128,printascii"`
	AcknowledgedAt string `json:"acknowledged_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}
