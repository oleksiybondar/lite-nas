package main

import (
	"regexp"
	"testing"

	alertcontract "lite-nas/shared/contracts/loggingmanager"
)

// Requirements: operational-safety-automation/operator-surface.
func TestSelectorCombinesIDMessageValueAndState(t *testing.T) {
	rule := &compiledRule{}
	step := selector{IDPrefix: "thermal", State: "active"}
	step.Message.Contains = "CPU"
	minimum := 80.0
	step.Value.NumGTE = &minimum
	rule.spec.CorrelateBy = "event_id"
	rule.spec.Steps = []selector{step}
	rule.patterns = []*regexp.Regexp{nil}
	value := 85.0
	event := observedEvent{alert: alertcontract.AlertPayload{EventID: "thermal_1", Message: "CPU hot"}, valueNum: &value, state: "active"}
	key, matched := rule.matches(0, "state", "thermal_1", event)
	if !matched || key != "thermal_1" {
		t.Fatalf("selector result = %q, %t", key, matched)
	}
}
