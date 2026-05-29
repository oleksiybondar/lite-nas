package main

import (
	"context"

	servicemodules "lite-nas/services/system-email-notifier/modules"
	sharedcontracts "lite-nas/shared/contracts"
)

const (
	packagedConfigPath = "/etc/lite-nas/system-email-notifier.conf"
	serviceName        = sharedcontracts.ServiceSystemEmailNotifier
)

// run assembles the system-email-notifier runtime and keeps the process alive
// until shutdown while the notifier contract surface is being added.
func run(ctx context.Context) error {
	infra, err := servicemodules.NewInfraModule(
		packagedConfigPath,
		serviceName,
	)
	if err != nil {
		return err
	}
	defer infra.Close()

	infra.Logger.Info("system email notifier service started", "config", packagedConfigPath)

	<-ctx.Done()
	infra.Logger.Info("system email notifier service stopping")
	return ctx.Err()
}
