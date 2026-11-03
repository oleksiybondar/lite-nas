package controllers

import (
	"context"
	"errors"
	"testing"

	brdto "lite-nas/services/web-gateway/dto/backup_restore"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/shared/authtoken"
	contract "lite-nas/shared/contracts/backuprestore"
)

type backupRestoreStub struct{ err error }

func (s backupRestoreStub) Applications(context.Context, string) ([]string, error) {
	return []string{"app"}, s.err
}

func (s backupRestoreStub) BackedUp(context.Context, string) ([]string, error) {
	return []string{"app"}, s.err
}

func (s backupRestoreStub) Versions(context.Context, string, string) ([]string, error) {
	return []string{"v1"}, s.err
}

func (s backupRestoreStub) Metadata(context.Context, string, string, string) (contract.BackupMetadataResponse, error) {
	return contract.BackupMetadataResponse{Application: "app", Version: "v1", CreatedAt: "2026-01-01T00:00:00Z"}, s.err
}

func (s backupRestoreStub) Delete(context.Context, string, string, string) error {
	return s.err
}

func (s backupRestoreStub) Backup(context.Context, string, string, string, string) (string, error) {
	return "v2", s.err
}

func (s backupRestoreStub) BackupAll(context.Context, string, string, string) ([]contract.BackupAllResult, error) {
	return []contract.BackupAllResult{{Application: "app", Status: "completed"}}, s.err
}

func (s backupRestoreStub) Diff(context.Context, string, string, string) ([]contract.Difference, error) {
	return []contract.Difference{{Path: "/etc/app", Status: "changed"}}, s.err
}

func (s backupRestoreStub) Restore(context.Context, string, string, string, string) (contract.RestoreResponse, error) {
	return contract.RestoreResponse{Restored: true}, s.err
}

// RestoreByName supplies one selected named version to controller tests.
func (s backupRestoreStub) RestoreByName(context.Context, string, string, string, string) (contract.RestoreByNameResponse, error) {
	return contract.RestoreByNameResponse{Version: "v1", Restored: true}, s.err
}

// RestoreAllByName supplies per-application outcomes to controller tests.
func (s backupRestoreStub) RestoreAllByName(context.Context, string, string, string) (contract.RestoreAllByNameResponse, error) {
	return contract.RestoreAllByNameResponse{Results: []contract.RestoreAllByNameResult{{Application: "app", Status: "completed", Version: "v1"}}}, s.err
}

func (s backupRestoreStub) MergeStart(context.Context, string, string, string) (contract.MergeStartResponse, error) {
	return contract.MergeStartResponse{SessionID: "0123456789abcdef0123456789abcdef"}, s.err
}

func (s backupRestoreStub) MergeRead(context.Context, contract.MergeReadRequest) (contract.MergeReadResponse, error) {
	return contract.MergeReadResponse{Done: true}, s.err
}

func (s backupRestoreStub) MergeStatus(context.Context, contract.MergeActionRequest) (contract.MergeStartResponse, error) {
	return contract.MergeStartResponse{SessionID: "0123456789abcdef0123456789abcdef"}, s.err
}

func (s backupRestoreStub) MergeWrite(context.Context, contract.MergeWriteRequest) (contract.MergeWriteResponse, error) {
	return contract.MergeWriteResponse{NextOffset: 1}, s.err
}

func (s backupRestoreStub) MergeApply(context.Context, contract.MergeActionRequest) (contract.MergeActionResponse, error) {
	return contract.MergeActionResponse{Completed: true}, s.err
}

func (s backupRestoreStub) MergeSave(context.Context, contract.MergeSaveRequest) (contract.BackupResponse, error) {
	return contract.BackupResponse{Version: "v2"}, s.err
}

func (s backupRestoreStub) MergeDiscard(context.Context, contract.MergeActionRequest) (contract.MergeActionResponse, error) {
	return contract.MergeActionResponse{Completed: true}, s.err
}

func backupContext() context.Context {
	return middlewares.NewAuthenticatedContext(context.Background(), "token", authtoken.AccessClaims{})
}

// Requirements: backup-restore/IR-004.
// TestBackupRestoreControllerMapsSuccessfulRPCResults covers browser-to-service request mapping.
func TestBackupRestoreControllerMapsSuccessfulRPCResults(t *testing.T) {
	c := NewBackupRestoreController(backupRestoreStub{})
	ctx := backupContext()
	checks := []struct {
		name string
		run  func() error
	}{
		{"applications", func() error {
			o, e := c.Applications(ctx, &brdto.EmptyInput{})
			if e == nil && len(o.Body.Data) != 1 {
				t.Fatal("applications missing")
			}
			return e
		}},
		{"backed-up", func() error {
			o, e := c.BackedUp(ctx, &brdto.EmptyInput{})
			if e == nil && len(o.Body.Data) != 1 {
				t.Fatal("backed-up applications missing")
			}
			return e
		}},
		{"versions", func() error {
			o, e := c.Versions(ctx, &brdto.ApplicationInput{Application: "app"})
			if e == nil && len(o.Body.Data) != 1 {
				t.Fatal("versions missing")
			}
			return e
		}},
		{"metadata", func() error {
			o, e := c.Metadata(ctx, &brdto.VersionInput{Application: "app", Version: "v1"})
			if e == nil && o.Body.Data.Version != "v1" {
				t.Fatal("archive metadata missing")
			}
			return e
		}},
		{"delete", func() error {
			o, e := c.Delete(ctx, &brdto.VersionInput{Application: "app", Version: "v1"})
			if e == nil && !o.Body.Data.Deleted {
				t.Fatal("archive deletion not confirmed")
			}
			return e
		}},
		{"backup", func() error {
			o, e := c.Backup(ctx, &brdto.BackupInput{Application: "app"})
			if e == nil && o.Body.Data["version"] != "v2" {
				t.Fatal("version missing")
			}
			return e
		}},
		{"backup-all", func() error { _, e := c.BackupAll(ctx, &brdto.BackupAllInput{}); return e }},
		{"diff", func() error { _, e := c.Diff(ctx, &brdto.VersionInput{Application: "app", Version: "v1"}); return e }},
		{"restore", func() error {
			input := &brdto.RestoreInput{Application: "app", Version: "v1"}
			input.Body.Mode = "merge"
			_, e := c.Restore(ctx, input)
			return e
		}},
		{"merge-start", func() error {
			_, e := c.MergeStart(ctx, &brdto.MergeStartInput{Application: "app", Version: "v1"})
			return e
		}},
		{"merge-status", func() error { _, e := c.MergeStatus(ctx, &brdto.SessionInput{}); return e }},
		{"merge-read", func() error { _, e := c.MergeRead(ctx, &brdto.MergeReadInput{}); return e }},
		{"merge-write", func() error { _, e := c.MergeWrite(ctx, &brdto.MergeWriteInput{}); return e }},
		{"merge-apply", func() error { _, e := c.MergeApply(ctx, &brdto.SessionInput{}); return e }},
		{"merge-save", func() error { _, e := c.MergeSave(ctx, &brdto.MergeSaveInput{}); return e }},
		{"merge-discard", func() error { _, e := c.MergeDiscard(ctx, &brdto.SessionInput{}); return e }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.run(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Requirements: backup-restore/SR-003.
// TestBackupRestoreControllerRejectsMissingToken checks the authentication boundary.
func TestBackupRestoreControllerRejectsMissingToken(t *testing.T) {
	c := NewBackupRestoreController(backupRestoreStub{})
	_, err := c.Applications(context.Background(), &brdto.EmptyInput{})
	if err == nil {
		t.Fatal("expected missing-token error")
	}
}

// Requirements: backup-restore/IR-004.
// TestBackupRestoreControllerMapsBackendFailure maps service errors to HTTP gateway errors.
func TestBackupRestoreControllerMapsBackendFailure(t *testing.T) {
	c := NewBackupRestoreController(backupRestoreStub{err: errors.New("backend")})
	_, err := c.Applications(backupContext(), &brdto.EmptyInput{})
	if err == nil {
		t.Fatal("expected backend error")
	}
}
