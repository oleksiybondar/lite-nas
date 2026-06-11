package main

import (
	"context"
	"os"

	sharedcontracts "lite-nas/shared/contracts"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	sharedloggingmanagercli "lite-nas/shared/loggingmanagercli"
)

const (
	defaultConfigPath = "/etc/lite-nas/system-logging-manager-cli.conf"
	appName           = sharedcontracts.AppSystemLoggingManagerCLI
)

// run executes CLI command flow.
func run(ctx context.Context, args []string) error {
	subjects := sharedloggingmanagercli.Subjects{
		AlertSubject:                            systemloggingmanagercontract.AlertSubject,
		AlertOccurrenceSubject:                  systemloggingmanagercontract.AlertOccurrenceSubject,
		GetAlertsRPCSubject:                     systemloggingmanagercontract.GetAlertsRPCSubject,
		GetAlertRPCSubject:                      systemloggingmanagercontract.GetAlertRPCSubject,
		GetActiveAlertsRPCSubject:               systemloggingmanagercontract.GetActiveAlertsRPCSubject,
		GetUnacknowledgedActiveAlertsRPCSubject: systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		UpdateAlertStateRPCSubject:              systemloggingmanagercontract.UpdateAlertStateRPCSubject,
		AcknowledgeAlertRPCSubject:              systemloggingmanagercontract.AcknowledgeAlertRPCSubject,
		MuteAlertRPCSubject:                     systemloggingmanagercontract.MuteAlertRPCSubject,
	}
	return sharedloggingmanagercli.Run(ctx, args, defaultConfigPath, appName, subjects, loadInfra, os.Stdout)
}

func loadInfra(configPath string, serviceName string) (func(), sharedloggingmanagercli.MessagingClient, error) {
	return sharedloggingmanagercli.LoadInfra(configPath, serviceName)
}
