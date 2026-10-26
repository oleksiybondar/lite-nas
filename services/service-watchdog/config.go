package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
	"lite-nas/shared/systemd"
)

// protectedUnit is an independent Watchdog allowlist entry; recovery and
// enablement expectations are explicit rather than trusted from NATS.
type protectedUnit struct {
	Name               string
	ExpectedEnablement string
	Recover            bool
	Critical           bool
}

// watchdogConfig bounds reconciliation and recovery attempts.
type watchdogConfig struct {
	Messaging     sharedconfig.MessagingConfig
	Auth          sharedconfig.AuthConfig
	Logging       sharedconfig.LoggingConfig
	PollInterval  time.Duration
	RecoveryDelay time.Duration
	Backoff       time.Duration
	MaxAttempts   int
	Units         map[string]protectedUnit
}

// loadWatchdogConfig validates every unit and policy before the service can
// inspect or start any configured service.
func loadWatchdogConfig(reader fileio.Reader) (watchdogConfig, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (watchdogConfig, error) {
		section := file.Section("watchdog")
		poll, err := time.ParseDuration(section.Key("poll_interval").MustString("10s"))
		if err != nil || poll < time.Second || poll > time.Hour {
			return watchdogConfig{}, fmt.Errorf("invalid watchdog poll interval")
		}
		delay, err := time.ParseDuration(section.Key("recovery_delay").MustString("30s"))
		if err != nil || delay < time.Second || delay > 24*time.Hour {
			return watchdogConfig{}, fmt.Errorf("invalid watchdog recovery delay")
		}
		backoff, err := time.ParseDuration(section.Key("backoff").MustString("1m"))
		if err != nil || backoff < time.Second || backoff > 24*time.Hour {
			return watchdogConfig{}, fmt.Errorf("invalid watchdog backoff")
		}
		maxAttempts, err := strconv.Atoi(section.Key("max_attempts").MustString("3"))
		if err != nil || maxAttempts < 1 || maxAttempts > 10 {
			return watchdogConfig{}, fmt.Errorf("max_attempts must be 1-10")
		}
		if shared.Auth.ServiceLogin == "" {
			return watchdogConfig{}, fmt.Errorf("watchdog auth service_login is required")
		}
		units := make(map[string]protectedUnit)
		for _, value := range strings.Split(section.Key("units").String(), ",") {
			name := strings.TrimSpace(value)
			if !systemd.ValidUnitName(name) || units[name].Name != "" {
				return watchdogConfig{}, fmt.Errorf("invalid or duplicate protected unit %q", name)
			}
			unitSection, err := file.GetSection("unit." + name)
			if err != nil {
				return watchdogConfig{}, fmt.Errorf("missing policy section for %s", name)
			}
			enablement := strings.TrimSpace(unitSection.Key("expected_enablement").String())
			switch enablement {
			case "enabled", "disabled", "static", "indirect", "masked":
			default:
				return watchdogConfig{}, fmt.Errorf("invalid expected enablement for %s", name)
			}
			recoverEnabled, err := strconv.ParseBool(unitSection.Key("recover").String())
			if err != nil {
				return watchdogConfig{}, fmt.Errorf("invalid recover policy for %s", name)
			}
			critical, err := strconv.ParseBool(unitSection.Key("critical").String())
			if err != nil {
				return watchdogConfig{}, fmt.Errorf("invalid critical policy for %s", name)
			}
			units[name] = protectedUnit{Name: name, ExpectedEnablement: enablement, Recover: recoverEnabled, Critical: critical}
			if len(units) > 32 {
				return watchdogConfig{}, fmt.Errorf("too many protected units")
			}
		}
		return watchdogConfig{
			Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging,
			PollInterval: poll, RecoveryDelay: delay, Backoff: backoff, MaxAttempts: maxAttempts, Units: units,
		}, nil
	})
}
