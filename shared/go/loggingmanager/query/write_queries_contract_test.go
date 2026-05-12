package query

import (
	"strings"
	"testing"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
)

func TestWriteQueriesContractCoverage(t *testing.T) {
	t.Parallel()

	for _, tc := range writeQueryContractCasesData {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertWriteQueryContract(t, tc.query, tc.sqlMustHave, tc.argsMustHave)
		})
	}
}

type writeQueryContractCase struct {
	name         string
	query        Query
	sqlMustHave  string
	argsMustHave int
}

var writeQueryContractCasesData = []writeQueryContractCase{
	{
		name: "upsert event",
		query: UpsertEvent(dto.EventRow{
			RecID:     1,
			EventID:   "perf_1",
			Category:  "system",
			Severity:  enum.SeverityWarning,
			Priority:  2,
			CreatedAt: "2026-05-12T10:00:00Z",
			Source:    "unknown",
		}),
		sqlMustHave:  "INSERT INTO events",
		argsMustHave: 7,
	},
	{
		name: "upsert lifecycle",
		query: UpsertLifecycle(dto.LifecycleRow{
			RecID:          1,
			EventID:        "perf_1",
			EventRecID:     1,
			Acknowledged:   false,
			AcknowledgedBy: "",
			AcknowledgedAt: "2026-05-12T10:00:00Z",
			Muted:          false,
			MutedBy:        "",
			MutedAt:        "2026-05-12T10:00:00Z",
		}),
		sqlMustHave:  "INSERT INTO lifecycle",
		argsMustHave: 9,
	},
	{
		name: "upsert event state",
		query: UpsertEventState(dto.EventStateRow{
			RecID:      1,
			EventID:    "perf_1",
			EventRecID: 1,
			Status:     enum.StatusActive,
			Message:    "",
		}),
		sqlMustHave:  "INSERT INTO event_state",
		argsMustHave: 5,
	},
	{
		name: "insert occurrence",
		query: InsertOccurrence(dto.OccurrenceRow{
			EventID:    "perf_1",
			EventRecID: 1,
			Timestamp:  "2026-05-12T10:00:00Z",
			ValueType:  enum.ValueTypeFloat,
			ValueNum:   float64TestPtr(99.9),
			ValueText:  stringTestPtr("text"),
			ValueBool:  boolTestPtr(true),
			ValueUnit:  stringTestPtr("%"),
		}),
		sqlMustHave:  "INSERT INTO occurrences",
		argsMustHave: 8,
	},
	{
		name: "upsert event meta",
		query: UpsertEventMeta(dto.EventMetaRow{
			EventID:   "perf_1",
			MetaKey:   "host",
			MetaValue: "rpi",
		}),
		sqlMustHave:  "INSERT INTO event_meta",
		argsMustHave: 3,
	},
	{
		name: "upsert runtime state",
		query: UpsertRuntimeState(dto.RuntimeStateRow{
			Key:   RuntimeStateCurrentEventSeqKey,
			Value: "1",
		}),
		sqlMustHave:  "INSERT INTO runtime_state",
		argsMustHave: 2,
	},
}

func float64TestPtr(value float64) *float64 {
	return &value
}

func stringTestPtr(value string) *string {
	return &value
}

func boolTestPtr(value bool) *bool {
	return &value
}

func assertWriteQueryContract(t *testing.T, builtQuery Query, sqlMustHave string, argsMustHave int) {
	t.Helper()
	if !strings.Contains(builtQuery.SQL, sqlMustHave) {
		t.Fatalf("query SQL %q must contain %q", builtQuery.SQL, sqlMustHave)
	}
	if len(builtQuery.Args) != argsMustHave {
		t.Fatalf("len(query.Args) = %d, want %d", len(builtQuery.Args), argsMustHave)
	}
}

func TestBuildRuntimeStateSeedQueriesContractCoverage(t *testing.T) {
	t.Parallel()

	queries := BuildRuntimeStateSeedQueries(1, 2, "event")
	if len(queries) != 3 {
		t.Fatalf("len(queries) = %d, want 3", len(queries))
	}
	for _, builtQuery := range queries {
		if !strings.Contains(builtQuery.SQL, "INSERT OR IGNORE INTO runtime_state") {
			t.Fatalf("unexpected seed query SQL: %q", builtQuery.SQL)
		}
		if len(builtQuery.Args) != 2 {
			t.Fatalf("len(seed query args) = %d, want 2", len(builtQuery.Args))
		}
	}
}
