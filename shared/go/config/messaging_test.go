package config_test

import (
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
)

func TestLoadMessagingConfigParsedFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(config.MessagingConfig) any
		want any
	}{
		{name: "url", got: func(cfg config.MessagingConfig) any { return cfg.URL }, want: "nats://localhost:4222"},
		{name: "client name", got: func(cfg config.MessagingConfig) any { return cfg.ClientName }, want: "system-metrics"},
		{name: "ca path", got: func(cfg config.MessagingConfig) any { return cfg.CA }, want: "/tmp/ca.pem"},
		{name: "cert path", got: func(cfg config.MessagingConfig) any { return cfg.Cert }, want: "/tmp/cert.pem"},
		{name: "key path", got: func(cfg config.MessagingConfig) any { return cfg.Key }, want: "/tmp/key.pem"},
		{name: "timeout", got: func(cfg config.MessagingConfig) any { return cfg.Timeout }, want: 8 * time.Second},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadMessagingConfigFixture(t, validMessagingConfigINI())
			if got := testCase.got(cfg); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func TestLoadMessagingConfigUsesDefaultTimeout(t *testing.T) {
	t.Parallel()

	cfg := loadMessagingConfigFixture(t, "[messaging]\nurl=nats://localhost:4222\n")
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("cfg.Timeout = %v, want 5s", cfg.Timeout)
	}
}

func TestLoadMessagingConfigRejectsInvalidTimeout(t *testing.T) {
	t.Parallel()

	cfgFile, err := ini.Load([]byte("[messaging]\ntimeout=not-a-duration\n"))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	if _, err = config.LoadMessagingConfig(cfgFile); err == nil {
		t.Fatal("expected timeout parsing error")
	}
}

func validMessagingConfigINI() string {
	return "[messaging]\n" +
		"url=nats://localhost:4222\n" +
		"client_name=system-metrics\n" +
		"ca=/tmp/ca.pem\n" +
		"cert=/tmp/cert.pem\n" +
		"key=/tmp/key.pem\n" +
		"timeout=8s\n"
}

func loadMessagingConfigFixture(t *testing.T, iniContent string) config.MessagingConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadMessagingConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadMessagingConfig() error = %v", err)
	}

	return cfg
}
