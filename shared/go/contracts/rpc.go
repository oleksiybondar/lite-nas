package contracts

import (
	authcontract "lite-nas/shared/contracts/auth"
	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	rbaccontract "lite-nas/shared/contracts/rbac"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
)

// RPCContract describes one request/reply endpoint contract.
type RPCContract struct {
	Subject  string
	Request  any
	Response any
}

type loggingManagerRPCSubjects struct {
	GetAlerts                     string
	GetAlert                      string
	GetAlertOccurrences           string
	GetActiveAlerts               string
	GetUnacknowledgedActiveAlerts string
	UpdateAlertState              string
	AcknowledgeAlert              string
	MuteAlert                     string
}

// RPCByService defines known RPC endpoints per service identity.
var RPCByService = map[string]map[string]RPCContract{
	ServiceDiskMetrics: {
		"get_snapshot": {
			Subject:  diskmetricscontract.SnapshotRPCSubject,
			Request:  diskmetricscontract.GetSnapshotRequest{},
			Response: diskmetricscontract.GetSnapshotResponse{},
		},
		"get_history": {
			Subject:  diskmetricscontract.HistoryRPCSubject,
			Request:  diskmetricscontract.GetHistoryRequest{},
			Response: diskmetricscontract.GetHistoryResponse{},
		},
	},
	ServiceNetworkMetrics: {
		"get_snapshot": {
			Subject:  networkmetricscontract.SnapshotRPCSubject,
			Request:  networkmetricscontract.GetSnapshotRequest{},
			Response: networkmetricscontract.GetSnapshotResponse{},
		},
		"get_history": {
			Subject:  networkmetricscontract.HistoryRPCSubject,
			Request:  networkmetricscontract.GetHistoryRequest{},
			Response: networkmetricscontract.GetHistoryResponse{},
		},
	},
	ServiceServiceMetrics: {
		"get_snapshot": {
			Subject:  servicemetricscontract.SnapshotRPCSubject,
			Request:  servicemetricscontract.GetSnapshotRequest{},
			Response: servicemetricscontract.GetSnapshotResponse{},
		},
		"get_history": {
			Subject:  servicemetricscontract.HistoryRPCSubject,
			Request:  servicemetricscontract.GetHistoryRequest{},
			Response: servicemetricscontract.GetHistoryResponse{},
		},
	},
	ServiceSystemMetrics: {
		"get_snapshot": {
			Subject:  systemmetricscontract.SnapshotRPCSubject,
			Request:  systemmetricscontract.GetSnapshotRequest{},
			Response: systemmetricscontract.GetSnapshotResponse{},
		},
		"get_history": {
			Subject:  systemmetricscontract.HistoryRPCSubject,
			Request:  systemmetricscontract.GetHistoryRequest{},
			Response: systemmetricscontract.GetHistoryResponse{},
		},
	},
	ServiceSystemLoggingManager: buildLoggingManagerRPCContracts(loggingManagerRPCSubjects{
		GetAlerts:                     systemloggingmanagercontract.GetAlertsRPCSubject,
		GetAlert:                      systemloggingmanagercontract.GetAlertRPCSubject,
		GetAlertOccurrences:           systemloggingmanagercontract.GetAlertOccurrencesRPCSubject,
		GetActiveAlerts:               systemloggingmanagercontract.GetActiveAlertsRPCSubject,
		GetUnacknowledgedActiveAlerts: systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		UpdateAlertState:              systemloggingmanagercontract.UpdateAlertStateRPCSubject,
		AcknowledgeAlert:              systemloggingmanagercontract.AcknowledgeAlertRPCSubject,
		MuteAlert:                     systemloggingmanagercontract.MuteAlertRPCSubject,
	}),
	ServiceSecurityLoggingManager: buildLoggingManagerRPCContracts(loggingManagerRPCSubjects{
		GetAlerts:                     securityloggingmanagercontract.GetAlertsRPCSubject,
		GetAlert:                      securityloggingmanagercontract.GetAlertRPCSubject,
		GetAlertOccurrences:           securityloggingmanagercontract.GetAlertOccurrencesRPCSubject,
		GetActiveAlerts:               securityloggingmanagercontract.GetActiveAlertsRPCSubject,
		GetUnacknowledgedActiveAlerts: securityloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		UpdateAlertState:              securityloggingmanagercontract.UpdateAlertStateRPCSubject,
		AcknowledgeAlert:              securityloggingmanagercontract.AcknowledgeAlertRPCSubject,
		MuteAlert:                     securityloggingmanagercontract.MuteAlertRPCSubject,
	}),
	ServiceAuth: {
		"login": {
			Subject:  authcontract.LoginRPCSubject,
			Request:  authcontract.LoginRequest{},
			Response: authcontract.LoginResponse{},
		},
		"refresh": {
			Subject:  authcontract.RefreshRPCSubject,
			Request:  authcontract.RefreshRequest{},
			Response: authcontract.RefreshResponse{},
		},
		"logout": {
			Subject:  authcontract.LogoutRPCSubject,
			Request:  authcontract.LogoutRequest{},
			Response: authcontract.LogoutResponse{},
		},
		"validate_access_token": {
			Subject:  authcontract.ValidateAccessTokenRPCSubject,
			Request:  authcontract.ValidateAccessTokenRequest{},
			Response: authcontract.ValidateAccessTokenResponse{},
		},
		"login_service_token": {
			Subject:  authcontract.ServiceTokenLoginRPCSubject,
			Request:  authcontract.ServiceTokenLoginRequest{},
			Response: authcontract.ServiceTokenLoginResponse{},
		},
		"refresh_service_token": {
			Subject:  authcontract.ServiceTokenRefreshRPCSubject,
			Request:  authcontract.ServiceTokenRefreshRequest{},
			Response: authcontract.ServiceTokenRefreshResponse{},
		},
		"set_lockdown": {
			Subject:  authcontract.SetLockdownRPCSubject,
			Request:  authcontract.SetLockdownRequest{},
			Response: authcontract.SetLockdownResponse{},
		},
	},
	ServiceRBAC: {
		"get_subject_roles": {
			Subject:  rbaccontract.GetSubjectRolesRPCSubject,
			Request:  rbaccontract.GetSubjectRolesRequest{},
			Response: rbaccontract.GetSubjectRolesResponse{},
		},
		"check_path_read": {
			Subject:  rbaccontract.CanReadPathRPCSubject,
			Request:  rbaccontract.CheckPathRequest{},
			Response: rbaccontract.DecisionResponse{},
		},
		"check_path_write": {
			Subject:  rbaccontract.CanWritePathRPCSubject,
			Request:  rbaccontract.CheckPathRequest{},
			Response: rbaccontract.DecisionResponse{},
		},
		"check_path_exec": {
			Subject:  rbaccontract.CanExecPathRPCSubject,
			Request:  rbaccontract.CheckPathRequest{},
			Response: rbaccontract.DecisionResponse{},
		},
		"check_sudo_exec": {
			Subject:  rbaccontract.CanSudoExecRPCSubject,
			Request:  rbaccontract.CheckSudoExecRequest{},
			Response: rbaccontract.DecisionResponse{},
		},
		"invalidate_cache": {
			Subject:  rbaccontract.InvalidateCacheRPCSubject,
			Request:  rbaccontract.InvalidateCacheRequest{},
			Response: rbaccontract.InvalidateCacheResponse{},
		},
	},
}

func buildLoggingManagerRPCContracts(subjects loggingManagerRPCSubjects) map[string]RPCContract {
	contracts := buildLoggingManagerReadRPCContracts(subjects)
	for name, contract := range buildLoggingManagerWriteRPCContracts(subjects) {
		contracts[name] = contract
	}
	return contracts
}

func buildLoggingManagerReadRPCContracts(subjects loggingManagerRPCSubjects) map[string]RPCContract {
	return map[string]RPCContract{
		"get_alerts":                       newRPCContract(subjects.GetAlerts, loggingmanagercontract.ListAlertsInput{}, loggingmanagercontract.ListAlertsResponse{}),
		"get_alert":                        newRPCContract(subjects.GetAlert, loggingmanagercontract.GetAlertInput{}, loggingmanagercontract.GetAlertResponse{}),
		"get_alert_occurrences":            newRPCContract(subjects.GetAlertOccurrences, loggingmanagercontract.GetAlertOccurrencesInput{}, loggingmanagercontract.GetAlertOccurrencesResponse{}),
		"get_active_alerts":                newRPCContract(subjects.GetActiveAlerts, loggingmanagercontract.ListAlertsInput{}, loggingmanagercontract.ListAlertsResponse{}),
		"get_active_unacknowledged_alerts": newRPCContract(subjects.GetUnacknowledgedActiveAlerts, loggingmanagercontract.ListAlertsInput{}, loggingmanagercontract.ListAlertsResponse{}),
	}
}

func buildLoggingManagerWriteRPCContracts(subjects loggingManagerRPCSubjects) map[string]RPCContract {
	return map[string]RPCContract{
		"update_alert_state": newRPCContract(subjects.UpdateAlertState, loggingmanagercontract.UpdateAlertStateInput{}, loggingmanagercontract.OKResponse{}),
		"acknowledge_alert":  newRPCContract(subjects.AcknowledgeAlert, loggingmanagercontract.AcknowledgeAlertInput{}, loggingmanagercontract.OKResponse{}),
		"mute_alert":         newRPCContract(subjects.MuteAlert, loggingmanagercontract.MuteAlertInput{}, loggingmanagercontract.OKResponse{}),
	}
}

func newRPCContract(subject string, request any, response any) RPCContract {
	return RPCContract{Subject: subject, Request: request, Response: response}
}
