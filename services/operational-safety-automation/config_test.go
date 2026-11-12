package main

import "testing"

// TestPackagedConfigNamesOperationalSafetyRules verifies the packaged startup path.
func TestPackagedConfigNamesOperationalSafetyRules(t *testing.T) {
	config, err := loadServiceConfig("../../configs/etc/lite-nas/operational-safety-automation.conf")
	if err != nil {
		t.Fatal(err)
	}
	if config.RulesFile != "/etc/lite-nas/operational-safety-automation/rules.json" {
		t.Fatalf("rules file = %q", config.RulesFile)
	}
}
