package query

import (
	"testing"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
)

func TestInsertOccurrenceBoolMapping(t *testing.T) {
	t.Parallel()

	value := true
	query := InsertOccurrence(dto.OccurrenceRow{
		EventID:    "perf_10",
		EventRecID: 10,
		Timestamp:  "2026-05-11T12:01:00Z",
		ValueType:  enum.ValueTypeBool,
		ValueBool:  &value,
	})

	if len(query.Args) != 8 {
		t.Fatalf("len(query.Args) = %d, want 8", len(query.Args))
	}
	mapped, ok := query.Args[6].(*int)
	if !ok || *mapped != 1 {
		t.Fatalf("bool mapping = %#v, want *int(1)", query.Args[6])
	}
}
