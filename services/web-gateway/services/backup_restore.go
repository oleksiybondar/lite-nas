package services

import (
	"context"
	"fmt"
	"time"

	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/messaging"
)

// BackupRestoreService exposes the backup service's authenticated browser operations.
type BackupRestoreService interface {
	Applications(context.Context, string) ([]string, error)
	BackedUp(context.Context, string) ([]string, error)
	Versions(context.Context, string, string) ([]string, error)
	Metadata(context.Context, string, string, string) (contract.BackupMetadataResponse, error)
	Delete(context.Context, string, string, string) error
	Backup(context.Context, string, string, string, string) (string, error)
	BackupAll(context.Context, string, string, string) ([]contract.BackupAllResult, error)
	Diff(context.Context, string, string, string) ([]contract.Difference, error)
	Restore(context.Context, string, string, string, string) (contract.RestoreResponse, error)
	RestoreByName(context.Context, string, string, string, string) (contract.RestoreByNameResponse, error)
	RestoreAllByName(context.Context, string, string, string) (contract.RestoreAllByNameResponse, error)
	MergeStart(context.Context, string, string, string) (contract.MergeStartResponse, error)
	MergeRead(context.Context, contract.MergeReadRequest) (contract.MergeReadResponse, error)
	MergeStatus(context.Context, contract.MergeActionRequest) (contract.MergeStartResponse, error)
	MergeWrite(context.Context, contract.MergeWriteRequest) (contract.MergeWriteResponse, error)
	MergeApply(context.Context, contract.MergeActionRequest) (contract.MergeActionResponse, error)
	MergeSave(context.Context, contract.MergeSaveRequest) (contract.BackupResponse, error)
	MergeDiscard(context.Context, contract.MergeActionRequest) (contract.MergeActionResponse, error)
}

type backupRestoreService struct{ client messaging.Client }

// Restore can stage a complete archive and restart its configured services.
const restoreRPCTimeout = 5 * time.Minute

// An all-application restore can include many sequential archive writes and restarts.
const restoreAllRPCTimeout = 30 * time.Minute

// NewBackupRestoreService binds backup operations to the shared NATS client.
func NewBackupRestoreService(client messaging.Client) BackupRestoreService {
	return backupRestoreService{client: client}
}

func (s backupRestoreService) request(ctx context.Context, subject string, input, output any) error {
	return s.client.Request(ctx, subject, input, output)
}

func (s backupRestoreService) Applications(ctx context.Context, token string) ([]string, error) {
	var out contract.ApplicationsResponse
	err := s.request(ctx, contract.ApplicationsSubject, contract.ApplicationsRequest{AccessToken: token}, &out)
	return out.Applications, err
}

func (s backupRestoreService) BackedUp(ctx context.Context, token string) ([]string, error) {
	var out contract.ApplicationsResponse
	err := s.request(ctx, contract.BackedUpSubject, contract.ApplicationsRequest{AccessToken: token}, &out)
	return out.Applications, err
}

func (s backupRestoreService) Versions(ctx context.Context, token, app string) ([]string, error) {
	var out contract.VersionsResponse
	err := s.request(ctx, contract.VersionsSubject, contract.ApplicationRequest{AccessToken: token, Application: app}, &out)
	return out.Versions, err
}

// Metadata reads archive labels and attribution without exposing archived files.
func (s backupRestoreService) Metadata(ctx context.Context, token, app, version string) (contract.BackupMetadataResponse, error) {
	var out contract.BackupMetadataResponse
	err := s.request(ctx, contract.MetadataSubject, contract.VersionRequest{AccessToken: token, Application: app, Version: version}, &out)
	return out, err
}

// Delete removes one stored archive selected by its validated application and version.
func (s backupRestoreService) Delete(ctx context.Context, token, app, version string) error {
	var out contract.DeleteResponse
	err := s.request(ctx, contract.DeleteSubject, contract.VersionRequest{AccessToken: token, Application: app, Version: version}, &out)
	if err == nil && !out.Deleted {
		return fmt.Errorf("backup deletion was not confirmed")
	}
	return err
}

func (s backupRestoreService) Backup(ctx context.Context, token, app, name, description string) (string, error) {
	var out contract.BackupResponse
	err := s.request(ctx, contract.BackupSubject, contract.BackupRequest{AccessToken: token, Application: app, Name: name, Description: description}, &out)
	return out.Version, err
}

func (s backupRestoreService) BackupAll(ctx context.Context, token, name, description string) ([]contract.BackupAllResult, error) {
	var out contract.BackupAllResponse
	err := s.request(ctx, contract.BackupAllSubject, contract.BackupAllRequest{AccessToken: token, Name: name, Description: description}, &out)
	return out.Results, err
}

func (s backupRestoreService) Diff(ctx context.Context, token, app, version string) ([]contract.Difference, error) {
	var out contract.DiffResponse
	err := s.request(ctx, contract.DiffSubject, contract.VersionRequest{AccessToken: token, Application: app, Version: version}, &out)
	return out.Differences, err
}

func (s backupRestoreService) Restore(ctx context.Context, token, app, version, mode string) (contract.RestoreResponse, error) {
	var out contract.RestoreResponse
	ctx, cancel := context.WithTimeout(ctx, restoreRPCTimeout)
	defer cancel()
	err := s.request(ctx, contract.RestoreSubject, contract.RestoreRequest{AccessToken: token, Application: app, Version: version, Mode: mode}, &out)
	return out, err
}

// RestoreByName resolves and applies one named backup through the owning service.
func (s backupRestoreService) RestoreByName(ctx context.Context, token, app, name, mode string) (contract.RestoreByNameResponse, error) {
	var out contract.RestoreByNameResponse
	ctx, cancel := context.WithTimeout(ctx, restoreRPCTimeout)
	defer cancel()
	err := s.request(ctx, contract.RestoreByNameSubject, contract.RestoreByNameRequest{AccessToken: token, Application: app, Name: name, Mode: mode}, &out)
	return out, err
}

// RestoreAllByName resolves and applies matching backups through the owning service.
func (s backupRestoreService) RestoreAllByName(ctx context.Context, token, name, mode string) (contract.RestoreAllByNameResponse, error) {
	var out contract.RestoreAllByNameResponse
	ctx, cancel := context.WithTimeout(ctx, restoreAllRPCTimeout)
	defer cancel()
	err := s.request(ctx, contract.RestoreAllByNameSubject, contract.RestoreAllByNameRequest{AccessToken: token, Name: name, Mode: mode}, &out)
	return out, err
}

func (s backupRestoreService) MergeStart(ctx context.Context, token, app, version string) (contract.MergeStartResponse, error) {
	var out contract.MergeStartResponse
	err := s.request(ctx, contract.MergeStartSubject, contract.MergeStartRequest{AccessToken: token, Application: app, Version: version}, &out)
	return out, err
}

func (s backupRestoreService) MergeRead(ctx context.Context, input contract.MergeReadRequest) (contract.MergeReadResponse, error) {
	var out contract.MergeReadResponse
	err := s.request(ctx, contract.MergeReadSubject, input, &out)
	return out, err
}

func (s backupRestoreService) MergeStatus(ctx context.Context, input contract.MergeActionRequest) (contract.MergeStartResponse, error) {
	var out contract.MergeStartResponse
	err := s.request(ctx, contract.MergeStatusSubject, input, &out)
	return out, err
}

func (s backupRestoreService) MergeWrite(ctx context.Context, input contract.MergeWriteRequest) (contract.MergeWriteResponse, error) {
	var out contract.MergeWriteResponse
	err := s.request(ctx, contract.MergeWriteSubject, input, &out)
	return out, err
}

func (s backupRestoreService) MergeApply(ctx context.Context, input contract.MergeActionRequest) (contract.MergeActionResponse, error) {
	var out contract.MergeActionResponse
	ctx, cancel := context.WithTimeout(ctx, restoreRPCTimeout)
	defer cancel()
	err := s.request(ctx, contract.MergeApplySubject, input, &out)
	return out, err
}

func (s backupRestoreService) MergeSave(ctx context.Context, input contract.MergeSaveRequest) (contract.BackupResponse, error) {
	var out contract.BackupResponse
	err := s.request(ctx, contract.MergeSaveSubject, input, &out)
	return out, err
}

func (s backupRestoreService) MergeDiscard(ctx context.Context, input contract.MergeActionRequest) (contract.MergeActionResponse, error) {
	var out contract.MergeActionResponse
	err := s.request(ctx, contract.MergeDiscardSubject, input, &out)
	return out, err
}
