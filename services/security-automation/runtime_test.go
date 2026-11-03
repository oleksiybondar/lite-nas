package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/authtoken"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/messagingtest"
)

// pipelineServer exposes registered handlers after both alert subscriptions exist.
type pipelineServer struct {
	*messagingtest.RecordingServer
	handlers map[string]messaging.MessageHandler
	ready    chan struct{}
}

// Subscribe records a real callback for the end-to-end worker test.
func (server *pipelineServer) Subscribe(subject string, handler messaging.MessageHandler) error {
	server.handlers[subject] = handler
	if len(server.handlers) == 2 {
		close(server.ready)
	}
	return nil
}

// matchLogger captures an observed FSM completion without publishing an unverified effect.
type matchLogger struct {
	logger.NopLogger
	matched chan struct{}
}

// Warn signals the match reporter while ignoring expected startup and shutdown warnings.
func (log matchLogger) Warn(message string, _ ...any) {
	if message == "automation match requires enrollment before effect" {
		select {
		case log.matched <- struct{}{}:
		default:
		}
	}
}

// Requirements: security-automation/IR-002.
// TestAutomationWorkersDrainAfterCancellation checks startup registration and channel owner shutdown.
func TestAutomationWorkersDrainAfterCancellation(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	server := &messagingtest.RecordingServer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = runAutomationWorkers(ctx, server, &messagingtest.RecordingClient{}, authtoken.Verifier{}, logger.NopLogger{}, rules, nil, nil, nil, nil, usbEffects{})
	if !errors.Is(err, context.Canceled) || server.DrainCalls != 1 {
		t.Fatalf("shutdown error = %v; drains = %d", err, server.DrainCalls)
	}
}

// Requirements: security-automation/IR-003.
// TestAutomationWorkersStopOnRPCRegistrationFailure checks startup never accepts events after an RPC registration failure.
func TestAutomationWorkersStopOnRPCRegistrationFailure(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	server := &messagingtest.RecordingServer{RegisterErr: errors.New("unavailable")}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := runAutomationWorkers(ctx, server, &messagingtest.RecordingClient{}, authtoken.Verifier{}, logger.NopLogger{}, rules, nil, nil, nil, nil, usbEffects{}); err == nil {
		t.Fatal("registration failure ignored")
	}
}

// Requirements: security-automation/IR-002.
// TestAutomationWorkersDetectSubscribedAlert checks subscription, listener, FSM, and outbound logging as one flow.
func TestAutomationWorkersDetectSubscribedAlert(t *testing.T) {
	path := mustRuleCopy(t, `"id": "ssh_brute_force", "name": "SSH brute force", "description": "Aggregated SSH failures from one source address.", "enabled": false`, `"id": "ssh_brute_force", "name": "SSH brute force", "description": "Aggregated SSH failures from one source address.", "enabled": true`)
	rules, err := loadAutomationRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
		t.Fatal(err)
	}
	server := &pipelineServer{RecordingServer: &messagingtest.RecordingServer{}, handlers: make(map[string]messaging.MessageHandler), ready: make(chan struct{})}
	log := matchLogger{matched: make(chan struct{}, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		finished <- runAutomationWorkers(ctx, server, &messagingtest.RecordingClient{}, authtoken.Verifier{}, log, rules, nil, nil, nil, nil, usbEffects{})
	}()
	select {
	case <-server.ready:
	case <-time.After(time.Second):
		t.Fatal("alert subscriptions were not registered")
	}
	payload := []byte(`{"access_token":"token","event_id":"secssh_1","category":"password_guessing","source":"log-analysis","message":"Security rule triggered: ssh_preauth_probing from 192.0.2.1"}`)
	if err := server.handlers[securitycontract.AlertSubject](ctx, messaging.Envelope{Payload: payload}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-log.matched:
	case <-time.After(time.Second):
		t.Fatal("subscribed alert did not complete the FSM")
	}
	cancel()
	select {
	case err := <-finished:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("worker shutdown = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker shutdown did not complete")
	}
}

// TestAutomationWorkersPublishCorrelatedSequence checks two validated source
// alerts complete a sequence and reach the outbound alert action.
// Requirements: security-automation/FR-002, security-automation/FR-007.
func TestAutomationWorkersPublishCorrelatedSequence(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	for id, rule := range rules {
		if id != "ssh_probe_then_root_login" {
			rule.enabled.Store(false)
		}
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
		t.Fatal(err)
	}
	server := &pipelineServer{RecordingServer: &messagingtest.RecordingServer{}, handlers: make(map[string]messaging.MessageHandler), ready: make(chan struct{})}
	reported := make(chan detection.Match, 1)
	report := func(_ context.Context, _ automationRuleSpec, match detection.Match) error {
		reported <- match
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() {
		finished <- runAutomationWorkers(ctx, server, &messagingtest.RecordingClient{}, authtoken.Verifier{}, logger.NopLogger{}, rules, nil, nil, report, nil, usbEffects{})
	}()
	select {
	case <-server.ready:
	case <-time.After(time.Second):
		t.Fatal("alert subscriptions were not registered")
	}
	for _, payload := range [][]byte{
		[]byte(`{"access_token":"token","event_id":"secssh_1","category":"password_guessing","source":"log-analysis","message":"Security rule triggered: ssh_preauth_probing from 198.51.100.10"}`),
		[]byte(`{"access_token":"token","event_id":"secssh_2","category":"privilege_escalation","source":"log-analysis","message":"Security rule triggered: ssh_root_login from 198.51.100.10"}`),
	} {
		if err := server.handlers[securitycontract.AlertSubject](ctx, messaging.Envelope{Payload: payload}); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case match := <-reported:
		if match.Key != "198.51.100.10" || len(match.AlertIDs) != 2 {
			t.Fatalf("published match = %+v", match)
		}
	case <-time.After(time.Second):
		t.Fatal("correlated sequence was not published")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("worker shutdown did not complete")
	}
}
