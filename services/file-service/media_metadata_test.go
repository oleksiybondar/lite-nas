package main

import (
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/fileaccess"
)

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
