package services

import (
	"context"
	"testing"

	contract "lite-nas/shared/contracts/backuprestore"
)

// Requirements: backup-restore/IR-004.
// TestBackupRestoreServiceUsesContractSubjects exercises every outbound backup RPC mapping.
func TestBackupRestoreServiceUsesContractSubjects(t *testing.T) {
	client := &alertsClientStub{requestFunc: func(_ context.Context, _ string, _ any, _ any) error { return nil }}
	s := NewBackupRestoreService(client)
	checks := []struct {
		name, subject string
		run           func() error
	}{
		{"applications", contract.ApplicationsSubject, func() error { _, e := s.Applications(context.Background(), "token"); return e }},
		{"backed-up", contract.BackedUpSubject, func() error { _, e := s.BackedUp(context.Background(), "token"); return e }},
		{"versions", contract.VersionsSubject, func() error { _, e := s.Versions(context.Background(), "token", "app"); return e }},
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
