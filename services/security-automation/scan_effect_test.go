package main

import (
	"testing"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/logger"
)

// TestScanSequenceQueuesAllTrafficBlock checks the completed sequence persists
// its distinct UFW action before the effect worker touches the host.
// Requirements: security-automation/FR-007.
func TestScanSequenceQueuesAllTrafficBlock(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	store := mustIncidentStore(t)
	runtime := &automationRuntime{
		rules: rules, incidents: store, operatorActions: make(chan effectCommand, 1), log: logger.NopLogger{},
		physicalInterfaces: func() ([]string, error) { return []string{"physical0"}, nil },
	}
	intent := detectionIntent{RuleID: "network_scan_sequence", Match: detection.Match{Key: "198.51.100.10", AlertIDs: []string{"secnet_1", "secnet_2"}, LastMessage: "Blocked TCP port scan from 198.51.100.10 to 192.0.2.10: 12 distinct ports in 2m0s"}}
	if err := runtime.queueAutomaticScanUFW(intent); err != nil {
		t.Fatal(err)
	}
	got := store.list()
	if len(got) != 1 || got[0].ActionName != "block_scan_source_ip" || got[0].SourceIP != "198.51.100.10" {
		t.Fatalf("scan incident = %+v", got)
	}
}

// TestScanSequenceRejectsChangedSource keeps a mismatched final alert from
// authorizing a source-wide UFW block.
// Requirements: security-automation/SR-002.
func TestScanSequenceRejectsChangedSource(t *testing.T) {
	runtime := &automationRuntime{}
	intent := detectionIntent{RuleID: "network_scan_sequence", Match: detection.Match{Key: "198.51.100.10", AlertIDs: []string{"secnet_1", "secnet_2"}, LastMessage: "Blocked TCP port scan from 198.51.100.11 to 192.0.2.10: 12 distinct ports in 2m0s"}}
	if err := runtime.queueAutomaticScanUFW(intent); err == nil {
		t.Fatal("mismatched scan source accepted")
	}
}
