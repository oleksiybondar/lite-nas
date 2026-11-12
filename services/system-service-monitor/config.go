package main

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// monitorConfig lists the systemd units this unprivileged monitor observes.
type monitorConfig struct {
	Messaging    sharedconfig.MessagingConfig
	Auth         sharedconfig.AuthConfig
	Logging      sharedconfig.LoggingConfig
	PollInterval time.Duration
	Units        []string
}

// loadMonitorConfig validates every configured unit before any systemctl call.
func loadMonitorConfig(reader fileio.Reader) (monitorConfig, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (monitorConfig, error) {
		section := file.Section("monitor")
		interval, err := time.ParseDuration(section.Key("poll_interval").MustString("10s"))
		if err != nil || interval < time.Second || interval > time.Hour {
			return monitorConfig{}, fmt.Errorf("monitor poll interval must be between 1s and 1h")
		}
		if shared.Auth.ServiceLogin == "" {
			return monitorConfig{}, fmt.Errorf("monitor auth service_login is required")
		}
		values := strings.Split(section.Key("units").String(), ",")
		if len(values) == 0 || len(values) > 128 {
			return monitorConfig{}, fmt.Errorf("monitor requires 1-128 units")
		}
		units := make([]string, 0, len(values))
		seen := make(map[string]bool)
		for _, value := range values {
			unit := strings.TrimSpace(value)
			if !validUnitName(unit) || seen[unit] {
				return monitorConfig{}, fmt.Errorf("invalid or duplicate service unit %q", unit)
			}
			seen[unit] = true
			units = append(units, unit)
		}
		return monitorConfig{Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging, PollInterval: interval, Units: units}, nil
	})
}
