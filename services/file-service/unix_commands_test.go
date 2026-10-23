package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"lite-nas/shared/fileaccess"
)

// TestValidateWriterArgumentsRejectsUnexpectedOperationsAndPaths covers writer arguments.
func TestValidateWriterArgumentsRejectsUnexpectedOperationsAndPaths(t *testing.T) {
	tests := []struct {
		operation string
		record    *uploadRecord
		data      []byte
	}{
		{operation: "--help", record: &uploadRecord{Target: "/tmp/file"}},
		{operation: "create", record: nil},
		{operation: "create", record: &uploadRecord{Target: "relative/file"}},
		{operation: "write", record: &uploadRecord{Target: "/tmp/file", Offset: -1}},
		{operation: "write", record: &uploadRecord{Target: "/tmp/file"}, data: make([]byte, 16*1024*1024+1)},
	}
	for _, test := range tests {
		if err := validateWriterArguments(test.operation, test.record, test.data); err == nil {
			t.Errorf("validateWriterArguments(%q, %#v) accepted invalid input", test.operation, test.record)
		}
	}
}

// TestValidateUserCommandArgumentsRejectsUnexpectedOperands covers fixed user commands.
func TestValidateUserCommandArgumentsRejectsUnexpectedOperands(t *testing.T) {
	tests := []struct {
		binary string
		paths  []string
	}{
		{binary: "/bin/sh", paths: []string{"/tmp/file"}},
		{binary: "/bin/rm", paths: []string{"-rf"}},
		{binary: "/bin/rm", paths: []string{"/tmp/one", "/tmp/two"}},
		{binary: "/bin/mkdir", paths: []string{"-m", "/tmp/folder"}},
		{binary: "/bin/mkdir", paths: []string{"0777", "/tmp/folder"}},
		{binary: "/bin/mkdir", paths: []string{"0775", "--help"}},
	}
	for _, test := range tests {
		if err := validateUserCommandArguments(test.binary, test.paths); err == nil {
			t.Errorf("validateUserCommandArguments(%q, %q) accepted invalid input", test.binary, test.paths)
		}
	}
}

// TestInvokeWriterRejectsUnsupportedOperationBeforeSudo covers operation validation.
func TestInvokeWriterRejectsUnsupportedOperationBeforeSudo(t *testing.T) {
	if _, err := invokeWriter(t.Context(), fileaccess.Subject{}, "--help", &uploadRecord{Target: "/tmp/file"}, nil); err == nil {
		t.Fatal("invokeWriter() accepted unsupported operation")
	}
}

// TestInvokeWriterRejectsInvalidIdentityBeforeSudo covers user validation.
func TestInvokeWriterRejectsInvalidIdentityBeforeSudo(t *testing.T) {
	if _, err := invokeWriter(t.Context(), fileaccess.Subject{}, "create", &uploadRecord{Target: "/tmp/file"}, nil); err == nil {
		t.Fatal("invokeWriter() accepted unverified Unix identity")
	}
}

// TestOpenCatRejectsInvalidPathBeforeSudo covers read path operands.
func TestOpenCatRejectsInvalidPathBeforeSudo(t *testing.T) {
	if _, err := openCat(t.Context(), fileaccess.Subject{}, "relative"); err == nil {
		t.Fatal("openCat() accepted relative path")
	}
}

// TestOpenCatRejectsInvalidIdentityBeforeSudo covers read user identity.
func TestOpenCatRejectsInvalidIdentityBeforeSudo(t *testing.T) {
	if _, err := openCat(t.Context(), fileaccess.Subject{}, "/tmp/file"); err == nil {
		t.Fatal("openCat() accepted unverified Unix identity")
	}
}

// TestCatProcessReapsCompletedReader covers a successful streamed subprocess.
func TestCatProcessReapsCompletedReader(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	process := mustStartCatProcess(t, "/bin/cat", path)
	t.Cleanup(func() { _ = process.Close() })
	data, err := io.ReadAll(process)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("catProcess.Read() = %q", data)
	}
}

// TestCatProcessCloseStopsActiveReader covers request cancellation cleanup.
func TestCatProcessCloseStopsActiveReader(t *testing.T) {
	process := mustStartCatProcess(t, "/bin/sleep", "10")
	if err := process.Close(); err != nil {
		t.Fatalf("catProcess.Close() error = %v", err)
	}
}

// mustStartCatProcess starts a controlled subprocess for reader lifecycle tests.
func mustStartCatProcess(t *testing.T, name string, args ...string) *catProcess {
	t.Helper()
	command := exec.CommandContext(t.Context(), name, args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	return &catProcess{ReadCloser: stdout, command: command}
}
