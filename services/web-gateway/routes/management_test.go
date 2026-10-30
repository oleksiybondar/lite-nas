package routes

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"lite-nas/services/web-gateway/controllers"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
	webtest "lite-nas/services/web-gateway/testutil"
	"lite-nas/shared/authtoken"
	servicecontract "lite-nas/shared/contracts/servicesmanager"
	sessioncontract "lite-nas/shared/contracts/sessionmanager"
)

type managementRouteStub struct{ calls int }

func (s *managementRouteStub) List(context.Context, string) ([]servicecontract.Service, error) {
	s.calls++
	return []servicecontract.Service{{Name: "ssh.service"}}, nil
}

func (s *managementRouteStub) Action(context.Context, string, string, string) error {
	s.calls++
	return nil
}

type sessionRouteStub struct{ calls int }

func (s *sessionRouteStub) List(context.Context, string) ([]sessioncontract.Session, error) {
	s.calls++
	return []sessioncontract.Session{{ID: "3"}}, nil
}

func (s *sessionRouteStub) Action(context.Context, string, string, string) error {
	s.calls++
	return nil
}

func managementRouter(role string, services *managementRouteStub, sessions *sessionRouteStub) http.Handler {
	return NewRouter("gateway", "test", modules.Controllers{ServicesManagement: controllers.NewServicesManagementController(services), SessionsManagement: controllers.NewSessionsManagementController(sessions)}, middlewares.AuthenticationOptions{Verifier: routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{role}}}})
}

func managementRequest(handler http.Handler, method, path string, authenticated bool) int {
	request := webtest.NewRequest(method, path, nil)
	if authenticated {
		request.Header.Set("Authorization", "Bearer test-token")
	}
	return webtest.ServeRequest(handler, request).Code
}

// Requirements: web-gateway/FR-003, web-gateway/IR-001, services-manager/SR-001, session-manager/SR-001.
func TestManagementRoutes(t *testing.T) {
	cases := []struct {
		name, role, method, path string
		authenticated            bool
		want                     int
		calls                    int
	}{
		{"services list", "admin", "GET", "/api/services", true, 200, 1},
		{"services start", "admin", "POST", "/api/services/ssh.service/start", true, 200, 1},
		{"services stop", "admin", "POST", "/api/services/ssh.service/stop", true, 200, 1},
		{"services restart", "admin", "POST", "/api/services/ssh.service/restart", true, 200, 1},
		{"services enable", "admin", "POST", "/api/services/ssh.service/enable", true, 200, 1},
		{"services disable", "admin", "POST", "/api/services/ssh.service/disable", true, 200, 1},
		{"sessions list", "lite-nas-security", "GET", "/api/sessions", true, 200, 1},
		{"sessions terminate", "lite-nas-security", "POST", "/api/sessions/3/terminate", true, 200, 1},
		{"sessions lock", "lite-nas-security", "POST", "/api/sessions/3/lock", true, 200, 1},
		{"sessions unlock", "lite-nas-security", "POST", "/api/sessions/3/unlock", true, 200, 1},
		{"missing token", "admin", "GET", "/api/services", false, 401, 0},
		{"forbidden service", "lite-nas-security", "GET", "/api/services", true, 403, 0},
		{"forbidden session", "operator", "GET", "/api/sessions", true, 403, 0},
		{"invalid unit", "admin", "POST", "/api/services/-bad.service/start", true, 422, 0},
		{"missing unit suffix", "admin", "POST", "/api/services/sshXservice/start", true, 422, 0},
		{"invalid unit dollar", "admin", "POST", "/api/services/ssh%24.service/start", true, 422, 0},
		{"invalid session", "lite-nas-security", "POST", "/api/sessions/-bad/lock", true, 422, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			services := &managementRouteStub{}
			sessions := &sessionRouteStub{}
			handler := managementRouter(tc.role, services, sessions)
			got := managementRequest(handler, tc.method, tc.path, tc.authenticated)
			if got != tc.want {
				t.Fatalf("status=%d, want %d", got, tc.want)
			}
		})
	}
}

// Requirements: web-gateway/IR-001.
func TestManagementOpenAPIDocumentsRoutes(t *testing.T) {
	handler := managementRouter("admin", &managementRouteStub{}, &sessionRouteStub{})
	recorder := webtest.ServeRequest(handler, webtest.NewRequest("GET", "/api/openapi.json", nil))
	for _, path := range []string{"/services", "/sessions", "/service-metrics"} {
		if !strings.Contains(recorder.Body.String(), path) {
			t.Errorf("OpenAPI missing %s", path)
		}
	}
}
