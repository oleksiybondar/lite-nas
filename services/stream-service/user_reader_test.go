package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

// mustMediaOperands returns the inode pair the parent process would approve.
func mustMediaOperands(t *testing.T, path string) (string, string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat := info.Sys().(*syscall.Stat_t)
	return strconv.FormatUint(uint64(stat.Dev), 10), strconv.FormatUint(stat.Ino, 10)
}

// mustCaptureMediaStdout runs a helper and returns its bounded output bytes.
func mustCaptureMediaStdout(t *testing.T, run func() error) []byte {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "helper-output-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	previous := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = previous }()
	if err := run(); err != nil {
		t.Fatal(err)
	}
	if _, err := output.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Requirements: stream-service-svc/SR-001 and TR-001.
func TestReaderRejectsMalformedOperands(t *testing.T) {
	for _, args := range [][]string{
		{"-option", "1", "1", "0", "0"},
		{"/tmp/file", "not-a-device", "1", "0", "0"},
		{"/tmp/file", "1", "1", "-1", "0"},
		{"/tmp/file", "1", "1", "0", "-1"},
		{"/tmp/file", "1", "1", "0"},
	} {
		if _, err := parseReaderInput(args); err == nil {
			t.Fatalf("accepted malformed operands: %v", args)
		}
	}
}

// Requirements: stream-service-svc/SR-001.
func TestReaderRefusesReplacedInode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "movie.mov")
	if err := os.WriteFile(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat := before.Sys().(*syscall.Stat_t)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := openExpectedFile(path, uint64(stat.Dev), stat.Ino)
	if file != nil {
		file.Close()
	}
	if err == nil {
		t.Fatal("reader accepted a replacement inode")
	}
}

// Requirements: stream-service-svc/FR-004.
func TestReaderInputAcceptsValidRange(t *testing.T) {
	input, err := parseReaderInput([]string{"/home/operator/video.avi", "2", "3", "100", "50"})
	if err != nil || input.Offset != 100 || input.Count != 50 {
		t.Fatalf("parsed input=%+v err=%v", input, err)
	}
}

// Requirements: stream-service-svc/SR-001 and FR-005.
func TestTranscoderRejectsOptionShapedPath(t *testing.T) {
	if err := runUserTranscoder([]string{"-option", "1", "2", "0"}); err == nil {
		t.Fatal("transcoder accepted an option-shaped path")
	}
}

// TestUserReaderStreamsApprovedRange covers the verified owner reader helper.
func TestUserReaderStreamsApprovedRange(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("the reader intentionally refuses root")
	}
	path := filepath.Join(t.TempDir(), "movie.mov")
	if err := os.WriteFile(path, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	device, inode := mustMediaOperands(t, path)
	data := mustCaptureMediaStdout(t, func() error {
		return runUserReader([]string{path, device, inode, "2", "3"})
	})
	if string(data) != "cde" {
		t.Fatalf("reader bytes = %q, want cde", data)
	}
}

// TestUserReaderRejectsChangedInode covers the helper's second identity check.
func TestUserReaderRejectsChangedInode(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("the reader intentionally refuses root")
	}
	path := filepath.Join(t.TempDir(), "movie.mov")
	if err := os.WriteFile(path, []byte("abcdef"), 0o600); err != nil {
		t.Fatal(err)
	}
	device, _ := mustMediaOperands(t, path)
	if err := runUserReader([]string{path, device, "1", "0", "0"}); err == nil {
		t.Fatal("reader accepted a mismatched inode")
	}
}

// TestReaderProcessReportsFailedChild covers a child that exits without media bytes.
func TestReaderProcessReportsFailedChild(t *testing.T) {
	command := exec.CommandContext(t.Context(), "/bin/false")
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	process := &readerProcess{ReadCloser: stdout, command: command}
	defer process.Close()
	var first [1]byte
	if _, err := process.Read(first[:]); err == nil || err == io.EOF {
		t.Fatalf("failed child read error = %v", err)
	}
}
