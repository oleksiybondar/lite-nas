package modules

import (
	"context"
	"database/sql"
	"time"

	loggingmanagerconfig "lite-nas/shared/config/loggingmanager"
	"lite-nas/shared/loggingmanager"
	sharedworkers "lite-nas/shared/workers"

	_ "modernc.org/sqlite"
)

// LoggingManagerCore groups initialized logging-manager dependencies that are
// shared by security/system logging-manager services.
//
// Scope:
//   - Includes local persistence, writer path, core facade, and cleanup timer
//     worker wiring.
//   - Excludes configuration loading and NATS integration, which stay in
//     service-level composition.
type LoggingManagerCore struct {
	DB             *sql.DB
	Executor       *loggingmanager.SQLTransactionExecutor
	Writer         *loggingmanager.Writer
	Core           *loggingmanager.Core
	WriterInputCh  chan loggingmanager.WriteRequest
	WriterFlushCh  chan struct{}
	CleanupTicksCh chan struct{}
	CleanupTimer   sharedworkers.TimerWorker
}

// NewLoggingManagerCoreModule builds the local logging-manager runtime module
// from already-loaded logging-manager configuration.
func NewLoggingManagerCoreModule(
	ctx context.Context,
	cfg loggingmanagerconfig.LoggingManagerConfig,
) (LoggingManagerCore, error) {
	db, err := newLoggingManagerDB(ctx, cfg.Storage.SQLitePath)
	if err != nil {
		return LoggingManagerCore{}, err
	}

	module, err := buildLoggingManagerCoreModule(ctx, db, cfg)
	if err != nil {
		_ = db.Close()
		return LoggingManagerCore{}, err
	}
	return module, nil
}

func buildLoggingManagerCoreModule(
	ctx context.Context,
	db *sql.DB,
	cfg loggingmanagerconfig.LoggingManagerConfig,
) (LoggingManagerCore, error) {
	executor, err := loggingmanager.NewSQLTransactionExecutor(db)
	if err != nil {
		return LoggingManagerCore{}, err
	}

	writerInputCh := make(chan loggingmanager.WriteRequest, cfg.Writer.BatchSize*100)
	writerFlushCh := make(chan struct{}, 1)
	writer, err := newLoggingManagerWriter(executor, writerInputCh, writerFlushCh, cfg)
	if err != nil {
		return LoggingManagerCore{}, err
	}

	core, err := newLoggingManagerCore(ctx, db, writerInputCh, cfg)
	if err != nil {
		return LoggingManagerCore{}, err
	}

	cleanupTicksCh, cleanupTimer, err := newLoggingManagerCleanupTimer(cfg)
	if err != nil {
		return LoggingManagerCore{}, err
	}

	return LoggingManagerCore{
		DB:             db,
		Executor:       executor,
		Writer:         writer,
		Core:           core,
		WriterInputCh:  writerInputCh,
		WriterFlushCh:  writerFlushCh,
		CleanupTicksCh: cleanupTicksCh,
		CleanupTimer:   cleanupTimer,
	}, nil
}

// Close releases persistence resources owned by the module.
func (module LoggingManagerCore) Close() error {
	if module.DB == nil {
		return nil
	}
	return module.DB.Close()
}

// applyLoggingManagerSQLitePragmas applies SQLite runtime settings used by
// logging-manager write/read workloads.
func applyLoggingManagerSQLitePragmas(ctx context.Context, db *sql.DB) error {
	pragmaStatements := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA temp_store=MEMORY;",
		"PRAGMA busy_timeout=5000;",
	}

	for _, statement := range pragmaStatements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func newLoggingManagerDB(ctx context.Context, sqlitePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		return nil, err
	}
	if err = applyLoggingManagerSQLitePragmas(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func newLoggingManagerWriter(
	executor *loggingmanager.SQLTransactionExecutor,
	writerInputCh chan loggingmanager.WriteRequest,
	writerFlushCh chan struct{},
	cfg loggingmanagerconfig.LoggingManagerConfig,
) (*loggingmanager.Writer, error) {
	return loggingmanager.NewWriter(
		executor,
		loggingmanager.DefaultTransactionBuilder{},
		writerInputCh,
		writerFlushCh,
		cfg.Writer.BatchSize,
		cfg.Storage.MaxOccurrences,
	)
}

func newLoggingManagerCore(
	ctx context.Context,
	db *sql.DB,
	writerInputCh chan loggingmanager.WriteRequest,
	cfg loggingmanagerconfig.LoggingManagerConfig,
) (*loggingmanager.Core, error) {
	return loggingmanager.NewCore(ctx, loggingmanager.CoreDeps{
		DB:             db,
		WriterInputCh:  writerInputCh,
		Clock:          time.Now,
		MaxEvents:      cfg.Storage.MaxEvents,
		MaxOccurrences: cfg.Storage.MaxOccurrences,
		EventIDPrefix:  cfg.Storage.EventIDPrefix,
	})
}

func newLoggingManagerCleanupTimer(
	cfg loggingmanagerconfig.LoggingManagerConfig,
) (chan struct{}, sharedworkers.TimerWorker, error) {
	cleanupTicksCh := make(chan struct{}, 1)
	cleanupTimer, err := sharedworkers.NewTimerWorker(
		sharedworkers.TimerConfig{
			Interval:    cfg.Cleanup.Interval,
			EmitOnStart: false,
		},
		cleanupTicksCh,
	)
	if err != nil {
		return nil, sharedworkers.TimerWorker{}, err
	}
	return cleanupTicksCh, cleanupTimer, nil
}
