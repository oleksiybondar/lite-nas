package config_test

import (
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadMessagingConfigParsedFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.MessagingConfig]{
		{Name: "url", Got: func(cfg config.MessagingConfig) any { return cfg.URL }, Want: "nats://localhost:4222"},
		{Name: "client name", Got: func(cfg config.MessagingConfig) any { return cfg.ClientName }, Want: "system-metrics"},
		{Name: "ca path", Got: func(cfg config.MessagingConfig) any { return cfg.CA }, Want: "/tmp/ca.pem"},
		{Name: "cert path", Got: func(cfg config.MessagingConfig) any { return cfg.Cert }, Want: "/tmp/cert.pem"},
		{Name: "key path", Got: func(cfg config.MessagingConfig) any { return cfg.Key }, Want: "/tmp/key.pem"},
		{Name: "timeout", Got: func(cfg config.MessagingConfig) any { return cfg.Timeout }, Want: 8 * time.Second},
	}

	testcasetest.RunFieldCases(t, func(t *testing.T) config.MessagingConfig {
		t.Helper()
		return loadMessagingConfigFixture(t, validMessagingConfigINI())
	}, testCases)
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
