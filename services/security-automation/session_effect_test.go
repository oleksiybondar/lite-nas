package main

import (
	"context"
	"testing"
	"time"

	"lite-nas/services/security-automation/detection"
	authcontract "lite-nas/shared/contracts/auth"
	contract "lite-nas/shared/contracts/sessionmanager"
)

// sessionEffectClient records whether the Session Manager terminate RPC was
// reached after inventory validation.
type sessionEffectClient struct {
	sessions   []contract.Session
	terminated string
}

func (client *sessionEffectClient) Publish(context.Context, string, any) error { return nil }

func (client *sessionEffectClient) Request(_ context.Context, subject string, request any, response any) error {
	switch subject {
	case authcontract.ServiceTokenLoginRPCSubject:
		*response.(*authcontract.ServiceTokenLoginResponse) = authcontract.ServiceTokenLoginResponse{AccessToken: "service-token", ExpiresAt: time.Now().Add(time.Hour)}
	case contract.ListSubject:
		*response.(*contract.ListResponse) = contract.ListResponse{Sessions: client.sessions}
	case contract.TerminateSubject:
		client.terminated = request.(contract.ActionRequest).Session
		*response.(*contract.ActionResponse) = contract.ActionResponse{Completed: true}
	}
	return nil
}

func (client *sessionEffectClient) Drain() error { return nil }
func (client *sessionEffectClient) Close()       {}

// TestSessionTerminationRequiresUniqueOwnerAndTTY checks an ambiguous
// inventory cannot terminate an arbitrary matching login session.
// Requirements: security-automation/FR-007.
func TestSessionTerminationRequiresUniqueOwnerAndTTY(t *testing.T) {
	client := &sessionEffectClient{sessions: []contract.Session{{ID: "4", User: "alice", TTY: "pts/2"}, {ID: "5", User: "alice", TTY: "pts/2"}}}
	terminate, err := newSessionTermination(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	_, matched, err := terminate(context.Background(), "alice", "pts/2")
	if err != nil || matched || client.terminated != "" {
		t.Fatalf("ambiguous session termination = %t, %q, %v", matched, client.terminated, err)
	}
}

// TestSessionTerminationUsesExactOwnerAndTTY checks the verified session ID
// is the only operand sent to the existing Session Manager action RPC.
// Requirements: security-automation/FR-007.
func TestSessionTerminationUsesExactOwnerAndTTY(t *testing.T) {
	client := &sessionEffectClient{sessions: []contract.Session{{ID: "4", User: "bob", TTY: "pts/2"}, {ID: "5", User: "alice", TTY: "pts/2"}}}
	terminate, err := newSessionTermination(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	_, matched, err := terminate(context.Background(), "alice", "pts/2")
	if err != nil || !matched || client.terminated != "5" {
		t.Fatalf("verified session termination = %t, %q, %v", matched, client.terminated, err)
	}
}

// TestSudoAlertQueuesExactAccountAndTTY verifies the detector key and alert
// evidence agree before a one-way session effect is journaled.
// Requirements: security-automation/FR-007.
func TestSudoAlertQueuesExactAccountAndTTY(t *testing.T) {
	runtime := &automationRuntime{incidents: mustIncidentStore(t), operatorActions: make(chan effectCommand, 1)}
	intent := detectionIntent{RuleID: "repeated_sudo_denial", Match: detection.Match{Key: "5:alicepts/2", AlertIDs: []string{"secpriv1", "secpriv2", "secpriv3"}, LastMessage: "Security rule triggered: sudo_policy_violation for account alice tty pts/2"}}
	if err := runtime.queueSessionTermination(intent); err != nil {
		t.Fatal(err)
	}
	if got := runtime.incidents.list(); len(got) != 1 || got[0].Username != "alice" || got[0].TTY != "pts/2" {
		t.Fatalf("journaled session target = %+v", got)
	}
}
