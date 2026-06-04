package main

import (
	"context"

	sharedcontracts "lite-nas/shared/contracts"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	sharedloggingmanagernats "lite-nas/shared/loggingmanager/nats"
	sharedloggingmanagerservice "lite-nas/shared/loggingmanagerservice"
	"lite-nas/shared/roleauth"
)

const (
	packagedConfigPath = "/etc/lite-nas/security-logging-manager.conf"
	serviceName        = sharedcontracts.ServiceSecurityLoggingManager
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
		"security logging manager",
	)
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

func buildAuthorizationPolicy(subjects sharedloggingmanagernats.Subjects) sharedloggingmanagernats.AuthorizationPolicy {
	allowedRoles := roleauth.AllowedRoles(roleauth.RequirementSecurity)

	return sharedloggingmanagernats.AuthorizationPolicy{
		RPCRolesBySubject: map[string][]string{
			subjects.GetAlertsRPCSubject:                     allowedRoles,
			subjects.GetAlertRPCSubject:                      allowedRoles,
			subjects.GetActiveAlertsRPCSubject:               allowedRoles,
			subjects.GetUnacknowledgedActiveAlertsRPCSubject: allowedRoles,
			subjects.UpdateAlertStateRPCSubject:              allowedRoles,
			subjects.AcknowledgeAlertRPCSubject:              allowedRoles,
			subjects.MuteAlertRPCSubject:                     allowedRoles,
		},
		SubscriptionRolesBySubject: map[string][]string{
			subjects.AlertSubject:           allowedRoles,
			subjects.AlertOccurrenceSubject: allowedRoles,
		},
	}
}
