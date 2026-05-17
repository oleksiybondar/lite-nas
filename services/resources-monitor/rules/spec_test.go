package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRulesSuccess(t *testing.T) {
	t.Parallel()

	path := writeRulesFile(t, `[
		{
			"event":"system.metrics.events.stats",
			"event_prefix":"syscpu",
			"field":"snapshot.cpu.totalUsagePct",
			"condition":">=",
			"values":90,
			"message":"cpu high",
			"category":"system.metrics.cpu.total",
			"severity":"warning",
			"priority":2,
			"source":"system-metrics"
		}
	]`)

	loadedRules, err := LoadRules([]string{path})
	if err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}
	if len(loadedRules) != 1 {
		t.Fatalf("len(loadedRules) = %d, want 1", len(loadedRules))
	}
}

func TestLoadRulesRejectsInvalidInValues(t *testing.T) {
	t.Parallel()

	path := writeRulesFile(t, `[
		{
			"event":"system.metrics.events.stats",
			"event_prefix":"syscpu",
			"field":"snapshot.cpu.totalUsagePct",
			"condition":"in",
			"values":90,
			"message":"cpu high",
			"category":"system.metrics.cpu.total",
			"severity":"warning",
			"priority":2,
			"source":"system-metrics"
		}
	]`)

	if _, err := LoadRules([]string{path}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadRulesRejectsNonNumericComparisonValue(t *testing.T) {
	t.Parallel()

	path := writeRulesFile(t, `[
		{
			"event":"system.metrics.events.stats",
			"event_prefix":"syscpu",
			"field":"snapshot.cpu.totalUsagePct",
			"condition":">=",
			"values":"90",
			"message":"cpu high",
			"category":"system.metrics.cpu.total",
			"severity":"warning",
			"priority":2,
			"source":"system-metrics"
		}
	]`)

	if _, err := LoadRules([]string{path}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadRulesRejectsMissingFiles(t *testing.T) {
	t.Parallel()

	if _, err := LoadRules(nil); err == nil {
		t.Fatal("expected error for empty files input")
	}
}

func writeRulesFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rules.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return path
}
