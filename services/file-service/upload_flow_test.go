package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/fileaccess"
)

// mustStartResumableUpload creates one active upload through the public HTTP route.
func mustStartResumableUpload(t *testing.T, service *fileService, target, content string) string {
	t.Helper()
	sum := sha256.Sum256([]byte(content))
	metadata := fmt.Sprintf(`{"destination":%q,"filename":"report.txt","size":%d,"sha256":%q}`,
		target, len(content), hex.EncodeToString(sum[:]))
	response := mustWorkflowRequest(t, service, http.MethodPost, "/proprietary/upload", metadata,
		map[string]string{"Idempotency-Key": "request-123"})
	if response.Code != http.StatusCreated {
		t.Fatalf("start upload = %d %s", response.Code, response.Body.String())
	}
	var progress struct {
		UploadID string `json:"upload_id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &progress); err != nil {
		t.Fatal(err)
	}
	return progress.UploadID
}

// mustWriteUploadChunk advances a fixture upload to its complete byte count.
func mustWriteUploadChunk(t *testing.T, service *fileService, id, content string) {
	t.Helper()
	response := mustWorkflowRequest(t, service, http.MethodPatch, "/proprietary/upload/"+id, content,
		map[string]string{"Upload-Offset": "0"})
	if response.Code != http.StatusNoContent {
		t.Fatalf("write = %d", response.Code)
	}
}

// mustCompleteUpload publishes a fully written fixture upload.
func mustCompleteUpload(t *testing.T, service *fileService, id string) {
	t.Helper()
	response := mustWorkflowRequest(t, service, http.MethodPut, "/proprietary/upload/"+id+"/completion", "", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("complete = %d", response.Code)
	}
}

// TestUploadResourceReportsActiveState covers GET and HEAD for active uploads.
func TestUploadResourceReportsActiveState(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	path := "/proprietary/upload/" + mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			response := mustWorkflowRequest(t, service, method, path, "", nil)
			if response.Code != http.StatusOK {
				t.Fatalf("active %s = %d", method, response.Code)
			}
		})
	}
}

// TestUploadResourceAcceptsIdenticalChunkReplay covers retrying a written chunk.
func TestUploadResourceAcceptsIdenticalChunkReplay(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	mustWriteUploadChunk(t, service, id, "hello")
	path := "/proprietary/upload/" + id
	replay := mustWorkflowRequest(t, service, http.MethodPatch, path, "hello", map[string]string{"Upload-Offset": "0"})
	if replay.Code != http.StatusNoContent {
		t.Fatalf("replayed chunk = %d", replay.Code)
	}
}

// TestUploadResourceReportsCompletedState covers GET and HEAD after publication.
func TestUploadResourceReportsCompletedState(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	mustWriteUploadChunk(t, service, id, "hello")
	mustCompleteUpload(t, service, id)
	path := "/proprietary/upload/" + id
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			response := mustWorkflowRequest(t, service, method, path, "", nil)
			if response.Code != http.StatusOK {
				t.Fatalf("completed %s = %d", method, response.Code)
			}
		})
	}
}

// TestUploadResourceAcceptsRepeatedCompletion covers an idempotent completion request.
func TestUploadResourceAcceptsRepeatedCompletion(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	mustWriteUploadChunk(t, service, id, "hello")
	mustCompleteUpload(t, service, id)
	path := "/proprietary/upload/" + id
	if response := mustWorkflowRequest(t, service, http.MethodPut, path+"/completion", "", nil); response.Code != http.StatusOK {
		t.Fatalf("repeated completion = %d", response.Code)
	}
}

// TestUploadResourceRejectsWrongOwner covers record ownership.
func TestUploadResourceRejectsWrongOwner(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	record := service.store.records[id]
	record.Owner = "bob"
	if response := mustWorkflowRequest(t, service, http.MethodGet, "/proprietary/upload/"+id, "", nil); response.Code != http.StatusNotFound {
		t.Fatalf("wrong owner = %d", response.Code)
	}
}

// TestUploadResourceRejectsExpiredRecord covers upload expiry.
func TestUploadResourceRejectsExpiredRecord(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	service.store.records[id].Expires = time.Now().Add(-time.Hour)
	if response := mustWorkflowRequest(t, service, http.MethodGet, "/proprietary/upload/"+id, "", nil); response.Code != http.StatusGone {
		t.Fatalf("expired upload = %d", response.Code)
	}
}

// TestPatchUploadRejectsBadChunkPositions covers bounded input before writer invocation.
func TestPatchUploadRejectsBadChunkPositions(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	subject := fileaccess.Subject{Username: "alice", UID: "1001"}
	for _, test := range []struct {
		name   string
		state  string
		offset string
		body   string
		want   int
	}{
		{name: "inactive", state: "pending", offset: "0", body: "hello", want: http.StatusConflict},
		{name: "invalid offset", state: "active", offset: "invalid", body: "hello", want: http.StatusBadRequest},
		{name: "conflicting offset", state: "active", offset: "1", body: "hello", want: http.StatusConflict},
		{name: "oversize", state: "active", offset: "0", body: "hello world", want: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			record := &uploadRecord{State: test.state, Size: 5, ChunkSize: 5}
			request := httptest.NewRequest(http.MethodPatch, "/upload", strings.NewReader(test.body))
			request.Header.Set("Upload-Offset", test.offset)
			response := httptest.NewRecorder()
			service.patchUpload(response, request, subject, record)
			if response.Code != test.want {
				t.Fatalf("patchUpload() = %d %s", response.Code, response.Body.String())
			}
		})
	}
}

// TestCancelUploadRemovesStageAndMetadata covers explicit active-upload cleanup.
func TestCancelUploadRemovesStageAndMetadata(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	response := mustWorkflowRequest(t, service, http.MethodDelete, "/proprietary/upload/"+id, "", nil)
	if len(service.store.records) != 0 {
		t.Fatalf("cancel upload left %d records (status %d)", len(service.store.records), response.Code)
	}
}

// TestCompleteUploadRejectsIncompleteStage covers checksum gating before publication.
func TestCompleteUploadRejectsIncompleteStage(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	record := &uploadRecord{State: "active", Size: 5}
	response := httptest.NewRecorder()
	service.completeUpload(response, httptest.NewRequest(http.MethodPut, "/completion", nil), fileaccess.Subject{}, record)
	if response.Code != http.StatusConflict {
		t.Fatalf("completeUpload() = %d", response.Code)
	}
}

// TestCleanupLoopStopsAfterInitialSweep covers startup cleanup and cancellation.
func TestCleanupLoopStopsAfterInitialSweep(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	service.cfg.CleanupEvery = time.Hour
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	service.cleanupLoop(ctx)
}

// mustExistingUploadFixture prepares an occupied target and matching request metadata.
func mustExistingUploadFixture(t *testing.T) (*fileService, string) {
	t.Helper()
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	target := filepath.Join(root, "report.txt")
	_ = mustStartResumableUpload(t, service, target, "hello")
	sum := sha256.Sum256([]byte("hello"))
	metadata := fmt.Sprintf(`{"destination":%q,"filename":"report.txt","size":5,"sha256":%q}`,
		target, hex.EncodeToString(sum[:]))
	return service, metadata
}

// TestCreateUploadHandlesIdempotency covers repeated request keys.
func TestCreateUploadHandlesIdempotency(t *testing.T) {
	service, metadata := mustExistingUploadFixture(t)
	for _, test := range []struct {
		name string
		key  string
		body string
		want int
	}{
		{name: "same request", key: "request-123", body: metadata, want: http.StatusOK},
		{name: "changed metadata", key: "request-123", body: strings.Replace(metadata, `"size":5`, `"size":6`, 1), want: http.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := mustWorkflowRequest(t, service, http.MethodPost, "/proprietary/upload", test.body,
				map[string]string{"Idempotency-Key": test.key})
			if response.Code != test.want {
				t.Fatalf("create upload = %d %s, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

// TestCreateUploadRejectsReservedTarget covers a new key for an occupied target.
func TestCreateUploadRejectsReservedTarget(t *testing.T) {
	service, metadata := mustExistingUploadFixture(t)
	response := mustWorkflowRequest(t, service, http.MethodPost, "/proprietary/upload", metadata,
		map[string]string{"Idempotency-Key": "request-456"})
	if response.Code != http.StatusConflict {
		t.Fatalf("reserved target = %d %s", response.Code, response.Body.String())
	}
}

// TestCreateUploadReportsStageFailure covers a writer error before metadata persistence.
func TestCreateUploadReportsStageFailure(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	service.writerCommand = func(context.Context, fileaccess.Subject, string, *uploadRecord, []byte) (writerResult, error) {
		return writerResult{}, fmt.Errorf("writer unavailable")
	}
	sum := sha256.Sum256([]byte("hello"))
	metadata := fmt.Sprintf(`{"destination":%q,"filename":"report.txt","size":5,"sha256":%q}`,
		filepath.Join(root, "report.txt"), hex.EncodeToString(sum[:]))
	response := mustWorkflowRequest(t, service, http.MethodPost, "/proprietary/upload", metadata,
		map[string]string{"Idempotency-Key": "request-123"})
	if response.Code != http.StatusForbidden {
		t.Fatalf("failed stage create = %d %s", response.Code, response.Body.String())
	}
}

// TestCancelUploadPreservesRecordWhenCleanupFails covers retryable stage cleanup.
func TestCancelUploadPreservesRecordWhenCleanupFails(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	id := mustStartResumableUpload(t, service, filepath.Join(root, "report.txt"), "hello")
	service.writerCommand = func(context.Context, fileaccess.Subject, string, *uploadRecord, []byte) (writerResult, error) {
		return writerResult{}, fmt.Errorf("cleanup failed")
	}
	response := mustWorkflowRequest(t, service, http.MethodDelete, "/proprietary/upload/"+id, "", nil)
	if len(service.store.records) != 1 {
		t.Fatalf("failed cleanup left %d records (status %d)", len(service.store.records), response.Code)
	}
}

// TestCompleteUploadRejectsChecksumAndPublicationFailures covers final stage checks.
func TestCompleteUploadRejectsChecksumAndPublicationFailures(t *testing.T) {
	root := t.TempDir()
	for _, test := range []struct {
		name   string
		sha    string
		failAt string
	}{
		{name: "checksum mismatch", sha: strings.Repeat("0", 64)},
		{name: "publication failure", failAt: "publish"},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, writer := mustWorkflowService(t, root)
			target := filepath.Join(root, test.name+".txt")
			writer.contents = map[string][]byte{target: []byte("hello")}
			sum := sha256.Sum256([]byte("hello"))
			checksum := hex.EncodeToString(sum[:])
			if test.sha != "" {
				checksum = test.sha
			}
			record := &uploadRecord{State: "active", Target: target, Size: 5, Offset: 5, SHA256: checksum}
			service.writerCommand = func(ctx context.Context, subject fileaccess.Subject, operation string, record *uploadRecord, data []byte) (writerResult, error) {
				if operation == test.failAt {
					return writerResult{}, fmt.Errorf("publish unavailable")
				}
				return writer.command(ctx, subject, operation, record, data)
			}
			response := httptest.NewRecorder()
			service.completeUpload(response, httptest.NewRequest(http.MethodPut, "/completion", nil), fileaccess.Subject{}, record)
			if response.Code != http.StatusConflict {
				t.Fatalf("completeUpload() = %d %s", response.Code, response.Body.String())
			}
		})
	}
}

// TestPatchUploadRejectsChangedReplay covers replay content comparison.
func TestPatchUploadRejectsChangedReplay(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	record := &uploadRecord{
		State: "active", Size: 10, ChunkSize: 5, Offset: 5,
		LastChunkSize: 5, LastChunkSHA256: strings.Repeat("0", 64),
	}
	request := httptest.NewRequest(http.MethodPatch, "/upload", strings.NewReader("other"))
	request.Header.Set("Upload-Offset", "0")
	response := httptest.NewRecorder()
	service.patchUpload(response, request, fileaccess.Subject{}, record)
	if response.Code != http.StatusConflict {
		t.Fatalf("changed replay = %d", response.Code)
	}
}

// TestPatchUploadReportsFullWindow covers concurrent chunk admission.
func TestPatchUploadReportsFullWindow(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	record := &uploadRecord{State: "active", Size: 10, ChunkSize: 5, Offset: 5}
	service.window.used = chunkBudget
	request := httptest.NewRequest(http.MethodPatch, "/upload", strings.NewReader("hello"))
	request.Header.Set("Upload-Offset", "5")
	response := httptest.NewRecorder()
	service.patchUpload(response, request, fileaccess.Subject{}, record)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("full window = %d", response.Code)
	}
}

// existingUploadPermission holds the validated identity of a fixture file.
type existingUploadPermission struct {
	target string
	exists bool
	device uint64
	inode  uint64
}

// mustExistingUploadPermission resolves the fixture file through service policy.
func mustExistingUploadPermission(t *testing.T) (string, existingUploadPermission) {
	t.Helper()
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	subject := fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}
	file := filepath.Join(root, "existing.txt")
	mustWriteWorkflowFile(t, file, "old")
	target, exists, device, inode, err := service.uploadPermission(t.Context(), subject, file)
	if err != nil {
		t.Fatal(err)
	}
	return file, existingUploadPermission{target: target, exists: exists, device: device, inode: inode}
}

// TestUploadPermissionFindsExistingFileIdentity covers returned destination fields.
func TestUploadPermissionFindsExistingFileIdentity(t *testing.T) {
	file, permission := mustExistingUploadPermission(t)
	t.Run("target", func(t *testing.T) {
		if permission.target != file {
			t.Fatalf("target = %q", permission.target)
		}
	})
	t.Run("exists", func(t *testing.T) {
		if !permission.exists {
			t.Fatal("existing file was reported missing")
		}
	})
	t.Run("device", func(t *testing.T) {
		if permission.device == 0 {
			t.Fatal("device identity was empty")
		}
	})
	t.Run("inode", func(t *testing.T) {
		if permission.inode == 0 {
			t.Fatal("inode identity was empty")
		}
	})
}

// TestUploadPermissionRejectsDirectoryDestination covers directory targets.
func TestUploadPermissionRejectsDirectoryDestination(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	subject := fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}}
	if _, _, _, _, err := service.uploadPermission(t.Context(), subject, root); err == nil {
		t.Fatal("uploadPermission() accepted directory destination")
	}
}
