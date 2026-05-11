package loggingmanager

// Query represents one SQL statement and its positional arguments.
//
// Contract:
//   - SQL must contain a non-empty statement string.
//   - Args are bound positionally by the SQL driver when the statement executes.
//
// Architectural role:
//   - Query is the transport unit between producers, batch writer, and
//     transaction executor.
type Query struct {
	SQL  string
	Args []any
}

// TransactionSQL contains one transaction unit prepared for execution.
//
// Contract:
//   - Begin and Commit are framing markers for transaction boundaries.
//   - Queries are executed in-order as one atomic persistence unit by the
//     executor.
//
// Architectural role:
//   - This struct is data-only. It does not execute SQL and does not own driver
//     resources.
type TransactionSQL struct {
	Begin   string
	Queries []Query
	Commit  string
}

// BuildTransactionSQL creates a transaction envelope that frames the provided
// queries between raw SQL BEGIN/COMMIT markers.
//
// Behavior:
//   - Copies the input slice header and values so caller-side append/re-slice
//     operations on the original slice do not mutate the returned batch list.
//
// Preconditions:
//   - Caller is responsible for supplying semantically valid queries.
//
// Side effects:
//   - None. This function performs no I/O and allocates only in-memory data.
func BuildTransactionSQL(queries []Query) TransactionSQL {
	copiedQueries := make([]Query, len(queries))
	copy(copiedQueries, queries)

	return TransactionSQL{
		Begin:   "BEGIN",
		Queries: copiedQueries,
		Commit:  "COMMIT",
	}
}
