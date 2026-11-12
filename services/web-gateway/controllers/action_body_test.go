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

	background "lite-nas/shared/contracts/backgroundtasks"
)

// usbBodyCapture records the USBGuard policy flag passed across the service boundary.
type usbBodyCapture struct {
	managementUSBContractStub
	permanent bool
}

// Action records whether a browser action requests a persistent policy rule.
func (s *usbBodyCapture) Action(_ context.Context, _, _, _ string, permanent bool) error {
	s.permanent = permanent
	return nil
}

// scanBodyCapture records the selected folder passed to the antivirus service.
type scanBodyCapture struct {
	securityTaskContractStub
	folder string
}

// StartAntivirusScan records the folder Huma decoded from the JSON body.
func (s *scanBodyCapture) StartAntivirusScan(_ context.Context, _, folder string) (background.StartResponse, error) {
	s.folder = folder
	return background.StartResponse{}, nil
}

// TestUSBActionJSONBodyReachesService verifies that permanent policy survives HTTP decoding.
func TestUSBActionJSONBodyReachesService(t *testing.T) {
	service := &usbBodyCapture{}
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("management", "1"))
	huma.Register(api, huma.Operation{OperationID: "usb-allow", Method: http.MethodPost, Path: "/devices/{device_id}/allow"}, NewUSBManagerController(service).Allow)
	request := httptest.NewRequest(http.MethodPost, "/devices/1/allow", strings.NewReader(`{"permanent":true}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(backupContext())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if !service.permanent {
		t.Fatalf("permanent = false; HTTP %d: %s", response.Code, response.Body.String())
	}
}

// TestAntivirusScanJSONBodyReachesService verifies that the selected folder reaches the task service.
func TestAntivirusScanJSONBodyReachesService(t *testing.T) {
	service := &scanBodyCapture{}
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("security tasks", "1"))
	huma.Register(api, huma.Operation{OperationID: "antivirus-scan", Method: http.MethodPost, Path: "/antivirus/scans"}, NewSecurityTasksController(service).StartAntivirusScan)
	request := httptest.NewRequest(http.MethodPost, "/antivirus/scans", strings.NewReader(`{"folder":"/home/media"}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(backupContext())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if service.folder != "/home/media" {
		t.Fatalf("folder = %q; HTTP %d: %s", service.folder, response.Code, response.Body.String())
	}
}

// TestAntivirusScanRejectsQueryOnlyFolder verifies the folder is validated as JSON.
func TestAntivirusScanRejectsQueryOnlyFolder(t *testing.T) {
	service := &scanBodyCapture{}
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("security tasks", "1"))
	huma.Register(api, huma.Operation{OperationID: "antivirus-scan", Method: http.MethodPost, Path: "/antivirus/scans"}, NewSecurityTasksController(service).StartAntivirusScan)
	request := httptest.NewRequest(http.MethodPost, "/antivirus/scans?folder=/home/media", nil)
	request = request.WithContext(backupContext())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want 400: %s", response.Code, response.Body.String())
	}
}
