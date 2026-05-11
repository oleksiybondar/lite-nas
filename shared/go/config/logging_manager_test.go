package config_test

import (
	"testing"
	"time"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadLoggingManagerConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, config.LoadLoggingManagerConfig)
}

func TestLoadLoggingManagerConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.LoggingManagerConfig]{
		{Name: "url", Got: func(cfg config.LoggingManagerConfig) any { return cfg.Messaging.URL }, Want: "nats://localhost:4222"},
		{Name: "client name", Got: func(cfg config.LoggingManagerConfig) any { return cfg.Messaging.ClientName }, Want: "security-logging-manager"},
		{Name: "ca path", Got: func(cfg config.LoggingManagerConfig) any { return cfg.Messaging.CA }, Want: "/tmp/ca.pem"},
		{Name: "cert path", Got: func(cfg config.LoggingManagerConfig) any { return cfg.Messaging.Cert }, Want: "/tmp/cert.pem"},
		{Name: "key path", Got: func(cfg config.LoggingManagerConfig) any { return cfg.Messaging.Key }, Want: "/tmp/key.pem"},
		{Name: "timeout", Got: func(cfg config.LoggingManagerConfig) any { return cfg.Messaging.Timeout }, Want: 7 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadLoggingManagerConfigFixture, testCases)
}

func TestLoadLoggingManagerConfigEventStoreFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.LoggingManagerConfig]{
		{Name: "sqlite path", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Storage.SQLitePath }, Want: "/var/lib/lite-nas/security-events.db"},
		{Name: "max events", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Storage.MaxEvents }, Want: 10000},
		{Name: "max occurrences", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Storage.MaxOccurrences }, Want: 750000},
		{Name: "writer batch size", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Writer.BatchSize }, Want: 100},
		{Name: "writer flush interval", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Writer.FlushInterval }, Want: 100 * time.Millisecond},
		{Name: "cleanup batch size", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Cleanup.BatchSize }, Want: 5000},
		{Name: "cleanup interval", Got: func(cfg config.LoggingManagerConfig) any { return cfg.EventStore.Cleanup.Interval }, Want: 2 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadLoggingManagerConfigFixture, testCases)
}

func TestLoadLoggingManagerConfigRejectsInvalidEventStoreValues(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		config.LoadLoggingManagerConfig,
		"[messaging]\n"+
			"url=nats://localhost:4222\n"+
			"client_name=security-logging-manager\n"+
			"ca=/tmp/ca.pem\n"+
			"cert=/tmp/cert.pem\n"+
			"key=/tmp/key.pem\n"+
			"timeout=7s\n"+
			"[eventstore]\n"+
			"sqlite_path=\n"+
			"max_events=10000\n"+
			"max_occurrences=750000\n"+
			"[eventstore_writer]\n"+
			"batch_size=100\n"+
			"flush_interval=100ms\n"+
			"[eventstore_cleanup]\n"+
			"batch_size=5000\n"+
			"interval=2s\n",
	)
}

func loadLoggingManagerConfigFixture(t *testing.T) config.LoggingManagerConfig {
	t.Helper()

	cfg, err := config.LoadLoggingManagerConfig(fileiotest.Reader{
		Data: []byte(
			"[messaging]\n" +
				"url=nats://localhost:4222\n" +
				"client_name=security-logging-manager\n" +
				"ca=/tmp/ca.pem\n" +
				"cert=/tmp/cert.pem\n" +
				"key=/tmp/key.pem\n" +
				"timeout=7s\n" +
				"[eventstore]\n" +
				"sqlite_path=/var/lib/lite-nas/security-events.db\n" +
				"max_events=10000\n" +
				"max_occurrences=750000\n" +
				"[eventstore_writer]\n" +
				"batch_size=100\n" +
				"flush_interval=100ms\n" +
				"[eventstore_cleanup]\n" +
				"batch_size=5000\n" +
				"interval=2s\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadLoggingManagerConfig() error = %v", err)
	}

	return cfg
}
