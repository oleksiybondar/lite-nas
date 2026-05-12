package modules

import (
	"context"

	sharedconfig "lite-nas/shared/config"
	loggingmanagerconfig "lite-nas/shared/config/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	sharedmodules "lite-nas/shared/modules"
)

// Config defines system-logging-manager service runtime configuration.
type Config struct {
	Messaging      sharedconfig.MessagingConfig
	Logging        sharedconfig.LoggingConfig
	LoggingManager loggingmanagerconfig.LoggingManagerConfig
}

// Infra groups service infrastructure and runtime dependencies.
type Infra struct {
	sharedmodules.CoreInfra
	Config             Config
	LoggingManagerCore sharedmodules.LoggingManagerCore
}

// NewInfraModule loads configuration and builds shared runtime dependencies.
func NewInfraModule(ctx context.Context, configPath string, serviceName string) (Infra, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return Infra{}, err
	}

	cfgFile, err := sharedconfig.LoadINI(cfgReader)
	if err != nil {
		return Infra{}, err
	}
	messagingCfg, err := sharedconfig.LoadMessagingConfig(cfgFile)
	if err != nil {
		return Infra{}, err
	}
	loggingCfg, err := sharedconfig.LoadLoggingConfig(cfgFile)
	if err != nil {
		return Infra{}, err
	}
	loggingManagerCfg, err := loggingmanagerconfig.LoadLoggingManagerConfig(cfgFile)
	if err != nil {
		return Infra{}, err
	}
	cfg := Config{
		Messaging:      messagingCfg,
		Logging:        loggingCfg,
		LoggingManager: loggingManagerCfg,
	}

	coreInfra, err := sharedmodules.NewCoreClientServerInfra(serviceName, cfg.Logging, cfg.Messaging)
	if err != nil {
		return Infra{}, err
	}

	loggingManagerCore, err := sharedmodules.NewLoggingManagerCoreModule(ctx, cfg.LoggingManager)
	if err != nil {
		coreInfra.Close()
		return Infra{}, err
	}

	return Infra{
		CoreInfra:          coreInfra,
		Config:             cfg,
		LoggingManagerCore: loggingManagerCore,
	}, nil
}

// Close releases all owned resources.
func (infra Infra) Close() {
	_ = infra.LoggingManagerCore.Close()
	infra.CoreInfra.Close()
}
