package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/browserauth"
)

// authStub supplies the shared auth interface for router boundary tests.
type authStub struct{ session browserauth.Session }

// Login is unused by the protected route test.
func (stub authStub) Login(context.Context, time.Time, string, string, browserauth.AuthRequestContext) (browserauth.Session, error) {
	return stub.session, nil
}

// Refresh is unused by the protected route test.
func (authStub) Refresh(context.Context, time.Time, string, browserauth.AuthRequestContext) (browserauth.Session, error) {
	return browserauth.Session{}, nil
}

// Logout is unused by the protected route test.
func (authStub) Logout(context.Context, time.Time, string, browserauth.AuthRequestContext) (browserauth.Session, error) {
	return browserauth.Session{}, nil
}

// Me rejects missing access credentials for the protected route test.
func (authStub) Me(time.Time, string) (browserauth.Session, error) {
	return browserauth.Session{}, browserauth.ErrUnauthorized
}

// Requirements: moments/FR-001
func TestMeRequiresAuthentication(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	response := httptest.NewRecorder()
	newRouter(authStub{}, "").ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

// Requirements: moments/FR-002
func TestAuthMutationRejectsForeignOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.Header.Set("Origin", "https://foreign.example")
	response := httptest.NewRecorder()
	newRouter(authStub{}, "").ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

// Requirements: moments/FR-001, browserauth/FR-003
func TestLoginUsesConfiguredCookieDomain(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"login":"alice","password":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	session := browserauth.Session{UserID: "alice", AccessToken: "access", RefreshToken: "refresh", AccessExpires: time.Now().Add(time.Hour), RefreshExpiry: time.Now().Add(24 * time.Hour)}
	newRouter(authStub{session: session}, "example.com").ServeHTTP(response, request)
	if !strings.Contains(response.Header().Get("Set-Cookie"), "Domain=example.com") {
		t.Fatalf("Set-Cookie lacks configured domain: %q", response.Header().Get("Set-Cookie"))
	}
}

// Requirements: moments/FR-001
func TestLoginReturnsGatewayCompatibleSessionData(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"login":"alice","password":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(authStub{session: browserauth.Session{UserID: "alice"}}, "").ServeHTTP(response, request)
	var body struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.User.ID != "alice" {
		t.Fatalf("user ID = %q, want alice", body.Data.User.ID)
	}
}

// Requirements: moments/IR-002
func TestListenerRejectsExternalBind(t *testing.T) {
	if err := validateListener("0.0.0.0:9091"); err == nil {
		t.Fatal("expected external listener to be rejected")
	}
}
