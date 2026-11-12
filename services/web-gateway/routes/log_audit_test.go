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
	contract "lite-nas/shared/contracts/loganalysis"
)

// logAuditRouteStub provides the catalog expected by the protected route.
type logAuditRouteStub struct{}

// Rules returns one configured rule after the HTTP authorization middleware runs.
func (logAuditRouteStub) Rules(context.Context, string) (contract.LogAnalysisRulesResponse, error) {
	return contract.LogAnalysisRulesResponse{Rules: []contract.LogAnalysisRule{{ID: "sample"}}}, nil
}

// TestLogAuditRulesRouteAllowsSecurityRole verifies the requested path is mounted.
func TestLogAuditRulesRouteAllowsSecurityRole(t *testing.T) {
	handler := NewRouter("gateway", "test", modules.Controllers{LogAudit: controllers.NewLogAuditController(logAuditRouteStub{})}, middlewares.AuthenticationOptions{Verifier: routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{"lite-nas-security"}}}})
	request := webtest.NewRequest(http.MethodGet, "/api/security/log-audit/rules", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	if got := webtest.ServeRequest(handler, request).Code; got != http.StatusOK {
		t.Fatalf("status=%d", got)
	}
}

// TestLogAuditRulesRouteRejectsOperatorRole verifies the security-only HTTP boundary.
func TestLogAuditRulesRouteRejectsOperatorRole(t *testing.T) {
	handler := NewRouter("gateway", "test", modules.Controllers{LogAudit: controllers.NewLogAuditController(logAuditRouteStub{})}, middlewares.AuthenticationOptions{Verifier: routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{"lite-nas-operator"}}}})
	request := webtest.NewRequest(http.MethodGet, "/api/security/log-audit/rules", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	if got := webtest.ServeRequest(handler, request).Code; got != http.StatusForbidden {
		t.Fatalf("status=%d", got)
	}
}
