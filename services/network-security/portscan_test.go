package main

import (
	"context"
	"testing"
	"time"

	"lite-nas/shared/contracts/networksecurity"
)

// mustPortScanPlugin creates the configured FSM for one detector test.
func mustPortScanPlugin(t *testing.T, now *time.Time) *portScanPlugin {
	t.Helper()
	plugin, err := newPortScanPlugin(3, 16, time.Minute, 5*time.Minute, func() time.Time { return *now })
	if err != nil {
		t.Fatalf("create port scan plugin: %v", err)
	}
	return plugin
}

// mustObserve sends one valid attempt and returns its effects.
func mustObserve(t *testing.T, plugin *portScanPlugin, port uint16) []scanDetection {
	return mustObserveTarget(t, plugin, "192.0.2.10", port)
}

// mustObserveTarget sends a valid attempt to one destination.
func mustObserveTarget(t *testing.T, plugin *portScanPlugin, target string, port uint16) []scanDetection {
	t.Helper()
	effects, err := plugin.Observe(context.Background(), networksecurity.FirewallAttempt{
		SourceIP: "192.0.2.4", DestinationIP: target, DestinationPort: port, Protocol: "TCP",
	})
	if err != nil {
		t.Fatalf("observe port %d: %v", port, err)
	}
	return effects
}

// mustAdvance applies expected timeout transitions after moving the clock.
func mustAdvance(t *testing.T, plugin *portScanPlugin) {
	t.Helper()
	if err := plugin.Advance(context.Background()); err != nil {
		t.Fatalf("advance: %v", err)
	}
}

// TestDistinctPortsRaiseOneFinding verifies the threshold counts unique ports.
// Requirements: network-security/FR-002.
func TestDistinctPortsRaiseOneFinding(t *testing.T) {
	now := time.Now()
	plugin := mustPortScanPlugin(t, &now)
	mustObserve(t, plugin, 21)
	mustObserve(t, plugin, 22)
	effects := mustObserve(t, plugin, 23)
	if len(effects) != 1 || effects[0].PortCount != 3 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestRepeatedPortDoesNotReachThreshold verifies duplicate probes are ignored.
// Requirements: network-security/FR-002.
func TestRepeatedPortDoesNotReachThreshold(t *testing.T) {
	now := time.Now()
	plugin := mustPortScanPlugin(t, &now)
	mustObserve(t, plugin, 21)
	mustObserve(t, plugin, 21)
	effects := mustObserve(t, plugin, 22)
	if len(effects) != 0 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestDifferentDestinationsRemainSeparate verifies one host's ports do not
// contribute to another host's scan threshold.
// Requirements: network-security/FR-002.
func TestDifferentDestinationsRemainSeparate(t *testing.T) {
	now := time.Now()
	plugin := mustPortScanPlugin(t, &now)
	mustObserve(t, plugin, 21)
	mustObserve(t, plugin, 22)
	effects := mustObserveTarget(t, plugin, "192.0.2.11", 23)
	if len(effects) != 0 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestCooldownSuppressesFurtherFindings verifies one burst yields one alert.
// Requirements: network-security/FR-003.
func TestCooldownSuppressesFurtherFindings(t *testing.T) {
	now := time.Now()
	plugin := mustPortScanPlugin(t, &now)
	mustObserve(t, plugin, 21)
	mustObserve(t, plugin, 22)
	mustObserve(t, plugin, 23)
	effects := mustObserve(t, plugin, 24)
	if len(effects) != 0 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestCooldownAllowsNewBurstAfterExpiry verifies a later burst is eligible.
// Requirements: network-security/FR-003.
func TestCooldownAllowsNewBurstAfterExpiry(t *testing.T) {
	now := time.Now()
	plugin := mustPortScanPlugin(t, &now)
	mustObserve(t, plugin, 21)
	mustObserve(t, plugin, 22)
	mustObserve(t, plugin, 23)
	now = now.Add(6 * time.Minute)
	mustAdvance(t, plugin)
	mustObserve(t, plugin, 24)
	mustObserve(t, plugin, 25)
	effects := mustObserve(t, plugin, 26)
	if len(effects) != 1 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestWindowExpiryResetsPortSet verifies old ports do not remain counted.
// Requirements: network-security/FR-002.
func TestWindowExpiryResetsPortSet(t *testing.T) {
	now := time.Now()
	plugin := mustPortScanPlugin(t, &now)
	mustObserve(t, plugin, 21)
	mustObserve(t, plugin, 22)
	now = now.Add(2 * time.Minute)
	mustAdvance(t, plugin)
	effects := mustObserve(t, plugin, 23)
	if len(effects) != 0 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestHorizontalScanCountsDestinations verifies one port across hosts.
// Requirements: network-security/FR-005.
func TestHorizontalScanCountsDestinations(t *testing.T) {
	now := time.Now()
	plugin, err := newPatternPlugin("horizontal_scan", 3, 16, time.Minute, 5*time.Minute, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"192.0.2.10", "192.0.2.11"} {
		if _, err := plugin.Observe(context.Background(), networksecurity.FirewallAttempt{SourceIP: "192.0.2.4", DestinationIP: target, DestinationPort: 80, Protocol: "TCP"}); err != nil {
			t.Fatal(err)
		}
	}
	effects, err := plugin.Observe(context.Background(), networksecurity.FirewallAttempt{SourceIP: "192.0.2.4", DestinationIP: "192.0.2.12", DestinationPort: 80, Protocol: "TCP"})
	if err != nil {
		t.Fatal(err)
	}
	if len(effects) != 1 || effects[0].Kind != "horizontal_scan" {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestRepeatedAttemptsCountsSameTarget verifies repeated blocked connections.
// Requirements: network-security/FR-006.
func TestRepeatedAttemptsCountsSameTarget(t *testing.T) {
	now := time.Now()
	plugin, err := newPatternPlugin("repeated_attempts", 3, 16, time.Minute, 5*time.Minute, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	attempt := networksecurity.FirewallAttempt{SourceIP: "192.0.2.4", DestinationIP: "192.0.2.10", DestinationPort: 80, Protocol: "TCP"}
	for range 2 {
		if _, err := plugin.Observe(context.Background(), attempt); err != nil {
			t.Fatal(err)
		}
	}
	effects, err := plugin.Observe(context.Background(), attempt)
	if err != nil {
		t.Fatal(err)
	}
	if len(effects) != 1 || effects[0].Kind != "repeated_attempts" {
		t.Fatalf("findings = %+v", effects)
	}
}

// mustPatternPlugin creates an additional detector with a controllable clock.
func mustPatternPlugin(t *testing.T, kind string, now *time.Time) *portScanPlugin {
	t.Helper()
	plugin, err := newPatternPlugin(kind, 3, 16, time.Minute, 5*time.Minute, func() time.Time { return *now })
	if err != nil {
		t.Fatalf("create %s plugin: %v", kind, err)
	}
	return plugin
}

// TestPatternPluginRejectsUnknownKind checks compiled detector selection.
// Requirements: network-security/FR-005.
func TestPatternPluginRejectsUnknownKind(t *testing.T) {
	_, err := newPatternPlugin("unknown", 3, 16, time.Minute, 5*time.Minute, time.Now)
	if err == nil {
		t.Fatal("unknown detector kind was accepted")
	}
}

// mustObserveAttempt advances one detector with a trusted attempt.
func mustObserveAttempt(t *testing.T, plugin *portScanPlugin, source, destination string, port uint16) []scanDetection {
	t.Helper()
	effects, err := plugin.Observe(context.Background(), networksecurity.FirewallAttempt{
		SourceIP: source, DestinationIP: destination, DestinationPort: port, Protocol: "TCP",
	})
	if err != nil {
		t.Fatalf("observe attempt: %v", err)
	}
	return effects
}

// TestHorizontalScanIgnoresDuplicateDestinations checks unique-host counting.
// Requirements: network-security/FR-005.
func TestHorizontalScanIgnoresDuplicateDestinations(t *testing.T) {
	now := time.Now()
	plugin := mustPatternPlugin(t, "horizontal_scan", &now)
	mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.10", 80)
	mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.10", 80)
	effects := mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.11", 80)
	if len(effects) != 0 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestHorizontalScanSeparatesSourcesAndPorts checks correlation keys.
// Requirements: network-security/FR-005.
func TestHorizontalScanSeparatesSourcesAndPorts(t *testing.T) {
	for _, testCase := range []struct {
		name, source string
		port         uint16
	}{
		{name: "source", source: "192.0.2.5", port: 80},
		{name: "port", source: "192.0.2.4", port: 81},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			now := time.Now()
			plugin := mustPatternPlugin(t, "horizontal_scan", &now)
			mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.10", 80)
			mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.11", 80)
			effects := mustObserveAttempt(t, plugin, testCase.source, "192.0.2.12", testCase.port)
			if len(effects) != 0 {
				t.Fatalf("findings = %+v", effects)
			}
		})
	}
}

// TestRepeatedAttemptsSeparatesPorts checks per-service correlation.
// Requirements: network-security/FR-006.
func TestRepeatedAttemptsSeparatesPorts(t *testing.T) {
	now := time.Now()
	plugin := mustPatternPlugin(t, "repeated_attempts", &now)
	mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.10", 80)
	mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.10", 80)
	effects := mustObserveAttempt(t, plugin, "192.0.2.4", "192.0.2.10", 81)
	if len(effects) != 0 {
		t.Fatalf("findings = %+v", effects)
	}
}

// TestAdditionalDetectorsCooldown checks suppression and later eligibility.
// Requirements: network-security/FR-003, network-security/FR-005, network-security/FR-006.
func TestAdditionalDetectorsCooldown(t *testing.T) {
	for _, kind := range []string{"horizontal_scan", "repeated_attempts"} {
		t.Run(kind, func(t *testing.T) {
			now := time.Now()
			plugin := mustPatternPlugin(t, kind, &now)
			for i := 0; i < 3; i++ {
				destination := "192.0.2.10"
				if kind == "horizontal_scan" {
					destination = []string{"192.0.2.10", "192.0.2.11", "192.0.2.12"}[i]
				}
				mustObserveAttempt(t, plugin, "192.0.2.4", destination, 80)
			}
			destination := "192.0.2.10"
			if kind == "horizontal_scan" {
				destination = "192.0.2.13"
			}
			effects := mustObserveAttempt(t, plugin, "192.0.2.4", destination, 80)
			if len(effects) != 0 {
				t.Fatalf("findings during cooldown = %+v", effects)
			}
		})
	}
}

// TestAdditionalDetectorsCooldownExpires checks a fresh burst after timeout.
// Requirements: network-security/FR-003, network-security/FR-005, network-security/FR-006.
func TestAdditionalDetectorsCooldownExpires(t *testing.T) {
	for _, kind := range []string{"horizontal_scan", "repeated_attempts"} {
		t.Run(kind, func(t *testing.T) {
			now := time.Now()
			plugin := mustPatternPlugin(t, kind, &now)
			for i := 0; i < 3; i++ {
				destination := "192.0.2.10"
				if kind == "horizontal_scan" {
					destination = []string{"192.0.2.10", "192.0.2.11", "192.0.2.12"}[i]
				}
				mustObserveAttempt(t, plugin, "192.0.2.4", destination, 80)
			}
			now = now.Add(6 * time.Minute)
			mustAdvance(t, plugin)
			for i := 0; i < 2; i++ {
				destination := "192.0.2.10"
				if kind == "horizontal_scan" {
					destination = []string{"192.0.2.13", "192.0.2.14"}[i]
				}
				mustObserveAttempt(t, plugin, "192.0.2.4", destination, 80)
			}
			destination := "192.0.2.10"
			if kind == "horizontal_scan" {
				destination = "192.0.2.15"
			}
			effects := mustObserveAttempt(t, plugin, "192.0.2.4", destination, 80)
			if len(effects) != 1 {
				t.Fatalf("findings after cooldown = %+v", effects)
			}
		})
	}
}
