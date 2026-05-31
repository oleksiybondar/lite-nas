package config

import (
	"errors"
	"strings"

	"gopkg.in/ini.v1"
)

var errMissingEmailFrom = errors.New("email from is required")

// EmailConfig defines the [email] bootstrap settings used by notifier services.
type EmailConfig struct {
	To            []string
	CC            []string
	From          string
	SubjectPrefix string
}

// LoadEmailConfig extracts and validates the [email] section from the INI file.
func LoadEmailConfig(cfgFile *ini.File) (EmailConfig, error) {
	section := cfgFile.Section("email")

	config := EmailConfig{
		To:            parseCommaSeparatedValues(section.Key("to").String()),
		CC:            parseCommaSeparatedValues(section.Key("cc").String()),
		From:          strings.TrimSpace(section.Key("from").String()),
		SubjectPrefix: strings.TrimSpace(section.Key("subject_prefix").String()),
	}

	if config.From == "" {
		return EmailConfig{}, errMissingEmailFrom
	}

	return config, nil
}
