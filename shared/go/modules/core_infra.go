package modules

import (
	"lite-nas/shared/applog"
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// CoreInfra groups the runtime infrastructure dependencies shared by LiteNAS
// services and apps.
type CoreInfra struct {
	Logger     sharedlogger.Logger
	logCleanup func()
	Client     messaging.Client
	Server     messaging.Server
}

// CoreInfraConstructor builds the shared logger and messaging dependencies for
// one runtime.
type CoreInfraConstructor func(
	serviceName string,
	loggingConfig sharedconfig.LoggingConfig,
	messagingConfig sharedconfig.MessagingConfig,
) (CoreInfra, error)

// NewCoreClientInfra constructs the shared logger and outbound messaging
// client for a runtime that only requires client-side messaging.
func NewCoreClientInfra(
	serviceName string,
	loggingConfig sharedconfig.LoggingConfig,
	messagingConfig sharedconfig.MessagingConfig,
) (CoreInfra, error) {
	log, logCleanup, err := applog.NewAppLogger(serviceName, loggingConfig)
	if err != nil {
		return CoreInfra{}, err
	}

	client, err := messaging.NewClient(messagingConfig, log, messaging.NewJSONCodec())
	if err != nil {
		logCleanup()
		return CoreInfra{}, err
	}

	return CoreInfra{
		Logger:     log,
		logCleanup: logCleanup,
		Client:     client,
	}, nil
}

// NewCoreClientServerInfra constructs the shared logger, outbound client, and
// inbound messaging server for a runtime that serves RPCs or subscriptions.
func NewCoreClientServerInfra(
	serviceName string,
	loggingConfig sharedconfig.LoggingConfig,
	messagingConfig sharedconfig.MessagingConfig,
) (CoreInfra, error) {
	core, err := NewCoreClientInfra(serviceName, loggingConfig, messagingConfig)
	if err != nil {
		return CoreInfra{}, err
	}

	server, err := messaging.NewServer(messagingConfig, core.Logger, messaging.NewJSONCodec())
	if err != nil {
		core.Close()
		return CoreInfra{}, err
	}

	core.Server = server
	return core, nil
}

// LoadCoreInfra reads runtime configuration and constructs shared
// infrastructure using the supplied service-specific config loader.
func LoadCoreInfra[T any](
	configPath string,
	serviceName string,
	loadConfig func(fileio.Reader) (T, error),
	constructor CoreInfraConstructor,
	messagingConfig func(T) sharedconfig.MessagingConfig,
	loggingConfig func(T) sharedconfig.LoggingConfig,
) (CoreInfra, T, error) {
	cfgReader, err := fileio.NewFileReader(configPath)
	if err != nil {
		var zero T
		return CoreInfra{}, zero, err
	}

	cfg, err := loadConfig(cfgReader)
	if err != nil {
		var zero T
		return CoreInfra{}, zero, err
	}

	core, err := constructor(serviceName, loggingConfig(cfg), messagingConfig(cfg))
	if err != nil {
		var zero T
		return CoreInfra{}, zero, err
	}

	return core, cfg, nil
}

// Close releases infrastructure resources created by the shared constructor
// helpers.
func (m CoreInfra) Close() {
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
