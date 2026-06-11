package dto

import "lite-nas/shared/loggingmanager/enum"

// EventStateRow models one row in the event_state table.
type EventStateRow struct {
	RecID      int64
	EventID    string
	EventRecID int64
	Status     enum.Status
	Message    string
}
