package main

import (
	"context"
	"os"

	sharedcontracts "lite-nas/shared/contracts"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	sharedloggingmanagercli "lite-nas/shared/loggingmanagercli"
)

const (
	defaultConfigPath = "/etc/lite-nas/security-logging-manager-cli.conf"
	appName           = sharedcontracts.AppSecurityLoggingMgrCLI
)

// run executes CLI command flow.
func run(ctx context.Context, args []string) error {
	subjects := sharedloggingmanagercli.Subjects{
		AlertSubject:                            securityloggingmanagercontract.AlertSubject,
		AlertOccurrenceSubject:                  securityloggingmanagercontract.AlertOccurrenceSubject,
		GetAlertsRPCSubject:                     securityloggingmanagercontract.GetAlertsRPCSubject,
		GetAlertRPCSubject:                      securityloggingmanagercontract.GetAlertRPCSubject,
		GetActiveAlertsRPCSubject:               securityloggingmanagercontract.GetActiveAlertsRPCSubject,
		GetUnacknowledgedActiveAlertsRPCSubject: securityloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		UpdateAlertStateRPCSubject:              securityloggingmanagercontract.UpdateAlertStateRPCSubject,
		AcknowledgeAlertRPCSubject:              securityloggingmanagercontract.AcknowledgeAlertRPCSubject,
		MuteAlertRPCSubject:                     securityloggingmanagercontract.MuteAlertRPCSubject,
	}
	return sharedloggingmanagercli.Run(ctx, args, defaultConfigPath, appName, subjects, loadInfra, os.Stdout)
}

func loadInfra(configPath string, serviceName string) (func(), sharedloggingmanagercli.MessagingClient, error) {
	return sharedloggingmanagercli.LoadInfra(configPath, serviceName)
}
