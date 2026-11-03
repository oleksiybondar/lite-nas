package modules

import (
	"time"

	"lite-nas/shared/loggingmanagerauth"

	serviceconfig "lite-nas/services/resources-monitor/config"
	sharedfileio "lite-nas/shared/fileio"
	"lite-nas/shared/messaging"
	sharedmodules "lite-nas/shared/modules"
)

const defaultAuthRefreshInterval = 24 * time.Hour

// Infra groups runtime infrastructure dependencies for resources-monitor.
type Infra struct {
	sharedmodules.CoreClientAuthInfra
	Config serviceconfig.Config
}

// NewInfraModule loads service config and constructs shared runtime
// infrastructure.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	cfg, err := loadConfigWithDefaultAuthService(configPath, serviceName)
	if err != nil {
		return Infra{}, err
	}

	core, err := buildCoreAuthInfra(serviceName, cfg)
	if err != nil {
		return Infra{}, err
	}

	return Infra{
		CoreClientAuthInfra: core,
		Config:              cfg,
	}, nil
}

func loadConfigWithDefaultAuthService(
	configPath string,
	serviceName string,
) (serviceconfig.Config, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return serviceconfig.Config{}, err
	}

	cfg, err := serviceconfig.LoadConfig(cfgReader)
	if err != nil {
		return serviceconfig.Config{}, err
	}

	authServiceName := cfg.Auth.ServiceName
	if authServiceName == "" {
		authServiceName = serviceName
	}
	cfg.Auth.ServiceName = authServiceName

	return cfg, nil
}

func buildCoreAuthInfra(
	serviceName string,
	cfg serviceconfig.Config,
) (sharedmodules.CoreClientAuthInfra, error) {
	core, err := sharedmodules.NewCoreClientAuthInfra(
		serviceName,
		cfg.Logging,
		cfg.Messaging,
		cfg.Auth,
		defaultAuthRefreshInterval,
	)
	if err != nil {
		return sharedmodules.CoreClientAuthInfra{}, err
	}

	server, err := messaging.NewServer(cfg.Messaging, core.Logger, messaging.NewJSONCodec())
	if err != nil {
		core.Close()
		return sharedmodules.CoreClientAuthInfra{}, err
	}

	core.Server = server
	core.Client = loggingmanagerauth.NewClient(core.Client, core.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken)

	return core, nil
}
