package loggingmanager

import (
	"context"
	"database/sql"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/model"
	"lite-nas/shared/loggingmanager/query"
)

// ListEventsPage contains one page of alert rows plus the total number of
// matching rows before pagination.
type ListEventsPage struct {
	Items      []model.Event
	TotalCount int
}

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
	page, err := core.ListEventsPage(input)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// ListEventsPage returns paginated events plus the total count for the
// provided filters.
func (core *Core) ListEventsPage(input dto.ListEventsInput) (ListEventsPage, error) {
	if err := core.validator.Struct(input); err != nil {
		return ListEventsPage{}, err
	}
	builtQuery, err := query.BuildListEventsQuery(input)
	if err != nil {
		return ListEventsPage{}, err
	}
	countQuery, err := query.BuildCountEventsQuery(input)
	if err != nil {
		return ListEventsPage{}, err
	}
	return core.listEventsPage(context.Background(), builtQuery, countQuery)
}

// ListActiveEvents returns active events only.
func (core *Core) ListActiveEvents(input dto.ListEventsInput) ([]model.Event, error) {
	page, err := core.ListActiveEventsPage(input)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// ListActiveEventsPage returns active events only plus the total count.
func (core *Core) ListActiveEventsPage(input dto.ListEventsInput) (ListEventsPage, error) {
	if err := core.validator.Struct(input); err != nil {
		return ListEventsPage{}, err
	}
	builtQuery, err := query.BuildListActiveEventsQuery(input)
	if err != nil {
		return ListEventsPage{}, err
	}
	countQuery, err := query.BuildCountActiveEventsQuery(input)
	if err != nil {
		return ListEventsPage{}, err
	}
	return core.listEventsPage(context.Background(), builtQuery, countQuery)
}

// ListActiveUnacknowledgedEvents returns active and unacknowledged events.
func (core *Core) ListActiveUnacknowledgedEvents(input dto.ListEventsInput) ([]model.Event, error) {
	page, err := core.ListActiveUnacknowledgedEventsPage(input)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

// ListActiveUnacknowledgedEventsPage returns active and unacknowledged events
// plus the total count.
func (core *Core) ListActiveUnacknowledgedEventsPage(input dto.ListEventsInput) (ListEventsPage, error) {
	if err := core.validator.Struct(input); err != nil {
		return ListEventsPage{}, err
	}
	builtQuery, err := query.BuildListActiveUnacknowledgedEventsQuery(input)
	if err != nil {
		return ListEventsPage{}, err
	}
	countQuery, err := query.BuildCountActiveUnacknowledgedEventsQuery(input)
	if err != nil {
		return ListEventsPage{}, err
	}
	return core.listEventsPage(context.Background(), builtQuery, countQuery)
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

func (core *Core) listEventsPage(
	ctx context.Context,
	listQuery query.Query,
	countQuery query.Query,
) (ListEventsPage, error) {
	items, err := core.listEventsQuery(ctx, listQuery)
	if err != nil {
		return ListEventsPage{}, err
	}

	totalCount, err := core.countEventsQuery(ctx, countQuery)
	if err != nil {
		return ListEventsPage{}, err
	}

	return ListEventsPage{
		Items:      items,
		TotalCount: totalCount,
	}, nil
}

func (core *Core) countEventsQuery(ctx context.Context, builtQuery query.Query) (int, error) {
	var count int
	err := core.db.QueryRowContext(ctx, builtQuery.SQL, builtQuery.Args...).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}
