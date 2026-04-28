package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	sharedfileio "lite-nas/shared/fileio"
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
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return Infra{}, err
	}

	cfg, err := serviceconfig.LoadConfig(cfgReader)
	if err != nil {
		return Infra{}, err
	}

	core, err := sharedmodules.NewCoreClientServerInfra(serviceName, cfg.Logging, cfg.Messaging)
	if err != nil {
		return Infra{}, err
	}

	return Infra{
		CoreInfra: core,
		Config:    cfg,
	}, nil
}
