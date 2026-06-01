package emailnotifier

import (
	"context"
	"errors"
	"os"

	sharedconfig "lite-nas/shared/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedfileio "lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmodules "lite-nas/shared/modules"
)

// Infra groups the shared runtime dependencies used by email notifier services.
type Infra struct {
	sharedmodules.CoreInfra
	Config sharedconfig.SharedEmailConfig
}

// ServiceRuntimeConfig defines the service-specific bootstrap inputs for one
// email notifier process.
type ServiceRuntimeConfig struct {
	ConfigPath      string
	ServiceName     string
	TemplatesPath   string
	AlertSubject    string
	StartupMessage  string
	ShutdownMessage string
	InputBufferSize int
}

// NewInfraModule loads shared email-notifier configuration and constructs core runtime infrastructure.
func NewInfraModule(configPath string, serviceName string) (Infra, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return Infra{}, err
	}

	cfgFile, err := sharedconfig.LoadINI(cfgReader)
	if err != nil {
		return Infra{}, err
	}

	cfg, err := sharedconfig.LoadSharedEmailConfig(cfgFile)
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

// RunService executes one email notifier service runtime from bootstrap to shutdown.
func RunService(ctx context.Context, config ServiceRuntimeConfig) error {
	infra, err := NewInfraModule(config.ConfigPath, config.ServiceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	validate, input, worker, err := BuildWorkerRuntime(
		config.TemplatesPath,
		infra.Config.Email,
		infra.Config.SMTP,
		config.InputBufferSize,
	)
	if err != nil {
		return err
	}

	if err = infra.Server.Subscribe(
		config.AlertSubject,
		NewAlertSubscriptionHandler(validate, input),
	); err != nil {
		return err
	}

	infra.Logger.Info(
		config.StartupMessage,
		"config", config.ConfigPath,
		"subject", config.AlertSubject,
		"templates_path", config.TemplatesPath,
	)

	return RunWorker(ctx, worker, infra.Logger, config.ShutdownMessage)
}

// BuildWorkerRuntime resolves host-local worker dependencies for one notifier process.
func BuildWorkerRuntime(
	templatesPath string,
	emailConfig sharedconfig.EmailConfig,
	smtpConfig sharedconfig.SMTPConfig,
	inputBufferSize int,
) (
	sharedloggingmanager.InputValidator,
	chan loggingmanagercontract.AlertPayload,
	Worker,
	error,
) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, nil, Worker{}, err
	}

	validate, err := sharedloggingmanager.NewInputValidator()
	if err != nil {
		return nil, nil, Worker{}, err
	}

	input := make(chan loggingmanagercontract.AlertPayload, inputBufferSize)
	worker, err := NewWorker(WorkerConfig{
		Hostname:      hostname,
		TemplatesPath: templatesPath,
		Email:         emailConfig,
		SMTP:          smtpConfig,
	}, input)
	if err != nil {
		return nil, nil, Worker{}, err
	}

	return validate, input, worker, nil
}

// RunWorker executes the notifier worker and logs clean shutdown paths.
func RunWorker(
	ctx context.Context,
	worker Worker,
	logger sharedlogger.Logger,
	shutdownMessage string,
) error {
	err := worker.Run(ctx)
	if err == nil || errors.Is(err, context.Canceled) {
		logger.Info(shutdownMessage)
	}

	return err
}
