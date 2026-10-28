package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
	"lite-nas/shared/mountinfo"
)

// antivirusConfig requires explicit mount and filesystem allowlists.
type antivirusConfig struct {
	Messaging             sharedconfig.MessagingConfig
	Auth                  sharedconfig.AuthConfig
	Logging               sharedconfig.LoggingConfig
	FSTypes               map[string]bool
	MountPrefixes         []string
	ExcludedMountPrefixes []string
	ScanTimeout           time.Duration
}

// loadAntivirusConfig rejects missing or unsafe scanner selection settings.
func loadAntivirusConfig(reader fileio.Reader) (antivirusConfig, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (antivirusConfig, error) {
		section := file.Section("scan")
		timeout, err := time.ParseDuration(section.Key("timeout").MustString("30m"))
		if err != nil || timeout < time.Minute || timeout > 24*time.Hour {
			return antivirusConfig{}, fmt.Errorf("scan timeout must be between 1m and 24h")
		}
		if shared.Auth.ServiceLogin == "" {
			return antivirusConfig{}, fmt.Errorf("antivirus auth service_login is required")
		}
		types := make(map[string]bool)
		for _, value := range strings.Split(section.Key("fs_types").String(), ",") {
			fs := strings.TrimSpace(value)
			if fs == "" || len(fs) > 128 || strings.ContainsAny(fs, "/\\ \t\n\r") || types[fs] {
				return antivirusConfig{}, fmt.Errorf("invalid or duplicate fs_type %q", fs)
			}
			types[fs] = true
		}
		prefixes := make([]string, 0)
		for _, value := range strings.Split(section.Key("mount_prefixes").String(), ",") {
			prefix := strings.TrimSpace(value)
			if !filepath.IsAbs(prefix) || filepath.Clean(prefix) != prefix || prefix == "/" {
				return antivirusConfig{}, fmt.Errorf("invalid mount prefix %q", prefix)
			}
			prefixes = append(prefixes, prefix)
		}
		excludedPrefixes := make([]string, 0)
		for _, value := range strings.Split(section.Key("excluded_mount_prefixes").String(), ",") {
			prefix := strings.TrimSpace(value)
			if prefix == "" {
				continue
			}
			if !filepath.IsAbs(prefix) || filepath.Clean(prefix) != prefix || prefix == "/" {
				return antivirusConfig{}, fmt.Errorf("invalid excluded mount prefix %q", prefix)
			}
			excludedPrefixes = append(excludedPrefixes, prefix)
		}
		if len(types) > 32 || len(prefixes) > 16 || len(excludedPrefixes) > 16 {
			return antivirusConfig{}, fmt.Errorf("antivirus allowlist too large")
		}
		return antivirusConfig{
			Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging,
			FSTypes: types, MountPrefixes: prefixes, ExcludedMountPrefixes: excludedPrefixes, ScanTimeout: timeout,
		}, nil
	})
}

// eligibleMount checks only validated configuration and mount metadata; live
// identity and symlink checks happen again immediately before scanning.
func eligibleMount(cfg antivirusConfig, mount mountinfo.Record) bool {
	if !cfg.FSTypes[mount.FSType] || !filepath.IsAbs(mount.Point) || filepath.Clean(mount.Point) != mount.Point {
		return false
	}
	if excludedMountPrefix(cfg, mount.Point) != "" {
		return false
	}
	for _, prefix := range cfg.MountPrefixes {
		relative, err := filepath.Rel(prefix, mount.Point)
		if err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// excludedMountPrefix returns the configured exclusion containing the mount point.
func excludedMountPrefix(cfg antivirusConfig, mountPoint string) string {
	for _, prefix := range cfg.ExcludedMountPrefixes {
		relative, err := filepath.Rel(prefix, mountPoint)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return prefix
		}
	}
	return ""
}
