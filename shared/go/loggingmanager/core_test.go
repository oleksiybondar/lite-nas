package loggingmanager

import (
	"database/sql"
	"testing"
	"time"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanager/query"
)

func TestNormalizeCoreDependenciesAppliesDefaults(t *testing.T) {
	t.Parallel()

	validate := mustInputValidator(t)
	writerInputCh := make(chan WriteRequest, 1)
	deps := CoreDeps{
		DB:             nil,
		WriterInputCh:  writerInputCh,
		Validator:      validate,
		Clock:          time.Now,
		MaxEvents:      10,
		MaxOccurrences: 100,
	}

	_, err := normalizeCoreDependencies(deps)
	if err == nil {
		t.Fatal("expected error for nil DB")
	}
}

func TestNormalizeCoreDependenciesSetsDefaultPrefixAndValidator(t *testing.T) {
	t.Parallel()

	writerInputCh := make(chan WriteRequest, 1)
	deps := CoreDeps{
		DB:             newTestDBPtr(),
		WriterInputCh:  writerInputCh,
		Clock:          time.Now,
		MaxEvents:      10,
		MaxOccurrences: 100,
	}

	normalized, err := normalizeCoreDependencies(deps)
	if err != nil {
		t.Fatalf("normalizeCoreDependencies() error = %v", err)
	}
	if normalized.EventIDPrefix != defaultEventPrefix {
		t.Fatalf("EventIDPrefix = %q, want %q", normalized.EventIDPrefix, defaultEventPrefix)
	}
	if normalized.Validator == nil {
		t.Fatal("expected non-nil validator")
	}
}

func TestNextEventIdentityWrapsRecIDAndSequence(t *testing.T) {
	t.Parallel()

	core := &Core{
		maxEvents:       3,
		currentEventRec: 3,
		currentEventSeq: query.EventIDMaxSequence,
		currentIDPrefix: "perf",
	}

	recID, seq, eventID, err := core.nextEventIdentity()
	if err != nil {
		t.Fatalf("nextEventIdentity() error = %v", err)
	}
	if recID != 1 {
		t.Fatalf("recID = %d, want 1", recID)
	}
	if seq != 0 {
		t.Fatalf("seq = %d, want 0", seq)
	}
	if eventID != "perf_0" {
		t.Fatalf("eventID = %q, want %q", eventID, "perf_0")
	}
}

func TestCreateEventEnqueuesThreeRequestsWithRuntimeUpdates(t *testing.T) {
	t.Parallel()

	core, writerInputCh := newCreateEventTestCore(t, "perf")

	eventID, err := core.CreateEvent(dto.CreateEventInput{
		Category: "system",
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if eventID != "perf_1" {
		t.Fatalf("eventID = %q, want perf_1", eventID)
	}

	requestOne := <-writerInputCh
	requestTwo := <-writerInputCh
	requestThree := <-writerInputCh

	if len(requestOne.RuntimeStateUpdates) != 3 {
		t.Fatalf("runtime updates len = %d, want 3", len(requestOne.RuntimeStateUpdates))
	}
	assertRuntimeStateUpdate(t, requestOne.RuntimeStateUpdates, query.RuntimeStateCurrentEventRecIDKey, "1")
	assertRuntimeStateUpdate(t, requestOne.RuntimeStateUpdates, query.RuntimeStateCurrentEventSeqKey, "1")
	assertRuntimeStateUpdate(t, requestOne.RuntimeStateUpdates, query.RuntimeStateEventIDPrefixKey, "perf")
	if requestTwo.Query.SQL == "" || requestThree.Query.SQL == "" {
		t.Fatal("expected non-empty lifecycle/state SQL")
	}
}

func TestCreateEventUsesProvidedEventIDAndStillAdvancesRuntimeState(t *testing.T) {
	t.Parallel()

	core, writerInputCh := newCreateEventTestCore(t, "perf")

	eventID, err := core.CreateEvent(dto.CreateEventInput{
		EventID:  "t1778675852000000000",
		Category: "system",
	})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if eventID != "t1778675852000000000" {
		t.Fatalf("eventID = %q, want %q", eventID, "t1778675852000000000")
	}

	requestOne := <-writerInputCh
	if len(requestOne.RuntimeStateUpdates) != 3 {
		t.Fatalf("runtime updates len = %d, want 3", len(requestOne.RuntimeStateUpdates))
	}
	assertRuntimeStateUpdate(t, requestOne.RuntimeStateUpdates, query.RuntimeStateCurrentEventRecIDKey, "1")
	assertRuntimeStateUpdate(t, requestOne.RuntimeStateUpdates, query.RuntimeStateCurrentEventSeqKey, "1")
}

func TestCreateEventAppliesDefaults(t *testing.T) {
	t.Parallel()

	core, writerInputCh := newCreateEventTestCore(t, "event")

	_, err := core.CreateEvent(dto.CreateEventInput{Category: "security"})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	eventRequest := <-writerInputCh
	if got := eventRequest.Query.Args[3]; got != string(enum.SeverityInfo) {
		t.Fatalf("severity arg = %v, want %q", got, enum.SeverityInfo)
	}
	if got := eventRequest.Query.Args[4]; got != defaultEventPriority {
		t.Fatalf("priority arg = %v, want %d", got, defaultEventPriority)
	}
	if got := eventRequest.Query.Args[6]; got != defaultEventSource {
		t.Fatalf("source arg = %v, want %q", got, defaultEventSource)
	}
}

func TestAddOccurrenceEnqueuesOccurrenceWrite(t *testing.T) {
	t.Parallel()

	writerInputCh := make(chan WriteRequest, 1)
	core := &Core{writerInputCh: writerInputCh}
	value := 10.5

	err := core.AddOccurrence(dto.OccurrenceRow{
		EventID:    "perf_1",
		EventRecID: 1,
		Timestamp:  "2026-05-12T10:00:00Z",
		ValueType:  enum.ValueTypeFloat,
		ValueNum:   &value,
	})
	if err != nil {
		t.Fatalf("AddOccurrence() error = %v", err)
	}

	request := <-writerInputCh
	if !request.TouchesOccurrences {
		t.Fatal("expected TouchesOccurrences=true")
	}
	if request.Query.SQL == "" {
		t.Fatal("expected non-empty SQL")
	}
}

func TestApplyRuntimeStateKeyValueSetsFields(t *testing.T) {
	t.Parallel()

	for _, testCase := range runtimeStateApplySuccessCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			core := &Core{currentIDPrefix: "event"}
			err := core.applyRuntimeStateKeyValue(testCase.key, testCase.value)
			if err != nil {
				t.Fatalf("applyRuntimeStateKeyValue() error = %v", err)
			}
			testCase.assert(t, core)
		})
	}
}

func TestApplyRuntimeStateKeyValueRejectsInvalidRecID(t *testing.T) {
	t.Parallel()

	core := &Core{}
	if err := core.applyRuntimeStateKeyValue(query.RuntimeStateCurrentEventRecIDKey, "x"); err == nil {
		t.Fatal("expected parse error for rec id")
	}
}

func TestApplyRuntimeStateKeyValueRejectsInvalidSeq(t *testing.T) {
	t.Parallel()

	core := &Core{}
	if err := core.applyRuntimeStateKeyValue(query.RuntimeStateCurrentEventSeqKey, "x"); err == nil {
		t.Fatal("expected parse error for seq")
	}
}

func assertRuntimeStateUpdate(t *testing.T, rows []dto.RuntimeStateRow, key string, value string) {
	t.Helper()
	for _, row := range rows {
		if row.Key == key {
			if row.Value != value {
				t.Fatalf("runtime state value for %q = %q, want %q", key, row.Value, value)
			}
			return
		}
	}
	t.Fatalf("runtime state key %q not found", key)
}

func fixedClock() time.Time {
	return time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
}

func newTestDBPtr() *sql.DB { return &sql.DB{} }

func newCreateEventTestCore(t *testing.T, eventIDPrefix string) (*Core, chan WriteRequest) {
	t.Helper()

	validate := mustInputValidator(t)
	writerInputCh := make(chan WriteRequest, 3)
	return &Core{
		writerInputCh:   writerInputCh,
		validator:       validate,
		clock:           fixedClock,
		maxEvents:       100,
		currentIDPrefix: eventIDPrefix,
	}, writerInputCh
}

type runtimeStateApplySuccessCase struct {
	name   string
	key    string
	value  string
	assert func(t *testing.T, core *Core)
}

var runtimeStateApplySuccessCases = []runtimeStateApplySuccessCase{
	{
		name:  "sets current event rec id",
		key:   query.RuntimeStateCurrentEventRecIDKey,
		value: "42",
		assert: func(t *testing.T, core *Core) {
			t.Helper()
			if core.currentEventRec != 42 {
				t.Fatalf("currentEventRec = %d, want 42", core.currentEventRec)
			}
		},
	},
	{
		name:  "sets current event seq",
		key:   query.RuntimeStateCurrentEventSeqKey,
		value: "7",
		assert: func(t *testing.T, core *Core) {
			t.Helper()
			if core.currentEventSeq != 7 {
				t.Fatalf("currentEventSeq = %d, want 7", core.currentEventSeq)
			}
		},
	},
	{
		name:  "sets event id prefix",
		key:   query.RuntimeStateEventIDPrefixKey,
		value: "perf",
		assert: func(t *testing.T, core *Core) {
			t.Helper()
			if core.currentIDPrefix != "perf" {
				t.Fatalf("currentIDPrefix = %q, want perf", core.currentIDPrefix)
			}
		},
	},
}
