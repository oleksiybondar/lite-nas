package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/messagingtest"
)

// tokenValidationClient returns an Auth Service validation result for cache tests.
type tokenValidationClient struct {
	messagingtest.RecordingClient
	valid bool
}

// loginClient returns a user access token and unused refresh token.
type loginClient struct {
	messagingtest.RecordingClient
	calls int
}

// conversationClient models Auth Service's additional PAM prompt.
type conversationClient struct {
	messagingtest.RecordingClient
	calls []string
}

// Request advances the login conversation and returns the final access token.
func (client *conversationClient) Request(_ context.Context, subject string, _ any, response any) error {
	client.calls = append(client.calls, subject)
	result := response.(*authcontract.LoginResponse)
	if subject == authcontract.LoginRPCSubject {
		*result = authcontract.LoginResponse{Status: authcontract.StatusPasswordChangeRequired, Conversation: &authcontract.PasswordConversation{ID: "id", Sequence: 1, Prompt: "New password", Secret: true}}
	} else if subject == authcontract.PasswordChangeReplyRPCSubject {
		*result = authcontract.LoginResponse{Status: authcontract.StatusAuthenticated, AccessToken: "access"}
	} else {
		return os.ErrInvalid
	}
	return nil
}

// Request records the login RPC without exposing a real credential.
func (client *loginClient) Request(_ context.Context, subject string, _ any, response any) error {
	if subject != authcontract.LoginRPCSubject {
		return os.ErrInvalid
	}
	client.calls++
	*response.(*authcontract.LoginResponse) = authcontract.LoginResponse{Status: authcontract.StatusAuthenticated, AccessToken: "access", RefreshToken: "refresh"}
	return nil
}

// Request supplies the validation outcome without accepting unrelated RPCs.
func (client *tokenValidationClient) Request(_ context.Context, subject string, _ any, response any) error {
	if subject != authcontract.ValidateAccessTokenRPCSubject {
		return os.ErrInvalid
	}
	*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{Valid: client.valid, Status: authcontract.StatusAuthenticated}
	return nil
}

// Requirements: backup-restore/SR-003.
func TestCachedTokenUsesOnlyValidAccessToken(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		valid bool
		want  string
	}{{"valid", true, "token"}, {"expired", false, ""}} {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token")
			if err := os.WriteFile(path, []byte("token"), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := cachedToken(context.Background(), &tokenValidationClient{valid: testCase.valid}, path)
			if err != nil || got != testCase.want {
				t.Fatalf("cached token = %q, %v", got, err)
			}
		})
	}
}

// Requirements: backup-restore/SR-003.
func TestWriteTokenRestrictsCachePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	if err := writeToken(path, "token"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("cache permissions = %o", info.Mode().Perm())
	}
}

// Requirements: backup-restore/SR-003.
func TestLoginKeepsOnlyAccessToken(t *testing.T) {
	client := &loginClient{}
	answers := []string{"admin", "password"}
	index := 0
	got, err := loginWithPrompt(context.Background(), client, func(_ string, _ bool) (string, error) { answer := answers[index]; index++; return answer, nil })
	if err != nil || got != "access" || client.calls != 1 {
		t.Fatalf("login = %q, calls = %d, error = %v", got, client.calls, err)
	}
}

// Requirements: backup-restore/SR-003.
func TestExpiredTokenPromptsAndReplacesCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(path, []byte("expired"), 0o600); err != nil {
		t.Fatal(err)
	}
	client := &tokenValidationClient{valid: false}
	got, err := accessTokenAtPath(context.Background(), client, path, func(context.Context, messaging.Client) (string, error) { return "fresh", nil })
	if err != nil {
		t.Fatal(err)
	}
	stored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != "fresh" || string(stored) != "fresh" {
		t.Fatalf("access token = %q, cache = %q", got, stored)
	}
}

// Requirements: backup-restore/SR-003.
func TestLoginCompletesAdditionalPAMPrompt(t *testing.T) {
	client := &conversationClient{}
	answers := []string{"admin", "password", "new-password"}
	index := 0
	got, err := loginWithPrompt(context.Background(), client, func(_ string, _ bool) (string, error) { answer := answers[index]; index++; return answer, nil })
	if err != nil || got != "access" || len(client.calls) != 2 {
		t.Fatalf("PAM login = %q, calls = %v, error = %v", got, client.calls, err)
	}
}

// Requirements: backup-restore/SR-003.
func TestTokenPathCreatesPrivateUserCache(t *testing.T) {
	runtimeDir := t.TempDir()
	if err := os.Chmod(runtimeDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path, err := tokenPathInRuntime(runtimeDir, os.Geteuid())
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("cache directory mode = %o", info.Mode().Perm())
	}
}

// Requirements: backup-restore/SR-003.
func TestTokenPathRejectsSymlinkRuntimeDirectory(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "runtime")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	_, err := tokenPathInRuntime(link, os.Geteuid())
	if err == nil {
		t.Fatal("symlink runtime directory accepted")
	}
}
