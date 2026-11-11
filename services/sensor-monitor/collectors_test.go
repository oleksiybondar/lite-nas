package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestLinuxCollectorPublishesAvailableReadings verifies one failed source does
// not remove another reading from the same collection cycle (FR-002).
func TestLinuxCollectorPublishesAvailableReadings(t *testing.T) {
	root := t.TempDir()
	chip := filepath.Join(root, "hwmon", "hwmon0")
	mustWriteSensorFile(t, filepath.Join(chip, "name"), "testchip\n")
	mustWriteSensorFile(t, filepath.Join(chip, "temp1_input"), "42000\n")
	mustWriteSensorFile(t, filepath.Join(chip, "in0_input"), "invalid\n")
	collector := LinuxCollector{
		hwmonRoot:   filepath.Join(root, "hwmon"),
		thermalRoot: filepath.Join(root, "thermal"),
		clockRoot:   filepath.Join(root, "cpufreq"),
	}
	snapshot, err := collector.Collect(context.Background())
	if err == nil || len(snapshot.Temperatures) != 1 || snapshot.Temperatures[0].Value != 42 {
		t.Fatalf("Collect() = (%+v, %v), want one valid temperature and an error", snapshot, err)
	}
}

// TestParseThrottleFlagsDecodesLatchedUndervoltage verifies the documented
// firmware history flag remains observable even when current voltage is normal.
func TestParseThrottleFlagsDecodesLatchedUndervoltage(t *testing.T) {
	flags, err := parseThrottleFlags("throttled=0x10000\n", nil)
	if err != nil {
		t.Fatalf("parseThrottleFlags(): %v", err)
	}
	if !flags[4].Active {
		t.Fatal("under_voltage_occurred = false, want true")
	}
}

// TestParsePMICVoltageRejectsUnrelatedChannels ensures a Pi firmware response
// cannot substitute another rail for the configured supply channel.
func TestParsePMICVoltageRejectsUnrelatedChannels(t *testing.T) {
	if _, err := parsePMICVoltage("VDD_CORE_V volt(15)=0.75186740V\n"); err == nil {
		t.Fatal("parsePMICVoltage() error = nil for unrelated channel")
	}
}

// TestFirmwareRunnerRejectsOptionShapedOperation verifies unsupported command
// operands never reach the external executable.
func TestFirmwareRunnerRejectsOptionShapedOperation(t *testing.T) {
	collector := RaspberryPiCollector{command: "/usr/bin/vcgencmd"}
	if _, err := collector.run(context.Background(), "--help"); err == nil {
		t.Fatal("run() accepted option-shaped operation")
	}
}

// TestHistoryStoreKeepsLatestWithoutRetention verifies latest and history are
// independent when retention is disabled (FR-004).
func TestHistoryStoreKeepsLatestWithoutRetention(t *testing.T) {
	store := NewHistoryStore(0)
	store.Add(emptySnapshot())
	_, available := store.Latest()
	if !available {
		t.Fatal("Latest() available = false")
	}
}

// TestNameCatalogFallsBackToRawKey verifies an unmapped sensor stays identifiable.
func TestNameCatalogFallsBackToRawKey(t *testing.T) {
	if got := (NameCatalog{}).Resolve("voltage", "rail-x", ""); got != "rail-x" {
		t.Fatalf("Resolve() = %q, want rail-x", got)
	}
}

// mustWriteSensorFile creates a file in the collector fixture.
func mustWriteSensorFile(t *testing.T, path, value string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(): %v", err)
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}
}
