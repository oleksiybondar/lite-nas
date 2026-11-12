package config

import (
	"errors"
	"path/filepath"
	"strings"

	"lite-nas/shared/fileaccess"

	"gopkg.in/ini.v1"
)

// LoadACLConfig parses the shared app folder and global enforcement rules.
func LoadACLConfig(cfgFile *ini.File) (fileaccess.ReadACLPolicy, error) {
	section := cfgFile.Section("acl")
	applications := map[string]string{
		"moments_folders":     "moments",
		"admin_panel_folders": "admin-panel",
	}
	policy := fileaccess.ReadACLPolicy{ApplicationFolders: make(map[string][]string)}
	for key, application := range applications {
		paths, err := parseACLPaths(section.Key(key).String())
		if err != nil {
			return fileaccess.ReadACLPolicy{}, errors.New("acl application folders must contain clean absolute paths")
		}
		if len(paths) > 0 {
			policy.ApplicationFolders[application] = paths
		}
	}
	paths, err := parseACLPaths(section.Key("enforce_paths").String())
	if err != nil {
		return fileaccess.ReadACLPolicy{}, errors.New("acl.enforce_paths must contain clean absolute paths")
	}
	policy.EnforcePaths = paths
	if len(policy.ApplicationFolders) == 0 {
		policy.ApplicationFolders = nil
	}
	return policy, nil
}

// parseACLPaths parses comma-separated clean absolute host paths.
func parseACLPaths(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	paths := make([]string, 0)
	for _, part := range strings.Split(value, ",") {
		path := strings.TrimSpace(part)
		if fileaccess.ValidateHostPath(path) != nil || filepath.Clean(path) != path {
			return nil, errors.New("path is not a clean absolute host path")
		}
		paths = append(paths, path)
	}
	return paths, nil
}
