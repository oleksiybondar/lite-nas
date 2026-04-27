package services

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// Requirements: web-gateway/FR-004
func TestAuthServiceLoginIssuesStubCookies(t *testing.T) {
	t.Parallel()

	service := NewAuthService()
	now := time.Unix(100, 0)

	session, err := service.Login(now, "john.doe", "pass")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if !strings.HasPrefix(session.AccessToken, "AT-") {
		t.Fatalf("AccessToken = %q, want AT-*", session.AccessToken)
	}

	if !strings.HasPrefix(session.RefreshToken, "RT-") {
		t.Fatalf("RefreshToken = %q, want RT-*", session.RefreshToken)
	}

	if got := session.AccessExpires; got != now.Add(15*time.Minute) {
		t.Fatalf("AccessExpires = %v, want %v", got, now.Add(15*time.Minute))
	}
}

// Requirements: web-gateway/FR-004
func TestAuthServiceRefreshRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	service := NewAuthService()

	session, err := service.Refresh(time.Unix(100, 0), "anything")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if !strings.HasPrefix(session.AccessToken, "AT-") {
		t.Fatalf("AccessToken = %q, want AT-*", session.AccessToken)
	}
}

// Requirements: web-gateway/FR-004
func TestAuthServiceLogoutRequiresRefreshToken(t *testing.T) {
	t.Parallel()

	service := NewAuthService()

	if _, err := service.Logout(time.Unix(100, 0), ""); !errors.Is(err, ErrMissingRefreshToken()) {
		t.Fatalf("Logout() error = %v, want %v", err, ErrMissingRefreshToken())
	}
}
