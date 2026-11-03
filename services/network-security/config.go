package main

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// securityConfig contains the fixed, startup-validated port scan policy.
type securityConfig struct {
	Messaging  sharedconfig.MessagingConfig
	Auth       sharedconfig.AuthConfig
	Logging    sharedconfig.LoggingConfig
	Threshold  int
	MaxPairs   int
	Window     time.Duration
	Cooldown   time.Duration
	Mitigate   string
	Horizontal patternConfig
	Repeated   patternConfig
}

// patternConfig bounds one additional blocked-attempt detector.
type patternConfig struct {
	Threshold int
	MaxPairs  int
	Window    time.Duration
	Cooldown  time.Duration
	Mitigate  string
}

// mitigationText validates administrator-authored guidance before it enters alert payloads.
func mitigationText(section *ini.Section) (string, error) {
	value := section.Key("mitigate").String()
	if value == "" || len(value) > 4096 || strings.TrimSpace(value) != value {
		return "", errors.New("invalid " + section.Name() + " mitigation guidance")
	}
	for _, char := range value {
		if char < 32 || char == 127 {
			return "", errors.New("invalid " + section.Name() + " mitigation guidance")
		}
	}
	return value, nil
}

// loadPatternConfig reads and validates a detector's independent policy.
func loadPatternConfig(file *ini.File, name string) (patternConfig, error) {
	section := file.Section(name)
	threshold, thresholdErr := strconv.Atoi(section.Key("threshold").MustString("12"))
	maxPairs, maxPairsErr := strconv.Atoi(section.Key("max_pairs").MustString("4096"))
	window, windowErr := time.ParseDuration(section.Key("window").MustString("2m"))
	cooldown, cooldownErr := time.ParseDuration(section.Key("cooldown").MustString("30m"))
	if thresholdErr != nil || maxPairsErr != nil || windowErr != nil || cooldownErr != nil || threshold < 2 || threshold > 1024 ||
		maxPairs < 1 || maxPairs > 16384 || window < time.Second || window > time.Hour || cooldown < time.Second || cooldown > 24*time.Hour {
		return patternConfig{}, errors.New("invalid " + name + " configuration")
	}
	mitigate, err := mitigationText(section)
	if err != nil {
		return patternConfig{}, err
	}
	return patternConfig{Threshold: threshold, MaxPairs: maxPairs, Window: window, Cooldown: cooldown, Mitigate: mitigate}, nil
}

// loadSecurityConfig rejects invalid policy before subscribing to events.
func loadSecurityConfig(reader fileio.Reader) (securityConfig, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (securityConfig, error) {
		section := file.Section("port_scan")
		threshold, thresholdErr := strconv.Atoi(section.Key("distinct_ports").MustString("12"))
		maxPairs, maxPairsErr := strconv.Atoi(section.Key("max_pairs").MustString("4096"))
		window, windowErr := time.ParseDuration(section.Key("window").MustString("2m"))
		cooldown, cooldownErr := time.ParseDuration(section.Key("cooldown").MustString("30m"))
		if thresholdErr != nil || maxPairsErr != nil || windowErr != nil || cooldownErr != nil ||
			threshold < 2 || threshold > 1024 || maxPairs < 1 || maxPairs > 16384 ||
			window < time.Second || window > time.Hour || cooldown < time.Second || cooldown > 24*time.Hour ||
			shared.Auth.ServiceLogin == "" {
			return securityConfig{}, errors.New("invalid network security configuration")
		}
		mitigate, err := mitigationText(section)
		if err != nil {
			return securityConfig{}, err
		}
		horizontal, err := loadPatternConfig(file, "horizontal_scan")
		if err != nil {
			return securityConfig{}, err
		}
		repeated, err := loadPatternConfig(file, "repeated_attempts")
		if err != nil {
			return securityConfig{}, err
		}
		return securityConfig{
			Messaging: shared.Messaging, Auth: shared.Auth, Logging: shared.Logging,
			Threshold: threshold, MaxPairs: maxPairs, Window: window, Cooldown: cooldown, Mitigate: mitigate,
			Horizontal: horizontal, Repeated: repeated,
		}, nil
	})
}
