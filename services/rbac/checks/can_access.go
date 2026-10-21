package checks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	idparser "lite-nas/shared/parsers/id/identity"
)

// CheckPathOperation evaluates a directory-entry mutation or traversal without modifying files.
func CheckPathOperation(ctx context.Context, runner Runner, uid string, path string, operation string) (bool, error) {
	identity, numericUID, err := pathIdentity(ctx, runner, uid, path)
	if err != nil {
		return false, err
	}
	if operation == "traverse" {
		return checkTraversal(ctx, runner, identity, path)
	}
	if operation != "create" && operation != "replace" && operation != "delete" {
		return false, errors.New("unsupported path operation")
	}
	if path == string(filepath.Separator) {
		return false, nil
	}
	allowed, err := checkTraversal(ctx, runner, identity, filepath.Dir(path))
	if err != nil || !allowed {
		return false, err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return false, err
	}
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return false, err
	}
	parentACL, err := readACLForResolvedPath(ctx, runner, parent)
	if err != nil {
		return false, err
	}
	if !evaluateAccess(identity, parentACL, "write") {
		return false, nil
	}
	targetInfo, err := os.Lstat(path)
	if operation == "create" {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}
	if err != nil {
		return false, err
	}
	if !targetInfo.Mode().IsRegular() {
		return false, nil
	}
	if parentInfo.Mode()&os.ModeSticky != 0 {
		parentOwner, ownerErr := fileOwner(parentInfo)
		if ownerErr != nil {
			return false, ownerErr
		}
		targetOwner, ownerErr := fileOwner(targetInfo)
		if ownerErr != nil {
			return false, ownerErr
		}
		if numericUID != parentOwner && numericUID != targetOwner {
			return false, nil
		}
	}
	return true, nil
}

// canAccess checks a resolved existing inode after verifying caller traversal.
func canAccess(ctx context.Context, runner Runner, uid string, path string, operation string) (bool, error) {
	identity, _, err := pathIdentity(ctx, runner, uid, path)
	if err != nil {
		return false, err
	}
	allowed, err := checkTraversal(ctx, runner, identity, filepath.Dir(path))
	if err != nil || !allowed {
		return false, err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false, err
	}
	allowed, err = checkResolvedAncestors(ctx, runner, identity, resolved, make(map[string]bool))
	if err != nil || !allowed {
		return false, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() && !info.IsDir() {
		return false, nil
	}
	if operation == "write" && !info.Mode().IsRegular() {
		return false, nil
	}
	acl, err := readACLForResolvedPath(ctx, runner, resolved)
	if err != nil {
		return false, err
	}
	if !evaluateAccess(identity, acl, operation) {
		return false, nil
	}
	// A directory listing needs search on the listed directory as well as read.
	return !info.IsDir() || operation != "read" || evaluateAccess(identity, acl, "execute"), nil
}

// pathIdentity validates a non-root UID and an unambiguous host path before host lookups.
func pathIdentity(ctx context.Context, runner Runner, uid string, path string) (idparser.Identity, uint32, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || strings.IndexByte(path, 0) >= 0 {
		return idparser.Identity{}, 0, errors.New("path must be clean and absolute")
	}
	numericUID, err := strconv.ParseUint(uid, 10, 32)
	if err != nil || numericUID == 0 {
		return idparser.Identity{}, 0, errors.New("UID must be a nonzero decimal number")
	}
	identity, err := ResolveIdentityByUID(ctx, runner, uid)
	if err != nil {
		return idparser.Identity{}, 0, err
	}
	if identity.UID != uid {
		return idparser.Identity{}, 0, errors.New("resolved UID does not match request")
	}
	return identity, uint32(numericUID), nil
}

// checkTraversal checks lexical directories and the ancestors of resolved symlink targets.
func checkTraversal(ctx context.Context, runner Runner, identity idparser.Identity, path string) (bool, error) {
	checked := make(map[string]bool)
	current := string(filepath.Separator)
	parts := strings.Split(strings.TrimPrefix(path, current), current)
	for _, part := range append([]string{""}, parts...) {
		if part != "" {
			current = filepath.Join(current, part)
		}
		resolved, err := filepath.EvalSymlinks(current)
		if err != nil {
			return false, err
		}
		allowed, err := checkResolvedDirectory(ctx, runner, identity, resolved, checked)
		if err != nil || !allowed {
			return false, err
		}
	}
	return true, nil
}

// checkResolvedAncestors verifies the directory chain to a resolved path.
func checkResolvedAncestors(ctx context.Context, runner Runner, identity idparser.Identity, path string, checked map[string]bool) (bool, error) {
	for directory := filepath.Dir(path); ; directory = filepath.Dir(directory) {
		allowed, err := checkResolvedDirectory(ctx, runner, identity, directory, checked)
		if err != nil || !allowed {
			return false, err
		}
		if directory == string(filepath.Separator) {
			break
		}
	}
	return true, nil
}

// checkResolvedDirectory verifies search on one directory and its ancestors.
func checkResolvedDirectory(ctx context.Context, runner Runner, identity idparser.Identity, path string, checked map[string]bool) (bool, error) {
	if checked[path] {
		return true, nil
	}
	if path != string(filepath.Separator) {
		allowed, err := checkResolvedAncestors(ctx, runner, identity, path, checked)
		if err != nil || !allowed {
			return false, err
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}
	acl, err := readACLForResolvedPath(ctx, runner, path)
	if err != nil {
		return false, err
	}
	if !evaluateAccess(identity, acl, "execute") {
		return false, nil
	}
	checked[path] = true
	return true, nil
}

// fileOwner reads the owner needed for a sticky-directory mutation decision.
func fileOwner(info os.FileInfo) (uint32, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, errors.New("file owner unavailable")
	}
	return stat.Uid, nil
}
