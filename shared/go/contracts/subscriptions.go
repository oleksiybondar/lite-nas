package contracts

import (
	authcontract "lite-nas/shared/contracts/auth"
	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
)

// SubscriptionContract describes one fire-and-forget messaging contract.
type SubscriptionContract struct {
	Subject string
	Payload any
}

// SubscriptionsByService defines known subscribed subjects per service
// identity.
var SubscriptionsByService = map[string]map[string]SubscriptionContract{
	ServiceResourcesMonitor: {
		"network_metrics_snapshot": {
			Subject: networkmetricscontract.SnapshotEventSubject,
			Payload: networkmetricscontract.SnapshotUpdatedEvent{},
		},
		"system_metrics_snapshot": {
			Subject: systemmetricscontract.SnapshotEventSubject,
			Payload: systemmetricscontract.SnapshotUpdatedEvent{},
		},
		"service_metrics_snapshot": {
			Subject: servicemetricscontract.SnapshotEventSubject,
			Payload: servicemetricscontract.SnapshotUpdatedEvent{},
		},
		"disk_metrics_snapshot": {
			Subject: diskmetricscontract.SnapshotEventSubject,
			Payload: diskmetricscontract.SnapshotUpdatedEvent{},
		},
		"zfs_metrics_snapshot": {
			Subject: zfsmetricscontract.SnapshotEventSubject,
			Payload: zfsmetricscontract.SnapshotUpdatedEvent{},
		},
	},
	ServiceSystemLoggingManager: {
		"alert_create": {
			Subject: systemloggingmanagercontract.AlertSubject,
			Payload: loggingmanagercontract.AlertPayload{},
		},
		"alert_occurrence": {
			Subject: systemloggingmanagercontract.AlertOccurrenceSubject,
			Payload: loggingmanagercontract.AlertOccurrencePayload{},
		},
	},
	ServiceSecurityLoggingManager: {
		"alert_create": {
			Subject: securityloggingmanagercontract.AlertSubject,
			Payload: loggingmanagercontract.AlertPayload{},
		},
		"alert_occurrence": {
			Subject: securityloggingmanagercontract.AlertOccurrenceSubject,
			Payload: loggingmanagercontract.AlertOccurrencePayload{},
		},
	},
	ServiceAuth: {
		"lockdown_changed": {
			Subject: authcontract.LockdownChangedEventSubject,
			Payload: authcontract.LockdownChangedEvent{},
		},
	},
}
