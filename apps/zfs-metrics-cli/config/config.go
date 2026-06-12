package config

import (
	"lite-nas/shared/fileio"
	sharedmetricscli "lite-nas/shared/metricscli"
)

// Config defines runtime configuration for zfs-metrics-cli.
type Config = sharedmetricscli.Config

// LoadConfig loads shared messaging/logging sections from the provided reader.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return sharedmetricscli.LoadConfig(reader)
}
