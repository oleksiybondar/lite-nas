package routes

import (
	"net/http"
	"testing"

	webtest "lite-nas/services/web-gateway/testutil"
	"lite-nas/shared/authtoken"
)

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyRoutesRequireAuthentication(t *testing.T) {
	for _, path := range []string{"/api/system/automations/rules", "/api/system/automations/tasks"} {
		t.Run(path, func(t *testing.T) {
			response := webtest.ServeRequest(routerFixture(nil), webtest.NewRequest(http.MethodGet, path, nil))
			webtest.AssertStatus(t, response, http.StatusUnauthorized)
		})
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyRoutesRequireSecurityRole(t *testing.T) {
	router := routerFixtureWithVerifier(nil, routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{"viewer"}}})
	assertAuthenticatedRouteStatus(t, router, http.MethodGet, "/api/system/automations/tasks", nil, http.StatusForbidden)
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyReadRoutesReachService(t *testing.T) {
	for _, path := range []string{"/api/system/automations/rules", "/api/system/automations/tasks"} {
		t.Run(path, func(t *testing.T) {
			assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodGet, path, nil, http.StatusOK)
		})
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyForceSetRejectsUnknownLevel(t *testing.T) {
	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodPost, "/api/system/automations/tasks/abc123/force-set", []byte(`{"level":"unsafe"}`), http.StatusUnprocessableEntity)
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyRuleStateRequiresEnabledValue(t *testing.T) {
	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodPut, "/api/system/automations/rules/power/enabled", []byte(`{}`), http.StatusUnprocessableEntity)
}
