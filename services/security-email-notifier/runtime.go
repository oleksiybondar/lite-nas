package main

import (
	"context"

	servicemodules "lite-nas/services/security-email-notifier/modules"
	sharedcontracts "lite-nas/shared/contracts"
)

const (
	packagedConfigPath = "/etc/lite-nas/security-email-notifier.conf"
	serviceName        = sharedcontracts.ServiceSecurityEmailNotifier
)

// run assembles the security-email-notifier runtime and keeps the process alive
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

	infra.Logger.Info("security email notifier service started", "config", packagedConfigPath)

	<-ctx.Done()
	infra.Logger.Info("security email notifier service stopping")
	return ctx.Err()
}
