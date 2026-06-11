package contracts

import (
	authcontract "lite-nas/shared/contracts/auth"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	rbaccontract "lite-nas/shared/contracts/rbac"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
)

// RPCContract describes one request/reply endpoint contract.
//
// Request and Response are zero-value typed placeholders that encode endpoint
// DTO ownership in one registry map.
type RPCContract struct {
	Subject  string
	Request  any
	Response any
}

type loggingManagerRPCSubjects struct {
	GetAlerts                     string
	GetAlert                      string
	GetActiveAlerts               string
	GetUnacknowledgedActiveAlerts string
	UpdateAlertState              string
	AcknowledgeAlert              string
	MuteAlert                     string
}

// RPCByService defines known RPC endpoints per service identity.
var RPCByService = map[string]map[string]RPCContract{
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
		GetActiveAlerts:               systemloggingmanagercontract.GetActiveAlertsRPCSubject,
		GetUnacknowledgedActiveAlerts: systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		UpdateAlertState:              systemloggingmanagercontract.UpdateAlertStateRPCSubject,
		AcknowledgeAlert:              systemloggingmanagercontract.AcknowledgeAlertRPCSubject,
		MuteAlert:                     systemloggingmanagercontract.MuteAlertRPCSubject,
	}),
	ServiceSecurityLoggingManager: buildLoggingManagerRPCContracts(loggingManagerRPCSubjects{
		GetAlerts:                     securityloggingmanagercontract.GetAlertsRPCSubject,
		GetAlert:                      securityloggingmanagercontract.GetAlertRPCSubject,
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
	return map[string]RPCContract{
		"get_alerts": {
			Subject:  subjects.GetAlerts,
			Request:  loggingmanagercontract.ListAlertsInput{},
			Response: loggingmanagercontract.ListAlertsResponse{},
		},
		"get_alert": {
			Subject:  subjects.GetAlert,
			Request:  loggingmanagercontract.GetAlertInput{},
			Response: loggingmanagercontract.GetAlertResponse{},
		},
		"get_active_alerts": {
			Subject:  subjects.GetActiveAlerts,
			Request:  loggingmanagercontract.ListAlertsInput{},
			Response: loggingmanagercontract.ListAlertsResponse{},
		},
		"get_active_unacknowledged_alerts": {
			Subject:  subjects.GetUnacknowledgedActiveAlerts,
			Request:  loggingmanagercontract.ListAlertsInput{},
			Response: loggingmanagercontract.ListAlertsResponse{},
		},
		"update_alert_state": {
			Subject:  subjects.UpdateAlertState,
			Request:  loggingmanagercontract.UpdateAlertStateInput{},
			Response: loggingmanagercontract.OKResponse{},
		},
		"acknowledge_alert": {
			Subject:  subjects.AcknowledgeAlert,
			Request:  loggingmanagercontract.AcknowledgeAlertInput{},
			Response: loggingmanagercontract.OKResponse{},
		},
		"mute_alert": {
			Subject:  subjects.MuteAlert,
			Request:  loggingmanagercontract.MuteAlertInput{},
			Response: loggingmanagercontract.OKResponse{},
		},
	}
}
