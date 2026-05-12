package query

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
