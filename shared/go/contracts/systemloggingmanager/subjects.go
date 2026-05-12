package systemloggingmanager

const (
	AlertSubject           = "system-alert"
	AlertOccurrenceSubject = "system-alert-occurrence"

	GetAlertsRPCSubject                     = "system-logging-manager.getAlerts"
	GetActiveAlertsRPCSubject               = "system-logging-manager.getActiveAlerts"
	GetUnacknowledgedActiveAlertsRPCSubject = "system-logging-manager.getUnacknowledgedActiveAlerts"
	UpdateAlertStateRPCSubject              = "system-logging-manager.updateAlertState"
	AcknowledgeAlertRPCSubject              = "system-logging-manager.acknowledgeAlert"
	MuteAlertRPCSubject                     = "system-logging-manager.muteAlert"
)
