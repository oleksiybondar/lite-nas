package query

import (
	"strings"
	"testing"
)

func TestBuildSchemaMigrationQueriesIncludesCoreTables(t *testing.T) {
	t.Parallel()
	sqlText := flattenSQL(BuildSchemaMigrationQueries())
	for _, snippet := range []string{
		"CREATE TABLE IF NOT EXISTS events",
		"CREATE TABLE IF NOT EXISTS lifecycle",
		"CREATE TABLE IF NOT EXISTS event_state",
		"CREATE TABLE IF NOT EXISTS occurrences",
		"CREATE TABLE IF NOT EXISTS runtime_state",
	} {
		if !strings.Contains(sqlText, snippet) {
			t.Fatalf("missing migration snippet %q", snippet)
		}
	}
}

func TestBuildRuntimeStateSeedQueries(t *testing.T) {
	t.Parallel()
	queries := BuildRuntimeStateSeedQueries(7, 9, "event")
	if len(queries) != 3 {
		t.Fatalf("len(queries) = %d, want 3", len(queries))
	}

	assertSeedPair(t, queries[0], RuntimeStateCurrentEventRecIDKey, "7")
	assertSeedPair(t, queries[1], RuntimeStateCurrentEventSeqKey, "9")
	assertSeedPair(t, queries[2], RuntimeStateEventIDPrefixKey, "event")
}

func flattenSQL(queries []Query) string {
	parts := make([]string, 0, len(queries))
	for _, query := range queries {
		parts = append(parts, query.SQL)
	}
	return strings.Join(parts, "\n")
}

func assertSeedPair(t *testing.T, query Query, wantKey string, wantValue string) {
	t.Helper()
	if got := query.Args[0]; got != wantKey {
		t.Fatalf("key = %#v, want %#v", got, wantKey)
	}
	if got := query.Args[1]; got != wantValue {
		t.Fatalf("value = %#v, want %#v", got, wantValue)
	}
}
