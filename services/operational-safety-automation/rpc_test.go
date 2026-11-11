package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"lite-nas/shared/authtoken"
	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/messagingtest"
)

// mustSafetyOperator creates a live-token-compatible RPC fixture.
func mustSafetyOperator(t *testing.T, roles []string) (*safetyRuntime, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := authtoken.NewIssuer(authtoken.IssuerOptions{Issuer: "test", Audience: "test", AccessLifetime: time.Minute}, private)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: "test", Audience: "test"}, public)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := issuer.Issue(authtoken.Principal{Subject: "123", Login: "alice", Roles: roles})
	if err != nil {
		t.Fatal(err)
	}
	return &safetyRuntime{rules: map[string]*compiledRule{}, order: []string{}, tasks: map[string]*safetyTask{}, verifier: verifier, client: &messagingtest.RecordingClient{AuthUsername: "alice"}, log: logger.NopLogger{}}, token
}

// mustSafetyEnvelope encodes one RPC request for handler tests.
func mustSafetyEnvelope(t *testing.T, value any) messaging.Envelope {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.Envelope{Payload: data}
}

// TestOperationalSafetyRegistersOperatorSubjects checks every gateway RPC is registered.
func TestOperationalSafetyRegistersOperatorSubjects(t *testing.T) {
	server := &messagingtest.RecordingServer{}
	if err := new(safetyRuntime).register(server); err != nil {
		t.Fatal(err)
	}
	for _, subject := range []string{contract.RulesListSubject, contract.TasksListSubject, contract.RuleSetEnabledSubject, contract.TaskForceSetSubject, contract.TaskForceStopSubject, contract.CPUSetCapSubject} {
		t.Run(subject, func(t *testing.T) {
			if server.Handlers[subject] == nil {
				t.Fatal("RPC handler missing")
			}
		})
	}
}

// TestOperationalSafetyManualCapRequiresFreshRuntime checks that active task state cannot be bypassed.
func TestOperationalSafetyManualCapRequiresFreshRuntime(t *testing.T) {
	runtime, token := mustSafetyOperator(t, []string{"lite-nas-security"})
	runtime.tasks["known"] = &safetyTask{}
	response, err := runtime.setCPUCap(context.Background(), mustSafetyEnvelope(t, contract.OperationalSafetyCPUSetCapRequest{AccessToken: token, CapKHz: 1900000}))
	if err != nil {
		t.Fatal(err)
	}
	if got := response.(contract.OperationalSafetyCPUSetCapResponse).State; got != contract.OutcomeBlocked {
		t.Fatalf("manual cap state = %q", got)
	}
}

// TestOperationalSafetyManualCapUsesVerifiedEffect checks the RPC-to-worker path.
func TestOperationalSafetyManualCapUsesVerifiedEffect(t *testing.T) {
	runtime, token := mustSafetyOperator(t, []string{"lite-nas-security"})
	effects := make(chan effectIntent, 1)
	runtime.intents = effects
	runtime.applyManual = func(_ context.Context, cap uint64) error {
		if cap != 1900000 {
			return errors.New("unexpected cap")
		}
		return nil
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for intent := range effects {
			runtime.applyIntent(context.Background(), intent)
		}
	}()
	t.Cleanup(func() { close(effects); <-done })
	response, err := runtime.setCPUCap(context.Background(), mustSafetyEnvelope(t, contract.OperationalSafetyCPUSetCapRequest{AccessToken: token, CapKHz: 1900000}))
	if err != nil {
		t.Fatal(err)
	}
	if got := response.(contract.OperationalSafetyCPUSetCapResponse); got.State != contract.OutcomeCompleted || got.CapKHz != 1900000 || runtime.lastCap != 1900000 {
		t.Fatalf("manual cap response = %#v, last cap = %d", got, runtime.lastCap)
	}
}

// TestOperationalSafetyRPCRejectsMalformedInput checks strict boundary decoding.
func TestOperationalSafetyRPCRejectsMalformedInput(t *testing.T) {
	runtime, _ := mustSafetyOperator(t, []string{"lite-nas-security"})
	for _, payload := range []string{"", "{}", `{"access_token":"x","extra":1}`, `{"access_token":"x"} {}`} {
		t.Run(payload, func(t *testing.T) {
			if _, err := runtime.listRules(context.Background(), messaging.Envelope{Payload: []byte(payload)}); err == nil {
				t.Fatal("malformed RPC accepted")
			}
		})
	}
}

// TestOperationalSafetyRuleStateRequiresEnabledField checks direct NATS callers cannot omit the desired state.
func TestOperationalSafetyRuleStateRequiresEnabledField(t *testing.T) {
	runtime, token := mustSafetyOperator(t, []string{"lite-nas-security"})
	envelope := mustSafetyEnvelope(t, map[string]any{"access_token": token, "rule_id": "power"})
	if _, err := runtime.setEnabled(context.Background(), envelope); err == nil {
		t.Fatal("missing enabled value accepted")
	}
}

// TestOperationalSafetyRPCRequiresSecurityRole checks live-token role enforcement.
func TestOperationalSafetyRPCRequiresSecurityRole(t *testing.T) {
	runtime, token := mustSafetyOperator(t, []string{"lite-nas-operator"})
	envelope := mustSafetyEnvelope(t, contract.OperationalSafetyListRequest{AccessToken: token})
	if _, err := runtime.listRules(context.Background(), envelope); err == nil {
		t.Fatal("operator-only token accepted")
	}
}

// TestOperationalSafetyRuleListAcceptsSecurityRole checks authorized read flow.
func TestOperationalSafetyRuleListAcceptsSecurityRole(t *testing.T) {
	runtime, token := mustSafetyOperator(t, []string{"lite-nas-security"})
	envelope := mustSafetyEnvelope(t, contract.OperationalSafetyListRequest{AccessToken: token})
	response, err := runtime.listRules(context.Background(), envelope)
	if err != nil {
		t.Fatal(err)
	}
	got := response.(contract.OperationalSafetyRulesResponse)
	if len(got.Rules) != 0 {
		t.Fatalf("rules = %#v", got.Rules)
	}
}
