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
	if err := server.RegisterRPC(subjects.GetAlertsRPCSubject, handleGetAlertsRPC(core)); err != nil {
		return err
	}
	if err := server.RegisterRPC(subjects.GetActiveAlertsRPCSubject, handleGetActiveAlertsRPC(core)); err != nil {
		return err
	}
	if err := server.RegisterRPC(subjects.GetUnacknowledgedActiveAlertsRPCSubject, handleGetUnacknowledgedActiveAlertsRPC(core)); err != nil {
		return err
	}
	if err := server.RegisterRPC(subjects.UpdateAlertStateRPCSubject, handleUpdateAlertStateRPC(core)); err != nil {
		return err
	}
	if err := server.RegisterRPC(subjects.AcknowledgeAlertRPCSubject, handleAcknowledgeAlertRPC(core)); err != nil {
		return err
	}
	if err := server.RegisterRPC(subjects.MuteAlertRPCSubject, handleMuteAlertRPC(core)); err != nil {
		return err
	}
	return nil
}
