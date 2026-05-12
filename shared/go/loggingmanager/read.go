package loggingmanager

import (
	"context"
	"database/sql"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanager/model"
	"lite-nas/shared/loggingmanager/query"
)

func (core *Core) listEventsQuery(ctx context.Context, builtQuery query.Query) ([]model.Event, error) {
	rows, err := core.db.QueryContext(ctx, builtQuery.SQL, builtQuery.Args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]model.Event, 0)
	for rows.Next() {
		item, scanErr := scanEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func scanEvent(rows *sql.Rows) (model.Event, error) {
	var scanned eventScanResult
	if err := scanEventRow(rows, &scanned); err != nil {
		return model.Event{}, err
	}
	return buildEvent(scanned), nil
}

type eventScanResult struct {
	event model.Event

	acknowledgedInt int
	mutedInt        int
	statusRaw       string
	severityRaw     string

	lastValueTs   sql.NullString
	lastValueType sql.NullString
	lastValueNum  sql.NullFloat64
	lastValueText sql.NullString
	lastValueBool sql.NullInt64
	lastValueUnit sql.NullString
	stateMessage  sql.NullString
	createdAt     string
	ackAt         string
	mutedAt       string
}

func scanEventRow(rows *sql.Rows, scanned *eventScanResult) error {
	return rows.Scan(
		&scanned.event.Event.RecID,
		&scanned.event.Event.EventID,
		&scanned.event.Event.Category,
		&scanned.severityRaw,
		&scanned.event.Event.Priority,
		&scanned.createdAt,
		&scanned.event.Event.Source,
		&scanned.acknowledgedInt,
		&scanned.event.Lifecycle.AcknowledgedBy,
		&scanned.ackAt,
		&scanned.mutedInt,
		&scanned.event.Lifecycle.MutedBy,
		&scanned.mutedAt,
		&scanned.statusRaw,
		&scanned.stateMessage,
		&scanned.lastValueTs,
		&scanned.lastValueType,
		&scanned.lastValueNum,
		&scanned.lastValueText,
		&scanned.lastValueBool,
		&scanned.lastValueUnit,
	)
}

func buildEvent(scanned eventScanResult) model.Event {
	row := scanned.event
	applyCoreEventFields(&row, scanned)
	applyStateMessage(&row, scanned.stateMessage)
	row.LastValue = scanLastOccurrence(
		row.Event.EventID,
		row.Event.RecID,
		scanned.lastValueTs,
		scanned.lastValueType,
		scanned.lastValueNum,
		scanned.lastValueText,
		scanned.lastValueBool,
		scanned.lastValueUnit,
	)
	return row
}

func applyCoreEventFields(row *model.Event, scanned eventScanResult) {
	row.Event.Severity = enum.Severity(scanned.severityRaw)
	row.Event.CreatedAt = scanned.createdAt
	row.Lifecycle.RecID = row.Event.RecID
	row.Lifecycle.EventID = row.Event.EventID
	row.Lifecycle.EventRecID = row.Event.RecID
	row.Lifecycle.Acknowledged = scanned.acknowledgedInt == 1
	row.Lifecycle.AcknowledgedAt = scanned.ackAt
	row.Lifecycle.Muted = scanned.mutedInt == 1
	row.Lifecycle.MutedAt = scanned.mutedAt
	row.State.RecID = row.Event.RecID
	row.State.EventID = row.Event.EventID
	row.State.EventRecID = row.Event.RecID
	row.State.Status = enum.Status(scanned.statusRaw)
}

func applyStateMessage(row *model.Event, stateMessage sql.NullString) {
	if stateMessage.Valid {
		row.State.Message = stateMessage.String
	}
}

func scanLastOccurrence(
	eventID string,
	eventRecID int64,
	lastValueTs sql.NullString,
	lastValueType sql.NullString,
	lastValueNum sql.NullFloat64,
	lastValueText sql.NullString,
	lastValueBool sql.NullInt64,
	lastValueUnit sql.NullString,
) *dto.OccurrenceRow {
	if !lastValueType.Valid {
		return nil
	}

	occurrence := dto.OccurrenceRow{
		EventID:    eventID,
		EventRecID: eventRecID,
		ValueType:  enum.ValueType(lastValueType.String),
	}
	setOccurrenceTimestamp(&occurrence, lastValueTs)
	setOccurrenceNumber(&occurrence, lastValueNum)
	setOccurrenceText(&occurrence, lastValueText)
	setOccurrenceBool(&occurrence, lastValueBool)
	setOccurrenceUnit(&occurrence, lastValueUnit)
	return &occurrence
}

func setOccurrenceTimestamp(occurrence *dto.OccurrenceRow, value sql.NullString) {
	if value.Valid {
		occurrence.Timestamp = value.String
	}
}

func setOccurrenceNumber(occurrence *dto.OccurrenceRow, value sql.NullFloat64) {
	if value.Valid {
		raw := value.Float64
		occurrence.ValueNum = &raw
	}
}

func setOccurrenceText(occurrence *dto.OccurrenceRow, value sql.NullString) {
	if value.Valid {
		raw := value.String
		occurrence.ValueText = &raw
	}
}

func setOccurrenceBool(occurrence *dto.OccurrenceRow, value sql.NullInt64) {
	if value.Valid {
		raw := value.Int64 == 1
		occurrence.ValueBool = &raw
	}
}

func setOccurrenceUnit(occurrence *dto.OccurrenceRow, value sql.NullString) {
	if value.Valid {
		raw := value.String
		occurrence.ValueUnit = &raw
	}
}
