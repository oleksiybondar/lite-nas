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
