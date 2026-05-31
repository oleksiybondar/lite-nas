package main

import (
	"context"
	"os"

	servicemodules "lite-nas/services/system-email-notifier/modules"
	sharedcontracts "lite-nas/shared/contracts"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	sharedemailnotifier "lite-nas/shared/emailnotifier"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
)

const (
	packagedConfigPath    = "/etc/lite-nas/system-email-notifier.conf"
	packagedTemplatesPath = "/etc/lite-nas/system-email-notifier"
	inputBufferSize       = 16
	serviceName           = sharedcontracts.ServiceSystemEmailNotifier
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

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	validate, err := sharedloggingmanager.NewInputValidator()
	if err != nil {
		return err
	}

	input := make(chan loggingmanagercontract.AlertPayload, inputBufferSize)
	worker, err := sharedemailnotifier.NewWorker(sharedemailnotifier.WorkerConfig{
		Hostname:      hostname,
		TemplatesPath: packagedTemplatesPath,
		Email:         infra.Config.Email,
		SMTP:          infra.Config.SMTP,
	}, input)
	if err != nil {
		return err
	}

	if err = infra.Server.Subscribe(
		systemloggingmanagercontract.AlertSubject,
		sharedemailnotifier.NewAlertSubscriptionHandler(validate, input),
	); err != nil {
		return err
	}

	infra.Logger.Info(
		"system email notifier service started",
		"config", packagedConfigPath,
		"subject", systemloggingmanagercontract.AlertSubject,
		"templates_path", packagedTemplatesPath,
	)

	err = worker.Run(ctx)
	if err == nil || err == context.Canceled {
		infra.Logger.Info("system email notifier service stopping")
	}

	return err
}
