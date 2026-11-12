package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

const defaultExcludedMountPrefixes = "/run/user,/run/netns/litenas-scan-test,/run/docker/netns,/var/lib/docker/zfs/graph"

// mountConfig uses the shared INI sections and validated mount settings.
type mountConfig struct {
	Messaging             sharedconfig.MessagingConfig
	Auth                  sharedconfig.AuthConfig
	Logging               sharedconfig.LoggingConfig
	PollInterval          time.Duration
	ExcludedMountPrefixes []string
}

// loadMountConfig rejects intervals that would flood or stall observations.
func loadMountConfig(reader fileio.Reader) (mountConfig, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (mountConfig, error) {
		interval, err := time.ParseDuration(file.Section("mounts").Key("poll_interval").MustString("5s"))
		if err != nil || interval < time.Second || interval > time.Hour {
			return mountConfig{}, fmt.Errorf("mount poll interval must be between 1s and 1h")
		}
		if shared.Auth.ServiceLogin == "" {
			return mountConfig{}, fmt.Errorf("mount monitor auth service_login is required")
		}
		excludedValue := defaultExcludedMountPrefixes
		if file.Section("mounts").HasKey("excluded_mount_prefixes") {
			excludedValue = file.Section("mounts").Key("excluded_mount_prefixes").String()
		}
		prefixes := make([]string, 0)
		for _, value := range strings.Split(excludedValue, ",") {
			prefix := strings.TrimSpace(value)
			if prefix == "" {
				continue
			}
			if !filepath.IsAbs(prefix) || filepath.Clean(prefix) != prefix || prefix == "/" {
				return mountConfig{}, fmt.Errorf("invalid excluded mount prefix %q", prefix)
			}
			prefixes = append(prefixes, prefix)
		}
		if len(prefixes) > 16 {
			return mountConfig{}, fmt.Errorf("excluded mount prefix list too large")
		}
		return mountConfig{Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging, PollInterval: interval, ExcludedMountPrefixes: prefixes}, nil
	})
}
