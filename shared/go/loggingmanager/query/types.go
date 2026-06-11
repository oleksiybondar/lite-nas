package query

// Query represents one SQL statement with positional arguments.
type Query struct {
	SQL  string
	Args []any
}

// TransactionSQL contains one transaction unit prepared for execution.
type TransactionSQL struct {
	Begin   string
	Queries []Query
	Commit  string
}

// BuildTransactionSQL frames the provided queries with BEGIN/COMMIT markers.
func BuildTransactionSQL(queries []Query) TransactionSQL {
	copiedQueries := make([]Query, len(queries))
	copy(copiedQueries, queries)

	return TransactionSQL{
		Begin:   "BEGIN",
		Queries: copiedQueries,
		Commit:  "COMMIT",
	}
}
