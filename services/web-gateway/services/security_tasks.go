package services

import (
	"context"

	antiviruscontract "lite-nas/shared/contracts/antivirus"
	backgroundtasks "lite-nas/shared/contracts/backgroundtasks"
	integritycontract "lite-nas/shared/contracts/litenasintegrity"
	"lite-nas/shared/messaging"
)

// SecurityTasksService exposes authenticated task operations for security scans.
type SecurityTasksService interface {
	StartAntivirusScan(context.Context, string, string) (backgroundtasks.StartResponse, error)
	StartIntegrityCheck(context.Context, string) (backgroundtasks.StartResponse, error)
	StartIntegrityBaseline(context.Context, string) (backgroundtasks.StartResponse, error)
	AntivirusTaskStatus(context.Context, string, string) (backgroundtasks.StatusCheckResponse, error)
	IntegrityTaskStatus(context.Context, string, string) (backgroundtasks.StatusCheckResponse, error)
	AntivirusInfo(context.Context, string) (antiviruscontract.InfoResponse, error)
	UpdateAntivirusSignatures(context.Context, string) (antiviruscontract.UpdateSignaturesResponse, error)
	IntegrityInfo(context.Context, string) (integritycontract.InfoResponse, error)
	RetryAntivirusTask(context.Context, string, string) (backgroundtasks.RetryResponse, error)
	RetryIntegrityTask(context.Context, string, string) (backgroundtasks.RetryResponse, error)
}

// StartIntegrityBaseline queues an authorized AIDE baseline generation.
func (service securityTasksService) StartIntegrityBaseline(ctx context.Context, token string) (backgroundtasks.StartResponse, error) {
	var response integritycontract.StartBaselineResponse
	err := service.client.Request(ctx, integritycontract.StartBaselineTaskSubject, integritycontract.StartBaselineRequest{AccessToken: token}, &response)
	return backgroundtasks.StartResponse{TaskID: response.TaskID, State: response.State}, err
}

// securityTasksService adapts gateway task operations to service-specific NATS subjects.
type securityTasksService struct{ client messaging.Client }

// AntivirusInfo reads ClamAV engine and signature state through its service RPC.
func (service securityTasksService) AntivirusInfo(ctx context.Context, token string) (antiviruscontract.InfoResponse, error) {
	var response antiviruscontract.InfoResponse
	err := service.client.Request(ctx, antiviruscontract.InfoSubject, antiviruscontract.InfoRequest{AccessToken: token}, &response)
	return response, err
}

// UpdateAntivirusSignatures requests one authorized FreshClam signature update.
func (service securityTasksService) UpdateAntivirusSignatures(ctx context.Context, token string) (antiviruscontract.UpdateSignaturesResponse, error) {
	var response antiviruscontract.UpdateSignaturesResponse
	err := service.client.Request(ctx, antiviruscontract.UpdateSignaturesSubject, antiviruscontract.UpdateSignaturesRequest{AccessToken: token}, &response)
	return response, err
}

// IntegrityInfo reads active AIDE baseline state through its service RPC.
func (service securityTasksService) IntegrityInfo(ctx context.Context, token string) (integritycontract.InfoResponse, error) {
	var response integritycontract.InfoResponse
	err := service.client.Request(ctx, integritycontract.InfoSubject, integritycontract.InfoRequest{AccessToken: token}, &response)
	return response, err
}

// NewSecurityTasksService binds security task operations to the gateway NATS client.
func NewSecurityTasksService(client messaging.Client) SecurityTasksService {
	return securityTasksService{client: client}
}

// StartAntivirusScan starts a folder scan in the antivirus service.
func (service securityTasksService) StartAntivirusScan(ctx context.Context, token, folder string) (backgroundtasks.StartResponse, error) {
	var response antiviruscontract.StartScanResponse
	err := service.client.Request(ctx, antiviruscontract.StartScanSubject, antiviruscontract.StartScanRequest{AccessToken: token, Folder: folder}, &response)
	return backgroundtasks.StartResponse{TaskID: response.TaskID, State: response.State}, err
}

// StartIntegrityCheck queues an AIDE check in the integrity service.
func (service securityTasksService) StartIntegrityCheck(ctx context.Context, token string) (backgroundtasks.StartResponse, error) {
	var response integritycontract.StartCheckResponse
	err := service.client.Request(ctx, integritycontract.StartCheckTaskSubject, integritycontract.StartCheckRequest{AccessToken: token}, &response)
	return backgroundtasks.StartResponse{TaskID: response.TaskID, State: response.State}, err
}

// AntivirusTaskStatus reads one antivirus task snapshot.
func (service securityTasksService) AntivirusTaskStatus(ctx context.Context, token, taskID string) (backgroundtasks.StatusCheckResponse, error) {
	var response antiviruscontract.StatusResponse
	err := service.client.Request(ctx, antiviruscontract.ScanStatusSubject, antiviruscontract.TaskRequest{AccessToken: token, TaskID: taskID}, &response)
	return response, err
}

// IntegrityTaskStatus reads one integrity task snapshot.
func (service securityTasksService) IntegrityTaskStatus(ctx context.Context, token, taskID string) (backgroundtasks.StatusCheckResponse, error) {
	var response integritycontract.TaskStatusResponse
	err := service.client.Request(ctx, integritycontract.CheckTaskStatusSubject, integritycontract.TaskRequest{AccessToken: token, TaskID: taskID}, &response)
	return response, err
}

// RetryAntivirusTask starts another attempt using the saved scan folder.
func (service securityTasksService) RetryAntivirusTask(ctx context.Context, token, taskID string) (backgroundtasks.RetryResponse, error) {
	var response antiviruscontract.RetryResponse
	err := service.client.Request(ctx, antiviruscontract.RetryScanSubject, antiviruscontract.TaskRequest{AccessToken: token, TaskID: taskID}, &response)
	return response, err
}

// RetryIntegrityTask starts another attempt of the saved AIDE check.
func (service securityTasksService) RetryIntegrityTask(ctx context.Context, token, taskID string) (backgroundtasks.RetryResponse, error) {
	var response integritycontract.TaskRetryResponse
	err := service.client.Request(ctx, integritycontract.RetryCheckTaskSubject, integritycontract.TaskRequest{AccessToken: token, TaskID: taskID}, &response)
	return response, err
}
