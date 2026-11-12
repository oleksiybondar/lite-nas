package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"lite-nas/shared/fileaccess"
)

// preview converts one authorized HEIC still to a bounded-size JPEG response.
func (s *streamService) preview(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, path string, info os.FileInfo) {
	extension := strings.ToLower(filepath.Ext(path))
	if extension != ".heic" && extension != ".heif" {
		http.Error(writer, "preview_requires_heic", http.StatusUnsupportedMediaType)
		return
	}
	select {
	case s.transcodes <- struct{}{}:
		defer func() { <-s.transcodes }()
	default:
		http.Error(writer, "preview_busy", http.StatusTooManyRequests)
		return
	}
	reader, err := startUserPreview(request.Context(), subject, path, info)
	if err != nil {
		http.Error(writer, "preview_unavailable", http.StatusServiceUnavailable)
		return
	}
	defer reader.Close()
	var first [32768]byte
	count, err := reader.Read(first[:])
	if count == 0 || err != nil && err != io.EOF {
		http.Error(writer, "preview_failed", http.StatusUnprocessableEntity)
		return
	}
	writer.Header().Set("Content-Type", "image/jpeg")
	writer.Header().Set("Content-Disposition", "inline")
	writer.Header().Set("X-Accel-Buffering", "no")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(first[:count])
	_, _ = io.Copy(writer, reader)
}

// startUserPreview executes only the fixed converter under the RBAC-matched Unix user.
func startUserPreview(ctx context.Context, subject fileaccess.Subject, path string, info os.FileInfo) (io.ReadCloser, error) {
	if fileaccess.ValidateHostPath(path) != nil {
		return nil, fileaccess.ErrInvalidPath
	}
	username, err := fileaccess.VerifyUnixIdentity(subject)
	if err != nil {
		return nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Ino == 0 {
		return nil, errors.New("file identity unavailable")
	}
	command := exec.CommandContext(ctx, "/usr/bin/sudo", "-n", "-u", username, "--",
		"/usr/libexec/lite-nas/stream-service", "--preview-file", path,
		strconv.FormatUint(uint64(stat.Dev), 10), strconv.FormatUint(stat.Ino, 10))
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	output, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := command.Start(); err != nil {
		_ = output.Close()
		return nil, err
	}
	return &readerProcess{ReadCloser: output, command: command}, nil
}

// runUserPreview checks the approved inode before starting the host decoder.
func runUserPreview(args []string) error {
	if os.Geteuid() == 0 || len(args) != 3 {
		return errors.New("preview must run as a non-root user with fixed operands")
	}
	input, err := parseReaderInput(append(append([]string(nil), args...), "0", "0"))
	if err != nil {
		return err
	}
	extension := strings.ToLower(filepath.Ext(input.Path))
	if extension != ".heic" && extension != ".heif" {
		return fileaccess.ErrInvalidPath
	}
	if err := requireLiveParent(); err != nil {
		return err
	}
	file, err := openExpectedFile(input.Path, input.Device, input.Inode)
	if err != nil {
		return err
	}
	defer file.Close()
	return convertHEICFile(context.Background(), file, os.Stdout)
}

// convertHEICFile assembles a HEIC still on private tmpfs and copies its JPEG bytes.
func convertHEICFile(ctx context.Context, file *os.File, output io.Writer) error {
	// heif-dec composes HEIC grids; ffmpeg exposes their individual tiles.
	// The service's private, size-limited tmpfs holds the JPEG only for this request.
	directory, err := os.MkdirTemp("/tmp", "lite-nas-preview-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	previewPath := filepath.Join(directory, "preview.jpg")
	command := exec.CommandContext(ctx, "/usr/bin/heif-dec", "--quiet", "--codec-threads", "2",
		"-q", "80", "--", "/proc/self/fd/0", previewPath)
	command.Stdin = file
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	command.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	if err := command.Run(); err != nil {
		return err
	}
	preview, err := os.Open(previewPath)
	if err != nil {
		return err
	}
	defer preview.Close()
	info, err := preview.Stat()
	if err != nil || info.Size() == 0 || info.Size() > 32<<20 {
		return errors.New("preview size is unavailable or exceeds the response limit")
	}
	var start, end [2]byte
	if _, err := preview.ReadAt(start[:], 0); err != nil {
		return err
	}
	if _, err := preview.ReadAt(end[:], info.Size()-2); err != nil {
		return err
	}
	if start != [2]byte{0xff, 0xd8} || end != [2]byte{0xff, 0xd9} {
		return errors.New("preview decoder did not produce a complete JPEG")
	}
	_, err = io.Copy(output, preview)
	return err
}
