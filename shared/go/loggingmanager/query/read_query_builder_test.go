package query

import (
	"strings"
	"testing"

	"lite-nas/shared/loggingmanager/dto"
)

func TestBuildListEventsQueryAppliesDefaultPageSize(t *testing.T) {
	t.Parallel()

	query, err := BuildListEventsQuery(dto.ListEventsInput{
		Page: 1,
	})
	if err != nil {
		t.Fatalf("BuildListEventsQuery() error = %v", err)
	}

	if !strings.Contains(query.SQL, "LIMIT ? OFFSET ?") {
		t.Fatalf("expected pagination placeholders, sql=%q", query.SQL)
	}
	if len(query.Args) != 2 {
		t.Fatalf("len(query.Args) = %d, want 2", len(query.Args))
	}
	if got := query.Args[0]; got != defaultListEventsPageSize {
		t.Fatalf("limit arg = %v, want %d", got, defaultListEventsPageSize)
	}
	if got := query.Args[1]; got != 0 {
		t.Fatalf("offset arg = %v, want 0", got)
	}
}

func TestBuildListEventsQueryBuildsFiltersAndOffset(t *testing.T) {
	t.Parallel()

	query, err := BuildListEventsQuery(dto.ListEventsInput{
		Page:     2,
		PageSize: 10,
		Filters: []dto.Filter{
			{Key: dto.FilterKeyCategory, Condition: dto.FilterConditionEQ, Values: []string{"system"}},
			{Key: dto.FilterKeyAcknowledged, Condition: dto.FilterConditionEQ, Values: []string{"false"}},
		},
	})
	if err != nil {
		t.Fatalf("BuildListEventsQuery() error = %v", err)
	}

	assertWhereClause(t, query.SQL, "e.category = ? AND l.acknowledged = ?")
	assertQueryArgs(t, query.Args, []any{"system", 0, 10, 10})
}

func TestBuildListEventsQueryRejectsInvalidBooleanFilter(t *testing.T) {
	t.Parallel()

	_, err := BuildListEventsQuery(dto.ListEventsInput{
		Page: 1,
		Filters: []dto.Filter{
			{Key: dto.FilterKeyAcknowledged, Condition: dto.FilterConditionEQ, Values: []string{"yes"}},
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBuildListActiveEventsQueryForcesNonNormalStates(t *testing.T) {
	t.Parallel()

	query, err := BuildListActiveEventsQuery(dto.ListEventsInput{Page: 1})
	if err != nil {
		t.Fatalf("BuildListActiveEventsQuery() error = %v", err)
	}
	if !strings.Contains(query.SQL, "s.status IN (?, ?, ?, ?)") {
		t.Fatalf("active-state filter missing, sql=%q", query.SQL)
	}
}

func TestBuildListActiveUnacknowledgedEventsQueryForcesAcknowledgedFalse(t *testing.T) {
	t.Parallel()

	query, err := BuildListActiveUnacknowledgedEventsQuery(dto.ListEventsInput{Page: 1})
	if err != nil {
		t.Fatalf("BuildListActiveUnacknowledgedEventsQuery() error = %v", err)
	}
	if !strings.Contains(query.SQL, "l.acknowledged = ?") {
		t.Fatalf("unacknowledged filter missing, sql=%q", query.SQL)
	}
}

func TestBuildGetEventHistoryQuery(t *testing.T) {
	t.Parallel()

	query := BuildGetEventHistoryQuery(dto.GetEventHistoryInput{EventID: "perf_1"})
	if !strings.Contains(query.SQL, "WHERE e.event_id = ?") {
		t.Fatalf("history event filter missing, sql=%q", query.SQL)
	}
	if len(query.Args) != 1 || query.Args[0] != "perf_1" {
		t.Fatalf("unexpected args: %#v", query.Args)
	}
}

func assertWhereClause(t *testing.T, sql string, expected string) {
	t.Helper()
	if !strings.Contains(sql, expected) {
		t.Fatalf("unexpected where clause, sql=%q", sql)
	}
}

func assertQueryArgs(t *testing.T, got []any, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(args) = %d, want %d", len(got), len(want))
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("args[%d] = %v, want %v; args=%#v", idx, got[idx], want[idx], got)
		}
	}
}
