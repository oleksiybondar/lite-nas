package plugins

import (
	"context"
	"testing"
	"time"

	sharedlogger "lite-nas/shared/logger"
)

func pluginFixture(t *testing.T, kind string, threshold int, keyBy string) (*Plugin, chan Detection) {
	t.Helper()
	output := make(chan Detection, 16)
	plugin, err := New(Config{
		ID: "auth1", Type: kind, Threshold: threshold, Window: 2 * time.Minute,
		Cooldown: 10 * time.Minute, Idle: 30 * time.Minute, KeyBy: keyBy, MaxKeys: 16,
	}, output, sharedlogger.NewNop())
	if err != nil {
		t.Fatalf("new plugin: %v", err)
	}
	return plugin, output
}

func authObservation(account string) Observation {
	return Observation{RuleID: "auth_fail", Category: "auth_failure", Severity: "warning", Kind: "failure", Account: account, At: time.Now()}
}

// Requirements: task.md/SVC-001.
func TestImmediatePluginEmitsOnFirstObservation(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "immediate", 1, "global")
	plugin.process(context.Background(), authObservation("alice"))
	if len(output) != 1 {
		t.Fatalf("detections = %d", len(output))
	}
}

// Requirements: task.md/SVC-001.
func TestThresholdPluginWaitsForThirdObservation(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "threshold", 3, "global")
	for range 2 {
		plugin.process(context.Background(), authObservation("alice"))
	}
	if len(output) != 0 {
		t.Fatalf("detections = %d", len(output))
	}
}

// Requirements: task.md/SVC-001.
func TestThresholdPluginTriggersAtBoundary(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "threshold", 3, "global")
	for range 3 {
		plugin.process(context.Background(), authObservation("alice"))
	}
	if len(output) != 1 {
		t.Fatalf("detections = %d", len(output))
	}
}

// Requirements: task.md/SVC-001.
func TestKeyedPluginKeepsAccountsIndependent(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "keyed_threshold", 3, "account")
	for range 2 {
		plugin.process(context.Background(), authObservation("alice"))
	}
	plugin.process(context.Background(), authObservation("bob"))
	if len(output) != 0 {
		t.Fatalf("detections = %d", len(output))
	}
}

// Requirements: task.md/SVC-001.
func TestKeyedPluginNormalizesAfterSuccess(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "keyed_threshold", 3, "account")
	for range 3 {
		plugin.process(context.Background(), authObservation("alice"))
	}
	success := authObservation("alice")
	success.Kind = "success"
	plugin.process(context.Background(), success)
	if len(output) != 2 {
		t.Fatalf("detections = %d", len(output))
	}
}

// Requirements: task.md/SVC-001.
func TestKeyedPluginDoesNotMergeMissingAccounts(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "keyed_threshold", 3, "account")
	for range 3 {
		plugin.process(context.Background(), authObservation(""))
	}
	if len(output) != 0 {
		t.Fatalf("detections = %d", len(output))
	}
}

// Requirements: task.md/SVC-001.
func TestCooldownSuppressesRepeatDetection(t *testing.T) {
	t.Parallel()
	plugin, output := pluginFixture(t, "threshold", 3, "global")
	for range 5 {
		plugin.process(context.Background(), authObservation("alice"))
	}
	if len(output) != 1 {
		t.Fatalf("detections = %d", len(output))
	}
}
