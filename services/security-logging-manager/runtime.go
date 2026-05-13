package main

import (
	"context"

	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedloggingmanagerservice "lite-nas/shared/loggingmanagerservice"
)

const (
	packagedConfigPath = "/etc/lite-nas/security-logging-manager.conf"
	serviceName        = "security-logging-manager"
)

// run assembles and starts the system logging-manager runtime.
func run(ctx context.Context) error {
	infra, err := sharedloggingmanagerservice.NewInfraModule(ctx, packagedConfigPath, serviceName)
	if err != nil {
		return err
	}
	defer infra.Close()

	return sharedloggingmanagerservice.Run(ctx, infra, buildNATSSubjects(), packagedConfigPath, "security logging manager")
}

func buildNATSSubjects() sharedloggingmanagernats.Subjects {
	return sharedloggingmanagernats.Subjects{
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
}
