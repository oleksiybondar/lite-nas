package main

import (
	"os"
	"strings"
	"testing"

	"lite-nas/shared/testutil/fileiotest"
)

// mustAnalysisConfigBytes reads the packaged configuration used by boundary tests.
func mustAnalysisConfigBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../configs/etc/lite-nas/log-analysis.conf")
	if err != nil {
		t.Fatalf("read packaged analysis config: %v", err)
	}
	return data
}

// Requirements: security-monitoring/FR-001.
func TestPackagedAnalysisConfigLoads(t *testing.T) {
	t.Parallel()
	cfg, err := loadAnalysisConfig(fileiotest.Reader{Data: mustAnalysisConfigBytes(t)})
	if err != nil {
		t.Fatalf("load analysis config: %v", err)
	}
	if cfg.RulesFile != "/etc/lite-nas/log-analysis/rules.json" {
		t.Fatalf("rules file = %q", cfg.RulesFile)
	}
}

// Requirements: security-monitoring/FR-001.
func TestAnalysisConfigRejectsInvalidBoundaryValues(t *testing.T) {
	t.Parallel()
	cases := map[string][2]string{
		"relative rules path": {"rules_file=/etc/lite-nas/log-analysis/rules.json", "rules_file=rules.json"},
		"invalid sender":      {"sender_users=syslog,root", "sender_users=-bad"},
		"missing login":       {"service_login=lite-nas-log-analysis", "service_login="},
	}
	for name, replacement := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			data := strings.Replace(string(mustAnalysisConfigBytes(t)), replacement[0], replacement[1], 1)
			if _, err := loadAnalysisConfig(fileiotest.Reader{Data: []byte(data)}); err == nil {
				t.Fatal("invalid analysis configuration accepted")
			}
		})
	}
}
