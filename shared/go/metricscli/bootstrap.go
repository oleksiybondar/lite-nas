package metricscli

import (
	"context"
	"fmt"
	"io"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
	sharedmodules "lite-nas/shared/modules"
)

// Config defines the shared logging and messaging configuration used by
// metrics CLI applications.
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

// Infra groups the shared runtime infrastructure dependencies used by metrics
// CLI applications.
type Infra struct {
	sharedmodules.CoreInfra
	Config Config
}

// RequestClient defines the messaging client contract required by metrics CLI
// runtimes.
type RequestClient interface {
	Request(ctx context.Context, subject string, request any, response any) error
}

// RequestRPC sends a request and decodes the typed RPC response payload.
func RequestRPC[Response any](
	ctx context.Context,
	client RequestClient,
	subject string,
	request any,
) (Response, error) {
	var response Response
	if err := client.Request(ctx, subject, request, &response); err != nil {
		return response, err
	}

	return response, nil
}

// LoadConfig reads shared messaging and logging configuration from the
// provided reader.
func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := sharedconfig.LoadSharedConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging: sharedCfg.Messaging,
		Logging:   sharedCfg.Logging,
	}, nil
}

// LoadInfra loads CLI configuration and constructs shared messaging client
// infrastructure for a metrics CLI runtime.
func LoadInfra(configPath string, serviceName string) (Infra, error) {
	cfgReader, err := fileio.NewFileReader(configPath)
	if err != nil {
		return Infra{}, err
	}

	cfg, err := LoadConfig(cfgReader)
	if err != nil {
		return Infra{}, err
	}

	core, err := sharedmodules.NewCoreClientInfra(serviceName, cfg.Logging, cfg.Messaging)
	if err != nil {
		return Infra{}, err
	}

	return Infra{
		CoreInfra: core,
		Config:    cfg,
	}, nil
}

// Run parses arguments, initializes shared runtime infrastructure, and
// executes the selected metrics CLI command.
func Run[Invocation any](
	ctx context.Context,
	args []string,
	stdout io.Writer,
	serviceName string,
	processArgs func([]string) (Invocation, error),
	configPath func(Invocation) string,
	isHelpRequested func(error) bool,
	printUsage func(io.Writer),
	execute func(context.Context, Invocation, RequestClient, io.Writer) error,
) error {
	invocation, err := processArgs(args)
	if err != nil {
		if isHelpRequested(err) {
			printUsage(stdout)
			return context.Canceled
		}

		return err
	}

	infra, err := LoadInfra(configPath(invocation), serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	return execute(ctx, invocation, infra.Client, stdout)
}

// ExecuteMode dispatches a metrics CLI invocation between current-snapshot and
// history execution paths.
func ExecuteMode[Mode ~string](
	mode Mode,
	current Mode,
	history Mode,
	executeCurrent func() error,
	executeHistory func() error,
) error {
	switch mode {
	case current:
		return executeCurrent()
	case history:
		return executeHistory()
	default:
		return fmt.Errorf("unsupported invocation mode: %s", mode)
	}
}
