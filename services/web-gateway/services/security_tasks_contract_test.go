package services

import (
	"context"
	"testing"

	antivirus "lite-nas/shared/contracts/antivirus"
	integrity "lite-nas/shared/contracts/litenasintegrity"
)

// mustRunSecurityTaskRPC invokes a service operation expected to reach its backend.
func mustRunSecurityTaskRPC(t *testing.T, action func() error) {
	t.Helper()
	if err := action(); err != nil {
		t.Fatal(err)
	}
}

// TestSecurityTaskRPCSubjects verifies each authorized operation reaches its owning backend RPC.
func TestSecurityTaskRPCSubjects(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name, subject string
		call          func(SecurityTasksService) error
	}{
		{"start antivirus", antivirus.StartScanSubject, func(s SecurityTasksService) error { _, err := s.StartAntivirusScan(ctx, "token", "/home"); return err }},
		{"start integrity", integrity.StartCheckTaskSubject, func(s SecurityTasksService) error { _, err := s.StartIntegrityCheck(ctx, "token"); return err }},
		{"start baseline", integrity.StartBaselineTaskSubject, func(s SecurityTasksService) error { _, err := s.StartIntegrityBaseline(ctx, "token"); return err }},
		{"antivirus status", antivirus.ScanStatusSubject, func(s SecurityTasksService) error { _, err := s.AntivirusTaskStatus(ctx, "token", "job"); return err }},
		{"integrity status", integrity.CheckTaskStatusSubject, func(s SecurityTasksService) error { _, err := s.IntegrityTaskStatus(ctx, "token", "job"); return err }},
		{"antivirus info", antivirus.InfoSubject, func(s SecurityTasksService) error { _, err := s.AntivirusInfo(ctx, "token"); return err }},
		{"update signatures", antivirus.UpdateSignaturesSubject, func(s SecurityTasksService) error { _, err := s.UpdateAntivirusSignatures(ctx, "token"); return err }},
		{"integrity info", integrity.InfoSubject, func(s SecurityTasksService) error { _, err := s.IntegrityInfo(ctx, "token"); return err }},
		{"retry antivirus", antivirus.RetryScanSubject, func(s SecurityTasksService) error { _, err := s.RetryAntivirusTask(ctx, "token", "job"); return err }},
		{"retry integrity", integrity.RetryCheckTaskSubject, func(s SecurityTasksService) error { _, err := s.RetryIntegrityTask(ctx, "token", "job"); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &alertsClientStub{}
			mustRunSecurityTaskRPC(t, func() error { return tc.call(NewSecurityTasksService(client)) })
			if client.subject != tc.subject {
				t.Fatalf("subject = %q, want %q", client.subject, tc.subject)
			}
		})
	}
}
