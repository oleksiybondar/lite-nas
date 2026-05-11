package query

import "lite-nas/shared/loggingmanager/dto"

// InsertOccurrence builds an insert query for one occurrences row.
func InsertOccurrence(row dto.OccurrenceRow) Query {
	return Query{
		SQL: "INSERT INTO occurrences (event_rec_id, ts, value_type, value_num, value_text, value_bool, value_unit) VALUES (?, ?, ?, ?, ?, ?, ?)",
		Args: []any{
			row.EventRecID,
			row.Timestamp,
			string(row.ValueType),
			row.ValueNum,
			row.ValueText,
			boolPtrToIntPtr(row.ValueBool),
			row.ValueUnit,
		},
	}
}
