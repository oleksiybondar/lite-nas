package main

import (
	"testing"

	detection "lite-nas/shared/automationdetection"
)

// Requirements: operational-safety-automation/operator-surface.
func TestRuleAllowsIDPrefixWithoutAdditionalConditions(t *testing.T) {
	spec := ruleSpec{CorrelateBy: "event_id", Steps: []selector{{IDPrefix: "thermal"}}}
	enabled := true
	spec.Enabled = &enabled
	spec.Action.Kind = "emit"
	spec.Output.IDPrefix = "safety"
	spec.Output.Category = "system"
	spec.Output.Severity = "warning"
	spec.Output.Message = "Thermal alert"
	spec.FSM.Type = detection.CountWithinWindow
	spec.FSM.Window = "1m"
	spec.FSM.Threshold = 1
	spec.FSM.MaxKeys = 16
	spec.FSM.MaxIdle = "2m"
	if _, err := compileRule(spec); err != nil {
		t.Fatalf("ID prefix selector rejected: %v", err)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestRuleGraphRejectsEmittedAlertCycle(t *testing.T) {
	first := &compiledRule{}
	first.spec.Output.IDPrefix = "alpha"
	first.spec.Steps = []selector{{IDPrefix: "beta"}}
	second := &compiledRule{}
	second.spec.Output.IDPrefix = "beta"
	second.spec.Steps = []selector{{IDPrefix: "alpha"}}
	if err := validateRuleGraph(map[string]*compiledRule{"first": first, "second": second}); err == nil {
		t.Fatal("publication cycle accepted")
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestRuleGraphAllowsDownstreamFSM(t *testing.T) {
	first := &compiledRule{}
	first.spec.Output.IDPrefix = "alpha"
	first.spec.Steps = []selector{{IDPrefix: "source"}}
	second := &compiledRule{}
	second.spec.Output.IDPrefix = "beta"
	second.spec.Steps = []selector{{IDPrefix: "alpha"}}
	if err := validateRuleGraph(map[string]*compiledRule{"first": first, "second": second}); err != nil {
		t.Fatalf("downstream FSM rejected: %v", err)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-001.
func TestRaspberryPiRuleExampleCompiles(t *testing.T) {
	rules, order, err := loadRules("../../docs/examples/operational-safety-rpi-cpu-rules.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 4 || len(order) != 4 {
		t.Fatalf("example has %d rules and %d ordered IDs", len(rules), len(order))
	}
}
