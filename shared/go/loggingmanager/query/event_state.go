package query

import "lite-nas/shared/loggingmanager/dto"

// UpsertEventState builds an upsert query for one event_state row.
func UpsertEventState(row dto.EventStateRow) Query {
	return Query{
		SQL: "INSERT INTO event_state (rec_id, event_id, event_rec_id, status, message) VALUES (?, ?, ?, ?, ?) " +
			"ON CONFLICT(rec_id) DO UPDATE SET " +
			"event_id = excluded.event_id, " +
			"event_rec_id = excluded.event_rec_id, " +
			"status = excluded.status, " +
			"message = excluded.message",
		Args: []any{row.RecID, row.EventID, row.EventRecID, string(row.Status), row.Message},
	}
}
