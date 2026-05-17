package main

import (
	"context"

	sharedcontracts "lite-nas/shared/contracts"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedloggingmanagerservice "lite-nas/shared/loggingmanagerservice"
)

const (
	packagedConfigPath = "/etc/lite-nas/system-logging-manager.conf"
	serviceName        = sharedcontracts.ServiceSystemLoggingManager
)

// run assembles and starts the system logging-manager runtime.
func run(ctx context.Context) error {
	infra, err := sharedloggingmanagerservice.NewInfraModule(ctx, packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	return sharedloggingmanagerservice.Run(ctx, infra, buildNATSSubjects(), packagedConfigPath, "system logging manager")
}

func buildNATSSubjects() sharedloggingmanagernats.Subjects {
	return sharedloggingmanagernats.Subjects{
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
}
