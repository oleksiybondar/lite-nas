package dto

import "lite-nas/shared/loggingmanager/enum"

// OccurrenceRow models one row in the occurrences table.
type OccurrenceRow struct {
	RecID      int64
	EventID    string
	EventRecID int64
	Timestamp  string
	ValueType  enum.ValueType
	ValueNum   *float64
	ValueText  *string
	ValueBool  *bool
	ValueUnit  *string
}
