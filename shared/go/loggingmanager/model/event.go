package model

import "lite-nas/shared/loggingmanager/dto"

// Event aggregates current event-facing data for reads.
type Event struct {
	Event     dto.EventRow
	Lifecycle dto.LifecycleRow
	State     dto.EventStateRow
	LastValue *dto.OccurrenceRow
	Meta      []dto.EventMetaRow
}
