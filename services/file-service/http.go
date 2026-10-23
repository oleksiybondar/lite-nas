package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"lite-nas/services/file-service/filesystem"
	"lite-nas/shared/fileaccess"
	sharedlogger "lite-nas/shared/logger"
)

// fileService owns the HTTP adapter, runtime policy, and upload progress manager.
type fileService struct {
	cfg    serviceConfig
	policy filesystem.Policy
	auth   *fileaccess.Authorizer
	store  *uploadStore
	jobs   *operationManager
	ranges *rangeLimiter
	window *chunkWindow
	logger sharedlogger.Logger
	// Command functions are service-owned so request tests can exercise workflows without host sudo.
	writerCommand func(context.Context, fileaccess.Subject, string, *uploadRecord, []byte) (writerResult, error)
	userCommand   func(context.Context, fileaccess.Subject, string, ...string) error
	readerCommand func(context.Context, fileaccess.Subject, string) (io.ReadCloser, error)
}

// newFileService prepares durable upload state before accepting HTTP traffic.
func newFileService(cfg serviceConfig, policy filesystem.Policy, auth *fileaccess.Authorizer) (*fileService, error) {
	store, err := newUploadStore(cfg.MetadataDir)
	if err != nil {
		return nil, err
	}
	service := &fileService{
		cfg: cfg, policy: policy, auth: auth, store: store,
		ranges: newRangeLimiter(16), window: &chunkWindow{},
		writerCommand: invokeWriter, userCommand: runUserCommand, readerCommand: openCat,
	}
	service.jobs = newOperationManager(service)
	return service, nil
}

// ServeHTTP authenticates every protected request before dispatching its route.
func (s *fileService) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/docs" || strings.HasPrefix(request.URL.Path, "/docs/") || request.URL.Path == "/openapi.json" {
		if !s.cfg.DocsEnabled {
			writeError(writer, http.StatusNotFound, "not_found")
			return
		}
	}
	subject, cookieAuth, err := s.authenticate(request)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) || errors.Is(err, fileaccess.ErrUnauthenticated) {
			writeError(writer, http.StatusUnauthorized, "unauthenticated")
		} else {
			if s.logger != nil {
				s.logger.Warn("authentication unavailable", "cause", err)
			}
			writeError(writer, http.StatusServiceUnavailable, "authentication_unavailable")
		}
		return
	}
	if cookieAuth && isMutation(request.Method) && !sameOrigin(request) {
		writeError(writer, http.StatusForbidden, "csrf_rejected")
		return
	}
	path := request.URL.Path
	switch {
	case path == "/docs" || strings.HasPrefix(path, "/docs/") || path == "/openapi.json":
		s.serveDocs(writer, request)
	case path == "/upload" && request.Method == http.MethodPost:
		s.conventionalUpload(writer, request, subject)
	case path == "/mkdir" && request.Method == http.MethodPost:
		s.createDirectory(writer, request, subject)
	case (path == "/copy" || path == "/move") && request.Method == http.MethodPost:
		s.createTransfer(writer, request, subject)
	case strings.HasPrefix(path, "/jobs/") && request.Method == http.MethodGet:
		s.getJob(writer, request, subject)
	case path == "/proprietary/upload-window" && request.Method == http.MethodGet:
		s.uploadWindow(writer)
	case path == "/proprietary/upload" && request.Method == http.MethodPost:
		s.createUpload(writer, request, subject)
	case strings.HasPrefix(path, "/proprietary/upload/"):
		s.uploadResource(writer, request, subject)
	case strings.HasPrefix(path, "/ls/") && request.Method == http.MethodGet:
		s.list(writer, request, subject)
	case strings.HasPrefix(path, "/download/") && (request.Method == http.MethodGet || request.Method == http.MethodHead):
		s.download(writer, request, subject, true)
	case strings.HasPrefix(path, "/download/") || strings.HasPrefix(path, "/ls/"):
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
	case strings.HasPrefix(path, "/") && (request.Method == http.MethodGet || request.Method == http.MethodHead):
		s.download(writer, request, subject, false)
	case strings.HasPrefix(path, "/") && request.Method == http.MethodDelete:
		s.deleteFile(writer, request, subject)
	default:
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
	}
}

// authenticate resolves a Bearer token or the existing browser access cookie.
func (s *fileService) authenticate(request *http.Request) (fileaccess.Subject, bool, error) {
	var token string
	var cookieAuth bool
	authorization := request.Header.Get("Authorization")
	if authorization != "" {
		if !strings.HasPrefix(authorization, "Bearer ") || strings.Count(authorization, " ") != 1 {
			return fileaccess.Subject{}, false, fileaccess.ErrUnauthenticated
		}
		token = strings.TrimPrefix(authorization, "Bearer ")
	} else {
		cookie, err := request.Cookie("lite-nas-at")
		if err != nil {
			return fileaccess.Subject{}, false, err
		}
		token, cookieAuth = cookie.Value, true
	}
	subject, err := s.auth.Resolve(request.Context(), token)
	return subject, cookieAuth, err
}

// sameOrigin enforces an Origin or Referer match on browser cookie mutations.
func sameOrigin(request *http.Request) bool {
	source := request.Header.Get("Origin")
	if source == "" {
		source = request.Header.Get("Referer")
	}
	parsed, err := url.Parse(source)
	if err != nil || parsed.Host == "" || parsed.Host != request.Host {
		return false
	}
	proxyScheme := request.Header.Get("X-Forwarded-Proto")
	if proxyScheme == "" {
		proxyScheme = "http"
	}
	return parsed.Scheme == proxyScheme
}

// isMutation identifies requests that can change upload or file state.
func isMutation(method string) bool {
	return method == http.MethodPost || method == http.MethodPatch ||
		method == http.MethodPut || method == http.MethodDelete
}

// writeError emits the stable public error envelope without a host path.
func writeError(writer http.ResponseWriter, status int, code string) {
	writeJSON(writer, status, map[string]any{"error": map[string]any{"code": code}})
}

// writeJSON serializes one response with a declared status code.
func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

// decisionError maps known authorization outcomes without exposing path details.
func decisionError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, fileaccess.ErrDenied),
		errors.Is(err, filesystem.ErrDeniedPath),
		errors.Is(err, filesystem.ErrOutsideWriteRoots),
		errors.Is(err, filesystem.ErrProtectedPath),
		errors.Is(err, os.ErrPermission):
		writeError(writer, http.StatusForbidden, "forbidden")
	case errors.Is(err, fileaccess.ErrInvalidPath):
		writeError(writer, http.StatusBadRequest, "invalid_path")
	default:
		writeError(writer, http.StatusServiceUnavailable, "authorization_unavailable")
	}
}
