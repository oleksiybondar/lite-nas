package query

// DeleteOldestEventsBeyondLimit removes the oldest retained events and keeps
// only the newest limit event identities.
func DeleteOldestEventsBeyondLimit(limit int) Query {
	return Query{
		SQL: "DELETE FROM events " +
			"WHERE rec_id IN (" +
			"SELECT rec_id FROM events ORDER BY rec_id DESC LIMIT -1 OFFSET ?" +
			")",
		Args: []any{limit},
	}
}

// DeleteOrphanLifecycle removes lifecycle rows that no longer belong to any
// retained event identity.
func DeleteOrphanLifecycle() Query {
	return Query{
		SQL: "DELETE FROM lifecycle " +
			"WHERE NOT EXISTS (" +
			"SELECT 1 FROM events e WHERE e.event_id = lifecycle.event_id" +
			")",
	}
}

// DeleteOrphanEventState removes state rows that no longer belong to any
// retained event identity.
func DeleteOrphanEventState() Query {
	return Query{
		SQL: "DELETE FROM event_state " +
			"WHERE NOT EXISTS (" +
			"SELECT 1 FROM events e WHERE e.event_id = event_state.event_id" +
			")",
	}
}

// DeleteOrphanOccurrences removes occurrences that no longer belong to any
// active event identity.
func DeleteOrphanOccurrences() Query {
	return Query{
		SQL: "DELETE FROM occurrences " +
			"WHERE NOT EXISTS (" +
			"SELECT 1 FROM events e WHERE e.event_id = occurrences.event_id" +
			")",
	}
}

// DeleteOccurrencesPerEventBeyondLimit removes older occurrence records and
// keeps only the latest limit rows for each event_id independently.
func DeleteOccurrencesPerEventBeyondLimit(limit int) Query {
	return Query{
		SQL: "DELETE FROM occurrences " +
			"WHERE rec_id IN (" +
			"SELECT rec_id FROM (" +
			"SELECT rec_id, ROW_NUMBER() OVER (PARTITION BY event_id ORDER BY rec_id DESC) AS row_num " +
			"FROM occurrences" +
			") ranked WHERE row_num > ?" +
			")",
		Args: []any{limit},
	}
}

// DeleteOrphanEventMeta removes event metadata rows that no longer belong to
// any active event identity.
func DeleteOrphanEventMeta() Query {
	return Query{
		SQL: "DELETE FROM event_meta " +
			"WHERE NOT EXISTS (" +
			"SELECT 1 FROM events e WHERE e.event_id = event_meta.event_id" +
			")",
	}
}
