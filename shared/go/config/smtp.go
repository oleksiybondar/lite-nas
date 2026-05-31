package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

const defaultSMTPHELO = "localhost"

var (
	errMissingSMTPHost = errors.New("smtp host is required")
	errInvalidSMTPPort = errors.New("smtp port must be between 1 and 65535")
)

// SMTPConfig defines the [smtp] bootstrap settings used by notifier services.
type SMTPConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
	HELO    string
}

// LoadSMTPConfig extracts and validates the [smtp] section from the INI file.
func LoadSMTPConfig(cfgFile *ini.File) (SMTPConfig, error) {
	section := cfgFile.Section("smtp")

	timeout, err := time.ParseDuration(section.Key("timeout").MustString("10s"))
	if err != nil {
		return SMTPConfig{}, err
	}

	config := SMTPConfig{
		Host:    strings.TrimSpace(section.Key("host").String()),
		Port:    section.Key("port").MustInt(25),
		Timeout: timeout,
		HELO:    strings.TrimSpace(section.Key("helo").MustString(defaultSMTPHELO)),
	}

	if config.Host == "" {
		return SMTPConfig{}, errMissingSMTPHost
	}

	if config.Port < 1 || config.Port > 65535 {
		return SMTPConfig{}, fmt.Errorf("%w: %d", errInvalidSMTPPort, config.Port)
	}

	if config.HELO == "" {
		config.HELO = defaultSMTPHELO
	}

	return config, nil
}
