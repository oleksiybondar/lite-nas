package main

import (
	"errors"
	"path/filepath"

	"lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

const packagedConfig = "/etc/lite-nas/operational-safety-automation.conf"

// serviceConfig contains the configured rule location and shared transport credentials.
type serviceConfig struct {
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
	Auth      config.AuthConfig
	Tokens    config.AuthTokenConfig
	RulesFile string
}

// loadServiceConfig validates startup configuration before NATS registration.
func loadServiceConfig(path string) (serviceConfig, error) {
	reader, err := fileio.NewFileReader(path)
	if err != nil {
		return serviceConfig{}, err
	}
	file, err := config.LoadINI(reader)
	if err != nil {
		return serviceConfig{}, err
	}
	messaging, err := config.LoadMessagingConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	logging, err := config.LoadLoggingConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	auth, err := config.LoadAuthConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	tokens, err := config.LoadAuthTokenConfig(file)
	if err != nil {
		return serviceConfig{}, err
	}
	rulesFile := file.Section("automation").Key("rules_file").String()
	if !filepath.IsAbs(rulesFile) || filepath.Clean(rulesFile) != rulesFile || rulesFile == "/" {
		return serviceConfig{}, errors.New("automation.rules_file must be a clean absolute path")
	}
	return serviceConfig{Messaging: messaging, Logging: logging, Auth: auth, Tokens: tokens, RulesFile: rulesFile}, nil
}
