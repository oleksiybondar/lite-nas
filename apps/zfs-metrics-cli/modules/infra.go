package modules

import (
	sharedmetricscli "lite-nas/shared/metricscli"
)

// Infra groups CLI infrastructure dependencies.
type Infra = sharedmetricscli.Infra

// NewInfraModule loads configuration and constructs CLI infrastructure.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	return sharedmetricscli.LoadInfra(configPath, serviceName)
}
