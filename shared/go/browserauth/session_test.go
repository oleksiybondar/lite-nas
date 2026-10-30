package browserauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
)

// authRPCStub supplies the auth lifecycle replies without a running NATS server.
type authRPCStub struct {
	login  authcontract.LoginResponse
	rotate authcontract.RefreshResponse
	logout authcontract.LogoutResponse
	err    error
}

// Publish is unused by the browser auth service.
func (authRPCStub) Publish(context.Context, string, any) error { return nil }

// Request returns the configured reply for the requested auth operation.
func (stub authRPCStub) Request(_ context.Context, subject string, _ any, response any) error {
	if stub.err != nil {
		return stub.err
	}
	switch subject {
	case authcontract.LoginRPCSubject:
		*response.(*authcontract.LoginResponse) = stub.login
	case authcontract.RefreshRPCSubject:
		*response.(*authcontract.RefreshResponse) = stub.rotate
	case authcontract.LogoutRPCSubject:
		*response.(*authcontract.LogoutResponse) = stub.logout
	}
	return nil
}

// Drain is unused by the browser auth service.
func (authRPCStub) Drain() error { return nil }

// Close is unused by the browser auth service.
func (authRPCStub) Close() {}

// claimsVerifier returns one already verified identity for lifecycle tests.
type claimsVerifier struct{ claims authtoken.AccessClaims }

// Verify maps the test token to its configured claims.
func (verifier claimsVerifier) Verify(string) (authtoken.AccessClaims, error) {
	return verifier.claims, nil
}

// mustBrowserService builds a local auth service with one valid access identity.
func mustBrowserService(stub authRPCStub) AuthService {
	return NewAuthService(stub, claimsVerifier{claims: authtoken.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "alice",
			ExpiresAt: jwt.NewNumericDate(time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)),
		},
		Login: "alice",
	}})
}

// TestLoginRejectsMissingCredentials covers browserauth/FR-001.
func TestLoginRejectsMissingCredentials(t *testing.T) {
	service := mustBrowserService(authRPCStub{})
	_, err := service.Login(t.Context(), time.Now(), " ", "secret", AuthRequestContext{})
	if !errors.Is(err, ErrMissingCredentials()) {
		t.Fatalf("login error = %v, want missing credentials", err)
	}
}

// TestLoginMapsIssuedSession covers browserauth/FR-001.
func TestLoginMapsIssuedSession(t *testing.T) {
	service := mustBrowserService(authRPCStub{login: authcontract.LoginResponse{
		Status: authcontract.StatusAuthenticated, AccessToken: "access", RefreshToken: "refresh",
	}})
	session, err := service.Login(t.Context(), time.Now(), "alice", "secret", AuthRequestContext{})
	if err != nil {
		t.Fatal(err)
	}
	if session.UserID != "alice" || session.RefreshToken != "refresh" {
		t.Fatalf("login session = %+v", session)
	}
}

// TestLoginRejectsDeniedReply covers browserauth/FR-001.
func TestLoginRejectsDeniedReply(t *testing.T) {
	service := mustBrowserService(authRPCStub{login: authcontract.LoginResponse{Status: authcontract.StatusDenied}})
	_, err := service.Login(t.Context(), time.Now(), "alice", "secret", AuthRequestContext{})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("login error = %v, want unauthorized", err)
	}
}

// TestBeginLoginPreservesPasswordPrompt exposes expiry without a browser session.
func TestBeginLoginPreservesPasswordPrompt(t *testing.T) {
	service := mustBrowserService(authRPCStub{login: authcontract.LoginResponse{
		Status:       authcontract.StatusPasswordChangeRequired,
		Conversation: &authcontract.PasswordConversation{ID: "pending", Sequence: 1},
	}})
	result, err := service.BeginLogin(t.Context(), time.Now(), "alice", "secret", AuthRequestContext{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Conversation == nil {
		t.Fatal("conversation is nil, want pending prompt")
	}
}

// TestRefreshRotatesSession covers browserauth/FR-001.
func TestRefreshRotatesSession(t *testing.T) {
	service := mustBrowserService(authRPCStub{rotate: authcontract.RefreshResponse{
		AccessToken: "new-access", RefreshToken: "new-refresh",
	}})
	session, err := service.Refresh(t.Context(), time.Now(), "old-refresh", AuthRequestContext{})
	if err != nil {
		t.Fatal(err)
	}
	if session.RefreshToken != "new-refresh" {
		t.Fatalf("refresh token = %q", session.RefreshToken)
	}
}

// TestRefreshRejectsMissingToken covers browserauth/FR-001.
func TestRefreshRejectsMissingToken(t *testing.T) {
	service := mustBrowserService(authRPCStub{})
	_, err := service.Refresh(t.Context(), time.Now(), " ", AuthRequestContext{})
	if !errors.Is(err, ErrMissingRefreshToken()) {
		t.Fatalf("refresh error = %v, want missing token", err)
	}
}

// TestLogoutExpiresRevokedSession covers browserauth/FR-001.
func TestLogoutExpiresRevokedSession(t *testing.T) {
	service := mustBrowserService(authRPCStub{logout: authcontract.LogoutResponse{LoggedOut: true}})
	now := time.Now()
	session, err := service.Logout(t.Context(), now, "refresh", AuthRequestContext{})
	if err != nil {
		t.Fatal(err)
	}
	if !session.RefreshExpiry.Equal(now) {
		t.Fatalf("refresh expiry = %v, want %v", session.RefreshExpiry, now)
	}
}

// TestLogoutRejectsUnrevokedSession covers browserauth/FR-001.
func TestLogoutRejectsUnrevokedSession(t *testing.T) {
	service := mustBrowserService(authRPCStub{})
	_, err := service.Logout(t.Context(), time.Now(), "refresh", AuthRequestContext{})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("logout error = %v, want unauthorized", err)
	}
}

// TestMeReturnsVerifiedIdentity covers browserauth/FR-001.
func TestMeReturnsVerifiedIdentity(t *testing.T) {
	service := mustBrowserService(authRPCStub{})
	session, err := service.Me(time.Now(), "access")
	if err != nil {
		t.Fatal(err)
	}
	if session.Login != "alice" {
		t.Fatalf("current user = %q, want alice", session.Login)
	}
}

// TestMeRejectsMissingAccessToken covers browserauth/FR-001.
func TestMeRejectsMissingAccessToken(t *testing.T) {
	service := mustBrowserService(authRPCStub{})
	_, err := service.Me(time.Now(), " ")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("current-user error = %v, want unauthorized", err)
	}
}
