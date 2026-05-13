package loggingmanagercli

import (
	sharedconfig "lite-nas/shared/config"
	sharedfileio "lite-nas/shared/fileio"
	sharedmodules "lite-nas/shared/modules"
)

// LoadInfra constructs messaging client infra for a logging-manager CLI app.
func LoadInfra(configPath string, appName string) (func(), MessagingClient, error) {
	cfgReader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return nil, nil, err
	}

	cfgFile, err := sharedconfig.LoadINI(cfgReader)
	if err != nil {
		return nil, nil, err
	}

	cfg, err := sharedconfig.LoadSharedConfig(cfgFile)
	if err != nil {
		return nil, nil, err
	}

	core, err := sharedmodules.NewCoreClientInfra(appName, cfg.Logging, cfg.Messaging)
	if err != nil {
		return nil, nil, err
	}

	return core.Close, core.Client, nil
}
