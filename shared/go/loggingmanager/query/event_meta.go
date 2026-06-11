package query

import "lite-nas/shared/loggingmanager/dto"

// UpsertEventMeta builds an upsert query for one event_meta row.
func UpsertEventMeta(row dto.EventMetaRow) Query {
	return Query{
		SQL:  "INSERT INTO event_meta (event_id, key, value) VALUES (?, ?, ?) ON CONFLICT(event_id, key) DO UPDATE SET value = excluded.value",
		Args: []any{row.EventID, row.MetaKey, row.MetaValue},
	}
}
