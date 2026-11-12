package fileaccess

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestWithinRootsUsesComponentBoundaries covers file-service-svc/FR-004 and FR-009.
func TestWithinRootsUsesComponentBoundaries(t *testing.T) {
	for _, test := range []struct {
		path string
		want bool
	}{
		{"/home", true},
		{"/home/alice/file", true},
		{"/home-old/alice/file", false},
		{"/home/../etc/passwd", false},
		{"relative/path", false},
	} {
		if got := WithinRoots(test.path, []string{"/home"}); got != test.want {
			t.Errorf("WithinRoots(%q) = %t, want %t", test.path, got, test.want)
		}
	}
}

// TestResolveExistingUsesCanonicalTarget covers file-service-svc/FR-003.
func TestResolveExistingUsesCanonicalTarget(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveExisting(link)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != target {
		t.Fatalf("resolved path = %q, want %q", resolved, target)
	}
	if _, err := ResolveExisting(root + "/../" + filepath.Base(root) + "/target"); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("unclean path error = %v", err)
	}
}

// TestResolveNewFileRejectsUnsupportedTarget covers file-service-svc/FR-004.
func TestResolveNewFileRejectsUnsupportedTarget(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink("missing", link); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveNewFile(link); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("symlink target error = %v", err)
	}
	newPath := filepath.Join(root, "new")
	resolved, err := ResolveNewFile(newPath)
	if err != nil || resolved != newPath {
		t.Fatalf("new target = %q, error = %v", resolved, err)
	}
}
