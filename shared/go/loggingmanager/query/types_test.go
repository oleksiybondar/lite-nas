package query

import "testing"

func TestBuildTransactionSQL(t *testing.T) {
	t.Parallel()

	queries := []Query{{SQL: "INSERT INTO events(event_id) VALUES(?)", Args: []any{"cpu.high"}}}
	tx := BuildTransactionSQL(queries)
	if tx.Begin != "BEGIN" || tx.Commit != "COMMIT" {
		t.Fatalf("unexpected tx markers: %#v", tx)
	}
	if len(tx.Queries) != 1 {
		t.Fatalf("len(tx.Queries) = %d, want 1", len(tx.Queries))
	}
}
