package main

import (
	"fmt"
	"time"

	"gopkg.in/ini.v1"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// mountConfig uses the shared INI sections and one validated poll interval.
type mountConfig struct {
	Messaging    sharedconfig.MessagingConfig
	Auth         sharedconfig.AuthConfig
	Logging      sharedconfig.LoggingConfig
	PollInterval time.Duration
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
		return mountConfig{Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging, PollInterval: interval}, nil
	})
}
