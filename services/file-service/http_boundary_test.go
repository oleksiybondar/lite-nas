package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lite-nas/shared/fileaccess"
)

// TestSameOriginChecksBrowserSource covers cookie-authenticated mutation origins.
func TestSameOriginChecksBrowserSource(t *testing.T) {
	for _, test := range []struct {
		name    string
		origin  string
		referer string
		proto   string
		want    bool
	}{
		{name: "same origin", origin: "http://example.test", want: true},
		{name: "referer fallback", referer: "http://example.test/page", want: true},
		{name: "wrong host", origin: "http://other.test"},
		{name: "wrong scheme", origin: "https://example.test"},
		{name: "forwarded HTTPS", origin: "https://example.test", proto: "https", want: true},
		{name: "missing origin"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://example.test/upload", nil)
			request.Header.Set("Origin", test.origin)
			request.Header.Set("Referer", test.referer)
			request.Header.Set("X-Forwarded-Proto", test.proto)
			if got := sameOrigin(request); got != test.want {
				t.Fatalf("sameOrigin() = %v, want %v", got, test.want)
			}
		})
	}
}

// TestTransferRejectsSymlinksAndRecursiveDestinations covers unsafe move shapes.
func TestTransferRejectsSymlinksAndRecursiveDestinations(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	if err := os.Mkdir(source, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	service, _ := mustWorkflowService(t, root)
	for _, test := range []struct {
		name   string
		source string
		dest   string
		want   int
	}{
		{name: "symlink source", source: link, dest: filepath.Join(root, "copy"), want: http.StatusBadRequest},
		{name: "same destination", source: source, dest: source, want: http.StatusConflict},
		{name: "recursive destination", source: source, dest: filepath.Join(source, "child"), want: http.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"source":%q,"destination":%q}`, test.source, test.dest)
			response := mustWorkflowRequest(t, service, http.MethodPost, "/copy", body, nil)
			if response.Code != test.want {
				t.Fatalf("createTransfer() = %d %s, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

// TestDeleteDirectoryQueuesRecursiveRemoval covers directory deletion dispatch.
func TestDeleteDirectoryQueuesRecursiveRemoval(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "empty")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	service, _ := mustWorkflowService(t, root)
	service.userCommand = func(_ context.Context, _ fileaccess.Subject, _ string, paths ...string) error {
		return os.Remove(paths[0])
	}
	response := mustWorkflowRequest(t, service, http.MethodDelete, directory, "", nil)
	if response.Code != http.StatusAccepted {
		t.Fatalf("directory delete = %d %s", response.Code, response.Body.String())
	}
	mustWaitForJob(t, service, response)
}

// TestDeleteFileRejectsUnsafeTargetsAndCommandFailure covers mutation denial paths.
func TestDeleteFileRejectsUnsafeTargetsAndCommandFailure(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}
	service, _ := mustWorkflowService(t, root)
	service.userCommand = func(context.Context, fileaccess.Subject, string, ...string) error {
		return errors.New("command failed")
	}
	for _, test := range []struct {
		name string
		path string
		want int
	}{
		{name: "symlink", path: link, want: http.StatusBadRequest},
		{name: "missing", path: filepath.Join(root, "missing"), want: http.StatusNotFound},
		{name: "outside root", path: "/etc/passwd", want: http.StatusForbidden},
		{name: "command failure", path: file, want: http.StatusConflict},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := mustWorkflowRequest(t, service, http.MethodDelete, test.path, "", nil)
			if response.Code != test.want {
				t.Fatalf("delete = %d %s, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

// TestDownloadHonorsPreconditionsAndReaderFailures covers response guards.
func TestDownloadHonorsPreconditionsAndReaderFailures(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "read.txt")
	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, _ := mustWorkflowService(t, root)
	for _, test := range []struct {
		name    string
		headers map[string]string
		reader  func(context.Context, fileaccess.Subject, string) (io.ReadCloser, error)
		want    int
	}{
		{name: "stale match", headers: map[string]string{"If-Match": "stale"}, want: http.StatusPreconditionFailed},
		{name: "reader failed", reader: func(context.Context, fileaccess.Subject, string) (io.ReadCloser, error) {
			return nil, errors.New("reader offline")
		}, want: http.StatusForbidden},
		{name: "short reader", reader: func(context.Context, fileaccess.Subject, string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("")), nil
		}, want: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			service.readerCommand = test.reader
			response := mustWorkflowRequest(t, service, http.MethodGet, path, "", test.headers)
			if response.Code != test.want {
				t.Fatalf("download = %d %s, want %d", response.Code, response.Body.String(), test.want)
			}
		})
	}
}

// TestDecisionErrorMapsKnownFailures covers stable HTTP authorization errors.
func TestDecisionErrorMapsKnownFailures(t *testing.T) {
	for _, test := range []struct {
		err  error
		want int
	}{
		{err: fileaccess.ErrDenied, want: http.StatusForbidden},
		{err: fileaccess.ErrInvalidPath, want: http.StatusBadRequest},
		{err: errors.New("RPC offline"), want: http.StatusServiceUnavailable},
	} {
		response := httptest.NewRecorder()
		decisionError(response, test.err)
		if response.Code != test.want {
			t.Errorf("decisionError(%v) = %d, want %d", test.err, response.Code, test.want)
		}
	}
}
