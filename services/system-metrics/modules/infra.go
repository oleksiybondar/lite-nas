package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	"lite-nas/shared/applog"
	sharedfileio "lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// Infra groups service infrastructure dependencies.
//
// The exported fields expose constructed runtime dependencies directly. They
// are expected to be treated as logically read-only after initialization.
type Infra struct {
	Config     serviceconfig.Config
	Logger     sharedlogger.Logger
	logCleanup func()
	Client     messaging.Client
	Server     messaging.Server
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

	log, logCleanup, err := applog.NewAppLogger(serviceName, cfg.Logging)
	if err != nil {
		return Infra{}, err
	}

	codec := messaging.NewJSONCodec()
	client, err := messaging.NewClient(cfg.Messaging, log, codec)
	if err != nil {
		logCleanup()
		return Infra{}, err
	}

	server, err := messaging.NewServer(cfg.Messaging, log, codec)
	if err != nil {
		_ = client.Drain()
		client.Close()
		logCleanup()
		return Infra{}, err
	}

	return Infra{
		Config:     cfg,
		Logger:     log,
		logCleanup: logCleanup,
		Client:     client,
		Server:     server,
	}, nil
}

// Close releases infrastructure resources created by NewInfraModule.
func (m Infra) Close() {
	if m.Client != nil {
		_ = m.Client.Drain()
		m.Client.Close()
	}

	if m.Server != nil {
		_ = m.Server.Drain()
		m.Server.Close()
	}

	if m.logCleanup != nil {
		m.logCleanup()
	}
}
