package contracts

import (
	authcontract "lite-nas/shared/contracts/auth"
	backupcontract "lite-nas/shared/contracts/backuprestore"
	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	litenasintegritycontract "lite-nas/shared/contracts/litenasintegrity"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	processmetricscontract "lite-nas/shared/contracts/processmetrics"
	rbaccontract "lite-nas/shared/contracts/rbac"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
	servicesmanagercontract "lite-nas/shared/contracts/servicesmanager"
	sessionmanagercontract "lite-nas/shared/contracts/sessionmanager"
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
	ServiceBackupRestore: {
		"applications":  {Subject: backupcontract.ApplicationsSubject, Request: backupcontract.ApplicationsRequest{}, Response: backupcontract.ApplicationsResponse{}},
		"backed_up":     {Subject: backupcontract.BackedUpSubject, Request: backupcontract.ApplicationsRequest{}, Response: backupcontract.ApplicationsResponse{}},
		"versions":      {Subject: backupcontract.VersionsSubject, Request: backupcontract.ApplicationRequest{}, Response: backupcontract.VersionsResponse{}},
		"metadata":      {Subject: backupcontract.MetadataSubject, Request: backupcontract.VersionRequest{}, Response: backupcontract.BackupMetadataResponse{}},
		"delete":        {Subject: backupcontract.DeleteSubject, Request: backupcontract.VersionRequest{}, Response: backupcontract.DeleteResponse{}},
		"backup":        {Subject: backupcontract.BackupSubject, Request: backupcontract.BackupRequest{}, Response: backupcontract.BackupResponse{}},
		"backup_all":    {Subject: backupcontract.BackupAllSubject, Request: backupcontract.BackupAllRequest{}, Response: backupcontract.BackupAllResponse{}},
		"diff":          {Subject: backupcontract.DiffSubject, Request: backupcontract.VersionRequest{}, Response: backupcontract.DiffResponse{}},
		"restore":       {Subject: backupcontract.RestoreSubject, Request: backupcontract.RestoreRequest{}, Response: backupcontract.RestoreResponse{}},
		"merge_start":   {Subject: backupcontract.MergeStartSubject, Request: backupcontract.MergeStartRequest{}, Response: backupcontract.MergeStartResponse{}},
		"merge_status":  {Subject: backupcontract.MergeStatusSubject, Request: backupcontract.MergeActionRequest{}, Response: backupcontract.MergeStartResponse{}},
		"merge_read":    {Subject: backupcontract.MergeReadSubject, Request: backupcontract.MergeReadRequest{}, Response: backupcontract.MergeReadResponse{}},
		"merge_write":   {Subject: backupcontract.MergeWriteSubject, Request: backupcontract.MergeWriteRequest{}, Response: backupcontract.MergeWriteResponse{}},
		"merge_apply":   {Subject: backupcontract.MergeApplySubject, Request: backupcontract.MergeActionRequest{}, Response: backupcontract.MergeActionResponse{}},
		"merge_save":    {Subject: backupcontract.MergeSaveSubject, Request: backupcontract.MergeSaveRequest{}, Response: backupcontract.BackupResponse{}},
		"merge_discard": {Subject: backupcontract.MergeDiscardSubject, Request: backupcontract.MergeActionRequest{}, Response: backupcontract.MergeActionResponse{}},
	},
	ServiceServicesManager: {
		"list":    {Subject: servicesmanagercontract.ListSubject, Request: servicesmanagercontract.ListRequest{}, Response: servicesmanagercontract.ListResponse{}},
		"start":   {Subject: servicesmanagercontract.StartSubject, Request: servicesmanagercontract.ActionRequest{}, Response: servicesmanagercontract.ActionResponse{}},
		"stop":    {Subject: servicesmanagercontract.StopSubject, Request: servicesmanagercontract.ActionRequest{}, Response: servicesmanagercontract.ActionResponse{}},
		"restart": {Subject: servicesmanagercontract.RestartSubject, Request: servicesmanagercontract.ActionRequest{}, Response: servicesmanagercontract.ActionResponse{}},
		"enable":  {Subject: servicesmanagercontract.EnableSubject, Request: servicesmanagercontract.ActionRequest{}, Response: servicesmanagercontract.ActionResponse{}},
		"disable": {Subject: servicesmanagercontract.DisableSubject, Request: servicesmanagercontract.ActionRequest{}, Response: servicesmanagercontract.ActionResponse{}},
	},
	ServiceSessionManager: {
		"list":      {Subject: sessionmanagercontract.ListSubject, Request: sessionmanagercontract.ListRequest{}, Response: sessionmanagercontract.ListResponse{}},
		"terminate": {Subject: sessionmanagercontract.TerminateSubject, Request: sessionmanagercontract.ActionRequest{}, Response: sessionmanagercontract.ActionResponse{}},
		"lock":      {Subject: sessionmanagercontract.LockSubject, Request: sessionmanagercontract.ActionRequest{}, Response: sessionmanagercontract.ActionResponse{}},
		"unlock":    {Subject: sessionmanagercontract.UnlockSubject, Request: sessionmanagercontract.ActionRequest{}, Response: sessionmanagercontract.ActionResponse{}},
	},
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
	ServiceProcessMetrics: {
		"get_snapshot": {
			Subject:  processmetricscontract.SnapshotRPCSubject,
			Request:  processmetricscontract.GetSnapshotRequest{},
			Response: processmetricscontract.GetSnapshotResponse{},
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
		"password_change_start": {
			Subject:  authcontract.PasswordChangeStartRPCSubject,
			Request:  authcontract.PasswordChangeStartRequest{},
			Response: authcontract.LoginResponse{},
		},
		"password_change_reply": {
			Subject:  authcontract.PasswordChangeReplyRPCSubject,
			Request:  authcontract.PasswordChangeReplyRequest{},
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
		"issue_caption_grant": {
			Subject:  authcontract.CaptionGrantIssueRPCSubject,
			Request:  authcontract.CaptionGrantIssueRequest{},
			Response: authcontract.CaptionGrantIssueResponse{},
		},
		"validate_caption_grant": {
			Subject:  authcontract.CaptionGrantValidateRPCSubject,
			Request:  authcontract.CaptionGrantValidateRequest{},
			Response: authcontract.CaptionGrantValidateResponse{},
		},
		"revoke_caption_grant": {
			Subject:  authcontract.CaptionGrantRevokeRPCSubject,
			Request:  authcontract.CaptionGrantRevokeRequest{},
			Response: authcontract.CaptionGrantRevokeResponse{},
		},
		"set_lockdown": {
			Subject:  authcontract.SetLockdownRPCSubject,
			Request:  authcontract.SetLockdownRequest{},
			Response: authcontract.SetLockdownResponse{},
		},
	},
	ServiceLiteNASIntegrity: {
		"regenerate_baseline": {
			Subject:  litenasintegritycontract.RegenerateBaselineRPCSubject,
			Request:  litenasintegritycontract.RegenerateBaselineRequest{},
			Response: litenasintegritycontract.RegenerateBaselineResponse{},
		},
		"start_baseline_task": {
			Subject:  litenasintegritycontract.StartBaselineTaskSubject,
			Request:  litenasintegritycontract.StartBaselineRequest{},
			Response: litenasintegritycontract.StartBaselineResponse{},
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
		"check_path_operation": {
			Subject:  rbaccontract.CheckPathOperationRPCSubject,
			Request:  rbaccontract.CheckPathOperationRequest{},
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
