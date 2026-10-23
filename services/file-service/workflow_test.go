package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

// memoryWriter models the documented writer result and file contents for HTTP workflow tests.
type memoryWriter struct {
	contents map[string][]byte
}

// command applies one upload operation without requiring host sudo in service tests.
func (writer *memoryWriter) command(_ context.Context, _ fileaccess.Subject, operation string, record *uploadRecord, data []byte) (writerResult, error) {
	if writer.contents == nil {
		writer.contents = make(map[string][]byte)
	}
	result := writerResult{Operation: operation, Device: 1, Inode: 2, UID: 1001}
	switch operation {
	case "create":
		writer.contents[record.Target] = nil
	case "write":
		content := writer.contents[record.Target]
		if record.Offset > int64(len(content)) {
			return writerResult{}, fmt.Errorf("offset exceeds stage size")
		}
		content = append(content[:record.Offset], data...)
		writer.contents[record.Target] = content
		result.NextOffset = int64(len(content))
		result.BytesWritten = int64(len(data))
	case "checksum":
		content := writer.contents[record.Target]
		sum := sha256.Sum256(content)
		result.Size = int64(len(content))
		result.SHA256 = hex.EncodeToString(sum[:])
	case "publish":
		mode := os.FileMode(0o664)
		if record.Private {
			mode = 0o660
		}
		if record.PublicAvatar {
			mode = 0o644
		}
		if err := os.WriteFile(record.Target, writer.contents[record.Target], mode); err != nil {
			return writerResult{}, err
		}
		if err := os.Chmod(record.Target, mode); err != nil {
			return writerResult{}, err
		}
	case "cleanup":
		delete(writer.contents, record.Target)
	}
	return result, nil
}

// mutationRequests grants the existing Auth and RBAC checks for a temporary test root.
func mutationRequests() requestStub {
	readOnly := authorizedRequests(new(int), "sudo")
	return func(ctx context.Context, subject string, request any, response any) error {
		if subject == rbaccontract.CheckPathOperationRPCSubject {
			response.(*rbaccontract.DecisionResponse).Allowed = true
			return nil
		}
		return readOnly(ctx, subject, request, response)
	}
}

// mustWorkflowService creates a service with a deterministic writer and active upload expiry.
func mustWorkflowService(t *testing.T, root string) (*fileService, *memoryWriter) {
	t.Helper()
	service := mustService(t, root, mutationRequests())
	service.cfg.UploadTTL = time.Hour
	writer := &memoryWriter{}
	service.writerCommand = writer.command
	return service, writer
}

// mustWriteWorkflowFile creates one fixture file for a service operation.
func mustWriteWorkflowFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// mustReadWorkflowFile reads a file created by a successful workflow action.
func mustReadWorkflowFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// openWorkflowFile reads a source under the already-authorized test policy.
func openWorkflowFile(_ context.Context, _ fileaccess.Subject, path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// workflowJob builds a filesystem operation for the test subject.
func workflowJob(kind, source, destination string) *operationJob {
	return &operationJob{
		kind: kind, source: source, destination: destination,
		subject: fileaccess.Subject{Username: "alice", UID: "1001", Groups: []string{"sudo"}},
	}
}

// mustRunWorkflowOperation executes one successful fixture operation.
func mustRunWorkflowOperation(t *testing.T, service *fileService, job *operationJob) {
	t.Helper()
	if err := service.runOperation(t.Context(), job, make(chan operationUpdate, 32)); err != nil {
		t.Fatalf("runOperation(%s) error = %v", job.kind, err)
	}
}

// mustWorkflowRequest sends a protected request with optional headers and returns the HTTP result.
func mustWorkflowRequest(t *testing.T, service *fileService, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer token")
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	return response
}

// mustWaitForJob waits for a queued filesystem operation to reach its final state.
func mustWaitForJob(t *testing.T, service *fileService, response *httptest.ResponseRecorder) {
	t.Helper()
	var job operationJob
	if err := json.Unmarshal(response.Body.Bytes(), &job); err != nil || job.ID == "" {
		t.Fatalf("queued job response = %s, %v", response.Body.String(), err)
	}
	deadline := time.After(time.Second)
	for {
		service.jobs.mu.Lock()
		state := service.jobs.jobs[job.ID].State
		service.jobs.mu.Unlock()
		if state == "done" {
			return
		}
		if state == "error" {
			t.Fatalf("queued job %s failed", job.ID)
		}
		select {
		case <-deadline:
			t.Fatalf("queued job %s remained in %s", job.ID, state)
		case <-time.After(time.Millisecond):
		}
	}
}

// TestResumableUploadPublishesWrittenContent covers the final file result.
func TestResumableUploadPublishesWrittenContent(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "report.txt")
	service, _ := mustWorkflowService(t, root)
	content := "hello"
	id := mustStartResumableUpload(t, service, target, content)
	mustWriteUploadChunk(t, service, id, content)
	mustCompleteUpload(t, service, id)
	if data := mustReadWorkflowFile(t, target); data != content {
		t.Fatalf("published content = %q", data)
	}
}

// mustConventionalUpload submits a successful destination-first upload.
func mustConventionalUpload(t *testing.T, service *fileService, target string) {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	destination, err := form.CreateFormField("destination")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(destination, target); err != nil {
		t.Fatal(err)
	}
	file, err := form.CreateFormFile("file", "report.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, "hello"); err != nil {
		t.Fatal(err)
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body.String(), map[string]string{"Content-Type": form.FormDataContentType()})
	if response.Code != http.StatusCreated {
		t.Fatalf("conventional upload = %d %s", response.Code, response.Body.String())
	}
}

// TestConventionalUploadAcceptsPrivateField covers file-service-svc/FR-005 multipart privacy.
func TestConventionalUploadAcceptsPrivateField(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	target := filepath.Join(root, "photo.jpg")
	body, contentType := multipartFields(t, [][2]string{{"destination", target}, {"private", "true"}, {"file", "hello"}})
	response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body, map[string]string{"Content-Type": contentType})
	if response.Code != http.StatusCreated {
		t.Fatalf("upload = %d %s", response.Code, response.Body.String())
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o660 {
		t.Fatalf("mode = %04o", got)
	}
}

// TestConventionalUploadPublishesMultipartFile covers the destination-first upload flow.
func TestConventionalUploadPublishesMultipartFile(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "report.txt")
	service, _ := mustWorkflowService(t, root)
	mustConventionalUpload(t, service, target)
	if data := mustReadWorkflowFile(t, target); data != "hello" {
		t.Fatalf("published content = %q", data)
	}
}

// TestDownloadSupportsFullAndRangeResponses covers authorized file streaming.
func TestDownloadSupportsFullAndRangeResponses(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "read.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, _ := mustWorkflowService(t, root)
	service.readerCommand = func(_ context.Context, _ fileaccess.Subject, _ string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("hello world")), nil
	}
	for _, test := range []struct {
		name    string
		method  string
		rangeAt string
		want    int
	}{
		{name: "full", method: http.MethodGet, want: http.StatusOK},
		{name: "range", method: http.MethodGet, rangeAt: "bytes=0-4", want: http.StatusPartialContent},
		{name: "head", method: http.MethodHead, want: http.StatusOK},
		{name: "invalid range", method: http.MethodGet, rangeAt: "bytes=99-100", want: http.StatusRequestedRangeNotSatisfiable},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := mustWorkflowRequest(t, service, test.method, path, "", map[string]string{"Range": test.rangeAt})
			if response.Code != test.want {
				t.Fatalf("download status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

// TestUserScopedDirectoryCreation covers the fixed mkdir invocation.
func TestUserScopedDirectoryCreation(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, binary string, paths ...string) error {
		switch binary {
		case "/bin/mkdir":
			return os.Mkdir(paths[1], 0o700)
		case "/bin/rm":
			return os.Remove(paths[0])
		default:
			return fmt.Errorf("unexpected binary %q", binary)
		}
	}
	directory := filepath.Join(root, "created")
	created := mustWorkflowRequest(t, service, http.MethodPost, "/mkdir", fmt.Sprintf(`{"path":%q}`, directory), nil)
	if created.Code != http.StatusCreated {
		t.Fatalf("mkdir = %d %s", created.Code, created.Body.String())
	}
}

// TestUserScopedDirectoryCreationUsesPrivateMode covers file-service-svc/FR-010.
func TestUserScopedDirectoryCreationUsesPrivateMode(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	var mode string
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, _ string, paths ...string) error {
		mode = paths[0]
		return os.Mkdir(paths[1], 0o700)
	}
	response := mustWorkflowRequest(t, service, http.MethodPost, "/mkdir",
		fmt.Sprintf(`{"path":%q,"private":true}`, filepath.Join(root, "private")), nil)
	if response.Code != http.StatusCreated {
		t.Fatalf("mkdir = %d %s", response.Code, response.Body.String())
	}
	if mode != "0770" {
		t.Fatalf("mkdir mode = %q", mode)
	}
}

// TestUserScopedFileDeletion covers the fixed rm invocation.
func TestUserScopedFileDeletion(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, _ string, paths ...string) error {
		return os.Remove(paths[0])
	}
	file := filepath.Join(root, "remove.txt")
	mustWriteWorkflowFile(t, file, "data")
	removed := mustWorkflowRequest(t, service, http.MethodDelete, file, "", nil)
	if removed.Code != http.StatusNoContent {
		t.Fatalf("delete = %d %s", removed.Code, removed.Body.String())
	}
}

// TestRunOperationCopiesFileAsAuthorizedUser covers queued file copy execution.
func TestRunOperationCopiesFileAsAuthorizedUser(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.txt")
	destination := filepath.Join(root, "destination.txt")
	mustWriteWorkflowFile(t, source, "hello")
	service, _ := mustWorkflowService(t, root)
	service.readerCommand = openWorkflowFile
	mustRunWorkflowOperation(t, service, workflowJob("copy", source, destination))
	if data := mustReadWorkflowFile(t, destination); data != "hello" {
		t.Fatalf("copied content = %q", data)
	}
}

// mustQueueWorkflowCopy submits and awaits one authorized copy request.
func mustQueueWorkflowCopy(t *testing.T, service *fileService, source, destination string) {
	t.Helper()
	body := fmt.Sprintf(`{"source":%q,"destination":%q}`, source, destination)
	response := mustWorkflowRequest(t, service, http.MethodPost, "/copy", body, nil)
	if response.Code != http.StatusAccepted {
		t.Fatalf("copy request = %d %s", response.Code, response.Body.String())
	}
	mustWaitForJob(t, service, response)
}

// TestTransferQueuesAuthorizedCopy covers the HTTP operation request and worker result.
func TestTransferQueuesAuthorizedCopy(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.txt")
	destination := filepath.Join(root, "destination.txt")
	mustWriteWorkflowFile(t, source, "copy me")
	service, _ := mustWorkflowService(t, root)
	service.readerCommand = openWorkflowFile
	mustQueueWorkflowCopy(t, service, source, destination)
	if content := mustReadWorkflowFile(t, destination); content != "copy me" {
		t.Fatalf("copied content = %q", content)
	}
}

// mustMoveDirectoryTree prepares and moves a nested fixture tree.
func mustMoveDirectoryTree(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "destination")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	mustWriteWorkflowFile(t, filepath.Join(source, "entry.txt"), "data")
	service, _ := mustWorkflowService(t, root)
	service.readerCommand = openWorkflowFile
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, binary string, paths ...string) error {
		if binary == "/bin/mkdir" {
			return os.Mkdir(paths[1], 0o700)
		}
		return os.Remove(paths[0])
	}
	mustRunWorkflowOperation(t, service, workflowJob("move", source, destination))
	return source, destination
}

// TestRunOperationMovesDirectoryTree copies the nested file.
func TestRunOperationMovesDirectoryTree(t *testing.T) {
	_, destination := mustMoveDirectoryTree(t)
	content := mustReadWorkflowFile(t, filepath.Join(destination, "entry.txt"))
	if content != "data" {
		t.Fatalf("moved entry = %q", content)
	}
}

// TestRunOperationRemovesSourceDirectory removes the original tree after moving.
func TestRunOperationRemovesSourceDirectory(t *testing.T) {
	source, _ := mustMoveDirectoryTree(t)
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists after move: %v", err)
	}
}

// TestExpiredUploadCleanupRemovesRecords covers startup expiry for each record state.
func TestExpiredUploadCleanupRemovesRecords(t *testing.T) {
	for _, state := range []string{"completed", "active"} {
		t.Run(state, func(t *testing.T) {
			root := t.TempDir()
			service, writer := mustWorkflowService(t, root)
			id := strings.Repeat("a", 48)
			record := &uploadRecord{
				ID: id, Owner: "alice", UID: "1001", Target: filepath.Join(root, state),
				State: state, Expires: time.Now().Add(-time.Hour),
			}
			if err := service.store.persist(record); err != nil {
				t.Fatal(err)
			}
			service.store.add(record)
			writer.contents = map[string][]byte{record.Target: nil}
			service.cleanupExpired(t.Context())
			if len(service.store.records) != 0 {
				t.Fatalf("cleanupExpired() left %d records", len(service.store.records))
			}
		})
	}
}

// TestRunOperationDeletesDirectoryTree covers reverse-order user-scoped deletion.
func TestRunOperationDeletesDirectoryTree(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "remove")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	mustWriteWorkflowFile(t, filepath.Join(source, "entry.txt"), "data")
	service, _ := mustWorkflowService(t, root)
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, _ string, paths ...string) error {
		return os.Remove(paths[0])
	}
	mustRunWorkflowOperation(t, service, workflowJob("delete", source, ""))
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("deleted directory still exists: %v", err)
	}
}

// TestRunOperationReplacesExistingDestination covers copy onto an authorized file.
func TestRunOperationReplacesExistingDestination(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source.txt")
	destination := filepath.Join(root, "destination.txt")
	mustWriteWorkflowFile(t, source, "new")
	mustWriteWorkflowFile(t, destination, "old")
	service, _ := mustWorkflowService(t, root)
	service.readerCommand = openWorkflowFile
	mustRunWorkflowOperation(t, service, workflowJob("copy", source, destination))
	if data := mustReadWorkflowFile(t, destination); data != "new" {
		t.Fatalf("replaced content = %q", data)
	}
}

// TestConventionalUploadRejectsMalformedParts covers multipart ordering and input checks.
func TestConventionalUploadRejectsMalformedParts(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	for _, test := range []struct {
		name        string
		destination string
		firstFile   bool
		want        int
	}{
		{name: "file first", destination: filepath.Join(root, "first.txt"), firstFile: true, want: http.StatusBadRequest},
		{name: "empty destination", want: http.StatusBadRequest},
		{name: "missing file", destination: filepath.Join(root, "missing.txt"), want: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			if test.firstFile {
				_, _ = form.CreateFormFile("file", "upload.txt")
			} else {
				part, err := form.CreateFormField("destination")
				if err != nil {
					t.Fatal(err)
				}
				_, _ = io.WriteString(part, test.destination)
			}
			if err := form.Close(); err != nil {
				t.Fatal(err)
			}
			response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body.String(),
				map[string]string{"Content-Type": form.FormDataContentType()})
			if response.Code != test.want {
				t.Fatalf("conventional upload = %d %s, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

// multipartFields builds a destination-first form for upload boundary tests.
func multipartFields(t *testing.T, fields [][2]string) (string, string) {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for _, field := range fields {
		part, err := form.CreateFormField(field[0])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(part, field[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	return body.String(), form.FormDataContentType()
}

// TestConventionalUploadRejectsReservedTarget covers an existing reservation.
func TestConventionalUploadRejectsReservedTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "report.txt")
	service, _ := mustWorkflowService(t, root)
	service.store.targets[target] = "reserved"
	body, contentType := multipartFields(t, [][2]string{{"destination", target}, {"file", "hello"}})
	response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body,
		map[string]string{"Content-Type": contentType})
	if response.Code != http.StatusConflict {
		t.Fatalf("conventional upload = %d %s", response.Code, response.Body.String())
	}
}

// TestConventionalUploadReportsWriterFailures covers stage command errors.
func TestConventionalUploadReportsWriterFailures(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "report.txt")
	for _, test := range []struct {
		name   string
		failAt string
		want   int
	}{
		{name: "stage create", failAt: "create", want: http.StatusForbidden},
		{name: "stage write", failAt: "write", want: http.StatusServiceUnavailable},
		{name: "checksum", failAt: "checksum", want: http.StatusServiceUnavailable},
		{name: "publish", failAt: "publish", want: http.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, writer := mustWorkflowService(t, root)
			service.writerCommand = func(ctx context.Context, subject fileaccess.Subject, operation string, record *uploadRecord, data []byte) (writerResult, error) {
				if operation == test.failAt {
					return writerResult{}, fmt.Errorf("simulated %s failure", operation)
				}
				return writer.command(ctx, subject, operation, record, data)
			}
			body, contentType := multipartFields(t, [][2]string{{"destination", target}, {"file", "hello"}})
			response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body,
				map[string]string{"Content-Type": contentType})
			if response.Code != test.want {
				t.Fatalf("conventional upload = %d %s, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

// TestConventionalUploadRejectsExtraPart covers unexpected content after the file.
func TestConventionalUploadRejectsExtraPart(t *testing.T) {
	root := t.TempDir()
	service, _ := mustWorkflowService(t, root)
	body, contentType := multipartFields(t, [][2]string{
		{"destination", filepath.Join(root, "report.txt")}, {"file", "hello"}, {"extra", "bad"},
	})
	response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body,
		map[string]string{"Content-Type": contentType})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("extra upload part = %d %s", response.Code, response.Body.String())
	}
}

// TestConventionalUploadRejectsInvalidContentType covers multipart parsing.
func TestConventionalUploadRejectsInvalidContentType(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	if response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", "not multipart", nil); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid multipart = %d", response.Code)
	}
}

// TestConventionalUploadRejectsInvalidDestinations covers upload path validation.
func TestConventionalUploadRejectsInvalidDestinations(t *testing.T) {
	service, _ := mustWorkflowService(t, t.TempDir())
	for _, destination := range []string{"relative", "/" + strings.Repeat("a", 4097)} {
		t.Run(destination, func(t *testing.T) {
			body, contentType := multipartFields(t, [][2]string{{"destination", destination}, {"file", "hello"}})
			response := mustWorkflowRequest(t, service, http.MethodPost, "/upload", body,
				map[string]string{"Content-Type": contentType})
			if response.Code != http.StatusBadRequest {
				t.Fatalf("invalid destination %q = %d", destination, response.Code)
			}
		})
	}
}
