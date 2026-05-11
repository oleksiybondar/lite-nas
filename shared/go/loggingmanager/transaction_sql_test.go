package loggingmanager

import "testing"

func TestBuildTransactionSQLReturnsTransactionPieces(t *testing.T) {
	t.Parallel()

	queries := []Query{
		{SQL: "INSERT INTO events(event_id) VALUES(?)", Args: []any{"cpu_high"}},
		{SQL: "UPDATE events SET state = ? WHERE event_id = ?", Args: []any{"active", "cpu_high"}},
	}

	transactionSQL := BuildTransactionSQL(queries)

	if transactionSQL.Begin != "BEGIN" {
		t.Fatalf("transactionSQL.Begin = %q, want %q", transactionSQL.Begin, "BEGIN")
	}
	if transactionSQL.Commit != "COMMIT" {
		t.Fatalf("transactionSQL.Commit = %q, want %q", transactionSQL.Commit, "COMMIT")
	}
	if len(transactionSQL.Queries) != 2 {
		t.Fatalf("len(transactionSQL.Queries) = %d, want 2", len(transactionSQL.Queries))
	}
}

func TestBuildTransactionSQLCopiesQueries(t *testing.T) {
	t.Parallel()

	queries := []Query{{SQL: "INSERT INTO occurrences(message) VALUES(?)", Args: []any{"initial"}}}
	transactionSQL := BuildTransactionSQL(queries)

	queries[0].SQL = "UPDATED"
	if transactionSQL.Queries[0].SQL != "INSERT INTO occurrences(message) VALUES(?)" {
		t.Fatalf("transactionSQL.Queries[0].SQL = %q, want original query", transactionSQL.Queries[0].SQL)
	}
}
