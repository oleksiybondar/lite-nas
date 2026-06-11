package query

import (
	"strings"
	"testing"
)

func TestDeleteOrphanOccurrences(t *testing.T) {
	t.Parallel()

	builtQuery := DeleteOrphanOccurrences()
	if !strings.Contains(builtQuery.SQL, "DELETE FROM occurrences") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if !strings.Contains(builtQuery.SQL, "events e WHERE e.event_id = occurrences.event_id") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if len(builtQuery.Args) != 0 {
		t.Fatalf("len(args) = %d, want 0", len(builtQuery.Args))
	}
}

func TestDeleteOldestEventsBeyondLimit(t *testing.T) {
	t.Parallel()

	builtQuery := DeleteOldestEventsBeyondLimit(5000)
	assertDeleteWithLimitQuery(t, builtQuery, "DELETE FROM events", "OFFSET ?", 5000)
}

func TestDeleteOrphanLifecycle(t *testing.T) {
	t.Parallel()

	builtQuery := DeleteOrphanLifecycle()
	if !strings.Contains(builtQuery.SQL, "DELETE FROM lifecycle") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if !strings.Contains(builtQuery.SQL, "events e WHERE e.event_id = lifecycle.event_id") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
}

func TestDeleteOrphanEventState(t *testing.T) {
	t.Parallel()

	builtQuery := DeleteOrphanEventState()
	if !strings.Contains(builtQuery.SQL, "DELETE FROM event_state") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if !strings.Contains(builtQuery.SQL, "events e WHERE e.event_id = event_state.event_id") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
}

func TestDeleteOrphanEventMeta(t *testing.T) {
	t.Parallel()

	builtQuery := DeleteOrphanEventMeta()
	if !strings.Contains(builtQuery.SQL, "DELETE FROM event_meta") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if !strings.Contains(builtQuery.SQL, "events e WHERE e.event_id = event_meta.event_id") {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if len(builtQuery.Args) != 0 {
		t.Fatalf("len(args) = %d, want 0", len(builtQuery.Args))
	}
}

func TestDeleteOccurrencesPerEventBeyondLimit(t *testing.T) {
	t.Parallel()

	builtQuery := DeleteOccurrencesPerEventBeyondLimit(500)
	assertDeleteWithLimitQuery(
		t,
		builtQuery,
		"DELETE FROM occurrences",
		"ROW_NUMBER() OVER (PARTITION BY event_id ORDER BY rec_id DESC)",
		500,
	)
}

func assertDeleteWithLimitQuery(t *testing.T, builtQuery Query, wantDeleteSQL string, wantFilterSQL string, wantLimit int) {
	t.Helper()

	if !strings.Contains(builtQuery.SQL, wantDeleteSQL) {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if !strings.Contains(builtQuery.SQL, wantFilterSQL) {
		t.Fatalf("unexpected SQL: %q", builtQuery.SQL)
	}
	if len(builtQuery.Args) != 1 {
		t.Fatalf("len(args) = %d, want 1", len(builtQuery.Args))
	}
	if got := builtQuery.Args[0]; got != wantLimit {
		t.Fatalf("args[0] = %v, want %d", got, wantLimit)
	}
}
