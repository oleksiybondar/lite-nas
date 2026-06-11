package dto

import "lite-nas/shared/loggingmanager/enum"

// EventRow models one row in the events table.
//
// Contract:
//   - EventID is a generated business identifier in "<prefix>_<seq>" format.
//   - EventID length must not exceed 20 characters.
//   - Priority uses range 0..5, where lower values mean higher priority.
//
// Architectural role:
//   - This DTO maps persistence fields exactly and carries no SQL behavior.
type EventRow struct {
	RecID     int64
	EventID   string
	Category  string
	Severity  enum.Severity
	Priority  int
	CreatedAt string
	Source    string
}
