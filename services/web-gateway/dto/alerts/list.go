package alerts

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

const (
	// DefaultPage defines the browser-facing default page number for alert lists.
	DefaultPage = 1
	// DefaultSize defines the browser-facing default page size for alert lists.
	DefaultSize = 20
	// MaxSize defines the largest supported browser-facing page size.
	MaxSize = 500
)

// ListInput defines the shared list query accepted by alert collection routes.
type ListInput struct {
	Page int `query:"page" default:"1" minimum:"1" doc:"Page number. Defaults to 1 when omitted."`
	Size int `query:"size" default:"20" minimum:"1" maximum:"500" doc:"Page size. Defaults to 20 when omitted."`
}

// ListOutput returns one browser-facing alert page with wrapped pagination metadata.
type ListOutput struct {
	Body ListBody
}

// ListBody defines the browser-facing alert-list response envelope.
type ListBody struct {
	dto.ResponseMeta
	Data ListData `json:"data"`
}

// ListData contains one page of alert items plus browser-facing paging metadata.
type ListData struct {
	Items    []loggingmanagercontract.ListAlertItem `json:"items"`
	Metadata ListMetadata                           `json:"metadata"`
}

// ListMetadata contains the browser-facing page metadata returned with alert lists.
type ListMetadata struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// NewListBody creates the browser-facing alert-list response body.
func NewListBody(now time.Time, items []loggingmanagercontract.ListAlertItem, metadata ListMetadata) ListBody {
	return ListBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: ListData{
			Items:    items,
			Metadata: metadata,
		},
	}
}
