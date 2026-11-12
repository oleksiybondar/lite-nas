package query

import "lite-nas/shared/loggingmanager/dto"

// UpsertEventState builds an upsert keyed by event ID while retaining the
// original row identity for duplicate event deliveries.
func UpsertEventState(row dto.EventStateRow) Query {
	return Query{
		SQL: "INSERT INTO event_state (rec_id, event_id, event_rec_id, status, message) VALUES (?, ?, ?, ?, ?) " +
			"ON CONFLICT(event_id) DO UPDATE SET " +
			"status = excluded.status, " +
			"message = excluded.message",
		Args: []any{row.RecID, row.EventID, row.EventRecID, string(row.Status), row.Message},
	}
}
