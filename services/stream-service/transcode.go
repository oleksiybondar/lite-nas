package main

import (
	"bytes"
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

// transcode streams a verified user's H.264/AAC fragmented MP4. HEIC decoding
// uses the bounded private tmpfs before ffmpeg starts. Each invocation follows
// its own Auth and RBAC approval.
func (s *streamService) transcode(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject, path string, info os.FileInfo, start int64) {
	select {
	case s.transcodes <- struct{}{}:
		defer func() { <-s.transcodes }()
	default:
		http.Error(writer, "transcode_busy", http.StatusTooManyRequests)
		return
	}
	reader, err := startUserTranscoder(request.Context(), subject, path, info, start)
	if err != nil {
		http.Error(writer, "transcode_unavailable", http.StatusServiceUnavailable)
		return
	}
	defer reader.Close()
	var first [32768]byte
	count, err := reader.Read(first[:])
	if count == 0 || err != nil && err != io.EOF {
		http.Error(writer, "transcode_failed", http.StatusUnprocessableEntity)
		return
	}
	writer.Header().Set("Content-Type", "video/mp4")
	// Browsers may send Range on an initial video request; a 200 response
	// supplies the sequential stream while time seeks use a fresh URL.
	writer.Header().Set("Accept-Ranges", "none")
	writer.Header().Set("Content-Disposition", "inline")
	writer.Header().Set("X-Accel-Buffering", "no")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(first[:count])
	_, _ = io.Copy(writer, reader)
}

// startUserTranscoder executes only the fixed helper as the Auth/RBAC-matched
// Unix account. The approved inode and time are validated again in the helper.
func startUserTranscoder(ctx context.Context, subject fileaccess.Subject, path string, info os.FileInfo, start int64) (io.ReadCloser, error) {
	if fileaccess.ValidateHostPath(path) != nil || start < 0 || start > 86400 {
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
		"/usr/libexec/lite-nas/stream-service", "--transcode-file", path,
		strconv.FormatUint(uint64(stat.Dev), 10), strconv.FormatUint(stat.Ino, 10),
		strconv.FormatInt(start, 10))
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

// runUserTranscoder validates CLI operands, opens the approved inode, and runs
// the host codec using a seekable descriptor visible only inside this process.
func runUserTranscoder(args []string) error {
	if os.Geteuid() == 0 || len(args) != 4 {
		return errors.New("transcoder must run as a non-root user with fixed operands")
	}
	input, err := parseReaderInput(append(append([]string(nil), args...), "0"))
	if err != nil || input.Offset > 86400 {
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
	extension := strings.ToLower(filepath.Ext(input.Path))
	if extension == ".heic" || extension == ".heif" {
		if input.Offset != 0 {
			return fileaccess.ErrInvalidPath
		}
		return convertHEICToVideo(context.Background(), file, os.Stdout)
	}
	return convertFile(context.Background(), file, input.Offset, os.Stdout)
}

// convertHEICToVideo composes the still before presenting it as a short MP4.
// A HEIC without motion frames cannot become a moving Live Photo.
func convertHEICToVideo(ctx context.Context, file *os.File, output io.Writer) error {
	var jpeg bytes.Buffer
	if err := convertHEICFile(ctx, file, &jpeg); err != nil {
		return err
	}
	return convertJPEGToVideo(ctx, jpeg.Bytes(), output)
}

// convertJPEGToVideo sends three still frames through the fixed image demuxer.
func convertJPEGToVideo(ctx context.Context, jpeg []byte, output io.Writer) error {
	if len(jpeg) == 0 || len(jpeg) > 32<<20 {
		return errors.New("JPEG frame is unavailable or too large")
	}
	command := exec.CommandContext(ctx, "/usr/bin/ffmpeg",
		"-hide_banner", "-loglevel", "error", "-nostdin", "-max_alloc", "134217728",
		"-filter_threads", "2", "-threads", "2",
		"-framerate", "1", "-f", "image2pipe", "-c:v", "mjpeg", "-i", "pipe:0", "-frames:v", "3",
		"-vf", "scale=1280:1280:force_original_aspect_ratio=decrease:force_divisible_by=2",
		"-threads", "2", "-c:v", "libx264", "-preset", "veryfast", "-crf", "24", "-pix_fmt", "yuv420p",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof", "-f", "mp4", "pipe:1")
	command.Stdin = io.MultiReader(bytes.NewReader(jpeg), bytes.NewReader(jpeg), bytes.NewReader(jpeg))
	command.Stdout = output
	command.Stderr = io.Discard
	command.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	return command.Run()
}

// convertFile keeps the original seekable and pipes only converted bytes.
// Its fixed demuxer list excludes playlist formats that can open other files.
func convertFile(ctx context.Context, file *os.File, start int64, output io.Writer) error {
	args := []string{
		"-hide_banner", "-loglevel", "error", "-nostdin", "-max_alloc", "134217728",
		"-filter_threads", "2", "-threads", "2",
		"-ss", strconv.FormatInt(start, 10),
		"-protocol_whitelist", "file",
		"-format_whitelist", "avi,mov,matroska,webm,mpegts,mpeg,asf,flv,ogg",
		"-i", "/proc/self/fd/0",
		"-map", "0:v:0", "-map", "0:a:0?",
		"-threads", "2", "-c:v", "libx264", "-preset", "veryfast", "-crf", "24", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "128k",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof", "-f", "mp4", "pipe:1",
	}
	command := exec.CommandContext(ctx, "/usr/bin/ffmpeg", args...)
	command.Stdin = file
	command.Stdout = output
	command.Stderr = io.Discard
	command.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	return command.Run()
}
