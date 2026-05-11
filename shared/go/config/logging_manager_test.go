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

	configtest.RunReaderErrorCase(t, config.LoadLoggingManagerServiceConfig)
}

func TestLoadLoggingManagerConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.LoggingManagerServiceConfig]{
		{Name: "url", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.Messaging.URL }, Want: "nats://localhost:4222"},
		{Name: "client name", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.Messaging.ClientName }, Want: "security-logging-manager"},
		{Name: "ca path", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.Messaging.CA }, Want: "/tmp/ca.pem"},
		{Name: "cert path", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.Messaging.Cert }, Want: "/tmp/cert.pem"},
		{Name: "key path", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.Messaging.Key }, Want: "/tmp/key.pem"},
		{Name: "timeout", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.Messaging.Timeout }, Want: 7 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadLoggingManagerConfigFixture, testCases)
}

func TestLoadLoggingManagerConfigLoggingManagerFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.LoggingManagerServiceConfig]{
		{Name: "sqlite path", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Storage.SQLitePath }, Want: "/var/lib/lite-nas/security-events.db"},
		{Name: "max events", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Storage.MaxEvents }, Want: 10000},
		{Name: "max occurrences", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Storage.MaxOccurrences }, Want: 750000},
		{Name: "writer batch size", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Writer.BatchSize }, Want: 100},
		{Name: "writer flush interval", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Writer.FlushInterval }, Want: 100 * time.Millisecond},
		{Name: "cleanup batch size", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Cleanup.BatchSize }, Want: 5000},
		{Name: "cleanup interval", Got: func(cfg config.LoggingManagerServiceConfig) any { return cfg.LoggingManager.Cleanup.Interval }, Want: 2 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadLoggingManagerConfigFixture, testCases)
}

func TestLoadLoggingManagerConfigRejectsInvalidLoggingManagerValues(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		config.LoadLoggingManagerServiceConfig,
		"[messaging]\n"+
			"url=nats://localhost:4222\n"+
			"client_name=security-logging-manager\n"+
			"ca=/tmp/ca.pem\n"+
			"cert=/tmp/cert.pem\n"+
			"key=/tmp/key.pem\n"+
			"timeout=7s\n"+
			"[loggingmanager]\n"+
			"sqlite_path=\n"+
			"max_events=10000\n"+
			"max_occurrences=750000\n"+
			"[loggingmanager_writer]\n"+
			"batch_size=100\n"+
			"flush_interval=100ms\n"+
			"[loggingmanager_cleanup]\n"+
			"batch_size=5000\n"+
			"interval=2s\n",
	)
}

func loadLoggingManagerConfigFixture(t *testing.T) config.LoggingManagerServiceConfig {
	t.Helper()

	cfg, err := config.LoadLoggingManagerServiceConfig(fileiotest.Reader{
		Data: []byte(
			"[messaging]\n" +
				"url=nats://localhost:4222\n" +
				"client_name=security-logging-manager\n" +
				"ca=/tmp/ca.pem\n" +
				"cert=/tmp/cert.pem\n" +
				"key=/tmp/key.pem\n" +
				"timeout=7s\n" +
				"[loggingmanager]\n" +
				"sqlite_path=/var/lib/lite-nas/security-events.db\n" +
				"max_events=10000\n" +
				"max_occurrences=750000\n" +
				"[loggingmanager_writer]\n" +
				"batch_size=100\n" +
				"flush_interval=100ms\n" +
				"[loggingmanager_cleanup]\n" +
				"batch_size=5000\n" +
				"interval=2s\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadLoggingManagerServiceConfig() error = %v", err)
	}

	return cfg
}
