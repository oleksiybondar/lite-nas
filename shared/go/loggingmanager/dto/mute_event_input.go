package dto

// MuteEventInput defines lifecycle mute transition input.
type MuteEventInput struct {
	EventID string `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
	MutedBy string `json:"muted_by" validate:"required,min=1,max=128,printascii"`
	MutedAt string `json:"muted_at,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}
