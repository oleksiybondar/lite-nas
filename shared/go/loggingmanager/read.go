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
	var (
		row model.Event

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
	)

	err := rows.Scan(
		&row.Event.RecID,
		&row.Event.EventID,
		&row.Event.Category,
		&severityRaw,
		&row.Event.Priority,
		&createdAt,
		&row.Event.Source,
		&acknowledgedInt,
		&row.Lifecycle.AcknowledgedBy,
		&ackAt,
		&mutedInt,
		&row.Lifecycle.MutedBy,
		&mutedAt,
		&statusRaw,
		&stateMessage,
		&lastValueTs,
		&lastValueType,
		&lastValueNum,
		&lastValueText,
		&lastValueBool,
		&lastValueUnit,
	)
	if err != nil {
		return model.Event{}, err
	}

	row.Event.Severity = enum.Severity(severityRaw)
	row.Event.CreatedAt = createdAt
	row.Lifecycle.RecID = row.Event.RecID
	row.Lifecycle.EventID = row.Event.EventID
	row.Lifecycle.EventRecID = row.Event.RecID
	row.Lifecycle.Acknowledged = acknowledgedInt == 1
	row.Lifecycle.AcknowledgedAt = ackAt
	row.Lifecycle.Muted = mutedInt == 1
	row.Lifecycle.MutedAt = mutedAt
	row.State.RecID = row.Event.RecID
	row.State.EventID = row.Event.EventID
	row.State.EventRecID = row.Event.RecID
	row.State.Status = enum.Status(statusRaw)
	if stateMessage.Valid {
		row.State.Message = stateMessage.String
	}
	row.LastValue = scanLastOccurrence(row.Event.EventID, row.Event.RecID, lastValueTs, lastValueType, lastValueNum, lastValueText, lastValueBool, lastValueUnit)

	return row, nil
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
	if lastValueTs.Valid {
		occurrence.Timestamp = lastValueTs.String
	}
	if lastValueNum.Valid {
		value := lastValueNum.Float64
		occurrence.ValueNum = &value
	}
	if lastValueText.Valid {
		value := lastValueText.String
		occurrence.ValueText = &value
	}
	if lastValueBool.Valid {
		value := lastValueBool.Int64 == 1
		occurrence.ValueBool = &value
	}
	if lastValueUnit.Valid {
		value := lastValueUnit.String
		occurrence.ValueUnit = &value
	}
	return &occurrence
}
