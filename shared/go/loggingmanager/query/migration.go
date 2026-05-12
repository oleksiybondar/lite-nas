package query

// BuildSchemaMigrationQueries returns idempotent schema creation queries.
//
// Schema contract highlights:
//   - event_id is TEXT with explicit max-length checks for bounded identifiers.
//   - occurrences persist event ownership by both event_id and event_rec_id.
//   - runtime_state is a key-value table used for generated-ID runtime pointers.
func BuildSchemaMigrationQueries() []Query {
	return []Query{
		{SQL: "CREATE TABLE IF NOT EXISTS events (rec_id INTEGER PRIMARY KEY, event_id TEXT NOT NULL UNIQUE CHECK(length(event_id) <= 20), category TEXT NOT NULL, severity TEXT NOT NULL CHECK(severity IN ('info', 'warning', 'error', 'critical')), priority INTEGER NOT NULL CHECK(priority >= 0 AND priority <= 5), created_at TEXT NOT NULL, source TEXT NOT NULL)"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_events_category ON events(category)"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_events_severity ON events(severity)"},
		{SQL: "CREATE TABLE IF NOT EXISTS lifecycle (rec_id INTEGER PRIMARY KEY, event_id TEXT NOT NULL UNIQUE CHECK(length(event_id) <= 20), event_rec_id INTEGER NOT NULL UNIQUE, acknowledged INTEGER NOT NULL CHECK(acknowledged IN (0, 1)), acknowledged_by TEXT NOT NULL, acknowledged_at TEXT NOT NULL, muted INTEGER NOT NULL CHECK(muted IN (0, 1)), muted_by TEXT NOT NULL, muted_at TEXT NOT NULL)"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_lifecycle_acknowledged ON lifecycle(acknowledged)"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_lifecycle_muted ON lifecycle(muted)"},
		{SQL: "CREATE TABLE IF NOT EXISTS event_state (rec_id INTEGER PRIMARY KEY, event_id TEXT NOT NULL UNIQUE CHECK(length(event_id) <= 20), event_rec_id INTEGER NOT NULL UNIQUE, status TEXT NOT NULL CHECK(status IN ('high', 'low', 'normal', 'active', 'failure')), message TEXT NOT NULL CHECK(length(message) <= 256))"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_event_state_status ON event_state(status)"},
		{SQL: "CREATE TABLE IF NOT EXISTS occurrences (rec_id INTEGER PRIMARY KEY AUTOINCREMENT, event_id TEXT NOT NULL CHECK(length(event_id) <= 20), event_rec_id INTEGER NOT NULL, ts TEXT NOT NULL, value_type TEXT NOT NULL CHECK(value_type IN ('int', 'float', 'text', 'bool')), value_num REAL, value_text TEXT, value_bool INTEGER CHECK(value_bool IN (0, 1)), value_unit TEXT)"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_occ_event_id ON occurrences(event_id)"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_occ_event ON occurrences(event_rec_id)"},
		{SQL: "CREATE TABLE IF NOT EXISTS event_meta (rec_id INTEGER PRIMARY KEY AUTOINCREMENT, event_id TEXT NOT NULL CHECK(length(event_id) <= 20), key TEXT NOT NULL, value TEXT NOT NULL, UNIQUE(event_id, key))"},
		{SQL: "CREATE INDEX IF NOT EXISTS idx_event_meta_event_id ON event_meta(event_id)"},
		{SQL: "CREATE TABLE IF NOT EXISTS runtime_state (key TEXT PRIMARY KEY, value TEXT NOT NULL)"},
	}
}
