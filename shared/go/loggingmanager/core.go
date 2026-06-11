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

// normalizeCoreDependencies validates and normalizes constructor dependencies
// so NewCore can consume one coherent dependency shape.
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

// verifyCoreObjects validates required object references.
func verifyCoreObjects(deps CoreDeps) error {
	if deps.DB == nil {
		return errNilDB
	}
	if deps.Clock == nil {
		return errNilClock
	}
	return nil
}

// verifyCoreChannels validates required channel dependencies.
func verifyCoreChannels(deps CoreDeps) error {
	if deps.WriterInputCh == nil {
		return errNilWriterInput
	}
	return nil
}

// normalizeCoreProperties validates and normalizes scalar constructor options.
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

// ensureCoreValidator provides the default validator when none was injected.
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

// Cleanup enforces event and occurrence retention limits and removes rows that
// no longer belong to retained events.
func (core *Core) Cleanup(ctx context.Context) error {
	queries := make([]query.Query, 0, 6)
	if core.maxEvents > 0 {
		queries = append(queries, query.DeleteOldestEventsBeyondLimit(core.maxEvents))
	}
	if core.maxOccurrences > 0 {
		queries = append(queries, query.DeleteOccurrencesPerEventBeyondLimit(core.maxOccurrences))
	}
	queries = append(queries,
		query.DeleteOrphanLifecycle(),
		query.DeleteOrphanEventState(),
		query.DeleteOrphanOccurrences(),
		query.DeleteOrphanEventMeta(),
	)
	return executeQueryBatch(ctx, core.db, queries)
}

// CreateEvent creates one new retained event identity and enqueues all writes
// required to persist event/lifecycle/state rows. Runtime counters are passed
// as deferred transaction-tail updates for single-write persistence per flush.
func (core *Core) CreateEvent(input dto.CreateEventInput) (string, error) {
	if err := core.validator.Struct(input); err != nil {
		return "", err
	}

	recID, seq, generatedEventID, err := core.nextEventIdentity()
	if err != nil {
		return "", err
	}
	eventID := generatedEventID
	if input.EventID != "" {
		eventID = input.EventID
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

// initialize executes startup orchestration for schema and runtime-state.
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

// loadRuntimeState reads persisted runtime pointers and applies them to the
// in-memory tracker state.
func (core *Core) loadRuntimeState(ctx context.Context) error {
	builtQuery := query.SelectRuntimeStatePointers()
	rows, err := core.db.QueryContext(ctx, builtQuery.SQL, builtQuery.Args...)
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

// applyRuntimeStateRow scans one key/value row and applies it to runtime state.
func (core *Core) applyRuntimeStateRow(rows *sql.Rows) error {
	var key string
	var value string
	if scanErr := rows.Scan(&key, &value); scanErr != nil {
		return fmt.Errorf("scan runtime_state: %w", scanErr)
	}
	return core.applyRuntimeStateKeyValue(key, value)
}

// applyRuntimeStateKeyValue routes one runtime-state entry to its typed applier.
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

// applyCurrentEventRecID parses and stores the current rotation rec_id pointer.
func (core *Core) applyCurrentEventRecID(value string) error {
	parsed, err := parseRuntimeStateInt64(value)
	if err != nil {
		return fmt.Errorf("parse current_event_rec_id: %w", err)
	}
	core.currentEventRec = parsed
	return nil
}

// applyCurrentEventSeq parses and stores the current generated sequence.
func (core *Core) applyCurrentEventSeq(value string) error {
	parsed, err := parseRuntimeStateUint32(value)
	if err != nil {
		return fmt.Errorf("parse current_event_seq: %w", err)
	}
	core.currentEventSeq = parsed
	return nil
}

// applyEventIDPrefix updates the active event-id prefix when value is non-empty.
func (core *Core) applyEventIDPrefix(value string) {
	if value != "" {
		core.currentIDPrefix = value
	}
}

// parseRuntimeStateInt64 parses a runtime-state value into int64.
func parseRuntimeStateInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

// parseRuntimeStateUint32 parses a runtime-state value into uint32.
func parseRuntimeStateUint32(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}

// nextEventIdentity advances rec_id and sequence counters and returns the
// resulting business event identifier.
func (core *Core) nextEventIdentity() (int64, uint32, string, error) {
	core.stateMu.Lock()
	defer core.stateMu.Unlock()

	recID := core.currentEventRec + 1

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

// executeQueryBatch executes query items in-order as one startup batch step.
func executeQueryBatch(ctx context.Context, db *sql.DB, queries []query.Query) error {
	for _, builtQuery := range queries {
		if _, err := db.ExecContext(ctx, builtQuery.SQL, builtQuery.Args...); err != nil {
			return fmt.Errorf("exec %q: %w", builtQuery.SQL, err)
		}
	}
	return nil
}

// resolveCreatedAt returns input value or the supplied fallback timestamp.
func resolveCreatedAt(createdAt string, fallback string) string {
	if createdAt == "" {
		return fallback
	}
	return createdAt
}

// resolveSeverity returns input value or default informational severity.
func resolveSeverity(severity enum.Severity) enum.Severity {
	if severity == "" {
		return enum.SeverityInfo
	}
	return severity
}

// resolvePriority returns input value or the default event priority.
func resolvePriority(priority *int) int {
	if priority == nil {
		return defaultEventPriority
	}
	return *priority
}

// resolveSource returns input value or the default source label.
func resolveSource(source string) string {
	if source == "" {
		return defaultEventSource
	}
	return source
}

// enqueueCreateEventWrites emits the create-event write set into the writer
// input channel, including deferred runtime-state tail updates.
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
