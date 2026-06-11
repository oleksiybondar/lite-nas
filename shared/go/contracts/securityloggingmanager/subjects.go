package securityloggingmanager

const (
	AlertSubject           = "security-alert"
	AlertOccurrenceSubject = "security-alert-occurrence"

	GetAlertsRPCSubject                     = "security-logging-manager.getAlerts"
	GetAlertRPCSubject                      = "security-logging-manager.getAlert"
	GetActiveAlertsRPCSubject               = "security-logging-manager.getActiveAlerts"
	GetUnacknowledgedActiveAlertsRPCSubject = "security-logging-manager.getUnacknowledgedActiveAlerts"
	UpdateAlertStateRPCSubject              = "security-logging-manager.updateAlertState"
	AcknowledgeAlertRPCSubject              = "security-logging-manager.acknowledgeAlert"
	MuteAlertRPCSubject                     = "security-logging-manager.muteAlert"
)
