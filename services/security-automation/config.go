package main

import (
	"errors"
	"path/filepath"

	"lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

const packagedAutomationConfig = "/etc/lite-nas/security-automation.conf"

// automationConfig binds the reviewed rule file and action policy to shared service infrastructure.
type automationConfig struct {
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
	Auth      config.AuthConfig
	Tokens    config.AuthTokenConfig
	RulesFile string
	Mode      string
}

// loadAutomationConfig validates startup paths and the required service-wide action mode.
func loadAutomationConfig(path string) (automationConfig, error) {
	reader, err := fileio.NewFileReader(path)
	if err != nil {
		return automationConfig{}, err
	}
	file, err := config.LoadINI(reader)
	if err != nil {
		return automationConfig{}, err
	}
	messaging, err := config.LoadMessagingConfig(file)
	if err != nil {
		return automationConfig{}, err
	}
	logging, err := config.LoadLoggingConfig(file)
	if err != nil {
		return automationConfig{}, err
	}
	auth, err := config.LoadAuthConfig(file)
	if err != nil {
		return automationConfig{}, err
	}
	tokens, err := config.LoadAuthTokenConfig(file)
	if err != nil {
		return automationConfig{}, err
	}
	section := file.Section("automation")
	rulesFile := section.Key("rules_file").String()
	mode := section.Key("mode").String()
	if !filepath.IsAbs(rulesFile) || filepath.Clean(rulesFile) != rulesFile || rulesFile == "/" {
		return automationConfig{}, errors.New("automation rules_file must be a clean absolute path")
	}
	if mode != "strict" && mode != "permissive" {
		return automationConfig{}, errors.New("automation mode must be strict or permissive")
	}
	return automationConfig{Messaging: messaging, Logging: logging, Auth: auth, Tokens: tokens, RulesFile: rulesFile, Mode: mode}, nil
}
