package main

import (
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/fileaccess"
)

// mustAuthorizedVideoSource prepares a regular video file inside an authorized storage root.
func mustAuthorizedVideoSource(t *testing.T) (*fileService, string) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "clip.mp4")
	if err := os.WriteFile(path, []byte("video"), 0o600); err != nil {
		t.Fatal(err)
	}
	return mustService(t, root, authorizedRequests(new(int), "sudo")), path
}

// mustReadSourceJSON performs an authorized source read and decodes its expected JSON response.
func mustReadSourceJSON[T any](t *testing.T, service *fileService, route string) T {
	t.Helper()
	response := mustProtectedRequest(t, service, http.MethodGet, route, "")
	if response.Code != http.StatusOK {
		t.Fatalf("read status = %d: %s", response.Code, response.Body.String())
	}
	var result T
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

// TestMediaAndStatRoutesRejectUnsupportedRequests checks both read-only input boundaries.
func TestMediaAndStatRoutesRejectUnsupportedRequests(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int), "sudo"))
	for _, test := range []struct {
		name, method, path string
		want               int
	}{
		{name: "metadata method", method: http.MethodPost, path: "/metadata/tmp/photo.jpg", want: http.StatusMethodNotAllowed},
		{name: "metadata query", method: http.MethodGet, path: "/metadata/tmp/photo.jpg?unexpected=1", want: http.StatusBadRequest},
		{name: "stat method", method: http.MethodPost, path: "/stat/tmp/photo.jpg", want: http.StatusMethodNotAllowed},
		{name: "stat query", method: http.MethodGet, path: "/stat/tmp/photo.jpg?unexpected=1", want: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := mustProtectedRequest(t, service, test.method, test.path, "")
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

// TestMediaMetadataReadsAuthorizedVideo covers metadata for sources without image dimensions.
func TestMediaMetadataReadsAuthorizedVideo(t *testing.T) {
	service, path := mustAuthorizedVideoSource(t)
	metadata := mustReadSourceJSON[mediaMetadata](t, service, "/metadata"+path)
	if metadata.MediaType != "VIDEO" {
		t.Fatalf("media type = %q, want VIDEO", metadata.MediaType)
	}
}

// TestMediaMetadataRejectsDirectories requires a regular source file.
func TestMediaMetadataRejectsDirectories(t *testing.T) {
	root := t.TempDir()
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	response := mustProtectedRequest(t, service, http.MethodGet, "/metadata"+root, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("metadata status = %d, want 404", response.Code)
	}
}

// TestMediaMetadataRejectsReaderFailure keeps unreadable image bytes private.
func TestMediaMetadataRejectsReaderFailure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(path, []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	service.readerCommand = func(context.Context, fileaccess.Subject, string) (io.ReadCloser, error) {
		return nil, errors.New("reader unavailable")
	}
	response := mustProtectedRequest(t, service, http.MethodGet, "/metadata"+path, "")
	if response.Code != http.StatusForbidden {
		t.Fatalf("metadata status = %d, want 403", response.Code)
	}
}

// TestStatResourceReturnsAuthorizedFileKind covers the file branch of source stat.
func TestStatResourceReturnsAuthorizedFileKind(t *testing.T) {
	service, path := mustAuthorizedVideoSource(t)
	result := mustReadSourceJSON[sourceStatForTest](t, service, "/stat"+path)
	if result.Kind != "file" {
		t.Fatalf("kind = %q, want file", result.Kind)
	}
}

// TestReadMetadataRoutesReportMissingSource keeps absent paths out of read responses.
func TestReadMetadataRoutesReportMissingSource(t *testing.T) {
	root := t.TempDir()
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	for _, route := range []string{"/stat", "/metadata"} {
		t.Run(route, func(t *testing.T) {
			response := mustProtectedRequest(t, service, http.MethodGet, route+filepath.Join(root, "missing.jpg"), "")
			if response.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", response.Code)
			}
		})
	}
}

// TestClassifyMediaCoversNonImageSources preserves extension classification for scans.
func TestClassifyMediaCoversNonImageSources(t *testing.T) {
	for _, test := range []struct{ path, want string }{
		{path: "sound.MP3", want: "AUDIO"},
		{path: "notes.txt", want: "OTHER"},
	} {
		t.Run(test.path, func(t *testing.T) {
			if got := classifyMedia(test.path); got != test.want {
				t.Fatalf("media type = %q, want %q", got, test.want)
			}
		})
	}
}

// TestMediaMetadataReadsAuthorizedDimensions covers moments-backend/FR-003.
func TestMediaMetadataReadsAuthorizedDimensions(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "photo.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	service.readerCommand = func(_ context.Context, _ fileaccess.Subject, path string) (io.ReadCloser, error) {
		return os.Open(path)
	}
	response := mustProtectedRequest(t, service, http.MethodGet, "/metadata"+path, "")
	if response.Code != http.StatusOK {
		t.Fatalf("metadata status = %d: %s", response.Code, response.Body.String())
	}
	var metadata mediaMetadata
	if err := json.Unmarshal(response.Body.Bytes(), &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.Width == nil || *metadata.Width != 2 {
		t.Fatalf("width = %v, want 2", metadata.Width)
	}
}

// TestStatResourceReturnsAuthorizedDirectoryKind covers moments-backend/FR-003.
func TestStatResourceReturnsAuthorizedDirectoryKind(t *testing.T) {
	root := t.TempDir()
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	response := mustProtectedRequest(t, service, http.MethodGet, "/stat"+root, "")
	if response.Code != http.StatusOK {
		t.Fatalf("stat status = %d: %s", response.Code, response.Body.String())
	}
	var result sourceStatForTest
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Kind != "directory" {
		t.Fatalf("kind = %q, want directory", result.Kind)
	}
}

// sourceStatForTest is the small public stat contract observed by callers.
type sourceStatForTest struct {
	Kind string `json:"kind"`
}
