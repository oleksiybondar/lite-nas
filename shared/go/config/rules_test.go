package config_test

import (
	"strings"
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
)

func TestLoadRulesConfigParsesCommaSeparatedFiles(t *testing.T) {
	t.Parallel()

	cfg := loadRulesConfigFixture(
		t,
		"[rules]\nfiles=/etc/lite-nas/rules/system.json, /etc/lite-nas/rules/network.json\n",
	)

	if len(cfg.Files) != 2 {
		t.Fatalf("len(cfg.Files) = %d, want 2", len(cfg.Files))
	}

	if cfg.Files[0] != "/etc/lite-nas/rules/system.json" {
		t.Fatalf("cfg.Files[0] = %q, want %q", cfg.Files[0], "/etc/lite-nas/rules/system.json")
	}

	if cfg.Files[1] != "/etc/lite-nas/rules/network.json" {
		t.Fatalf("cfg.Files[1] = %q, want %q", cfg.Files[1], "/etc/lite-nas/rules/network.json")
	}
}

func TestLoadRulesConfigRejectsMissingFiles(t *testing.T) {
	t.Parallel()

	cfgFile, err := ini.Load([]byte("[rules]\n"))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	_, err = config.LoadRulesConfig(cfgFile)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "rules files are required") {
		t.Fatalf("error = %q, want substring %q", err, "rules files are required")
	}
}

func TestLoadRulesConfigRejectsBlankFiles(t *testing.T) {
	t.Parallel()

	cfgFile, err := ini.Load([]byte("[rules]\nfiles= ,  \n"))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	_, err = config.LoadRulesConfig(cfgFile)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), "rules files are required") {
		t.Fatalf("error = %q, want substring %q", err, "rules files are required")
	}
}

func loadRulesConfigFixture(t *testing.T, iniContent string) config.RulesConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadRulesConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadRulesConfig() error = %v", err)
	}

	return cfg
}
