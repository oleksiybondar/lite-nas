package modules

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	loggingmanagerconfig "lite-nas/shared/config/loggingmanager"
)

func TestNewLoggingManagerCoreModuleBuildsDependencies(t *testing.T) {
	t.Parallel()

	cfg := newLoggingManagerCoreConfigFixture(t)

	module, err := NewLoggingManagerCoreModule(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewLoggingManagerCoreModule() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := module.Close(); closeErr != nil {
			t.Fatalf("Close() error = %v", closeErr)
		}
	})

	assertLoggingManagerCoreModuleInitialized(t, module)
}

func newLoggingManagerCoreConfigFixture(t *testing.T) loggingmanagerconfig.LoggingManagerConfig {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "logging-manager.db")
	return loggingmanagerconfig.LoggingManagerConfig{
		Storage: loggingmanagerconfig.LoggingManagerStorageConfig{
			SQLitePath:     dbPath,
			MaxEvents:      100,
			MaxOccurrences: 1000,
			EventIDPrefix:  "event",
		},
		Writer: loggingmanagerconfig.LoggingManagerWriterConfig{
			BatchSize:     10,
			FlushInterval: 100 * time.Millisecond,
		},
		Cleanup: loggingmanagerconfig.LoggingManagerCleanupConfig{
			BatchSize: 100,
			Interval:  time.Second,
		},
	}
}

func assertLoggingManagerCoreModuleInitialized(t *testing.T, module LoggingManagerCore) {
	t.Helper()
	assertNotNil(t, module.DB, "expected DB to be initialized")
	assertNotNil(t, module.Executor, "expected Executor to be initialized")
	assertNotNil(t, module.Writer, "expected Writer to be initialized")
	assertNotNil(t, module.Core, "expected Core to be initialized")
	assertNotNil(t, module.WriterInputCh, "expected WriterInputCh to be initialized")
	assertNotNil(t, module.WriterFlushCh, "expected WriterFlushCh to be initialized")
	assertNotNil(t, module.CleanupTicksCh, "expected CleanupTicksCh to be initialized")
}

func assertNotNil(t *testing.T, value any, message string) {
	t.Helper()
	if value == nil {
		t.Fatal(message)
	}
}
