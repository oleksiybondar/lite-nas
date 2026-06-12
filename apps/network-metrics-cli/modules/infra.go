package modules

import (
	sharedmetricscli "lite-nas/shared/metricscli"
)

// Infra groups the CLI infrastructure dependencies.
type Infra = sharedmetricscli.Infra

// NewInfraModule loads configuration and constructs infrastructure shared by
// the CLI runtime.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	return sharedmetricscli.LoadInfra(configPath, serviceName)
}
