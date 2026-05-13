package loggingmanagerservice

import (
	"context"

	sharedconfig "lite-nas/shared/config"
	loggingmanagerconfig "lite-nas/shared/config/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	sharedmodules "lite-nas/shared/modules"
)

// Config defines runtime configuration required by a logging-manager service.
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
	cfg, err := loadConfig(configPath)
	if err != nil {
		return Infra{}, err
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

func loadConfig(configPath string) (Config, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return Config{}, err
	}

	cfgFile, err := sharedconfig.LoadINI(cfgReader)
	if err != nil {
		return Config{}, err
	}

	messagingCfg, err := sharedconfig.LoadMessagingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	loggingCfg, err := sharedconfig.LoadLoggingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	loggingManagerCfg, err := loggingmanagerconfig.LoadLoggingManagerConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging:      messagingCfg,
		Logging:        loggingCfg,
		LoggingManager: loggingManagerCfg,
	}, nil
}

// Close releases all owned resources.
func (infra Infra) Close() {
	_ = infra.LoggingManagerCore.Close()
	infra.CoreInfra.Close()
}
