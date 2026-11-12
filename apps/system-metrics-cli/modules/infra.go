package modules

import (
	sharedmetricscli "lite-nas/shared/metricscli"
)

// Infra groups the CLI infrastructure dependencies.
//
// The exported fields expose constructed runtime dependencies directly. They
// are expected to be treated as logically read-only after initialization.
type Infra = sharedmetricscli.Infra

// NewInfraModule loads configuration and constructs infrastructure shared by
// the CLI runtime.
//
// Parameters:
//   - configPath: filesystem path to the CLI INI configuration file
//   - serviceName: application name used to initialize the logger
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	return sharedmetricscli.LoadInfra(configPath, serviceName)
}
