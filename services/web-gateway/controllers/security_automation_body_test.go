package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	contract "lite-nas/shared/contracts/securityautomation"
)

// mitigationBodyCapture records the request handed from the HTTP controller to the RPC service.
type mitigationBodyCapture struct {
	automationControllerStub
	request contract.ForceMitigateRequest
}

// ForceMitigate captures the validated operator request without contacting NATS.
func (s *mitigationBodyCapture) ForceMitigate(_ context.Context, request contract.ForceMitigateRequest) (contract.OperationResponse, error) {
	s.request = request
	return contract.OperationResponse{Completed: true}, nil
}

// TestForceMitigateJSONBodyReachesService verifies the rule and operands survive HTTP decoding.
func TestForceMitigateJSONBodyReachesService(t *testing.T) {
	service := &mitigationBodyCapture{}
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("security automation", "1"))
	huma.Register(api, huma.Operation{OperationID: "force-mitigate-body", Method: http.MethodPost, Path: "/rules/{rule_id}/mitigate"}, NewSecurityAutomationController(service).ForceMitigate)
	request := httptest.NewRequest(http.MethodPost, "/rules/network_reconnaissance/mitigate", strings.NewReader(`{"operands":{"requester_ip_address":"192.0.2.1"}}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(backupContext())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.request.RuleID != "network_reconnaissance" || service.request.Operands["requester_ip_address"] != "192.0.2.1" {
		t.Fatalf("HTTP %d, RPC request = %+v", response.Code, service.request)
	}
}
