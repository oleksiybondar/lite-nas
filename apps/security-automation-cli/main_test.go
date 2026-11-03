package main

import "testing"

// TestValidateCommandRejectsMalformedArguments checks the CLI boundary before login.
func TestValidateCommandRejectsMalformedArguments(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"missing", nil},
		{"unknown", []string{"unknown"}},
		{"extra list operand", []string{"rules", "extra"}},
		{"option incident", []string{"rollback", "--help"}},
		{"option rule", []string{"enable", "-rule"}},
		{"blank key", []string{"cooldown-reset", "rule", ""}},
		{"option key", []string{"cooldown-reset", "rule", "--key"}},
		{"control key", []string{"cooldown-reset", "rule", "a\nb"}},
		{"missing confirmation", []string{"force-mitigate", "rule", "requester_ip_address=192.0.2.1"}},
		{"option operand", []string{"force-mitigate", "rule", "requester_ip_address=--bad", "--yes"}},
		{"duplicate operand", []string{"force-mitigate", "rule", "device_id=1", "device_id=2", "--yes"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := validateCommand(test.args); err == nil {
				t.Fatal("malformed command accepted")
			}
		})
	}
}

// TestValidateCommandAcceptsOperatorActions verifies each advertised action shape.
func TestValidateCommandAcceptsOperatorActions(t *testing.T) {
	cases := [][]string{
		{"rules"},
		{"active"},
		{"rollback", "incident1"},
		{"force-normalize", "incident1"},
		{"enable", "rule_1"},
		{"disable", "rule_1"},
		{"cooldown-reset", "rule_1", "192.0.2.1"},
		{"force-mitigate", "rule_1", "requester_ip_address=192.0.2.1", "--yes"},
		{"force-mitigate", "rule_1", "--yes"},
	}
	for _, args := range cases {
		t.Run(args[0], func(t *testing.T) {
			if err := validateCommand(args); err != nil {
				t.Fatal(err)
			}
		})
	}
}
