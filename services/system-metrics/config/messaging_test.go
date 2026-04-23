package config

import (
	"errors"
	"testing"
	"time"
)

type fakeReader struct {
	data []byte
	err  error
}

func (r fakeReader) Read() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.data, nil
}

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read failed")

	if _, err := LoadConfig(fakeReader{err: expectedErr}); !errors.Is(err, expectedErr) {
		t.Fatalf("LoadConfig() error = %v, want %v", err, expectedErr)
	}
}

func TestLoadConfigMetricsFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Config) any
		want any
	}{
		{name: "poll interval", got: func(cfg Config) any { return cfg.Metrics.PollInterval }, want: 2 * time.Second},
		{name: "history size", got: func(cfg Config) any { return cfg.Metrics.HistorySize }, want: 10},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadConfigFixture(t)
			if got := testCase.got(cfg); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Config) any
		want any
	}{
		{name: "url", got: func(cfg Config) any { return cfg.Messaging.URL }, want: "nats://localhost:4222"},
		{name: "client name", got: func(cfg Config) any { return cfg.Messaging.ClientName }, want: "system-metrics"},
		{name: "timeout", got: func(cfg Config) any { return cfg.Messaging.Timeout }, want: 9 * time.Second},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadConfigFixture(t)
			if got := testCase.got(cfg); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Config) any
		want any
	}{
		{name: "level", got: func(cfg Config) any { return cfg.Logging.Level }, want: "debug"},
		{name: "output", got: func(cfg Config) any { return cfg.Logging.Output }, want: "file"},
		{name: "file path", got: func(cfg Config) any { return cfg.Logging.FilePath }, want: "/var/log/liteNAS/system-metrics.log"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadConfigFixture(t)
			if got := testCase.got(cfg); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestLoadConfigRejectsInvalidMetricsValues(t *testing.T) {
	t.Parallel()

	reader := fakeReader{
		data: []byte("[metrics]\npoll_interval=nope\nhistory_size=10\n"),
	}

	if _, err := LoadConfig(reader); err == nil {
		t.Fatal("expected invalid metrics error")
	}
}

func TestLoadConfigRejectsInvalidLoggingValues(t *testing.T) {
	t.Parallel()

	reader := fakeReader{
		data: []byte(
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

	cfg, err := LoadConfig(fakeReader{
		data: []byte(
			"[metrics]\n" +
				"poll_interval=2s\n" +
				"history_size=10\n" +
				"[messaging]\n" +
				"url=nats://localhost:4222\n" +
				"client_name=system-metrics\n" +
				"timeout=9s\n" +
				"[logging]\n" +
				"level=debug\n" +
				"format=rfc5424\n" +
				"output=file\n" +
				"file_path=/var/log/liteNAS/system-metrics.log\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}
