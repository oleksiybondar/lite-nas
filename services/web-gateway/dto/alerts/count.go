package alerts

import (
	"time"

	"lite-nas/services/web-gateway/dto"
)

// CountInput defines the shared list query accepted by alert count routes.
type CountInput struct {
	Page int `query:"page" default:"1" minimum:"1" doc:"Ignored by the count endpoint. Present for transport consistency."`
	Size int `query:"size" default:"20" minimum:"1" maximum:"500" doc:"Ignored by the count endpoint. Present for transport consistency."`
}

// CountOutput returns the simplified browser-facing alert count response.
type CountOutput struct {
	Body CountBody
}

// CountBody defines the browser-facing alert-count response envelope.
type CountBody struct {
	dto.ResponseMeta
	Data CountData `json:"data"`
}

// CountData contains the simplified count-only payload for alert dashboards.
type CountData struct {
	Count int `json:"count"`
}

// NewCountBody creates the browser-facing alert-count response body.
func NewCountBody(now time.Time, count int) CountBody {
	return CountBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: CountData{Count: count},
	}
}
