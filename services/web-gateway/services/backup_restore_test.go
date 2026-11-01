package services

import (
	"context"
	"testing"
	"time"

	contract "lite-nas/shared/contracts/backuprestore"
)

// Requirements: backup-restore/IR-004.
// TestBackupRestoreServiceUsesContractSubjects exercises every outbound backup RPC mapping.
func TestBackupRestoreServiceUsesContractSubjects(t *testing.T) {
	client := &alertsClientStub{requestFunc: func(_ context.Context, _ string, _ any, output any) error {
		if response, ok := output.(*contract.DeleteResponse); ok {
			response.Deleted = true
		}
		return nil
	}}
	s := NewBackupRestoreService(client)
	checks := []struct {
		name, subject string
		run           func() error
	}{
		{"applications", contract.ApplicationsSubject, func() error { _, e := s.Applications(context.Background(), "token"); return e }},
		{"backed-up", contract.BackedUpSubject, func() error { _, e := s.BackedUp(context.Background(), "token"); return e }},
		{"versions", contract.VersionsSubject, func() error { _, e := s.Versions(context.Background(), "token", "app"); return e }},
		{"metadata", contract.MetadataSubject, func() error { _, e := s.Metadata(context.Background(), "token", "app", "version"); return e }},
		{"delete", contract.DeleteSubject, func() error { return s.Delete(context.Background(), "token", "app", "version") }},
		{"backup", contract.BackupSubject, func() error { _, e := s.Backup(context.Background(), "token", "app", "name", "desc"); return e }},
		{"backup-all", contract.BackupAllSubject, func() error { _, e := s.BackupAll(context.Background(), "token", "name", "desc"); return e }},
		{"diff", contract.DiffSubject, func() error { _, e := s.Diff(context.Background(), "token", "app", "version"); return e }},
		{"restore", contract.RestoreSubject, func() error { _, e := s.Restore(context.Background(), "token", "app", "version", "merge"); return e }},
		{"merge-start", contract.MergeStartSubject, func() error { _, e := s.MergeStart(context.Background(), "token", "app", "version"); return e }},
		{"merge-read", contract.MergeReadSubject, func() error { _, e := s.MergeRead(context.Background(), contract.MergeReadRequest{}); return e }},
		{"merge-status", contract.MergeStatusSubject, func() error { _, e := s.MergeStatus(context.Background(), contract.MergeActionRequest{}); return e }},
		{"merge-write", contract.MergeWriteSubject, func() error { _, e := s.MergeWrite(context.Background(), contract.MergeWriteRequest{}); return e }},
		{"merge-apply", contract.MergeApplySubject, func() error { _, e := s.MergeApply(context.Background(), contract.MergeActionRequest{}); return e }},
		{"merge-save", contract.MergeSaveSubject, func() error { _, e := s.MergeSave(context.Background(), contract.MergeSaveRequest{}); return e }},
		{"merge-discard", contract.MergeDiscardSubject, func() error { _, e := s.MergeDiscard(context.Background(), contract.MergeActionRequest{}); return e }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.run(); err != nil {
				t.Fatal(err)
			}
			if client.subject != check.subject {
				t.Fatalf("subject=%q, want %q", client.subject, check.subject)
			}
		})
	}
}

// Requirements: backup-restore/IR-004.
// TestRestoreActionsAllowArchiveWork checks that restore calls outlive the short default RPC timeout.
func TestRestoreActionsAllowArchiveWork(t *testing.T) {
	for _, action := range []struct {
		name string
		run  func(context.Context, BackupRestoreService) error
	}{
		{name: "restore", run: func(ctx context.Context, service BackupRestoreService) error {
			_, err := service.Restore(ctx, "token", "app", "version", "merge")
			return err
		}},
		{name: "merge apply", run: func(ctx context.Context, service BackupRestoreService) error {
			_, err := service.MergeApply(ctx, contract.MergeActionRequest{})
			return err
		}},
	} {
		t.Run(action.name, func(t *testing.T) {
			client := &alertsClientStub{requestFunc: func(ctx context.Context, _ string, _ any, _ any) error {
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) < 4*time.Minute {
					t.Errorf("restore RPC deadline = %v, want at least four minutes remaining", deadline)
				}
				return nil
			}}
			if err := action.run(context.Background(), NewBackupRestoreService(client)); err != nil {
				t.Fatal(err)
			}
		})
	}
}
