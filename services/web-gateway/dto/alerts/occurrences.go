package alerts

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

// OccurrencesInput defines the path parameters accepted by one alert-occurrence route.
type OccurrencesInput struct {
	ID string `path:"id" minLength:"1" maxLength:"20" doc:"Alert business record ID."`
}

// OccurrencesOutput returns one browser-facing alert-occurrences response.
type OccurrencesOutput struct {
	Body OccurrencesBody
}

// OccurrencesBody defines the browser-facing alert-occurrences response envelope.
type OccurrencesBody struct {
	dto.ResponseMeta
	Data []loggingmanagercontract.AlertOccurrenceItem `json:"data"`
}

// NewOccurrencesBody creates the browser-facing alert-occurrences response body.
func NewOccurrencesBody(now time.Time, items []loggingmanagercontract.AlertOccurrenceItem) OccurrencesBody {
	return OccurrencesBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: items,
	}
}
