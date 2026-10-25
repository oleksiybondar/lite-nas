package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/shared/fileaccess"
)

// streamService authenticates each media request and sends only authorized files.
type streamService struct {
	cfg           serviceConfig
	auth          *fileaccess.Authorizer
	open          func(string, os.FileInfo) (*os.File, error)
	transcodes    chan struct{}
	readAsUser    func(context.Context, fileaccess.Subject, string, os.FileInfo, int64, int64) (io.ReadCloser, error)
	transcodeFile func(http.ResponseWriter, *http.Request, fileaccess.Subject, string, os.FileInfo, int64)
	previewFile   func(http.ResponseWriter, *http.Request, fileaccess.Subject, string, os.FileInfo)
}

// mediaInput validates the route suffix before any file lookup.
type mediaInput struct {
	Path string `json:"path" minLength:"2" maxLength:"4096"`
}

// transcodeInput validates the seek position at the HTTP adapter boundary.
type transcodeInput struct {
	Start int64 `json:"start" minimum:"0" maximum:"86400"`
}

// ServeHTTP handles original media requests and repeats Auth and RBAC checks for ranges.
func (s *streamService) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Cache-Control", "private, no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	transcode := strings.HasPrefix(request.URL.Path, "/transcode/")
	preview := strings.HasPrefix(request.URL.Path, "/preview/")
	if request.Method != http.MethodGet && (request.Method != http.MethodHead || transcode || preview) {
		http.Error(writer, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}
	var path string
	var start int64
	var err error
	if preview {
		path, err = requestPreviewPath(request)
	} else {
		path, start, err = requestMediaPath(request, transcode)
	}
	if err != nil {
		http.Error(writer, "invalid_path", http.StatusBadRequest)
		return
	}
	subject, err := s.authenticate(request)
	if err != nil {
		status := http.StatusUnauthorized
		if !errors.Is(err, fileaccess.ErrUnauthenticated) && !errors.Is(err, http.ErrNoCookie) {
			status = http.StatusServiceUnavailable
		}
		http.Error(writer, http.StatusText(status), status)
		return
	}
	roots, err := s.readRoots()
	if err != nil {
		http.Error(writer, "storage_unavailable", http.StatusServiceUnavailable)
		return
	}
	if !s.inReadScope(path, roots) {
		http.NotFound(writer, request)
		return
	}
	resolved, err := fileaccess.ResolveExisting(path)
	if err != nil || !s.inReadScope(resolved, roots) {
		http.NotFound(writer, request)
		return
	}
	before, err := os.Stat(resolved)
	if err != nil || !before.Mode().IsRegular() {
		http.NotFound(writer, request)
		return
	}
	if err := s.auth.Read(request.Context(), subject, resolved); err != nil {
		if errors.Is(err, fileaccess.ErrDenied) {
			http.NotFound(writer, request)
		} else {
			http.Error(writer, "authorization_unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	if transcode && start != 0 && (strings.EqualFold(filepath.Ext(resolved), ".heic") ||
		strings.EqualFold(filepath.Ext(resolved), ".heif")) {
		http.Error(writer, "still_video_cannot_seek", http.StatusBadRequest)
		return
	}
	if preview && request.Header.Get("Range") != "" {
		http.Error(writer, "preview_does_not_support_ranges", http.StatusBadRequest)
		return
	}
	if !transcode && !preview {
		if _, _, _, rangeErr := singleRange(request.Header.Get("Range"), before.Size()); rangeErr != nil {
			writer.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", before.Size()))
			http.Error(writer, "invalid_range", http.StatusRequestedRangeNotSatisfiable)
			return
		}
	}
	if transcode {
		s.transcodeFile(writer, request, subject, resolved, before, start)
		return
	}
	if preview {
		s.previewFile(writer, request, subject, resolved, before)
		return
	}
	file, err := s.open(resolved, before)
	if err != nil {
		if errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM) {
			s.serveUserOriginal(writer, request, subject, resolved, before)
			return
		}
		http.NotFound(writer, request)
		return
	}
	defer file.Close()
	writer.Header().Set("Content-Type", mediaContentType(resolved))
	writer.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": filepath.Base(resolved)}))
	http.ServeContent(writer, request, filepath.Base(resolved), before.ModTime(), file)
}

// serveUserOriginal handles a private file using the verified owner's Unix ACL.
func (s *streamService) serveUserOriginal(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, path string, info os.FileInfo) {
	start, end, partial, err := singleRange(request.Header.Get("Range"), info.Size())
	if err != nil {
		writer.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", info.Size()))
		http.Error(writer, "invalid_range", http.StatusRequestedRangeNotSatisfiable)
		return
	}
	length := end - start + 1
	if info.Size() == 0 {
		length = 0
	}
	status := http.StatusOK
	if partial {
		status = http.StatusPartialContent
		writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
	}
	if request.Method == http.MethodHead || length == 0 {
		setOriginalHeaders(writer, path, length)
		writer.WriteHeader(status)
		return
	}
	reader, err := s.readAsUser(request.Context(), subject, path, info, start, length)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer reader.Close()
	var first [1]byte
	if _, err := io.ReadFull(reader, first[:]); err != nil {
		http.NotFound(writer, request)
		return
	}
	setOriginalHeaders(writer, path, length)
	writer.WriteHeader(status)
	_, _ = writer.Write(first[:])
	_, _ = io.CopyN(writer, reader, length-1)
}

// setOriginalHeaders prevents transformed content or request data from defining media metadata.
func setOriginalHeaders(writer http.ResponseWriter, path string, length int64) {
	writer.Header().Set("Content-Type", mediaContentType(path))
	writer.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": filepath.Base(path)}))
	writer.Header().Set("Accept-Ranges", "bytes")
	writer.Header().Set("Content-Length", strconv.FormatInt(length, 10))
}

// requestMediaPath validates the original or explicit transcoding URL contract.
func requestMediaPath(request *http.Request, transcode bool) (string, int64, error) {
	if !transcode {
		path, err := mediaPath(request)
		return path, 0, err
	}
	values, queryErr := url.ParseQuery(request.URL.RawQuery)
	if queryErr != nil || len(values) > 1 || len(values["start"]) > 1 ||
		len(values) == 1 && len(values["start"]) == 0 {
		return "", 0, fileaccess.ErrInvalidPath
	}
	start := int64(0)
	if value := values.Get("start"); value != "" {
		if len(value) > 5 || strings.Trim(value, "0123456789") != "" {
			return "", 0, fileaccess.ErrInvalidPath
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed > 86400 {
			return "", 0, fileaccess.ErrInvalidPath
		}
		start = parsed
	} else if _, present := values["start"]; present {
		return "", 0, fileaccess.ErrInvalidPath
	}
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)
	schema := registry.Schema(reflect.TypeOf(transcodeInput{}), true, "StreamTranscodeInput")
	result := &huma.ValidateResult{}
	huma.Validate(registry, schema, huma.NewPathBuffer(nil, 0), huma.ModeWriteToServer, map[string]any{"start": start}, result)
	if len(result.Errors) != 0 {
		return "", 0, fileaccess.ErrInvalidPath
	}
	clone := request.Clone(request.Context())
	clone.URL.Path = strings.TrimPrefix(request.URL.Path, "/transcode")
	clone.URL.RawPath = strings.TrimPrefix(request.URL.RawPath, "/transcode")
	clone.URL.RawQuery = ""
	path, err := mediaPath(clone)
	return path, start, err
}

// requestPreviewPath validates the source path for a converted still image.
func requestPreviewPath(request *http.Request) (string, error) {
	clone := request.Clone(request.Context())
	clone.URL.Path = strings.TrimPrefix(request.URL.Path, "/preview")
	clone.URL.RawPath = strings.TrimPrefix(request.URL.RawPath, "/preview")
	return mediaPath(clone)
}

// readRoots combines configured storage with mounted ZFS datasets for each request.
func (s *streamService) readRoots() ([]string, error) {
	zfsRoots, err := fileaccess.MountedZFSRoots()
	if err != nil {
		return nil, err
	}
	return append(append([]string(nil), s.cfg.ReadRoots...), zfsRoots...), nil
}

// inReadScope applies the same storage boundary to every caller, including admins.
func (s *streamService) inReadScope(path string, roots []string) bool {
	return fileaccess.WithinRoots(path, roots) && !fileaccess.WithinRoots(path, s.cfg.DeniedPaths)
}

// authenticate accepts a single Bearer token or the existing browser access cookie.
func (s *streamService) authenticate(request *http.Request) (fileaccess.Subject, error) {
	authorization := request.Header.Get("Authorization")
	if authorization != "" {
		if !strings.HasPrefix(authorization, "Bearer ") || strings.Count(authorization, " ") != 1 {
			return fileaccess.Subject{}, fileaccess.ErrUnauthenticated
		}
		return s.auth.Resolve(request.Context(), strings.TrimPrefix(authorization, "Bearer "))
	}
	cookie, err := request.Cookie("lite-nas-at")
	if err != nil {
		return fileaccess.Subject{}, err
	}
	return s.auth.Resolve(request.Context(), cookie.Value)
}

// mediaPath maps Nginx's stripped /s/ suffix to one canonical host path.
func mediaPath(request *http.Request) (string, error) {
	escaped := request.URL.EscapedPath()
	lower := strings.ToLower(escaped)
	if strings.Contains(lower, "%2f") || strings.Contains(lower, "%5c") {
		return "", fileaccess.ErrInvalidPath
	}
	path, err := url.PathUnescape(escaped)
	if err != nil || fileaccess.ValidateHostPath(path) != nil || strings.Contains(path, "\\") || request.URL.RawQuery != "" {
		return "", fileaccess.ErrInvalidPath
	}
	registry := huma.NewMapRegistry("#/components/schemas/", huma.DefaultSchemaNamer)
	schema := registry.Schema(reflect.TypeOf(mediaInput{}), true, "StreamMediaInput")
	result := &huma.ValidateResult{}
	huma.Validate(registry, schema, huma.NewPathBuffer(nil, 0), huma.ModeWriteToServer, map[string]any{"path": path}, result)
	if len(result.Errors) != 0 {
		return "", fmt.Errorf("media path validation: %w", fileaccess.ErrInvalidPath)
	}
	return path, nil
}

// mediaContentType uses a passive media allowlist; unknown formats remain downloadable bytes.
func mediaContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".heic":
		return "image/heic"
	case ".heif":
		return "image/heif"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
