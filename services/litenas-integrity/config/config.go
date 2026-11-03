package config

import (
	"errors"
	"fmt"
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

var (
	errMissingAIDESchedule = errors.New("AIDE schedule is required")
	errInvalidAIDESchedule = errors.New("AIDE schedule must be greater than zero hours")
)

// AIDEConfig defines integrity-check runtime settings.
type AIDEConfig struct {
	Schedule time.Duration
}

// Config defines runtime configuration for the LiteNAS integrity service.
//
// It groups configuration by responsibility:
//   - Messaging: NATS connectivity for RPC and alert delivery
//   - Auth: service identity used to acquire background service tokens
//   - AIDE: integrity-check scheduling policy
//   - Logging: application logging behavior
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Auth      sharedconfig.AuthConfig
	AIDE      AIDEConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads integrity-service configuration from a file abstraction and
// returns a validated Config value.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return sharedconfig.LoadConfigWithSharedSections(
		reader,
		func(cfgFile *ini.File, sharedCfg sharedconfig.SharedConfig) (Config, error) {
			aideConfig, err := loadAIDEConfig(cfgFile.Section("AIDE"))
			if err != nil {
				return Config{}, err
			}

			return Config{
				Messaging: sharedCfg.Messaging,
				Auth:      sharedCfg.Auth,
				AIDE:      aideConfig,
				Logging:   sharedCfg.Logging,
			}, nil
		},
	)
}

// loadAIDEConfig validates and parses the `[AIDE]` section.
func loadAIDEConfig(section *ini.Section) (AIDEConfig, error) {
	scheduleHours, err := section.Key("schedule").Int()
	if err != nil {
		return AIDEConfig{}, fmt.Errorf("AIDE schedule: %w", errMissingAIDESchedule)
	}
	if scheduleHours <= 0 {
		return AIDEConfig{}, errInvalidAIDESchedule
	}

	return AIDEConfig{
		Schedule: time.Duration(scheduleHours) * time.Hour,
	}, nil
}
