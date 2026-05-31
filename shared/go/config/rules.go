package config

import (
	"errors"

	"gopkg.in/ini.v1"
)

var errMissingRulesFiles = errors.New("rules files are required")

// RulesConfig defines plain [rules] INI settings used by rule-driven workers.
type RulesConfig struct {
	Files []string
}

// LoadRulesConfig extracts and validates the [rules] section from the INI file.
//
// Expected keys:
//   - files: comma-separated list of rule JSON file paths
func LoadRulesConfig(cfgFile *ini.File) (RulesConfig, error) {
	section := cfgFile.Section("rules")
	normalizedFiles := parseCommaSeparatedValues(section.Key("files").String())
	if len(normalizedFiles) == 0 {
		return RulesConfig{}, errMissingRulesFiles
	}

	return RulesConfig{Files: normalizedFiles}, nil
}
