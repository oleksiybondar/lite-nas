// Package journal defines browser-facing journal query and vacuum DTOs.
package journal

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	contract "lite-nas/shared/contracts/journalmanager"
)

// QueryInput validates browser query parameters before calling the service.
type QueryInput struct {
	Since    string `query:"since" format:"date-time" doc:"Optional UTC lower time bound."`
	Until    string `query:"until" format:"date-time" doc:"Optional UTC upper time bound."`
	Message  string `query:"message" maxLength:"256" doc:"Literal message text to search."`
	Unit     string `query:"unit" maxLength:"256" pattern:"^[A-Za-z0-9][A-Za-z0-9_.@:-]*\\.service$" doc:"Optional systemd service unit."`
	Priority string `query:"priority" pattern:"^[0-7]?$" doc:"Optional maximum journal priority."`
	Cursor   string `query:"cursor" maxLength:"1024" pattern:"^[A-Za-z0-9;=_.:-]*$" doc:"Opaque cursor returned by the previous chunk."`
	Limit    int    `query:"limit" minimum:"1" maximum:"100" default:"50" doc:"Maximum entries in one chunk."`
}

// QueryOutput wraps one journal chunk for the browser.
type QueryOutput struct{ Body QueryBody }

// UnitsInput validates filters used to discover matching journal service units.
type UnitsInput struct {
	Since    string `query:"since" format:"date-time" doc:"Optional UTC lower time bound."`
	Until    string `query:"until" format:"date-time" doc:"Optional UTC upper time bound."`
	Message  string `query:"message" maxLength:"256" doc:"Literal message text to search."`
	Priority string `query:"priority" pattern:"^[0-7]?$" doc:"Optional maximum journal priority."`
}

// UnitsOutput wraps the matching journal service-unit list.
type UnitsOutput struct{ Body UnitsBody }

// UnitsBody contains matching service-unit names.
type UnitsBody struct {
	dto.ResponseMeta
	Data contract.UnitsResponse `json:"data"`
}

// NewUnitsBody wraps a validated unit discovery response.
func NewUnitsBody(data contract.UnitsResponse) UnitsBody {
	return UnitsBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}
}

// QueryBody contains entries and continuation metadata.
type QueryBody struct {
	dto.ResponseMeta
	Data contract.QueryResponse `json:"data"`
}

// NewQueryBody wraps a validated service response.
func NewQueryBody(data contract.QueryResponse) QueryBody {
	return QueryBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}
}

// VacuumInput validates the retention days in the browser request.
type VacuumInput struct {
	Body struct {
		Days int `json:"days" minimum:"1" maximum:"365" required:"true"`
	}
}

// VacuumOutput acknowledges the completed vacuum and alert publication.
type VacuumOutput struct{ Body VacuumBody }

// VacuumBody contains a success response.
type VacuumBody struct {
	dto.ResponseMeta
	Completed bool `json:"completed"`
}

// NewVacuumBody wraps a completed vacuum action.
func NewVacuumBody() VacuumBody {
	return VacuumBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Completed: true}
}
