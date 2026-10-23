package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lite-nas/shared/fileaccess"
)

// mustProtectedRequest sends one authenticated HTTP request through the public adapter.
func mustProtectedRequest(t *testing.T, service *fileService, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	return response
}

// TestOperationInputsRejectMalformedRequests covers file-service-svc/FR-010.
func TestOperationInputsRejectMalformedRequests(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	for _, tc := range []struct {
		path string
		body string
	}{
		{path: "/mkdir", body: ""},
		{path: "/mkdir", body: `{"path":"relative"}`},
		{path: "/mkdir", body: `{"path":"/"}`},
		{path: "/mkdir", body: `{"path":"/tmp/x","extra":true}`},
		{path: "/mkdir", body: `{"path":1}`},
		{path: "/copy", body: `{"source":"/","destination":"/tmp/new"}`},
		{path: "/copy", body: `{"source":"/tmp/a","destination":"relative"}`},
		{path: "/move", body: `{"source":"relative","destination":"/tmp/new"}`},
		{path: "/copy", body: `{"source":"/tmp/a"} {"destination":"/tmp/b"}`},
	} {
		response := mustProtectedRequest(t, service, http.MethodPost, tc.path, tc.body)
		if response.Code != http.StatusBadRequest {
			t.Errorf("POST %s body %q = %d %s", tc.path, tc.body, response.Code, response.Body.String())
		}
	}
	response := mustProtectedRequest(t, service, http.MethodPost, "/copy", `{"source":"/missing-source","destination":"/tmp/new"}`)
	if response.Code != http.StatusNotFound {
		t.Fatalf("missing source = %d %s", response.Code, response.Body.String())
	}
}

// TestOperationJobIsVisibleOnlyToItsOwner covers file-service-svc/FR-011.
func TestOperationJobIsVisibleOnlyToItsOwner(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	id := strings.Repeat("a", 48)
	service.jobs.mu.Lock()
	service.jobs.jobs[id] = &operationJob{ID: id, State: "running", Percent: 42, owner: "alice", uid: "1001"}
	service.jobs.mu.Unlock()
	for _, tc := range []struct {
		path string
		want int
	}{
		{path: "/jobs/bad", want: http.StatusBadRequest},
		{path: "/jobs/" + strings.Repeat("b", 48), want: http.StatusNotFound},
		{path: "/jobs/" + id, want: http.StatusOK},
	} {
		response := mustProtectedRequest(t, service, http.MethodGet, tc.path, "")
		if response.Code != tc.want {
			t.Errorf("GET %s = %d %s, want %d", tc.path, response.Code, response.Body.String(), tc.want)
		}
	}
	response := mustProtectedRequest(t, service, http.MethodGet, "/jobs/"+id, "")
	var job operationJob
	if err := json.Unmarshal(response.Body.Bytes(), &job); err != nil || job.Percent != 42 {
		t.Fatalf("job = %+v, decode error %v", job, err)
	}
	service.jobs.mu.Lock()
	service.jobs.jobs[id].owner = "bob"
	service.jobs.mu.Unlock()
	if response := mustProtectedRequest(t, service, http.MethodGet, "/jobs/"+id, ""); response.Code != http.StatusNotFound {
		t.Fatalf("other owner's job = %d", response.Code)
	}
}

// TestDocumentationRoutesRequireAuthentication covers file-service-svc/IR-001 and SR-001.
func TestDocumentationRoutesRequireAuthentication(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	service.cfg.DocsEnabled = true
	for _, path := range []string{"/docs", "/docs/init.js", "/docs/swagger-ui.css", "/openapi.json"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		unauthorized := httptest.NewRecorder()
		service.ServeHTTP(unauthorized, request)
		if unauthorized.Code != http.StatusUnauthorized {
			t.Errorf("anonymous GET %s = %d", path, unauthorized.Code)
		}
		response := mustProtectedRequest(t, service, http.MethodGet, path, "")
		if response.Code != http.StatusOK || response.Body.Len() == 0 || response.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("authenticated GET %s = %d %v", path, response.Code, response.Header())
		}
	}
	for _, tc := range []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/docs/missing", want: http.StatusNotFound},
		{method: http.MethodPost, path: "/docs", want: http.StatusMethodNotAllowed},
	} {
		response := mustProtectedRequest(t, service, tc.method, tc.path, "")
		if response.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.path, response.Code, tc.want)
		}
	}
	service.cfg.DocsEnabled = false
	if response := mustProtectedRequest(t, service, http.MethodGet, "/docs", ""); response.Code != http.StatusNotFound {
		t.Fatalf("disabled docs = %d", response.Code)
	}
}

// TestNewDirectoryResolutionRejectsUnsafeTargets covers file-service-svc/FR-010.
func TestNewDirectoryResolutionRejectsUnsafeTargets(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "new")
	if got, err := resolveNewDirectory(path); err != nil || got != path {
		t.Fatalf("new directory = %q, %v", got, err)
	}
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if got, err := resolveNewDirectory(path); err != nil || got != path {
		t.Fatalf("existing directory = %q, %v", got, err)
	}
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{"relative", filepath.Join(root, "missing-parent", "new"), file} {
		if _, err := resolveNewDirectory(invalid); err == nil {
			t.Errorf("accepted invalid directory target %q", invalid)
		}
	}
}

// TestUploadBoundaryRejectsInvalidMetadata covers file-service-svc/FR-004.
func TestUploadBoundaryRejectsInvalidMetadata(t *testing.T) {
	valid := `{"destination":"/tmp/photo.jpg","filename":"photo.jpg","size":0,"sha256":"` + strings.Repeat("0", 64) + `"}`
	for _, body := range []string{"", `{"destination":"relative"}`, `{"destination":"/tmp/photo.jpg","filename":"../bad","size":0,"sha256":"` + strings.Repeat("0", 64) + `"}`, valid[:len(valid)-1] + `,"unknown":true}`} {
		var input createInput
		if validateCreateDTO([]byte(body), &input) {
			t.Errorf("accepted invalid upload metadata %q", body)
		}
	}
	var input createInput
	if !validateCreateDTO([]byte(valid), &input) || input.Filename != "photo.jpg" {
		t.Fatalf("valid upload metadata rejected: %+v", input)
	}
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	for _, key := range []string{"", "short"} {
		request := httptest.NewRequest(http.MethodPost, "/proprietary/upload", strings.NewReader(valid))
		request.Header.Set("Authorization", "Bearer token")
		request.Header.Set("Idempotency-Key", key)
		response := httptest.NewRecorder()
		service.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("upload key %q = %d", key, response.Code)
		}
	}
}

// TestChunkWindowAndUnixIdentityBoundaries covers file-service-svc/FR-005 and SR-001.
func TestChunkWindowAndUnixIdentityBoundaries(t *testing.T) {
	window := &chunkWindow{}
	if window.reserve(-1) || !window.reserve(minChunkSize) || window.available() != chunkBudget-minChunkSize {
		t.Fatal("unexpected chunk reservation behavior")
	}
	window.release(minChunkSize)
	if window.available() != chunkBudget || window.reserve(chunkBudget+1) {
		t.Fatal("unexpected chunk release behavior")
	}
	if err := runUserCommand(context.Background(), fileaccess.Subject{}, "/bin/mkdir", "/tmp/noop"); err == nil {
		t.Fatal("missing Unix identity reached command invocation")
	}
}

// TestUploadMetadataSurvivesStoreReload covers file-service-svc/FR-005 and RR-002.
func TestUploadMetadataSurvivesStoreReload(t *testing.T) {
	directory := t.TempDir()
	store, err := newUploadStore(directory)
	if err != nil {
		t.Fatal(err)
	}
	id, err := newUploadID()
	if err != nil || len(id) != 48 {
		t.Fatalf("upload ID = %q, %v", id, err)
	}
	record := &uploadRecord{ID: id, Owner: "alice", UID: "1001", Target: "/tmp/photo.jpg", Filename: "photo.jpg", IdempotencyKey: "request-key", State: "active", ChunkSize: minChunkSize}
	if err := store.persist(record); err != nil {
		t.Fatal(err)
	}
	store.add(record)
	if found, err := store.find(id); err != nil || found != record {
		t.Fatalf("indexed record = %+v, %v", found, err)
	}
	reloaded, err := newUploadStore(directory)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.targets[record.Target] != id || reloaded.keys["alice\x00request-key"] != id {
		t.Fatalf("reloaded store lost active record: %+v", reloaded)
	}
	if err := reloaded.remove(reloaded.records[id]); err != nil {
		t.Fatal(err)
	}
	if _, err := reloaded.find(id); err == nil {
		t.Fatal("removed upload remains indexed")
	}
	if _, err := os.Stat(filepath.Join(directory, id+".json")); !os.IsNotExist(err) {
		t.Fatalf("removed metadata remains on disk: %v", err)
	}
}

// TestUploadStoreRejectsCorruptMetadata covers file-service-svc/RR-002.
func TestUploadStoreRejectsCorruptMetadata(t *testing.T) {
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{name: "malformed JSON", data: []byte("{")},
		{name: "missing owner", data: []byte(`{"id":"one","target":"/tmp/photo.jpg"}`)},
		{name: "wrong ID", data: []byte(`{"id":"other","owner":"alice","target":"/tmp/photo.jpg"}`)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "one.json"), tc.data, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := newUploadStore(directory); err == nil {
				t.Fatal("corrupt metadata was accepted")
			}
		})
	}
}
