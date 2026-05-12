package query

import "lite-nas/shared/loggingmanager/dto"

// UpsertLifecycle builds an upsert query for one lifecycle row.
func UpsertLifecycle(row dto.LifecycleRow) Query {
	return Query{
		SQL: "INSERT INTO lifecycle (rec_id, event_id, event_rec_id, acknowledged, acknowledged_by, acknowledged_at, muted, muted_by, muted_at) " +
			"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) " +
			"ON CONFLICT(rec_id) DO UPDATE SET " +
			"event_id = excluded.event_id, " +
			"event_rec_id = excluded.event_rec_id, " +
			"acknowledged = excluded.acknowledged, " +
			"acknowledged_by = excluded.acknowledged_by, " +
			"acknowledged_at = excluded.acknowledged_at, " +
			"muted = excluded.muted, " +
			"muted_by = excluded.muted_by, " +
			"muted_at = excluded.muted_at",
		Args: []any{
			row.RecID, row.EventID, row.EventRecID, boolToInt(row.Acknowledged), row.AcknowledgedBy, row.AcknowledgedAt,
			boolToInt(row.Muted), row.MutedBy, row.MutedAt,
		},
	}
}

// SelectLifecycleByEventID builds a read query that resolves lifecycle row by
// event_id.
func SelectLifecycleByEventID(eventID string) Query {
	return Query{
		SQL: "SELECT rec_id, event_id, event_rec_id, acknowledged, acknowledged_by, acknowledged_at, muted, muted_by, muted_at " +
			"FROM lifecycle WHERE event_id = ?",
		Args: []any{eventID},
	}
}
