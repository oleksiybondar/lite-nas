package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	"lite-nas/shared/fileaccess"
)

// mustStreamHEIC returns the absolute path of the synthetic decoder fixture.
func mustStreamHEIC(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", "blue-64.HEIC"))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Requirements: stream-service-svc/FR-005 and TR-001.
func TestFFmpegConvertsOriginalVideoToFragmentedMP4(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("host ffmpeg is not installed")
	}
	for _, format := range []string{"avi", "mov"} {
		t.Run(format, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "source."+format)
			command := exec.Command("/usr/bin/ffmpeg", "-nostdin", "-loglevel", "error", "-y",
				"-f", "lavfi", "-i", "testsrc=size=64x64:rate=1", "-t", "1",
				"-c:v", "mpeg4", "-f", format, path)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("create source media: %v: %s", err, output)
			}
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			var output bytes.Buffer
			if err := convertFile(context.Background(), file, 0, &output); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), "ftyp") ||
				!strings.Contains(output.String(), "moof") || !strings.Contains(output.String(), "mdat") {
				t.Fatalf("transcode output has %d bytes without MP4 fragments", output.Len())
			}
		})
	}
}

// Requirements: stream-service-svc/FR-005.
func TestFFmpegConvertsStillJPEGToFragmentedMP4(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("host ffmpeg is not installed")
	}
	path := filepath.Join(t.TempDir(), "still.jpg")
	command := exec.Command("/usr/bin/ffmpeg", "-nostdin", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "color=c=blue:s=64x64", "-frames:v", "1", path)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("create JPEG still: %v: %s", err, output)
	}
	jpeg, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := convertJPEGToVideo(context.Background(), jpeg, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "ftyp") ||
		!strings.Contains(output.String(), "moof") || !strings.Contains(output.String(), "mdat") {
		t.Fatalf("still video has %d bytes without MP4 fragments", output.Len())
	}
}

// TestUserPreviewConvertsApprovedHEIC covers the inode-bound decoder helper.
func TestUserPreviewConvertsApprovedHEIC(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("the preview helper intentionally refuses root")
	}
	if _, err := exec.LookPath("heif-dec"); err != nil {
		t.Skip("host heif-dec is not installed")
	}
	path := mustStreamHEIC(t)
	device, inode := mustMediaOperands(t, path)
	data := mustCaptureMediaStdout(t, func() error {
		return runUserPreview([]string{path, device, inode})
	})
	if !bytes.HasPrefix(data, []byte{0xff, 0xd8}) || !bytes.HasSuffix(data, []byte{0xff, 0xd9}) {
		t.Fatalf("preview did not return a complete JPEG (%d bytes)", len(data))
	}
}

// TestUserTranscoderConvertsApprovedHEIC covers the still-video helper path.
func TestUserTranscoderConvertsApprovedHEIC(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("the transcoder helper intentionally refuses root")
	}
	if _, err := exec.LookPath("heif-dec"); err != nil {
		t.Skip("host heif-dec is not installed")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("host ffmpeg is not installed")
	}
	path := mustStreamHEIC(t)
	device, inode := mustMediaOperands(t, path)
	data := mustCaptureMediaStdout(t, func() error {
		return runUserTranscoder([]string{path, device, inode, "0"})
	})
	if !strings.Contains(string(data), "ftyp") || !strings.Contains(string(data), "moof") {
		t.Fatalf("still transcode did not return fragmented MP4 (%d bytes)", len(data))
	}
}

// TestConversionStartersRejectUntrustedInputs checks the subprocess boundary.
func TestConversionStartersRejectUntrustedInputs(t *testing.T) {
	path := mustStreamHEIC(t)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		start func(string) error
	}{
		{"preview path", func(path string) error {
			_, err := startUserPreview(t.Context(), fileaccess.Subject{}, path, info)
			return err
		}},
		{"transcode path", func(path string) error {
			_, err := startUserTranscoder(t.Context(), fileaccess.Subject{}, path, info, 0)
			return err
		}},
		{"reader path", func(path string) error {
			_, err := userReader(t.Context(), fileaccess.Subject{}, path, info, 0, 1)
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.start("-option"); err == nil {
				t.Fatal("accepted option-shaped path")
			}
		})
		t.Run(test.name+" identity", func(t *testing.T) {
			if err := test.start(path); err == nil {
				t.Fatal("accepted an unverified Unix identity")
			}
		})
	}
}

// TestConversionStartersLaunchForVerifiedAccount checks the fixed subprocess boundary.
func TestConversionStartersLaunchForVerifiedAccount(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Unix helper identities must be non-root")
	}
	account, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	path := mustStreamHEIC(t)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	subject := fileaccess.Subject{Username: account.Username, UID: account.Uid}
	for _, test := range []struct {
		name  string
		start func() (io.ReadCloser, error)
	}{
		{"preview", func() (io.ReadCloser, error) {
			return startUserPreview(t.Context(), subject, path, info)
		}},
		{"transcode", func() (io.ReadCloser, error) {
			return startUserTranscoder(t.Context(), subject, path, info, 0)
		}},
		{"reader", func() (io.ReadCloser, error) {
			return userReader(t.Context(), subject, path, info, 0, 1)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader, err := test.start()
			if err != nil {
				t.Fatal(err)
			}
			if err := reader.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
