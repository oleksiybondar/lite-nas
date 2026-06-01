package main

import (
	"context"

	sharedcontracts "lite-nas/shared/contracts"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	sharedemailnotifier "lite-nas/shared/emailnotifier"
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
	return sharedemailnotifier.RunService(ctx, sharedemailnotifier.ServiceRuntimeConfig{
		ConfigPath:      packagedConfigPath,
		ServiceName:     serviceName,
		TemplatesPath:   packagedTemplatesPath,
		AlertSubject:    systemloggingmanagercontract.AlertSubject,
		StartupMessage:  "system email notifier service started",
		ShutdownMessage: "system email notifier service stopping",
		InputBufferSize: inputBufferSize,
	})
}
