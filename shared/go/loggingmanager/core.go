package loggingmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"lite-nas/shared/loggingmanager/dto"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanager/query"
)

const (
	defaultEventPriority = 2
	defaultEventSource   = "unknown"
	defaultEventPrefix   = "event"
)

var (
	errNilDB                      = errors.New("loggingmanager core database is required")
	errNilWriterInput             = errors.New("loggingmanager core writer input channel is required")
	errInvalidMaxEvents           = errors.New("loggingmanager core max events must be greater than zero")
	errInvalidMaxOccurrences      = errors.New("loggingmanager core max occurrences must be greater than zero")
	errInvalidEventIDPrefixConfig = errors.New("loggingmanager core event id prefix is invalid")
	errNilClock                   = errors.New("loggingmanager core clock is required")
)

// Core encapsulates logging-manager orchestration logic.
//
// Ownership:
//   - Lifecycle ownership of writer goroutine and timers is external.
//   - Core owns initialization, runtime-state tracking, read/write query
//     composition, and writer request production.
type Core struct {
	db             *sql.DB
	writerInputCh  chan<- WriteRequest
	validator      InputValidator
	clock          func() time.Time
	maxEvents      int
	maxOccurrences int

	stateMu         sync.Mutex
	currentEventRec int64
	currentEventSeq uint32
	currentIDPrefix string
}

// CoreDeps defines required dependencies and runtime bounds.
type CoreDeps struct {
	DB             *sql.DB
	WriterInputCh  chan<- WriteRequest
	Validator      InputValidator
	Clock          func() time.Time
	MaxEvents      int
	MaxOccurrences int
	EventIDPrefix  string
}

// NewCore initializes schema/runtime state and returns a ready logging-manager
// orchestrator facade.
func NewCore(ctx context.Context, deps CoreDeps) (*Core, error) {
	normalizedDeps, err := normalizeCoreDependencies(deps)
	if err != nil {
		return nil, err
	}

	core := &Core{
		db:              normalizedDeps.DB,
		writerInputCh:   normalizedDeps.WriterInputCh,
		validator:       normalizedDeps.Validator,
		clock:           normalizedDeps.Clock,
		maxEvents:       normalizedDeps.MaxEvents,
		maxOccurrences:  normalizedDeps.MaxOccurrences,
		currentIDPrefix: normalizedDeps.EventIDPrefix,
	}

	if err := core.initialize(ctx); err != nil {
		return nil, err
	}

	return core, nil
}

func normalizeCoreDependencies(deps CoreDeps) (CoreDeps, error) {
	if err := verifyCoreObjects(deps); err != nil {
		return CoreDeps{}, err
	}
	if err := verifyCoreChannels(deps); err != nil {
		return CoreDeps{}, err
	}
	normalizedDeps, err := normalizeCoreProperties(deps)
	if err != nil {
		return CoreDeps{}, err
	}
	if err := ensureCoreValidator(&normalizedDeps); err != nil {
		return CoreDeps{}, err
	}
	return normalizedDeps, nil
}

func verifyCoreObjects(deps CoreDeps) error {
	if deps.DB == nil {
		return errNilDB
	}
	if deps.Clock == nil {
		return errNilClock
	}
	return nil
}

func verifyCoreChannels(deps CoreDeps) error {
	if deps.WriterInputCh == nil {
		return errNilWriterInput
	}
	return nil
}

func normalizeCoreProperties(deps CoreDeps) (CoreDeps, error) {
	if deps.MaxEvents <= 0 {
		return CoreDeps{}, errInvalidMaxEvents
	}
	if deps.MaxOccurrences <= 0 {
		return CoreDeps{}, errInvalidMaxOccurrences
	}

	normalizedDeps := deps
	if normalizedDeps.EventIDPrefix == "" {
		normalizedDeps.EventIDPrefix = defaultEventPrefix
	}
	if len(normalizedDeps.EventIDPrefix) > query.EventIDPrefixMaxLength {
		return CoreDeps{}, errInvalidEventIDPrefixConfig
	}

	return normalizedDeps, nil
}

func ensureCoreValidator(deps *CoreDeps) error {
	if deps.Validator != nil {
		return nil
	}
	validate, err := NewInputValidator()
	if err != nil {
		return err
	}
	deps.Validator = validate
	return nil
}

// Cleanup removes orphan rows for rotated-out events.
func (core *Core) Cleanup(ctx context.Context) error {
	return executeQueryBatch(ctx, core.db, []query.Query{
		query.DeleteOrphanOccurrences(),
		query.DeleteOrphanEventMeta(),
	})
}

// CreateEvent creates or rotates one event slot and enqueues all writes
// required to persist event/lifecycle/state rows. Runtime pointers are passed
// as deferred transaction-tail updates for single-write persistence per flush.
func (core *Core) CreateEvent(input dto.CreateEventInput) (string, error) {
	if err := core.validator.Struct(input); err != nil {
		return "", err
	}

	recID, seq, eventID, err := core.nextEventIdentity()
	if err != nil {
		return "", err
	}

	now := core.clock().UTC().Format(time.RFC3339)
	createdAt := resolveCreatedAt(input.CreatedAt, now)
	severity := resolveSeverity(input.Severity)
	priority := resolvePriority(input.Priority)
	source := resolveSource(input.Source)

	eventQuery := query.UpsertEvent(dto.EventRow{
		RecID:     recID,
		EventID:   eventID,
		Category:  input.Category,
		Severity:  severity,
		Priority:  priority,
		CreatedAt: createdAt,
		Source:    source,
	})

	core.enqueueCreateEventWrites(recID, seq, eventID, createdAt, eventQuery)

	return eventID, nil
}

// AddOccurrence enqueues one occurrence write.
func (core *Core) AddOccurrence(row dto.OccurrenceRow) error {
	core.writerInputCh <- WriteRequest{
		Query:              query.InsertOccurrence(row),
		TouchesOccurrences: true,
	}
	return nil
}

func (core *Core) initialize(ctx context.Context) error {
	if err := executeQueryBatch(ctx, core.db, query.BuildSchemaMigrationQueries()); err != nil {
		return err
	}

	if err := core.Cleanup(ctx); err != nil {
		return err
	}

	seedQueries := query.BuildRuntimeStateSeedQueries(0, 0, core.currentIDPrefix)
	if err := executeQueryBatch(ctx, core.db, seedQueries); err != nil {
		return err
	}

	return core.loadRuntimeState(ctx)
}

func (core *Core) loadRuntimeState(ctx context.Context) error {
	rows, err := core.db.QueryContext(
		ctx,
		"SELECT key, value FROM runtime_state WHERE key IN (?, ?, ?)",
		query.RuntimeStateCurrentEventRecIDKey,
		query.RuntimeStateCurrentEventSeqKey,
		query.RuntimeStateEventIDPrefixKey,
	)
	if err != nil {
		return fmt.Errorf("query runtime_state: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		if err = core.applyRuntimeStateRow(rows); err != nil {
			return err
		}
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("iterate runtime_state: %w", err)
	}

	return nil
}

func (core *Core) applyRuntimeStateRow(rows *sql.Rows) error {
	var key string
	var value string
	if scanErr := rows.Scan(&key, &value); scanErr != nil {
		return fmt.Errorf("scan runtime_state: %w", scanErr)
	}
	return core.applyRuntimeStateKeyValue(key, value)
}

func (core *Core) applyRuntimeStateKeyValue(key string, value string) error {
	switch key {
	case query.RuntimeStateCurrentEventRecIDKey:
		return core.applyCurrentEventRecID(value)
	case query.RuntimeStateCurrentEventSeqKey:
		return core.applyCurrentEventSeq(value)
	case query.RuntimeStateEventIDPrefixKey:
		core.applyEventIDPrefix(value)
		return nil
	default:
		return nil
	}
}

func (core *Core) applyCurrentEventRecID(value string) error {
	parsed, err := parseRuntimeStateInt64(value)
	if err != nil {
		return fmt.Errorf("parse current_event_rec_id: %w", err)
	}
	core.currentEventRec = parsed
	return nil
}

func (core *Core) applyCurrentEventSeq(value string) error {
	parsed, err := parseRuntimeStateUint32(value)
	if err != nil {
		return fmt.Errorf("parse current_event_seq: %w", err)
	}
	core.currentEventSeq = parsed
	return nil
}

func (core *Core) applyEventIDPrefix(value string) {
	if value != "" {
		core.currentIDPrefix = value
	}
}

func parseRuntimeStateInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseRuntimeStateUint32(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}

func (core *Core) nextEventIdentity() (int64, uint32, string, error) {
	core.stateMu.Lock()
	defer core.stateMu.Unlock()

	recID := core.currentEventRec + 1
	if recID > int64(core.maxEvents) {
		recID = 1
	}

	seq := core.currentEventSeq + 1
	if core.currentEventSeq >= query.EventIDMaxSequence {
		seq = 0
	}

	eventID, err := query.BuildEventID(core.currentIDPrefix, seq)
	if err != nil {
		return 0, 0, "", err
	}

	core.currentEventRec = recID
	core.currentEventSeq = seq
	return recID, seq, eventID, nil
}

func executeQueryBatch(ctx context.Context, db *sql.DB, queries []query.Query) error {
	for _, builtQuery := range queries {
		if _, err := db.ExecContext(ctx, builtQuery.SQL, builtQuery.Args...); err != nil {
			return fmt.Errorf("exec %q: %w", builtQuery.SQL, err)
		}
	}
	return nil
}

func resolveCreatedAt(createdAt string, fallback string) string {
	if createdAt == "" {
		return fallback
	}
	return createdAt
}

func resolveSeverity(severity enum.Severity) enum.Severity {
	if severity == "" {
		return enum.SeverityInfo
	}
	return severity
}

func resolvePriority(priority *int) int {
	if priority == nil {
		return defaultEventPriority
	}
	return *priority
}

func resolveSource(source string) string {
	if source == "" {
		return defaultEventSource
	}
	return source
}

func (core *Core) enqueueCreateEventWrites(
	recID int64,
	seq uint32,
	eventID string,
	createdAt string,
	eventQuery query.Query,
) {
	lifecycleQuery := query.UpsertLifecycle(dto.LifecycleRow{
		RecID:          recID,
		EventID:        eventID,
		EventRecID:     recID,
		Acknowledged:   false,
		AcknowledgedBy: "",
		AcknowledgedAt: createdAt,
		Muted:          false,
		MutedBy:        "",
		MutedAt:        createdAt,
	})

	stateQuery := query.UpsertEventState(dto.EventStateRow{
		RecID:      recID,
		EventID:    eventID,
		EventRecID: recID,
		Status:     enum.StatusActive,
		Message:    "",
	})

	runtimeStateUpdates := []dto.RuntimeStateRow{
		{Key: query.RuntimeStateCurrentEventRecIDKey, Value: strconv.FormatInt(recID, 10)},
		{Key: query.RuntimeStateCurrentEventSeqKey, Value: strconv.FormatUint(uint64(seq), 10)},
		{Key: query.RuntimeStateEventIDPrefixKey, Value: core.currentIDPrefix},
	}

	core.writerInputCh <- WriteRequest{Query: eventQuery, RuntimeStateUpdates: runtimeStateUpdates}
	core.writerInputCh <- WriteRequest{Query: lifecycleQuery}
	core.writerInputCh <- WriteRequest{Query: stateQuery}
}
