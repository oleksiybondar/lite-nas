// Package filesystem enforces File Service mutation boundaries before RBAC checks.
package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"lite-nas/shared/fileaccess"
	"lite-nas/shared/roleauth"
)

var (
	// ErrOutsideWriteRoots means a mutation target is outside configured storage.
	ErrOutsideWriteRoots = errors.New("target is outside writable storage roots")
	// ErrProtectedPath means a mutation target is inside a protected OS or operator path.
	ErrProtectedPath = errors.New("target is protected")
	// ErrDeniedPath means configuration hides a path from this caller.
	ErrDeniedPath = errors.New("target is denied")
)

// Policy holds canonical roots used to check each mutation target.
type Policy struct {
	writeRoots           []string
	readRoots            []string
	protectedRoots       []string
	readDeniedNonAdmin   []string
	readDeniedAll        []string
	knownPathDirectories []string
}

// ReadRules names configured caller exclusions and known-path directory masks.
type ReadRules struct {
	ReadRoots            []string
	NonAdminDenied       []string
	AllDenied            []string
	KnownPathDirectories []string
}

// NewPolicy validates configured path roots and caller-specific rules.
// Write and read roots must exist at startup; protected paths may be absent.
func NewPolicy(writeRoots []string, protectedPaths []string, rules ...ReadRules) (Policy, error) {
	if len(writeRoots) == 0 {
		return Policy{}, errors.New("at least one write root is required")
	}
	canonicalWrite, err := canonicalDirectories(writeRoots)
	if err != nil {
		return Policy{}, fmt.Errorf("write roots: %w", err)
	}
	canonicalProtected, err := canonicalProtectedRoots(protectedPaths)
	if err != nil {
		return Policy{}, fmt.Errorf("protected paths: %w", err)
	}
	policy := Policy{writeRoots: canonicalWrite, protectedRoots: canonicalProtected}
	if len(rules) > 1 {
		return Policy{}, errors.New("only one read rule set is allowed")
	}
	if len(rules) == 1 {
		canonicalRead, err := canonicalDirectories(rules[0].ReadRoots)
		if err != nil {
			return Policy{}, fmt.Errorf("read roots: %w", err)
		}
		policy.readRoots = canonicalRead
		for _, pattern := range append(append(append([]string{}, rules[0].NonAdminDenied...), rules[0].AllDenied...), rules[0].KnownPathDirectories...) {
			if err := ValidatePathMask(pattern); err != nil {
				return Policy{}, err
			}
		}
		policy.readDeniedNonAdmin = rules[0].NonAdminDenied
		policy.readDeniedAll = rules[0].AllDenied
		policy.knownPathDirectories = rules[0].KnownPathDirectories
	}
	return policy, nil
}

// ValidatePathMask accepts clean absolute paths with whole-component stars only.
func ValidatePathMask(pattern string) error {
	if pattern == "/" || fileaccess.ValidateHostPath(strings.ReplaceAll(pattern, "*", "mask")) != nil {
		return errors.New("path mask must be clean and absolute")
	}
	for _, segment := range strings.Split(strings.TrimPrefix(pattern, "/"), "/") {
		if strings.ContainsAny(segment, "?[]") || strings.Contains(segment, "*") && segment != "*" {
			return errors.New("path mask permits only whole-component stars")
		}
	}
	return nil
}

// matchesMask reports an exact directory or file match without crossing path components.
func matchesMask(pattern string, target string) bool {
	matched, err := path.Match(pattern, target)
	return err == nil && matched
}

// deniedByMask reports whether any mask covers a path or one of its ancestors.
func deniedByMask(patterns []string, target string) bool {
	for current := target; current != "/"; current = filepath.Dir(current) {
		for _, pattern := range patterns {
			if matchesMask(pattern, current) {
				return true
			}
		}
	}
	return false
}

// DirectoryMode selects the created directory's preset; explicit privacy wins.
func (p Policy) DirectoryMode(target string, private bool) string {
	if private {
		return "0770"
	}
	for _, pattern := range p.knownPathDirectories {
		if matchesMask(pattern, target) {
			return "0771"
		}
	}
	return "0775"
}

// AllowSubject denies configured path trees before any read or mutation operation.
func (p Policy) AllowSubject(target string, groups []string) error {
	if err := fileaccess.ValidateHostPath(target); err != nil {
		return err
	}
	if deniedByMask(p.readDeniedAll, target) ||
		!roleauth.MatchesRequirement(groups, roleauth.RequirementAdministrator) && deniedByMask(p.readDeniedNonAdmin, target) {
		return ErrDeniedPath
	}
	return nil
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

// canonicalProtectedRoots resolves existing paths and retains absent future roots.
func canonicalProtectedRoots(paths []string) ([]string, error) {
	roots := make([]string, 0, len(paths))
	for _, path := range paths {
		if err := fileaccess.ValidateHostPath(path); err != nil {
			return nil, err
		}
		resolved, err := filepath.EvalSymlinks(path)
		if errors.Is(err, os.ErrNotExist) {
			parent, parentErr := filepath.EvalSymlinks(filepath.Dir(path))
			if parentErr != nil {
				return nil, parentErr
			}
			resolved = filepath.Join(parent, filepath.Base(path))
			err = nil
		}
		if err != nil {
			return nil, err
		}
		roots = append(roots, resolved)
	}
	return roots, nil
}
