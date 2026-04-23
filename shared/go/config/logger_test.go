package config_test

import (
	"strings"
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
)

type loggingFieldAssertion struct {
	name string
	got  func(config.LoggingConfig) any
}

var loggingFieldAssertions = []loggingFieldAssertion{
	{name: "level", got: func(cfg config.LoggingConfig) any { return cfg.Level }},
	{name: "format", got: func(cfg config.LoggingConfig) any { return cfg.Format }},
	{name: "output", got: func(cfg config.LoggingConfig) any { return cfg.Output }},
	{name: "file path", got: func(cfg config.LoggingConfig) any { return cfg.FilePath }},
}

func TestLoadLoggingConfigSuccessCases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		ini        string
		assertions map[string]any
	}{
		{
			name: "defaults",
			ini:  "",
			assertions: map[string]any{
				"level":     "info",
				"format":    "rfc5424",
				"output":    "stdout",
				"file path": "",
			},
		},
		{
			name: "file output",
			ini:  "[logging]\nlevel=debug\nformat=rfc5424\noutput=file\nfile_path=/tmp/lite-nas.log\n",
			assertions: map[string]any{
				"level":     "debug",
				"format":    "rfc5424",
				"output":    "file",
				"file path": "/tmp/lite-nas.log",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadLoggingConfigFixture(t, testCase.ini)
			assertLoggingConfigFields(t, cfg, testCase.assertions)
		})
	}
}

func TestLoadLoggingConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "invalid level",
			ini:  "[logging]\nlevel=trace\n",
			want: "unsupported logging level",
		},
		{
			name: "invalid format",
			ini:  "[logging]\nformat=json\n",
			want: "unsupported logging format",
		},
		{
			name: "invalid output",
			ini:  "[logging]\noutput=socket\n",
			want: "unsupported logging output",
		},
		{
			name: "missing file path",
			ini:  "[logging]\noutput=file\n",
			want: "file_path is required",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assertLoadLoggingConfigError(t, testCase.ini, testCase.want)
		})
	}
}

func loadLoggingConfigFixture(t *testing.T, iniContent string) config.LoggingConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadLoggingConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadLoggingConfig() error = %v", err)
	}

	return cfg
}

func assertLoggingConfigFields(t *testing.T, cfg config.LoggingConfig, want map[string]any) {
	t.Helper()

	for _, assertion := range loggingFieldAssertions {
		wantValue, ok := want[assertion.name]
		if !ok {
			continue
		}

		if got := assertion.got(cfg); got != wantValue {
			t.Fatalf("%s = %#v, want %#v", assertion.name, got, wantValue)
		}
	}
}

func assertLoadLoggingConfigError(t *testing.T, iniContent string, want string) {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	_, err = config.LoadLoggingConfig(cfgFile)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", err, want)
	}
}
