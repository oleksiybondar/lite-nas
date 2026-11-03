package config

import (
	"lite-nas/shared/fileio"
	sharedmetricscli "lite-nas/shared/metricscli"
)

// Config defines runtime configuration for the network metrics CLI.
type Config = sharedmetricscli.Config

// LoadConfig reads CLI configuration from the provided reader.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return sharedmetricscli.LoadConfig(reader)
}
