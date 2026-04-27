package modules

import (
	serviceconfig "lite-nas/services/web-gateway/config"
	"lite-nas/shared/applog"
	sharedfileio "lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// Infra groups gateway infrastructure dependencies.
//
// The exported fields expose constructed runtime dependencies directly. They
// are expected to be treated as logically read-only after initialization.
type Infra struct {
	Config     serviceconfig.Config
	Logger     sharedlogger.Logger
	logCleanup func()
	Client     messaging.Client
}

// NewInfraModule loads configuration and constructs infrastructure shared by
// the gateway runtime.
//
// Parameters:
//   - configPath: filesystem path to the gateway INI configuration file
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

	log, logCleanup, err := applog.NewAppLogger(serviceName, cfg.Logging)
	if err != nil {
		return Infra{}, err
	}

	client, err := messaging.NewClient(cfg.Messaging, log, messaging.NewJSONCodec())
	if err != nil {
		logCleanup()
		return Infra{}, err
	}

	return Infra{
		Config:     cfg,
		Logger:     log,
		logCleanup: logCleanup,
		Client:     client,
	}, nil
}

// Close releases infrastructure resources created by NewInfraModule.
func (m Infra) Close() {
	if m.Client != nil {
		_ = m.Client.Drain()
		m.Client.Close()
	}

	if m.logCleanup != nil {
		m.logCleanup()
	}
}
