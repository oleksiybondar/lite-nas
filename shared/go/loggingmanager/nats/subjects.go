package nats

// Subjects defines messaging subject names used by one logging-manager domain.
type Subjects struct {
	AlertSubject                            string
	AlertOccurrenceSubject                  string
	GetAlertsRPCSubject                     string
	GetActiveAlertsRPCSubject               string
	GetUnacknowledgedActiveAlertsRPCSubject string
	UpdateAlertStateRPCSubject              string
	AcknowledgeAlertRPCSubject              string
	MuteAlertRPCSubject                     string
}
