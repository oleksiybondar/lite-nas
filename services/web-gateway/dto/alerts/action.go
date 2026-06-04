package alerts

import (
	"time"

	"lite-nas/services/web-gateway/dto"
)

// ActionInput defines the path parameters accepted by one alert action route.
type ActionInput struct {
	ID   string        `path:"id" minLength:"1" maxLength:"20" doc:"Alert business record ID."`
	Body ActionRequest `json:"-"`
}

// ActionRequest is intentionally empty because acknowledge and mute are item-scoped commands.
type ActionRequest struct{}

// ActionOutput returns the browser-facing acknowledgement of one alert action.
type ActionOutput struct {
	Body ActionBody
}

// ActionBody defines the browser-facing alert-action response envelope.
type ActionBody struct {
	dto.ResponseMeta
}

// NewActionBody creates the browser-facing alert-action response body.
func NewActionBody(now time.Time, message string) ActionBody {
	return ActionBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
			Message:   message,
		},
	}
}
