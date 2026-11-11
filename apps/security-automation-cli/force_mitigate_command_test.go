package main

import (
	"context"
	"errors"
	"testing"

	safetycontract "lite-nas/shared/contracts/operationalsafety"
	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/testutil/messagingtest"
)

// mitigationClient records the rule lookup and mitigation request in a command test.
type mitigationClient struct {
	messagingtest.RecordingClient
	subject     string
	request     contract.ForceMitigateRequest
	lastRequest any
	cpuRequest  safetycontract.OperationalSafetyCPUSetCapRequest
	cpuState    safetycontract.OperationalSafetyOutcomeState
	failOn      string
}

// Request supplies the rule definition and captures the final effect request.
func (client *mitigationClient) Request(_ context.Context, subject string, request any, response any) error {
	if subject == client.failOn {
		return errors.New("RPC unavailable")
	}
	switch subject {
	case contract.RulesListSubject:
		client.subject = subject
		*response.(*contract.RulesListResponse) = contract.RulesListResponse{Rules: []contract.Rule{{ID: "rule_1", RequiredOperands: []string{":username"}}}}
	case contract.ActiveListSubject:
		client.subject = subject
		*response.(*contract.ActiveListResponse) = contract.ActiveListResponse{}
	case contract.RollbackSubject, contract.ForceNormalizeSubject, contract.RuleSetEnabledSubject, contract.CooldownResetSubject:
		client.subject, client.lastRequest = subject, request
		*response.(*contract.OperationResponse) = contract.OperationResponse{Completed: true}
	case contract.ForceMitigateSubject:
		client.subject = subject
		client.request = request.(contract.ForceMitigateRequest)
		*response.(*contract.OperationResponse) = contract.OperationResponse{Completed: true}
	case safetycontract.CPUSetCapSubject:
		client.subject = subject
		client.cpuRequest = request.(safetycontract.OperationalSafetyCPUSetCapRequest)
		state := client.cpuState
		if state == "" {
			state = safetycontract.OutcomeCompleted
		}
		*response.(*safetycontract.OperationalSafetyCPUSetCapResponse) = safetycontract.OperationalSafetyCPUSetCapResponse{State: state, CapKHz: client.cpuRequest.CapKHz}
	default:
		return errors.New("unexpected RPC subject")
	}
	return nil
}

// Requirements: security-automation/IR-001.
// TestOperatorCommandsUseExpectedRPC checks each operator action routes to its contract subject.
func TestOperatorCommandsUseExpectedRPC(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		subject string
	}{
		{"rules", []string{"rules"}, contract.RulesListSubject},
		{"active", []string{"active"}, contract.ActiveListSubject},
		{"rollback", []string{"rollback", "incident1"}, contract.RollbackSubject},
		{"force normalize", []string{"force-normalize", "incident1"}, contract.ForceNormalizeSubject},
		{"enable", []string{"enable", "rule_1"}, contract.RuleSetEnabledSubject},
		{"disable", []string{"disable", "rule_1"}, contract.RuleSetEnabledSubject},
		{"cooldown reset", []string{"cooldown-reset", "rule_1", "alice"}, contract.CooldownResetSubject},
		{"manual CPU cap", []string{"cpu-set-cap", "1900000", "--yes"}, safetycontract.CPUSetCapSubject},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			client := &mitigationClient{}
			if err := executeCommand(context.Background(), client, "token", test.args); err != nil {
				t.Fatal(err)
			}
			if client.subject != test.subject {
				t.Fatalf("subject = %q, want %q", client.subject, test.subject)
			}
		})
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-012.
func TestManualCPUCapCommandUsesKHzRPCOperand(t *testing.T) {
	client := &mitigationClient{}
	if err := executeCommand(context.Background(), client, "token", []string{"cpu-set-cap", "1900000", "--yes"}); err != nil {
		t.Fatal(err)
	}
	if client.cpuRequest.CapKHz != 1900000 || client.cpuRequest.AccessToken != "token" {
		t.Fatalf("manual cap RPC = %#v", client.cpuRequest)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-012.
func TestManualCPUCapCommandFailsWhenRPCBlocksChange(t *testing.T) {
	client := &mitigationClient{cpuState: safetycontract.OutcomeBlocked}
	if err := executeCommand(context.Background(), client, "token", []string{"cpu-set-cap", "1900000", "--yes"}); err == nil {
		t.Fatal("blocked manual cap returned CLI success")
	}
}

// Requirements: security-automation/IR-001.
// TestOperatorCommandReportsRPCFailure checks transport errors are not reported as successful actions.
func TestOperatorCommandReportsRPCFailure(t *testing.T) {
	client := &mitigationClient{failOn: contract.RulesListSubject}
	if err := executeCommand(context.Background(), client, "token", []string{"rules"}); err == nil {
		t.Fatal("RPC failure was ignored")
	}
}

// Requirements: security-automation/IR-001.
// TestOperatorCommandRejectsMalformedInputBeforeRPC checks the command boundary precedes transport.
func TestOperatorCommandRejectsMalformedInputBeforeRPC(t *testing.T) {
	client := &mitigationClient{}
	if err := executeCommand(context.Background(), client, "token", []string{"enable", "--bad"}); err == nil || client.subject != "" {
		t.Fatalf("malformed command result = %v, subject = %q", err, client.subject)
	}
}

// Requirements: security-automation/IR-001.
// TestForceMitigateCommandRejectsUnknownRule checks a missing rule cannot reach the effect RPC.
func TestForceMitigateCommandRejectsUnknownRule(t *testing.T) {
	client := &mitigationClient{}
	err := executeCommand(context.Background(), client, "token", []string{"force-mitigate", "other", "--yes"})
	if err == nil || client.subject == contract.ForceMitigateSubject {
		t.Fatalf("unknown rule result = %v, subject = %q", err, client.subject)
	}
}

// Requirements: security-automation/IR-001.
// TestForceMitigateCommandUsesRuleOperands checks the CLI sends only operands defined by the selected rule.
func TestForceMitigateCommandUsesRuleOperands(t *testing.T) {
	client := &mitigationClient{}
	if err := executeCommand(context.Background(), client, "token", []string{"force-mitigate", "rule_1", "username=alice", "--yes"}); err != nil {
		t.Fatal(err)
	}
	if client.subject != contract.ForceMitigateSubject || client.request.Operands["username"] != "alice" {
		t.Fatalf("mitigation RPC = %q, %#v", client.subject, client.request)
	}
}
