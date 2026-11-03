package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"

	authcontract "lite-nas/shared/contracts/auth"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

// fakeRequester controls the Auth and RBAC decisions without a live NATS server.
type fakeRequester struct {
	allowed bool
	valid   bool
	calls   []string
	inputs  []any
	failRPC string
}

// Request returns one contract response and records its RPC subject.
func (f *fakeRequester) Request(_ context.Context, subject string, input any, response any) error {
	f.calls = append(f.calls, subject)
	f.inputs = append(f.inputs, input)
	if subject == f.failRPC {
		return errors.New("RPC unavailable")
	}
	switch subject {
	case authcontract.ValidateAccessTokenRPCSubject:
		*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{
			Valid: f.valid, Status: authcontract.StatusAuthenticated, Username: "operator",
		}
	case rbaccontract.GetSubjectRolesRPCSubject:
		*response.(*rbaccontract.GetSubjectRolesResponse) = rbaccontract.GetSubjectRolesResponse{UID: "1000"}
	case rbaccontract.CanReadPathRPCSubject:
		*response.(*rbaccontract.DecisionResponse) = rbaccontract.DecisionResponse{Allowed: f.allowed}
	}
	return nil
}

// mustStreamFixture creates a file and an isolated service with controlled decisions.
func mustStreamFixture(t *testing.T, allowed bool) (*streamService, string, *fakeRequester) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "movie.avi")
	if err := os.WriteFile(path, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	client := &fakeRequester{allowed: allowed, valid: true}
	auth, err := fileaccess.NewAuthorizer(client)
	if err != nil {
		t.Fatal(err)
	}
	return &streamService{
		cfg: serviceConfig{ReadRoots: []string{root}}, auth: auth,
		open: openAuthorizedFile, readAsUser: userReader, transcodes: make(chan struct{}, 2),
	}, path, client
}

// requestMedia executes one loopback request with an optional access token.
func requestMedia(service *streamService, path, token, byteRange string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if byteRange != "" {
		request.Header.Set("Range", byteRange)
	}
	recorder := httptest.NewRecorder()
	service.ServeHTTP(recorder, request)
	return recorder
}

// Requirements: stream-service-svc/FR-001.
func TestAnonymousCannotRead(t *testing.T) {
	service, path, client := mustStreamFixture(t, true)
	response := requestMedia(service, path, "", "")
	if response.Code != http.StatusUnauthorized || len(client.calls) != 0 {
		t.Fatalf("anonymous response=%d, rpc calls=%d", response.Code, len(client.calls))
	}
}

// Requirements: stream-service-svc/FR-003 and FR-004.
func TestOwnerCanSeekOriginalAVI(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	response := requestMedia(service, path, "valid", "bytes=2-3")
	if response.Code != http.StatusPartialContent || response.Body.String() != "cd" {
		t.Fatalf("response=%d body=%q", response.Code, response.Body.String())
	}
}

// Requirements: stream-service-svc/FR-003 and SR-001.
func TestDeniedFileHasNoBytes(t *testing.T) {
	service, path, _ := mustStreamFixture(t, false)
	response := requestMedia(service, path, "valid", "bytes=0-0")
	if response.Code != http.StatusNotFound || strings.Contains(response.Body.String(), "abcdef") {
		t.Fatalf("response=%d body=%q", response.Code, response.Body.String())
	}
}

// Requirements: stream-service-svc/FR-004.
func TestInvalidRangeDoesNotReadBeyondFile(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	response := requestMedia(service, path, "valid", "bytes=999-")
	if response.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("response=%d", response.Code)
	}
}

// Requirements: stream-service-svc/FR-003 and SR-001.
func TestEncodedSeparatorIsRejected(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	escaped := strings.Replace(path, "/movie.avi", "%2fmovie.avi", 1)
	response := requestMedia(service, escaped, "valid", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("response=%d", response.Code)
	}
}

// Requirements: stream-service-svc/FR-001 and FR-005.
func TestAnonymousCannotStartTranscode(t *testing.T) {
	service, path, client := mustStreamFixture(t, true)
	response := requestMedia(service, "/transcode"+path+"?start=10", "", "")
	if response.Code != http.StatusUnauthorized || len(client.calls) != 0 {
		t.Fatalf("response=%d, rpc calls=%d", response.Code, len(client.calls))
	}
}

// Requirements: stream-service-svc/FR-005.
func TestStillVideoRejectsTimeSeek(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	still := strings.TrimSuffix(path, ".avi") + ".heic"
	if err := os.WriteFile(still, []byte("still"), 0o600); err != nil {
		t.Fatal(err)
	}
	response := requestMedia(service, "/transcode"+still+"?start=1", "valid", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("response=%d", response.Code)
	}
}

// Requirements: stream-service-svc/FR-003 and FR-005.
func TestAuthorizedHEICTranscodeUsesApprovedPath(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	still := strings.TrimSuffix(path, ".avi") + ".HEIC"
	if err := os.Rename(path, still); err != nil {
		t.Fatal(err)
	}
	var approved string
	service.transcodeFile = func(writer http.ResponseWriter, _ *http.Request, _ fileaccess.Subject, path string, _ os.FileInfo, _ int64) {
		approved = path
		writer.WriteHeader(http.StatusOK)
	}
	response := requestMedia(service, "/transcode"+still+"?start=0", "valid", "")
	if response.Code != http.StatusOK || approved != still {
		t.Fatalf("response=%d approved=%q", response.Code, approved)
	}
}

// Requirements: stream-service-svc/IR-001.
func TestTranscodeSeekRejectsOptionShapedInput(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	response := requestMedia(service, "/transcode"+path+"?start=-1", "valid", "")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("response=%d", response.Code)
	}
}

// Requirements: stream-service-svc/FR-004 and SR-001.
func TestPrivateFileFallsBackToVerifiedUserReader(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	service.open = func(string, os.FileInfo) (*os.File, error) { return nil, syscall.EACCES }
	service.readAsUser = func(_ context.Context, _ fileaccess.Subject, _ string, _ os.FileInfo, offset, count int64) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("abcdef"[offset : offset+count])), nil
	}
	response := requestMedia(service, path, "valid", "bytes=2-3")
	if response.Code != http.StatusPartialContent || response.Body.String() != "cd" {
		t.Fatalf("response=%d body=%q", response.Code, response.Body.String())
	}
}

// Requirements: stream-service-svc/FR-005 and TR-001.
func TestAuthorizedTranscodeReceivesRequestedTime(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	var start int64 = -1
	service.transcodeFile = func(writer http.ResponseWriter, _ *http.Request, _ fileaccess.Subject, _ string, _ os.FileInfo, position int64) {
		start = position
		writer.WriteHeader(http.StatusOK)
	}
	response := requestMedia(service, "/transcode"+path+"?start=30", "valid", "")
	if response.Code != http.StatusOK || start != 30 {
		t.Fatalf("response=%d start=%d", response.Code, start)
	}
}

// Requirements: stream-service-svc/FR-005 and TR-001.
func TestBrowserRangeHeaderStillReceivesTranscode(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	service.transcodeFile = func(writer http.ResponseWriter, _ *http.Request, _ fileaccess.Subject, _ string, _ os.FileInfo, _ int64) {
		writer.WriteHeader(http.StatusOK)
	}
	response := requestMedia(service, "/transcode"+path+"?start=0", "valid", "bytes=0-")
	if response.Code != http.StatusOK {
		t.Fatalf("response=%d", response.Code)
	}
}

// Requirements: stream-service-svc/FR-002 and FR-004.
func TestAuthorizedPreviewReceivesApprovedPath(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	preview := strings.TrimSuffix(path, ".avi") + ".HEIC"
	if err := os.Rename(path, preview); err != nil {
		t.Fatal(err)
	}
	var received string
	service.previewFile = func(writer http.ResponseWriter, _ *http.Request, _ fileaccess.Subject, approved string, _ os.FileInfo) {
		received = approved
		writer.WriteHeader(http.StatusOK)
	}
	response := requestMedia(service, "/preview"+preview, "valid", "")
	if response.Code != http.StatusOK || received != preview {
		t.Fatalf("response=%d path=%q", response.Code, received)
	}
}

// Requirements: stream-service-svc/FR-002 and SR-001.
func TestDeniedPreviewDoesNotStartConversion(t *testing.T) {
	service, path, _ := mustStreamFixture(t, false)
	started := false
	service.previewFile = func(http.ResponseWriter, *http.Request, fileaccess.Subject, string, os.FileInfo) {
		started = true
	}
	response := requestMedia(service, "/preview"+path, "valid", "")
	if response.Code != http.StatusNotFound || started {
		t.Fatalf("response=%d started=%t", response.Code, started)
	}
}

// Requirements: stream-service-svc/FR-001, FR-002, and IR-001.
func TestPreviewUsesAuthAndRBACContracts(t *testing.T) {
	service, path, client := mustStreamFixture(t, true)
	service.previewFile = func(writer http.ResponseWriter, _ *http.Request, _ fileaccess.Subject, _ string, _ os.FileInfo) {
		writer.WriteHeader(http.StatusOK)
	}
	response := requestMedia(service, "/preview"+path, "verified-token", "")
	got := struct {
		status int
		calls  []string
		inputs []any
	}{response.Code, client.calls, client.inputs}
	want := struct {
		status int
		calls  []string
		inputs []any
	}{
		http.StatusOK,
		[]string{
			authcontract.ValidateAccessTokenRPCSubject,
			rbaccontract.GetSubjectRolesRPCSubject,
			rbaccontract.CanReadPathRPCSubject,
		},
		[]any{
			authcontract.ValidateAccessTokenRequest{AccessToken: "verified-token"},
			rbaccontract.GetSubjectRolesRequest{Username: "operator"},
			rbaccontract.CheckPathRequest{UID: "1000", Path: path},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RPC chain = %+v, want %+v", got, want)
	}
}

// Requirements: stream-service-svc/FR-001, FR-002, and RR-001.
func TestPreviewFailsClosedOnAuthOrRBACFailure(t *testing.T) {
	for _, test := range []struct {
		name    string
		failed  string
		valid   bool
		allowed bool
		status  int
		calls   int
	}{
		{"invalid token", "", false, true, http.StatusUnauthorized, 1},
		{"Auth unavailable", authcontract.ValidateAccessTokenRPCSubject, true, true, http.StatusServiceUnavailable, 1},
		{"RBAC identity unavailable", rbaccontract.GetSubjectRolesRPCSubject, true, true, http.StatusServiceUnavailable, 2},
		{"RBAC read unavailable", rbaccontract.CanReadPathRPCSubject, true, true, http.StatusServiceUnavailable, 3},
		{"RBAC read denied", "", true, false, http.StatusNotFound, 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, path, client := mustStreamFixture(t, test.allowed)
			client.valid = test.valid
			client.failRPC = test.failed
			service.previewFile = func(http.ResponseWriter, *http.Request, fileaccess.Subject, string, os.FileInfo) {
				t.Fatal("failed authorization started preview")
			}
			response := requestMedia(service, "/preview"+path, "token", "")
			got := struct{ status, calls int }{response.Code, len(client.calls)}
			want := struct{ status, calls int }{test.status, test.calls}
			if got != want {
				t.Fatalf("result = %+v, want %+v", got, want)
			}
		})
	}
}

// TestMediaContentTypeUsesPassiveAllowlist covers supported and unknown media types.
func TestMediaContentTypeUsesPassiveAllowlist(t *testing.T) {
	for _, test := range []struct{ path, want string }{
		{"movie.MP4", "video/mp4"},
		{"movie.mov", "video/quicktime"},
		{"movie.avi", "video/x-msvideo"},
		{"movie.webm", "video/webm"},
		{"movie.mkv", "video/x-matroska"},
		{"photo.jpeg", "image/jpeg"},
		{"photo.heic", "image/heic"},
		{"photo.heif", "image/heif"},
		{"photo.png", "image/png"},
		{"unknown.txt", "application/octet-stream"},
	} {
		t.Run(test.path, func(t *testing.T) {
			if got := mediaContentType(test.path); got != test.want {
				t.Fatalf("media type = %q, want %q", got, test.want)
			}
		})
	}
}

// TestPreviewRejectsUnsupportedMedia checks the endpoint before starting a decoder.
func TestPreviewRejectsUnsupportedMedia(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	response := httptest.NewRecorder()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	service.preview(response, httptest.NewRequest(http.MethodGet, "/preview", nil), fileaccess.Subject{}, path, info)
	if response.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("preview response = %d", response.Code)
	}
}

// TestConversionEndpointsLimitConcurrentWork checks both conversion queues.
func TestConversionEndpointsLimitConcurrentWork(t *testing.T) {
	for _, test := range []struct {
		name    string
		preview bool
	}{
		{"preview", true},
		{"transcode", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, path, _ := mustStreamFixture(t, true)
			service.transcodes = make(chan struct{}, 1)
			service.transcodes <- struct{}{}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.preview {
				service.preview(response, request, fileaccess.Subject{}, path+".heic", info)
			} else {
				service.transcode(response, request, fileaccess.Subject{}, path, info, 0)
			}
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("conversion response = %d", response.Code)
			}
		})
	}
}

// TestConversionEndpointsRejectUnverifiedIdentity checks subprocess startup.
func TestConversionEndpointsRejectUnverifiedIdentity(t *testing.T) {
	for _, test := range []struct {
		name    string
		preview bool
	}{
		{"preview", true},
		{"transcode", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, path, _ := mustStreamFixture(t, true)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.preview {
				service.preview(response, request, fileaccess.Subject{}, path+".heic", info)
			} else {
				service.transcode(response, request, fileaccess.Subject{}, path, info, 0)
			}
			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("conversion response = %d", response.Code)
			}
		})
	}
}

// TestMediaRequestRejectsInvalidRoutes checks route validation before file access.
func TestMediaRequestRejectsInvalidRoutes(t *testing.T) {
	for _, test := range []struct {
		name, method, prefix, query string
		status                      int
	}{
		{"write method", http.MethodPost, "", "", http.StatusMethodNotAllowed},
		{"preview head", http.MethodHead, "/preview", "", http.StatusMethodNotAllowed},
		{"unknown query", http.MethodGet, "", "?unknown=1", http.StatusBadRequest},
		{"invalid seek", http.MethodGet, "/transcode", "?start=86401", http.StatusBadRequest},
		{"duplicate seek", http.MethodGet, "/transcode", "?start=1&start=2", http.StatusBadRequest},
		{"extra query", http.MethodGet, "/transcode", "?other=1", http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, path, _ := mustStreamFixture(t, true)
			request := httptest.NewRequest(test.method, test.prefix+path+test.query, nil)
			response := httptest.NewRecorder()
			service.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("media response = %d, want %d", response.Code, test.status)
			}
		})
	}
}

// TestMediaRequestChecksReadScopeAndFileType verifies filesystem boundaries.
func TestMediaRequestChecksReadScopeAndFileType(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*streamService, string) string
	}{
		{"outside storage", func(service *streamService, path string) string {
			service.cfg.ReadRoots = []string{filepath.Dir(filepath.Dir(path)) + "/elsewhere"}
			return path
		}},
		{"denied storage", func(service *streamService, path string) string {
			service.cfg.DeniedPaths = []string{filepath.Dir(path)}
			return path
		}},
		{"directory", func(_ *streamService, path string) string { return filepath.Dir(path) }},
		{"missing file", func(_ *streamService, path string) string { return path + ".missing" }},
		{"symlink outside", func(_ *streamService, path string) string {
			link := path + ".link"
			if err := os.Symlink("/etc/hosts", link); err != nil {
				t.Fatal(err)
			}
			return link
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, path, _ := mustStreamFixture(t, true)
			response := requestMedia(service, test.edit(service, path), "valid", "")
			if response.Code != http.StatusNotFound {
				t.Fatalf("media response = %d", response.Code)
			}
		})
	}
}

// TestPrivateOriginalHandlesHeadAndEmptyBody checks authorized metadata responses.
func TestPrivateOriginalHandlesHeadAndEmptyBody(t *testing.T) {
	for _, test := range []struct {
		name, method, content string
	}{
		{"head", http.MethodHead, "abcdef"},
		{"empty", http.MethodGet, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			service, path, _ := mustStreamFixture(t, true)
			if err := os.WriteFile(path, []byte(test.content), 0o600); err != nil {
				t.Fatal(err)
			}
			service.open = func(string, os.FileInfo) (*os.File, error) { return nil, syscall.EACCES }
			request := httptest.NewRequest(test.method, path, nil)
			request.Header.Set("Authorization", "Bearer valid")
			response := httptest.NewRecorder()
			service.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("private response = %d", response.Code)
			}
		})
	}
}

// TestPrivateOriginalFailsClosedWhenReaderCannotStart checks the user fallback.
func TestPrivateOriginalFailsClosedWhenReaderCannotStart(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	service.open = func(string, os.FileInfo) (*os.File, error) { return nil, syscall.EACCES }
	service.readAsUser = func(context.Context, fileaccess.Subject, string, os.FileInfo, int64, int64) (io.ReadCloser, error) {
		return nil, errors.New("reader unavailable")
	}
	response := requestMedia(service, path, "valid", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("private response = %d", response.Code)
	}
}

// TestPrivateOriginalFailsClosedWhenReaderProducesNoBytes checks early EOF.
func TestPrivateOriginalFailsClosedWhenReaderProducesNoBytes(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	service.open = func(string, os.FileInfo) (*os.File, error) { return nil, syscall.EACCES }
	service.readAsUser = func(context.Context, fileaccess.Subject, string, os.FileInfo, int64, int64) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("")), nil
	}
	response := requestMedia(service, path, "valid", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("private response = %d", response.Code)
	}
}

// TestPreviewRejectsRanges checks conversion's response contract.
func TestPreviewRejectsRanges(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	preview := strings.TrimSuffix(path, ".avi") + ".heic"
	if err := os.Rename(path, preview); err != nil {
		t.Fatal(err)
	}
	response := requestMedia(service, "/preview"+preview, "valid", "bytes=0-1")
	if response.Code != http.StatusBadRequest {
		t.Fatalf("preview response = %d", response.Code)
	}
}

// TestMediaOpenFailsClosedOnUnexpectedError checks the direct file opener.
func TestMediaOpenFailsClosedOnUnexpectedError(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	service.open = func(string, os.FileInfo) (*os.File, error) {
		return nil, errors.New("file replaced")
	}
	response := requestMedia(service, path, "valid", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("media response = %d", response.Code)
	}
}

// TestBrowserCookieAuthenticatesMedia checks the same access token as the UI.
func TestBrowserCookieAuthenticatesMedia(t *testing.T) {
	service, path, _ := mustStreamFixture(t, true)
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "valid"})
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("cookie response = %d", response.Code)
	}
}
