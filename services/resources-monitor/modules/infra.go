package modules

import (
	serviceconfig "lite-nas/services/resources-monitor/config"
	sharedfileio "lite-nas/shared/fileio"
	sharedmodules "lite-nas/shared/modules"
)

// Infra groups runtime infrastructure dependencies for resources-monitor.
type Infra struct {
	sharedmodules.CoreInfra
	Config serviceconfig.Config
}

// NewInfraModule loads service config and constructs shared runtime
// infrastructure.
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
