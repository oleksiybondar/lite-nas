package routes

import (
	"context"
	"net/http"
	"testing"

	"lite-nas/services/web-gateway/controllers"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
	webtest "lite-nas/services/web-gateway/testutil"
	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
)

type authSessionsRouteStub struct{}

func (s *authSessionsRouteStub) List(context.Context, string) ([]authcontract.SessionSummary, error) {
	return []authcontract.SessionSummary{{SessionID: "session_1", Login: "admin", Subject: "1000"}}, nil
}

func (s *authSessionsRouteStub) Revoke(context.Context, string, string) (bool, error) {
	return true, nil
}

// TestAuthSessionsRoutes verifies administrator access and path validation.
func TestAuthSessionsRoutes(t *testing.T) {
	cases := []struct {
		name, role, method, path string
		want                     int
	}{
		{"list as admin", "admin", "GET", "/api/auth/sessions", 200},
		{"revoke as admin", "admin", "POST", "/api/auth/sessions/session_1/revoke", 200},
		{"security officer forbidden", "lite-nas-security", "GET", "/api/auth/sessions", 403},
		{"security officer cannot revoke", "lite-nas-security", "POST", "/api/auth/sessions/session_1/revoke", 403},
		{"invalid session id", "admin", "POST", "/api/auth/sessions/%24bad/revoke", 422},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &authSessionsRouteStub{}
			handler := NewRouter("gateway", "test", modules.Controllers{AuthSessions: controllers.NewAuthSessionsController(stub)}, middlewares.AuthenticationOptions{Verifier: routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{tc.role}}}})
			request := webtest.NewRequest(tc.method, tc.path, nil)
			request.Header.Set("Authorization", "Bearer test-token")
			got := webtest.ServeRequest(handler, request).Code
			if got != tc.want {
				t.Fatalf("status=%d, want %d", got, tc.want)
			}
		})
	}
}

// TestAuthSessionsRouteRequiresAuthentication verifies missing credentials are rejected.
func TestAuthSessionsRouteRequiresAuthentication(t *testing.T) {
	stub := &authSessionsRouteStub{}
	handler := NewRouter("gateway", "test", modules.Controllers{AuthSessions: controllers.NewAuthSessionsController(stub)}, middlewares.AuthenticationOptions{Verifier: routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{"admin"}}}})
	got := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/auth/sessions", nil)).Code
	if got != http.StatusUnauthorized {
		t.Fatalf("status=%d", got)
	}
}
