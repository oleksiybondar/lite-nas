package modules

import (
	processconfig "lite-nas/services/process-metrics/config"
	sharedconfig "lite-nas/shared/config"
	sharedmodules "lite-nas/shared/modules"
)

// Infra groups service infrastructure dependencies.
type Infra struct {
	sharedmodules.CoreInfra
	Config processconfig.Config
}

// NewInfraModule loads configuration and constructs infrastructure shared by
// the process-metrics runtime.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	core, cfg, err := sharedmodules.LoadCoreInfra(
		configPath,
		serviceName,
		processconfig.LoadConfig,
		sharedmodules.NewCoreClientServerInfra,
		func(cfg processconfig.Config) sharedconfig.MessagingConfig {
			return cfg.Messaging
		},
		func(cfg processconfig.Config) sharedconfig.LoggingConfig {
			return cfg.Logging
		},
	)
	if err != nil {
		return Infra{}, err
	}

	return Infra{
		CoreInfra: core,
		Config:    cfg,
	}, nil
}
