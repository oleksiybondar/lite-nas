// Package journalmanager defines authenticated journal query and vacuum RPCs.
package journalmanager

const (
	QuerySubject  = "journal-manager.rpc.query"
	UnitsSubject  = "journal-manager.rpc.units"
	VacuumSubject = "journal-manager.rpc.vacuum"
)

// QueryRequest selects a bounded newest-first slice of the host journal.
type QueryRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Since       string `json:"since,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Until       string `json:"until,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Message     string `json:"message,omitempty" validate:"max=256"`
	Unit        string `json:"unit,omitempty" validate:"max=256"`
	Priority    *int   `json:"priority,omitempty" validate:"omitempty,gte=0,lte=7"`
	Cursor      string `json:"cursor,omitempty" validate:"max=1024"`
	Limit       int    `json:"limit" validate:"required,gte=1,lte=100"`
}

// Entry contains selected structured fields from one journal record.
type Entry struct {
	Cursor     string `json:"cursor"`
	Timestamp  string `json:"timestamp"`
	Message    string `json:"message"`
	Unit       string `json:"unit,omitempty"`
	Priority   string `json:"priority,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	PID        string `json:"pid,omitempty"`
}

// QueryResponse returns a chunk and the cursor needed to continue browsing.
type QueryResponse struct {
	Entries    []Entry `json:"entries"`
	NextCursor string  `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

// UnitsRequest selects journal units with entries matching non-unit filters.
type UnitsRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Since       string `json:"since,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Until       string `json:"until,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Message     string `json:"message,omitempty" validate:"max=256"`
	Priority    *int   `json:"priority,omitempty" validate:"omitempty,gte=0,lte=7"`
}

// UnitsResponse lists service units present in the matching journal records.
type UnitsResponse struct {
	Units []string `json:"units"`
}

// VacuumRequest selects retention in whole days.
type VacuumRequest struct {
	AccessToken string `json:"access_token" validate:"required,max=8192"`
	Days        int    `json:"days" validate:"required,gte=1,lte=365"`
}

// VacuumResponse reports whether rotation, vacuum, and alert publication completed.
type VacuumResponse struct {
	Completed bool `json:"completed"`
}
