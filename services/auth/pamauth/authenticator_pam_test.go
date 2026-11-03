package pamauth

import (
	"errors"
	"testing"

	pam "github.com/msteinert/pam/v2"
)

// TestNewAuthenticatorRejectsMissingService guards the PAM service boundary.
func TestNewAuthenticatorRejectsMissingService(t *testing.T) {
	if _, err := NewAuthenticator(" "); !errors.Is(err, errMissingServiceName) {
		t.Fatalf("service error = %v", err)
	}
}

// TestNewAuthenticatorAcceptsService binds authentication to the chosen stack.
func TestNewAuthenticatorAcceptsService(t *testing.T) {
	auth, err := NewAuthenticator("litenas-auth")
	if err != nil || auth == nil {
		t.Fatalf("authenticator = %v, error = %v", auth, err)
	}
}

// TestPAMConversationUsesCredentials checks prompt response routing.
func TestPAMConversationUsesCredentials(t *testing.T) {
	conversation := newAuthenticateConversation(AuthenticateRequest{Username: "alice", Password: "secret"})
	response, err := conversation.respond(pam.PromptEchoOff, "Password:")
	if err != nil || response != "secret" {
		t.Fatalf("response = %q, error = %v", response, err)
	}
}

// TestPAMConversationRecordsMessages checks PAM feedback classification.
func TestPAMConversationRecordsMessages(t *testing.T) {
	for _, test := range []struct {
		name  string
		style pam.Style
		level MessageLevel
	}{
		{"information", pam.TextInfo, MessageLevelInfo},
		{"error", pam.ErrorMsg, MessageLevelError},
	} {
		t.Run(test.name, func(t *testing.T) {
			conversation := newAuthenticateConversation(AuthenticateRequest{Username: "alice"})
			_, _ = conversation.respond(test.style, "message")
			if len(conversation.messages) != 1 || conversation.messages[0].Level != test.level {
				t.Fatalf("messages = %+v", conversation.messages)
			}
		})
	}
}

// TestPAMConversationReturnsUsernameForVisiblePrompt checks PAM identity input.
func TestPAMConversationReturnsUsernameForVisiblePrompt(t *testing.T) {
	conversation := newAuthenticateConversation(AuthenticateRequest{Username: "alice"})
	response, err := conversation.respond(pam.PromptEchoOn, "Username:")
	if err != nil || response != "alice" {
		t.Fatalf("response = %q, error = %v", response, err)
	}
}

// TestPAMAuthErrorIsDenied checks a failed PAM password result.
func TestPAMAuthErrorIsDenied(t *testing.T) {
	result := newAuthenticateConversation(AuthenticateRequest{}).resultForAuthError("alice", pam.ErrAuth)
	if result.Code != OutcomeDenied {
		t.Fatalf("result = %+v", result)
	}
}

// TestPAMConversationRejectsUnavailableService checks a failed host stack.
func TestPAMConversationRejectsUnavailableService(t *testing.T) {
	result, err := RunConversation("missing-lite-nas-pam-service", AuthenticateRequest{Username: "alice", Password: "bad"}, false, func(Prompt) (string, error) { return "", nil }, func(string) {})
	if err == nil && result.Code == OutcomeAuthenticated {
		t.Fatalf("unexpected authentication result = %+v", result)
	}
}

// TestPAMAuthenticatorDeniesUnknownService checks direct authentication failure.
func TestPAMAuthenticatorDeniesUnknownService(t *testing.T) {
	auth, err := NewAuthenticator("missing-lite-nas-pam-service")
	if err != nil {
		t.Fatalf("construct authenticator: %v", err)
	}
	result, err := auth.Authenticate(AuthenticateRequest{Username: "alice", Password: "bad"})
	if err == nil && result.Code == OutcomeAuthenticated {
		t.Fatalf("unexpected authentication result = %+v", result)
	}
}

// TestAccountErrorsOnlyRequirePasswordChangeForThePAMExpiryStatus keeps
// unrelated account failures from being offered a password update.
func TestAccountErrorsOnlyRequirePasswordChangeForThePAMExpiryStatus(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want OutcomeCode
	}{
		{name: "password expired", err: pam.ErrNewAuthtokReqd, want: OutcomePasswordChangeNeeded},
		{name: "account expired", err: pam.ErrAcctExpired, want: OutcomeDenied},
		{name: "permission denied", err: pam.ErrPermDenied, want: OutcomeDenied},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := newAuthenticateConversation(AuthenticateRequest{}).resultForAccountError("alice", test.err)
			if result.Code != test.want {
				t.Fatalf("Code = %q, want %q", result.Code, test.want)
			}
		})
	}
}
