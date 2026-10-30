package main

import (
	"context"
	"testing"
	"time"

	"lite-nas/services/auth/pamauth"
	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
)

// TestPasswordChangeStartRejectsMissingConversationManager checks the RPC boundary.
func TestPasswordChangeStartRejectsMissingConversationManager(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	response, err := handlePasswordChangeStartRPC(t.Context(), deps, rpcEnvelope(t, authcontract.PasswordChangeStartRequest{AccessToken: "invalid"}))
	if err != nil || response.Status != authcontract.StatusDenied {
		t.Fatalf("start response = %+v, error = %v", response, err)
	}
}

// TestPasswordChangeReplyRejectsUnknownConversation checks replay-safe lookup.
func TestPasswordChangeReplyRejectsUnknownConversation(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	deps.Conversations = newPasswordConversations(context.Background(), "missing-service", nil)
	response, err := handlePasswordChangeReplyRPC(t.Context(), deps, rpcEnvelope(t, authcontract.PasswordChangeReplyRequest{ConversationID: "missing", Sequence: 1, Answer: "answer"}))
	if err != nil || response.Status != authcontract.StatusDenied {
		t.Fatalf("reply response = %+v, error = %v", response, err)
	}
}

// TestPasswordConversationPromptCarriesSequence checks the browser prompt contract.
func TestPasswordConversationPromptCarriesSequence(t *testing.T) {
	flow := &passwordFlow{id: "pending", login: authcontract.LoginRequest{Username: "alice"}, expiresAt: time.Now().Add(time.Minute)}
	flow.sequence.Store(2)
	response := loginResponseForConversation(authRuntime{}, passwordEvent{prompt: &pamauth.Prompt{Text: "New password", Secret: true}}, flow)
	if response.Conversation == nil || response.Conversation.Sequence != 2 {
		t.Fatalf("conversation = %+v", response.Conversation)
	}
}

// TestOnDemandPasswordChangeDoesNotIssueSession protects token creation policy.
func TestOnDemandPasswordChangeDoesNotIssueSession(t *testing.T) {
	flow := &passwordFlow{onDemand: true}
	response := loginResponseForConversation(authRuntime{}, passwordEvent{result: pamauth.Result{Code: pamauth.OutcomeAuthenticated}}, flow)
	if response.AccessToken != "" {
		t.Fatalf("on-demand access token = %q", response.AccessToken)
	}
}

// TestCompletedPasswordConversationIssuesSession only after normal login.
func TestCompletedPasswordConversationIssuesSession(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	flow := &passwordFlow{login: authcontract.LoginRequest{Username: "alice", UserAgent: "test"}}
	response := loginResponseForConversation(deps, passwordEvent{result: pamauth.Result{Code: pamauth.OutcomeAuthenticated, Username: "alice"}}, flow)
	if response.AccessToken == "" {
		t.Fatal("completed login did not issue an access token")
	}
}

// TestPasswordChangeStartRejectsInvalidToken checks verified identity before PAM.
func TestPasswordChangeStartRejectsInvalidToken(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	deps.Conversations = newPasswordConversations(context.Background(), "missing-service", nil)
	response, err := handlePasswordChangeStartRPC(t.Context(), deps, rpcEnvelope(t, authcontract.PasswordChangeStartRequest{AccessToken: "invalid"}))
	if err != nil || response.Status != authcontract.StatusDenied {
		t.Fatalf("start response = %+v, error = %v", response, err)
	}
}

// TestPasswordChangeStartUsesVerifiedUser checks the on-demand PAM boundary.
func TestPasswordChangeStartUsesVerifiedUser(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	deps.Conversations = newPasswordConversations(context.Background(), "missing-lite-nas-pam-service", nil)
	token, _, err := deps.Tokens.Issuer.Issue(authtoken.Principal{Subject: "alice", Login: "alice"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	response, err := handlePasswordChangeStartRPC(t.Context(), deps, rpcEnvelope(t, authcontract.PasswordChangeStartRequest{AccessToken: token}))
	if err != nil || response.Status == authcontract.StatusAuthenticated {
		t.Fatalf("start response = %+v, error = %v", response, err)
	}
}

// TestPasswordChangeReplyDeliversAnswer checks the pending worker handoff.
func TestPasswordChangeReplyDeliversAnswer(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	flow := &passwordFlow{id: "pending", ctx: ctx, answers: make(chan string), events: make(chan passwordEvent), expiresAt: time.Now().Add(time.Minute), onDemand: true}
	flow.sequence.Store(1)
	deps.Conversations = &passwordConversations{flows: map[string]*passwordFlow{"pending": flow}}
	go func() {
		<-flow.answers
		flow.events <- passwordEvent{result: pamauth.Result{Code: pamauth.OutcomeDenied}}
	}()
	response, err := handlePasswordChangeReplyRPC(ctx, deps, rpcEnvelope(t, authcontract.PasswordChangeReplyRequest{ConversationID: "pending", Sequence: 1, Answer: "secret"}))
	if err != nil || response.Status != authcontract.StatusDenied {
		t.Fatalf("reply response = %+v, error = %v", response, err)
	}
}

// TestPasswordChangeStartRejectsLockdown checks emergency policy before PAM.
func TestPasswordChangeStartRejectsLockdown(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	deps.Conversations = newPasswordConversations(context.Background(), "missing-service", nil)
	deps.Lockdown = &lockdownStore{state: lockdownState{Lockdown: true}}
	response, err := handlePasswordChangeStartRPC(t.Context(), deps, rpcEnvelope(t, authcontract.PasswordChangeStartRequest{AccessToken: "invalid"}))
	if err != nil || response.Status != authcontract.StatusDenied {
		t.Fatalf("start response = %+v, error = %v", response, err)
	}
}

// TestCompletedConversationRejectsLockdown prevents session issuance during lockdown.
func TestCompletedConversationRejectsLockdown(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	deps.Lockdown = &lockdownStore{state: lockdownState{Lockdown: true}}
	flow := &passwordFlow{login: authcontract.LoginRequest{Username: "alice"}}
	response := loginResponseForConversation(deps, passwordEvent{result: pamauth.Result{Code: pamauth.OutcomeAuthenticated}}, flow)
	if response.Status != authcontract.StatusDenied {
		t.Fatalf("lockdown response = %+v", response)
	}
}

// TestFailedPasswordConversationDoesNotIssueTokens checks terminal PAM denial.
func TestFailedPasswordConversationDoesNotIssueTokens(t *testing.T) {
	response := loginResponseForConversation(authRuntime{}, passwordEvent{result: pamauth.Result{Code: pamauth.OutcomeDenied, Username: "alice"}}, &passwordFlow{})
	if response.AccessToken != "" {
		t.Fatalf("denied access token = %q", response.AccessToken)
	}
}

// TestLoginRPCUsesConversationWorker checks password policy on the login path.
func TestLoginRPCUsesConversationWorker(t *testing.T) {
	deps := authRuntimeFixture(t, pamauth.Result{})
	deps.Conversations = newPasswordConversations(context.Background(), "missing-lite-nas-pam-service", nil)
	response, err := handleLoginRPCWithContext(t.Context(), deps, rpcEnvelope(t, authcontract.LoginRequest{Username: "alice", Password: "bad"}))
	if err != nil || response.Status == authcontract.StatusAuthenticated {
		t.Fatalf("login response = %+v, error = %v", response, err)
	}
}

// TestPasswordConversationStartHandlesUnavailablePAM checks worker completion.
func TestPasswordConversationStartHandlesUnavailablePAM(t *testing.T) {
	manager := newPasswordConversations(context.Background(), "missing-lite-nas-pam-service", nil)
	event, _, accepted := manager.start(t.Context(), authcontract.LoginRequest{Username: "alice", Password: "bad"}, false)
	if !accepted || event.result.Code == pamauth.OutcomeAuthenticated {
		t.Fatalf("conversation accepted = %t, result = %+v", accepted, event.result)
	}
}
