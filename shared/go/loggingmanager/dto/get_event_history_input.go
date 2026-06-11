package dto

// GetEventHistoryInput defines history retrieval request input for one event.
type GetEventHistoryInput struct {
	EventID string `json:"event_id" validate:"required,max=20,loggingmanager_event_id"`
}
