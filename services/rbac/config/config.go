package config

import (
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

const (
	defaultInvalidateInterval    = time.Hour
	defaultRealUserTTL           = 24 * time.Hour
	defaultNonInteractiveUserTTL = 7 * 24 * time.Hour
)

type CacheConfig struct {
	InvalidateInterval    time.Duration
	RealUserTTL           time.Duration
	NonInteractiveUserTTL time.Duration
}

type Config struct {
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
	Cache     CacheConfig
}

func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	messagingConfig, err := sharedconfig.LoadMessagingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}
	loggingConfig, err := sharedconfig.LoadLoggingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	cacheSection := cfgFile.Section("cache")
	invalidateInterval := cacheSection.Key("invalidate_interval").MustDuration(defaultInvalidateInterval)
	realUserTTL := cacheSection.Key("real_user_ttl").MustDuration(defaultRealUserTTL)
	nonInteractiveUserTTL := cacheSection.Key("non_interactive_user_ttl").MustDuration(defaultNonInteractiveUserTTL)

	return Config{
		Messaging: messagingConfig,
		Logging:   loggingConfig,
		Cache: CacheConfig{
			InvalidateInterval:    invalidateInterval,
			RealUserTTL:           realUserTTL,
			NonInteractiveUserTTL: nonInteractiveUserTTL,
		},
	}, nil
}
