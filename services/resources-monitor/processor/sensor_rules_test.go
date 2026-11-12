package processor

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	servicerules "lite-nas/services/resources-monitor/rules"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/eventmanager"
	sharedlogger "lite-nas/shared/logger"
)

// TestSensorSeverityBands verifies the packaged voltage and CPU temperature
// bands select exactly one severity and the requested email policy.
// Requirements: resources-monitor/FR-006.
func TestSensorSeverityBands(t *testing.T) {
	tests := []struct {
		name, category, key, want string
		value                     float64
	}{
		{"voltage warning", "voltages", "EXT5V_V", "warning:false", 4.89},
		{"voltage warning boundary", "voltages", "EXT5V_V", "warning:false", 4.85},
		{"voltage error", "voltages", "EXT5V_V", "error:true", 4.84},
		{"voltage critical boundary", "voltages", "EXT5V_V", "critical:true", 4.81},
		{"voltage critical", "voltages", "EXT5V_V", "critical:true", 4.80},
		{"temperature warning boundary", "temperatures", "thermal:cpu-thermal", "warning:false", 65},
		{"temperature error boundary", "temperatures", "thermal:cpu-thermal", "error:true", 70},
		{"temperature critical boundary", "temperatures", "thermal:cpu-thermal", "critical:true", 75},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := processSensorReading(t, test.category, test.key, test.value)
			alert := mustSingleSensorAlert(t, client)
			got := string(alert.Severity) + ":" + boolText(alert.Email)
			if got != test.want {
				t.Fatalf("alert severity/email = %q, want %q", got, test.want)
			}
		})
	}
}

// TestSensorReadingsBelowAlertRange verifies the upper boundaries are excluded
// from their warning bands. Requirements: resources-monitor/FR-006.
func TestSensorReadingsBelowAlertRange(t *testing.T) {
	tests := []struct {
		name, category, key string
		value               float64
	}{
		{"nominal voltage", "voltages", "EXT5V_V", 4.9},
		{"nominal CPU temperature", "temperatures", "thermal:cpu-thermal", 64.99},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := processSensorReading(t, test.category, test.key, test.value)
			if len(client.published) != 0 {
				t.Fatalf("published %d alerts for nominal reading", len(client.published))
			}
		})
	}
}

// TestSensorAlertKeepsIdentityWhenArrayPositionChanges verifies optional
// readings cannot create a second alert for the same physical sensor.
// Requirements: resources-monitor/FR-006.
func TestSensorAlertKeepsIdentityWhenArrayPositionChanges(t *testing.T) {
	rules := mustLoadSensorRules(t)
	client := &recordingClient{}
	processor := New(rules, eventmanager.NewManager(0), client, sharedlogger.NewNop())
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", sensorVoltagePayload(
		map[string]any{"key": "core-voltage", "value": 0.8},
		map[string]any{"key": "EXT5V_V", "value": 4.89},
	))
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", sensorVoltagePayload(
		map[string]any{"key": "EXT5V_V", "value": 4.89},
	))
	want := []string{systemloggingmanagercontract.AlertSubject, systemloggingmanagercontract.AlertOccurrenceSubject}
	if !slices.Equal(client.publishSubjects, want) {
		t.Fatalf("publish subjects = %v, want %v", client.publishSubjects, want)
	}
}

// TestSensorSeverityTransitionNormalizesLowerBand verifies an error reading
// replaces its prior warning lifecycle after the configured cooldown.
// Requirements: resources-monitor/FR-006.
func TestSensorSeverityTransitionNormalizesLowerBand(t *testing.T) {
	rules := mustLoadSensorRules(t)
	client := &recordingClient{requestResponse: loggingmanagercontract.OKResponse{OK: true}}
	manager := eventmanager.NewManager(0)
	processor := New(rules, manager, client, sharedlogger.NewNop())
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	processor.clock = func() time.Time { return now }
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", sensorVoltagePayload(
		map[string]any{"key": "EXT5V_V", "value": 4.89},
	))
	now = now.Add(time.Second)
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", sensorVoltagePayload(
		map[string]any{"key": "EXT5V_V", "value": 4.84},
	))
	now = now.Add(29 * time.Second)
	processor.ProcessCooldowns(context.Background())
	_, warningActive := manager.FindEvent("sensor.monitor.events.snapshot", "snapshot.voltages[].value", "<", "EXT5V_V", "snsvwrn")
	_, errorActive := manager.FindEvent("sensor.monitor.events.snapshot", "snapshot.voltages[].value", "<", "EXT5V_V", "snsverr")
	if warningActive || !errorActive {
		t.Fatalf("warning active = %t, error active = %t; want only error", warningActive, errorActive)
	}
}

// TestMissingSensorReadingKeepsActiveAlert verifies an absent reading does not
// imply recovery. Requirements: resources-monitor/FR-006.
func TestMissingSensorReadingKeepsActiveAlert(t *testing.T) {
	rules := mustLoadSensorRules(t)
	client := &recordingClient{}
	manager := eventmanager.NewManager(0)
	processor := New(rules, manager, client, sharedlogger.NewNop())
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", sensorVoltagePayload(
		map[string]any{"key": "EXT5V_V", "value": 4.89},
	))
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", sensorVoltagePayload())
	_, warningActive := manager.FindEvent("sensor.monitor.events.snapshot", "snapshot.voltages[].value", "<", "EXT5V_V", "snsvwrn")
	if !warningActive {
		t.Fatal("warning alert normalized when the reading disappeared")
	}
}

// TestSensorRulesIgnoreUnrelatedReadings verifies a low core rail cannot be
// mistaken for the 5 V input supply.
// Requirements: resources-monitor/FR-006.
func TestSensorRulesIgnoreUnrelatedReadings(t *testing.T) {
	client := processSensorReading(t, "voltages", "core-voltage", 0.8)
	if len(client.published) != 0 {
		t.Fatalf("published %d alerts for unrelated voltage", len(client.published))
	}
}

// TestSensorRulesIgnoreOtherTemperatures verifies the CPU thresholds do not
// classify an unrelated hot sensor.
// Requirements: resources-monitor/FR-006.
func TestSensorRulesIgnoreOtherTemperatures(t *testing.T) {
	client := processSensorReading(t, "temperatures", "gpu-temperature", 80)
	if len(client.published) != 0 {
		t.Fatalf("published %d alerts for unrelated temperature", len(client.published))
	}
}

// processSensorReading runs all packaged sensor rules against one keyed reading.
func processSensorReading(t *testing.T, category, key string, value float64) *recordingClient {
	t.Helper()
	rules := mustLoadSensorRules(t)
	client := &recordingClient{}
	processor := New(rules, eventmanager.NewManager(0), client, sharedlogger.NewNop())
	processor.ProcessSnapshot(context.Background(), "sensor.monitor.events.snapshot", map[string]any{
		"snapshot": map[string]any{category: []any{map[string]any{"key": key, "value": value}}},
	})
	return client
}

// mustLoadSensorRules loads the packaged rules for integration-oriented cases.
func mustLoadSensorRules(t *testing.T) []servicerules.Rule {
	t.Helper()
	path := filepath.Join("..", "..", "..", "configs", "etc", "lite-nas", "resources-monitor", "rules", "sensor-monitor.json")
	rules, err := servicerules.LoadRules([]string{path})
	if err != nil {
		t.Fatalf("LoadRules(): %v", err)
	}
	// These cases check sensor classification and lifecycle independently of
	// the configurable match window, which has dedicated coverage.
	for index := range rules {
		rules[index].MatchWindow = 0
		rules[index].MinMatches = 0
	}
	return rules
}

// sensorVoltagePayload makes an indexed voltage array for identity checks.
func sensorVoltagePayload(readings ...map[string]any) map[string]any {
	return map[string]any{"snapshot": map[string]any{"voltages": readings}}
}

// mustSingleSensorAlert returns the only expected alert creation.
func mustSingleSensorAlert(t *testing.T, client *recordingClient) loggingmanagercontract.AlertPayload {
	t.Helper()
	if len(client.published) != 1 {
		t.Fatalf("published %d alerts, want one", len(client.published))
	}
	alert, ok := client.published[0].(loggingmanagercontract.AlertPayload)
	if !ok {
		t.Fatalf("published type %T, want AlertPayload", client.published[0])
	}
	return alert
}

// boolText describes the alert's explicit email policy for the band assertion.
func boolText(value *bool) string {
	if value == nil {
		return "nil"
	}
	if *value {
		return "true"
	}
	return "false"
}
