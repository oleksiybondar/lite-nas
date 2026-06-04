package loggingmanager

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanager/model"
	"lite-nas/shared/loggingmanager/query"

	_ "modernc.org/sqlite"
)

func TestListEventsReturnsCreatedEvent(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRig(t)
	eventID, err := rig.core.CreateEvent(dto.CreateEventInput{Category: "system"})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	rig.flush(t)
	waitForListedEvents(t, rig.core, 1)

	events, err := rig.core.ListEvents(dto.ListEventsInput{Page: 1})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("ListEvents len = %d, want 1", len(events))
	}
	if events[0].Event.EventID != eventID {
		t.Fatalf("eventID = %q, want %q", events[0].Event.EventID, eventID)
	}
}

func TestListEventsPageReturnsTotalCountIndependentOfPageSize(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRig(t)
	if _, err := rig.core.CreateEvent(dto.CreateEventInput{Category: "system"}); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if _, err := rig.core.CreateEvent(dto.CreateEventInput{Category: "system"}); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	rig.flush(t)
	waitForListedEvents(t, rig.core, 2)

	page, err := rig.core.ListEventsPage(dto.ListEventsInput{Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("ListEventsPage() error = %v", err)
	}
	if page.TotalCount != 2 {
		t.Fatalf("total_count = %d, want 2", page.TotalCount)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(page.Items))
	}
}

func TestGetEventReturnsCreatedEventByEventID(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRig(t)
	eventID := mustCreateEvent(t, rig, "system")

	item, found, err := rig.core.GetEvent(dto.GetEventHistoryInput{EventID: eventID})
	if err != nil {
		t.Fatalf("GetEvent() error = %v", err)
	}
	if !found {
		t.Fatal("GetEvent() found = false, want true")
	}
	if item.Event.EventID != eventID {
		t.Fatalf("eventID = %q, want %q", item.Event.EventID, eventID)
	}
}

func TestSetStateUpdatesEventState(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRig(t)
	eventID := mustCreateEvent(t, rig, "system")

	message := "node is under pressure"
	err := rig.core.SetState(dto.SetStateInput{
		EventID: eventID,
		Status:  enum.StatusFailure,
		Message: &message,
	})
	if err != nil {
		t.Fatalf("SetState() error = %v", err)
	}
	rig.flush(t)

	events, err := rig.core.ListEvents(dto.ListEventsInput{Page: 1})
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(events) != 1 || events[0].State.Status != enum.StatusFailure || events[0].State.Message != message {
		t.Fatal("state was not updated")
	}
}

func TestAcknowledgeRemovesEventFromUnacknowledgedList(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRig(t)
	eventID := mustCreateEvent(t, rig, "system")

	err := rig.core.AcknowledgeEvent(dto.AcknowledgeEventInput{
		EventID:        eventID,
		AcknowledgedBy: "ops",
	})
	if err != nil {
		t.Fatalf("AcknowledgeEvent() error = %v", err)
	}
	rig.flush(t)

	unacked, err := rig.core.ListActiveUnacknowledgedEvents(dto.ListEventsInput{Page: 1})
	if err != nil {
		t.Fatalf("ListActiveUnacknowledgedEvents() error = %v", err)
	}
	if len(unacked) != 0 {
		t.Fatalf("ListActiveUnacknowledgedEvents len = %d, want 0", len(unacked))
	}
}

func TestActiveEventReadCases(t *testing.T) {
	t.Parallel()

	for _, tc := range activeEventReadTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rig := newRunningCoreRig(t)
			eventID := mustCreateEvent(t, rig, "system")
			tc.act(t, rig, eventID)
			tc.assert(t, mustGetSingleActiveEvent(t, rig.core))
		})
	}
}

type activeEventReadTestCase struct {
	name   string
	act    func(t *testing.T, rig runningCoreRig, eventID string)
	assert func(t *testing.T, event model.Event)
}

func activeEventReadTestCases() []activeEventReadTestCase {
	return []activeEventReadTestCase{
		{
			name:   "mute marks lifecycle as muted",
			act:    actMuteEvent,
			assert: assertLifecycleMuted,
		},
		{
			name:   "add occurrence sets last value",
			act:    actAddOccurrence,
			assert: assertLastOccurrencePresent,
		},
	}
}

func actMuteEvent(t *testing.T, rig runningCoreRig, eventID string) {
	t.Helper()
	if err := rig.core.MuteEvent(dto.MuteEventInput{EventID: eventID, MutedBy: "ops"}); err != nil {
		t.Fatalf("MuteEvent() error = %v", err)
	}
	rig.flush(t)
}

func assertLifecycleMuted(t *testing.T, event model.Event) {
	t.Helper()
	if !event.Lifecycle.Muted {
		t.Fatal("expected muted lifecycle")
	}
}

func actAddOccurrence(t *testing.T, rig runningCoreRig, eventID string) {
	t.Helper()
	valueNum := 92.5
	valueUnit := "%"
	if err := rig.core.AddOccurrence(dto.OccurrenceRow{
		EventID:    eventID,
		EventRecID: 1,
		Timestamp:  fixedClock().Format(time.RFC3339),
		ValueType:  enum.ValueTypeFloat,
		ValueNum:   &valueNum,
		ValueUnit:  &valueUnit,
	}); err != nil {
		t.Fatalf("AddOccurrence() error = %v", err)
	}
	rig.flush(t)
}

func assertLastOccurrencePresent(t *testing.T, event model.Event) {
	t.Helper()
	if event.LastValue == nil {
		t.Fatal("expected last occurrence")
	}
}

func TestCleanupSucceeds(t *testing.T) {
	t.Parallel()
	rig := newRunningCoreRig(t)
	if err := rig.core.Cleanup(context.Background()); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
}

func TestCleanupDeletesOldestEventsBeyondMaxEvents(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRigWithLimits(t, 2, 1000)

	firstID := createEventAndWaitCount(t, rig, "system", 1)
	secondID := createEventAndWaitCount(t, rig, "system", 2)
	thirdID := createEventAndWaitCount(t, rig, "system", 3)

	if err := rig.core.Cleanup(context.Background()); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
	waitForListedEvents(t, rig.core, 2)

	if _, found, err := rig.core.GetEvent(dto.GetEventHistoryInput{EventID: firstID}); err != nil {
		t.Fatalf("GetEvent(%q) error = %v", firstID, err)
	} else if found {
		t.Fatalf("GetEvent(%q) found = true, want false", firstID)
	}

	assertEventStillPresent(t, rig.core, secondID)
	assertEventStillPresent(t, rig.core, thirdID)
}

func TestCleanupKeepsLatestOccurrencesPerEvent(t *testing.T) {
	t.Parallel()

	rig := newRunningCoreRigWithLimits(t, 10, 2)
	eventID := mustCreateEvent(t, rig, "system")

	addOccurrenceWithValue(t, rig, eventID, 1)
	addOccurrenceWithValue(t, rig, eventID, 2)
	addOccurrenceWithValue(t, rig, eventID, 3)
	addOccurrenceWithValue(t, rig, eventID, 4)

	if got := countOccurrencesByEventID(t, rig.core.db, eventID); got != 4 {
		t.Fatalf("occurrence count before cleanup = %d, want 4", got)
	}

	if err := rig.core.Cleanup(context.Background()); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if got := countOccurrencesByEventID(t, rig.core.db, eventID); got != 2 {
		t.Fatalf("occurrence count after cleanup = %d, want 2", got)
	}
}

func TestCoreMethodsReturnNotFoundForUnknownEvent(t *testing.T) {
	t.Parallel()

	core := newCoreWithInitializedDB(t)

	err := core.SetState(dto.SetStateInput{
		EventID: "event_1",
		Status:  enum.StatusFailure,
	})
	if !errors.Is(err, errEventNotFound) {
		t.Fatalf("SetState() err = %v, want %v", err, errEventNotFound)
	}

	err = core.AcknowledgeEvent(dto.AcknowledgeEventInput{
		EventID:        "event_1",
		AcknowledgedBy: "ops",
	})
	if !errors.Is(err, errEventNotFound) {
		t.Fatalf("AcknowledgeEvent() err = %v, want %v", err, errEventNotFound)
	}

	err = core.MuteEvent(dto.MuteEventInput{
		EventID: "event_1",
		MutedBy: "ops",
	})
	if !errors.Is(err, errEventNotFound) {
		t.Fatalf("MuteEvent() err = %v, want %v", err, errEventNotFound)
	}
}

func TestNewCoreLoadsPreSeededRuntimeState(t *testing.T) {
	t.Parallel()

	db := openTestSQLiteDB(t)
	defer func() { _ = db.Close() }()

	if err := executeQueryBatch(context.Background(), db, query.BuildSchemaMigrationQueries()); err != nil {
		t.Fatalf("schema migration error = %v", err)
	}
	if err := executeQueryBatch(context.Background(), db, []query.Query{
		query.UpsertRuntimeState(dto.RuntimeStateRow{Key: query.RuntimeStateCurrentEventRecIDKey, Value: "5"}),
		query.UpsertRuntimeState(dto.RuntimeStateRow{Key: query.RuntimeStateCurrentEventSeqKey, Value: "8"}),
		query.UpsertRuntimeState(dto.RuntimeStateRow{Key: query.RuntimeStateEventIDPrefixKey, Value: "boot"}),
	}); err != nil {
		t.Fatalf("seed runtime state error = %v", err)
	}

	writerInputCh := make(chan WriteRequest, 8)
	core, err := NewCore(context.Background(), CoreDeps{
		DB:             db,
		WriterInputCh:  writerInputCh,
		Clock:          fixedClock,
		MaxEvents:      10,
		MaxOccurrences: 1000,
		EventIDPrefix:  "ignored",
	})
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}

	eventID, err := core.CreateEvent(dto.CreateEventInput{Category: "system"})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if eventID != "boot_9" {
		t.Fatalf("eventID = %q, want %q", eventID, "boot_9")
	}
}

func TestNewCoreReloadsRuntimeStateSeededByCreateEventTransaction(t *testing.T) {
	t.Parallel()

	db := openTestSQLiteDB(t)
	defer func() { _ = db.Close() }()

	writerInputCh, writerFlushCh, _, _ := startRunningWriter(t, db)
	core := mustNewCoreWithDeps(t, db, writerInputCh, "alert")
	assertCreatedEventID(t, core, dto.CreateEventInput{
		EventID:  "t1778675852000000000",
		Category: "system",
	}, "t1778675852000000000")
	flushWriter(t, writerFlushCh)

	reloadedCore := mustNewCoreWithDeps(t, db, make(chan WriteRequest, 8), "ignored")
	assertCreatedEventID(t, reloadedCore, dto.CreateEventInput{Category: "system"}, "alert_2")
}

func mustNewCoreWithDeps(t *testing.T, db *sql.DB, writerInputCh chan WriteRequest, prefix string) *Core {
	t.Helper()

	core, err := NewCore(context.Background(), CoreDeps{
		DB:             db,
		WriterInputCh:  writerInputCh,
		Clock:          fixedClock,
		MaxEvents:      10,
		MaxOccurrences: 1000,
		EventIDPrefix:  prefix,
	})
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}
	return core
}

func assertCreatedEventID(t *testing.T, core *Core, input dto.CreateEventInput, want string) {
	t.Helper()

	eventID, err := core.CreateEvent(input)
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	if eventID != want {
		t.Fatalf("eventID = %q, want %q", eventID, want)
	}
}

func newCoreWithInitializedDB(t *testing.T) *Core {
	t.Helper()

	db := openTestSQLiteDB(t)
	t.Cleanup(func() { _ = db.Close() })

	writerInputCh := make(chan WriteRequest, 8)
	core, err := NewCore(context.Background(), CoreDeps{
		DB:             db,
		WriterInputCh:  writerInputCh,
		Clock:          fixedClock,
		MaxEvents:      10,
		MaxOccurrences: 1000,
		EventIDPrefix:  "event",
	})
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}
	return core
}

type runningCoreRig struct {
	core    *Core
	flushCh chan struct{}
	cancel  context.CancelFunc
	done    <-chan error
}

func newRunningCoreRig(t *testing.T) runningCoreRig {
	return newRunningCoreRigWithLimits(t, 10, 1000)
}

func newRunningCoreRigWithLimits(t *testing.T, maxEvents int, maxOccurrences int) runningCoreRig {
	t.Helper()

	db := openTestSQLiteDB(t)
	t.Cleanup(func() { _ = db.Close() })

	writerInputCh, writerFlushCh, cancel, done := startRunningWriter(t, db)
	core, err := newCoreForRigWithLimits(writerInputCh, db, maxEvents, maxOccurrences)
	if err != nil {
		t.Fatalf("NewCore() error = %v", err)
	}

	return runningCoreRig{
		core:    core,
		flushCh: writerFlushCh,
		cancel:  cancel,
		done:    done,
	}
}

func startRunningWriter(t *testing.T, db *sql.DB) (chan WriteRequest, chan struct{}, context.CancelFunc, <-chan error) {
	t.Helper()

	writerInputCh := make(chan WriteRequest, 64)
	writerFlushCh := make(chan struct{}, 8)
	executor, err := NewSQLTransactionExecutor(db)
	if err != nil {
		t.Fatalf("NewSQLTransactionExecutor() error = %v", err)
	}
	writer, err := NewWriter(executor, DefaultTransactionBuilder{}, writerInputCh, writerFlushCh, 1, 1000)
	if err != nil {
		t.Fatalf("NewWriter() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := runWriterAsync(writer, ctx)
	t.Cleanup(func() {
		cancel()
		waitDone(t, done)
	})

	return writerInputCh, writerFlushCh, cancel, done
}

func newCoreForRigWithLimits(
	writerInputCh chan WriteRequest,
	db *sql.DB,
	maxEvents int,
	maxOccurrences int,
) (*Core, error) {
	return NewCore(context.Background(), CoreDeps{
		DB:             db,
		WriterInputCh:  writerInputCh,
		Clock:          fixedClock,
		MaxEvents:      maxEvents,
		MaxOccurrences: maxOccurrences,
		EventIDPrefix:  "alert",
	})
}

func (rig runningCoreRig) flush(t *testing.T) {
	t.Helper()
	flushWriter(t, rig.flushCh)
}

func mustCreateEvent(t *testing.T, rig runningCoreRig, category string) string {
	t.Helper()
	eventID, err := rig.core.CreateEvent(dto.CreateEventInput{Category: category})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	rig.flush(t)
	waitForListedEvents(t, rig.core, 1)
	return eventID
}

func createEventAndWaitCount(t *testing.T, rig runningCoreRig, category string, wantCount int) string {
	t.Helper()

	eventID, err := rig.core.CreateEvent(dto.CreateEventInput{Category: category})
	if err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}
	rig.flush(t)
	waitForListedEvents(t, rig.core, wantCount)
	return eventID
}

func addOccurrenceWithValue(t *testing.T, rig runningCoreRig, eventID string, value float64) {
	t.Helper()

	valueUnit := "%"
	if err := rig.core.AddOccurrence(dto.OccurrenceRow{
		EventID:    eventID,
		EventRecID: 1,
		Timestamp:  fixedClock().Format(time.RFC3339),
		ValueType:  enum.ValueTypeFloat,
		ValueNum:   &value,
		ValueUnit:  &valueUnit,
	}); err != nil {
		t.Fatalf("AddOccurrence() error = %v", err)
	}
	rig.flush(t)
}

func assertEventStillPresent(t *testing.T, core *Core, eventID string) {
	t.Helper()

	if _, found, err := core.GetEvent(dto.GetEventHistoryInput{EventID: eventID}); err != nil {
		t.Fatalf("GetEvent(%q) error = %v", eventID, err)
	} else if !found {
		t.Fatalf("GetEvent(%q) found = false, want true", eventID)
	}
}

func countOccurrencesByEventID(t *testing.T, db *sql.DB, eventID string) int {
	t.Helper()

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM occurrences WHERE event_id = ?", eventID).Scan(&count); err != nil {
		t.Fatalf("count occurrences query error = %v", err)
	}
	return count
}

func openTestSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "loggingmanager-test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	return db
}

func flushWriter(t *testing.T, flushCh chan<- struct{}) {
	t.Helper()
	select {
	case flushCh <- struct{}{}:
	case <-time.After(time.Second):
		t.Fatal("timed out sending writer flush signal")
	}
	time.Sleep(10 * time.Millisecond)
}

func waitForListedEvents(t *testing.T, core *Core, want int) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		items, err := core.ListEvents(dto.ListEventsInput{Page: 1})
		if err == nil && len(items) == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for listed events count %d", want)
}

func mustGetSingleActiveEvent(t *testing.T, core *Core) model.Event {
	t.Helper()

	active, err := core.ListActiveEvents(dto.ListEventsInput{Page: 1})
	if err != nil {
		t.Fatalf("ListActiveEvents() error = %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("ListActiveEvents len = %d, want 1", len(active))
	}
	return active[0]
}
