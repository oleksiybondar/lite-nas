package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	"lite-nas/shared/applog"
	sharedfileio "lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// Infra groups service infrastructure dependencies.
type Infra struct {
	config     serviceconfig.Config
	logger     sharedlogger.Logger
	logCleanup func()
	client     messaging.Client
	server     messaging.Server
}

// NewInfraModule builds the service infrastructure dependencies.
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
		config:     cfg,
		logger:     log,
		logCleanup: logCleanup,
		client:     client,
		server:     server,
	}, nil
}

// Config returns the loaded service configuration.
func (m Infra) Config() serviceconfig.Config {
	return m.config
}

// Logger returns the application logger.
func (m Infra) Logger() sharedlogger.Logger {
	return m.logger
}

// Client returns the messaging client.
func (m Infra) Client() messaging.Client {
	return m.client
}

// Server returns the messaging server.
func (m Infra) Server() messaging.Server {
	return m.server
}

// Close releases infrastructure resources.
func (m Infra) Close() {
	if m.client != nil {
		_ = m.client.Drain()
		m.client.Close()
	}

	if m.server != nil {
		_ = m.server.Drain()
		m.server.Close()
	}

	if m.logCleanup != nil {
		m.logCleanup()
	}
}
