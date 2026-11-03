package main

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"lite-nas/shared/fileaccess"
)

// mediaTag is a source-backed value extracted from a media file.
type mediaTag struct {
	Category string `json:"category"`
	Value    string `json:"value"`
}

// mediaMetadata is the caller-authorized metadata contract for one source file.
type mediaMetadata struct {
	Path        string       `json:"path"`
	MediaType   string       `json:"mediaType"`
	ContentType string       `json:"contentType"`
	Width       *int         `json:"width"`
	Height      *int         `json:"height"`
	Location    *gpsLocation `json:"location,omitempty"`
	Tags        []mediaTag   `json:"tags"`
}

// classifyMedia distinguishes source media classes without claiming a Live Photo pair.
func classifyMedia(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".heic", ".heif", ".tif", ".tiff", ".bmp":
		return "IMAGE"
	case ".mov", ".mp4", ".m4v", ".avi", ".mkv", ".webm", ".ogv", ".mpeg", ".mpg":
		return "VIDEO"
	case ".mp3", ".wav", ".m4a", ".aac", ".flac", ".ogg":
		return "AUDIO"
	default:
		return "OTHER"
	}
}

// mediaMetadataResource authenticates and authorizes every metadata extraction.
func (s *fileService) mediaMetadataResource(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if request.URL.RawQuery != "" {
		writeError(writer, http.StatusBadRequest, "invalid_query")
		return
	}
	resolved, scope, ok := s.resolveReadScope(writer, request, subject, "/metadata")
	if !ok {
		return
	}
	if !scope.AllowsFile(resolved) {
		decisionError(writer, fileaccess.ErrDenied)
		return
	}
	if err := s.auth.Read(request.Context(), subject, resolved); err != nil {
		decisionError(writer, err)
		return
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	result := mediaMetadata{
		Path: resolved, MediaType: classifyMedia(resolved),
		ContentType: safeContentType(resolved, false), Tags: []mediaTag{},
	}
	if result.MediaType == "IMAGE" {
		reader, err := s.readerCommand(request.Context(), subject, resolved)
		if err != nil {
			decisionError(writer, fileaccess.ErrDenied)
			return
		}
		defer reader.Close()
		data, readErr := io.ReadAll(io.LimitReader(reader, 16<<20))
		if readErr != nil {
			writeError(writer, http.StatusServiceUnavailable, "metadata_unavailable")
			return
		}
		imageInfo, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			result.Width, result.Height = &imageInfo.Width, &imageInfo.Height
		}
		result.Tags = jpegEXIFTags(data)
		result.Location = jpegGPSLocation(data)
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "private, no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	_ = json.NewEncoder(writer).Encode(result)
}
