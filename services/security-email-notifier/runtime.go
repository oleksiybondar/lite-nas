package main

import (
	"context"

	sharedcontracts "lite-nas/shared/contracts"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	sharedemailnotifier "lite-nas/shared/emailnotifier"
)

const (
	packagedConfigPath    = "/etc/lite-nas/security-email-notifier.conf"
	packagedTemplatesPath = "/etc/lite-nas/security-email-notifier"
	inputBufferSize       = 16
	serviceName           = sharedcontracts.ServiceSecurityEmailNotifier
)

// run assembles the security-email-notifier runtime and keeps the process alive
// until shutdown while the notifier contract surface is being added.
func run(ctx context.Context) error {
	return sharedemailnotifier.RunService(ctx, sharedemailnotifier.ServiceRuntimeConfig{
		ConfigPath:      packagedConfigPath,
		ServiceName:     serviceName,
		TemplatesPath:   packagedTemplatesPath,
		AlertSubject:    securityloggingmanagercontract.AlertSubject,
		StartupMessage:  "security email notifier service started",
		ShutdownMessage: "security email notifier service stopping",
		InputBufferSize: inputBufferSize,
	})
}
