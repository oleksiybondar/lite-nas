package modules

import (
	serviceconfig "lite-nas/services/rbac/config"
	sharedconfig "lite-nas/shared/config"
	sharedmodules "lite-nas/shared/modules"
)

type Infra struct {
	sharedmodules.CoreInfra
	Config serviceconfig.Config
}

func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	core, cfg, err := sharedmodules.LoadCoreInfra(
		configPath,
		serviceName,
		serviceconfig.LoadConfig,
		sharedmodules.NewCoreClientServerInfra,
		func(cfg serviceconfig.Config) sharedconfig.MessagingConfig {
			return cfg.Messaging
		},
		func(cfg serviceconfig.Config) sharedconfig.LoggingConfig {
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
