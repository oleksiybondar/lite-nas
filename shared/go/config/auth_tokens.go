package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

const (
	defaultAuthTokenAccessLifetime = 15 * time.Minute
	defaultAuthTokenClockSkew      = 30 * time.Second
)

var (
	errMissingAuthTokenIssuer    = errors.New("auth token issuer is required")
	errMissingAuthTokenAudience  = errors.New("auth token audience is required")
	errInvalidAuthTokenLifetime  = errors.New("auth token access_lifetime must be greater than zero")
	errInvalidAuthTokenClockSkew = errors.New("auth token clock_skew must not be negative")
)

// AuthTokenConfig defines [auth_tokens] settings for JWT access-token issuing
// and verification.
//
// Key and certificate paths are parsed as plain strings. File presence, PEM
// format, key type, and certificate/key matching are validated by token setup.
type AuthTokenConfig struct {
	Issuer           string
	Audience         string
	AccessLifetime   time.Duration
	ClockSkew        time.Duration
	SigningKey       string
	SigningCert      string
	VerificationCert string
}

// LoadAuthTokenConfig extracts and validates token policy values from the
// [auth_tokens] section of an INI file.
func LoadAuthTokenConfig(cfgFile *ini.File) (AuthTokenConfig, error) {
	section := cfgFile.Section("auth_tokens")

	config := AuthTokenConfig{
		Issuer:           strings.TrimSpace(section.Key("issuer").String()),
		Audience:         strings.TrimSpace(section.Key("audience").String()),
		SigningKey:       strings.TrimSpace(section.Key("signing_key").String()),
		SigningCert:      strings.TrimSpace(section.Key("signing_cert").String()),
		VerificationCert: strings.TrimSpace(section.Key("verification_cert").String()),
	}

	var err error
	config.AccessLifetime, err = parseAuthTokenDuration(section, "access_lifetime", defaultAuthTokenAccessLifetime)
	if err != nil {
		return AuthTokenConfig{}, err
	}

	config.ClockSkew, err = parseAuthTokenDuration(section, "clock_skew", defaultAuthTokenClockSkew)
	if err != nil {
		return AuthTokenConfig{}, err
	}

	if err := validateAuthTokenConfig(config); err != nil {
		return AuthTokenConfig{}, err
	}

	return config, nil
}

func parseAuthTokenDuration(section *ini.Section, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := time.ParseDuration(section.Key(key).MustString(defaultValue.String()))
	if err != nil {
		return 0, fmt.Errorf("parse auth token %s: %w", key, err)
	}

	return value, nil
}

func validateAuthTokenConfig(config AuthTokenConfig) error {
	switch {
	case config.Issuer == "":
		return errMissingAuthTokenIssuer
	case config.Audience == "":
		return errMissingAuthTokenAudience
	case config.AccessLifetime <= 0:
		return errInvalidAuthTokenLifetime
	case config.ClockSkew < 0:
		return errInvalidAuthTokenClockSkew
	default:
		return nil
	}
}
