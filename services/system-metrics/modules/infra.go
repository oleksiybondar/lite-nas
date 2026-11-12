package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	sharedconfig "lite-nas/shared/config"
	sharedmodules "lite-nas/shared/modules"
)

// Infra groups service infrastructure dependencies.
//
// The exported fields expose constructed runtime dependencies directly. They
// are expected to be treated as logically read-only after initialization.
type Infra struct {
	sharedmodules.CoreInfra
	Config serviceconfig.Config
}

// NewInfraModule loads configuration and constructs infrastructure shared by
// the system-metrics runtime.
//
// Parameters:
//   - configPath: filesystem path to the service INI configuration file
//   - serviceName: application name used to initialize the logger
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
