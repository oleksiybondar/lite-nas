package query

import (
	"lite-nas/shared/loggingmanager/dto"
)

// UpsertEvent builds an upsert query for one events row.
//
// Contract:
//   - row.EventID must follow the configured generated-ID policy and stay
//     within schema limits.
//   - row.RecID is the rotation-slot identity used for overwrite semantics.
//
// Side effects:
//   - None. The function only returns SQL + args data.
func UpsertEvent(row dto.EventRow) Query {
	return Query{
		SQL: "INSERT INTO events (rec_id, event_id, category, severity, priority, created_at, source) " +
			"VALUES (?, ?, ?, ?, ?, ?, ?) " +
			"ON CONFLICT(rec_id) DO UPDATE SET " +
			"event_id = excluded.event_id, " +
			"category = excluded.category, " +
			"severity = excluded.severity, " +
			"priority = excluded.priority, " +
			"created_at = excluded.created_at, " +
			"source = excluded.source",
		Args: []any{row.RecID, row.EventID, row.Category, string(row.Severity), row.Priority, row.CreatedAt, row.Source},
	}
}
