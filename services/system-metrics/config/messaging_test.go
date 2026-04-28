package config

import (
	"errors"
	"testing"
	"time"

	"lite-nas/shared/testutil/fileiotest"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read failed")

	if _, err := LoadConfig(fileiotest.Reader{Err: expectedErr}); !errors.Is(err, expectedErr) {
		t.Fatalf("LoadConfig() error = %v, want %v", err, expectedErr)
	}
}

func TestLoadConfigMetricsFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "poll interval", Got: func(cfg Config) any { return cfg.Metrics.PollInterval }, Want: 2 * time.Second},
		{Name: "history size", Got: func(cfg Config) any { return cfg.Metrics.HistorySize }, Want: 10},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "url", Got: func(cfg Config) any { return cfg.Messaging.URL }, Want: "nats://localhost:4222"},
		{Name: "client name", Got: func(cfg Config) any { return cfg.Messaging.ClientName }, Want: "system-metrics"},
		{Name: "ca path", Got: func(cfg Config) any { return cfg.Messaging.CA }, Want: "/etc/lite-nas/certificates/root-ca.crt"},
		{Name: "cert path", Got: func(cfg Config) any { return cfg.Messaging.Cert }, Want: "/etc/lite-nas/certificates/lite-nas-system-metrics/client.crt"},
		{Name: "key path", Got: func(cfg Config) any { return cfg.Messaging.Key }, Want: "/etc/lite-nas/certificates/lite-nas-system-metrics/client.key"},
		{Name: "timeout", Got: func(cfg Config) any { return cfg.Messaging.Timeout }, Want: 9 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "level", Got: func(cfg Config) any { return cfg.Logging.Level }, Want: "debug"},
		{Name: "output", Got: func(cfg Config) any { return cfg.Logging.Output }, Want: "file"},
		{Name: "file path", Got: func(cfg Config) any { return cfg.Logging.FilePath }, Want: "/var/lib/lite-nas/system-metrics.log"},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func TestLoadConfigRejectsInvalidMetricsValues(t *testing.T) {
	t.Parallel()

	reader := fileiotest.Reader{
		Data: []byte("[metrics]\npoll_interval=nope\nhistory_size=10\n"),
	}

	if _, err := LoadConfig(reader); err == nil {
		t.Fatal("expected invalid metrics error")
	}
}

func TestLoadConfigRejectsInvalidLoggingValues(t *testing.T) {
	t.Parallel()

	reader := fileiotest.Reader{
		Data: []byte(
			"[metrics]\n" +
				"poll_interval=1s\n" +
				"history_size=10\n" +
				"[logging]\n" +
				"output=file\n",
		),
	}

	if _, err := LoadConfig(reader); err == nil {
		t.Fatal("expected invalid logging error")
	}
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{
		Data: []byte(
			"[metrics]\n" +
				"poll_interval=2s\n" +
				"history_size=10\n" +
				"[messaging]\n" +
				"url=nats://localhost:4222\n" +
				"client_name=system-metrics\n" +
				"ca=/etc/lite-nas/certificates/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/lite-nas-system-metrics/client.crt\n" +
				"key=/etc/lite-nas/certificates/lite-nas-system-metrics/client.key\n" +
				"timeout=9s\n" +
				"[logging]\n" +
				"level=debug\n" +
				"format=rfc5424\n" +
				"output=file\n" +
				"file_path=/var/lib/lite-nas/system-metrics.log\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}
