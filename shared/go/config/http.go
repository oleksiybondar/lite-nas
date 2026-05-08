package config

import (
	"errors"
	"strings"

	"gopkg.in/ini.v1"
)

const defaultHTTPAddress = "127.0.0.1:9090"

var errMissingHTTPAddress = errors.New("http address is required")

// HTTPConfig defines the plain [http] INI settings used at bootstrap time.
type HTTPConfig struct {
	Address string
}

// LoadHTTPConfig extracts and validates the [http] section from the INI file.
func LoadHTTPConfig(cfgFile *ini.File) (HTTPConfig, error) {
	section := cfgFile.Section("http")
	addressKey, err := section.GetKey("address")
	address := ""
	if err != nil {
		address = defaultHTTPAddress
	} else {
		address = strings.TrimSpace(addressKey.String())
	}

	cfg := HTTPConfig{
		Address: address,
	}

	if cfg.Address == "" {
		return HTTPConfig{}, errMissingHTTPAddress
	}

	return cfg, nil
}
