package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config holds sensor selection, retention, and shared service infrastructure settings.
type Config struct {
	Sensors   SensorConfig
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

// SensorConfig controls polling, source selection, and readable-name lookup.
// Empty key lists select all discovered readings in that category.
type SensorConfig struct {
	PollInterval    time.Duration
	HistorySize     int
	NamesFile       string
	FirmwareEnabled bool
	TemperatureKeys []string
	VoltageKeys     []string
	FanKeys         []string
	ClockKeys       []string
	ThrottlingKeys  []string
}

// LoadConfig parses the service INI file and validates sensor-specific settings.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return sharedconfig.LoadConfigWithSharedSections(reader, func(file *ini.File, shared sharedconfig.SharedConfig) (Config, error) {
		section := file.Section("sensors")
		interval, err := time.ParseDuration(section.Key("poll_interval").MustString("300ms"))
		if err != nil || interval <= 0 {
			return Config{}, fmt.Errorf("invalid sensors.poll_interval")
		}
		historySize, err := section.Key("history_size").Int()
		if err != nil || historySize < 0 || historySize > 100000 {
			return Config{}, fmt.Errorf("invalid sensors.history_size")
		}
		firmware := true
		if section.HasKey("firmware_enabled") {
			firmware, err = section.Key("firmware_enabled").Bool()
			if err != nil {
				return Config{}, fmt.Errorf("invalid sensors.firmware_enabled: %w", err)
			}
		}
		namesFile := section.Key("names_file").String()
		if namesFile == "" {
			return Config{}, fmt.Errorf("sensors.names_file is required")
		}
		return Config{
			Sensors: SensorConfig{
				PollInterval: interval, HistorySize: historySize,
				NamesFile: namesFile, FirmwareEnabled: firmware,
				TemperatureKeys: splitKeys(section.Key("temperature_keys").String()),
				VoltageKeys:     splitKeys(section.Key("voltage_keys").String()),
				FanKeys:         splitKeys(section.Key("fan_keys").String()),
				ClockKeys:       splitKeys(section.Key("clock_keys").String()),
				ThrottlingKeys:  splitKeys(section.Key("throttling_keys").String()),
			},
			Messaging: shared.Messaging, Logging: shared.Logging,
		}, nil
	})
}

// splitKeys parses comma-separated stable source keys without empty entries.
func splitKeys(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		if key := strings.TrimSpace(part); key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}

// NameEntry supplies a readable header for one stable sensor key.
type NameEntry struct {
	Header string `json:"hdr"`
}

// NameCatalog maps a sensor category and stable key to a readable header.
type NameCatalog map[string]map[string]NameEntry

// LoadNames reads and checks the configured JSON name mapping at startup.
func LoadNames(path string) (NameCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var names NameCatalog
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("parse sensor names: %w", err)
	}
	for category, entries := range names {
		if !knownCategory(category) {
			return nil, fmt.Errorf("unknown sensor name category %q", category)
		}
		for key, entry := range entries {
			if key == "" || strings.TrimSpace(entry.Header) == "" {
				return nil, fmt.Errorf("empty sensor name for %s/%s", category, key)
			}
		}
	}
	return names, nil
}

// Resolve uses a source label, a configured name, or the stable key in that order.
func (catalog NameCatalog) Resolve(category, key, sourceLabel string) string {
	if label := strings.TrimSpace(sourceLabel); label != "" {
		return label
	}
	if entry, ok := catalog[category][key]; ok {
		return entry.Header
	}
	// hwmon device paths vary between boards and kernel boots, while the chip
	// name and attribute index remain useful for a configured readable label.
	if first, last := strings.IndexByte(key, ':'), strings.LastIndexByte(key, ':'); first >= 0 && last > first {
		if entry, ok := catalog[category][key[:first]+key[last:]]; ok {
			return entry.Header
		}
	}
	return key
}

// knownCategory restricts names JSON keys to the published snapshot sections.
func knownCategory(category string) bool {
	switch category {
	case "temperature", "voltage", "fan", "clock", "throttling":
		return true
	default:
		return false
	}
}
