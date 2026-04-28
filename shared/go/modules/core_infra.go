package modules

import (
	"lite-nas/shared/applog"
	sharedconfig "lite-nas/shared/config"
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
