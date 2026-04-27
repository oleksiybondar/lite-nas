package main

import (
	"context"

	"lite-nas/services/auth/modules"
)

const (
	packagedConfigPath = "/etc/lite-nas/auth.conf"
	serviceName        = "auth-service"
	pamServiceName     = "litenas-auth"
)

// run assembles the auth-service runtime and keeps the process alive until
// shutdown while the service contract surface is still being built out.
//
// Parameters:
//   - ctx: process-lifetime context cancelled by OS signal handling
func run(ctx context.Context) error {
	infra, err := modules.NewInfraModule(packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	authModule, err := modules.NewAuthModule(pamServiceName)
	if err != nil {
		return err
	}

	infra.Logger.Info(
		"auth service started",
		"config", packagedConfigPath,
		"pam_service", authModule.ServiceName,
	)

	<-ctx.Done()

	infra.Logger.Info("auth service stopping")
	return ctx.Err()
}
