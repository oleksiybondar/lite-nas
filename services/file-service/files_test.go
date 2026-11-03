package main

import (
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSafeContentTypeIncludesPassiveMedia covers file-service-svc/FR-007.
func TestSafeContentTypeIncludesPassiveMedia(t *testing.T) {
	for _, test := range []struct {
		path string
		want string
	}{
		{path: "/movie.mp4", want: "video/mp4"},
		{path: "/movie.m4v", want: "video/mp4"},
		{path: "/movie.mov", want: "video/quicktime"},
		{path: "/movie.avi", want: "video/x-msvideo"},
		{path: "/movie.mkv", want: "video/x-matroska"},
		{path: "/movie.webm", want: "video/webm"},
		{path: "/movie.ogv", want: "video/ogg"},
		{path: "/sound.mp3", want: "audio/mpeg"},
		{path: "/sound.wav", want: "audio/wav"},
		{path: "/sound.m4a", want: "audio/mp4"},
		{path: "/sound.aac", want: "audio/aac"},
		{path: "/sound.flac", want: "audio/flac"},
		{path: "/sound.ogg", want: "audio/ogg"},
	} {
		t.Run(test.path, func(t *testing.T) {
			if got := safeContentType(test.path, false); got != test.want {
				t.Fatalf("safeContentType(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

// TestDownloadQuotesHeaderSyntaxInFilename covers file-service-svc/FR-003 response headers.
func TestDownloadQuotesHeaderSyntaxInFilename(t *testing.T) {
	root := t.TempDir()
	name := "report\r\nX-Injected: yes.txt"
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("safe"), 0o600); err != nil {
		t.Fatal(err)
	}
	service, _ := mustWorkflowService(t, root)
	service.readerCommand = openWorkflowFile
	escaped := strings.ReplaceAll(strings.ReplaceAll(path, "\r", "%0D"), "\n", "%0A")
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.URL.Path = path
	request.URL.RawPath = escaped
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("download = %d %q, want 200", response.Code, response.Body.String())
	}
	_, parameters, err := mime.ParseMediaType(response.Header().Get("Content-Disposition"))
	if err != nil {
		t.Fatalf("parse content disposition: %v", err)
	}
	if parameters["filename"] != name {
		t.Fatalf("content disposition filename = %q, want %q", parameters["filename"], name)
	}
}

// TestDownloadServesPassiveMediaInline covers file-service-svc/FR-007.
func TestDownloadServesPassiveMediaInline(t *testing.T) {
	for _, test := range []struct {
		name         string
		mediaType    string
		requestRange string
		status       int
		contentRange string
	}{
		{name: "movie.mp4", mediaType: "video/mp4", status: http.StatusOK},
		{name: "sound.mp3", mediaType: "audio/mpeg", status: http.StatusOK},
		{name: "movie.webm", mediaType: "video/webm", requestRange: "bytes=0-4", status: http.StatusPartialContent, contentRange: "bytes 0-4/11"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, test.name)
			if err := os.WriteFile(path, []byte("media bytes"), 0o600); err != nil {
				t.Fatal(err)
			}
			service, _ := mustWorkflowService(t, root)
			service.readerCommand = openWorkflowFile
			response := mustWorkflowRequest(t, service, http.MethodGet, path, "", map[string]string{"Range": test.requestRange})
			got := struct {
				status       int
				mediaType    string
				disposition  string
				contentRange string
			}{response.Code, response.Header().Get("Content-Type"), strings.Split(response.Header().Get("Content-Disposition"), ";")[0], response.Header().Get("Content-Range")}
			want := struct {
				status       int
				mediaType    string
				disposition  string
				contentRange string
			}{test.status, test.mediaType, "inline", test.contentRange}
			if got != want {
				t.Fatalf("media response headers = %#v, want %#v", got, want)
			}
		})
	}
}
