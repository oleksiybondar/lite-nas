package alerts

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

// GetInput defines the path parameters accepted by one alert-detail route.
type GetInput struct {
	ID string `path:"id" minLength:"1" maxLength:"20" doc:"Alert business record ID."`
}

// GetOutput returns one browser-facing alert-detail response.
type GetOutput struct {
	Body GetBody
}

// GetBody defines the browser-facing alert-detail response envelope.
type GetBody struct {
	dto.ResponseMeta
	Data loggingmanagercontract.ListAlertItem `json:"data"`
}

// NewGetBody creates the browser-facing alert-detail response body.
func NewGetBody(now time.Time, item loggingmanagercontract.ListAlertItem) GetBody {
	return GetBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: item,
	}
}
