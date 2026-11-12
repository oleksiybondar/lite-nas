package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const repositoryConfig = "../../configs/etc/lite-nas/security-automation.conf"

// mustConfigCopy writes a test variant of the packaged startup configuration.
func mustConfigCopy(t *testing.T, replaceOld, replaceNew string) string {
	t.Helper()
	data, err := os.ReadFile(repositoryConfig)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "automation.conf")
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), replaceOld, replaceNew, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// Requirements: security-automation/OR-001.
// TestPackagedAutomationConfigLoads checks the shipped transport, auth, and mode sections parse together.
func TestPackagedAutomationConfigLoads(t *testing.T) {
	cfg, err := loadAutomationConfig(repositoryConfig)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != "permissive" || cfg.RulesFile != "/etc/lite-nas/security-automation/rules.json" {
		t.Fatalf("packaged automation config = %+v", cfg)
	}
}

// Requirements: security-automation/OR-001.
// TestAutomationConfigRejectsInvalidMode checks a misspelled action policy fails startup.
func TestAutomationConfigRejectsInvalidMode(t *testing.T) {
	path := mustConfigCopy(t, "mode=permissive", "mode=unsafe")
	if _, err := loadAutomationConfig(path); err == nil {
		t.Fatal("invalid automation mode accepted")
	}
}

// Requirements: security-automation/OR-001.
// TestAutomationConfigRejectsRelativeRules checks startup never resolves rules against process working directory.
func TestAutomationConfigRejectsRelativeRules(t *testing.T) {
	path := mustConfigCopy(t, "rules_file=/etc/lite-nas/security-automation/rules.json", "rules_file=rules.json")
	if _, err := loadAutomationConfig(path); err == nil {
		t.Fatal("relative rules path accepted")
	}
}
