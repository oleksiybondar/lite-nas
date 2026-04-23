package modules

import (
	serviceconfig "lite-nas/apps/system-metrics-cli/config"
	"lite-nas/shared/applog"
	sharedfileio "lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// Infra groups the CLI infrastructure dependencies.
type Infra struct {
	config     serviceconfig.Config
	logger     sharedlogger.Logger
	logCleanup func()
	client     messaging.Client
}

// NewInfraModule builds the CLI infrastructure dependencies.
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
		config:     cfg,
		logger:     log,
		logCleanup: logCleanup,
		client:     client,
	}, nil
}

// Config returns the loaded CLI configuration.
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

// Close releases infrastructure resources.
func (m Infra) Close() {
	if m.client != nil {
		_ = m.client.Drain()
		m.client.Close()
	}

	if m.logCleanup != nil {
		m.logCleanup()
	}
}
