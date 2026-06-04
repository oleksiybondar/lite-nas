package main

import (
	"context"

	sharedcontracts "lite-nas/shared/contracts"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedloggingmanagerservice "lite-nas/shared/loggingmanagerservice"
	"lite-nas/shared/roleauth"
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

	subjects := buildNATSSubjects()
	return sharedloggingmanagerservice.Run(
		ctx,
		infra,
		subjects,
		buildAuthorizationPolicy(subjects),
		packagedConfigPath,
		"system logging manager",
	)
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

func buildAuthorizationPolicy(subjects sharedloggingmanagernats.Subjects) sharedloggingmanagernats.AuthorizationPolicy {
	writeRoles := roleauth.AllowedRoles(roleauth.RequirementOperator)

	return sharedloggingmanagernats.AuthorizationPolicy{
		RPCRolesBySubject: map[string][]string{
			subjects.UpdateAlertStateRPCSubject: writeRoles,
			subjects.AcknowledgeAlertRPCSubject: writeRoles,
			subjects.MuteAlertRPCSubject:        writeRoles,
		},
		SubscriptionRolesBySubject: map[string][]string{
			subjects.AlertSubject:           writeRoles,
			subjects.AlertOccurrenceSubject: writeRoles,
		},
	}
}
