package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"

	contract "lite-nas/shared/contracts/backuprestore"
)

// backupBodyCapture records values received by the backend-facing service boundary.
type backupBodyCapture struct {
	backupRestoreStub
	got []string
}

// Backup records the decoded labels for one application snapshot.
func (s *backupBodyCapture) Backup(_ context.Context, _, application, name, description string) (string, error) {
	s.got = []string{application, name, description}
	return "v2", nil
}

// BackupAll records the labels shared by all requested snapshots.
func (s *backupBodyCapture) BackupAll(_ context.Context, _, name, description string) ([]contract.BackupAllResult, error) {
	s.got = []string{name, description}
	return []contract.BackupAllResult{}, nil
}

// Restore records the selected mode after Huma decodes the request body.
func (s *backupBodyCapture) Restore(_ context.Context, _, application, version, mode string) (contract.RestoreResponse, error) {
	s.got = []string{application, version, mode}
	return contract.RestoreResponse{Restored: true}, nil
}

// RestoreByName records the name and mode decoded for one application.
func (s *backupBodyCapture) RestoreByName(_ context.Context, _, application, name, mode string) (contract.RestoreByNameResponse, error) {
	s.got = []string{application, name, mode}
	return contract.RestoreByNameResponse{Version: "v1", Restored: true}, nil
}

// RestoreAllByName records the name and mode decoded for all applications.
func (s *backupBodyCapture) RestoreAllByName(_ context.Context, _, name, mode string) (contract.RestoreAllByNameResponse, error) {
	s.got = []string{name, mode}
	return contract.RestoreAllByNameResponse{Results: []contract.RestoreAllByNameResult{}}, nil
}

// MergeWrite records the decoded content chunk and checksum.
func (s *backupBodyCapture) MergeWrite(_ context.Context, request contract.MergeWriteRequest) (contract.MergeWriteResponse, error) {
	s.got = []string{request.SessionID, request.Path, request.Data, request.SHA256}
	return contract.MergeWriteResponse{NextOffset: 1}, nil
}

// MergeSave records the labels assigned to a saved merge.
func (s *backupBodyCapture) MergeSave(_ context.Context, request contract.MergeSaveRequest) (contract.BackupResponse, error) {
	s.got = []string{request.SessionID, request.Name, request.Description}
	return contract.BackupResponse{Version: "v2"}, nil
}

// TestBackupRestoreJSONBodiesReachService verifies Huma decodes browser JSON before the RPC boundary.
// Requirements: backup-restore/IR-004.
func TestBackupRestoreJSONBodiesReachService(t *testing.T) {
	const sessionID = "0123456789abcdef0123456789abcdef"
	const digest = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	tests := []struct {
		name, path, body string
		want             []string
	}{
		{"backup", "/applications/example/backups", `{"name":"named","description":"note"}`, []string{"example", "named", "note"}},
		{"backup-all", "/backups", `{"name":"named","description":"note"}`, []string{"named", "note"}},
		{"restore", "/applications/example/versions/v1/restore", `{"mode":"merge"}`, []string{"example", "v1", "merge"}},
		{"restore-by-name", "/applications/example/restore-by-name", `{"name":"release one","mode":"merge"}`, []string{"example", "release one", "merge"}},
		{"restore-all-by-name", "/restore-by-name", `{"name":"release one","mode":"replace"}`, []string{"release one", "replace"}},
		{"merge-write", "/merge-sessions/" + sessionID + "/content", `{"path":"/etc/example","offset":0,"data":"YQ==","final":true,"sha256":"` + digest + `"}`, []string{sessionID, "/etc/example", "YQ==", digest}},
		{"merge-save", "/merge-sessions/" + sessionID + "/save", `{"name":"named","description":"note"}`, []string{sessionID, "named", "note"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &backupBodyCapture{}
			router := chi.NewRouter()
			api := humachi.New(router, huma.DefaultConfig("backup", "1"))
			controller := NewBackupRestoreController(service)
			huma.Register(api, huma.Operation{OperationID: "backup", Method: http.MethodPost, Path: "/applications/{application}/backups"}, controller.Backup)
			huma.Register(api, huma.Operation{OperationID: "backup-all", Method: http.MethodPost, Path: "/backups"}, controller.BackupAll)
			huma.Register(api, huma.Operation{OperationID: "restore", Method: http.MethodPost, Path: "/applications/{application}/versions/{version}/restore"}, controller.Restore)
			huma.Register(api, huma.Operation{OperationID: "restore-by-name", Method: http.MethodPost, Path: "/applications/{application}/restore-by-name"}, controller.RestoreByName)
			huma.Register(api, huma.Operation{OperationID: "restore-all-by-name", Method: http.MethodPost, Path: "/restore-by-name"}, controller.RestoreAllByName)
			huma.Register(api, huma.Operation{OperationID: "merge-write", Method: http.MethodPut, Path: "/merge-sessions/{session_id}/content"}, controller.MergeWrite)
			huma.Register(api, huma.Operation{OperationID: "merge-save", Method: http.MethodPost, Path: "/merge-sessions/{session_id}/save"}, controller.MergeSave)
			method := http.MethodPost
			if test.name == "merge-write" {
				method = http.MethodPut
			}
			request := httptest.NewRequest(method, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			request = request.WithContext(backupContext())
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if !reflect.DeepEqual(service.got, test.want) {
				t.Fatalf("service input = %v, want %v; HTTP %d: %s", service.got, test.want, response.Code, response.Body.String())
			}
		})
	}
}

// TestBackupRestoreJSONBodyValidation rejects a label exceeding the HTTP contract.
// Requirements: backup-restore/IR-004.
func TestBackupRestoreJSONBodyValidation(t *testing.T) {
	service := &backupBodyCapture{}
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("backup", "1"))
	huma.Register(api, huma.Operation{OperationID: "backup", Method: http.MethodPost, Path: "/applications/{application}/backups"}, NewBackupRestoreController(service).Backup)
	request := httptest.NewRequest(http.MethodPost, "/applications/example/backups", strings.NewReader(`{"name":"`+strings.Repeat("a", 129)+`"}`))
	request.Header.Set("Content-Type", "application/json")
	request = request.WithContext(backupContext())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("HTTP status = %d, want 422: %s", response.Code, response.Body.String())
	}
}

// TestRestoreRejectsQueryOnlyMode verifies the selected mode must pass JSON validation.
// Requirements: backup-restore/IR-004.
func TestRestoreRejectsQueryOnlyMode(t *testing.T) {
	service := &backupBodyCapture{}
	router := chi.NewRouter()
	api := humachi.New(router, huma.DefaultConfig("backup", "1"))
	huma.Register(api, huma.Operation{OperationID: "restore", Method: http.MethodPost, Path: "/applications/{application}/versions/{version}/restore"}, NewBackupRestoreController(service).Restore)
	request := httptest.NewRequest(http.MethodPost, "/applications/example/versions/v1/restore?mode=merge", nil)
	request = request.WithContext(backupContext())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("HTTP status = %d, want 400: %s", response.Code, response.Body.String())
	}
}
