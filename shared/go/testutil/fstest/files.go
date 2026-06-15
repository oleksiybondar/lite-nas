package fstest

import (
	"os"
	"path/filepath"
	"testing"
)

// MustMkdirAll creates a directory tree for a test fixture or fails the test.
func MustMkdirAll(t *testing.T, path string, perm os.FileMode) {
	t.Helper()

	if err := os.MkdirAll(path, perm); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

// MustWriteFile writes a fixture file and creates parent directories when
// needed.
func MustWriteFile(t *testing.T, path string, data string, dirPerm os.FileMode, filePerm os.FileMode) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(data), filePerm); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

// MustSymlink creates a fixture symlink or fails the test.
func MustSymlink(t *testing.T, target string, path string) {
	t.Helper()

	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("Symlink(%q, %q) error = %v", target, path, err)
	}
}
