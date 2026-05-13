package loggingmanager

import (
	"context"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/model"
	"lite-nas/shared/loggingmanager/query"
)

// GetEvent returns one event by business event ID.
func (core *Core) GetEvent(input dto.GetEventHistoryInput) (model.Event, bool, error) {
	if err := core.validator.Struct(input); err != nil {
		return model.Event{}, false, err
	}

	builtQuery := query.BuildGetEventByIDQuery(input)
	items, err := core.listEventsQuery(context.Background(), builtQuery)
	if err != nil {
		return model.Event{}, false, err
	}
	if len(items) == 0 {
		return model.Event{}, false, nil
	}
	return items[0], true, nil
}

// ListEvents returns paginated events using the provided filters.
func (core *Core) ListEvents(input dto.ListEventsInput) ([]model.Event, error) {
	if err := core.validator.Struct(input); err != nil {
		return nil, err
	}
	builtQuery, err := query.BuildListEventsQuery(input)
	if err != nil {
		return nil, err
	}
	return core.listEventsQuery(context.Background(), builtQuery)
}

// ListActiveEvents returns active events only.
func (core *Core) ListActiveEvents(input dto.ListEventsInput) ([]model.Event, error) {
	if err := core.validator.Struct(input); err != nil {
		return nil, err
	}
	builtQuery, err := query.BuildListActiveEventsQuery(input)
	if err != nil {
		return nil, err
	}
	return core.listEventsQuery(context.Background(), builtQuery)
}

// ListActiveUnacknowledgedEvents returns active and unacknowledged events.
func (core *Core) ListActiveUnacknowledgedEvents(input dto.ListEventsInput) ([]model.Event, error) {
	if err := core.validator.Struct(input); err != nil {
		return nil, err
	}
	builtQuery, err := query.BuildListActiveUnacknowledgedEventsQuery(input)
	if err != nil {
		return nil, err
	}
	return core.listEventsQuery(context.Background(), builtQuery)
}

// SetState updates event state through the writer queue.
func (core *Core) SetState(input dto.SetStateInput) error {
	if err := core.validator.Struct(input); err != nil {
		return err
	}
	return core.setState(context.Background(), input)
}

// AcknowledgeEvent updates lifecycle acknowledgement through the writer queue.
func (core *Core) AcknowledgeEvent(input dto.AcknowledgeEventInput) error {
	if err := core.validator.Struct(input); err != nil {
		return err
	}
	return core.acknowledgeEvent(context.Background(), input)
}

// MuteEvent updates lifecycle mute state through the writer queue.
func (core *Core) MuteEvent(input dto.MuteEventInput) error {
	if err := core.validator.Struct(input); err != nil {
		return err
	}
	return core.muteEvent(context.Background(), input)
}
