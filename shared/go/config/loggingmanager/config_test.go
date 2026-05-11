package loggingmanager_test

import (
	"strings"
	"testing"
	"time"

	"gopkg.in/ini.v1"

	loggingmanagerconfig "lite-nas/shared/config/loggingmanager"
	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadLoggingManagerConfigParsedFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[loggingmanagerconfig.LoggingManagerConfig]{
		{Name: "sqlite path", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Storage.SQLitePath }, Want: "/var/lib/lite-nas/loggingmanager.db"},
		{Name: "max events", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Storage.MaxEvents }, Want: 5000},
		{Name: "max occurrences", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Storage.MaxOccurrences }, Want: 500000},
		{Name: "writer batch size", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Writer.BatchSize }, Want: 100},
		{Name: "writer flush interval", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Writer.FlushInterval }, Want: 100 * time.Millisecond},
		{Name: "cleanup batch size", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Cleanup.BatchSize }, Want: 5000},
		{Name: "cleanup interval", Got: func(cfg loggingmanagerconfig.LoggingManagerConfig) any { return cfg.Cleanup.Interval }, Want: 5 * time.Second},
	}

	testcasetest.RunFieldCases(t, func(t *testing.T) loggingmanagerconfig.LoggingManagerConfig {
		t.Helper()
		return loadLoggingManagerSectionConfigFixture(t, validLoggingManagerConfigINI())
	}, testCases)
}

func TestLoadLoggingManagerConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	for _, testCase := range invalidLoggingManagerConfigCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			configtest.RunINILoadRejectsCase(t, loggingmanagerconfig.LoadLoggingManagerConfig, testCase.ini, testCase.want)
		})
	}
}

func invalidLoggingManagerConfigCases() []struct {
	name string
	ini  string
	want string
} {
	cases := []struct {
		name string
		ini  string
		want string
	}{}

	cases = append(cases, invalidLoggingManagerStorageCases()...)
	cases = append(cases, invalidLoggingManagerWriterCases()...)
	cases = append(cases, invalidLoggingManagerCleanupCases()...)

	return cases
}

func invalidLoggingManagerStorageCases() []struct {
	name string
	ini  string
	want string
} {
	return []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "missing sqlite path",
			ini: strings.ReplaceAll(
				validLoggingManagerConfigINI(),
				"sqlite_path=/var/lib/lite-nas/loggingmanager.db\n",
				"sqlite_path=   \n",
			),
			want: "sqlite_path is required",
		},
		{
			name: "invalid max events",
			ini:  strings.ReplaceAll(validLoggingManagerConfigINI(), "max_events=5000\n", "max_events=0\n"),
			want: "max_events must be greater than zero",
		},
		{
			name: "invalid max occurrences",
			ini:  strings.ReplaceAll(validLoggingManagerConfigINI(), "max_occurrences=500000\n", "max_occurrences=-1\n"),
			want: "max_occurrences must be greater than zero",
		},
	}
}

func invalidLoggingManagerWriterCases() []struct {
	name string
	ini  string
	want string
} {
	return []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "invalid writer batch size",
			ini:  strings.ReplaceAll(validLoggingManagerConfigINI(), "batch_size=100\n", "batch_size=0\n"),
			want: "loggingmanager_writer batch_size must be greater than zero",
		},
		{
			name: "invalid writer flush interval duration",
			ini: strings.ReplaceAll(
				validLoggingManagerConfigINI(),
				"flush_interval=100ms\n",
				"flush_interval=not-a-duration\n",
			),
			want: "loggingmanager_writer flush_interval has invalid duration",
		},
	}
}

func invalidLoggingManagerCleanupCases() []struct {
	name string
	ini  string
	want string
} {
	return []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "invalid cleanup interval value",
			ini:  strings.ReplaceAll(validLoggingManagerConfigINI(), "interval=5s\n", "interval=0s\n"),
			want: "loggingmanager_cleanup interval must be greater than zero",
		},
	}
}

func validLoggingManagerConfigINI() string {
	return "[loggingmanager]\n" +
		"sqlite_path=/var/lib/lite-nas/loggingmanager.db\n" +
		"max_events=5000\n" +
		"max_occurrences=500000\n" +
		"[loggingmanager_writer]\n" +
		"batch_size=100\n" +
		"flush_interval=100ms\n" +
		"[loggingmanager_cleanup]\n" +
		"batch_size=5000\n" +
		"interval=5s\n"
}

func loadLoggingManagerSectionConfigFixture(t *testing.T, iniContent string) loggingmanagerconfig.LoggingManagerConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := loggingmanagerconfig.LoadLoggingManagerConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadLoggingManagerConfig() error = %v", err)
	}

	return cfg
}
