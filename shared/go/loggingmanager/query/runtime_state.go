package query

import (
	"strconv"

	"lite-nas/shared/loggingmanager/dto"
)

const (
	// RuntimeStateCurrentEventRecIDKey stores the rotation-slot pointer.
	RuntimeStateCurrentEventRecIDKey = "current_event_rec_id"
	// RuntimeStateCurrentEventSeqKey stores the generated event-id sequence.
	RuntimeStateCurrentEventSeqKey = "current_event_seq"
	// RuntimeStateEventIDPrefixKey stores the current event-id prefix.
	RuntimeStateEventIDPrefixKey = "event_id_prefix"
)

// UpsertRuntimeState builds an upsert query for one runtime_state row.
func UpsertRuntimeState(row dto.RuntimeStateRow) Query {
	return Query{
		SQL:  "INSERT INTO runtime_state (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
		Args: []any{row.Key, row.Value},
	}
}

// BuildRuntimeStateSeedQueries returns idempotent seed queries for required
// runtime-state keys used by rotation and event-id generation.
//
// Seed behavior:
//   - current_event_rec_id tracks the current rotation slot pointer.
//   - current_event_seq tracks the current generated event-id sequence.
//   - event_id_prefix tracks the default event-id prefix.
func BuildRuntimeStateSeedQueries(defaultCurrentEventRecID int64, defaultCurrentEventSeq uint32, defaultPrefix string) []Query {
	return []Query{
		{
			SQL:  "INSERT OR IGNORE INTO runtime_state (key, value) VALUES (?, ?)",
			Args: []any{RuntimeStateCurrentEventRecIDKey, strconv.FormatInt(defaultCurrentEventRecID, 10)},
		},
		{
			SQL:  "INSERT OR IGNORE INTO runtime_state (key, value) VALUES (?, ?)",
			Args: []any{RuntimeStateCurrentEventSeqKey, strconv.FormatUint(uint64(defaultCurrentEventSeq), 10)},
		},
		{
			SQL:  "INSERT OR IGNORE INTO runtime_state (key, value) VALUES (?, ?)",
			Args: []any{RuntimeStateEventIDPrefixKey, defaultPrefix},
		},
	}
}
