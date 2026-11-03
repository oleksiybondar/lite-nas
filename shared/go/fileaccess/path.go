package fileaccess

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ValidateHostPath rejects ambiguous path spellings before filesystem lookup.
// HTTP adapters must call this on the decoded path suffix before file access.
func ValidateHostPath(path string) error {
	if len(path) == 0 || len(path) > 4096 || !filepath.IsAbs(path) ||
		filepath.Clean(path) != path || strings.IndexByte(path, 0) >= 0 {
		return ErrInvalidPath
	}
	return nil
}

// ResolveExisting returns the canonical absolute path of an existing object.
// The caller must authorize the returned path and check the opened object again.
func ResolveExisting(path string) (string, error) {
	if err := ValidateHostPath(path); err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	if err := ValidateHostPath(resolved); err != nil {
		return "", err
	}
	return resolved, nil
}

// ResolveNewFile returns a canonical target under an existing real directory.
// It rejects a symbolic-link target, matching the writer's publication policy.
func ResolveNewFile(path string) (string, error) {
	if err := ValidateHostPath(path); err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	if err := ValidateHostPath(parent); err != nil {
		return "", err
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err == nil && !info.Mode().IsRegular() {
		return "", ErrInvalidPath
	}
	return filepath.Join(parent, filepath.Base(path)), nil
}

// WithinRoots reports whether path is a root or descendant of any canonical root.
// Component-aware comparison prevents /home-old from matching /home.
func WithinRoots(path string, roots []string) bool {
	if ValidateHostPath(path) != nil {
		return false
	}
	for _, root := range roots {
		if ValidateHostPath(root) != nil {
			continue
		}
		relative, err := filepath.Rel(root, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}
