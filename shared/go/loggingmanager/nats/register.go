package nats

import (
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

// RegisterSubscriptions registers event ingestion subscriptions.
func RegisterSubscriptions(
	server sharedmessaging.Server,
	core *sharedloggingmanager.Core,
	subjects Subjects,
) error {
	if err := server.Subscribe(subjects.AlertSubject, handleAlert(core)); err != nil {
		return err
	}
	if err := server.Subscribe(subjects.AlertOccurrenceSubject, handleAlertOccurrence(core)); err != nil {
		return err
	}
	return nil
}

// RegisterRPCHandlers registers query and mutation RPC handlers.
func RegisterRPCHandlers(
	server sharedmessaging.Server,
	core *sharedloggingmanager.Core,
	subjects Subjects,
) error {
	handlers := []rpcRegistration{
		{subject: subjects.GetAlertsRPCSubject, handler: handleGetAlertsRPC(core)},
		{subject: subjects.GetAlertRPCSubject, handler: handleGetAlertRPC(core)},
		{subject: subjects.GetActiveAlertsRPCSubject, handler: handleGetActiveAlertsRPC(core)},
		{subject: subjects.GetUnacknowledgedActiveAlertsRPCSubject, handler: handleGetUnacknowledgedActiveAlertsRPC(core)},
		{subject: subjects.UpdateAlertStateRPCSubject, handler: handleUpdateAlertStateRPC(core)},
		{subject: subjects.AcknowledgeAlertRPCSubject, handler: handleAcknowledgeAlertRPC(core)},
		{subject: subjects.MuteAlertRPCSubject, handler: handleMuteAlertRPC(core)},
	}

	for _, registration := range handlers {
		if err := server.RegisterRPC(registration.subject, registration.handler); err != nil {
			return err
		}
	}

	return nil
}

type rpcRegistration struct {
	subject string
	handler sharedmessaging.RPCHandler
}
