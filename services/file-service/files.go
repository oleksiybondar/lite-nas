package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"lite-nas/services/file-service/filesystem"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

// rangeLimiter bounds simultaneous responses for each principal and file.
type rangeLimiter struct {
	mu     sync.Mutex
	active map[string]int
	limit  int
}

// newRangeLimiter constructs a request-scoped stream admission tracker.
func newRangeLimiter(limit int) *rangeLimiter {
	return &rangeLimiter{active: make(map[string]int), limit: limit}
}

// acquire reserves one response slot until its release callback is used.
func (limiter *rangeLimiter) acquire(username string, path string) (func(), bool) {
	key := username + "\x00" + path
	limiter.mu.Lock()
	if limiter.active[key] >= limiter.limit {
		limiter.mu.Unlock()
		return nil, false
	}
	limiter.active[key]++
	limiter.mu.Unlock()
	return func() {
		limiter.mu.Lock()
		limiter.active[key]--
		if limiter.active[key] == 0 {
			delete(limiter.active, key)
		}
		limiter.mu.Unlock()
	}, true
}

// fileIdentity returns the stable host identity of a statted file.
func fileIdentity(info os.FileInfo) (uint64, uint64) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0
	}
	return uint64(stat.Dev), stat.Ino
}

// hostPath maps a decoded route suffix to one unambiguous absolute host path.
func hostPath(request *http.Request, prefix string) (string, error) {
	escaped := request.URL.EscapedPath()
	if strings.Contains(strings.ToLower(escaped), "%2f") || strings.Contains(strings.ToLower(escaped), "%5c") {
		return "", fileaccess.ErrInvalidPath
	}
	decoded, err := url.PathUnescape(escaped)
	if err != nil || !strings.HasPrefix(decoded, prefix) {
		return "", fileaccess.ErrInvalidPath
	}
	path := strings.TrimPrefix(decoded, prefix)
	if !strings.HasPrefix(path, "/") || fileaccess.ValidateHostPath(path) != nil {
		return "", fileaccess.ErrInvalidPath
	}
	return path, nil
}

// download authenticates and authorizes a new full or one-range file response.
func (s *fileService) download(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, attachment bool) {
	prefix := ""
	if attachment {
		prefix = "/download"
	}
	requested, err := hostPath(request, prefix)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_path")
		return
	}
	resolved, err := fileaccess.ResolveExisting(requested)
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	scope, err := s.policy.ReadScope(subject.Groups)
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "authorization_unavailable")
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
	release, ok := s.ranges.acquire(subject.Username, resolved)
	if !ok {
		writeError(writer, http.StatusTooManyRequests, "download_limit")
		return
	}
	defer release()
	device, inode := fileIdentity(info)
	etag := fmt.Sprintf("\"%x-%x-%x-%x\"", device, inode, info.Size(), info.ModTime().UnixNano())
	start, end, partial, rangeErr := parseRange(request.Header.Get("Range"), info.Size())
	if rangeErr != nil {
		writer.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", info.Size()))
		writeError(writer, http.StatusRequestedRangeNotSatisfiable, "invalid_range")
		return
	}
	if request.Header.Get("If-Range") != "" && request.Header.Get("If-Range") != etag {
		start, end, partial = 0, info.Size()-1, false
	}
	if request.Header.Get("If-Match") != "" && request.Header.Get("If-Match") != etag {
		writeError(writer, http.StatusPreconditionFailed, "file_changed")
		return
	}
	length := end - start + 1
	if info.Size() == 0 {
		length = 0
	}
	status := http.StatusOK
	if partial {
		status = http.StatusPartialContent
	}
	var stream io.ReadCloser
	var first [1]byte
	if request.Method == http.MethodGet {
		stream, err = openCat(request.Context(), subject, resolved)
		if err != nil {
			if s.logger != nil {
				s.logger.Warn("file reader start failed", "cause", err)
			}
			writeError(writer, http.StatusForbidden, "read_failed")
			return
		}
		defer stream.Close()
		if _, err := io.CopyN(io.Discard, stream, start); err != nil {
			if s.logger != nil {
				s.logger.Warn("file reader seek failed", "cause", err)
			}
			writeError(writer, http.StatusForbidden, "read_failed")
			return
		}
		count, readErr := stream.Read(first[:])
		if (length == 0 && (count != 0 || readErr != io.EOF)) ||
			(length > 0 && count != 1) {
			if s.logger != nil {
				s.logger.Warn("file reader first byte failed", "cause", readErr)
			}
			writeError(writer, http.StatusForbidden, "read_failed")
			return
		}
	}
	writer.Header().Set("ETag", etag)
	writer.Header().Set("Accept-Ranges", "bytes")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	writer.Header().Set("Content-Type", safeContentType(resolved, attachment))
	disposition := "inline"
	if attachment || writer.Header().Get("Content-Type") == "application/octet-stream" {
		disposition = "attachment"
	}
	writer.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": filepath.Base(resolved)}))
	if partial {
		writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
	}
	if request.Method == http.MethodHead {
		writer.WriteHeader(status)
		return
	}
	if length == 0 {
		writer.WriteHeader(status)
		return
	}
	writer.WriteHeader(status)
	_, _ = writer.Write(first[:])
	_, _ = io.CopyN(writer, stream, length-1)
}

// parseRange accepts one RFC 9110 byte range or the complete representation.
func parseRange(header string, size int64) (int64, int64, bool, error) {
	if header == "" {
		return 0, size - 1, false, nil
	}
	if !strings.HasPrefix(header, "bytes=") || strings.Contains(header, ",") || size == 0 {
		return 0, 0, false, errors.New("invalid range")
	}
	parts := strings.Split(strings.TrimPrefix(header, "bytes="), "-")
	if len(parts) != 2 {
		return 0, 0, false, errors.New("invalid range")
	}
	if parts[0] == "" {
		count, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || count <= 0 {
			return 0, 0, false, errors.New("invalid suffix")
		}
		if count > size {
			count = size
		}
		return size - count, size - 1, true, nil
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false, errors.New("invalid start")
	}
	end := size - 1
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil || end < start {
			return 0, 0, false, errors.New("invalid end")
		}
		if end >= size {
			end = size - 1
		}
	}
	return start, end, true, nil
}

// safeContentType permits a small passive inline set and attaches everything else.
func safeContentType(path string, attachment bool) string {
	if attachment {
		return "application/octet-stream"
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

// list returns only entries whose paths receive a fresh RBAC read allow.
func (s *fileService) list(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	requested, err := hostPath(request, "/ls")
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_path")
		return
	}
	resolved, err := fileaccess.ResolveExisting(requested)
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	scope, err := s.policy.ReadScope(subject.Groups)
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "authorization_unavailable")
		return
	}
	if !scope.AllowsDirectory(resolved) {
		decisionError(writer, fileaccess.ErrDenied)
		return
	}
	if err := s.auth.Read(request.Context(), subject, resolved); err != nil {
		decisionError(writer, err)
		return
	}
	files, directories := make([]string, 0), make([]string, 0)
	recursive := request.URL.Query().Has("recursive")
	filesOnly := request.URL.Query().Has("files-only")
	var visit func(string) error
	visit = func(directory string) error {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			path := filepath.Join(directory, entry.Name())
			if entry.IsDir() {
				if !scope.AllowsDirectory(path) {
					continue
				}
			} else if !scope.AllowsFile(path) {
				continue
			}
			if err := s.auth.Read(request.Context(), subject, path); err != nil {
				if errors.Is(err, fileaccess.ErrDenied) {
					continue
				}
				return err
			}
			if entry.IsDir() {
				if !filesOnly {
					directories = append(directories, path)
				}
				if recursive {
					if err := visit(path); err != nil {
						return err
					}
				}
			} else if entry.Type().IsRegular() {
				files = append(files, path)
			}
		}
		return nil
	}
	if err := visit(resolved); err != nil {
		writeError(writer, http.StatusServiceUnavailable, "listing_failed")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"files": files, "directories": directories})
}

// deleteFile checks policy and RBAC before one user-scoped regular-file removal.
func (s *fileService) deleteFile(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	requested, err := hostPath(request, "")
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_path")
		return
	}
	requestedInfo, err := os.Lstat(requested)
	if err == nil && requestedInfo.Mode()&os.ModeSymlink != 0 {
		writeError(writer, http.StatusBadRequest, "unsupported_target")
		return
	}
	resolved, err := fileaccess.ResolveExisting(requested)
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	if err := s.policy.AllowMutation(resolved); err != nil {
		writeError(writer, http.StatusForbidden, "protected_path")
		return
	}
	info, err := os.Lstat(resolved)
	if err != nil || (!info.Mode().IsRegular() && !info.IsDir()) {
		writeError(writer, http.StatusBadRequest, "unsupported_target")
		return
	}
	if err := s.auth.Operation(request.Context(), subject, resolved, rbaccontract.PathOperationDelete); err != nil {
		decisionError(writer, err)
		return
	}
	if info.IsDir() {
		job, err := s.jobs.enqueue(subject, "delete", resolved, "")
		if err != nil {
			writeError(writer, http.StatusServiceUnavailable, "queue_unavailable")
			return
		}
		writeJSON(writer, http.StatusAccepted, job)
		return
	}
	username, err := filesystem.VerifyUnixIdentity(subject)
	if err != nil {
		writeError(writer, http.StatusForbidden, "identity_mismatch")
		return
	}
	command := exec.CommandContext(request.Context(), "/usr/bin/sudo", "-n", "-u", username,
		"--", "/bin/rm", "--", resolved)
	if err := command.Run(); err != nil {
		writeError(writer, http.StatusConflict, "delete_failed")
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
