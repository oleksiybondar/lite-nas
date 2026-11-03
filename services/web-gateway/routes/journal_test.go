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
	contract "lite-nas/shared/contracts/journalmanager"
)

// journalRouteStub records calls made after the gateway authorization boundary.
type journalRouteStub struct{ calls int }

// Query returns an empty chunk for route authorization tests.
func (s *journalRouteStub) Query(context.Context, contract.QueryRequest) (contract.QueryResponse, error) {
	s.calls++
	return contract.QueryResponse{Entries: []contract.Entry{}}, nil
}

// Units returns matching service names for route authorization tests.
func (s *journalRouteStub) Units(context.Context, contract.UnitsRequest) (contract.UnitsResponse, error) {
	s.calls++
	return contract.UnitsResponse{Units: []string{}}, nil
}

// Vacuum confirms valid requests for route authorization tests.
func (s *journalRouteStub) Vacuum(context.Context, contract.VacuumRequest) error {
	s.calls++
	return nil
}

// journalRouter assembles the real gateway routes with a recording backend.
func journalRouter(role string, stub *journalRouteStub) http.Handler {
	return NewRouter("gateway", "test", modules.Controllers{Journal: controllers.NewJournalController(stub)}, middlewares.AuthenticationOptions{Verifier: routeAuthVerifier{claims: authtoken.AccessClaims{Roles: []string{role}}}})
}

// Requirements: journal-manager/SR-001, journal-manager/FR-001.
func TestJournalQueryRequiresSecurityRole(t *testing.T) {
	for _, tc := range []struct {
		role string
		want int
	}{{"lite-nas-security", 200}, {"admin", 200}, {"operator", 403}} {
		t.Run(tc.role, func(t *testing.T) {
			stub := &journalRouteStub{}
			handler := journalRouter(tc.role, stub)
			if got := managementRequest(handler, "GET", "/api/journal?limit=10", true); got != tc.want {
				t.Fatalf("status=%d, want %d", got, tc.want)
			}
		})
	}
}

// Requirements: journal-manager/SR-002.
func TestJournalRejectsInvalidQueryBeforeRPC(t *testing.T) {
	stub := &journalRouteStub{}
	handler := journalRouter("lite-nas-security", stub)
	if got := managementRequest(handler, "GET", "/api/journal?unit=--help", true); got != 422 {
		t.Fatalf("status=%d", got)
	}
}

// Requirements: journal-manager/FR-001.
func TestJournalAcceptsServiceUnitFilter(t *testing.T) {
	handler := journalRouter("lite-nas-security", &journalRouteStub{})
	if got := managementRequest(handler, "GET", "/api/journal?unit=ssh.service", true); got != 200 {
		t.Fatalf("status=%d", got)
	}
}

// Requirements: journal-manager/SR-001, journal-manager/FR-003.
func TestJournalUnitsRequiresSecurityRole(t *testing.T) {
	handler := journalRouter("operator", &journalRouteStub{})
	if got := managementRequest(handler, "GET", "/api/journal/units", true); got != 403 {
		t.Fatalf("status=%d", got)
	}
}

// Requirements: journal-manager/SR-002, journal-manager/FR-002.
func TestJournalRejectsInvalidVacuumDaysBeforeRPC(t *testing.T) {
	handler := journalRouter("lite-nas-security", &journalRouteStub{})
	request := webtest.NewAuthenticatedRequest(http.MethodPost, "/api/journal/vacuum", []byte(`{"days":0}`))
	request.Header.Set("Content-Type", "application/json")
	if got := webtest.ServeRequest(handler, request).Code; got != 422 {
		t.Fatalf("status=%d", got)
	}
}
