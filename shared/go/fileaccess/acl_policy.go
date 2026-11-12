package fileaccess

import (
	"path/filepath"
	"strings"
)

// ReadACLPolicy configures application folder authorization and global path enforcement.
type ReadACLPolicy struct {
	ApplicationFolders map[string][]string
	EnforcePaths       []string
}

// ReadDecisionPath selects the RBAC path for an app-scoped read request.
func (policy ReadACLPolicy) ReadDecisionPath(application string, path string) string {
	if policy.MustEnforce(path) {
		return path
	}
	var selected string
	for _, folder := range policy.ApplicationFolders[application] {
		if isWithinPath(folder, path) && len(folder) > len(selected) {
			selected = folder
		}
	}
	if selected != "" {
		return selected
	}
	return path
}

// ParentFolderCoversEntry reports whether a successful parent check covers a listed child.
func (policy ReadACLPolicy) ParentFolderCoversEntry(application string, directory string, entry string) bool {
	covered := false
	for _, folder := range policy.ApplicationFolders[application] {
		if isWithinPath(folder, directory) {
			covered = true
			break
		}
	}
	return covered && !policy.MustEnforce(directory) && !policy.MustEnforce(entry)
}

// MustEnforce reports whether a global enforcement path contains the requested path.
func (policy ReadACLPolicy) MustEnforce(path string) bool {
	for _, enforced := range policy.EnforcePaths {
		if isWithinPath(enforced, path) {
			return true
		}
	}
	return false
}

// isWithinPath matches a path itself or a descendant without prefix ambiguities.
func isWithinPath(root string, path string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && !filepath.IsAbs(relative) &&
		(relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))))
}
