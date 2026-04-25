package config_test

import (
	"strings"
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
)

// Requirements: web-gateway/OR-001
func TestLoadHTTPConfigParsedFields(t *testing.T) {
	t.Parallel()

	cfg := loadHTTPConfigFixture(t, "[http]\naddress=127.0.0.1:9191\n")
	if cfg.Address != "127.0.0.1:9191" {
		t.Fatalf("cfg.Address = %q, want %q", cfg.Address, "127.0.0.1:9191")
	}
}

// Requirements: web-gateway/OR-001
func TestLoadHTTPConfigUsesDefaultAddress(t *testing.T) {
	t.Parallel()

	cfg := loadHTTPConfigFixture(t, "[http]\n")
	if cfg.Address != "127.0.0.1:9090" {
		t.Fatalf("cfg.Address = %q, want %q", cfg.Address, "127.0.0.1:9090")
	}
}

// Requirements: web-gateway/OR-001
func TestLoadHTTPConfigRejectsBlankAddress(t *testing.T) {
	t.Parallel()

	cfgFile, err := ini.Load([]byte("[http]\naddress=   \n"))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	_, err = config.LoadHTTPConfig(cfgFile)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "http address is required") {
		t.Fatalf("error = %q, want substring %q", err, "http address is required")
	}
}

func loadHTTPConfigFixture(t *testing.T, iniContent string) config.HTTPConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadHTTPConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadHTTPConfig() error = %v", err)
	}

	return cfg
}
