package config_test

import (
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/configtest"
)

func TestLoadSMTPConfigParsedFields(t *testing.T) {
	t.Parallel()

	cfg := mustLoadSMTPConfig(
		t,
		"[smtp]\n"+
			"host=127.0.0.1\n"+
			"port=25\n"+
			"timeout=10s\n"+
			"helo=localhost\n",
	)

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("cfg.Host = %q, want %q", cfg.Host, "127.0.0.1")
	}

	if cfg.Port != 25 {
		t.Fatalf("cfg.Port = %d, want %d", cfg.Port, 25)
	}

	if cfg.Timeout != 10*time.Second {
		t.Fatalf("cfg.Timeout = %s, want %s", cfg.Timeout, 10*time.Second)
	}

	if cfg.HELO != "localhost" {
		t.Fatalf("cfg.HELO = %q, want %q", cfg.HELO, "localhost")
	}
}

func TestLoadSMTPConfigDefaultsHELO(t *testing.T) {
	t.Parallel()

	cfg := mustLoadSMTPConfig(t, "[smtp]\nhost=127.0.0.1\nport=25\n")

	if cfg.HELO != "localhost" {
		t.Fatalf("cfg.HELO = %q, want %q", cfg.HELO, "localhost")
	}
}

func TestLoadSMTPConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		ini  string
		want string
	}{
		{
			name: "missing host",
			ini:  "[smtp]\nhost=   \nport=25\n",
			want: "smtp host is required",
		},
		{
			name: "invalid port",
			ini:  "[smtp]\nhost=127.0.0.1\nport=70000\n",
			want: "smtp port must be between 1 and 65535",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			configtest.RunINILoadRejectsCase(t, config.LoadSMTPConfig, testCase.ini, testCase.want)
		})
	}
}

func mustLoadSMTPConfig(t *testing.T, raw string) config.SMTPConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(raw))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadSMTPConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadSMTPConfig() error = %v", err)
	}

	return cfg
}
