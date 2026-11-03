package main

import (
	"context"
	"testing"
	"time"

	alertcontract "lite-nas/shared/contracts/loggingmanager"
	"lite-nas/shared/logger"
)

// Requirements: security-automation/IR-002.
// TestListenerRoutesSecurityEventsAndTicks checks the bounded listener feeds its two dispatcher inputs.
func TestListenerRoutesSecurityEventsAndTicks(t *testing.T) {
	rules, err := loadAutomationRules(mustRuleCopy(t, `"id": "ssh_brute_force", "name": "SSH brute force", "description": "Aggregated SSH failures from one source address.", "enabled": false`, `"id": "ssh_brute_force", "name": "SSH brute force", "description": "Aggregated SSH failures from one source address.", "enabled": true`))
	if err != nil {
		t.Fatal(err)
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
		t.Fatal(err)
	}
	for id, rule := range rules {
		if id != "ssh_brute_force" {
			rule.enabled.Store(false)
		}
	}
	alerts := make(chan alertObservation, 1)
	timerTicks := make(chan struct{}, 1)
	observations, deadlines, done := startAutomationListener(context.Background(), alerts, timerTicks, rules, logger.NopLogger{})
	alerts <- alertObservation{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "secssh_1", Source: "log-analysis", Message: "Security rule triggered: ssh_preauth_probing from 192.0.2.1"}}
	timerTicks <- struct{}{}
	close(alerts)
	select {
	case item := <-observations:
		if item.RuleID != "ssh_brute_force" || item.Event.Key != "192.0.2.1" {
			t.Fatalf("routed observation = %+v", item)
		}
	case <-time.After(time.Second):
		t.Fatal("listener did not route alert")
	}
	select {
	case <-deadlines:
	case <-time.After(time.Second):
		t.Fatal("listener did not forward timer tick")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("listener did not drain and stop")
	}
}
