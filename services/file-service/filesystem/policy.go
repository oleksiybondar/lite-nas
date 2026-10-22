// Package filesystem enforces File Service mutation boundaries before RBAC checks.
package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"lite-nas/shared/fileaccess"
)

var (
	// ErrOutsideWriteRoots means a mutation target is outside configured storage.
	ErrOutsideWriteRoots = errors.New("target is outside writable storage roots")
	// ErrProtectedPath means a mutation target is inside a protected OS or operator path.
	ErrProtectedPath = errors.New("target is protected")
)

// builtinProtectedPaths cannot be removed by service configuration.
var builtinProtectedPaths = []string{
	"/bin", "/boot", "/dev", "/etc", "/lib", "/lib64", "/proc",
	"/root", "/run", "/sbin", "/sys", "/usr", "/var/lib",
}

// Policy holds canonical roots used to check each mutation target.
type Policy struct {
	writeRoots     []string
	protectedRoots []string
}

// NewPolicy validates configured roots and adds non-removable protected paths.
// An empty writeRoots list selects /home. Configured roots must exist at startup.
func NewPolicy(writeRoots []string, protectedPaths []string) (Policy, error) {
	if len(writeRoots) == 0 {
		writeRoots = []string{"/home"}
	}
	canonicalWrite, err := canonicalDirectories(writeRoots)
	if err != nil {
		return Policy{}, fmt.Errorf("write roots: %w", err)
	}
	canonicalProtected, err := canonicalExistingRoots(append(append([]string(nil), builtinProtectedPaths...), protectedPaths...))
	if err != nil {
		return Policy{}, fmt.Errorf("protected paths: %w", err)
	}
	return Policy{writeRoots: canonicalWrite, protectedRoots: canonicalProtected}, nil
}

// AllowMutation checks a resolved target against write and protected roots.
// Callers must resolve the target again immediately before publication or delete.
func (p Policy) AllowMutation(target string) error {
	if err := fileaccess.ValidateHostPath(target); err != nil {
		return err
	}
	if !fileaccess.WithinRoots(target, p.writeRoots) {
		return ErrOutsideWriteRoots
	}
	if fileaccess.WithinRoots(target, p.protectedRoots) {
		return ErrProtectedPath
	}
	return nil
}

// canonicalDirectories requires each writable root to exist as a directory.
func canonicalDirectories(paths []string) ([]string, error) {
	roots := make([]string, 0, len(paths))
	for _, path := range paths {
		if err := fileaccess.ValidateHostPath(path); err != nil {
			return nil, err
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, errors.New("root is not a directory")
		}
		roots = append(roots, resolved)
	}
	return roots, nil
}

// canonicalExistingRoots skips absent built-in platform paths but rejects bad configured paths.
func canonicalExistingRoots(paths []string) ([]string, error) {
	roots := make([]string, 0, len(paths))
	for index, path := range paths {
		if err := fileaccess.ValidateHostPath(path); err != nil {
			return nil, err
		}
		resolved, err := filepath.EvalSymlinks(path)
		if errors.Is(err, os.ErrNotExist) && index < len(builtinProtectedPaths) {
			continue
		}
		if err != nil {
			return nil, err
		}
		roots = append(roots, resolved)
	}
	return roots, nil
}
