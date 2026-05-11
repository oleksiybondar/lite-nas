package config_test

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadEventStoreConfigParsedFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.EventStoreConfig]{
		{Name: "sqlite path", Got: func(cfg config.EventStoreConfig) any { return cfg.Storage.SQLitePath }, Want: "/var/lib/lite-nas/eventstore.db"},
		{Name: "max events", Got: func(cfg config.EventStoreConfig) any { return cfg.Storage.MaxEvents }, Want: 5000},
		{Name: "max occurrences", Got: func(cfg config.EventStoreConfig) any { return cfg.Storage.MaxOccurrences }, Want: 500000},
		{Name: "writer batch size", Got: func(cfg config.EventStoreConfig) any { return cfg.Writer.BatchSize }, Want: 100},
		{Name: "writer flush interval", Got: func(cfg config.EventStoreConfig) any { return cfg.Writer.FlushInterval }, Want: 100 * time.Millisecond},
		{Name: "cleanup batch size", Got: func(cfg config.EventStoreConfig) any { return cfg.Cleanup.BatchSize }, Want: 5000},
		{Name: "cleanup interval", Got: func(cfg config.EventStoreConfig) any { return cfg.Cleanup.Interval }, Want: 5 * time.Second},
	}

	testcasetest.RunFieldCases(t, func(t *testing.T) config.EventStoreConfig {
		t.Helper()
		return loadEventStoreConfigFixture(t, validEventStoreConfigINI())
	}, testCases)
}

func TestLoadEventStoreConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "missing sqlite path",
			ini: strings.ReplaceAll(
				validEventStoreConfigINI(),
				"sqlite_path=/var/lib/lite-nas/eventstore.db\n",
				"sqlite_path=   \n",
			),
			want: "sqlite_path is required",
		},
		{
			name: "invalid max events",
			ini:  strings.ReplaceAll(validEventStoreConfigINI(), "max_events=5000\n", "max_events=0\n"),
			want: "max_events must be greater than zero",
		},
		{
			name: "invalid max occurrences",
			ini:  strings.ReplaceAll(validEventStoreConfigINI(), "max_occurrences=500000\n", "max_occurrences=-1\n"),
			want: "max_occurrences must be greater than zero",
		},
		{
			name: "invalid writer batch size",
			ini:  strings.ReplaceAll(validEventStoreConfigINI(), "batch_size=100\n", "batch_size=0\n"),
			want: "eventstore_writer batch_size must be greater than zero",
		},
		{
			name: "invalid writer flush interval duration",
			ini: strings.ReplaceAll(
				validEventStoreConfigINI(),
				"flush_interval=100ms\n",
				"flush_interval=not-a-duration\n",
			),
			want: "eventstore_writer flush_interval has invalid duration",
		},
		{
			name: "invalid cleanup interval value",
			ini:  strings.ReplaceAll(validEventStoreConfigINI(), "interval=5s\n", "interval=0s\n"),
			want: "eventstore_cleanup interval must be greater than zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			configtest.RunINILoadRejectsCase(t, config.LoadEventStoreConfig, testCase.ini, testCase.want)
		})
	}
}

func validEventStoreConfigINI() string {
	return "[eventstore]\n" +
		"sqlite_path=/var/lib/lite-nas/eventstore.db\n" +
		"max_events=5000\n" +
		"max_occurrences=500000\n" +
		"[eventstore_writer]\n" +
		"batch_size=100\n" +
		"flush_interval=100ms\n" +
		"[eventstore_cleanup]\n" +
		"batch_size=5000\n" +
		"interval=5s\n"
}

func loadEventStoreConfigFixture(t *testing.T, iniContent string) config.EventStoreConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadEventStoreConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadEventStoreConfig() error = %v", err)
	}

	return cfg
}
