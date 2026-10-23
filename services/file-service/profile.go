package main

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"lite-nas/services/file-service/filesystem"
	"lite-nas/shared/fileaccess"
)

const (
	profileDirectoryName  = ".lite-nas"
	maxProfileAvatarBytes = 5 << 20
)

// profileDocumentInput is the bounded Huma-validated JSON envelope for one settings file.
type profileDocumentInput struct {
	Document map[string]any `json:"document" required:"true"`
}

// profilePath binds fixed profile filenames to the authenticated Unix account.
func profilePath(subject fileaccess.Subject, name string) (string, error) {
	if _, err := filesystem.VerifyUnixIdentity(subject); err != nil {
		return "", err
	}
	account, err := user.Lookup(subject.Username)
	if err != nil || account.HomeDir != filepath.Join("/home", subject.Username) {
		return "", fileaccess.ErrInvalidPath
	}
	return filepath.Join(account.HomeDir, profileDirectoryName, name), nil
}

// sharedAvatarOwner recognizes only the fixed avatar path in one Unix home.
func sharedAvatarOwner(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) != 5 || parts[0] != "" || parts[1] != "home" ||
		parts[2] == "" || parts[3] != profileDirectoryName || parts[4] != "avatar.image" {
		return ""
	}
	return parts[2]
}

// serveSharedAvatar permits authenticated cross-user avatar reads only for a world-readable owner file.
func (s *fileService) serveSharedAvatar(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, path string) bool {
	owner := sharedAvatarOwner(path)
	if owner == "" || owner == subject.Username {
		return false
	}
	account, err := user.Lookup(owner)
	if err != nil || account.HomeDir != filepath.Join("/home", owner) {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	ownedDirectory, directoryErr := profileDirectoryOwnedBy(filepath.Dir(path), fileaccess.Subject{UID: account.Uid})
	if directoryErr != nil || !ownedDirectory {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	ownerUID, parseErr := strconv.ParseUint(account.Uid, 10, 32)
	if parseErr != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	fd, err := unix.Openat2(unix.AT_FDCWD, path, &unix.OpenHow{
		Flags:   unix.O_RDONLY | unix.O_CLOEXEC,
		Resolve: unix.RESOLVE_NO_SYMLINKS,
	})
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxProfileAvatarBytes || info.Mode().Perm()&0o004 == 0 {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || uint64(stat.Uid) != ownerUID {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	var signature [512]byte
	count, _ := file.Read(signature[:])
	mediaType := http.DetectContentType(signature[:count])
	if mediaType != "image/png" && mediaType != "image/jpeg" {
		writeError(writer, http.StatusNotFound, "not_found")
		return true
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeError(writer, http.StatusForbidden, "read_failed")
		return true
	}
	writer.Header().Set("Content-Type", mediaType)
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(writer, request, "avatar.image", info.ModTime(), file)
	return true
}

// profileResource serves the current account's fixed profile files without browser-supplied paths.
func (s *fileService) profileResource(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	var name string
	switch request.URL.Path {
	case "/profile/ui":
		name = "ui.json"
	case "/profile/admin-panel":
		name = "admin-panel.json"
	case "/profile/avatar":
		name = "avatar.image"
	default:
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	path, err := profilePath(subject, name)
	if err != nil {
		decisionError(writer, err)
		return
	}
	if request.Method == http.MethodGet || request.Method == http.MethodHead {
		exists, directoryErr := profileDirectoryOwnedBy(filepath.Dir(path), subject)
		if directoryErr != nil {
			decisionError(writer, directoryErr)
			return
		}
		if !exists {
			writeError(writer, http.StatusNotFound, "not_found")
			return
		}
		resolved := request.Clone(request.Context())
		resolved.URL.Path = path
		resolved.URL.RawPath = ""
		s.download(writer, resolved, subject, false)
		return
	}
	if request.Method != http.MethodPut {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	var data []byte
	if name == "avatar.image" {
		request.Body = http.MaxBytesReader(writer, request.Body, maxProfileAvatarBytes)
		data, err = io.ReadAll(request.Body)
		if err != nil {
			writeError(writer, http.StatusRequestEntityTooLarge, "avatar_too_large")
			return
		}
		if !validAvatar(data) {
			writeError(writer, http.StatusBadRequest, "invalid_avatar")
			return
		}
	} else {
		var input profileDocumentInput
		if !decodeOperationDTO(writer, request, &input) || input.Document == nil {
			writeError(writer, http.StatusBadRequest, "invalid_profile_document")
			return
		}
		data, err = json.Marshal(input.Document)
		if err != nil || len(data) > 8192 {
			writeError(writer, http.StatusBadRequest, "invalid_profile_document")
			return
		}
	}
	if err := s.ensureProfileDirectory(request, subject, filepath.Dir(path)); err != nil {
		if s.logger != nil {
			s.logger.Warn("profile directory setup failed", "cause", err)
		}
		decisionError(writer, err)
		return
	}
	s.uploadProfileBytes(writer, request, subject, path, data, name == "avatar.image")
}

// validAvatar fully decodes a bounded PNG or JPEG before it can be published.
func validAvatar(data []byte) bool {
	mediaType := http.DetectContentType(data)
	if mediaType != "image/png" && mediaType != "image/jpeg" {
		return false
	}
	configuration, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || configuration.Width < 1 || configuration.Height < 1 ||
		configuration.Width > 4096 || configuration.Height > 4096 ||
		(format != "png" && format != "jpeg") {
		return false
	}
	_, decodedFormat, err := image.Decode(bytes.NewReader(data))
	return err == nil && decodedFormat == format
}

// ensureProfileDirectory creates the owner's known-path directory once as that Unix user.
func (s *fileService) ensureProfileDirectory(request *http.Request, subject fileaccess.Subject, directory string) error {
	exists, err := profileDirectoryOwnedBy(directory, subject)
	if err != nil || exists {
		return err
	}
	resolved, err := resolveNewDirectory(directory)
	if err != nil || resolved != directory {
		return profileDirectoryRaceResult(directory, subject, fileaccess.ErrInvalidPath)
	}
	if err := s.allowDestination(request.Context(), subject, directory, true); err != nil {
		return err
	}
	if err := s.userCommand(request.Context(), subject, "/bin/mkdir", s.policy.DirectoryMode(directory, false), directory); err != nil {
		return profileDirectoryRaceResult(directory, subject, err)
	}
	return nil
}

// profileDirectoryOwnedBy accepts only a real directory owned by the current Unix user.
func profileDirectoryOwnedBy(directory string, subject fileaccess.Subject) (bool, error) {
	info, err := os.Lstat(directory)
	if err == nil {
		resolved, resolveErr := fileaccess.ResolveExisting(directory)
		uid, parseErr := strconv.ParseUint(subject.UID, 10, 32)
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !info.IsDir() || !ok || resolveErr != nil || resolved != directory || parseErr != nil || uint64(stat.Uid) != uid {
			return false, fileaccess.ErrDenied
		}
		return true, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}
	return false, nil
}

// profileDirectoryRaceResult accepts another successful profile request's matching directory.
func profileDirectoryRaceResult(directory string, subject fileaccess.Subject, original error) error {
	exists, err := profileDirectoryOwnedBy(directory, subject)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return original
}

// uploadProfileBytes uses the existing staged upload and authorization pipeline for a fixed destination.
func (s *fileService) uploadProfileBytes(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, path string, data []byte, avatar bool) {
	var body bytes.Buffer
	parts := multipart.NewWriter(&body)
	if err := parts.WriteField("destination", path); err != nil {
		writeError(writer, http.StatusInternalServerError, "profile_upload_failed")
		return
	}
	if err := parts.WriteField("private", strconv.FormatBool(!avatar)); err != nil {
		writeError(writer, http.StatusInternalServerError, "profile_upload_failed")
		return
	}
	filePart, err := parts.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "profile_upload_failed")
		return
	}
	if _, err := filePart.Write(data); err != nil {
		writeError(writer, http.StatusInternalServerError, "profile_upload_failed")
		return
	}
	if err := parts.Close(); err != nil {
		writeError(writer, http.StatusInternalServerError, "profile_upload_failed")
		return
	}
	forwarded := request.Clone(request.Context())
	forwarded.Body = io.NopCloser(&body)
	forwarded.Header = request.Header.Clone()
	forwarded.Header.Set("Content-Type", parts.FormDataContentType())
	forwarded.ContentLength = int64(body.Len())
	s.conventionalUploadWithAvatarMode(writer, forwarded, subject, avatar)
}
